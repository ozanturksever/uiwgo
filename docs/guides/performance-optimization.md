# Performance Optimization Guide for UIwGo

This guide provides comprehensive strategies for optimizing the performance of UIwGo applications, focusing on efficient use of helper functions and reactive patterns.

## Related Documentation

- **[Helper Functions Guide](./helper-functions.md)** - Comprehensive guide to all helper functions
- **[Quick Reference](./quick-reference.md)** - Concise syntax reference
- **[Real-World Examples](./real-world-examples.md)** - Practical application examples
- **[Integration Examples](./integration-examples.md)** - Complex multi-helper scenarios
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions

## Table of Contents

- [Performance Fundamentals](#performance-fundamentals)
- [Reactivity Optimization](#reactivity-optimization)
- [List Rendering Optimization](#list-rendering-optimization)
- [Memory Management](#memory-management)
- [Bundle Size Optimization](#bundle-size-optimization)
- [Runtime Performance](#runtime-performance)
- [Profiling and Monitoring](#profiling-and-monitoring)
- [Best Practices](#best-practices)

## Performance Fundamentals

### Understanding UIwGo's Reactive System

UIwGo's performance is built on efficient reactivity. Understanding how signals, memos, and effects work is crucial for optimization:

```go
// Signals: Store reactive state
count := reactivity.CreateSignal(0)

// Memos: Cache expensive computations
expensiveResult := reactivity.CreateMemo(func() int {
    // Only recalculates when dependencies change
    return expensiveCalculation(count.Get())
})

// Effects: Handle side effects
reactivity.CreateEffect(func() {
    // Runs when dependencies change
    logutil.Logf("Count changed to: %d", count.Get())
})
```

### Performance Metrics to Track

1. **Render Time**: Time taken to render components
2. **Memory Usage**: Heap size and garbage collection frequency
3. **Bundle Size**: JavaScript/WASM bundle size
4. **Reactivity Overhead**: Time spent in reactive updates
5. **DOM Operations**: Number and frequency of DOM manipulations

## Reactivity Optimization

### Use Memos for Expensive Computations

```go
// ❌ Bad: Expensive computation on every render
func renderStats() g.Node {
    total := 0
    for _, item := range items.Get() {
        total += complexCalculation(item) // Runs every render!
    }
    return g.Div(g.Text(fmt.Sprintf("Total: %d", total)))
}

// ✅ Good: Memoized expensive computation
totalMemo := reactivity.CreateMemo(func() int {
    total := 0
    for _, item := range items.Get() {
        total += complexCalculation(item) // Only runs when items change
    }
    return total
})

func renderStats() g.Node {
    return g.Div(g.Text(fmt.Sprintf("Total: %d", totalMemo.Get())))
}
```

### Optimize Signal Dependencies

```go
// ❌ Bad: Multiple signal accesses
func renderUserInfo() g.Node {
    user := currentUser.Get()
    settings := userSettings.Get()
    preferences := userPreferences.Get()
    
    return g.Div(
        g.H1(g.Text(user.Name)),
        g.P(g.Text(settings.Theme)),
        g.P(g.Text(preferences.Language)),
    )
}

// ✅ Good: Combined memo for related data
userData := reactivity.CreateMemo(func() UserData {
    return UserData{
        User:        currentUser.Get(),
        Settings:    userSettings.Get(),
        Preferences: userPreferences.Get(),
    }
})

func renderUserInfo() g.Node {
    data := userData.Get()
    return g.Div(
        g.H1(g.Text(data.User.Name)),
        g.P(g.Text(data.Settings.Theme)),
        g.P(g.Text(data.Preferences.Language)),
    )
}
```

### Batch Signal Updates

```go
// ❌ Bad: Multiple individual updates
func updateUserProfile(name, email, phone string) {
    userName.Set(name)     // Triggers update
    userEmail.Set(email)   // Triggers update
    userPhone.Set(phone)   // Triggers update
}

// ✅ Good: Batch updates using a single signal
type UserProfile struct {
    Name  string
    Email string
    Phone string
}

func updateUserProfile(name, email, phone string) {
    userProfile.Set(UserProfile{
        Name:  name,
        Email: email,
        Phone: phone,
    }) // Single update triggers one re-render
}
```

### Debounce Frequent Updates

```go
// ❌ Bad: Immediate updates on every keystroke
searchInput := reactivity.CreateSignal("")
dom.OnInput(func(value string) {
    searchInput.Set(value) // Triggers search on every character
})

// ✅ Good: Debounced updates
type DebouncedSignal[T any] struct {
    signal reactivity.Signal[T]
    timer  *time.Timer
}

func NewDebouncedSignal[T any](initial T, delay time.Duration) *DebouncedSignal[T] {
    return &DebouncedSignal[T]{
        signal: reactivity.CreateSignal(initial),
    }
}

func (d *DebouncedSignal[T]) Set(value T, delay time.Duration) {
    if d.timer != nil {
        d.timer.Stop()
    }
    d.timer = time.AfterFunc(delay, func() {
        d.signal.Set(value)
    })
}

func (d *DebouncedSignal[T]) Get() T {
    return d.signal.Get()
}

// Usage
debouncedSearch := NewDebouncedSignal("", 300*time.Millisecond)
dom.OnInput(func(value string) {
    debouncedSearch.Set(value, 300*time.Millisecond)
})
```

## List Rendering Optimization

### Efficient Key Functions

```go
// ❌ Bad: Using array index as key
comps.For(comps.ForProps[Item]{
    Items: items,
    Key: func(item Item) string {
        return fmt.Sprintf("%d", index) // Wrong! Index changes when list reorders
    },
    Children: func(item Item, index int) g.Node {
        return renderItem(item)
    },
})

// ✅ Good: Using stable, unique identifier
comps.For(comps.ForProps[Item]{
    Items: items,
    Key: func(item Item) string {
        return item.ID // Stable identifier
    },
    Children: func(item Item, index int) g.Node {
        return renderItem(item)
    },
})
```

### Virtual Scrolling for Large Lists

```go
type VirtualList struct {
    items       reactivity.Signal[[]Item]
    itemHeight  int
    visibleCount int
    scrollTop   reactivity.Signal[int]
}

func NewVirtualList(items []Item, itemHeight, visibleCount int) *VirtualList {
    return &VirtualList{
        items:        reactivity.CreateSignal(items),
        itemHeight:   itemHeight,
        visibleCount: visibleCount,
        scrollTop:    reactivity.CreateSignal(0),
    }
}

func (vl *VirtualList) render() g.Node {
    visibleItems := reactivity.CreateMemo(func() []ItemWithIndex {
        allItems := vl.items.Get()
        startIndex := vl.scrollTop.Get() / vl.itemHeight
        endIndex := startIndex + vl.visibleCount
        
        if startIndex < 0 {
            startIndex = 0
        }
        if endIndex > len(allItems) {
            endIndex = len(allItems)
        }
        
        var visible []ItemWithIndex
        for i := startIndex; i < endIndex; i++ {
            visible = append(visible, ItemWithIndex{
                Item:  allItems[i],
                Index: i,
            })
        }
        return visible
    })
    
    totalHeight := len(vl.items.Get()) * vl.itemHeight
    offsetY := (vl.scrollTop.Get() / vl.itemHeight) * vl.itemHeight
    
    return g.Div(
        g.Class("virtual-list"),
        g.Style(fmt.Sprintf("height: %dpx; overflow-y: auto;", vl.visibleCount*vl.itemHeight)),
        dom.OnScroll(func(e dom.Event) {
            scrollTop := e.Target().Get("scrollTop").Int()
            vl.scrollTop.Set(scrollTop)
        }),
        
        // Spacer for total height
        g.Div(
            g.Style(fmt.Sprintf("height: %dpx; position: relative;", totalHeight)),
            
            // Visible items container
            g.Div(
                g.Style(fmt.Sprintf("transform: translateY(%dpx);", offsetY)),
                comps.For(comps.ForProps[ItemWithIndex]{
                    Items: visibleItems,
                    Key: func(item ItemWithIndex) string {
                        return fmt.Sprintf("%d", item.Index)
                    },
                    Children: func(item ItemWithIndex, index int) g.Node {
                        return g.Div(
                            g.Style(fmt.Sprintf("height: %dpx;", vl.itemHeight)),
                            renderItem(item.Item),
                        )
                    },
                }),
            ),
        ),
    )
}
```

### Lazy Loading for Complex Items

```go
type LazyItem struct {
    id      string
    loaded  reactivity.Signal[bool]
    data    reactivity.Signal[*ItemData]
    loading reactivity.Signal[bool]
}

func NewLazyItem(id string) *LazyItem {
    return &LazyItem{
        id:      id,
        loaded:  reactivity.CreateSignal(false),
        data:    reactivity.CreateSignal[*ItemData](nil),
        loading: reactivity.CreateSignal(false),
    }
}

func (li *LazyItem) load() {
    if li.loaded.Get() || li.loading.Get() {
        return
    }
    
    li.loading.Set(true)
    go func() {
        data := fetchItemData(li.id) // Expensive operation
        li.data.Set(data)
        li.loaded.Set(true)
        li.loading.Set(false)
    }()
}

func (li *LazyItem) render() g.Node {
    return comps.Switch(comps.SwitchProps{
        When: reactivity.CreateMemo(func() string {
            if li.loading.Get() {
                return "loading"
            }
            if li.loaded.Get() {
                return "loaded"
            }
            return "unloaded"
        }),
        Children: []g.Node{
            comps.Match(comps.MatchProps{
                When: "unloaded",
                Children: g.Div(
                    g.Class("lazy-placeholder"),
                    g.Text("Click to load"),
                    dom.OnClick(func() {
                        li.load()
                    }),
                ),
            }),
            comps.Match(comps.MatchProps{
                When: "loading",
                Children: g.Div(
                    g.Class("loading-spinner"),
                    g.Text("Loading..."),
                ),
            }),
            comps.Match(comps.MatchProps{
                When: "loaded",
                Children: renderComplexItem(li.data.Get()),
            }),
        },
    })
}
```

## Memory Management

### Proper Signal Cleanup

```go
type Component struct {
    signals   []reactivity.Signal[any]
    memos     []reactivity.Memo[any]
    effects   []func() // Disposal functions
    disposed  bool
}

func (c *Component) addSignal(signal reactivity.Signal[any]) {
    c.signals = append(c.signals, signal)
}

func (c *Component) addMemo(memo reactivity.Memo[any]) {
    c.memos = append(c.memos, memo)
}

func (c *Component) addEffect(dispose func()) {
    c.effects = append(c.effects, dispose)
}

func (c *Component) dispose() {
    if c.disposed {
        return
    }
    
    // Dispose all effects
    for _, dispose := range c.effects {
        dispose()
    }
    
    // Clear references
    c.signals = nil
    c.memos = nil
    c.effects = nil
    c.disposed = true
}
```

### Weak References for Caches

```go
type WeakCache[K comparable, V any] struct {
    cache map[K]*WeakRef[V]
    mutex sync.RWMutex
}

type WeakRef[T any] struct {
    value T
    refs  int32
}

func NewWeakCache[K comparable, V any]() *WeakCache[K, V] {
    return &WeakCache[K, V]{
        cache: make(map[K]*WeakRef[V]),
    }
}

func (wc *WeakCache[K, V]) Get(key K, factory func() V) V {
    wc.mutex.RLock()
    if ref, exists := wc.cache[key]; exists {
        atomic.AddInt32(&ref.refs, 1)
        wc.mutex.RUnlock()
        return ref.value
    }
    wc.mutex.RUnlock()
    
    wc.mutex.Lock()
    defer wc.mutex.Unlock()
    
    // Double-check after acquiring write lock
    if ref, exists := wc.cache[key]; exists {
        atomic.AddInt32(&ref.refs, 1)
        return ref.value
    }
    
    value := factory()
    wc.cache[key] = &WeakRef[V]{
        value: value,
        refs:  1,
    }
    
    return value
}

func (wc *WeakCache[K, V]) Release(key K) {
    wc.mutex.RLock()
    ref, exists := wc.cache[key]
    wc.mutex.RUnlock()
    
    if !exists {
        return
    }
    
    if atomic.AddInt32(&ref.refs, -1) <= 0 {
        wc.mutex.Lock()
        delete(wc.cache, key)
        wc.mutex.Unlock()
    }
}
```

### Object Pooling for Frequent Allocations

```go
type ItemPool struct {
    pool sync.Pool
}

func NewItemPool() *ItemPool {
    return &ItemPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &Item{}
            },
        },
    }
}

func (p *ItemPool) Get() *Item {
    return p.pool.Get().(*Item)
}

func (p *ItemPool) Put(item *Item) {
    // Reset item state
    item.Reset()
    p.pool.Put(item)
}

// Usage in component
var itemPool = NewItemPool()

func processItems(data []ItemData) []Item {
    items := make([]Item, len(data))
    for i, d := range data {
        item := itemPool.Get()
        item.Initialize(d)
        items[i] = *item
        // Don't put back immediately - will be done when item is no longer needed
    }
    return items
}
```

## Bundle Size Optimization

### Tree Shaking Unused Code

```go
// ❌ Bad: Importing entire packages
import (
    "github.com/ozanturksever/uiwgo/comps" // Imports all helpers
    "github.com/ozanturksever/uiwgo/dom"   // Imports all DOM utilities
)

// ✅ Good: Import only what you need (if supported)
import (
    "github.com/ozanturksever/uiwgo/comps/for"
    "github.com/ozanturksever/uiwgo/comps/show"
    "github.com/ozanturksever/uiwgo/dom/events"
)
```

### Conditional Compilation

```go
//go:build !production
// +build !production

func debugLog(msg string) {
    logutil.Log(msg)
}

//go:build production
// +build production

func debugLog(msg string) {
    // No-op in production
}
```

### Lazy Loading Components

```go
type LazyComponent struct {
    loader   func() Component
    loaded   reactivity.Signal[bool]
    component reactivity.Signal[Component]
}

func NewLazyComponent(loader func() Component) *LazyComponent {
    return &LazyComponent{
        loader:    loader,
        loaded:    reactivity.CreateSignal(false),
        component: reactivity.CreateSignal[Component](nil),
    }
}

func (lc *LazyComponent) load() {
    if lc.loaded.Get() {
        return
    }
    
    go func() {
        component := lc.loader()
        lc.component.Set(component)
        lc.loaded.Set(true)
    }()
}

func (lc *LazyComponent) render() g.Node {
    return comps.Show(comps.ShowProps{
        When: lc.loaded,
        Children: func() g.Node {
            if comp := lc.component.Get(); comp != nil {
                return comp.render()
            }
            return g.Div(g.Text("Loading component..."))
        }(),
        Fallback: g.Button(
            g.Text("Load Component"),
            dom.OnClick(func() {
                lc.load()
            }),
        ),
    })
}
```

## Runtime Performance

### Minimize DOM Operations

```go
// ❌ Bad: Multiple DOM queries
func updateUI() {
    element1 := dom.GetWindow().Document().GetElementByID("item1")
    element2 := dom.GetWindow().Document().GetElementByID("item2")
    element3 := dom.GetWindow().Document().GetElementByID("item3")
    
    element1.SetTextContent("New text 1")
    element2.SetTextContent("New text 2")
    element3.SetTextContent("New text 3")
}

// ✅ Good: Batch DOM operations
func updateUI() {
    doc := dom.GetWindow().Document()
    
    // Batch queries
    elements := map[string]dom.Element{
        "item1": doc.GetElementByID("item1"),
        "item2": doc.GetElementByID("item2"),
        "item3": doc.GetElementByID("item3"),
    }
    
    // Batch updates
    updates := map[string]string{
        "item1": "New text 1",
        "item2": "New text 2",
        "item3": "New text 3",
    }
    
    for id, text := range updates {
        elements[id].SetTextContent(text)
    }
}
```

### Use RequestAnimationFrame for Smooth Animations

```go
type Animator struct {
    running   reactivity.Signal[bool]
    progress  reactivity.Signal[float64]
    duration  time.Duration
    startTime time.Time
}

func NewAnimator(duration time.Duration) *Animator {
    return &Animator{
        running:  reactivity.CreateSignal(false),
        progress: reactivity.CreateSignal(0.0),
        duration: duration,
    }
}

func (a *Animator) start() {
    if a.running.Get() {
        return
    }
    
    a.running.Set(true)
    a.startTime = time.Now()
    a.animate()
}

func (a *Animator) animate() {
    if !a.running.Get() {
        return
    }
    
    elapsed := time.Since(a.startTime)
    progress := float64(elapsed) / float64(a.duration)
    
    if progress >= 1.0 {
        a.progress.Set(1.0)
        a.running.Set(false)
        return
    }
    
    a.progress.Set(progress)
    
    // Schedule next frame
    dom.GetWindow().RequestAnimationFrame(func(float64) {
        a.animate()
    })
}
```

## Profiling and Monitoring

### Performance Monitoring

```go
type PerformanceMonitor struct {
    metrics map[string]*Metric
    mutex   sync.RWMutex
}

type Metric struct {
    Count    int64
    Total    time.Duration
    Min      time.Duration
    Max      time.Duration
    Average  time.Duration
}

func NewPerformanceMonitor() *PerformanceMonitor {
    return &PerformanceMonitor{
        metrics: make(map[string]*Metric),
    }
}

func (pm *PerformanceMonitor) Time(name string, fn func()) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        pm.recordMetric(name, duration)
    }()
    
    fn()
}

func (pm *PerformanceMonitor) recordMetric(name string, duration time.Duration) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    metric, exists := pm.metrics[name]
    if !exists {
        metric = &Metric{
            Min: duration,
            Max: duration,
        }
        pm.metrics[name] = metric
    }
    
    metric.Count++
    metric.Total += duration
    metric.Average = metric.Total / time.Duration(metric.Count)
    
    if duration < metric.Min {
        metric.Min = duration
    }
    if duration > metric.Max {
        metric.Max = duration
    }
}

func (pm *PerformanceMonitor) GetMetrics() map[string]*Metric {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    result := make(map[string]*Metric)
    for name, metric := range pm.metrics {
        result[name] = &Metric{
            Count:   metric.Count,
            Total:   metric.Total,
            Min:     metric.Min,
            Max:     metric.Max,
            Average: metric.Average,
        }
    }
    return result
}

// Usage
var monitor = NewPerformanceMonitor()

func expensiveOperation() {
    monitor.Time("expensive_operation", func() {
        // ... expensive code
    })
}
```

### Memory Usage Tracking

```go
type MemoryTracker struct {
    samples []MemorySample
    mutex   sync.Mutex
}

type MemorySample struct {
    Timestamp time.Time
    HeapSize  uint64
    GCCount   uint32
}

func NewMemoryTracker() *MemoryTracker {
    tracker := &MemoryTracker{}
    
    // Start periodic sampling
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            tracker.sample()
        }
    }()
    
    return tracker
}

func (mt *MemoryTracker) sample() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    sample := MemorySample{
        Timestamp: time.Now(),
        HeapSize:  m.HeapInuse,
        GCCount:   m.NumGC,
    }
    
    mt.mutex.Lock()
    mt.samples = append(mt.samples, sample)
    
    // Keep only last 100 samples
    if len(mt.samples) > 100 {
        mt.samples = mt.samples[1:]
    }
    mt.mutex.Unlock()
}

func (mt *MemoryTracker) GetTrend() (increasing bool, rate float64) {
    mt.mutex.Lock()
    defer mt.mutex.Unlock()
    
    if len(mt.samples) < 2 {
        return false, 0
    }
    
    first := mt.samples[0]
    last := mt.samples[len(mt.samples)-1]
    
    timeDiff := last.Timestamp.Sub(first.Timestamp).Seconds()
    sizeDiff := float64(int64(last.HeapSize) - int64(first.HeapSize))
    
    rate = sizeDiff / timeDiff
    increasing = rate > 0
    
    return
}
```

## Best Practices

### 1. Profile Before Optimizing

Always measure performance before making optimizations. Use browser dev tools and custom profiling to identify actual bottlenecks.

### 2. Optimize for the Common Case

Focus optimization efforts on code paths that run frequently or handle large amounts of data.

### 3. Use Appropriate Data Structures

Choose the right data structure for your use case:
- Maps for fast lookups
- Slices for ordered data
- Channels for communication
- Sync primitives for concurrent access

### 4. Minimize Allocations

Reduce garbage collection pressure by:
- Reusing objects when possible
- Using object pools for frequent allocations
- Avoiding unnecessary string concatenations
- Pre-allocating slices with known capacity

### 5. Batch Operations

Group related operations together to reduce overhead:
- Batch DOM updates
- Combine signal updates
- Group network requests

### 6. Use Lazy Loading

Load resources only when needed:
- Lazy load components
- Defer expensive computations
- Load data on demand

### 7. Monitor in Production

Implement monitoring to track performance in real-world usage:
- Track key metrics
- Set up alerts for performance regressions
- Analyze user behavior patterns

By following these optimization strategies and best practices, you can build UIwGo applications that perform well at scale while maintaining clean, maintainable code.