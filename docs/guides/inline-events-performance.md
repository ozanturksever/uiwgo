# Inline Events: Performance Considerations and Edge Cases

This guide covers performance optimization techniques, edge cases, and advanced considerations when using inline event binding in uiwgo applications.

## Table of Contents

- [Performance Overview](#performance-overview)
- [Memory Management](#memory-management)
- [Event Delegation Internals](#event-delegation-internals)
- [Performance Best Practices](#performance-best-practices)
- [Edge Cases and Solutions](#edge-cases-and-solutions)
- [Debugging and Profiling](#debugging-and-profiling)
- [Comparison with Traditional Binding](#comparison-with-traditional-binding)

## Performance Overview

### How Inline Events Work

Inline events in uiwgo use **event delegation** for optimal performance:

1. **Single Global Listeners**: Instead of attaching individual event listeners to each element, uiwgo attaches one listener per event type to the document root
2. **Event Bubbling**: Events bubble up from the target element to the document, where the global listener catches them
3. **Handler Registry**: Each inline event handler is stored in a global registry with a unique ID
4. **Attribute Matching**: The global listener matches the event target's data attributes to find the correct handler

```go
// This creates only ONE document-level click listener
// regardless of how many buttons you have
Button(
    Text("Button 1"),
    dom.OnClickInline(func(el dom.Element) { /* handler 1 */ }),
)
Button(
    Text("Button 2"),
    dom.OnClickInline(func(el dom.Element) { /* handler 2 */ }),
)
// ... 1000 more buttons
```

### Performance Benefits

1. **Constant Memory Usage**: Memory usage for event listeners doesn't scale with the number of elements
2. **Fast DOM Manipulation**: Adding/removing elements doesn't require listener management
3. **Reduced Browser Overhead**: Fewer actual DOM event listeners means less browser memory and processing
4. **Better Garbage Collection**: No need to manually clean up individual listeners

## Memory Management

### Handler Lifecycle

```go
// Handler registration happens during element creation
Button(
    Text("Click me"),
    dom.OnClickInline(func(el dom.Element) {
        // This function is stored in the global registry
        logutil.Log("Clicked!")
    }),
)

// Handlers are automatically cleaned up when:
// 1. The component is unmounted
// 2. dom.AttachInlineDelegates() is called again (full reset)
// 3. The page is refreshed
```

### Memory Optimization Strategies

#### 1. Avoid Capturing Large Objects

```go
// ❌ Bad: Captures entire large object
largeData := make([]byte, 1024*1024) // 1MB
Button(
    Text("Process"),
    dom.OnClickInline(func(el dom.Element) {
        // This keeps the entire 1MB in memory
        processData(largeData)
    }),
)

// ✅ Good: Extract only what you need
processID := extractProcessID(largeData)
Button(
    Text("Process"),
    dom.OnClickInline(func(el dom.Element) {
        // Only captures the small processID
        processDataByID(processID)
    }),
)
```

#### 2. Use Signals for Shared State

```go
// ✅ Good: Shared state via signals
sharedCounter := reactivity.NewSignal(0)

// Multiple handlers can reference the same signal
// without duplicating state
for i := 0; i < 100; i++ {
    buttons = append(buttons, Button(
        Text(fmt.Sprintf("Button %d", i)),
        dom.OnClickInline(func(el dom.Element) {
            sharedCounter.Set(sharedCounter.Get() + 1)
        }),
    ))
}
```

#### 3. Handler Cleanup Patterns

```go
// For long-running applications, consider periodic cleanup
func PeriodicCleanup() {
    // This clears all inline handlers and re-attaches delegates
    // Use sparingly and only when necessary
    dom.AttachInlineDelegates()
}

// Better: Design components to be naturally cleaned up
func ShortLivedComponent() g.Node {
    // Handlers are automatically cleaned when component unmounts
    return Button(
        Text("Temporary"),
        dom.OnClickInline(func(el dom.Element) {
            // This will be cleaned up automatically
        }),
    )
}
```

## Event Delegation Internals

### How Event Matching Works

```html
<!-- Generated HTML for inline events -->
<button data-click-inline="handler-123">Click me</button>
<input data-input-inline="handler-456" data-change-inline="handler-789">
<div data-keydown-inline-enter="handler-101" data-keydown-inline-escape="handler-102">
```

```javascript
// Simplified version of the delegation logic
document.addEventListener('click', function(event) {
    const handlerID = event.target.getAttribute('data-click-inline');
    if (handlerID && window.inlineClickHandlers[handlerID]) {
        window.inlineClickHandlers[handlerID](event.target);
    }
});
```

### Performance Characteristics

- **Event Lookup**: O(1) - Direct attribute access and hash map lookup
- **Memory per Handler**: ~100-200 bytes (function + registry entry)
- **Event Processing**: ~0.1ms per event (including bubbling and lookup)
- **DOM Overhead**: 1 attribute per event type per element

## Performance Best Practices

### 1. Minimize Handler Complexity

```go
// ❌ Avoid: Heavy computation in handlers
dom.OnClickInline(func(el dom.Element) {
    // Heavy computation blocks the UI thread
    result := heavyComputation()
    updateUI(result)
})

// ✅ Better: Defer heavy work
dom.OnClickInline(func(el dom.Element) {
    go func() {
        result := heavyComputation()
        // Update UI on main thread
        updateSignal.Set(result)
    }()
})

// ✅ Best: Use Web Workers for heavy computation
dom.OnClickInline(func(el dom.Element) {
    worker := js.Global().Get("Worker").New("worker.js")
    worker.Call("postMessage", data)
    worker.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        result := args[0].Get("data")
        updateSignal.Set(result.String())
        return nil
    }))
})
```

### 2. Batch DOM Updates

```go
// ❌ Avoid: Multiple individual updates
dom.OnClickInline(func(el dom.Element) {
    signal1.Set(value1) // Triggers re-render
    signal2.Set(value2) // Triggers re-render
    signal3.Set(value3) // Triggers re-render
})

// ✅ Better: Batch updates
dom.OnClickInline(func(el dom.Element) {
    reactivity.Batch(func() {
        signal1.Set(value1)
        signal2.Set(value2)
        signal3.Set(value3)
    }) // Single re-render
})
```

### 3. Optimize Event Frequency

```go
// ❌ Avoid: High-frequency events without throttling
dom.OnInputInline(func(el dom.Element) {
    // This fires on every keystroke
    expensiveSearch(el.Get("value").String())
})

// ✅ Better: Debounce high-frequency events
var debounceTimer *time.Timer
dom.OnInputInline(func(el dom.Element) {
    if debounceTimer != nil {
        debounceTimer.Stop()
    }
    
    value := el.Get("value").String()
    debounceTimer = time.AfterFunc(300*time.Millisecond, func() {
        expensiveSearch(value)
    })
})

// ✅ Alternative: Use throttling for continuous updates
var lastUpdate time.Time
dom.OnInputInline(func(el dom.Element) {
    now := time.Now()
    if now.Sub(lastUpdate) < 100*time.Millisecond {
        return // Skip this update
    }
    lastUpdate = now
    
    updateSearch(el.Get("value").String())
})
```

### 4. Efficient State Management

```go
// ✅ Use computed values for derived state
filteredItems := reactivity.NewComputed(func() []Item {
    search := searchTerm.Get()
    items := allItems.Get()
    
    if search == "" {
        return items
    }
    
    var filtered []Item
    searchLower := strings.ToLower(search)
    for _, item := range items {
        if strings.Contains(strings.ToLower(item.Name), searchLower) {
            filtered = append(filtered, item)
        }
    }
    return filtered
})

// Handler only updates the search term
dom.OnInputInline(func(el dom.Element) {
    searchTerm.Set(el.Get("value").String())
    // Filtering happens automatically via computed
})
```

## Edge Cases and Solutions

### 1. Event Timing Issues

```go
// Problem: Handler fires before DOM is ready
dom.OnClickInline(func(el dom.Element) {
    // This might fail if DOM isn't fully updated
    sibling := el.Get("nextElementSibling")
    if sibling.IsNull() {
        // Handle missing sibling
        return
    }
})

// Solution: Use requestAnimationFrame for DOM-dependent operations
dom.OnClickInline(func(el dom.Element) {
    js.Global().Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        // DOM is guaranteed to be updated
        sibling := el.Get("nextElementSibling")
        if !sibling.IsNull() {
            // Safe to manipulate sibling
        }
        return nil
    }))
})
```

### 2. Event Propagation Control

```go
// Problem: Event bubbles when it shouldn't
dom.OnClickInline(func(el dom.Element) {
    // This event will bubble to parent handlers
    handleClick()
})

// Solution: Stop propagation when needed
dom.OnClickInline(func(el dom.Element) {
    // Access the original event through the global context
    if event := js.Global().Get("event"); !event.IsUndefined() {
        event.Call("stopPropagation")
    }
    handleClick()
})

// Alternative: Use event parameter pattern
dom.OnClickInline(func(el dom.Element) {
    // Store event reference in element for access
    if eventObj := el.Get("__lastEvent"); !eventObj.IsUndefined() {
        eventObj.Call("preventDefault")
    }
    handleClick()
})
```

### 3. Dynamic Content Issues

```go
// Problem: Handlers on dynamically added content
func DynamicList() g.Node {
    items := reactivity.NewSignal([]string{"item1", "item2"})
    
    return Div(
        Button(
            Text("Add Item"),
            dom.OnClickInline(func(el dom.Element) {
                current := items.Get()
                newItem := fmt.Sprintf("item%d", len(current)+1)
                items.Set(append(current, newItem))
                // New items automatically get inline handlers
            }),
        ),
        comps.BindList(items, func(item string) g.Node {
            return Div(
                Text(item),
                Button(
                    Text("Delete"),
                    // This handler works even for dynamically added items
                    dom.OnClickInline(func(el dom.Element) {
                        current := items.Get()
                        var filtered []string
                        for _, i := range current {
                            if i != item {
                                filtered = append(filtered, i)
                            }
                        }
                        items.Set(filtered)
                    }),
                ),
            )
        }),
    )
}
```

### 4. Memory Leaks in Long-Running Apps

```go
// Problem: Handlers accumulate over time
var handlerCount int

func ComponentWithPotentialLeak() g.Node {
    // Each render creates new handlers
    handlerCount++
    
    return Button(
        Text(fmt.Sprintf("Handler #%d", handlerCount)),
        dom.OnClickInline(func(el dom.Element) {
            // This handler stays in memory even after re-render
            logutil.Log("Handler", handlerCount, "called")
        }),
    )
}

// Solution: Use stable references
func ComponentWithStableHandlers() g.Node {
    // Create handler once, reuse across renders
    handleClick := func(el dom.Element) {
        logutil.Log("Stable handler called")
    }
    
    return Button(
        Text("Stable Handler"),
        dom.OnClickInline(handleClick),
    )
}

// Better solution: Use signals for state
func ComponentWithSignals() g.Node {
    clickCount := reactivity.NewSignal(0)
    
    return Button(
        comps.BindText(func() string {
            return fmt.Sprintf("Clicked %d times", clickCount.Get())
        }),
        dom.OnClickInline(func(el dom.Element) {
            clickCount.Set(clickCount.Get() + 1)
        }),
    )
}
```

### 5. Cross-Browser Compatibility

```go
// Handle browser differences in event handling
dom.OnKeyDownInline(func(el dom.Element) {
    // Access event through multiple possible paths
    var event js.Value
    if global := js.Global(); !global.IsUndefined() {
        if evt := global.Get("event"); !evt.IsUndefined() {
            event = evt
        } else if evt := el.Get("event"); !evt.IsUndefined() {
            event = evt
        }
    }
    
    if !event.IsUndefined() {
        // Handle key codes vs key names
        var key string
        if keyProp := event.Get("key"); !keyProp.IsUndefined() {
            key = keyProp.String() // Modern browsers
        } else if keyCode := event.Get("keyCode"); !keyCode.IsUndefined() {
            // Fallback for older browsers
            switch keyCode.Int() {
            case 13:
                key = "Enter"
            case 27:
                key = "Escape"
            }
        }
        
        handleKeyPress(key)
    }
}, "Enter")
```

## Debugging and Profiling

### 1. Handler Registry Inspection

```go
// Add debugging helpers
func DebugInlineHandlers() {
    js.Global().Set("debugInlineHandlers", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        // Inspect handler registries
        clickHandlers := js.Global().Get("inlineClickHandlers")
        inputHandlers := js.Global().Get("inlineInputHandlers")
        
        logutil.Log("Click handlers:", clickHandlers.Length())
        logutil.Log("Input handlers:", inputHandlers.Length())
        
        return nil
    }))
}

// Call from browser console: debugInlineHandlers()
```

### 2. Performance Monitoring

```go
// Wrap handlers with performance monitoring
func MonitoredHandler(handler func(dom.Element)) func(dom.Element) {
    return func(el dom.Element) {
        start := time.Now()
        handler(el)
        duration := time.Since(start)
        
        if duration > 10*time.Millisecond {
            logutil.Logf("Slow handler detected: %v", duration)
        }
    }
}

// Usage
dom.OnClickInline(MonitoredHandler(func(el dom.Element) {
    // Your handler logic
}))
```

### 3. Memory Usage Tracking

```go
// Track handler memory usage
var handlerMemoryUsage int64

func TrackHandlerMemory(handler func(dom.Element)) func(dom.Element) {
    // Estimate memory usage (rough approximation)
    handlerMemoryUsage += 200 // bytes per handler
    
    return func(el dom.Element) {
        handler(el)
        
        // Log memory usage periodically
        if handlerMemoryUsage%10000 == 0 {
            logutil.Logf("Estimated handler memory: %d KB", handlerMemoryUsage/1024)
        }
    }
}
```

## Comparison with Traditional Binding

### Performance Metrics

| Metric | Inline Events | Traditional Binding |
|--------|---------------|--------------------|
| Memory per handler | ~200 bytes | ~1KB+ |
| DOM listeners | 1 per event type | 1 per element |
| Event processing | O(1) lookup | Direct call |
| Setup time | O(1) | O(n) elements |
| Cleanup complexity | Automatic | Manual required |

### When to Use Each Approach

#### Use Inline Events When:
- Building typical web applications
- You have many interactive elements
- You want automatic cleanup
- Performance and memory usage matter
- You prefer declarative event binding

#### Consider Traditional Binding When:
- You need direct access to the original event object
- You're integrating with third-party libraries that expect traditional listeners
- You have very specific event handling requirements
- You're working with non-standard events

### Migration Strategy

```go
// Before: Traditional binding
func OldComponent() g.Node {
    comps.OnMount(func() {
        btn := dom.GetElementByID("my-button")
        dom.BindClickToCallback(btn, func() {
            handleClick()
        })
    })
    
    return Button(ID("my-button"), Text("Click"))
}

// After: Inline events
func NewComponent() g.Node {
    return Button(
        Text("Click"),
        dom.OnClickInline(func(el dom.Element) {
            handleClick()
        }),
    )
}
```

This comprehensive guide covers the performance characteristics, edge cases, and optimization strategies for inline events. By following these guidelines, you can build high-performance, memory-efficient applications with uiwgo's inline event system.