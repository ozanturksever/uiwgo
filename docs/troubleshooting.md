# Troubleshooting & FAQ

This guide helps you diagnose and solve common issues when developing with UIwGo. It covers debugging techniques, performance problems, common pitfalls, and frequently asked questions.

## Table of Contents

- [Quick Debugging Checklist](#quick-debugging-checklist)
- [Common Issues](#common-issues)
- [Action System Issues](#action-system-issues)
- [Debugging Techniques](#debugging-techniques)
- [Performance Issues](#performance-issues)
- [Build & Development Issues](#build--development-issues)
- [Browser Compatibility](#browser-compatibility)
- [Frequently Asked Questions](#frequently-asked-questions)

## Quick Debugging Checklist

When something isn't working, check these items first:

### ✅ Component Basics
- [ ] Component is a function that returns a `gomponents.Node`.
- [ ] The component function is properly mounted with `comps.Mount("id", MyComponent)`.
- [ ] `gomponents` structure is valid Go code.

### ✅ Reactivity
- [ ] Signals are created with `reactivity.CreateSignal()`.
- [ ] Memos and Effects access signals with `.Get()` to track them as dependencies.
- [ ] Reactive content is rendered with helpers like `comps.BindText(signal.Get)`.
- [ ] No infinite loops in effects (e.g., an effect setting a signal that it depends on).

### ✅ DOM Interaction & Events
- [ ] DOM element access and event binding happens inside `comps.OnMount`.
- [ ] Element IDs used in `dom.GetElementByID("my-id")` match the `ID("my-id")` in the `gomponents` tree.
- [ ] Event handlers are bound to the correct `dom.Element` objects.
- [ ] `comps.OnCleanup` is used to remove global listeners or stop timers to prevent memory leaks.

### ✅ Browser Console
- [ ] Check for JavaScript errors.
- [ ] Verify the `.wasm` module loads successfully (check the Network tab).
- [ ] Look for any logs from `logutil`.

### ✅ Action System
- [ ] Action types are properly defined using [`DefineAction()`](action/types.go:14) or [`DefineQuery()`](action/types.go:25).
- [ ] Actions are dispatched to the correct bus instance (global vs local).
- [ ] Subscriptions are properly disposed to prevent memory leaks.
- [ ] Bridge signals are disposed when components unmount.
- [ ] Query handlers are registered before queries are dispatched.
- [ ] Observability is enabled for debugging action flow.

## Common Issues

### Component Not Rendering or Blank

**Problem**: The component appears blank or doesn't render at all.

**Solutions**:

1.  **Check Mounting**: Ensure `comps.Mount` is called correctly with the right element ID and component function.
    ```go
    // GOOD: Mount the component function
    func main() {
        comps.Mount("app", MyComponent)
        select {}
    }
    ```

2.  **Check HTML Host File**: Make sure you have an element in your `index.html` with the ID you're mounting to.
    ```html
    <!-- index.html -->
    <body>
        <div id="app"></div> <!-- This ID must match the Mount call -->
        <script src="wasm_exec.js"></script>
        <!-- ... -->
    </body>
    ```

3.  **Ensure Signal Has a Value**: If using `comps.BindText`, make sure the signal has an initial value that isn't empty.
    ```go
    // GOOD
    message := reactivity.CreateSignal("Hello, World!")

    // POTENTIAL ISSUE (will render nothing)
    message := reactivity.CreateSignal("")
    ```

### Reactivity Not Working

**Problem**: The UI doesn't update when a signal's value changes.

**Solutions**:

1.  **Check Dependencies**: Ensure your Memos and Effects are calling `.Get()` on the signals they should be tracking.
    ```go
    // BAD: Effect doesn't track the 'count' signal
reactivity.CreateEffect(func() {
    logutil.Log("This runs only once.")
})

// GOOD: Effect tracks the 'count' signal
reactivity.CreateEffect(func() {
    value := count.Get() // Tracks the signal
    logutil.Logf("Count is now: %d", value)
})
    ```

2.  **Verify Reactive Binding**: Make sure you are using a reactive helper like `comps.BindText`.
    ```go
    // BAD: This will only render the initial value
    return Span(Text(fmt.Sprintf("%d", count.Get())))

    // GOOD: This creates a reactive text node
    return Span(comps.BindText(func() string {
        return fmt.Sprintf("%d", count.Get())
    }))
    ```

3.  **Check for Disposed Effects**: An effect will stop working if its scope is cleaned up or if it's manually disposed. Use `comps.OnCleanup` to manage effect lifecycles.

### Event Handlers Not Working

**Problem**: Clicks or other events don't trigger the expected Go functions.

**Solutions**:

1.  **Bind Inside `OnMount`**: DOM elements do not exist until the component is mounted. All event binding must happen inside `comps.OnMount`.
    ```go
    // BAD: This will panic because the element is not yet in the DOM
    incrementBtn := dom.GetElementByID("increment-btn")
    dom.BindClickToCallback(incrementBtn, handler)

    // GOOD
    comps.OnMount(func() {
        incrementBtn := dom.GetElementByID("increment-btn")
        if incrementBtn != nil {
            dom.BindClickToCallback(incrementBtn, handler)
        }
    })
    ```

2.  **Verify Element ID**: Double-check that the ID used in `dom.GetElementByID` exactly matches the `ID()` attribute in your `gomponents` tree.
    ```go
    // In the component function's return
    Button(ID("increment-btn"), Text("+"))

    // In comps.OnMount
    comps.OnMount(func() {
        // The ID "increment-btn" must match
        btn := dom.GetElementByID("increment-btn")
        // ...
    })
    ```

3.  **Check for `nil` Elements**: Always check if `GetElementByID` returned `nil` before trying to bind an event to it. This tells you if the element was not found.
    ```go
    if btn == nil {
        logutil.Log("ERROR: Increment button not found in the DOM!")
        return
    }
    dom.BindClickToCallback(btn, handler)
    ```

### Memory Leaks

**Problem**: Application becomes slow over time, or the browser becomes unresponsive.

**Solutions**:

1.  **Implement Cleanup with `OnCleanup`**: If you set up manual event listeners, timers, or WebSocket connections, you must clean them up in a `comps.OnCleanup` hook.
    ```go
    func MyComponent() g.Node {
        ticker := time.NewTicker(1 * time.Second)

        comps.OnCleanup(func() {
            logutil.Log("Stopping ticker.")
            ticker.Stop() // IMPORTANT: Stop the ticker to prevent a leak.
        })

        go func() {
            for range ticker.C {
                logutil.Log("Tick")
            }
        }()
        // ...
        return Div()
    }
    ```

2.  **Avoid Infinite Effect Loops**: An effect that sets a signal it depends on will create an infinite loop.
    ```go
    // BAD: Infinite loop
    reactivity.CreateEffect(func() {
        count := counter.Get()
        counter.Set(count + 1) // This triggers the effect again!
    })

    // GOOD: Use a separate signal for the output
    reactivity.CreateEffect(func() {
        count := counter.Get()
        displayText.Set(fmt.Sprintf("Count: %d", count))
    })
    ```

### WASM Loading Issues

**Problem**: Application doesn't start or shows WASM-related errors in the browser console.

**Solutions**:

1.  **Check WASM File Path**: Ensure the `fetch("main.wasm")` path in your `index.html` is correct relative to where the HTML file is served.
    ```html
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
            .then((result) => go.run(result.instance))
            .catch((err) => console.error("WASM loading failed:", err));
    </script>
    ```

2.  **Verify Build Process**: Make sure the WASM file was compiled successfully.
    ```bash
    # Run the build for your example
    make build counter
    # Check that the output file exists
    ls examples/counter/main.wasm
    ```

3.  **Check Server MIME Types**: The development server handles this automatically, but for a custom server, ensure it serves `.wasm` files with the `Content-Type: application/wasm` header.

## Action System Issues

### Subscriptions Not Receiving Actions

**Problem**: Action handlers don't trigger when actions are dispatched.

**Solutions**:

1.  **Check Action Type Matching**: Ensure the action type in [`Dispatch()`](action/bus.go:123) matches exactly with [`Subscribe()`](action/bus.go:18).
    ```go
    // BAD: Mismatched action types
    userAction := action.DefineAction[User]("user.login")
    bus.Subscribe("user.signin", handler) // Wrong action type
    bus.Dispatch(action.Action[string]{Type: "user.login", Payload: userData})

    // GOOD: Matching action types
    userAction := action.DefineAction[User]("user.login")
    bus.Subscribe("user.login", handler)
    bus.Dispatch(action.Action[string]{Type: "user.login", Payload: userData})
    ```

2.  **Verify Bus Instance**: Make sure you're using the same bus instance for both subscription and dispatch.
    ```go
    // BAD: Using different bus instances
    bus1 := action.New()
    bus2 := action.New()
    bus1.Subscribe("test", handler)
    bus2.Dispatch("test") // Won't reach handler

    // GOOD: Using the same bus instance
    bus := action.New()
    bus.Subscribe("test", handler)
    bus.Dispatch("test")
    ```

3.  **Check Subscription Activity**: Verify the subscription hasn't been disposed.
    ```go
    sub := bus.Subscribe("user.action", handler)
    if !sub.IsActive() {
        logutil.Log("Subscription is not active!")
    }
    ```

### Bridge Signal Issues

**Problem**: [`ToSignal()`](action/bus.go:646) doesn't update when actions are dispatched.

**Solutions**:

1.  **Ensure Proper Signal Usage**: Bridge signals must be used within reactive contexts.
    ```go
    // BAD: Using bridge signal outside reactive context
    func MyComponent() g.Node {
        signal := action.ToSignal[string](bus, "user.name")
        value := signal.Get() // Won't be reactive
        return Div(Text(value))
    }

    // GOOD: Using bridge signal in reactive binding
    func MyComponent() g.Node {
        signal := action.ToSignal[string](bus, "user.name")
        return Div(comps.BindText(func() string {
            return signal.Get() // Properly reactive
        }))
    }
    ```

2.  **Check Signal Disposal**: Bridge signals need proper cleanup to prevent memory leaks.
    ```go
    func MyComponent() g.Node {
        signal := action.ToSignal[string](bus, "user.name")
        
        comps.OnCleanup(func() {
            if bridgeSig, ok := signal.(*action.bridgeSignal[string]); ok {
                bridgeSig.Dispose()
            }
        })
        
        return Div(comps.BindText(signal.Get))
    }
    ```

3.  **Verify Action Payload Type**: Ensure the action payload matches the signal type.
    ```go
    // Create typed signal
    signal := action.ToSignal[int](bus, "counter.value")
    
    // Dispatch compatible action
    bus.Dispatch(action.Action[string]{
        Type: "counter.value",
        Payload: "42", // String payload will be converted
    })
    ```

### Query System Issues

**Problem**: [`Ask()`](action/bus.go:829) queries timeout or return errors.

**Solutions**:

1.  **Register Query Handler First**: Query handlers must be registered before queries are sent.
    ```go
    // BAD: Query sent before handler registration
    result, err := bus.Ask("get.user", query)
    bus.HandleQuery("get.user", handler) // Too late

    // GOOD: Handler registered first
    bus.HandleQuery("get.user", func(action action.Action[string]) (any, error) {
        return getUserData(action.Payload), nil
    })
    result, err := bus.Ask("get.user", query)
    ```

2.  **Handle Query Errors Properly**: Query handlers should return meaningful errors.
    ```go
    bus.HandleQuery("get.user", func(action action.Action[string]) (any, error) {
        user, err := database.GetUser(action.Payload)
        if err != nil {
            return nil, fmt.Errorf("failed to get user %s: %w", action.Payload, err)
        }
        return user, nil
    })
    ```

3.  **Use Appropriate Timeouts**: Set reasonable timeouts for query operations.
    ```go
    result, err := bus.Ask("slow.query", query,
        action.WithTimeout(30*time.Second))
    if err == action.ErrTimeout {
        logutil.Log("Query timed out after 30 seconds")
    }
    ```

### Subscription Memory Leaks

**Problem**: Subscriptions are not properly disposed, causing memory leaks.

**Solutions**:

1.  **Use Lifecycle Management**: Always dispose subscriptions when components unmount.
    ```go
    func MyComponent() g.Node {
        var sub action.Subscription
        
        comps.OnMount(func() {
            sub = bus.Subscribe("user.action", func(a action.Action[string]) error {
                logutil.Log("Received action:", a.Payload)
                return nil
            })
        })
        
        comps.OnCleanup(func() {
            if sub != nil {
                sub.Dispose()
            }
        })
        
        return Div(Text("Component"))
    }
    ```

2.  **Use Auto-Disposing Helpers**: Leverage lifecycle utilities for automatic cleanup.
    ```go
    func MyComponent() g.Node {
        // Automatically disposed when component unmounts
        action.OnAction(bus, userAction, func(ctx action.Context, user User) {
            logutil.Log("User action:", user.Name)
        })
        
        return Div(Text("Component"))
    }
    ```

3.  **Check Subscription Status**: Monitor active subscriptions to detect leaks.
    ```go
    // Debug active subscriptions
    stats := action.GetObservabilityStats(bus)
    logutil.Logf("Active subscriptions: %+v", stats)
    ```

### Action Dispatch Errors

**Problem**: Actions fail to dispatch or cause runtime panics.

**Solutions**:

1.  **Handle Dispatch Errors**: Always check dispatch errors in production code.
    ```go
    err := bus.Dispatch(action.Action[string]{
        Type:    "user.update",
        Payload: userData,
    })
    if err != nil {
        logutil.Logf("Failed to dispatch action: %v", err)
        // Handle error appropriately
    }
    ```

2.  **Use Enhanced Error Handling**: Set up comprehensive error handling for the bus.
    ```go
    action.SetEnhancedErrorHandler(bus, func(ctx action.Context, err error, recovered any) {
        logutil.Logf("Action error in %s: %v", ctx.Scope, err)
        if recovered != nil {
            logutil.Logf("Panic recovered: %v", recovered)
        }
    })
    ```

3.  **Validate Action Payloads**: Ensure action payloads are valid before dispatch.
    ```go
    // Create action with validation
    if userData.Name == "" {
        return fmt.Errorf("user name cannot be empty")
    }
    
    err := bus.Dispatch(action.Action[string]{
        Type:    "user.create",
        Payload: userData.Name,
        Meta:    map[string]any{"validated": true},
    })
    ```

## Debugging Techniques

### Using logutil for Debugging

The `logutil` package is your best friend for debugging, as it works correctly in browser environments.

```go
import "github.com/ozanturksever/logutil"

// Log a simple message
logutil.Log("Component OnMount hook fired.")

// Log formatted strings and variables
logutil.Logf("Current count: %d", count.Get())

// Debug signal changes by subscribing to them
count.Subscribe(func(value int) {
    logutil.Logf("Signal 'count' changed to: %d", value)
})
```

### DOM Inspection

```go
// Check if an element exists within OnMount
comps.OnMount(func() {
    el := dom.GetElementByID("my-button")
    if el == nil {
        logutil.Log("Element with ID 'my-button' was not found!")
    } else {
        logutil.Log("Successfully found element:", el.TagName())
    }
})
```

### Effect Debugging

Wrap your effect's logic in logs to see when it triggers.

```go
// Debug an effect's execution
reactivity.CreateEffect(func() {
    logutil.Log("Counter display effect is running...")
    count := counter.Get() // This is the dependency
    displayText.Set(fmt.Sprintf("Count: %d", count))
})
```

### Action System Debugging

The action system provides comprehensive debugging tools through observability features.

#### Enable Development Logging

```go
import "github.com/ozanturksever/uiwgo/action"

// Enable detailed action logging
action.EnableDevLogger(bus, func(entry action.DevLogEntry) {
    logutil.Logf("[ACTION] %s from %s took %v with %d subscribers",
        entry.ActionType, entry.Source, entry.Duration, entry.SubscriberCount)
    if entry.Error != nil {
        logutil.Logf("[ERROR] Action failed: %v", entry.Error)
    }
})
```

#### Debug Ring Buffer

```go
// Enable debug buffer to track recent actions
action.EnableDebugRingBuffer(bus, 100) // Keep last 100 actions per type

// Later, inspect recent actions
entries := action.GetDebugRingBufferEntries(bus, "user.login")
for _, entry := range entries {
    logutil.Logf("Action: %s at %v with payload: %v",
        entry.ActionType, entry.Timestamp, entry.Payload)
}
```

#### Analytics Tap

```go
// Monitor all actions for debugging
tap := action.NewAnalyticsTap(bus, func(event action.AnalyticsEvent) {
    logutil.Logf("Action: %s from %s", event.ActionType, event.Source)
}, action.WithAnalyticsFilter(func(action any) bool {
    // Only log specific action types
    if act, ok := action.(action.Action[string]); ok {
        return strings.HasPrefix(act.Type, "user.")
    }
    return false
}))

// Remember to dispose when done
comps.OnCleanup(func() {
    tap.Dispose()
})
```

#### Bridge Signal Debugging

```go
// Debug bridge signal updates
signal := action.ToSignal[string](bus, "user.name")

// Monitor signal changes
signal.Subscribe(func(value string) {
    logutil.Logf("Signal updated to: %s", value)
})

// Check if bridge signal is active
if bridgeSig, ok := signal.(*action.bridgeSignal[string]); ok {
    logutil.Logf("Bridge signal active: %v", !bridgeSig.disposed)
}
```

#### Query Debugging

```go
// Debug query handling
bus.HandleQuery("get.user", func(action action.Action[string]) (any, error) {
    logutil.Logf("Handling query: %s with payload: %s",
        action.Type, action.Payload)
    
    result, err := getUserData(action.Payload)
    if err != nil {
        logutil.Logf("Query failed: %v", err)
        return nil, err
    }
    
    logutil.Logf("Query succeeded: %+v", result)
    return result, nil
})
```

#### Subscription Debugging

```go
// Debug subscription lifecycle
sub := bus.Subscribe("user.action", func(action action.Action[string]) error {
    logutil.Logf("Received action: %s with payload: %s",
        action.Type, action.Payload)
    return nil
}, action.WithPriority(10)) // Higher priority for debugging

// Monitor subscription status
if !sub.IsActive() {
    logutil.Log("WARNING: Subscription is not active!")
}

// Dispose with logging
comps.OnCleanup(func() {
    logutil.Log("Disposing subscription...")
    sub.Dispose()
})
```

## Performance Issues

### Excessive Re-renders

**Problem**: Components update too frequently

**Solution**: Use memos to optimize computations

```go
// BAD: Expensive computation in effect
reactivity.CreateEffect(func() {
    items := itemList.Get()
    
    // Expensive operation runs on every change
    var total int
    for _, item := range items {
        total += item.Price * item.Quantity
    }
    
    totalDisplay.Set(fmt.Sprintf("$%.2f", float64(total)/100))
})

// GOOD: Use memo for expensive computation
totalPrice := reactivity.CreateMemo(func() int {
    items := itemList.Get()
    var total int
    for _, item := range items {
        total += item.Price * item.Quantity
    }
    return total
})

totalDisplay := totalPrice.Map(func(total int) string {
    return fmt.Sprintf("$%.2f", float64(total)/100)
})
```

### Too Many Effects

**Problem**: Many effects running simultaneously

**Solution**: Batch updates and combine effects

```go
// BAD: Multiple effects for related updates
reactivity.CreateEffect(func() {
    user := currentUser.Get()
    userName.Set(user.Name)
})

reactivity.CreateEffect(func() {
    user := currentUser.Get()
    userEmail.Set(user.Email)
})

reactivity.CreateEffect(func() {
    user := currentUser.Get()
    userAvatar.Set(user.AvatarURL)
})

// GOOD: Single effect for related updates
reactivity.CreateEffect(func() {
    user := currentUser.Get()
    
    // Batch all updates
    reactivity.Batch(func() {
        userName.Set(user.Name)
        userEmail.Set(user.Email)
        userAvatar.Set(user.AvatarURL)
    })
})
```

### Large Lists

**Problem**: Rendering many items is slow

**Solution**: Implement virtual scrolling or pagination

```go
// For large lists, consider pagination
type PaginatedList struct {
    allItems    *reactivity.Signal[[]Item]
    currentPage *reactivity.Signal[int]
    pageSize    int
    
    visibleItems *reactivity.Memo[[]Item]
}

func NewPaginatedList(pageSize int) *PaginatedList {
    pl := &PaginatedList{
        allItems:    reactivity.CreateSignal([]Item{}),
        currentPage: reactivity.CreateSignal(0),
        pageSize:    pageSize,
    }
    
    pl.visibleItems = reactivity.CreateMemo(func() []Item {
        all := pl.allItems.Get()
        page := pl.currentPage.Get()
        
        start := page * pl.pageSize
        end := start + pl.pageSize
        
        if start >= len(all) {
            return []Item{}
        }
        if end > len(all) {
            end = len(all)
        }
        
        return all[start:end]
    })
    
    return pl
}
```

### Action System Performance

**Problem**: High-frequency action dispatching causes performance issues

**Solutions**:

1.  **Use DistinctUntilChanged**: Prevent unnecessary handler executions for duplicate payloads.
    ```go
    // Subscription only triggers if payload actually changes
    bus.Subscribe("ui.update", func(action action.Action[string]) error {
        updateUI(action.Payload)
        return nil
    }, action.WithDistinctUntilChanged())
    ```

2.  **Implement Manual Debouncing**: For frequent events like scrolling or input.
    ```go
    type DebouncedDispatcher struct {
        bus    action.Bus
        timer  *time.Timer
        mu     sync.Mutex
    }

    func (d *DebouncedDispatcher) DispatchDebounced(actionType string, payload string, delay time.Duration) {
        d.mu.Lock()
        defer d.mu.Unlock()
        
        if d.timer != nil {
            d.timer.Stop()
        }
        
        d.timer = time.AfterFunc(delay, func() {
            d.bus.Dispatch(action.Action[string]{
                Type:    actionType,
                Payload: payload,
            })
        })
    }
    ```

3.  **Use Async Dispatch for Non-Critical Actions**: Offload heavy processing from the main thread.
    ```go
    // Dispatch analytics events asynchronously
    err := bus.Dispatch(analyticsAction, action.WithAsync())
    if err != nil {
        logutil.Logf("Failed to dispatch analytics: %v", err)
    }
    ```

4.  **Optimize Subscription Filters**: Use efficient filtering to reduce handler execution.
    ```go
    // Filter actions at subscription level
    bus.Subscribe("user.action", handler, action.WithFilter(func(payload any) bool {
        if user, ok := payload.(User); ok {
            return user.IsActive && user.Role == "admin"
        }
        return false
    }))
    ```

5.  **Monitor Action Performance**: Use observability to identify bottlenecks.
    ```go
    // Track slow actions
    action.EnableDevLogger(bus, func(entry action.DevLogEntry) {
        if entry.Duration > 100*time.Millisecond {
            logutil.Logf("SLOW ACTION: %s took %v with %d subscribers",
                entry.ActionType, entry.Duration, entry.SubscriberCount)
        }
    })
    ```

6.  **Batch Related Actions**: Group multiple related updates into single actions.
    ```go
    // BAD: Multiple individual updates
    bus.Dispatch(action.Action[string]{Type: "user.name", Payload: user.Name})
    bus.Dispatch(action.Action[string]{Type: "user.email", Payload: user.Email})
    bus.Dispatch(action.Action[string]{Type: "user.role", Payload: user.Role})

    // GOOD: Single batched update
    bus.Dispatch(action.Action[User]{Type: "user.update", Payload: user})
    ```

**Note**: Advanced performance features like object pooling and microtask scheduling are only available in standard Go builds (`!js && !wasm`). For WebAssembly builds, focus on application-level optimizations.

## Build & Development Issues

### Dev Server Not Starting

**Problem**: `make run` fails or server doesn't start

**Solutions**:

1. **Check port availability**:
```bash
# Kill processes using port 8080
make kill

# Or check what's using the port
lsof -i :8080
```

2. **Verify Go installation**:
```bash
# Check Go version (need 1.21+)
go version

# Check WASM support
GOOS=js GOARCH=wasm go version
```

3. **Clean and rebuild**:
```bash
make clean
make run counter
```

### Build Failures

**Problem**: WASM compilation fails

**Solutions**:

1. **Check build tags**:
```go
// Make sure build tags are correct
//go:build js && wasm
// +build js,wasm

package main
```

2. **Verify imports**:
```go
// Use WASM-compatible packages
import (
    "honnef.co/go/js/dom/v2" // Good for WASM
    // "os"                   // May not work in WASM
)
```

3. **Check for unsupported features**:
```go
// BAD: File system operations in WASM
file, err := os.Open("data.txt")

// GOOD: Use fetch API or embed data
response, err := http.Get("/api/data")
```

### Test Failures

**Problem**: Browser tests fail or timeout

**Solutions**:

1. **Increase timeout**:
```go
// In test files
chromedpCtx := testhelpers.MustNewChromedpContext(
    testhelpers.ExtendedTimeoutConfig(), // 60s timeout
)
```

2. **Check test dependencies**:
```bash
# Make sure Chrome/Chromium is installed
which google-chrome
which chromium
```

3. **Debug test failures**:
```go
// Use visible browser for debugging
chromedpCtx := testhelpers.MustNewChromedpContext(
    testhelpers.VisibleConfig(), // Shows browser window
)
```

## Browser Compatibility

### WebAssembly Support

**Supported Browsers**:
- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+

**Check Support**:
```javascript
if (typeof WebAssembly === 'object') {
    console.log('WebAssembly is supported');
} else {
    console.error('WebAssembly is not supported');
}
```

### CORS Issues

**Problem**: WASM files fail to load due to CORS

**Solution**: Configure server properly

```go
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers if needed
        w.Header().Set("Access-Control-Allow-Origin", "*")
        
        // Set correct MIME type for WASM
        if strings.HasSuffix(r.URL.Path, ".wasm") {
            w.Header().Set("Content-Type", "application/wasm")
        }
        
        http.ServeFile(w, r, "."+r.URL.Path)
    })
    
    logutil.Log("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

## Frequently Asked Questions

### General Questions

**Q: How is UIwGo different from React/Vue/Angular?**

A: UIwGo is designed for Go developers who want to build web UIs without learning JavaScript. Key differences:
- Server-side rendering with client-side reactivity
- Go-based component logic
- HTML-first approach with data attributes
- Compiled to WebAssembly for browser execution

**Q: Can I use UIwGo with existing JavaScript libraries?**

A: Yes, but with limitations. You can:
- Call JavaScript functions from Go using `syscall/js`
- Use CSS frameworks (Bootstrap, Tailwind, etc.)
- Integrate with simple JavaScript libraries
- However, complex JavaScript frameworks may conflict

**Q: Is UIwGo production-ready?**

A: UIwGo is suitable for:
- Internal tools and dashboards
- Prototypes and MVPs
- Applications where Go expertise is more valuable than JavaScript
- Projects requiring strong typing and Go's ecosystem

Consider maturity and ecosystem size for large-scale public applications.

### Technical Questions

**Q: How do I handle forms and validation?**

A: Use two-way binding and reactive validation:

```go
type Form struct {
    email    *reactivity.Signal[string]
    password *reactivity.Signal[string]
    
    emailError    *reactivity.Memo[string]
    passwordError *reactivity.Memo[string]
    isValid       *reactivity.Memo[bool]
}

func NewForm() *Form {
    f := &Form{
        email:    reactivity.CreateSignal(""),
        password: reactivity.CreateSignal(""),
    }
    
    f.emailError = reactivity.CreateMemo(func() string {
        email := f.email.Get()
        if email == "" {
            return "Email is required"
        }
        if !strings.Contains(email, "@") {
            return "Invalid email format"
        }
        return ""
    })
    
    f.isValid = reactivity.CreateMemo(func() bool {
        return f.emailError.Get() == "" && f.passwordError.Get() == ""
    })
    
    return f
}
```

**Q: How do I make HTTP requests?**

A: Use Go's standard `net/http` package:

```go
func (c *Component) loadData() {
    c.loading.Set(true)
    
    go func() {
        defer c.loading.Set(false)
        
        resp, err := http.Get("/api/data")
        if err != nil {
            c.error.Set(err)
            return
        }
        defer resp.Body.Close()
        
        var data []Item
        if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
            c.error.Set(err)
            return
        }
        
        c.items.Set(data)
    }()
}
```

**Q: How do I handle routing?**

A: UIwGo includes a client-side router:

```go
import "github.com/ozanturksever/uiwgo/router"

func setupRoutes() {
    r := router.New()
    
    r.Route("/", NewHomePage())
    r.Route("/about", NewAboutPage())
    r.Route("/users/:id", NewUserPage())
    
    r.Start()
}
```

**Q: Can I use CSS frameworks?**

A: Yes! UIwGo works with any CSS framework:

```html
<!-- Include in your HTML -->
<link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
```

```go
// Use classes in components
func (c *Button) Render() g.Node {
    return h.Button(
        g.Class("bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"),
        g.Attr("data-click", "handleClick"),
        g.Text("Click me"),
    )
}
```

**Q: How do I optimize bundle size?**

A: Several strategies:

1. **Use build flags to exclude unused code**:
```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm
```

2. **Lazy load components**:
```go
// Only create expensive components when needed
if c.showAdvanced.Get() && c.advancedPanel == nil {
    c.advancedPanel = NewAdvancedPanel()
}
```

3. **Use gzip compression**:
```go
func main() {
    http.Handle("/", gzipHandler(http.FileServer(http.Dir("."))))
}
```

### Performance Questions

**Q: Is UIwGo fast enough for real applications?**

A: Performance depends on your use case:
- **Good for**: Forms, dashboards, admin panels, CRUD apps
- **Consider alternatives for**: Games, real-time graphics, heavy animations
- **Optimization tips**: Use memos, batch updates, implement virtual scrolling for large lists

**Q: How do I profile performance?**

A: Use browser dev tools and Go profiling:

```go
// Add timing to effects
reactivity.CreateEffect(func() {
    start := time.Now()
    defer func() {
        logutil.Logf("Effect took: %v", time.Since(start))
    }()
    
    // Effect logic here
})
```

**Q: How do I test action system components?**

A: Use the comprehensive testing utilities provided in [`action/TESTING.md`](action/TESTING.md):

```go
import "github.com/ozanturksever/uiwgo/action"

func TestMyActionFlow(t *testing.T) {
    // Create isolated test environment
    testBus := action.NewTestBus()
    bus := testBus.Bus()
    clock := testBus.Clock()
    
    // Set up mock subscriber
    mock := action.NewMockSubscriber[string]()
    bus.Subscribe("user.action", mock.Handler())
    
    // Dispatch test action
    bus.Dispatch(action.Action[string]{
        Type:    "user.action",
        Payload: "test-data",
    })
    
    // Verify results
    received := mock.GetReceivedActions()
    if len(received) != 1 {
        t.Errorf("Expected 1 action, got %d", len(received))
    }
    
    // Test with fake clock for deterministic timing
    future := action.NewTestFuture[string](testBus)
    go func() {
        time.Sleep(10 * time.Millisecond)
        future.Resolve("success")
    }()
    
    clock.Advance(20 * time.Millisecond)
    result, err := future.Await()
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}
```

**Q: How do I monitor action system performance in production?**

A: Use the observability features for production monitoring:

```go
// Enable observability for production monitoring
action.EnableDevLogger(bus, func(entry action.DevLogEntry) {
    // Log to your monitoring system
    if entry.Error != nil {
        metrics.Counter("action.errors").Inc()
        logger.Error("Action failed",
            "type", entry.ActionType,
            "error", entry.Error,
            "duration", entry.Duration)
    }
    
    if entry.Duration > 100*time.Millisecond {
        metrics.Counter("action.slow").Inc()
        logger.Warn("Slow action detected",
            "type", entry.ActionType,
            "duration", entry.Duration,
            "subscribers", entry.SubscriberCount)
    }
})

// Monitor action system health
stats := action.GetObservabilityStats(bus)
logutil.Logf("Action system stats: %+v", stats)

// Set up analytics for user behavior tracking
tap := action.NewAnalyticsTap(bus, func(event action.AnalyticsEvent) {
    // Send to analytics service
    analytics.Track(event.ActionType, map[string]any{
        "source":    event.Source,
        "timestamp": event.Timestamp,
        "meta":      event.Meta,
    })
}, action.WithAnalyticsFilter(func(action any) bool {
    // Only track user-facing actions
    if act, ok := action.(action.Action[string]); ok {
        return strings.HasPrefix(act.Type, "user.") ||
               strings.HasPrefix(act.Type, "ui.")
    }
    return false
}))
```

**Q: How do I debug subscription lifecycle issues?**

A: Use lifecycle debugging and proper cleanup patterns:

```go
func MyComponent() g.Node {
    logutil.Log("Component created")
    
    comps.OnMount(func() {
        logutil.Log("Component mounted, setting up subscriptions")
        
        // Use lifecycle-aware subscription helpers
        action.OnAction(bus, userAction, func(ctx action.Context, user User) {
            logutil.Logf("Received user action: %s", user.Name)
        })
        
        // Or manual subscription with proper cleanup
        sub := bus.Subscribe("debug.action", func(a action.Action[string]) error {
            logutil.Logf("Debug action: %s", a.Payload)
            return nil
        })
        
        comps.OnCleanup(func() {
            logutil.Log("Component cleaning up, disposing subscription")
            if err := sub.Dispose(); err != nil {
                logutil.Logf("Error disposing subscription: %v", err)
            }
        })
    })
    
    return Div(Text("Component"))
}

// Monitor subscription leaks
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := action.GetObservabilityStats(bus)
        if stats.DebugBufferActionTypes > 50 {
            logutil.Log("WARNING: High number of action types, possible subscription leak")
        }
    }
}()
```

---

## Getting Help

### Resources
- **Documentation**: Start with [Getting Started](./getting-started.md)
- **Examples**: Check the `examples/` directory
- **API Reference**: See [Core APIs](./api/core-apis.md)
- **Action System**: Review [Action System Testing](action/TESTING.md) and [Performance Guide](action/PERFORMANCE.md)
- **Action System**: Check [API Reference](action/API_REFERENCE.md) for comprehensive action system documentation

### Debugging Steps
1. Check browser console for errors
2. Verify component lifecycle (Render → Mount → Attach)
3. Test reactivity with simple examples
4. Use `logutil` for debugging
5. Check data attribute bindings
6. **Action System**: Enable [`DevLogger`](action/observability.go:120) and [`DebugRingBuffer`](action/observability.go:143) for action flow visibility
7. **Action System**: Verify bus instances match between dispatch and subscription
8. **Action System**: Check subscription disposal and lifecycle management

### Common Solutions
- **Nothing renders**: Check if `Attach()` is called
- **No reactivity**: Verify signal dependencies in effects/memos
- **Events don't work**: Check data attribute names and bindings
- **Memory leaks**: Implement `Cleanup()` and dispose effects
- **Build fails**: Check Go version and WASM support
- **Actions not received**: Verify action type matching and bus instance consistency
- **Bridge signals not updating**: Ensure signals are used within reactive contexts
- **Query timeouts**: Register handlers before sending queries and use appropriate timeouts
- **Subscription leaks**: Use [`OnAction()`](action/lifecycle.go:93) or manual disposal in [`OnCleanup`](action/lifecycle.go:86)

Next: Explore [Forms & Events](./guides/forms-events.md) or learn about [Performance Optimization](./guides/performance-optimization.md).

For action system specific guidance, see the [Action System Testing Guide](action/TESTING.md) and [Performance Documentation](action/PERFORMANCE.md).