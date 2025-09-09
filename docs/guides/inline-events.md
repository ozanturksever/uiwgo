# Inline Event Binding Guide

**Inline event binding is the preferred method for handling DOM events in uiwgo applications.** This approach provides a clean, declarative way to attach event handlers directly within your component definitions without requiring manual DOM queries or lifecycle hooks.

## Overview

Inline event binding allows you to attach event handlers directly to elements during component creation:

```go
Button(
    Text("Click me"),
    dom.OnClickInline(func(el dom.Element) {
        // Handle click event
    }),
)
```

This is simpler and more maintainable than the traditional approach:

```go
// Traditional approach (NOT recommended)
Button(ID("my-button"), Text("Click me")),
comps.OnMount(func() {
    btn := dom.GetElementByID("my-button")
    dom.BindClickToCallback(btn, func() {
        // Handle click event
    })
})
```

## Why Inline Events Are Preferred

### 1. **Declarative and Co-located**
Event handlers are defined right where the element is created, making the code more readable and maintainable.

### 2. **No Manual DOM Queries**
Eliminate the need for `ID` attributes and `dom.GetElementByID` calls.

### 3. **Automatic Cleanup**
Event handlers are automatically cleaned up when components are unmounted, preventing memory leaks.

### 4. **Works with Dynamic Content**
Event delegation ensures handlers work for both static and dynamically created elements.

### 5. **Type Safety**
Handlers receive properly typed `dom.Element` parameters.

## Available Inline Event Handlers

### Click Events
```go
dom.OnClickInline(func(el dom.Element) {
    // Handle click
})
```

### Input Events
```go
dom.OnInputInline(func(el dom.Element) {
    value := el.Underlying().Get("value").String()
    // Handle input change
})
```

### Change Events
```go
dom.OnChangeInline(func(el dom.Element) {
    value := el.Underlying().Get("value").String()
    // Handle change (for select, checkbox, etc.)
})
```

### Keyboard Events
```go
// Generic keydown with specific key
dom.OnKeyDownInline(func(el dom.Element) {
    // Handle keydown
}, "Enter")

// Convenience methods for common keys
dom.OnEnterInline(func(el dom.Element) {
    // Handle Enter key
})

dom.OnEscapeInline(func(el dom.Element) {
    // Handle Escape key
})
```

### File Upload Events
```go
// File selection handler
dom.OnFileSelectInline(func(el dom.Element, files []js.Value) {
    for _, file := range files {
        name := file.Get("name").String()
        size := file.Get("size").Int()
        // Process selected file
    }
})

// File drop handler
dom.OnFileDropInline(func(el dom.Element, files []js.Value) {
    for _, file := range files {
        name := file.Get("name").String()
        size := file.Get("size").Int()
        // Process dropped file
    }
})
```

## Complete Example

Here's a comprehensive example showing various inline event handlers:

```go
func InteractiveComponent() Node {
    count := reactivity.CreateSignal(0)
    name := reactivity.CreateSignal("")
    todos := reactivity.CreateSignal([]string{})
    
    return Div(
        // Counter with click handlers
        Div(
            H3(Text("Counter")),
            P(comps.BindText(func() string {
                return fmt.Sprintf("Count: %d", count.Get())
            })),
            Button(
                Text("+"),
                dom.OnClickInline(func(el dom.Element) {
                    count.Set(count.Get() + 1)
                }),
            ),
            Button(
                Text("-"),
                dom.OnClickInline(func(el dom.Element) {
                    count.Set(count.Get() - 1)
                }),
            ),
            Button(
                Text("Reset"),
                dom.OnClickInline(func(el dom.Element) {
                    count.Set(0)
                }),
            ),
        ),
        
        // Input with real-time updates
        Div(
            H3(Text("Name Input")),
            Input(
                Type("text"),
                Placeholder("Enter your name"),
                dom.OnInputInline(func(el dom.Element) {
                    value := el.Underlying().Get("value").String()
                    name.Set(value)
                }),
            ),
            P(comps.BindText(func() string {
                return fmt.Sprintf("Hello, %s!", name.Get())
            })),
        ),
        
        // Todo list with Enter/Escape handling
        Div(
            H3(Text("Todo List")),
            Input(
                Type("text"),
                Placeholder("Add todo and press Enter"),
                dom.OnEnterInline(func(el dom.Element) {
                    value := el.Underlying().Get("value").String()
                    if value != "" {
                        items := append(todos.Get(), value)
                        todos.Set(items)
                        el.Underlying().Set("value", "") // Clear input
                    }
                }),
                dom.OnEscapeInline(func(el dom.Element) {
                    el.Underlying().Set("value", "") // Clear input
                }),
            ),
            Ul(
                comps.BindHTML(func() Node {
                    items := make([]Node, 0)
                    for _, todo := range todos.Get() {
                        items = append(items, Li(Text(todo)))
                    }
                    return Group(items)
                }),
            ),
        ),
        
        // File upload with drag and drop
        Div(
            H3(Text("File Upload")),
            Div(
                Style("border: 2px dashed #ccc; padding: 20px; text-align: center; margin: 10px 0;"),
                Text("Drop files here or click to select"),
                dom.OnFileDropInline(func(el dom.Element, files []js.Value) {
                    for _, file := range files {
                        name := file.Get("name").String()
                        size := file.Get("size").Int()
                        logutil.Logf("Dropped file: %s (%d bytes)", name, size)
                    }
                }),
                dom.OnClickInline(func(el dom.Element) {
                    // Trigger file input
                    input := dom.GetDocument().CreateElement("input")
                    input.SetAttribute("type", "file")
                    input.SetAttribute("multiple", "true")
                    input.Underlying().Call("click")
                }),
            ),
            Input(
                Type("file"),
                Multiple(true),
                Style("margin: 10px 0;"),
                dom.OnFileSelectInline(func(el dom.Element, files []js.Value) {
                    for _, file := range files {
                        name := file.Get("name").String()
                        size := file.Get("size").Int()
                        logutil.Logf("Selected file: %s (%d bytes)", name, size)
                    }
                }),
            ),
        ),
    )
}
```

## Best Practices

### 1. **Use Inline Events by Default**
Always prefer inline event handlers over traditional DOM binding unless you have a specific reason not to.

### 2. **Keep Handlers Simple**
Inline handlers should be concise. For complex logic, extract to separate functions:

```go
func handleComplexClick(el dom.Element, state *AppState) {
    // Complex logic here
}

// In component:
Button(
    Text("Complex Action"),
    dom.OnClickInline(func(el dom.Element) {
        handleComplexClick(el, appState)
    }),
)
```

### 3. **Access Element Properties Safely**
Always check if properties exist before accessing them:

```go
dom.OnInputInline(func(el dom.Element) {
    underlying := el.Underlying()
    if !underlying.Get("value").IsUndefined() {
        value := underlying.Get("value").String()
        // Use value
    }
})
```

### 4. **Use Appropriate Event Types**
- `OnInputInline`: For real-time input changes (text inputs, textareas)
- `OnChangeInline`: For discrete changes (select dropdowns, checkboxes)
- `OnClickInline`: For button clicks and clickable elements
- `OnEnterInline`/`OnEscapeInline`: For keyboard shortcuts

### 5. **Combine with Reactive Signals**
Inline events work perfectly with reactive signals for state management:

```go
func FormComponent() Node {
    formData := reactivity.CreateSignal(FormData{})
    
    return Form(
        Input(
            Type("text"),
            dom.OnInputInline(func(el dom.Element) {
                value := el.Underlying().Get("value").String()
                data := formData.Get()
                data.Name = value
                formData.Set(data)
            }),
        ),
        Button(
            Text("Submit"),
            dom.OnClickInline(func(el dom.Element) {
                submitForm(formData.Get())
            }),
        ),
    )
}
```

## Performance Considerations

### Event Delegation
Inline events use event delegation, which means:
- **Single listener per event type**: Only one click listener is attached to the root element
- **Efficient for dynamic content**: New elements automatically get event handling
- **Memory efficient**: No per-element listeners

### Handler Registration
- Handlers are stored in memory maps during component creation
- Automatic cleanup prevents memory leaks
- Thread-safe with proper synchronization

### When to Avoid Inline Events

1. **High-frequency events**: For events like `mousemove` or `scroll`, consider traditional binding with throttling
2. **Complex event logic**: If you need to prevent bubbling or capture phases
3. **Third-party integration**: When integrating with libraries that expect specific event handling patterns

## Migration from Traditional Events

### Before (Traditional)
```go
func OldComponent() Node {
    count := reactivity.CreateSignal(0)
    
    comps.OnMount(func() {
        btn := dom.GetElementByID("increment-btn")
        dom.BindClickToCallback(btn, func() {
            count.Set(count.Get() + 1)
        })
    })
    
    return Div(
        Button(ID("increment-btn"), Text("+")),
        P(comps.BindText(func() string {
            return fmt.Sprintf("Count: %d", count.Get())
        })),
    )
}
```

### After (Inline Events)
```go
func NewComponent() Node {
    count := reactivity.CreateSignal(0)
    
    return Div(
        Button(
            Text("+"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(count.Get() + 1)
            }),
        ),
        P(comps.BindText(func() string {
            return fmt.Sprintf("Count: %d", count.Get())
        })),
    )
}
```

### Migration Steps
1. Remove `ID` attributes from elements
2. Remove `comps.OnMount` hooks used only for event binding
3. Replace `dom.Bind*` calls with `dom.*Inline` functions
4. Move event handlers inline with element creation
5. Test thoroughly to ensure behavior is preserved

## Advanced Usage

### Conditional Event Handlers
```go
func ConditionalComponent() Node {
    enabled := reactivity.CreateSignal(true)
    
    var clickHandler Node
    if enabled.Get() {
        clickHandler = dom.OnClickInline(func(el dom.Element) {
            // Handle click when enabled
        })
    }
    
    return Button(
        Text("Maybe Clickable"),
        clickHandler, // Will be nil if disabled
    )
}
```

### Multiple Event Handlers
```go
Input(
    Type("text"),
    dom.OnInputInline(func(el dom.Element) {
        // Handle input changes
    }),
    dom.OnEnterInline(func(el dom.Element) {
        // Handle Enter key
    }),
    dom.OnEscapeInline(func(el dom.Element) {
        // Handle Escape key
    }),
)
```

### Accessing Event Context
```go
dom.OnClickInline(func(el dom.Element) {
    // Access element properties
    tagName := el.Underlying().Get("tagName").String()
    className := el.Underlying().Get("className").String()
    
    // Access parent/child elements
    parent := el.Underlying().Get("parentElement")
    
    // Modify element
    el.Underlying().Set("disabled", true)
})
```

## Troubleshooting

### Handler Not Firing
1. **Check element rendering**: Ensure the element is actually rendered to the DOM
2. **Verify handler registration**: Check that the inline handler is properly attached
3. **Browser console**: Look for JavaScript errors that might prevent event delegation

### Memory Leaks
1. **Automatic cleanup**: Inline events are automatically cleaned up on component unmount
2. **Manual cleanup**: If needed, use `reactivity.OnCleanup` for additional cleanup
3. **Avoid circular references**: Don't capture large objects in handler closures

### Performance Issues
1. **Handler complexity**: Keep inline handlers simple and fast
2. **Frequent updates**: For high-frequency events, consider debouncing
3. **Large lists**: For very large lists, consider virtualization

## Conclusion

Inline event binding provides a modern, declarative approach to event handling in uiwgo applications. By co-locating event handlers with element definitions, you create more maintainable and readable code while benefiting from automatic cleanup and efficient event delegation.

**Always prefer inline events over traditional DOM binding** unless you have specific requirements that necessitate the traditional approach.