# Inline Events: Troubleshooting Guide

This guide helps you diagnose and resolve common issues when working with inline event binding in uiwgo applications.

## Table of Contents

- [Common Issues](#common-issues)
- [Debugging Techniques](#debugging-techniques)
- [Error Messages](#error-messages)
- [Performance Issues](#performance-issues)
- [Browser Compatibility](#browser-compatibility)
- [Testing Problems](#testing-problems)
- [Migration Issues](#migration-issues)

## Common Issues

### 1. Event Handler Not Firing

**Symptoms:**
- Clicking/interacting with elements does nothing
- No console errors
- Handler function is never called

**Possible Causes & Solutions:**

#### A. Missing `dom.AttachInlineDelegates()` Call

```go
// ❌ Problem: Delegates not attached
func main() {
    comps.Mount("app", MyApp())
    // Missing: dom.AttachInlineDelegates()
}

// ✅ Solution: Always call AttachInlineDelegates
func main() {
    comps.Mount("app", MyApp())
    dom.AttachInlineDelegates() // Required!
}
```

#### B. Element Not in DOM When Handler Attached

```go
// ❌ Problem: Handler attached before element exists
func BadComponent() g.Node {
    // This runs before the button is in the DOM
    comps.OnMount(func() {
        btn := dom.GetElementByID("my-btn")
        // btn might be null here
    })
    
    return Button(ID("my-btn"), Text("Click"))
}

// ✅ Solution: Use inline events (recommended)
func GoodComponent() g.Node {
    return Button(
        Text("Click"),
        dom.OnClickInline(func(el dom.Element) {
            // This always works
        }),
    )
}
```

#### C. Event Bubbling Stopped by Parent

```go
// ❌ Problem: Parent stops event propagation
Div(
    dom.OnClickInline(func(el dom.Element) {
        // This stops all child click events
        js.Global().Get("event").Call("stopPropagation")
    }),
    Button(
        Text("Child Button"),
        dom.OnClickInline(func(el dom.Element) {
            // This will never fire!
            logutil.Log("Child clicked")
        }),
    ),
)

// ✅ Solution: Be selective about stopping propagation
Div(
    dom.OnClickInline(func(el dom.Element) {
        // Only stop propagation when necessary
        if shouldStopPropagation() {
            js.Global().Get("event").Call("stopPropagation")
        }
    }),
    Button(
        Text("Child Button"),
        dom.OnClickInline(func(el dom.Element) {
            logutil.Log("Child clicked")
        }),
    ),
)
```

### 2. Handler Fires Multiple Times

**Symptoms:**
- Single click triggers handler multiple times
- Duplicate log messages
- Unexpected behavior

**Possible Causes & Solutions:**

#### A. Multiple `AttachInlineDelegates()` Calls

```go
// ❌ Problem: Multiple delegate attachments
func main() {
    comps.Mount("app", MyApp())
    dom.AttachInlineDelegates()
    
    // Later in code...
    dom.AttachInlineDelegates() // This creates duplicate listeners!
}

// ✅ Solution: Call AttachInlineDelegates only once
func main() {
    comps.Mount("app", MyApp())
    dom.AttachInlineDelegates() // Only once!
}
```

#### B. Event Bubbling Through Multiple Handlers

```go
// ❌ Problem: Nested elements with same event type
Div(
    dom.OnClickInline(func(el dom.Element) {
        logutil.Log("Outer clicked")
    }),
    Div(
        dom.OnClickInline(func(el dom.Element) {
            logutil.Log("Inner clicked")
            // Both handlers fire when inner div is clicked!
        }),
    ),
)

// ✅ Solution: Stop propagation when needed
Div(
    dom.OnClickInline(func(el dom.Element) {
        logutil.Log("Outer clicked")
    }),
    Div(
        dom.OnClickInline(func(el dom.Element) {
            logutil.Log("Inner clicked")
            js.Global().Get("event").Call("stopPropagation")
        }),
    ),
)
```

### 3. Handler Captures Wrong Values

**Symptoms:**
- Handler uses stale/incorrect values
- Loop variables have unexpected values
- Closure captures wrong state

**Possible Causes & Solutions:**

#### A. Loop Variable Capture Issue

```go
// ❌ Problem: All handlers capture the same loop variable
for i, item := range items {
    buttons = append(buttons, Button(
        Text(item.Name),
        dom.OnClickInline(func(el dom.Element) {
            // All handlers will use the final value of 'i' and 'item'!
            logutil.Log("Clicked item", i, item.Name)
        }),
    ))
}

// ✅ Solution 1: Capture variables in closure
for i, item := range items {
    i, item := i, item // Create local copies
    buttons = append(buttons, Button(
        Text(item.Name),
        dom.OnClickInline(func(el dom.Element) {
            logutil.Log("Clicked item", i, item.Name)
        }),
    ))
}

// ✅ Solution 2: Use function parameters
for i, item := range items {
    buttons = append(buttons, createButton(i, item))
}

func createButton(index int, item Item) g.Node {
    return Button(
        Text(item.Name),
        dom.OnClickInline(func(el dom.Element) {
            logutil.Log("Clicked item", index, item.Name)
        }),
    )
}

// ✅ Solution 3: Use BindList (recommended)
comps.BindList(itemsSignal, func(item Item) g.Node {
    return Button(
        Text(item.Name),
        dom.OnClickInline(func(el dom.Element) {
            // 'item' is correctly captured per iteration
            logutil.Log("Clicked", item.Name)
        }),
    )
})
```

#### B. Stale Signal Values

```go
// ❌ Problem: Capturing signal value instead of signal
currentValue := mySignal.Get() // Captures current value
Button(
    Text("Click"),
    dom.OnClickInline(func(el dom.Element) {
        // Uses stale value, not current signal value!
        logutil.Log("Value:", currentValue)
    }),
)

// ✅ Solution: Capture the signal, not its value
Button(
    Text("Click"),
    dom.OnClickInline(func(el dom.Element) {
        // Gets current value each time
        logutil.Log("Value:", mySignal.Get())
    }),
)
```

### 4. Memory Leaks

**Symptoms:**
- Memory usage grows over time
- Browser becomes slow
- Handler registry grows indefinitely

**Possible Causes & Solutions:**

#### A. Handlers Not Cleaned Up

```go
// ❌ Problem: Creating new handlers on every render
func LeakyComponent() g.Node {
    // This creates a new handler every time the component renders
    return Button(
        Text("Click"),
        dom.OnClickInline(func(el dom.Element) {
            // New handler instance each render
            handleClick()
        }),
    )
}

// ✅ Solution: Use stable handler references
var stableHandler = func(el dom.Element) {
    handleClick()
}

func NonLeakyComponent() g.Node {
    return Button(
        Text("Click"),
        dom.OnClickInline(stableHandler), // Reuses same handler
    )
}
```

#### B. Large Object Capture

```go
// ❌ Problem: Handler captures large objects
largeData := loadLargeDataset() // 10MB
Button(
    Text("Process"),
    dom.OnClickInline(func(el dom.Element) {
        // Keeps entire 10MB in memory!
        processData(largeData)
    }),
)

// ✅ Solution: Extract only needed data
processID := largeData.ID
Button(
    Text("Process"),
    dom.OnClickInline(func(el dom.Element) {
        // Only captures small ID
        processDataByID(processID)
    }),
)
```

## Debugging Techniques

### 1. Enable Debug Logging

```go
// Add debug logging to handlers
dom.OnClickInline(func(el dom.Element) {
    logutil.Log("Handler fired for element:", el.Get("tagName").String())
    logutil.Log("Element ID:", el.Get("id").String())
    logutil.Log("Handler ID:", el.Get("data-click-inline").String())
    
    // Your actual handler logic
    handleClick()
})
```

### 2. Inspect Handler Registry

```javascript
// Run in browser console to inspect handlers
console.log('Click handlers:', window.inlineClickHandlers);
console.log('Input handlers:', window.inlineInputHandlers);
console.log('Change handlers:', window.inlineChangeHandlers);
console.log('Keydown handlers:', window.inlineKeydownHandlers);

// Count total handlers
let totalHandlers = 0;
if (window.inlineClickHandlers) totalHandlers += Object.keys(window.inlineClickHandlers).length;
if (window.inlineInputHandlers) totalHandlers += Object.keys(window.inlineInputHandlers).length;
console.log('Total handlers:', totalHandlers);
```

### 3. Check DOM Attributes

```javascript
// Inspect element attributes in browser console
const element = document.querySelector('#my-button');
console.log('Click handler ID:', element.getAttribute('data-click-inline'));
console.log('All data attributes:', element.dataset);

// Find all elements with inline handlers
const elementsWithHandlers = document.querySelectorAll('[data-click-inline], [data-input-inline], [data-change-inline]');
console.log('Elements with handlers:', elementsWithHandlers.length);
```

### 4. Monitor Event Flow

```go
// Add event flow monitoring
dom.OnClickInline(func(el dom.Element) {
    logutil.Log("=== Event Debug Info ===")
    logutil.Log("Target:", el.Get("tagName").String())
    logutil.Log("ID:", el.Get("id").String())
    logutil.Log("Classes:", el.Get("className").String())
    
    // Check if event object is available
    if event := js.Global().Get("event"); !event.IsUndefined() {
        logutil.Log("Event type:", event.Get("type").String())
        logutil.Log("Event target:", event.Get("target").Get("tagName").String())
        logutil.Log("Current target:", event.Get("currentTarget").Get("tagName").String())
    }
    
    handleClick()
})
```

## Error Messages

### "Cannot read property 'addEventListener' of null"

**Cause:** Trying to attach event listeners before DOM is ready.

**Solution:**
```go
// ✅ Ensure DOM is ready
func main() {
    // Wait for DOM
    js.Global().Get("document").Call("addEventListener", "DOMContentLoaded", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        comps.Mount("app", MyApp())
        dom.AttachInlineDelegates()
        return nil
    }))
}
```

### "Handler function is not defined"

**Cause:** Handler registry corruption or timing issue.

**Solution:**
```go
// ✅ Ensure proper initialization order
func main() {
    // 1. Mount components first
    comps.Mount("app", MyApp())
    
    // 2. Then attach delegates
    dom.AttachInlineDelegates()
    
    // 3. Verify handlers are registered
    js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        logutil.Log("Handlers registered:", js.Global().Get("inlineClickHandlers").Length())
        return nil
    }), 100)
}
```

### "Maximum call stack size exceeded"

**Cause:** Infinite recursion in event handlers.

**Solution:**
```go
// ❌ Problem: Recursive handler calls
dom.OnClickInline(func(el dom.Element) {
    // This might trigger the same handler again!
    el.Call("click")
})

// ✅ Solution: Avoid triggering same events
dom.OnClickInline(func(el dom.Element) {
    // Use flags to prevent recursion
    if processing {
        return
    }
    processing = true
    defer func() { processing = false }()
    
    // Safe handler logic
    handleClick()
})
```

## Performance Issues

### Slow Event Handling

**Symptoms:**
- Delayed response to clicks/inputs
- UI freezes during interactions
- High CPU usage

**Debugging:**
```go
// Add performance monitoring
dom.OnClickInline(func(el dom.Element) {
    start := time.Now()
    
    handleClick()
    
    duration := time.Since(start)
    if duration > 10*time.Millisecond {
        logutil.Logf("Slow handler: %v", duration)
    }
})
```

**Solutions:**
```go
// ✅ Defer heavy work
dom.OnClickInline(func(el dom.Element) {
    go func() {
        heavyComputation()
    }()
})

// ✅ Use requestAnimationFrame for DOM updates
dom.OnClickInline(func(el dom.Element) {
    js.Global().Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        updateDOM()
        return nil
    }))
})
```

### Memory Growth

**Debugging:**
```javascript
// Monitor memory usage in browser console
setInterval(() => {
    const used = performance.memory.usedJSHeapSize / 1024 / 1024;
    console.log(`Memory usage: ${used.toFixed(2)} MB`);
}, 5000);
```

**Solutions:**
```go
// ✅ Periodic cleanup for long-running apps
func periodicCleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for range ticker.C {
            // Force garbage collection
            runtime.GC()
            
            // Optionally reset handlers (use sparingly)
            if shouldResetHandlers() {
                dom.AttachInlineDelegates()
            }
        }
    }()
}
```

## Browser Compatibility

### Internet Explorer Issues

**Problem:** Event delegation not working in IE.

**Solution:**
```go
// Add IE compatibility checks
func checkBrowserSupport() {
    userAgent := js.Global().Get("navigator").Get("userAgent").String()
    if strings.Contains(userAgent, "MSIE") || strings.Contains(userAgent, "Trident") {
        logutil.Log("Warning: Limited IE support for inline events")
        // Consider fallback to traditional binding
    }
}
```

### Mobile Safari Issues

**Problem:** Touch events not firing.

**Solution:**
```go
// Add touch event support
dom.OnClickInline(func(el dom.Element) {
    // Handle both click and touch
    handleInteraction()
})

// Also consider adding touch-specific handlers
func addTouchSupport(element g.Node) g.Node {
    return element.With(
        Style("cursor: pointer; -webkit-tap-highlight-color: transparent;"),
    )
}
```

## Testing Problems

### Handlers Not Firing in Tests

**Problem:** Event simulation doesn't trigger inline handlers.

**Solution:**
```go
// ✅ Ensure delegates are attached in tests
func TestMyComponent(t *testing.T) {
    // Setup
    server := devserver.NewServer("my_component", "localhost:0")
    if err := server.Start(); err != nil {
        t.Fatalf("Failed to start server: %v", err)
    }
    defer server.Stop()
    
    chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
    defer chromedpCtx.Cancel()
    
    err := chromedp.Run(chromedpCtx.Ctx,
        chromedp.Navigate(server.URL()),
        chromedp.WaitVisible("body"),
        
        // Wait for inline delegates to be attached
        chromedp.Sleep(100*time.Millisecond),
        
        // Now test interactions
        chromedp.Click("#my-button"),
        chromedp.WaitVisible("#result"),
    )
    
    if err != nil {
        t.Fatalf("Test failed: %v", err)
    }
}
```

### Race Conditions in Tests

**Problem:** Tests fail intermittently due to timing.

**Solution:**
```go
// ✅ Add proper waits
func TestWithProperWaits(t *testing.T) {
    err := chromedp.Run(ctx,
        chromedp.Navigate(url),
        
        // Wait for WASM to initialize
        testhelpers.Actions.WaitForWASMInit(),
        
        // Wait for specific element
        chromedp.WaitVisible("#my-button"),
        
        // Wait for handlers to be attached
        chromedp.Evaluate(`
            new Promise(resolve => {
                const check = () => {
                    if (window.inlineClickHandlers && Object.keys(window.inlineClickHandlers).length > 0) {
                        resolve(true);
                    } else {
                        setTimeout(check, 10);
                    }
                };
                check();
            })
        `, nil),
        
        // Now safe to test
        chromedp.Click("#my-button"),
    )
}
```

## Migration Issues

### Converting from Traditional Binding

**Problem:** Existing code uses traditional event binding.

**Migration Strategy:**
```go
// Before: Traditional binding
func OldComponent() g.Node {
    comps.OnMount(func() {
        btn := dom.GetElementByID("btn")
        dom.BindClickToCallback(btn, handleClick)
    })
    
    return Button(ID("btn"), Text("Click"))
}

// After: Inline events (step-by-step migration)
func NewComponent() g.Node {
    return Button(
        ID("btn"), // Keep ID for now (can remove later)
        Text("Click"),
        dom.OnClickInline(func(el dom.Element) {
            handleClick() // Same handler function
        }),
    )
    // Remove OnMount hook entirely
}
```

### Gradual Migration

```go
// ✅ Migrate one component at a time
func MixedComponent() g.Node {
    return Div(
        // New: Inline events
        Button(
            Text("New Style"),
            dom.OnClickInline(func(el dom.Element) {
                handleNewStyle()
            }),
        ),
        
        // Old: Traditional binding (still works)
        Button(
            ID("old-btn"),
            Text("Old Style"),
        ),
    )
}

// Keep traditional binding for now
func init() {
    comps.OnMount(func() {
        if btn := dom.GetElementByID("old-btn"); !btn.IsNull() {
            dom.BindClickToCallback(btn, handleOldStyle)
        }
    })
}
```

This troubleshooting guide covers the most common issues you'll encounter when working with inline events. Remember to check the browser console for JavaScript errors and use the debugging techniques provided to diagnose problems effectively.