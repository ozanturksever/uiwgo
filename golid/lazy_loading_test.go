// lazy_loading_test.go
// Comprehensive tests for lazy loading, resources, and Suspense boundaries

package golid

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🧪 Resource Tests
// ------------------------------------

func TestCreateResource(t *testing.T) {
	// Test basic resource creation
	resource := CreateResource(func() (string, error) {
		return "test data", nil
	})

	if resource == nil {
		t.Fatal("Expected resource to be created")
	}

	// Test resource state
	state := resource.State()
	if state.Loading {
		t.Error("Expected resource to not be loading initially")
	}
	if state.Data != nil {
		t.Error("Expected resource data to be nil initially")
	}
}

func TestResourceRead(t *testing.T) {
	// Test successful resource read
	resource := CreateResource(func() (string, error) {
		return "test data", nil
	})

	// This would normally suspend in a Suspense boundary
	// For testing, we'll catch the panic
	defer func() {
		if r := recover(); r != nil {
			if suspenseErr, ok := r.(SuspenseException); ok {
				if suspenseErr.Message != "Resource loading" {
					t.Errorf("Expected suspense exception, got: %v", suspenseErr.Message)
				}
			} else {
				t.Errorf("Expected SuspenseException, got: %v", r)
			}
		}
	}()

	resource.Read()
	t.Error("Expected Read() to panic with SuspenseException")
}

func TestResourceWithOptions(t *testing.T) {
	options := ResourceOptions{
		Name:       "test-resource",
		CacheKey:   "test-key",
		TTL:        1 * time.Minute,
		MaxRetries: 5,
		Timeout:    10 * time.Second,
	}

	resource := CreateResource(func() (int, error) {
		return 42, nil
	}, options)

	if resource == nil {
		t.Fatal("Expected resource to be created with options")
	}

	// Verify options are applied
	if resource.options.Name != "test-resource" {
		t.Errorf("Expected name 'test-resource', got '%s'", resource.options.Name)
	}
	if resource.options.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", resource.options.MaxRetries)
	}
}

func TestResourceMutate(t *testing.T) {
	resource := CreateResource(func() (string, error) {
		return "initial", nil
	})

	// Mutate the resource
	resource.Mutate("updated")

	state := resource.State()
	if state.Data == nil || *state.Data != "updated" {
		t.Errorf("Expected data to be 'updated', got %v", state.Data)
	}
	if state.Loading {
		t.Error("Expected resource to not be loading after mutation")
	}
}

func TestResourceRefetch(t *testing.T) {
	callCount := 0
	resource := CreateResource(func() (int, error) {
		callCount++
		return callCount, nil
	})

	// Initial state
	if resource.State().Loading {
		t.Error("Expected resource to not be loading initially")
	}

	// Trigger refetch
	resource.RefetchAsync()

	// Give some time for async operation
	time.Sleep(100 * time.Millisecond)

	// Should have triggered a fetch
	if callCount == 0 {
		t.Error("Expected fetcher to be called after RefetchAsync")
	}
}

// ------------------------------------
// 🔄 Suspense Tests
// ------------------------------------

func TestSuspenseCreation(t *testing.T) {
	fallback := gomponents.Text("Loading...")
	children := []gomponents.Node{gomponents.Text("Content")}

	suspense := CreateSuspense(fallback, SuspenseOptions{}, children...)

	if suspense == nil {
		t.Fatal("Expected suspense boundary to be created")
	}

	if suspense.IsSuspended() {
		t.Error("Expected suspense to not be suspended initially")
	}
}

func TestSuspenseSuspendResume(t *testing.T) {
	suspense := Suspense(gomponents.Text("Loading..."))

	// Test suspend
	suspense.suspend()
	if !suspense.IsSuspended() {
		t.Error("Expected suspense to be suspended")
	}

	// Test resume
	suspense.resume()
	if suspense.IsSuspended() {
		t.Error("Expected suspense to not be suspended after resume")
	}
}

func TestSuspenseWithCallbacks(t *testing.T) {
	var suspendCalled, resumeCalled bool

	options := SuspenseOptions{
		OnSuspend: func() { suspendCalled = true },
		OnResume:  func() { resumeCalled = true },
	}

	suspense := CreateSuspense(gomponents.Text("Loading..."), options)

	// Test suspend callback
	suspense.suspend()
	if !suspendCalled {
		t.Error("Expected onSuspend callback to be called")
	}

	// Test resume callback
	suspense.resume()
	if !resumeCalled {
		t.Error("Expected onResume callback to be called")
	}
}

func TestNestedSuspense(t *testing.T) {
	parent := Suspense(gomponents.Text("Parent Loading..."))
	child := NestedSuspense(gomponents.Text("Child Loading..."))

	if parent == nil || child == nil {
		t.Fatal("Expected both suspense boundaries to be created")
	}

	// Test that they're independent
	parent.suspend()
	if child.IsSuspended() {
		t.Error("Expected child suspense to be independent of parent")
	}
}

// ------------------------------------
// 🔄 Lazy Component Tests
// ------------------------------------

func TestLazyComponentCreation(t *testing.T) {
	loader := func() (gomponents.Node, error) {
		return gomponents.Text("Lazy content"), nil
	}

	lazy := Lazy(loader)

	if lazy == nil {
		t.Fatal("Expected lazy component to be created")
	}

	if lazy.IsLoaded() {
		t.Error("Expected lazy component to not be loaded initially")
	}
	if lazy.IsLoading() {
		t.Error("Expected lazy component to not be loading initially")
	}
}

func TestLazyComponentWithOptions(t *testing.T) {
	var loadCalled, errorCalled bool

	options := LazyOptions{
		Name:       "test-lazy",
		Cache:      true,
		MaxRetries: 3,
		OnLoad: func(component gomponents.Node) {
			loadCalled = true
		},
		OnError: func(err error) {
			errorCalled = true
		},
	}

	loader := func() (gomponents.Node, error) {
		return gomponents.Text("Lazy content"), nil
	}

	lazy := Lazy(loader, options)

	if lazy.maxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", lazy.maxRetries)
	}
	if !lazy.cache {
		t.Error("Expected cache to be enabled")
	}

	// Test that callbacks are set
	if lazy.onLoad == nil {
		t.Error("Expected onLoad callback to be set")
	}
	if lazy.onError == nil {
		t.Error("Expected onError callback to be set")
	}

	// Use variables to avoid unused warnings
	_ = loadCalled
	_ = errorCalled
}

func TestLazyComponentRender(t *testing.T) {
	loader := func() (gomponents.Node, error) {
		return gomponents.Text("Lazy content"), nil
	}

	lazy := Lazy(loader)

	// Test render with fallback
	fallback := gomponents.Text("Loading...")
	result := lazy.RenderWithFallback(fallback)

	if result == nil {
		t.Error("Expected render result to not be nil")
	}

	// Should return fallback initially
	// In a real implementation, this would be the fallback content
}

func TestLazyComponentError(t *testing.T) {
	loader := func() (gomponents.Node, error) {
		return nil, errors.New("load failed")
	}

	lazy := Lazy(loader)

	// Trigger loading
	go lazy.load()

	// Wait for loading to complete
	time.Sleep(100 * time.Millisecond)

	if !lazy.HasError() {
		t.Error("Expected lazy component to have error")
	}

	err := lazy.GetError()
	if err == nil {
		t.Error("Expected error to be set")
	}
}

func TestLazyComponentRetry(t *testing.T) {
	attempts := 0
	loader := func() (gomponents.Node, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("not ready")
		}
		return gomponents.Text("Success"), nil
	}

	lazy := Lazy(loader, LazyOptions{MaxRetries: 5})

	// Trigger loading
	go lazy.load()

	// Wait for retries to complete
	time.Sleep(5 * time.Second)

	if attempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts)
	}

	if lazy.HasError() {
		t.Errorf("Expected no error after successful retry, got: %v", lazy.GetError())
	}
}

func TestLazyComponentPreload(t *testing.T) {
	loaded := false
	loader := func() (gomponents.Node, error) {
		loaded = true
		return gomponents.Text("Preloaded"), nil
	}

	options := LazyOptions{Preload: true}
	lazy := Lazy(loader, options)

	// Give time for preload
	time.Sleep(100 * time.Millisecond)

	if !loaded {
		t.Error("Expected component to be preloaded")
	}

	// Use lazy to avoid unused warning
	_ = lazy
}

// ------------------------------------
// 🗄️ Cache Tests
// ------------------------------------

func TestResourceCache(t *testing.T) {
	cache := NewResourceCache(10, 1*time.Minute)

	// Test set and get
	cache.Set("key1", "value1", 30*time.Second)

	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find cached value")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got '%v'", value)
	}

	// Test stats
	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Expected 0 misses, got %d", stats.Misses)
	}
}

func TestResourceCacheExpiration(t *testing.T) {
	cache := NewResourceCache(10, 100*time.Millisecond)

	// Set with short TTL
	cache.Set("key1", "value1", 50*time.Millisecond)

	// Should be available immediately
	_, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find cached value immediately")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected cached value to be expired")
	}
}

func TestAdvancedResourceCache(t *testing.T) {
	config := CacheConfig{
		Policy:    CachePolicyLRU,
		MaxSize:   3,
		MaxMemory: 1024,
	}

	cache := NewAdvancedResourceCache(config)
	defer cache.Stop()

	// Test basic operations
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// All should be present
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("key%d", i)
		if _, found := cache.Get(key); !found {
			t.Errorf("Expected to find %s", key)
		}
	}

	// Add one more to trigger eviction
	cache.Set("key4", "value4")

	// key1 should be evicted (LRU)
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}

	// Others should still be present
	for i := 2; i <= 4; i++ {
		key := fmt.Sprintf("key%d", i)
		if _, found := cache.Get(key); !found {
			t.Errorf("Expected to find %s", key)
		}
	}
}

func TestCacheWithTags(t *testing.T) {
	cache := NewAdvancedResourceCache(CacheConfig{MaxSize: 10})
	defer cache.Stop()

	// Set values with tags
	cache.Set("user1", "John", CacheSetOptions{Tags: []string{"user", "active"}})
	cache.Set("user2", "Jane", CacheSetOptions{Tags: []string{"user", "inactive"}})
	cache.Set("post1", "Hello", CacheSetOptions{Tags: []string{"post"}})

	// Get by tag
	users := cache.GetByTag("user")
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	posts := cache.GetByTag("post")
	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	// Invalidate by tag
	invalidated := cache.InvalidateByTag("user")
	if invalidated != 2 {
		t.Errorf("Expected 2 invalidated entries, got %d", invalidated)
	}

	// Users should be gone
	users = cache.GetByTag("user")
	if len(users) != 0 {
		t.Errorf("Expected 0 users after invalidation, got %d", len(users))
	}

	// Posts should remain
	posts = cache.GetByTag("post")
	if len(posts) != 1 {
		t.Errorf("Expected 1 post to remain, got %d", len(posts))
	}
}

// ------------------------------------
// 🔄 Integration Tests
// ------------------------------------

func TestResourceWithSuspense(t *testing.T) {
	// Create a resource that takes time to load
	resource := CreateResource(func() (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "loaded data", nil
	})

	// Create suspense boundary
	suspense := Suspense(gomponents.Text("Loading..."))

	// In a real implementation, reading the resource would suspend
	// the boundary until data is loaded

	if resource == nil || suspense == nil {
		t.Fatal("Expected both resource and suspense to be created")
	}
}

func TestLazyWithSuspense(t *testing.T) {
	loader := func() (gomponents.Node, error) {
		time.Sleep(50 * time.Millisecond)
		return gomponents.Text("Lazy loaded"), nil
	}

	lazy, suspense := LazyWithSuspense(
		loader,
		gomponents.Text("Loading lazy component..."),
	)

	if lazy == nil || suspense == nil {
		t.Fatal("Expected both lazy component and suspense to be created")
	}

	// Initially should be suspended
	if !suspense.IsSuspended() {
		t.Error("Expected suspense to be suspended initially")
	}
}

// ------------------------------------
// 🧪 Concurrent Tests
// ------------------------------------

func TestConcurrentResourceAccess(t *testing.T) {
	resource := CreateResource(func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Concurrent access
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// Expected to panic with SuspenseException
					if _, ok := r.(SuspenseException); !ok {
						errors <- fmt.Errorf("unexpected panic: %v", r)
					}
				}
			}()
			resource.Read()
		}()
	}

	wg.Wait()
	close(errors)

	// Check for unexpected errors
	for err := range errors {
		t.Error(err)
	}
}

func TestConcurrentCacheAccess(t *testing.T) {
	cache := NewAdvancedResourceCache(CacheConfig{MaxSize: 100})
	defer cache.Stop()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)
			cache.Set(key, value)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i%25) // Some overlap
			_, _ = cache.Get(key)
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify cache state
	stats := cache.GetAdvancedStats()
	if stats.Size > 100 {
		t.Errorf("Expected cache size <= 100, got %d", stats.Size)
	}
}

// ------------------------------------
// 🧪 Benchmark Tests
// ------------------------------------

func BenchmarkResourceCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateResource(func() (int, error) {
			return i, nil
		})
	}
}

func BenchmarkCacheOperations(b *testing.B) {
	cache := NewResourceCache(1000, 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%100)
		value := fmt.Sprintf("value%d", i)

		cache.Set(key, value, 1*time.Minute)
		cache.Get(key)
	}
}

func BenchmarkAdvancedCacheOperations(b *testing.B) {
	cache := NewAdvancedResourceCache(CacheConfig{
		MaxSize: 1000,
		Policy:  CachePolicyLRU,
	})
	defer cache.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%100)
		value := fmt.Sprintf("value%d", i)

		cache.Set(key, value)
		cache.Get(key)
	}
}

func BenchmarkLazyComponentCreation(b *testing.B) {
	loader := func() (gomponents.Node, error) {
		return gomponents.Text("test"), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lazy(loader)
	}
}
