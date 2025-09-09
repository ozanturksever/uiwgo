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
- [ ] Component implements `Render()` and `Attach()` methods
- [ ] `Render()` returns valid HTML with data attributes
- [ ] `Attach()` binds all data attributes to signals/memos
- [ ] Component is properly mounted with `comps.Mount()`

### ✅ Reactivity
- [ ] Signals are created with `reactivity.NewSignal()`
- [ ] Memos access signals with `.Get()` inside the function
- [ ] Effects track dependencies by calling `.Get()` on signals
- [ ] No infinite loops in effects (signal updates triggering themselves)

### ✅ DOM Binding
- [ ] Data attributes match binding selectors exactly
- [ ] HTML elements exist when `Attach()` is called
- [ ] Event handlers are bound to existing elements
- [ ] No typos in data attribute names

### ✅ Browser Console
- [ ] Check for JavaScript errors in browser console
- [ ] Verify WASM module loads successfully
- [ ] Look for UIwGo-specific error messages

## Common Issues

### Component Not Rendering

**Problem**: Component appears blank or shows default content

**Symptoms**:
```html
<!-- Expected: <div>Hello, World!</div> -->
<!-- Actual: <div data-text="greeting">Loading...</div> -->
```

**Solutions**:

1. **Check if `Attach()` is called**:
```go
// BAD: Only renders, doesn't attach
func main() {
    comp := NewMyComponent()
    html := comp.Render()
    // Missing: comp.Attach()
}

// GOOD: Proper mounting
func main() {
    comp := NewMyComponent()
    comps.Mount("app", comp) // Calls both Render() and Attach()
}
```

2. **Verify data attribute binding**:
```go
// Check that selector matches HTML
func (c *MyComponent) Render() g.Node {
    return h.Div(
        g.Attr("data-text", "greeting"), // selector: "greeting"
        g.Text("Loading..."),
    )
}

func (c *MyComponent) Attach() {
    c.BindText("greeting", c.message) // Must match exactly
}
```

3. **Ensure signal has a value**:
```go
// Check signal initialization
message := reactivity.NewSignal("") // Empty string won't show
message := reactivity.NewSignal("Hello, World!") // Will show
```

### Reactivity Not Working

**Problem**: UI doesn't update when signals change

**Symptoms**:
```go
count.Set(5) // Signal updates
// But UI still shows old value
```

**Solutions**:

1. **Check effect dependencies**:
```go
// BAD: Effect doesn't track signal
reactivity.NewEffect(func() {
    value := 42 // Static value, no dependencies
    logutil.Logf("Value: %d", value)
})

// GOOD: Effect tracks signal
reactivity.NewEffect(func() {
    value := count.Get() // Tracks count signal
    logutil.Logf("Count: %d", value)
})
```

2. **Verify memo dependencies**:
```go
// BAD: Memo doesn't access signals
displayText := reactivity.NewMemo(func() string {
    return "Static text" // No signal access
})

// GOOD: Memo accesses signals
displayText := reactivity.NewMemo(func() string {
    return "Count: " + strconv.Itoa(count.Get()) // Tracks count
})
```

3. **Check for disposed effects**:
```go
effect := reactivity.NewEffect(func() {
    logutil.Logf("Count: %d", count.Get())
})

effect.Dispose() // Effect stops working after this
count.Set(10)    // Won't trigger the effect
```

### Event Handlers Not Working

**Problem**: Clicks and other events don't trigger handlers

**Solutions**:

1. **Check data attribute syntax**:
```html
<!-- BAD: Wrong attribute name -->
<button data-onclick="increment">+</button>

<!-- GOOD: Correct attribute name -->
<button data-click="increment">+</button>
```

2. **Verify handler binding**:
```go
func (c *Counter) Attach() {
    // Make sure selector matches data attribute
    c.BindClick("increment", c.increment) // data-click="increment"
}

func (c *Counter) increment() {
    c.count.Update(func(n int) int { return n + 1 })
}
```

3. **Check element exists**:
```go
// Debug: Check if element is found
func (c *Counter) Attach() {
    element := dom.QuerySelector(`[data-click="increment"]`)
    if element == nil {
        logutil.Log("ERROR: Increment button not found!")
    }
    
    c.BindClick("increment", c.increment)
}
```

### Memory Leaks

**Problem**: Application becomes slow over time

**Symptoms**:
- Increasing memory usage
- Slower response times
- Browser becomes unresponsive

**Solutions**:

1. **Implement proper cleanup**:
```go
type Component struct {
    effects []reactivity.Effect
    timer   *time.Timer
}

func (c *Component) setupEffects() {
    effect := reactivity.NewEffect(func() {
        // Effect logic
    })
    c.effects = append(c.effects, effect)
}

func (c *Component) Cleanup() {
    // Dispose all effects
    for _, effect := range c.effects {
        effect.Dispose()
    }
    
    // Stop timers
    if c.timer != nil {
        c.timer.Stop()
    }
}
```

2. **Avoid infinite effect loops**:
```go
// BAD: Infinite loop
reactivity.NewEffect(func() {
    count := counter.Get()
    counter.Set(count + 1) // Triggers effect again!
})

// GOOD: Use different signals
reactivity.NewEffect(func() {
    count := counter.Get()
    displayText.Set(fmt.Sprintf("Count: %d", count))
})
```

### WASM Loading Issues

**Problem**: Application doesn't start or shows WASM errors

**Solutions**:

1. **Check WASM file path**:
```html
<!-- Make sure path is correct -->
<script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
        .then((result) => {
            go.run(result.instance);
        })
        .catch((err) => {
            console.error("WASM loading failed:", err);
        });
</script>
```

2. **Verify build process**:
```bash
# Make sure WASM is built correctly
make build counter
# Check if main.wasm exists
ls examples/counter/main.wasm
```

3. **Check server MIME types**:
```go
// Ensure server serves .wasm files correctly
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, ".wasm") {
            w.Header().Set("Content-Type", "application/wasm")
        }
        http.ServeFile(w, r, "."+r.URL.Path)
    })
}
```

## Debugging Techniques

### Using logutil for Debugging

```go
// Use logutil instead of fmt.Println for WASM compatibility
import "github.com/ozanturksever/uiwgo/internal/logutil"

// Basic logging
logutil.Log("Component attached")
logutil.Logf("Count value: %d", count.Get())

// Debug signal changes
count.Subscribe(func(value int) {
    logutil.Logf("Count changed: %d", value)
})

// Debug effect execution
reactivity.NewEffect(func() {
    value := count.Get()
    logutil.Logf("Effect triggered with count: %d", value)
})
```

### Debugging Reactivity

```go
// Create a debug wrapper for signals
func DebugSignal[T any](name string, initial T) *reactivity.Signal[T] {
    signal := reactivity.NewSignal(initial)
    
    // Log all changes
    signal.Subscribe(func(value T) {
        logutil.Logf("Signal %s changed to: %v", name, value)
    })
    
    return signal
}

// Usage
count := DebugSignal("count", 0)
name := DebugSignal("name", "")
```

### DOM Inspection

```go
// Check if elements exist
func debugElement(selector string) {
    element := dom.QuerySelector(selector)
    if element == nil {
        logutil.Logf("Element not found: %s", selector)
    } else {
        logutil.Logf("Element found: %s", element.TagName())
    }
}

// Debug data attributes
func debugDataAttributes() {
    elements := dom.QuerySelectorAll("[data-text]")
    logutil.Logf("Found %d elements with data-text", len(elements))
    
    for i, el := range elements {
        attr := el.GetAttribute("data-text")
        logutil.Logf("Element %d: data-text=%s", i, attr)
    }
}
```

### Effect Debugging

```go
// Debug effect dependencies
func DebugEffect(name string, fn func()) reactivity.Effect {
    logutil.Logf("Creating effect: %s", name)
    
    return reactivity.NewEffect(func() {
        logutil.Logf("Effect %s triggered", name)
        fn()
        logutil.Logf("Effect %s completed", name)
    })
}

// Usage
DebugEffect("counter-display", func() {
    count := counter.Get()
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