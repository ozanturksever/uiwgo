// resource_cache.go
// Advanced resource caching and deduplication system for performance optimization

package golid

import (
	"crypto/md5"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

// ------------------------------------
// 🗄️ Advanced Cache Types
// ------------------------------------

// CachePolicy defines cache behavior policies
type CachePolicy int

const (
	CachePolicyLRU  CachePolicy = iota // Least Recently Used
	CachePolicyLFU                     // Least Frequently Used
	CachePolicyTTL                     // Time To Live
	CachePolicyFIFO                    // First In First Out
)

// CacheEntry represents an enhanced cache entry with metadata
type CacheEntryAdvanced struct {
	key         string
	value       interface{}
	timestamp   time.Time
	lastAccess  time.Time
	accessCount uint64
	ttl         time.Duration
	size        int64
	tags        []string
	metadata    map[string]interface{}
	mutex       sync.RWMutex
}

// AdvancedResourceCache provides sophisticated caching with multiple policies
type AdvancedResourceCache struct {
	entries     map[string]*CacheEntryAdvanced
	policy      CachePolicy
	maxSize     int
	maxMemory   int64
	currentSize int64
	ttl         time.Duration
	mutex       sync.RWMutex

	// Statistics
	hits      uint64
	misses    uint64
	evictions uint64
	sets      uint64

	// Event handlers
	onHit   func(key string, value interface{})
	onMiss  func(key string)
	onEvict func(key string, value interface{})
	onSet   func(key string, value interface{})

	// Background cleanup
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// CacheConfig provides configuration for advanced cache
type CacheConfig struct {
	Policy          CachePolicy
	MaxSize         int
	MaxMemory       int64
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
	OnHit           func(key string, value interface{})
	OnMiss          func(key string)
	OnEvict         func(key string, value interface{})
	OnSet           func(key string, value interface{})
}

// ------------------------------------
// 🏗️ Cache Creation
// ------------------------------------

// NewAdvancedResourceCache creates a new advanced resource cache
func NewAdvancedResourceCache(config CacheConfig) *AdvancedResourceCache {
	// Set defaults
	if config.MaxSize == 0 {
		config.MaxSize = 1000
	}
	if config.MaxMemory == 0 {
		config.MaxMemory = 100 * 1024 * 1024 // 100MB
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 10 * time.Minute
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = time.Minute
	}

	cache := &AdvancedResourceCache{
		entries:     make(map[string]*CacheEntryAdvanced),
		policy:      config.Policy,
		maxSize:     config.MaxSize,
		maxMemory:   config.MaxMemory,
		ttl:         config.DefaultTTL,
		onHit:       config.OnHit,
		onMiss:      config.OnMiss,
		onEvict:     config.OnEvict,
		onSet:       config.OnSet,
		stopCleanup: make(chan bool, 1),
	}

	// Start background cleanup
	cache.startCleanup(config.CleanupInterval)

	return cache
}

// ------------------------------------
// 🔍 Cache Operations
// ------------------------------------

// Get retrieves a value from the cache with advanced tracking
func (c *AdvancedResourceCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	entry, exists := c.entries[key]
	c.mutex.RUnlock()

	if !exists {
		c.recordMiss(key)
		return nil, false
	}

	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	// Check if entry is expired
	if c.isExpired(entry) {
		c.mutex.Lock()
		delete(c.entries, key)
		c.currentSize -= entry.size
		c.mutex.Unlock()
		c.recordMiss(key)
		return nil, false
	}

	// Update access metadata
	entry.lastAccess = time.Now()
	entry.accessCount++

	c.recordHit(key, entry.value)
	return entry.value, true
}

// Set stores a value in the cache with metadata
func (c *AdvancedResourceCache) Set(key string, value interface{}, options ...CacheSetOptions) {
	var opts CacheSetOptions
	if len(options) > 0 {
		opts = options[0]
	}

	ttl := opts.TTL
	if ttl == 0 {
		ttl = c.ttl
	}

	size := opts.Size
	if size == 0 {
		size = c.estimateSize(value)
	}

	entry := &CacheEntryAdvanced{
		key:         key,
		value:       value,
		timestamp:   time.Now(),
		lastAccess:  time.Now(),
		accessCount: 0,
		ttl:         ttl,
		size:        size,
		tags:        opts.Tags,
		metadata:    opts.Metadata,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict entries
	c.ensureCapacity(size)

	// Store the entry
	if existingEntry, exists := c.entries[key]; exists {
		c.currentSize -= existingEntry.size
	}

	c.entries[key] = entry
	c.currentSize += size
	c.sets++

	c.recordSet(key, value)
}

// CacheSetOptions provides options for cache set operations
type CacheSetOptions struct {
	TTL      time.Duration
	Size     int64
	Tags     []string
	Metadata map[string]interface{}
}

// ------------------------------------
// 🔄 Cache Management
// ------------------------------------

// ensureCapacity ensures cache has capacity for new entry
func (c *AdvancedResourceCache) ensureCapacity(newEntrySize int64) {
	// Check memory limit
	for c.currentSize+newEntrySize > c.maxMemory && len(c.entries) > 0 {
		c.evictOne()
	}

	// Check size limit
	for len(c.entries) >= c.maxSize && c.maxSize > 0 {
		c.evictOne()
	}
}

// evictOne evicts one entry based on the cache policy
func (c *AdvancedResourceCache) evictOne() {
	if len(c.entries) == 0 {
		return
	}

	var keyToEvict string
	var entryToEvict *CacheEntryAdvanced

	switch c.policy {
	case CachePolicyLRU:
		keyToEvict, entryToEvict = c.findLRU()
	case CachePolicyLFU:
		keyToEvict, entryToEvict = c.findLFU()
	case CachePolicyTTL:
		keyToEvict, entryToEvict = c.findExpired()
	case CachePolicyFIFO:
		keyToEvict, entryToEvict = c.findFIFO()
	default:
		keyToEvict, entryToEvict = c.findLRU()
	}

	if keyToEvict != "" {
		delete(c.entries, keyToEvict)
		c.currentSize -= entryToEvict.size
		c.evictions++
		c.recordEvict(keyToEvict, entryToEvict.value)
	}
}

// findLRU finds the least recently used entry
func (c *AdvancedResourceCache) findLRU() (string, *CacheEntryAdvanced) {
	var oldestKey string
	var oldestEntry *CacheEntryAdvanced
	var oldestTime time.Time

	for key, entry := range c.entries {
		entry.mutex.RLock()
		lastAccess := entry.lastAccess
		entry.mutex.RUnlock()

		if oldestKey == "" || lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestEntry = entry
			oldestTime = lastAccess
		}
	}

	return oldestKey, oldestEntry
}

// findLFU finds the least frequently used entry
func (c *AdvancedResourceCache) findLFU() (string, *CacheEntryAdvanced) {
	var leastKey string
	var leastEntry *CacheEntryAdvanced
	var leastCount uint64 = ^uint64(0) // Max uint64

	for key, entry := range c.entries {
		entry.mutex.RLock()
		accessCount := entry.accessCount
		entry.mutex.RUnlock()

		if accessCount < leastCount {
			leastKey = key
			leastEntry = entry
			leastCount = accessCount
		}
	}

	return leastKey, leastEntry
}

// findExpired finds an expired entry
func (c *AdvancedResourceCache) findExpired() (string, *CacheEntryAdvanced) {
	for key, entry := range c.entries {
		if c.isExpired(entry) {
			return key, entry
		}
	}
	// If no expired entry found, fall back to LRU
	return c.findLRU()
}

// findFIFO finds the first in, first out entry
func (c *AdvancedResourceCache) findFIFO() (string, *CacheEntryAdvanced) {
	var oldestKey string
	var oldestEntry *CacheEntryAdvanced
	var oldestTime time.Time

	for key, entry := range c.entries {
		entry.mutex.RLock()
		timestamp := entry.timestamp
		entry.mutex.RUnlock()

		if oldestKey == "" || timestamp.Before(oldestTime) {
			oldestKey = key
			oldestEntry = entry
			oldestTime = timestamp
		}
	}

	return oldestKey, oldestEntry
}

// isExpired checks if an entry is expired
func (c *AdvancedResourceCache) isExpired(entry *CacheEntryAdvanced) bool {
	if entry.ttl == 0 {
		return false
	}
	return time.Since(entry.timestamp) > entry.ttl
}

// ------------------------------------
// 🏷️ Tag-based Operations
// ------------------------------------

// GetByTag retrieves all entries with a specific tag
func (c *AdvancedResourceCache) GetByTag(tag string) map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string]interface{})
	for key, entry := range c.entries {
		entry.mutex.RLock()
		hasTag := c.hasTag(entry.tags, tag)
		value := entry.value
		entry.mutex.RUnlock()

		if hasTag && !c.isExpired(entry) {
			result[key] = value
		}
	}

	return result
}

// InvalidateByTag removes all entries with a specific tag
func (c *AdvancedResourceCache) InvalidateByTag(tag string) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0
	keysToDelete := make([]string, 0)

	for key, entry := range c.entries {
		entry.mutex.RLock()
		hasTag := c.hasTag(entry.tags, tag)
		entry.mutex.RUnlock()

		if hasTag {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		if entry, exists := c.entries[key]; exists {
			delete(c.entries, key)
			c.currentSize -= entry.size
			count++
		}
	}

	return count
}

// hasTag checks if tags slice contains a specific tag
func (c *AdvancedResourceCache) hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// ------------------------------------
// 🧹 Cache Cleanup
// ------------------------------------

// startCleanup starts background cleanup routine
func (c *AdvancedResourceCache) startCleanup(interval time.Duration) {
	c.cleanupTicker = time.NewTicker(interval)
	go c.cleanupLoop()
}

// cleanupLoop runs the cleanup routine
func (c *AdvancedResourceCache) cleanupLoop() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// cleanupExpired removes all expired entries
func (c *AdvancedResourceCache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	keysToDelete := make([]string, 0)
	for key, entry := range c.entries {
		if c.isExpired(entry) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		if entry, exists := c.entries[key]; exists {
			delete(c.entries, key)
			c.currentSize -= entry.size
			c.evictions++
		}
	}
}

// Stop stops the cache cleanup routine
func (c *AdvancedResourceCache) Stop() {
	select {
	case c.stopCleanup <- true:
	default:
	}
}

// Clear removes all entries from the cache
func (c *AdvancedResourceCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntryAdvanced)
	c.currentSize = 0
}

// ------------------------------------
// 📊 Statistics and Monitoring
// ------------------------------------

// GetAdvancedStats returns detailed cache statistics
func (c *AdvancedResourceCache) GetAdvancedStats() AdvancedCacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := AdvancedCacheStats{
		Size:        len(c.entries),
		MaxSize:     c.maxSize,
		CurrentSize: c.currentSize,
		MaxMemory:   c.maxMemory,
		Hits:        c.hits,
		Misses:      c.misses,
		Evictions:   c.evictions,
		Sets:        c.sets,
		Policy:      c.policy,
	}

	// Calculate hit ratio
	total := c.hits + c.misses
	if total > 0 {
		stats.HitRatio = float64(c.hits) / float64(total)
	}

	// Calculate memory usage
	stats.MemoryUsage = float64(c.currentSize) / float64(c.maxMemory)

	return stats
}

// AdvancedCacheStats provides detailed cache statistics
type AdvancedCacheStats struct {
	Size        int
	MaxSize     int
	CurrentSize int64
	MaxMemory   int64
	Hits        uint64
	Misses      uint64
	Evictions   uint64
	Sets        uint64
	HitRatio    float64
	MemoryUsage float64
	Policy      CachePolicy
}

// ------------------------------------
// 🔧 Utility Functions
// ------------------------------------

// estimateSize estimates the size of a value
func (c *AdvancedResourceCache) estimateSize(value interface{}) int64 {
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int32, int64, uint, uint32, uint64:
		return 8
	case float32, float64:
		return 8
	case bool:
		return 1
	default:
		return 64 // Default estimate
	}
}

// generateKey generates a cache key from multiple parts
func GenerateCacheKey(parts ...interface{}) string {
	hash := fnv.New64a()
	for _, part := range parts {
		hash.Write([]byte(fmt.Sprintf("%v", part)))
	}
	return fmt.Sprintf("%x", hash.Sum64())
}

// generateMD5Key generates an MD5 hash key
func GenerateMD5Key(data string) string {
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ------------------------------------
// 📈 Event Handlers
// ------------------------------------

// recordHit records a cache hit
func (c *AdvancedResourceCache) recordHit(key string, value interface{}) {
	c.hits++
	if c.onHit != nil {
		c.onHit(key, value)
	}
}

// recordMiss records a cache miss
func (c *AdvancedResourceCache) recordMiss(key string) {
	c.misses++
	if c.onMiss != nil {
		c.onMiss(key)
	}
}

// recordEvict records a cache eviction
func (c *AdvancedResourceCache) recordEvict(key string, value interface{}) {
	if c.onEvict != nil {
		c.onEvict(key, value)
	}
}

// recordSet records a cache set operation
func (c *AdvancedResourceCache) recordSet(key string, value interface{}) {
	if c.onSet != nil {
		c.onSet(key, value)
	}
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// MockAdvancedCache creates a mock cache for testing
func MockAdvancedCache() *AdvancedResourceCache {
	return NewAdvancedResourceCache(CacheConfig{
		Policy:    CachePolicyLRU,
		MaxSize:   100,
		MaxMemory: 1024 * 1024, // 1MB
	})
}

// FillCache fills the cache with test data
func (c *AdvancedResourceCache) FillCache(count int) {
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		value := fmt.Sprintf("test-value-%d", i)
		c.Set(key, value)
	}
}
