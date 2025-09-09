# Troubleshooting & FAQ

This guide helps you diagnose and solve common issues when developing with UIwGo. It covers debugging techniques, performance problems, common pitfalls, and frequently asked questions.

## Table of Contents

- [Quick Debugging Checklist](#quick-debugging-checklist)
- [Common Issues](#common-issues)
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
- [ ] Signals are created with `reactivity.NewSignal()`.
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
    message := reactivity.NewSignal("Hello, World!")

    // POTENTIAL ISSUE (will render nothing)
    message := reactivity.NewSignal("")
    ```

### Reactivity Not Working

**Problem**: The UI doesn't update when a signal's value changes.

**Solutions**:

1.  **Check Dependencies**: Ensure your Memos and Effects are calling `.Get()` on the signals they should be tracking.
    ```go
    // BAD: Effect doesn't track the 'count' signal
    reactivity.NewEffect(func() {
        logutil.Log("This runs only once.")
    })

    // GOOD: Effect tracks the 'count' signal
    reactivity.NewEffect(func() {
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
    reactivity.NewEffect(func() {
        count := counter.Get()
        counter.Set(count + 1) // This triggers the effect again!
    })

    // GOOD: Use a separate signal for the output
    reactivity.NewEffect(func() {
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
reactivity.NewEffect(func() {
    logutil.Log("Counter display effect is running...")
    count := counter.Get() // This is the dependency
    displayText.Set(fmt.Sprintf("Count: %d", count))
})
```

## Performance Issues

### Excessive Re-renders

**Problem**: Components update too frequently

**Solution**: Use memos to optimize computations

```go
// BAD: Expensive computation in effect
reactivity.NewEffect(func() {
    items := itemList.Get()
    
    // Expensive operation runs on every change
    var total int
    for _, item := range items {
        total += item.Price * item.Quantity
    }
    
    totalDisplay.Set(fmt.Sprintf("$%.2f", float64(total)/100))
})

// GOOD: Use memo for expensive computation
totalPrice := reactivity.NewMemo(func() int {
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
reactivity.NewEffect(func() {
    user := currentUser.Get()
    userName.Set(user.Name)
})

reactivity.NewEffect(func() {
    user := currentUser.Get()
    userEmail.Set(user.Email)
})

reactivity.NewEffect(func() {
    user := currentUser.Get()
    userAvatar.Set(user.AvatarURL)
})

// GOOD: Single effect for related updates
reactivity.NewEffect(func() {
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
        allItems:    reactivity.NewSignal([]Item{}),
        currentPage: reactivity.NewSignal(0),
        pageSize:    pageSize,
    }
    
    pl.visibleItems = reactivity.NewMemo(func() []Item {
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
        email:    reactivity.NewSignal(""),
        password: reactivity.NewSignal(""),
    }
    
    f.emailError = reactivity.NewMemo(func() string {
        email := f.email.Get()
        if email == "" {
            return "Email is required"
        }
        if !strings.Contains(email, "@") {
            return "Invalid email format"
        }
        return ""
    })
    
    f.isValid = reactivity.NewMemo(func() bool {
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
reactivity.NewEffect(func() {
    start := time.Now()
    defer func() {
        logutil.Logf("Effect took: %v", time.Since(start))
    }()
    
    // Effect logic here
})
```

---

## Getting Help

### Resources
- **Documentation**: Start with [Getting Started](./getting-started.md)
- **Examples**: Check the `examples/` directory
- **API Reference**: See [Core APIs](./api/core-apis.md)

### Debugging Steps
1. Check browser console for errors
2. Verify component lifecycle (Render → Mount → Attach)
3. Test reactivity with simple examples
4. Use `logutil` for debugging
5. Check data attribute bindings

### Common Solutions
- **Nothing renders**: Check if `Attach()` is called
- **No reactivity**: Verify signal dependencies in effects/memos
- **Events don't work**: Check data attribute names and bindings
- **Memory leaks**: Implement `Cleanup()` and dispose effects
- **Build fails**: Check Go version and WASM support

Next: Explore [Control Flow](./guides/control-flow.md) or [Forms & Events](./guides/forms-events.md).