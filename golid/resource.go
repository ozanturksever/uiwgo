// resource.go
// Core resource management and createResource implementation for async data loading

package golid

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 🔄 Resource Types
// ------------------------------------

// ResourceState represents the current state of a resource
type ResourceState[T any] struct {
	Loading bool
	Data    *T
	Error   error
	Version uint64 // For cache invalidation
}

// AsyncResource represents an async data resource with reactive state
type AsyncResource[T any] struct {
	id         uint64
	signal     *ReactiveSignal[ResourceState[T]]
	fetcher    func() (T, error)
	owner      *Owner
	cache      *ResourceCache
	options    ResourceOptions
	mutex      sync.RWMutex
	cancelFunc context.CancelFunc
	lastFetch  time.Time
	retryCount int
}

// ResourceOptions provides configuration for resource creation
type ResourceOptions struct {
	Name         string
	CacheKey     string
	TTL          time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	Timeout      time.Duration
	OnSuccess    func(interface{})
	OnError      func(error)
	OnLoading    func(bool)
	Preload      bool
	Dependencies []interface{} // Reactive dependencies that trigger refetch
}

// ResourceCache manages caching and deduplication of resources
type ResourceCache struct {
	entries   map[string]*CacheEntry
	maxSize   int
	ttl       time.Duration
	mutex     sync.RWMutex
	hits      uint64
	misses    uint64
	evictions uint64
}

// CacheEntry represents a cached resource value
type CacheEntry struct {
	value     interface{}
	timestamp time.Time
	ttl       time.Duration
	hits      uint64
	size      int64
	mutex     sync.RWMutex
}

// Global resource management
var (
	asyncResourceIdCounter uint64
	globalCache            *ResourceCache
	cacheOnce              sync.Once
)

// SuspenseException is thrown when a resource is loading and should suspend rendering
type SuspenseException struct {
	Message string
}

func (e SuspenseException) Error() string {
	return e.Message
}

// NewSuspenseException creates a new suspense exception
func NewSuspenseException(message string) SuspenseException {
	return SuspenseException{Message: message}
}

// ------------------------------------
// 🏗️ Resource Creation
// ------------------------------------

// CreateResource creates a new async resource with reactive state management
func CreateResource[T any](fetcher func() (T, error), options ...ResourceOptions) *AsyncResource[T] {
	var opts ResourceOptions
	if len(options) > 0 {
		opts = options[0]
	}

	// Set default options
	if opts.TTL == 0 {
		opts.TTL = 5 * time.Minute
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}
	if opts.RetryDelay == 0 {
		opts.RetryDelay = time.Second
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	// Create initial state
	initialState := ResourceState[T]{
		Loading: false,
		Data:    nil,
		Error:   nil,
		Version: 0,
	}

	// Create reactive signal for state
	stateSignal := &ReactiveSignal[ResourceState[T]]{
		id:          atomic.AddUint64(&asyncResourceIdCounter, 1),
		value:       initialState,
		subscribers: make(map[uint64]*Computation),
		owner:       getCurrentOwner(),
		mutex:       sync.RWMutex{},
	}

	resource := &AsyncResource[T]{
		id:      atomic.AddUint64(&asyncResourceIdCounter, 1),
		signal:  stateSignal,
		fetcher: fetcher,
		owner:   getCurrentOwner(),
		cache:   getGlobalCache(),
		options: opts,
	}

	// Register with owner for cleanup
	if resource.owner != nil {
		resource.owner.addCleanup(func() {
			resource.cleanupAsync()
		})
	}

	// Set up reactive dependencies if provided
	if len(opts.Dependencies) > 0 {
		CreateEffect(func() {
			// Track dependencies
			for _, dep := range opts.Dependencies {
				if signal, ok := dep.(interface{ Get() interface{} }); ok {
					signal.Get() // Track dependency
				}
			}
			// Refetch when dependencies change
			resource.RefetchAsync()
		}, resource.owner)
	}

	// Preload if requested
	if opts.Preload {
		go resource.fetchAsync()
	}

	return resource
}

// ------------------------------------
// 🔍 Resource Methods
// ------------------------------------

// Read returns the current data, triggering a fetch if not loaded
func (r *AsyncResource[T]) Read() T {
	state := r.signal.Get()

	// If we have data and it's not stale, return it
	if state.Data != nil && !r.isStale() {
		return *state.Data
	}

	// If not loading and no data, start loading
	if !state.Loading && state.Data == nil {
		go r.fetchAsync()
	}

	// If we have stale data, return it while fetching fresh data
	if state.Data != nil {
		if !state.Loading {
			go r.fetchAsync()
		}
		return *state.Data
	}

	// No data available, this will suspend in Suspense boundary
	panic(NewSuspenseException("Resource loading"))
}

// State returns the current resource state
func (r *AsyncResource[T]) State() ResourceState[T] {
	return r.signal.Get()
}

// RefetchAsync forces a new fetch, ignoring cache
func (r *AsyncResource[T]) RefetchAsync() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Cancel any ongoing fetch
	if r.cancelFunc != nil {
		r.cancelFunc()
	}

	// Clear cache entry if exists
	if r.options.CacheKey != "" {
		r.cache.Delete(r.options.CacheKey)
	}

	// Reset retry count
	r.retryCount = 0

	// Start new fetch
	go r.fetchAsync()
}

// Mutate updates the resource data directly (optimistic updates)
func (r *AsyncResource[T]) Mutate(value T) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	newState := ResourceState[T]{
		Loading: false,
		Data:    &value,
		Error:   nil,
		Version: r.signal.value.Version + 1,
	}

	r.signal.Set(newState)

	// Update cache if caching is enabled
	if r.options.CacheKey != "" {
		r.cache.Set(r.options.CacheKey, value, r.options.TTL)
	}
}

// ------------------------------------
// 🔄 Internal Methods
// ------------------------------------

// fetchAsync performs the actual data fetching with error handling and retries
func (r *AsyncResource[T]) fetchAsync() {
	r.mutex.Lock()

	// Check if already loading
	if r.signal.value.Loading {
		r.mutex.Unlock()
		return
	}

	// Set loading state
	loadingState := r.signal.value
	loadingState.Loading = true
	loadingState.Error = nil
	r.signal.Set(loadingState)
	r.mutex.Unlock()

	// Notify loading callback
	if r.options.OnLoading != nil {
		r.options.OnLoading(true)
	}

	// Check cache first
	if r.options.CacheKey != "" {
		if cached, found := r.cache.Get(r.options.CacheKey); found {
			if data, ok := cached.(T); ok {
				r.setSuccess(data)
				return
			}
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), r.options.Timeout)
	r.cancelFunc = cancel
	defer cancel()

	// Perform fetch with retry logic
	var result T
	var err error

	for attempt := 0; attempt <= r.options.MaxRetries; attempt++ {
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			r.setError(ctx.Err())
			return
		default:
		}

		// Perform the fetch
		result, err = r.fetcher()
		if err == nil {
			r.setSuccess(result)
			return
		}

		// If this isn't the last attempt, wait before retrying
		if attempt < r.options.MaxRetries {
			r.retryCount++
			select {
			case <-time.After(r.options.RetryDelay * time.Duration(attempt+1)): // Exponential backoff
			case <-ctx.Done():
				r.setError(ctx.Err())
				return
			}
		}
	}

	// All retries failed
	r.setError(fmt.Errorf("resource fetch failed after %d attempts: %w", r.options.MaxRetries+1, err))
}

// setSuccess updates the resource state with successful data
func (r *AsyncResource[T]) setSuccess(data T) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	newState := ResourceState[T]{
		Loading: false,
		Data:    &data,
		Error:   nil,
		Version: r.signal.value.Version + 1,
	}

	r.signal.Set(newState)
	r.lastFetch = time.Now()
	r.retryCount = 0

	// Update cache
	if r.options.CacheKey != "" {
		r.cache.Set(r.options.CacheKey, data, r.options.TTL)
	}

	// Notify success callback
	if r.options.OnSuccess != nil {
		r.options.OnSuccess(data)
	}

	// Notify loading callback
	if r.options.OnLoading != nil {
		r.options.OnLoading(false)
	}
}

// setError updates the resource state with error
func (r *AsyncResource[T]) setError(err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	newState := ResourceState[T]{
		Loading: false,
		Data:    r.signal.value.Data, // Keep existing data if any
		Error:   err,
		Version: r.signal.value.Version + 1,
	}

	r.signal.Set(newState)

	// Notify error callback
	if r.options.OnError != nil {
		r.options.OnError(err)
	}

	// Notify loading callback
	if r.options.OnLoading != nil {
		r.options.OnLoading(false)
	}
}

// isStale checks if the resource data is stale based on TTL
func (r *AsyncResource[T]) isStale() bool {
	if r.options.TTL == 0 {
		return false // No TTL means never stale
	}
	return time.Since(r.lastFetch) > r.options.TTL
}

// cleanupAsync cleans up the resource
func (r *AsyncResource[T]) cleanupAsync() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.cancelFunc != nil {
		r.cancelFunc()
	}
}

// ------------------------------------
// 🗄️ Cache Implementation
// ------------------------------------

// getGlobalCache returns the global resource cache
func getGlobalCache() *ResourceCache {
	cacheOnce.Do(func() {
		globalCache = NewResourceCache(1000, 10*time.Minute) // Default: 1000 entries, 10min TTL
	})
	return globalCache
}

// NewResourceCache creates a new resource cache
func NewResourceCache(maxSize int, defaultTTL time.Duration) *ResourceCache {
	return &ResourceCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     defaultTTL,
	}
}

// Get retrieves a value from the cache
func (c *ResourceCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	entry, exists := c.entries[key]
	c.mutex.RUnlock()

	if !exists {
		atomic.AddUint64(&c.misses, 1)
		return nil, false
	}

	entry.mutex.RLock()
	defer entry.mutex.RUnlock()

	// Check if entry is expired
	if entry.ttl > 0 && time.Since(entry.timestamp) > entry.ttl {
		// Entry is expired, remove it
		c.Delete(key)
		atomic.AddUint64(&c.misses, 1)
		return nil, false
	}

	// Update hit count
	atomic.AddUint64(&entry.hits, 1)
	atomic.AddUint64(&c.hits, 1)

	return entry.value, true
}

// Set stores a value in the cache
func (c *ResourceCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict entries
	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	if ttl == 0 {
		ttl = c.ttl
	}

	c.entries[key] = &CacheEntry{
		value:     value,
		timestamp: time.Now(),
		ttl:       ttl,
		hits:      0,
		size:      estimateSize(value),
	}
}

// Delete removes a value from the cache
func (c *ResourceCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *ResourceCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// evictLRU evicts the least recently used entry
func (c *ResourceCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		entry.mutex.RLock()
		timestamp := entry.timestamp
		entry.mutex.RUnlock()

		if oldestKey == "" || timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = timestamp
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		atomic.AddUint64(&c.evictions, 1)
	}
}

// Stats returns cache statistics
func (c *ResourceCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return CacheStats{
		Size:      len(c.entries),
		MaxSize:   c.maxSize,
		Hits:      atomic.LoadUint64(&c.hits),
		Misses:    atomic.LoadUint64(&c.misses),
		Evictions: atomic.LoadUint64(&c.evictions),
	}
}

// CacheStats provides cache statistics
type CacheStats struct {
	Size      int
	MaxSize   int
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

// estimateSize provides a rough estimate of value size
func estimateSize(value interface{}) int64 {
	// This is a simple estimation - in a real implementation,
	// you might want more sophisticated size calculation
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	default:
		return 64 // Default estimate
	}
}

// ------------------------------------
// 🔄 Utility Functions
// ------------------------------------

// PreloadResource starts loading a resource without reading it
func PreloadResource[T any](resource *AsyncResource[T]) {
	go resource.fetchAsync()
}

// InvalidateResource clears the cache for a resource
func InvalidateResource[T any](resource *AsyncResource[T]) {
	if resource.options.CacheKey != "" {
		resource.cache.Delete(resource.options.CacheKey)
	}
}

// CreateDerivedResource creates a resource that depends on other resources
func CreateDerivedResource[T, U any](
	sourceResource *AsyncResource[T],
	transform func(T) (U, error),
	options ...ResourceOptions,
) *AsyncResource[U] {
	return CreateResource(func() (U, error) {
		sourceData := sourceResource.Read()
		return transform(sourceData)
	}, options...)
}
