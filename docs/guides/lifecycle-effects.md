# Lifecycle & Effects

UIwGo components have a well-defined lifecycle that manages their creation, mounting, reactive behavior, and cleanup. This guide covers the complete component lifecycle, effect management, and best practices for resource management.

## Table of Contents

- [Component Lifecycle Overview](#component-lifecycle-overview)
- [Lifecycle Phases](#lifecycle-phases)
- [Effect Management](#effect-management)
- [Cleanup Patterns](#cleanup-patterns)
- [Resource Management](#resource-management)
- [Advanced Lifecycle Patterns](#advanced-lifecycle-patterns)
- [Performance Considerations](#performance-considerations)
- [Common Pitfalls](#common-pitfalls)

## Component Lifecycle Overview

### The Complete Lifecycle

UIwGo components follow a predictable lifecycle from creation to cleanup:

```
1. Construction  │ NewComponent()
   ↓             │ - Initialize signals
2. Render        │ - Set up memos
   ↓             │ - Create child components
3. Mount         │
   ↓             │ component.Render()
4. Attach        │ - Generate HTML string
   ↓             │ - Insert into DOM
5. Active        │
   ↓             │ component.Attach()
6. Cleanup       │ - Scan for data attributes
                 │ - Bind reactive behavior
                 │ - Set up event listeners
                 │
                 │ [Component is now live]
                 │ - Signals trigger updates
                 │ - User interactions work
                 │ - Effects run
                 │
                 │ component.Cleanup() (optional)
                 │ - Clean up resources
                 │ - Cancel subscriptions
                 │ - Remove event listeners
```

### Basic Component Structure

```go
type MyComponent struct {
    // 1. State (signals)
    data    *reactivity.Signal[string]
    visible *reactivity.Signal[bool]
    
    // 2. Computed state (memos)
    displayText *reactivity.Memo[string]
    
    // 3. Effects (for cleanup)
    effects []reactivity.Effect
    
    // 4. Child components
    child *ChildComponent
}

// Constructor - Phase 1
func NewMyComponent() *MyComponent {
    c := &MyComponent{
        data:    reactivity.NewSignal("initial"),
        visible: reactivity.NewSignal(true),
        effects: make([]reactivity.Effect, 0),
    }
    
    // Set up computed state
    c.displayText = reactivity.NewMemo(func() string {
        if !c.visible.Get() {
            return ""
        }
        return "Data: " + c.data.Get()
    })
    
    // Create child components
    c.child = NewChildComponent(c.data)
    
    return c
}

// Render - Phase 2
func (c *MyComponent) Render() g.Node {
    return h.Div(
        g.Class("my-component"),
        h.H2(
            g.Attr("data-text", "displayText"),
            g.Text("Loading..."),
        ),
        h.Div(
            g.Attr("data-show", "visible"),
            g.Class("content"),
            c.child.Render(),
        ),
        h.Button(
            g.Attr("data-click", "toggle"),
            g.Text("Toggle"),
        ),
    )
}

// Attach - Phase 4
func (c *MyComponent) Attach() {
    // Bind reactive behavior
    c.BindText("displayText", c.displayText)
    c.BindShow("visible", c.visible)
    c.BindClick("toggle", c.toggle)
    
    // Attach child components
    c.child.Attach()
    
    // Set up effects
    c.setupEffects()
}

// Cleanup - Phase 6
func (c *MyComponent) Cleanup() {
    // Clean up effects
    for _, effect := range c.effects {
        effect.Dispose()
    }
    
    // Clean up child components
    if c.child != nil {
        c.child.Cleanup()
    }
}

func (c *MyComponent) toggle() {
    c.visible.Set(!c.visible.Get())
}

func (c *MyComponent) setupEffects() {
    // Effect: Log visibility changes
    effect1 := reactivity.NewEffect(func() {
        visible := c.visible.Get()
        logutil.Logf("Component visibility: %t", visible)
    })
    
    c.effects = append(c.effects, effect1)
}
```

## Lifecycle Phases

### Phase 1: Construction

**Purpose**: Initialize component state and structure

```go
func NewUserProfile(userID int) *UserProfile {
    up := &UserProfile{
        userID:  userID,
        user:    reactivity.NewSignal(User{}),
        loading: reactivity.NewSignal(false),
        error:   reactivity.NewSignal(error(nil)),
    }
    
    // Set up computed state
    up.displayName = reactivity.NewMemo(func() string {
        user := up.user.Get()
        if user.Name != "" {
            return user.Name
        }
        return "Anonymous User"
    })
    
    up.isValid = reactivity.NewMemo(func() bool {
        user := up.user.Get()
        return user.ID != 0 && user.Email != ""
    })
    
    return up
}
```

**Best Practices**:
- Initialize all signals with sensible defaults
- Set up memos that depend on signals
- Create child components
- Don't perform side effects (save for Attach phase)

### Phase 2: Render

**Purpose**: Generate static HTML structure with binding markers

```go
func (up *UserProfile) Render() g.Node {
    return h.Div(
        g.Class("user-profile"),
        h.Div(
            g.Class("loading"),
            g.Attr("data-show", "loading"),
            g.Text("Loading user profile..."),
        ),
        h.Div(
            g.Class("error"),
            g.Attr("data-show", "hasError"),
            h.Span(
                g.Attr("data-text", "errorMessage"),
                g.Text("Error occurred"),
            ),
        ),
        h.Div(
            g.Class("profile"),
            g.Attr("data-show", "isLoaded"),
            h.Img(
                g.Attr("data-attr-src", "avatarURL"),
                g.Attr("alt", "User Avatar"),
            ),
            h.H2(
                g.Attr("data-text", "displayName"),
                g.Text("Loading..."),
            ),
            h.P(
                g.Attr("data-text", "email"),
                g.Text("email@example.com"),
            ),
            h.Button(
                g.Attr("data-click", "edit"),
                g.Attr("data-show", "canEdit"),
                g.Text("Edit Profile"),
            ),
        ),
    )
}
```

**Best Practices**:
- Use semantic HTML structure
- Include data attributes for all dynamic content
- Provide meaningful default content
- Keep render logic simple (move complexity to memos)

### Phase 3: Mount

**Purpose**: Insert rendered HTML into the DOM

```go
// Mounting happens automatically when you call comps.Mount()
func main() {
    userProfile := NewUserProfile(123)
    
    // This triggers: Render → Mount → Attach
    comps.Mount("user-profile-container", userProfile)
}
```

**What happens**:
1. `Render()` is called to generate HTML
2. HTML is inserted into the target DOM element
3. `Attach()` is called to make it reactive

### Phase 4: Attach

**Purpose**: Bind reactive behavior to the rendered HTML

```go
func (up *UserProfile) Attach() {
    // Bind text content
    up.BindText("displayName", up.displayName)
    up.BindText("email", up.user.Map(func(u User) string { return u.Email }))
    up.BindText("errorMessage", up.error.Map(func(e error) string {
        if e != nil {
            return e.Error()
        }
        return ""
    }))
    
    // Bind attributes
    up.BindAttr("avatarURL", "src", up.user.Map(func(u User) string {
        return u.AvatarURL
    }))
    
    // Bind visibility
    up.BindShow("loading", up.loading)
    up.BindShow("hasError", up.error.Map(func(e error) bool { return e != nil }))
    up.BindShow("isLoaded", up.isValid)
    up.BindShow("canEdit", up.user.Map(func(u User) bool { return u.CanEdit }))
    
    // Bind events
    up.BindClick("edit", up.startEditing)
    
    // Set up effects
    up.setupEffects()
    
    // Load initial data
    up.loadUser()
}
```

**Best Practices**:
- Bind all data attributes to signals or memos
- Set up event handlers
- Initialize effects
- Trigger initial data loading
- Attach child components

### Phase 5: Active

**Purpose**: Component is live and reactive

```go
// During the active phase, the component responds to:
// - Signal changes (automatic DOM updates)
// - User interactions (event handlers)
// - External events (effects)

func (up *UserProfile) startEditing() {
    up.editing.Set(true)
    // DOM automatically updates to show edit form
}

func (up *UserProfile) saveProfile(newData User) {
    up.loading.Set(true)
    
    go func() {
        err := saveUserProfile(newData)
        if err != nil {
            up.error.Set(err)
        } else {
            up.user.Set(newData)
            up.editing.Set(false)
        }
        up.loading.Set(false)
    }()
}
```

### Phase 6: Cleanup

**Purpose**: Clean up resources when component is destroyed

```go
func (up *UserProfile) Cleanup() {
    // Dispose of effects
    for _, effect := range up.effects {
        effect.Dispose()
    }
    
    // Cancel ongoing operations
    if up.cancelFunc != nil {
        up.cancelFunc()
    }
    
    // Clean up child components
    if up.editForm != nil {
        up.editForm.Cleanup()
    }
    
    // Clear references
    up.effects = nil
    up.editForm = nil
}
```

## Effect Management

### Creating and Managing Effects

```go
type EffectManager struct {
    effects []reactivity.Effect
}

func (em *EffectManager) AddEffect(effectFunc func()) {
    effect := reactivity.NewEffect(effectFunc)
    em.effects = append(em.effects, effect)
}

func (em *EffectManager) Cleanup() {
    for _, effect := range em.effects {
        effect.Dispose()
    }
    em.effects = nil
}

// Usage in component
type Component struct {
    effectManager *EffectManager
    data         *reactivity.Signal[string]
}

func (c *Component) setupEffects() {
    c.effectManager = &EffectManager{}
    
    // Effect 1: Log data changes
    c.effectManager.AddEffect(func() {
        data := c.data.Get()
        logutil.Logf("Data changed: %s", data)
    })
    
    // Effect 2: Update document title
    c.effectManager.AddEffect(func() {
        data := c.data.Get()
        if data != "" {
            dom.GetWindow().Document().SetTitle("App - " + data)
        }
    })
}

func (c *Component) Cleanup() {
    if c.effectManager != nil {
        c.effectManager.Cleanup()
    }
}
```

### Effect Patterns

#### Data Fetching Effect

```go
type DataLoader struct {
    url     *reactivity.Signal[string]
    data    *reactivity.Signal[interface{}]
    loading *reactivity.Signal[bool]
    error   *reactivity.Signal[error]
    
    cancelFunc context.CancelFunc
}

func (dl *DataLoader) setupDataFetchingEffect() {
    reactivity.NewEffect(func() {
        url := dl.url.Get()
        if url == "" {
            return
        }
        
        // Cancel previous request
        if dl.cancelFunc != nil {
            dl.cancelFunc()
        }
        
        // Create new context
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        dl.cancelFunc = cancel
        
        dl.loading.Set(true)
        dl.error.Set(nil)
        
        go func() {
            defer dl.loading.Set(false)
            
            data, err := fetchDataWithContext(ctx, url)
            if err != nil {
                if !errors.Is(err, context.Canceled) {
                    dl.error.Set(err)
                }
            } else {
                dl.data.Set(data)
            }
        }()
    })
}

func (dl *DataLoader) Cleanup() {
    if dl.cancelFunc != nil {
        dl.cancelFunc()
    }
}
```

#### Persistence Effect

```go
type PersistentSettings struct {
    theme    *reactivity.Signal[string]
    language *reactivity.Signal[string]
    
    debounceTimer *time.Timer
}

func (ps *PersistentSettings) setupPersistenceEffects() {
    // Debounced save to localStorage
    reactivity.NewEffect(func() {
        theme := ps.theme.Get()
        
        if ps.debounceTimer != nil {
            ps.debounceTimer.Stop()
        }
        
        ps.debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
            localStorage := dom.GetWindow().LocalStorage()
            localStorage.SetItem("theme", theme)
            logutil.Logf("Theme saved: %s", theme)
        })
    })
    
    reactivity.NewEffect(func() {
        language := ps.language.Get()
        
        // Immediate save for language (less frequent changes)
        localStorage := dom.GetWindow().LocalStorage()
        localStorage.SetItem("language", language)
        logutil.Logf("Language saved: %s", language)
    })
}

func (ps *PersistentSettings) Cleanup() {
    if ps.debounceTimer != nil {
        ps.debounceTimer.Stop()
    }
}
```

#### WebSocket Effect

```go
type WebSocketManager struct {
    url       *reactivity.Signal[string]
    connected *reactivity.Signal[bool]
    messages  *reactivity.Signal[[]Message]
    
    conn   *websocket.Conn
    done   chan struct{}
    cancel context.CancelFunc
}

func (wsm *WebSocketManager) setupWebSocketEffect() {
    reactivity.NewEffect(func() {
        url := wsm.url.Get()
        shouldConnect := wsm.connected.Get()
        
        if shouldConnect && url != "" {
            wsm.connect(url)
        } else {
            wsm.disconnect()
        }
    })
}

func (wsm *WebSocketManager) connect(url string) {
    if wsm.conn != nil {
        return // Already connected
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    wsm.cancel = cancel
    wsm.done = make(chan struct{})
    
    go func() {
        defer close(wsm.done)
        
        conn, err := websocket.Dial(url)
        if err != nil {
            logutil.Logf("WebSocket connection failed: %v", err)
            return
        }
        
        wsm.conn = conn
        logutil.Log("WebSocket connected")
        
        // Message reading loop
        for {
            select {
            case <-ctx.Done():
                return
            default:
                var msg Message
                if err := conn.ReadJSON(&msg); err != nil {
                    logutil.Logf("WebSocket read error: %v", err)
                    return
                }
                
                // Update messages signal
                current := wsm.messages.Get()
                updated := append(current, msg)
                wsm.messages.Set(updated)
            }
        }
    }()
}

func (wsm *WebSocketManager) disconnect() {
    if wsm.cancel != nil {
        wsm.cancel()
    }
    
    if wsm.conn != nil {
        wsm.conn.Close()
        wsm.conn = nil
    }
    
    if wsm.done != nil {
        <-wsm.done // Wait for goroutine to finish
    }
    
    logutil.Log("WebSocket disconnected")
}

func (wsm *WebSocketManager) Cleanup() {
    wsm.disconnect()
}
```

## Cleanup Patterns

### Automatic Cleanup with Defer

```go
type ResourceManager struct {
    resources []io.Closer
}

func (rm *ResourceManager) AddResource(resource io.Closer) {
    rm.resources = append(rm.resources, resource)
}

func (rm *ResourceManager) Cleanup() {
    for i := len(rm.resources) - 1; i >= 0; i-- {
        if err := rm.resources[i].Close(); err != nil {
            logutil.Logf("Error closing resource: %v", err)
        }
    }
    rm.resources = nil
}

// Usage
func (c *Component) setupResources() {
    rm := &ResourceManager{}
    
    // Add resources that need cleanup
    file, err := os.Open("data.txt")
    if err == nil {
        rm.AddResource(file)
    }
    
    conn, err := net.Dial("tcp", "localhost:8080")
    if err == nil {
        rm.AddResource(conn)
    }
    
    c.resourceManager = rm
}
```

### Effect Cleanup with Context

```go
type ContextualComponent struct {
    ctx    context.Context
    cancel context.CancelFunc
    
    data *reactivity.Signal[string]
}

func NewContextualComponent() *ContextualComponent {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &ContextualComponent{
        ctx:    ctx,
        cancel: cancel,
        data:   reactivity.NewSignal(""),
    }
}

func (cc *ContextualComponent) setupEffects() {
    // Effect with context-aware cleanup
    reactivity.NewEffect(func() {
        data := cc.data.Get()
        
        go func() {
            select {
            case <-cc.ctx.Done():
                return // Component is being cleaned up
            case <-time.After(1 * time.Second):
                // Perform delayed operation
                logutil.Logf("Delayed log: %s", data)
            }
        }()
    })
}

func (cc *ContextualComponent) Cleanup() {
    cc.cancel() // Cancels all context-aware operations
}
```

### Memory Leak Prevention

```go
type LeakProneComponent struct {
    timer    *time.Timer
    ticker   *time.Ticker
    effects  []reactivity.Effect
    channels []chan struct{}
}

func (lpc *LeakProneComponent) setupTimers() {
    // Timer that needs cleanup
    lpc.timer = time.AfterFunc(5*time.Second, func() {
        logutil.Log("Timer fired")
    })
    
    // Ticker that needs cleanup
    lpc.ticker = time.NewTicker(1 * time.Second)
    go func() {
        for range lpc.ticker.C {
            logutil.Log("Tick")
        }
    }()
}

func (lpc *LeakProneComponent) setupChannels() {
    ch := make(chan struct{})
    lpc.channels = append(lpc.channels, ch)
    
    go func() {
        for {
            select {
            case <-ch:
                return // Channel closed, exit goroutine
            case <-time.After(1 * time.Second):
                logutil.Log("Background work")
            }
        }
    }()
}

func (lpc *LeakProneComponent) Cleanup() {
    // Stop timer
    if lpc.timer != nil {
        lpc.timer.Stop()
    }
    
    // Stop ticker
    if lpc.ticker != nil {
        lpc.ticker.Stop()
    }
    
    // Close channels to stop goroutines
    for _, ch := range lpc.channels {
        close(ch)
    }
    
    // Dispose effects
    for _, effect := range lpc.effects {
        effect.Dispose()
    }
    
    // Clear references
    lpc.timer = nil
    lpc.ticker = nil
    lpc.effects = nil
    lpc.channels = nil
}
```

## Resource Management

### File Handling

```go
type FileProcessor struct {
    filename *reactivity.Signal[string]
    content  *reactivity.Signal[string]
    error    *reactivity.Signal[error]
    
    currentFile *os.File
}

func (fp *FileProcessor) setupFileEffect() {
    reactivity.NewEffect(func() {
        filename := fp.filename.Get()
        
        // Close previous file
        if fp.currentFile != nil {
            fp.currentFile.Close()
            fp.currentFile = nil
        }
        
        if filename == "" {
            fp.content.Set("")
            return
        }
        
        // Open new file
        file, err := os.Open(filename)
        if err != nil {
            fp.error.Set(err)
            return
        }
        
        fp.currentFile = file
        fp.error.Set(nil)
        
        // Read content
        content, err := io.ReadAll(file)
        if err != nil {
            fp.error.Set(err)
        } else {
            fp.content.Set(string(content))
        }
    })
}

func (fp *FileProcessor) Cleanup() {
    if fp.currentFile != nil {
        fp.currentFile.Close()
    }
}
```

### Network Resource Management

```go
type NetworkManager struct {
    endpoint *reactivity.Signal[string]
    client   *http.Client
    
    activeRequests map[string]context.CancelFunc
    mu            sync.Mutex
}

func NewNetworkManager() *NetworkManager {
    return &NetworkManager{
        endpoint:       reactivity.NewSignal(""),
        client:         &http.Client{Timeout: 30 * time.Second},
        activeRequests: make(map[string]context.CancelFunc),
    }
}

func (nm *NetworkManager) makeRequest(url string) {
    nm.mu.Lock()
    
    // Cancel previous request to same URL
    if cancel, exists := nm.activeRequests[url]; exists {
        cancel()
    }
    
    // Create new request context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    nm.activeRequests[url] = cancel
    nm.mu.Unlock()
    
    go func() {
        defer func() {
            nm.mu.Lock()
            delete(nm.activeRequests, url)
            nm.mu.Unlock()
        }()
        
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return
        }
        
        resp, err := nm.client.Do(req)
        if err != nil {
            if !errors.Is(err, context.Canceled) {
                logutil.Logf("Request failed: %v", err)
            }
            return
        }
        defer resp.Body.Close()
        
        // Process response...
    }()
}

func (nm *NetworkManager) Cleanup() {
    nm.mu.Lock()
    defer nm.mu.Unlock()
    
    // Cancel all active requests
    for url, cancel := range nm.activeRequests {
        cancel()
        delete(nm.activeRequests, url)
    }
}
```

## Advanced Lifecycle Patterns

### Lazy Initialization

```go
type LazyComponent struct {
    visible *reactivity.Signal[bool]
    
    // Lazily initialized
    expensiveChild *ExpensiveChild
    initialized    bool
}

func (lc *LazyComponent) Attach() {
    lc.BindShow("container", lc.visible)
    
    // Effect: Initialize expensive child only when visible
    reactivity.NewEffect(func() {
        visible := lc.visible.Get()
        
        if visible && !lc.initialized {
            lc.expensiveChild = NewExpensiveChild()
            lc.expensiveChild.Attach()
            lc.initialized = true
            
            logutil.Log("Expensive child initialized")
        }
    })
}

func (lc *LazyComponent) Cleanup() {
    if lc.expensiveChild != nil {
        lc.expensiveChild.Cleanup()
    }
}
```

### Conditional Lifecycle

```go
type ConditionalComponent struct {
    mode     *reactivity.Signal[string]
    children map[string]Component
    active   Component
}

func (cc *ConditionalComponent) setupModeEffect() {
    reactivity.NewEffect(func() {
        mode := cc.mode.Get()
        
        // Cleanup previous active component
        if cc.active != nil {
            cc.active.Cleanup()
            cc.active = nil
        }
        
        // Activate new component based on mode
        if child, exists := cc.children[mode]; exists {
            cc.active = child
            cc.active.Attach()
        }
    })
}

func (cc *ConditionalComponent) Cleanup() {
    if cc.active != nil {
        cc.active.Cleanup()
    }
    
    for _, child := range cc.children {
        child.Cleanup()
    }
}
```

## Performance Considerations

### Effect Batching

```go
type BatchedComponent struct {
    signals []*reactivity.Signal[int]
    sum     *reactivity.Memo[int]
}

func (bc *BatchedComponent) updateAll(values []int) {
    // Without batching: sum recomputes for each signal update
    // for i, value := range values {
    //     bc.signals[i].Set(value) // sum recomputes 10 times
    // }
    
    // With batching: sum recomputes only once
    reactivity.Batch(func() {
        for i, value := range values {
            bc.signals[i].Set(value)
        }
    }) // sum recomputes once after all updates
}
```

### Effect Debouncing

```go
type DebouncedComponent struct {
    searchTerm *reactivity.Signal[string]
    results    *reactivity.Signal[[]Result]
    
    debounceTimer *time.Timer
}

func (dc *DebouncedComponent) setupDebouncedSearch() {
    reactivity.NewEffect(func() {
        term := dc.searchTerm.Get()
        
        // Cancel previous timer
        if dc.debounceTimer != nil {
            dc.debounceTimer.Stop()
        }
        
        // Debounce search by 300ms
        dc.debounceTimer = time.AfterFunc(300*time.Millisecond, func() {
            if term != "" {
                dc.performSearch(term)
            } else {
                dc.results.Set([]Result{})
            }
        })
    })
}

func (dc *DebouncedComponent) Cleanup() {
    if dc.debounceTimer != nil {
        dc.debounceTimer.Stop()
    }
}
```

## Common Pitfalls

### Forgetting Cleanup

```go
// BAD: Memory leak
func (c *Component) setupTimer() {
    timer := time.NewTicker(1 * time.Second)
    go func() {
        for range timer.C {
            // This goroutine never stops!
            logutil.Log("Tick")
        }
    }()
}

// GOOD: Proper cleanup
type Component struct {
    ticker *time.Ticker
    done   chan struct{}
}

func (c *Component) setupTimer() {
    c.ticker = time.NewTicker(1 * time.Second)
    c.done = make(chan struct{})
    
    go func() {
        for {
            select {
            case <-c.ticker.C:
                logutil.Log("Tick")
            case <-c.done:
                return
            }
        }
    }()
}

func (c *Component) Cleanup() {
    if c.ticker != nil {
        c.ticker.Stop()
    }
    if c.done != nil {
        close(c.done)
    }
}
```

### Effect Dependency Issues

```go
// BAD: Effect doesn't track all dependencies
func (c *Component) setupEffect() {
    localVar := "static"
    
    reactivity.NewEffect(func() {
        data := c.data.Get() // Tracked dependency
        
        // This won't trigger effect re-run when otherSignal changes
        if localVar == "static" {
            other := c.otherSignal.Get() // Hidden dependency!
            logutil.Logf("Data: %s, Other: %s", data, other)
        }
    })
}

// GOOD: All dependencies are tracked
func (c *Component) setupEffect() {
    reactivity.NewEffect(func() {
        data := c.data.Get()  // Tracked
        other := c.otherSignal.Get() // Tracked
        
        if data != "" {
            logutil.Logf("Data: %s, Other: %s", data, other)
        }
    })
}
```

### Infinite Effect Loops

```go
// BAD: Infinite loop
func (c *Component) setupBadEffect() {
    reactivity.NewEffect(func() {
        count := c.count.Get()
        c.count.Set(count + 1) // This triggers the effect again!
    })
}

// GOOD: Use different signals or conditions
func (c *Component) setupGoodEffect() {
    reactivity.NewEffect(func() {
        count := c.count.Get()
        
        // Update different signal
        c.displayCount.Set(fmt.Sprintf("Count: %d", count))
        
        // Or use conditions
        if count < 10 {
            // Only update under certain conditions
            c.status.Set("low")
        }
    })
}
```

---

Next: Learn about [Core APIs](../api/core-apis.md) or explore [Control Flow](./control-flow.md) patterns.