# 🚀 Golid Lazy Loading & Suspense Guide

This guide covers the comprehensive lazy loading system implemented in Golid, including `createResource`, `Suspense` boundaries, and lazy component loading patterns.

## 📋 Table of Contents

- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Resource Management](#resource-management)
- [Suspense Boundaries](#suspense-boundaries)
- [Lazy Component Loading](#lazy-component-loading)
- [Advanced Caching](#advanced-caching)
- [Performance Optimization](#performance-optimization)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [API Reference](#api-reference)

## 🎯 Overview

The Golid lazy loading system provides:

- **Async Resource Loading** - `createResource` for data fetching with reactive state
- **Suspense Boundaries** - Declarative loading states with fallback rendering
- **Lazy Components** - Code-splitting and on-demand component loading
- **Advanced Caching** - Multi-policy caching with TTL and deduplication
- **Error Handling** - Comprehensive error boundaries and recovery mechanisms
- **Performance Optimization** - Preloading, batching, and intelligent caching

## 🧠 Core Concepts

### Reactive Resources

Resources are reactive primitives that manage async data loading:

```go
// Create a resource for fetching users
usersResource := golid.CreateResource(func() ([]User, error) {
    return fetchUsersFromAPI()
}, golid.ResourceOptions{
    Name:     "users",
    CacheKey: "users-list",
    TTL:      5 * time.Minute,
})

// Read the resource (will suspend if loading)
users := usersResource.Read()
```

### Suspense Boundaries

Suspense boundaries handle loading states declaratively:

```go
// Create a suspense boundary with fallback
suspense := golid.Suspense(
    golid.LoadingSpinner(), // Fallback content
    UserListComponent(),    // Children that might suspend
)
```

### Lazy Components

Lazy components enable code-splitting and on-demand loading:

```go
// Create a lazy component
lazyDashboard := golid.Lazy(func() (gomponents.Node, error) {
    return HeavyDashboardComponent(), nil
}, golid.LazyOptions{
    Cache: true,
    Preload: false,
})
```

## 📦 Resource Management

### Basic Resource Creation

```go
// Simple resource
dataResource := golid.CreateResource(func() (string, error) {
    return "Hello, World!", nil
})

// Resource with options
userResource := golid.CreateResource(fetchUser, golid.ResourceOptions{
    Name:       "user-profile",
    CacheKey:   "user-123",
    TTL:        10 * time.Minute,
    MaxRetries: 3,
    Timeout:    30 * time.Second,
    OnSuccess: func(data interface{}) {
        fmt.Println("User loaded successfully")
    },
    OnError: func(err error) {
        fmt.Printf("Failed to load user: %v\n", err)
    },
})
```

### Resource State Management

```go
// Check resource state
state := userResource.State()
if state.Loading {
    fmt.Println("Resource is loading...")
}
if state.Error != nil {
    fmt.Printf("Resource error: %v\n", state.Error)
}
if state.Data != nil {
    fmt.Printf("Resource data: %v\n", *state.Data)
}

// Manually refetch
userResource.RefetchAsync()

// Optimistic updates
userResource.Mutate(updatedUser)
```

### Reactive Dependencies

Resources can depend on other reactive values:

```go
pageSignal, setPage := golid.CreateSignal(1)

// Resource that refetches when page changes
itemsResource := golid.CreateResource(func() ([]Item, error) {
    page := pageSignal()
    return fetchItems(page)
}, golid.ResourceOptions{
    Dependencies: []interface{}{pageSignal},
})

// Changing page will trigger refetch
setPage(2)
```

## 🔄 Suspense Boundaries

### Basic Suspense

```go
// Simple suspense boundary
suspense := golid.Suspense(
    gomponents.Text("Loading..."), // Fallback
    DataComponent(),               // Content that might suspend
)
```

### Suspense with Options

```go
suspense := golid.CreateSuspense(
    golid.LoadingSpinner(),
    golid.SuspenseOptions{
        Name: "user-data-boundary",
        OnSuspend: func() {
            fmt.Println("Started loading...")
        },
        OnResume: func() {
            fmt.Println("Loading completed!")
        },
    },
    UserProfileComponent(),
    UserPostsComponent(),
)
```

### Nested Suspense

```go
// Parent suspense for main content
parentSuspense := golid.Suspense(
    golid.LoadingSkeleton(5),
    
    // Child suspense for specific sections
    golid.NestedSuspense(
        gomponents.Text("Loading comments..."),
        CommentsComponent(),
    ),
)
```

### Error Boundaries with Suspense

```go
errorBoundary := golid.CreateErrorBoundary(func(err error) interface{} {
    return golid.ErrorFallback(err.Error())
})

suspense := golid.CreateSuspense(
    golid.LoadingSpinner(),
    golid.SuspenseOptions{
        ErrorBoundary: errorBoundary,
    },
    RiskyComponent(),
)
```

## 🔄 Lazy Component Loading

### Basic Lazy Loading

```go
// Create lazy component
lazyComponent := golid.Lazy(func() (gomponents.Node, error) {
    // Simulate heavy loading
    time.Sleep(2 * time.Second)
    return HeavyComponent(), nil
})

// Render with fallback
content := lazyComponent.RenderWithFallback(
    gomponents.Text("Loading heavy component..."),
)
```

### Lazy with Options

```go
lazyDashboard := golid.Lazy(loadDashboard, golid.LazyOptions{
    Name:       "dashboard",
    Cache:      true,
    MaxRetries: 3,
    Preload:    false,
    Timeout:    30 * time.Second,
    OnLoad: func(component gomponents.Node) {
        fmt.Println("Dashboard loaded!")
    },
    OnError: func(err error) {
        fmt.Printf("Dashboard failed: %v\n", err)
    },
})
```

### Lazy with Suspense

```go
// Automatic suspense integration
lazyComp, suspense := golid.LazyWithSuspense(
    loadHeavyComponent,
    golid.LoadingSpinner(),
    golid.LazyOptions{Cache: true},
)
```

### Code Splitting

```go
// Split component by route
dashboardRoute := golid.CreateLazyRoute(func() (gomponents.Node, error) {
    return DashboardComponent(), nil
})

// Chunk-based loading
chunkLoader := golid.NewChunkLoader()
chunkLoader.RegisterChunk("dashboard", loadDashboard)
chunkLoader.RegisterChunk("profile", loadProfile)

// Load specific chunk
dashboardChunk := chunkLoader.LoadChunk("dashboard", golid.LazyOptions{
    Cache: true,
})
```

## 🗄️ Advanced Caching

### Basic Resource Cache

```go
// Create cache
cache := golid.NewResourceCache(1000, 10*time.Minute)

// Set and get
cache.Set("user-123", userData, 5*time.Minute)
if data, found := cache.Get("user-123"); found {
    fmt.Printf("Cached data: %v\n", data)
}

// Cache statistics
stats := cache.Stats()
fmt.Printf("Cache hits: %d, misses: %d\n", stats.Hits, stats.Misses)
```

### Advanced Cache with Policies

```go
// Create advanced cache with LRU policy
cache := golid.NewAdvancedResourceCache(golid.CacheConfig{
    Policy:      golid.CachePolicyLRU,
    MaxSize:     1000,
    MaxMemory:   100 * 1024 * 1024, // 100MB
    DefaultTTL:  10 * time.Minute,
    OnHit: func(key string, value interface{}) {
        fmt.Printf("Cache hit: %s\n", key)
    },
    OnEvict: func(key string, value interface{}) {
        fmt.Printf("Cache evicted: %s\n", key)
    },
})

// Set with options
cache.Set("user-data", userData, golid.CacheSetOptions{
    TTL:  5 * time.Minute,
    Tags: []string{"user", "profile"},
    Metadata: map[string]interface{}{
        "source": "api",
        "version": 1,
    },
})
```

### Tag-based Cache Operations

```go
// Set items with tags
cache.Set("user1", user1Data, golid.CacheSetOptions{
    Tags: []string{"user", "active"},
})
cache.Set("user2", user2Data, golid.CacheSetOptions{
    Tags: []string{"user", "inactive"},
})

// Get by tag
activeUsers := cache.GetByTag("user")
fmt.Printf("Found %d users\n", len(activeUsers))

// Invalidate by tag
invalidated := cache.InvalidateByTag("user")
fmt.Printf("Invalidated %d entries\n", invalidated)
```

## ⚡ Performance Optimization

### Preloading

```go
// Preload critical resources
criticalResource := golid.CreateResource(fetchCriticalData, golid.ResourceOptions{
    Preload: true, // Start loading immediately
})

// Preload lazy components
golid.PreloadComponent("dashboard")

// Preload all registered components
golid.PreloadAll()
```

### Resource Deduplication

```go
// Multiple components requesting same resource
// Only one network request will be made
userResource1 := golid.CreateResource(fetchUser, golid.ResourceOptions{
    CacheKey: "user-123",
})
userResource2 := golid.CreateResource(fetchUser, golid.ResourceOptions{
    CacheKey: "user-123", // Same cache key = deduplication
})
```

### Intelligent Caching

```go
// Cache with different policies
lruCache := golid.NewAdvancedResourceCache(golid.CacheConfig{
    Policy: golid.CachePolicyLRU, // Least Recently Used
})

lfuCache := golid.NewAdvancedResourceCache(golid.CacheConfig{
    Policy: golid.CachePolicyLFU, // Least Frequently Used
})

ttlCache := golid.NewAdvancedResourceCache(golid.CacheConfig{
    Policy: golid.CachePolicyTTL, // Time To Live
})
```

## 🎯 Best Practices

### 1. Resource Organization

```go
// Group related resources
type UserResources struct {
    Profile *golid.AsyncResource[User]
    Posts   *golid.AsyncResource[[]Post]
    Friends *golid.AsyncResource[[]User]
}

func NewUserResources(userID int) *UserResources {
    return &UserResources{
        Profile: golid.CreateResource(func() (User, error) {
            return fetchUserProfile(userID)
        }, golid.ResourceOptions{
            CacheKey: fmt.Sprintf("user-profile-%d", userID),
            TTL:      5 * time.Minute,
        }),
        Posts: golid.CreateResource(func() ([]Post, error) {
            return fetchUserPosts(userID)
        }, golid.ResourceOptions{
            CacheKey: fmt.Sprintf("user-posts-%d", userID),
            TTL:      2 * time.Minute,
        }),
    }
}
```

### 2. Error Handling Strategy

```go
// Comprehensive error handling
resource := golid.CreateResource(fetchData, golid.ResourceOptions{
    MaxRetries: 3,
    OnError: func(err error) {
        // Log error
        log.Printf("Resource error: %v", err)
        
        // Send to error tracking service
        errorTracker.Report(err)
    },
})

// Error boundary with fallback
errorBoundary := golid.CreateErrorBoundary(func(err error) interface{} {
    if isNetworkError(err) {
        return NetworkErrorComponent()
    }
    return GenericErrorComponent(err)
})
```

### 3. Cache Strategy

```go
// Different TTL for different data types
userCache := golid.ResourceOptions{
    TTL: 10 * time.Minute, // User data changes infrequently
}

postsCache := golid.ResourceOptions{
    TTL: 2 * time.Minute, // Posts change more frequently
}

realTimeCache := golid.ResourceOptions{
    TTL: 30 * time.Second, // Real-time data
}
```

### 4. Lazy Loading Strategy

```go
// Critical components - load immediately
criticalComponent := golid.Lazy(loadCritical, golid.LazyOptions{
    Preload: true,
    Cache:   true,
})

// Secondary components - load on demand
secondaryComponent := golid.Lazy(loadSecondary, golid.LazyOptions{
    Preload: false,
    Cache:   true,
})

// Heavy components - load with user interaction
heavyComponent := golid.Lazy(loadHeavy, golid.LazyOptions{
    Preload: false,
    Cache:   true,
    Timeout: 60 * time.Second,
})
```

## 📚 Examples

### Complete User Dashboard

```go
func UserDashboard(userID int) gomponents.Node {
    // Create resources
    userResource := golid.CreateResource(func() (User, error) {
        return fetchUser(userID)
    }, golid.ResourceOptions{
        CacheKey: fmt.Sprintf("user-%d", userID),
        TTL:      5 * time.Minute,
    })

    postsResource := golid.CreateResource(func() ([]Post, error) {
        return fetchUserPosts(userID)
    }, golid.ResourceOptions{
        CacheKey: fmt.Sprintf("posts-%d", userID),
        TTL:      2 * time.Minute,
        Dependencies: []interface{}{userResource}, // Depend on user
    })

    // Lazy load heavy analytics
    analyticsComponent := golid.Lazy(func() (gomponents.Node, error) {
        return AnalyticsComponent(userID), nil
    }, golid.LazyOptions{
        Cache: true,
        OnLoad: func(component gomponents.Node) {
            fmt.Println("Analytics loaded")
        },
    })

    return golid.Suspense(
        golid.LoadingSkeleton(3),
        
        gomponents.Div(
            gomponents.Class("user-dashboard"),
            
            // User profile section
            golid.Suspense(
                gomponents.Text("Loading profile..."),
                UserProfileComponent(userResource),
            ).Render(),
            
            // Posts section
            golid.Suspense(
                gomponents.Text("Loading posts..."),
                UserPostsComponent(postsResource),
            ).Render(),
            
            // Lazy analytics
            analyticsComponent.RenderWithFallback(
                gomponents.Text("Loading analytics..."),
            ),
        ),
    ).Render()
}
```

### Infinite Scroll with Resources

```go
func InfiniteScrollList() gomponents.Node {
    pageSignal, setPage := golid.CreateSignal(1)
    itemsSignal, setItems := golid.CreateSignal([]Item{})

    // Resource that loads more items
    moreItemsResource := golid.CreateResource(func() ([]Item, error) {
        page := pageSignal()
        newItems, err := fetchItems(page)
        if err != nil {
            return nil, err
        }
        
        // Append to existing items
        currentItems := itemsSignal()
        allItems := append(currentItems, newItems...)
        setItems(allItems)
        
        return newItems, nil
    }, golid.ResourceOptions{
        Dependencies: []interface{}{pageSignal},
    })

    return gomponents.Div(
        gomponents.Class("infinite-scroll"),
        
        // Render current items
        ItemListComponent(itemsSignal),
        
        // Load more button with suspense
        golid.Suspense(
            gomponents.Text("Loading more..."),
            gomponents.Button(
                gomponents.Text("Load More"),
                // OnClick: () => setPage(pageSignal() + 1)
            ),
        ).Render(),
    )
}
```

## 📖 API Reference

### Resource API

```go
// Create resource
func CreateResource[T any](fetcher func() (T, error), options ...ResourceOptions) *AsyncResource[T]

// Resource methods
func (r *AsyncResource[T]) Read() T                    // Read data (suspends if loading)
func (r *AsyncResource[T]) State() ResourceState[T]   // Get current state
func (r *AsyncResource[T]) RefetchAsync()             // Force refetch
func (r *AsyncResource[T]) Mutate(value T)            // Optimistic update

// Resource options
type ResourceOptions struct {
    Name         string
    CacheKey     string
    TTL          time.Duration
    MaxRetries   int
    Timeout      time.Duration
    OnSuccess    func(interface{})
    OnError      func(error)
    Dependencies []interface{}
    Preload      bool
}
```

### Suspense API

```go
// Create suspense boundary
func Suspense(fallback gomponents.Node, children ...gomponents.Node) *SuspenseBoundary
func CreateSuspense(fallback gomponents.Node, options SuspenseOptions, children ...gomponents.Node) *SuspenseBoundary

// Suspense methods
func (s *SuspenseBoundary) Render() gomponents.Node
func (s *SuspenseBoundary) IsSuspended() bool
func (s *SuspenseBoundary) SetFallback(fallback gomponents.Node)

// Suspense options
type SuspenseOptions struct {
    Name          string
    OnSuspend     func()
    OnResume      func()
    ErrorBoundary *ErrorBoundary
    Owner         *Owner
}
```

### Lazy Loading API

```go
// Create lazy component
func Lazy(loader func() (gomponents.Node, error), options ...LazyOptions) *LazyComponent

// Lazy methods
func (l *LazyComponent) Render() gomponents.Node
func (l *LazyComponent) RenderWithFallback(fallback gomponents.Node) gomponents.Node
func (l *LazyComponent) IsLoaded() bool
func (l *LazyComponent) IsLoading() bool
func (l *LazyComponent) Preload()
func (l *LazyComponent) Retry()

// Lazy options
type LazyOptions struct {
    Name       string
    Cache      bool
    MaxRetries int
    OnLoad     func(gomponents.Node)
    OnError    func(error)
    Preload    bool
    Timeout    time.Duration
}
```

### Cache API

```go
// Basic cache
func NewResourceCache(maxSize int, defaultTTL time.Duration) *ResourceCache
func (c *ResourceCache) Get(key string) (interface{}, bool)
func (c *ResourceCache) Set(key string, value interface{}, ttl time.Duration)
func (c *ResourceCache) Stats() CacheStats

// Advanced cache
func NewAdvancedResourceCache(config CacheConfig) *AdvancedResourceCache
func (c *AdvancedResourceCache) GetByTag(tag string) map[string]interface{}
func (c *AdvancedResourceCache) InvalidateByTag(tag string) int
func (c *AdvancedResourceCache) GetAdvancedStats() AdvancedCacheStats
```

## 🎉 Conclusion

The Golid lazy loading system provides a comprehensive solution for:

- **Async Data Management** - Reactive resources with automatic state management
- **Loading States** - Declarative Suspense boundaries with fallback rendering
- **Code Splitting** - Lazy component loading for performance optimization
- **Intelligent Caching** - Multi-policy caching with deduplication and TTL
- **Error Handling** - Robust error boundaries and recovery mechanisms

This system enables building highly performant, responsive applications with excellent user experience through optimized loading patterns and efficient resource management.

For more examples and advanced usage patterns, see the `examples/lazy_loading_demo/` directory.