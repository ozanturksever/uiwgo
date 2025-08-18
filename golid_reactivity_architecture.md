# Golid Reactivity System Architecture v2.0
## SolidJS-Inspired Fine-Grained Reactive System

### Executive Summary

This architecture redesigns Golid's reactivity system to eliminate critical performance bottlenecks by implementing a SolidJS-inspired fine-grained reactivity model with direct DOM manipulation, automatic dependency tracking, and robust lifecycle management.

### Key Problems Addressed

1. **Infinite Lifecycle-Signal-Observer Loops** - Causing 100% CPU usage
2. **Inefficient Virtual DOM Diffing** - String-based replacements causing performance degradation
3. **Unscoped Signals** - Leading to memory leaks
4. **Event System Leaks** - Lack of deterministic unsubscription
5. **Lifecycle Hook Cascades** - Exponential execution chains

---

## 1. Core Reactivity Primitives

### 1.1 Signal Primitive

```go
type Signal[T any] struct {
    id          uint64
    value       T
    subscribers map[uint64]*Computation
    owner       *Owner
    compareFn   func(prev, next T) bool
    mutex       sync.RWMutex
}

type SignalOptions[T any] struct {
    Name      string
    Equals    func(prev, next T) bool
    Owner     *Owner
}

// Core Signal API
func CreateSignal[T any](initial T, options ...SignalOptions[T]) (*Signal[T], func(T))
func (s *Signal[T]) Get() T
func (s *Signal[T]) Set(value T)
func (s *Signal[T]) Update(fn func(T) T)
func (s *Signal[T]) Subscribe(computation *Computation)
func (s *Signal[T]) Unsubscribe(computation *Computation)
```

### 1.2 Computation Primitive (Effect)

```go
type Computation struct {
    id           uint64
    fn           func()
    dependencies map[uint64]*Signal[any]
    owner        *Owner
    state        ComputationState
    cleanups     []func()
    context      *Context
}

type ComputationState int
const (
    Clean ComputationState = iota
    Check
    Dirty
)

// Core Effect API
func CreateEffect(fn func(), owner *Owner) *Computation
func CreateMemo[T any](fn func() T, owner *Owner) *Signal[T]
func CreateComputed[T any](fn func() T, owner *Owner) func() T
```

### 1.3 Owner Context (Scope Management)

```go
type Owner struct {
    id          uint64
    parent      *Owner
    children    []*Owner
    computations []*Computation
    signals     []*Signal[any]
    cleanups    []func()
    context     map[string]interface{}
    disposed    bool
    mutex       sync.RWMutex
}

// Owner API
func CreateRoot[T any](fn func() T) (T, func())
func CreateContext[T any](defaultValue T) *Context[T]
func RunWithOwner[T any](owner *Owner, fn func() T) T
func OnCleanup(fn func())
```

---

## 2. Dependency Tracking & Batching

### 2.1 Automatic Dependency Tracking

```go
type DependencyTracker struct {
    current     *Computation
    stack       []*Computation
    transitions map[uint64]*Transition
}

type Transition struct {
    id        uint64
    pending   []*Update
    effects   []*Computation
    completed chan bool
}

type Update struct {
    signal *Signal[any]
    value  interface{}
}

// Global tracking context
var tracker = &DependencyTracker{
    transitions: make(map[uint64]*Transition),
}

// Tracking API
func Track[T any](fn func() T) T
func Batch(fn func())
func StartTransition() *Transition
func (t *Transition) Complete()
```

### 2.2 Batched Updates with Scheduler

```go
type Scheduler struct {
    queue       *PriorityQueue
    microtask   chan *ScheduledTask
    running     bool
    batchDepth  int
    mutex       sync.Mutex
}

type ScheduledTask struct {
    priority    Priority
    computation *Computation
    timestamp   int64
}

type Priority int
const (
    UserBlocking Priority = iota  // Immediate (user input)
    Normal                        // Default priority
    Idle                         // Low priority
)

// Scheduler API
func (s *Scheduler) Schedule(task *ScheduledTask)
func (s *Scheduler) Flush()
func (s *Scheduler) RunBatch(fn func())
```

---

## 3. Direct DOM Manipulation

### 3.1 DOM Binding System

```go
type DOMBinding struct {
    element     js.Value
    property    string
    computation *Computation
    owner       *Owner
}

type DOMRenderer struct {
    bindings map[string]*DOMBinding
    templates map[string]*Template
    hydration *HydrationContext
}

// Direct DOM API
func BindAttribute(el js.Value, attr string, value func() string) *DOMBinding
func BindText(el js.Value, text func() string) *DOMBinding
func BindClass(el js.Value, classes func() map[string]bool) *DOMBinding
func BindStyle(el js.Value, styles func() map[string]string) *DOMBinding
```

### 3.2 Template Compilation

```go
type Template struct {
    id       string
    factory  func() js.Value
    bindings []BindingDescriptor
    static   bool
}

type BindingDescriptor struct {
    path     []int  // DOM tree path to element
    type     BindingType
    accessor func() interface{}
}

type BindingType int
const (
    TextBinding BindingType = iota
    AttributeBinding
    PropertyBinding
    EventBinding
)

// Template API
func CompileTemplate(html string) *Template
func (t *Template) Clone() js.Value
func (t *Template) Hydrate(root js.Value)
```

### 3.3 Fine-Grained DOM Updates

```go
type DOMPatcher struct {
    operations []DOMOperation
    batching   bool
}

type DOMOperation struct {
    type     OperationType
    target   js.Value
    property string
    value    interface{}
    oldValue interface{}
}

type OperationType int
const (
    SetText OperationType = iota
    SetAttribute
    RemoveAttribute
    InsertNode
    RemoveNode
    ReplaceNode
)

// Patching API
func (p *DOMPatcher) QueueOperation(op DOMOperation)
func (p *DOMPatcher) Flush()
func (p *DOMPatcher) BatchUpdates(fn func())
```

---

## 4. Component Lifecycle Management

### 4.1 Component System

```go
type Component struct {
    id       uint64
    owner    *Owner
    props    *Signal[map[string]interface{}]
    state    ComponentState
    render   func() js.Value
    mounted  js.Value
    context  *ComponentContext
}

type ComponentState int
const (
    Unmounted ComponentState = iota
    Mounting
    Mounted
    Updating
    Unmounting
)

type ComponentContext struct {
    scheduler   *Scheduler
    renderer    *DOMRenderer
    errorBoundary *ErrorBoundary
}

// Component API
func CreateComponent[P any](fn func(props P) js.Value) func(P) *Component
func (c *Component) Mount(container js.Value)
func (c *Component) Update(props map[string]interface{})
func (c *Component) Unmount()
```

### 4.2 Lifecycle Hooks

```go
type LifecycleHooks struct {
    onMount    []func()
    onCleanup  []func()
    onError    []func(error)
}

// Lifecycle API
func OnMount(fn func())
func OnCleanup(fn func())
func OnError(fn func(error))
func CreateLifecycle() *LifecycleHooks
```

### 4.3 Lifecycle Cascade Prevention

```go
type LifecycleGuard struct {
    depth    int
    maxDepth int
    visited  map[uint64]bool
}

func (g *LifecycleGuard) Enter(id uint64) bool {
    if g.depth >= g.maxDepth {
        return false
    }
    if g.visited[id] {
        return false
    }
    g.depth++
    g.visited[id] = true
    return true
}

func (g *LifecycleGuard) Exit(id uint64) {
    g.depth--
    delete(g.visited, id)
}
```

---

## 5. Memory Management & Cleanup

### 5.1 Resource Management

```go
type Resource[T any] struct {
    signal   *Signal[ResourceState[T]]
    fetcher  func() (T, error)
    owner    *Owner
    cache    *ResourceCache
}

type ResourceState[T any] struct {
    Loading bool
    Data    *T
    Error   error
}

type ResourceCache struct {
    entries map[string]*CacheEntry
    maxSize int
    ttl     time.Duration
}

// Resource API
func CreateResource[T any](fetcher func() (T, error)) *Resource[T]
func (r *Resource[T]) Read() T
func (r *Resource[T]) Refetch()
func (r *Resource[T]) Mutate(value T)
```

### 5.2 Automatic Cleanup

```go
type CleanupManager struct {
    registry map[uint64][]func()
    order    []uint64
    mutex    sync.RWMutex
}

func (m *CleanupManager) Register(owner uint64, cleanup func())
func (m *CleanupManager) Execute(owner uint64)
func (m *CleanupManager) ExecuteAll()

// Weak reference support for preventing leaks
type WeakRef[T any] struct {
    ptr unsafe.Pointer
}

func CreateWeakRef[T any](value *T) *WeakRef[T]
func (w *WeakRef[T]) Deref() *T
```

### 5.3 Memory Leak Prevention

```go
type LeakDetector struct {
    allocations map[uint64]*Allocation
    threshold   int
    enabled     bool
}

type Allocation struct {
    id        uint64
    type      string
    size      int
    timestamp time.Time
    stack     []uintptr
}

func (d *LeakDetector) Track(id uint64, typ string, size int)
func (d *LeakDetector) Release(id uint64)
func (d *LeakDetector) Report() []LeakReport
```

---

## 6. Error Handling & Recovery

### 6.1 Error Boundaries

```go
type ErrorBoundary struct {
    owner     *Owner
    fallback  func(error) js.Value
    onError   func(error, *ErrorInfo)
    recovered bool
}

type ErrorInfo struct {
    Error      error
    Component  string
    Stack      []uintptr
    Props      map[string]interface{}
    Timestamp  time.Time
}

// Error Boundary API
func CreateErrorBoundary(fallback func(error) js.Value) *ErrorBoundary
func (e *ErrorBoundary) Catch(fn func()) (err error)
func (e *ErrorBoundary) Reset()
```

### 6.2 Error Recovery Strategies

```go
type RecoveryStrategy int
const (
    Retry RecoveryStrategy = iota
    Fallback
    Propagate
    Ignore
)

type ErrorHandler struct {
    strategies map[string]RecoveryStrategy
    retries    map[string]int
    maxRetries int
}

func (h *ErrorHandler) Handle(err error, context string) bool
func (h *ErrorHandler) SetStrategy(pattern string, strategy RecoveryStrategy)
```

---

## 7. Event System Redesign

### 7.1 Event Delegation

```go
type EventDelegator struct {
    handlers map[string]map[uint64]EventHandler
    root     js.Value
    active   map[string]bool
}

type EventHandler struct {
    id       uint64
    selector string
    handler  func(js.Value)
    options  EventOptions
}

type EventOptions struct {
    Capture bool
    Once    bool
    Passive bool
}

// Event API
func (d *EventDelegator) On(event string, selector string, handler func(js.Value))
func (d *EventDelegator) Off(event string, id uint64)
func (d *EventDelegator) Delegate(event string)
```

### 7.2 Event Subscription Management

```go
type EventSubscription struct {
    id        uint64
    event     string
    handler   func(js.Value)
    cleanup   func()
    autoClean bool
}

type EventManager struct {
    subscriptions map[uint64]*EventSubscription
    delegator     *EventDelegator
    owner         *Owner
}

func (m *EventManager) Subscribe(event string, handler func(js.Value)) func()
func (m *EventManager) Unsubscribe(id uint64)
func (m *EventManager) UnsubscribeAll()
```

---

## 8. Integration with Existing Codebase

### 8.1 Migration Path

```go
// Compatibility layer for existing Signal API
type LegacySignal[T any] struct {
    *Signal[T]
}

func NewSignal[T any](initial T) *LegacySignal[T] {
    signal, _ := CreateSignal(initial)
    return &LegacySignal[T]{signal}
}

func (s *LegacySignal[T]) Get() T {
    return s.Signal.Get()
}

func (s *LegacySignal[T]) Set(value T) {
    s.Signal.Set(value)
}
```

### 8.2 Gomponents Integration

```go
type GomponentsAdapter struct {
    renderer *DOMRenderer
    compiler *TemplateCompiler
}

func (a *GomponentsAdapter) RenderNode(node gomponents.Node) js.Value
func (a *GomponentsAdapter) HydrateNode(node gomponents.Node, element js.Value)
func (a *GomponentsAdapter) CreateReactive(fn func() gomponents.Node) *Component
```

### 8.3 Router Integration

```go
type ReactiveRouter struct {
    routes   *Signal[[]Route]
    current  *Signal[string]
    params   *Signal[map[string]string]
    renderer *DOMRenderer
}

func (r *ReactiveRouter) Navigate(path string)
func (r *ReactiveRouter) Match() *RouteMatch
func (r *ReactiveRouter) Render() js.Value
```

---

## 9. Performance Optimizations

### 9.1 Compilation Optimizations

```go
type Optimizer struct {
    staticHoisting bool
    deadCodeElim   bool
    inlineExpand   bool
}

func (o *Optimizer) OptimizeTemplate(template *Template) *Template
func (o *Optimizer) OptimizeComponent(component *Component) *Component
```

### 9.2 Runtime Optimizations

```go
type RuntimeOptimizer struct {
    memoization  *MemoCache
    lazy         *LazyEvaluator
    prefetch     *Prefetcher
}

type MemoCache struct {
    cache map[uint64]interface{}
    hits  int
    miss  int
}

type LazyEvaluator struct {
    threshold time.Duration
    deferred  []*Computation
}
```

---

## 10. Implementation Phases

### Phase 1: Core Reactivity (Week 1-2)
- Implement Signal, Computation, Owner primitives
- Implement dependency tracking
- Implement batch updates and scheduler

### Phase 2: Direct DOM (Week 3-4)
- Implement DOM binding system
- Implement template compilation
- Remove virtual DOM dependencies

### Phase 3: Lifecycle Management (Week 5)
- Implement component system
- Implement lifecycle hooks
- Implement cascade prevention

### Phase 4: Memory & Cleanup (Week 6)
- Implement resource management
- Implement automatic cleanup
- Implement leak detection

### Phase 5: Error Handling (Week 7)
- Implement error boundaries
- Implement recovery strategies
- Implement error reporting

### Phase 6: Integration (Week 8)
- Implement compatibility layer
- Migrate existing components
- Performance testing

---

## 11. Testing Strategy

### 11.1 Unit Tests

```go
// Test automatic dependency tracking
func TestDependencyTracking(t *testing.T)

// Test batch updates
func TestBatchedUpdates(t *testing.T)

// Test memory cleanup
func TestAutomaticCleanup(t *testing.T)

// Test cascade prevention
func TestLifecycleCascadePrevention(t *testing.T)
```

### 11.2 Integration Tests

```go
// Test component lifecycle
func TestComponentLifecycle(t *testing.T)

// Test DOM updates
func TestDirectDOMManipulation(t *testing.T)

// Test error boundaries
func TestErrorBoundaries(t *testing.T)
```

### 11.3 Performance Benchmarks

```go
// Benchmark signal updates
func BenchmarkSignalUpdates(b *testing.B)

// Benchmark DOM operations
func BenchmarkDOMOperations(b *testing.B)

// Benchmark memory usage
func BenchmarkMemoryUsage(b *testing.B)
```

---

## 12. Migration Guide

### 12.1 Signal Migration

```go
// Old API
signal := golid.NewSignal(0)
value := signal.Get()
signal.Set(10)

// New API
signal, setSignal := CreateSignal(0)
value := signal.Get()
setSignal(10)
```

### 12.2 Component Migration

```go
// Old API
component := golid.NewComponent(func() Node {
    return Div(Text("Hello"))
})

// New API
component := CreateComponent(func(props struct{}) js.Value {
    return h("div", nil, "Hello")
})
```

### 12.3 Effect Migration

```go
// Old API
golid.Watch(func() {
    value := signal.Get()
    // side effect
})

// New API
CreateEffect(func() {
    value := signal.Get()
    // side effect
}, nil)
```

---

## Conclusion

This architecture provides a comprehensive solution to Golid's performance bottlenecks by:

1. **Eliminating infinite loops** through proper dependency tracking and cascade prevention
2. **Removing virtual DOM overhead** with direct DOM manipulation
3. **Preventing memory leaks** through scoped ownership and automatic cleanup
4. **Fixing event system leaks** with deterministic subscription management
5. **Stopping lifecycle cascades** with proper guards and depth limits

The design maintains compatibility with existing code while providing a clear migration path to the new reactive system. The fine-grained reactivity model inspired by SolidJS ensures optimal performance while preserving the developer experience.