# Reactivity & State Management

UIwGo's reactivity system is built on **signals** - a powerful primitive that enables fine-grained reactive updates. This guide covers everything you need to know about managing state and reactivity in UIwGo applications.

## Table of Contents

- [Understanding Signals](#understanding-signals)
- [Signal Types](#signal-types)
- [Computed Values (Memos)](#computed-values-memos)
- [Effects and Side Effects](#effects-and-side-effects)
- [State Management Patterns](#state-management-patterns)
- [Advanced Reactivity](#advanced-reactivity)
- [Performance Considerations](#performance-considerations)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)

## Understanding Signals

### What are Signals?

Signals are **reactive containers** that hold values and automatically notify dependents when the value changes. They form the foundation of UIwGo's reactivity system.

```go
// Create a signal with an initial value
count := reactivity.CreateSignal(0)

// Read the current value
value := count.Get() // Returns 0

// Update the value (triggers reactivity)
count.Set(42)

// All UI elements bound to this signal automatically update
```

### Signal Characteristics

**Type Safety**
```go
// Signals are strongly typed
name := reactivity.CreateSignal("Alice")     // Signal[string]
age := reactivity.CreateSignal(25)          // Signal[int]
active := reactivity.CreateSignal(true)     // Signal[bool]
items := reactivity.CreateSignal([]Item{})  // Signal[[]Item]
```

**Automatic Dependency Tracking**
```go
firstName := reactivity.CreateSignal("John")
lastName := reactivity.CreateSignal("Doe")

// This memo automatically tracks both firstName and lastName
fullName := reactivity.CreateMemo(func() string {
    return firstName.Get() + " " + lastName.Get()
})

// When either firstName or lastName changes, fullName recomputes
firstName.Set("Jane") // fullName becomes "Jane Doe"
```

**Synchronous Updates**
```go
count.Set(5)
// All dependent computations and DOM updates happen immediately
fmt.Println(count.Get()) // Always prints 5
```

### Basic Signal Operations

```go
type UserProfile struct {
    name  *reactivity.Signal[string]
    email *reactivity.Signal[string]
    age   *reactivity.Signal[int]
}

func NewUserProfile() *UserProfile {
    return &UserProfile{
        name:  reactivity.CreateSignal("Anonymous"),
        email: reactivity.CreateSignal(""),
        age:   reactivity.CreateSignal(0),
    }
}

func (u *UserProfile) UpdateName(newName string) {
    u.name.Set(newName) // Triggers all dependent updates
}

func (u *UserProfile) GetDisplayName() string {
    name := u.name.Get()
    if name == "" {
        return "Anonymous User"
    }
    return name
}
```

## Signal Types

### Primitive Types

```go
// Basic types
count := reactivity.CreateSignal(0)                    // int
price := reactivity.CreateSignal(19.99)               // float64
name := reactivity.CreateSignal("Product")             // string
active := reactivity.CreateSignal(true)               // bool
```

### Complex Types

```go
// Structs
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Email string `json:"email"`
}

user := reactivity.CreateSignal(User{
    ID:   1,
    Name: "Alice",
    Email: "alice@example.com",
})

// Slices
items := reactivity.CreateSignal([]string{"apple", "banana", "cherry"})

// Maps
scores := reactivity.CreateSignal(map[string]int{
    "alice": 100,
    "bob":   85,
})
```

### Working with Complex Types

```go
type TodoList struct {
    todos *reactivity.Signal[[]Todo]
}

type Todo struct {
    ID        int    `json:"id"`
    Text      string `json:"text"`
    Completed bool   `json:"completed"`
}

func (tl *TodoList) AddTodo(text string) {
    current := tl.todos.Get()
    newTodo := Todo{
        ID:        len(current) + 1,
        Text:      text,
        Completed: false,
    }
    
    // Create new slice with added todo
    updated := append(current, newTodo)
    tl.todos.Set(updated) // Triggers reactivity
}

func (tl *TodoList) ToggleTodo(id int) {
    current := tl.todos.Get()
    updated := make([]Todo, len(current))
    
    for i, todo := range current {
        if todo.ID == id {
            todo.Completed = !todo.Completed
        }
        updated[i] = todo
    }
    
    tl.todos.Set(updated)
}
```

### Signal Equality and Updates

```go
// Signals only trigger updates when the value actually changes
count := reactivity.CreateSignal(5)
count.Set(5) // No update triggered (same value)
count.Set(6) // Update triggered (different value)

// For complex types, use value equality
user := reactivity.CreateSignal(User{Name: "Alice"})
user.Set(User{Name: "Alice"}) // Update triggered (different instance)

// To avoid unnecessary updates, check before setting
func (u *UserProfile) SetNameIfDifferent(newName string) {
    if u.name.Get() != newName {
        u.name.Set(newName)
    }
}
```

## Computed Values (Memos)

### What are Memos?

Memos are **computed signals** that automatically recalculate when their dependencies change. They're perfect for derived state and expensive calculations.

```go
type ShoppingCart struct {
    items    *reactivity.Signal[[]CartItem]
    taxRate  *reactivity.Signal[float64]
    
    subtotal *reactivity.Memo[float64]
    tax      *reactivity.Memo[float64]
    total    *reactivity.Memo[float64]
}

type CartItem struct {
    Name     string
    Price    float64
    Quantity int
}

func NewShoppingCart() *ShoppingCart {
    cart := &ShoppingCart{
        items:   reactivity.CreateSignal([]CartItem{}),
        taxRate: reactivity.CreateSignal(0.08), // 8% tax
    }
    
    // Subtotal depends on items
    cart.subtotal = reactivity.CreateMemo(func() float64 {
        total := 0.0
        for _, item := range cart.items.Get() {
            total += item.Price * float64(item.Quantity)
        }
        return total
    })
    
    // Tax depends on subtotal and taxRate
    cart.tax = reactivity.CreateMemo(func() float64 {
        return cart.subtotal.Get() * cart.taxRate.Get()
    })
    
    // Total depends on subtotal and tax
    cart.total = reactivity.CreateMemo(func() float64 {
        return cart.subtotal.Get() + cart.tax.Get()
    })
    
    return cart
}
```

### Memo Benefits

**Automatic Caching**
```go
// Expensive computation only runs when dependencies change
expensiveResult := reactivity.CreateMemo(func() ComplexResult {
    // This only runs when input signals change
    return performExpensiveCalculation(input.Get())
})

// Multiple calls return cached result
result1 := expensiveResult.Get() // Computation runs
result2 := expensiveResult.Get() // Returns cached result
```

**Dependency Optimization**
```go
type Analytics struct {
    events *reactivity.Signal[[]Event]
    
    // Only recalculates when events change
    dailyStats   *reactivity.Memo[DailyStats]
    weeklyStats  *reactivity.Memo[WeeklyStats]
    monthlyStats *reactivity.Memo[MonthlyStats]
}

func (a *Analytics) setupMemos() {
    a.dailyStats = reactivity.CreateMemo(func() DailyStats {
        return calculateDailyStats(a.events.Get())
    })
    
    // Weekly stats depend on daily stats (not raw events)
    a.weeklyStats = reactivity.CreateMemo(func() WeeklyStats {
        return calculateWeeklyStats(a.dailyStats.Get())
    })
    
    // Monthly stats depend on weekly stats
    a.monthlyStats = reactivity.CreateMemo(func() MonthlyStats {
        return calculateMonthlyStats(a.weeklyStats.Get())
    })
}
```

### Conditional Memos

```go
type UserDashboard struct {
    user        *reactivity.Signal[User]
    showAdvanced *reactivity.Signal[bool]
    
    // Only calculates when showAdvanced is true
    advancedStats *reactivity.Memo[AdvancedStats]
}

func (ud *UserDashboard) setupMemos() {
    ud.advancedStats = reactivity.CreateMemo(func() AdvancedStats {
        if !ud.showAdvanced.Get() {
            return AdvancedStats{} // Return empty stats
        }
        
        // Expensive calculation only when needed
        return calculateAdvancedStats(ud.user.Get())
    })
}
```

## Effects and Side Effects

### Understanding Effects

Effects run side effects in response to signal changes. They're perfect for DOM manipulation, API calls, logging, and other side effects.

```go
type NotificationSystem struct {
    messages *reactivity.Signal[[]Message]
    unreadCount *reactivity.Memo[int]
}

func (ns *NotificationSystem) setupEffects() {
    // Effect: Update document title with unread count
    reactivity.CreateEffect(func() {
        count := ns.unreadCount.Get()
        title := "My App"
        if count > 0 {
            title = fmt.Sprintf("(%d) My App", count)
        }
        dom.GetWindow().Document().SetTitle(title)
    })
    
    // Effect: Log message changes
    reactivity.CreateEffect(func() {
        messages := ns.messages.Get()
        logutil.Logf("Messages updated: %d total", len(messages))
    })
    
    // Effect: Show browser notification for new messages
    var lastCount int
    reactivity.CreateEffect(func() {
        count := len(ns.messages.Get())
        if count > lastCount && lastCount > 0 {
            showBrowserNotification("New message received!")
        }
        lastCount = count
    })
}
```

### Effect Patterns

#### API Calls
```go
type UserLoader struct {
    userID *reactivity.Signal[int]
    user   *reactivity.Signal[User]
    loading *reactivity.Signal[bool]
    error  *reactivity.Signal[error]
}

func (ul *UserLoader) setupEffects() {
    // Effect: Load user when userID changes
    reactivity.CreateEffect(func() {
        id := ul.userID.Get()
        if id == 0 {
            return
        }
        
        ul.loading.Set(true)
        ul.error.Set(nil)
        
        go func() {
            user, err := fetchUser(id)
            if err != nil {
                ul.error.Set(err)
            } else {
                ul.user.Set(user)
            }
            ul.loading.Set(false)
        }()
    })
}
```

#### Local Storage Sync
```go
type Settings struct {
    theme    *reactivity.Signal[string]
    language *reactivity.Signal[string]
}

func (s *Settings) setupPersistence() {
    // Effect: Save theme to localStorage
    reactivity.CreateEffect(func() {
        theme := s.theme.Get()
        localStorage := dom.GetWindow().LocalStorage()
        localStorage.SetItem("theme", theme)
    })
    
    // Effect: Save language to localStorage
    reactivity.CreateEffect(func() {
        lang := s.language.Get()
        localStorage := dom.GetWindow().LocalStorage()
        localStorage.SetItem("language", lang)
    })
}

func (s *Settings) loadFromStorage() {
    localStorage := dom.GetWindow().LocalStorage()
    
    if theme := localStorage.GetItem("theme"); theme != "" {
        s.theme.Set(theme)
    }
    
    if lang := localStorage.GetItem("language"); lang != "" {
        s.language.Set(lang)
    }
}
```

#### Cleanup and Resource Management
```go
type WebSocketManager struct {
    connected *reactivity.Signal[bool]
    url       *reactivity.Signal[string]
    
    connection *websocket.Conn
}

func (wsm *WebSocketManager) setupEffects() {
    // Effect: Manage WebSocket connection
    reactivity.CreateEffect(func() {
        url := wsm.url.Get()
        shouldConnect := wsm.connected.Get()
        
        if shouldConnect && url != "" {
            // Connect
            if wsm.connection == nil {
                conn, err := websocket.Dial(url)
                if err == nil {
                    wsm.connection = conn
                    logutil.Log("WebSocket connected")
                }
            }
        } else {
            // Disconnect
            if wsm.connection != nil {
                wsm.connection.Close()
                wsm.connection = nil
                logutil.Log("WebSocket disconnected")
            }
        }
    })
}

func (wsm *WebSocketManager) Cleanup() {
    if wsm.connection != nil {
        wsm.connection.Close()
    }
}
```

## State Management Patterns

### Local Component State

```go
type Counter struct {
    count *reactivity.Signal[int]
    step  *reactivity.Signal[int]
}

func NewCounter() *Counter {
    return &Counter{
        count: reactivity.CreateSignal(0),
        step:  reactivity.CreateSignal(1),
    }
}

func (c *Counter) Increment() {
    c.count.Set(c.count.Get() + c.step.Get())
}

func (c *Counter) SetStep(step int) {
    c.step.Set(step)
}
```

### Shared State (Global Store)

```go
// Global application store
var AppStore = struct {
    User        *reactivity.Signal[User]
    Theme       *reactivity.Signal[string]
    Notifications *reactivity.Signal[[]Notification]
    Settings    *reactivity.Signal[AppSettings]
}{
    User:          reactivity.CreateSignal(User{}),
    Theme:         reactivity.CreateSignal("light"),
    Notifications: reactivity.CreateSignal([]Notification{}),
    Settings:      reactivity.CreateSignal(AppSettings{}),
}

// Components can access global state
type UserProfile struct {
    // Local state
    editing *reactivity.Signal[bool]
    
    // Computed from global state
    displayName *reactivity.Memo[string]
}

func NewUserProfile() *UserProfile {
    up := &UserProfile{
        editing: reactivity.CreateSignal(false),
    }
    
    up.displayName = reactivity.CreateMemo(func() string {
        user := AppStore.User.Get()
        if user.Name != "" {
            return user.Name
        }
        return "Anonymous User"
    })
    
    return up
}
```

### Context Pattern

```go
type AppContext struct {
    User     *reactivity.Signal[User]
    Theme    *reactivity.Signal[string]
    Router   *reactivity.Signal[Route]
    Settings *reactivity.Signal[Settings]
}

func NewAppContext() *AppContext {
    return &AppContext{
        User:     reactivity.CreateSignal(User{}),
        Theme:    reactivity.CreateSignal("light"),
        Router:   reactivity.CreateSignal(Route{}),
        Settings: reactivity.CreateSignal(Settings{}),
    }
}

type App struct {
    context *AppContext
    header  *Header
    content *Content
    footer  *Footer
}

func NewApp() *App {
    ctx := NewAppContext()
    
    return &App{
        context: ctx,
        header:  NewHeader(ctx),
        content: NewContent(ctx),
        footer:  NewFooter(ctx),
    }
}

// Child components receive context
type Header struct {
    context *AppContext
    
    // Local computed state
    userMenuVisible *reactivity.Signal[bool]
}

func NewHeader(ctx *AppContext) *Header {
    return &Header{
        context:         ctx,
        userMenuVisible: reactivity.CreateSignal(false),
    }
}


```

### Store Pattern with Actions

```go
type TodoStore struct {
    todos   *reactivity.Signal[[]Todo]
    filter  *reactivity.Signal[TodoFilter]
    
    // Computed state
    filteredTodos *reactivity.Memo[[]Todo]
    stats         *reactivity.Memo[TodoStats]
}

type TodoFilter string

const (
    FilterAll       TodoFilter = "all"
    FilterActive    TodoFilter = "active"
    FilterCompleted TodoFilter = "completed"
)

type TodoStats struct {
    Total     int
    Active    int
    Completed int
}

func NewTodoStore() *TodoStore {
    store := &TodoStore{
        todos:  reactivity.CreateSignal([]Todo{}),
        filter: reactivity.CreateSignal(FilterAll),
    }
    
    // Setup computed state
    store.filteredTodos = reactivity.CreateMemo(func() []Todo {
        todos := store.todos.Get()
        filter := store.filter.Get()
        
        switch filter {
        case FilterActive:
            return filterTodos(todos, func(t Todo) bool { return !t.Completed })
        case FilterCompleted:
            return filterTodos(todos, func(t Todo) bool { return t.Completed })
        default:
            return todos
        }
    })
    
    store.stats = reactivity.CreateMemo(func() TodoStats {
        todos := store.todos.Get()
        stats := TodoStats{Total: len(todos)}
        
        for _, todo := range todos {
            if todo.Completed {
                stats.Completed++
            } else {
                stats.Active++
            }
        }
        
        return stats
    })
    
    return store
}

// Actions
func (ts *TodoStore) AddTodo(text string) {
    current := ts.todos.Get()
    newTodo := Todo{
        ID:        generateID(),
        Text:      text,
        Completed: false,
        CreatedAt: time.Now(),
    }
    
    updated := append(current, newTodo)
    ts.todos.Set(updated)
}

func (ts *TodoStore) ToggleTodo(id int) {
    current := ts.todos.Get()
    updated := make([]Todo, len(current))
    
    for i, todo := range current {
        if todo.ID == id {
            todo.Completed = !todo.Completed
        }
        updated[i] = todo
    }
    
    ts.todos.Set(updated)
}

func (ts *TodoStore) SetFilter(filter TodoFilter) {
    ts.filter.Set(filter)
}

func (ts *TodoStore) ClearCompleted() {
    current := ts.todos.Get()
    active := filterTodos(current, func(t Todo) bool { return !t.Completed })
    ts.todos.Set(active)
}
```

## Advanced Reactivity

### Batched Updates

```go
type FormData struct {
    firstName *reactivity.Signal[string]
    lastName  *reactivity.Signal[string]
    email     *reactivity.Signal[string]
    
    isValid *reactivity.Memo[bool]
}

func (fd *FormData) UpdateFromAPI(data APIResponse) {
    // Multiple signal updates in sequence
    // Each triggers dependent computations
    fd.firstName.Set(data.FirstName) // isValid recomputes
    fd.lastName.Set(data.LastName)   // isValid recomputes again
    fd.email.Set(data.Email)         // isValid recomputes again
    
    // Consider batching for performance:
    // reactivity.Batch(func() {
    //     fd.firstName.Set(data.FirstName)
    //     fd.lastName.Set(data.LastName)
    //     fd.email.Set(data.Email)
    // }) // isValid recomputes only once
}
```

### Signal Composition

```go
type CombinedSignal struct {
    signals []reactivity.Signal[string]
    combined *reactivity.Memo[string]
}

func NewCombinedSignal(signals ...reactivity.Signal[string]) *CombinedSignal {
    cs := &CombinedSignal{
        signals: signals,
    }
    
    cs.combined = reactivity.CreateMemo(func() string {
        var parts []string
        for _, signal := range cs.signals {
            if value := signal.Get(); value != "" {
                parts = append(parts, value)
            }
        }
        return strings.Join(parts, " ")
    })
    
    return cs
}
```

### Async Signal Updates

```go
type AsyncDataLoader struct {
    url     *reactivity.Signal[string]
    data    *reactivity.Signal[interface{}]
    loading *reactivity.Signal[bool]
    error   *reactivity.Signal[error]
}

func (adl *AsyncDataLoader) setupEffects() {
    reactivity.CreateEffect(func() {
        url := adl.url.Get()
        if url == "" {
            return
        }
        
        adl.loading.Set(true)
        adl.error.Set(nil)
        
        go func() {
            data, err := fetchData(url)
            
            // Update signals from goroutine
            if err != nil {
                adl.error.Set(err)
            } else {
                adl.data.Set(data)
            }
            adl.loading.Set(false)
        }()
    })
}
```

## Performance Considerations

### Memo Optimization

```go
// Expensive computation - use memo
expensiveResult := reactivity.CreateMemo(func() ComplexResult {
    return performExpensiveCalculation(input.Get())
})

// Simple computation - direct access might be better
simpleResult := func() string {
    return "Hello, " + name.Get()
}
```

### Effect Cleanup

```go
type Component struct {
    effects []reactivity.Effect
}

func (c *Component) setupEffects() {
    effect1 := reactivity.CreateEffect(func() {
        // Effect logic
    })
    
    effect2 := reactivity.CreateEffect(func() {
        // Effect logic
    })
    
    c.effects = []reactivity.Effect{effect1, effect2}
}

func (c *Component) Cleanup() {
    for _, effect := range c.effects {
        effect.Dispose() // Clean up effect
    }
}
```

### Memory Management

```go
// Avoid creating signals in loops
for i := 0; i < 1000; i++ {
    // BAD: Creates 1000 signals
    signal := reactivity.CreateSignal(i)
}

// GOOD: Use a single signal for the collection
items := reactivity.CreateSignal(make([]int, 1000))
```

## Common Patterns

### Toggle Pattern

```go
type ToggleComponent struct {
    visible *reactivity.Signal[bool]
}

func (tc *ToggleComponent) Toggle() {
    tc.visible.Set(!tc.visible.Get())
}
```

### Loading State Pattern

```go
type AsyncComponent struct {
    loading *reactivity.Signal[bool]
    data    *reactivity.Signal[interface{}]
    error   *reactivity.Signal[error]
}

func (ac *AsyncComponent) LoadData() {
    ac.loading.Set(true)
    ac.error.Set(nil)
    
    go func() {
        defer ac.loading.Set(false)
        
        data, err := fetchData()
        if err != nil {
            ac.error.Set(err)
        } else {
            ac.data.Set(data)
        }
    }()
}
```

### Form Validation Pattern

```go
type FormValidator struct {
    email    *reactivity.Signal[string]
    password *reactivity.Signal[string]
    
    emailValid    *reactivity.Memo[bool]
    passwordValid *reactivity.Memo[bool]
    formValid     *reactivity.Memo[bool]
}

func NewFormValidator() *FormValidator {
    fv := &FormValidator{
        email:    reactivity.CreateSignal(""),
        password: reactivity.CreateSignal(""),
    }
    
    fv.emailValid = reactivity.CreateMemo(func() bool {
        email := fv.email.Get()
        return strings.Contains(email, "@") && len(email) > 5
    })
    
    fv.passwordValid = reactivity.CreateMemo(func() bool {
        password := fv.password.Get()
        return len(password) >= 8
    })
    
    fv.formValid = reactivity.CreateMemo(func() bool {
        return fv.emailValid.Get() && fv.passwordValid.Get()
    })
    
    return fv
}
```

## Troubleshooting

### Common Issues

**Infinite Update Loops**
```go
// BAD: Creates infinite loop
effect := reactivity.CreateEffect(func() {
    count := counter.Get()
    counter.Set(count + 1) // This triggers the effect again!
})

// GOOD: Use conditions or different signals
effect := reactivity.CreateEffect(func() {
    count := counter.Get()
    if count < 10 {
        otherSignal.Set(count + 1) // Update different signal
    }
})
```

**Memory Leaks**
```go
// BAD: Effect never cleaned up
func createComponent() {
    signal := reactivity.CreateSignal(0)
    reactivity.CreateEffect(func() {
        // This effect will never be disposed
        logutil.Log(signal.Get())
    })
}

// GOOD: Store and clean up effects
type Component struct {
    signal *reactivity.Signal[int]
    effect reactivity.Effect
}

func (c *Component) setup() {
    c.effect = reactivity.CreateEffect(func() {
        logutil.Log(c.signal.Get())
    })
}

func (c *Component) Cleanup() {
    c.effect.Dispose()
}
```

**Stale Closures**
```go
// BAD: Captures stale value
for i := 0; i < 3; i++ {
    reactivity.CreateEffect(func() {
        logutil.Log(i) // Always logs 3!
    })
}

// GOOD: Capture value properly
for i := 0; i < 3; i++ {
    index := i // Capture current value
    reactivity.CreateEffect(func() {
        logutil.Log(index) // Logs 0, 1, 2
    })
}
```

### Debugging Tips

**Signal Tracing**
```go
// Add logging to track signal changes
count := reactivity.CreateSignal(0)

reactivity.CreateEffect(func() {
    value := count.Get()
    logutil.Logf("Count changed to: %d", value)
})
```

**Dependency Tracking**
```go
// Log which signals a memo depends on
computed := reactivity.CreateMemo(func() string {
    logutil.Log("Computing result...")
    a := signalA.Get()
    b := signalB.Get()
    logutil.Logf("Dependencies: a=%v, b=%v", a, b)
    return fmt.Sprintf("%v-%v", a, b)
})
```

---

Next: Learn about [Lifecycle & Effects](./lifecycle-effects.md) or explore [Control Flow](./control-flow.md) patterns.