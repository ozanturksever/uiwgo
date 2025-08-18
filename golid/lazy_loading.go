// lazy_loading.go
// Lazy component loading and code-splitting utilities for performance optimization

package golid

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🔄 Lazy Loading Types
// ------------------------------------

// LazyComponent represents a lazily loaded component
type LazyComponent struct {
	id         uint64
	loader     func() (gomponents.Node, error)
	component  gomponents.Node
	loaded     bool
	loading    bool
	error      error
	cache      bool
	owner      *Owner
	mutex      sync.RWMutex
	onLoad     func(gomponents.Node)
	onError    func(error)
	retryCount int
	maxRetries int
}

// LazyOptions provides configuration for lazy components
type LazyOptions struct {
	Name       string
	Cache      bool
	MaxRetries int
	OnLoad     func(gomponents.Node)
	OnError    func(error)
	Owner      *Owner
	Preload    bool
	Timeout    time.Duration
}

// LazyRegistry manages lazy component registration and loading
type LazyRegistry struct {
	components map[string]*LazyComponent
	loaders    map[string]func() (gomponents.Node, error)
	mutex      sync.RWMutex
	stats      LazyStats
}

// LazyStats provides statistics about lazy loading
type LazyStats struct {
	TotalComponents int
	LoadedCount     int
	ErrorCount      int
	CacheHits       int
	CacheMisses     int
}

// Global lazy loading management
var (
	lazyIdCounter  uint64
	globalRegistry *LazyRegistry
	registryOnce   sync.Once
)

// ------------------------------------
// 🏗️ Lazy Component Creation
// ------------------------------------

// Lazy creates a new lazy component with the given loader function
func Lazy(loader func() (gomponents.Node, error), options ...LazyOptions) *LazyComponent {
	var opts LazyOptions
	if len(options) > 0 {
		opts = options[0]
	}

	// Set default options
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	owner := opts.Owner
	if owner == nil {
		owner = getCurrentOwner()
	}

	lazy := &LazyComponent{
		id:         atomic.AddUint64(&lazyIdCounter, 1),
		loader:     loader,
		loaded:     false,
		loading:    false,
		cache:      opts.Cache,
		owner:      owner,
		onLoad:     opts.OnLoad,
		onError:    opts.OnError,
		maxRetries: opts.MaxRetries,
	}

	// Register with owner for cleanup
	if owner != nil {
		owner.addCleanup(func() {
			lazy.cleanup()
		})
	}

	// Register with global registry
	getGlobalRegistry().registerComponent(opts.Name, lazy)

	// Preload if requested
	if opts.Preload {
		go lazy.load()
	}

	return lazy
}

// LazyWithSuspense creates a lazy component with automatic suspense handling
func LazyWithSuspense(
	loader func() (gomponents.Node, error),
	fallback gomponents.Node,
	options ...LazyOptions,
) (*LazyComponent, *SuspenseBoundary) {
	lazy := Lazy(loader, options...)
	boundary := Suspense(fallback)

	// Connect lazy loading to suspense
	CreateEffect(func() {
		if lazy.IsLoading() {
			boundary.suspend()
		} else if lazy.IsLoaded() {
			boundary.resume()
		}
	}, lazy.owner)

	return lazy, boundary
}

// ------------------------------------
// 🎭 Lazy Component Rendering
// ------------------------------------

// Render renders the lazy component, loading it if necessary
func (l *LazyComponent) Render() gomponents.Node {
	l.mutex.RLock()
	loaded := l.loaded
	loading := l.loading
	component := l.component
	err := l.error
	l.mutex.RUnlock()

	// If loaded, return the component
	if loaded && component != nil {
		return component
	}

	// If there's an error, handle it
	if err != nil {
		return l.renderError(err)
	}

	// If not loading, start loading
	if !loading {
		go l.load()
	}

	// Suspend while loading
	panic(NewSuspenseException("Lazy component loading"))
}

// RenderWithFallback renders with a fallback while loading
func (l *LazyComponent) RenderWithFallback(fallback gomponents.Node) gomponents.Node {
	l.mutex.RLock()
	loaded := l.loaded
	loading := l.loading
	component := l.component
	err := l.error
	l.mutex.RUnlock()

	// If loaded, return the component
	if loaded && component != nil {
		return component
	}

	// If there's an error, handle it
	if err != nil {
		return l.renderError(err)
	}

	// If not loading, start loading
	if !loading {
		go l.load()
	}

	// Return fallback while loading
	return fallback
}

// renderError renders an error state
func (l *LazyComponent) renderError(err error) gomponents.Node {
	return gomponents.El("div",
		gomponents.Attr("class", "lazy-error"),
		gomponents.Attr("style", "padding: 20px; border: 1px solid #ff6b6b; background: #ffe0e0; border-radius: 4px;"),
		gomponents.El("h3", gomponents.Text("Failed to load component")),
		gomponents.El("p", gomponents.Text(err.Error())),
		gomponents.El("button",
			gomponents.Attr("onclick", "this.parentElement.dispatchEvent(new CustomEvent('retry'))"),
			gomponents.Text("Retry"),
		),
	)
}

// ------------------------------------
// 🔄 Lazy Loading Logic
// ------------------------------------

// load performs the actual component loading
func (l *LazyComponent) load() {
	l.mutex.Lock()
	if l.loading || l.loaded {
		l.mutex.Unlock()
		return
	}
	l.loading = true
	l.mutex.Unlock()

	// Perform loading with retry logic
	var component gomponents.Node
	var err error

	for attempt := 0; attempt <= l.maxRetries; attempt++ {
		component, err = l.loader()
		if err == nil {
			l.setLoaded(component)
			return
		}

		// If this isn't the last attempt, wait before retrying
		if attempt < l.maxRetries {
			l.retryCount++
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
		}
	}

	// All retries failed
	l.setError(fmt.Errorf("lazy component loading failed after %d attempts: %w", l.maxRetries+1, err))
}

// setLoaded marks the component as loaded
func (l *LazyComponent) setLoaded(component gomponents.Node) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.component = component
	l.loaded = true
	l.loading = false
	l.error = nil

	// Update registry stats
	getGlobalRegistry().incrementLoaded()

	// Call onLoad callback
	if l.onLoad != nil {
		l.onLoad(component)
	}
}

// setError marks the component as failed to load
func (l *LazyComponent) setError(err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.error = err
	l.loading = false

	// Update registry stats
	getGlobalRegistry().incrementError()

	// Call onError callback
	if l.onError != nil {
		l.onError(err)
	}
}

// ------------------------------------
// 🔍 Lazy Component State
// ------------------------------------

// IsLoaded returns whether the component is loaded
func (l *LazyComponent) IsLoaded() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.loaded
}

// IsLoading returns whether the component is currently loading
func (l *LazyComponent) IsLoading() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.loading
}

// HasError returns whether the component failed to load
func (l *LazyComponent) HasError() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.error != nil
}

// GetError returns the loading error if any
func (l *LazyComponent) GetError() error {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.error
}

// Retry retries loading the component
func (l *LazyComponent) Retry() {
	l.mutex.Lock()
	l.error = nil
	l.retryCount = 0
	l.mutex.Unlock()
	go l.load()
}

// Preload starts loading the component without rendering it
func (l *LazyComponent) Preload() {
	go l.load()
}

// ------------------------------------
// 🧹 Cleanup
// ------------------------------------

// cleanup cleans up the lazy component
func (l *LazyComponent) cleanup() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.component = nil
	l.loader = nil
	l.onLoad = nil
	l.onError = nil
}

// ------------------------------------
// 🗂️ Lazy Registry Implementation
// ------------------------------------

// getGlobalRegistry returns the global lazy registry
func getGlobalRegistry() *LazyRegistry {
	registryOnce.Do(func() {
		globalRegistry = &LazyRegistry{
			components: make(map[string]*LazyComponent),
			loaders:    make(map[string]func() (gomponents.Node, error)),
		}
	})
	return globalRegistry
}

// registerComponent registers a component with the registry
func (r *LazyRegistry) registerComponent(name string, component *LazyComponent) {
	if name == "" {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.components[name] = component
	r.stats.TotalComponents++
}

// GetComponent retrieves a component by name
func (r *LazyRegistry) GetComponent(name string) (*LazyComponent, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	component, exists := r.components[name]
	return component, exists
}

// RegisterLoader registers a loader function for a component
func (r *LazyRegistry) RegisterLoader(name string, loader func() (gomponents.Node, error)) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.loaders[name] = loader
}

// LoadComponent loads a component by name
func (r *LazyRegistry) LoadComponent(name string) (*LazyComponent, error) {
	r.mutex.RLock()
	loader, exists := r.loaders[name]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no loader registered for component: %s", name)
	}

	return Lazy(loader), nil
}

// incrementLoaded increments the loaded count
func (r *LazyRegistry) incrementLoaded() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.stats.LoadedCount++
}

// incrementError increments the error count
func (r *LazyRegistry) incrementError() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.stats.ErrorCount++
}

// GetStats returns current lazy loading statistics
func (r *LazyRegistry) GetStats() LazyStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.stats
}

// ------------------------------------
// 🔄 Utility Functions
// ------------------------------------

// PreloadComponent preloads a component by name
func PreloadComponent(name string) error {
	registry := getGlobalRegistry()
	component, exists := registry.GetComponent(name)
	if !exists {
		return fmt.Errorf("component not found: %s", name)
	}

	component.Preload()
	return nil
}

// PreloadAll preloads all registered components
func PreloadAll() {
	registry := getGlobalRegistry()
	registry.mutex.RLock()
	components := make([]*LazyComponent, 0, len(registry.components))
	for _, component := range registry.components {
		components = append(components, component)
	}
	registry.mutex.RUnlock()

	for _, component := range components {
		go component.Preload()
	}
}

// CreateLazyRoute creates a lazy-loaded route component
func CreateLazyRoute(loader func() (gomponents.Node, error)) func() gomponents.Node {
	lazy := Lazy(loader, LazyOptions{Cache: true})

	return func() gomponents.Node {
		return lazy.RenderWithFallback(LoadingSpinner())
	}
}

// ------------------------------------
// 🎯 Code Splitting Utilities
// ------------------------------------

// SplitComponent creates a code-split component with dynamic imports
func SplitComponent(
	importPath string,
	componentName string,
	options ...LazyOptions,
) *LazyComponent {
	loader := func() (gomponents.Node, error) {
		// In a real implementation, this would use dynamic imports
		// For now, we'll simulate it
		return gomponents.El("div",
			gomponents.Text(fmt.Sprintf("Component %s from %s", componentName, importPath)),
		), nil
	}

	return Lazy(loader, options...)
}

// ChunkLoader manages loading of code chunks
type ChunkLoader struct {
	chunks map[string]func() (gomponents.Node, error)
	mutex  sync.RWMutex
}

// NewChunkLoader creates a new chunk loader
func NewChunkLoader() *ChunkLoader {
	return &ChunkLoader{
		chunks: make(map[string]func() (gomponents.Node, error)),
	}
}

// RegisterChunk registers a chunk loader
func (cl *ChunkLoader) RegisterChunk(name string, loader func() (gomponents.Node, error)) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	cl.chunks[name] = loader
}

// LoadChunk loads a chunk by name
func (cl *ChunkLoader) LoadChunk(name string, options ...LazyOptions) *LazyComponent {
	cl.mutex.RLock()
	loader, exists := cl.chunks[name]
	cl.mutex.RUnlock()

	if !exists {
		loader = func() (gomponents.Node, error) {
			return nil, fmt.Errorf("chunk not found: %s", name)
		}
	}

	return Lazy(loader, options...)
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// MockLazyComponent creates a mock lazy component for testing
func MockLazyComponent(component gomponents.Node) *LazyComponent {
	return &LazyComponent{
		id:        atomic.AddUint64(&lazyIdCounter, 1),
		component: component,
		loaded:    true,
		loading:   false,
		cache:     true,
	}
}

// SimulateLoadingDelay simulates a loading delay for testing
func SimulateLoadingDelay(duration time.Duration, component gomponents.Node) func() (gomponents.Node, error) {
	return func() (gomponents.Node, error) {
		time.Sleep(duration)
		return component, nil
	}
}

// SimulateLoadingError simulates a loading error for testing
func SimulateLoadingError(message string) func() (gomponents.Node, error) {
	return func() (gomponents.Node, error) {
		return nil, fmt.Errorf("simulated error: %s", message)
	}
}
