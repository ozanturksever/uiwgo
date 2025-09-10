# Troubleshooting UIwGo Helper Functions

This guide helps you diagnose and fix common issues when working with UIwGo helper functions.

## Related Documentation

- **[Helper Functions Guide](./helper-functions.md)** - Comprehensive guide to all helper functions
- **[Quick Reference](./quick-reference.md)** - Concise syntax reference
- **[Real-World Examples](./real-world-examples.md)** - Practical application examples
- **[Integration Examples](./integration-examples.md)** - Complex multi-helper scenarios
- **[Performance Optimization](./performance-optimization.md)** - Performance best practices

## Table of Contents

- [Common Issues](#common-issues)
- [Performance Problems](#performance-problems)
- [Memory Leaks](#memory-leaks)
- [Reactivity Issues](#reactivity-issues)
- [Rendering Problems](#rendering-problems)
- [Event Handling Issues](#event-handling-issues)
- [Testing Problems](#testing-problems)
- [Debugging Techniques](#debugging-techniques)

## Common Issues

### Issue: "Key function returns empty string"

**Problem**: Your `For` helper is not rendering items correctly or throwing errors.

**Symptoms**:
- Items not appearing in the DOM
- Console errors about duplicate keys
- Unexpected re-rendering behavior

**Solution**:
```go
// ❌ Bad: Key function returns empty string
comps.For(comps.ForProps[User]{
    Items: users,
    Key: func(user User) string { return "" }, // Wrong!
    Children: func(user User, index int) g.Node {
        return g.Div(g.Text(user.Name))
    },
})

// ✅ Good: Key function returns unique identifier
comps.For(comps.ForProps[User]{
    Items: users,
    Key: func(user User) string { return user.ID }, // Correct!
    Children: func(user User, index int) g.Node {
        return g.Div(g.Text(user.Name))
    },
})
```

**Prevention**:
- Always ensure your key function returns a unique, stable identifier
- Use IDs, UUIDs, or other unique properties
- Avoid using array indices as keys for dynamic lists

### Issue: "Signal not updating UI"

**Problem**: You update a signal but the UI doesn't reflect the change.

**Symptoms**:
- Signal value changes but DOM doesn't update
- Components seem "frozen" with old data
- Manual refresh shows correct data

**Solution**:
```go
// ❌ Bad: Mutating signal value directly
users := reactivity.CreateSignal([]User{{ID: "1", Name: "John"}})
// This won't trigger updates!
users.Get()[0].Name = "Jane"

// ✅ Good: Creating new value and setting signal
users := reactivity.CreateSignal([]User{{ID: "1", Name: "John"}})
currentUsers := users.Get()
currentUsers[0].Name = "Jane"
users.Set(currentUsers) // This triggers updates

// ✅ Better: Using immutable updates
users := reactivity.CreateSignal([]User{{ID: "1", Name: "John"}})
newUsers := make([]User, len(users.Get()))
copy(newUsers, users.Get())
newUsers[0].Name = "Jane"
users.Set(newUsers)
```

**Prevention**:
- Always call `Set()` to update signal values
- Avoid mutating signal values in place
- Consider using immutable update patterns

### Issue: "Infinite re-rendering loop"

**Problem**: Your component keeps re-rendering indefinitely.

**Symptoms**:
- Browser becomes unresponsive
- Console shows repeated render cycles
- High CPU usage

**Solution**:
```go
// ❌ Bad: Creating new signal in render function
func renderComponent() g.Node {
    // This creates a new signal on every render!
    count := reactivity.CreateSignal(0)
    return g.Div(g.Text(fmt.Sprintf("Count: %d", count.Get())))
}

// ✅ Good: Signal created outside render function
type Component struct {
    count reactivity.Signal[int]
}

func NewComponent() *Component {
    return &Component{
        count: reactivity.CreateSignal(0),
    }
}

func (c *Component) render() g.Node {
    return g.Div(g.Text(fmt.Sprintf("Count: %d", c.count.Get())))
}
```

**Prevention**:
- Create signals in component constructors, not render functions
- Use memos for expensive computations
- Avoid side effects in render functions

### Issue: "Memory leak with event listeners"

**Problem**: Event listeners are not being cleaned up properly.

**Symptoms**:
- Memory usage keeps increasing
- Multiple event handlers firing for the same event
- Performance degradation over time

**Solution**:
```go
// ❌ Bad: No cleanup of event listeners
func setupComponent() {
    button := dom.GetWindow().Document().GetElementByID("myButton")
    button.AddEventListener("click", false, func(event dom.Event) {
        // Handler code
    })
    // No cleanup!
}

// ✅ Good: Proper cleanup with disposal
type Component struct {
    disposers []func()
}

func (c *Component) setupEventListeners() {
    button := dom.GetWindow().Document().GetElementByID("myButton")
    disposer := button.AddEventListener("click", false, func(event dom.Event) {
        // Handler code
    })
    c.disposers = append(c.disposers, disposer)
}

func (c *Component) cleanup() {
    for _, dispose := range c.disposers {
        dispose()
    }
    c.disposers = nil
}
```

**Prevention**:
- Always clean up event listeners when components are destroyed
- Use the disposal functions returned by event listener setup
- Consider using lifecycle hooks for automatic cleanup

## Performance Problems

### Issue: "Slow list rendering"

**Problem**: Large lists render slowly or cause UI freezing.

**Symptoms**:
- Noticeable delay when rendering lists
- Browser becomes unresponsive
- Poor scrolling performance

**Solution**:
```go
// ❌ Bad: Rendering all items at once
comps.For(comps.ForProps[Item]{
    Items: allItems, // Could be thousands of items
    Key: func(item Item) string { return item.ID },
    Children: func(item Item, index int) g.Node {
        return renderComplexItem(item) // Expensive rendering
    },
})

// ✅ Good: Virtual scrolling or pagination
visibleItems := reactivity.CreateMemo(func() []Item {
    items := allItems.Get()
    start := scrollPosition.Get() / itemHeight
    end := start + visibleCount
    if end > len(items) {
        end = len(items)
    }
    return items[start:end]
})

comps.For(comps.ForProps[Item]{
    Items: visibleItems,
    Key: func(item Item) string { return item.ID },
    Children: func(item Item, index int) g.Node {
        return renderComplexItem(item)
    },
})
```

**Prevention**:
- Implement virtual scrolling for large lists
- Use pagination to limit rendered items
- Optimize individual item rendering
- Consider lazy loading for complex items

### Issue: "Expensive computations in render"

**Problem**: Heavy calculations are performed on every render.

**Symptoms**:
- Slow UI updates
- High CPU usage during interactions
- Delayed response to user input

**Solution**:
```go
// ❌ Bad: Expensive computation in render
func renderStats() g.Node {
    // This runs on every render!
    total := 0
    for _, item := range items.Get() {
        total += expensiveCalculation(item)
    }
    
    return g.Div(g.Text(fmt.Sprintf("Total: %d", total)))
}

// ✅ Good: Using memo for expensive computation
expensiveTotal := reactivity.CreateMemo(func() int {
    total := 0
    for _, item := range items.Get() {
        total += expensiveCalculation(item)
    }
    return total
})

func renderStats() g.Node {
    return g.Div(g.Text(fmt.Sprintf("Total: %d", expensiveTotal.Get())))
}
```

**Prevention**:
- Use `reactivity.CreateMemo()` for expensive computations
- Cache results when possible
- Debounce frequent updates
- Profile your application to identify bottlenecks

## Memory Leaks

### Issue: "Signals not being garbage collected"

**Problem**: Old signals remain in memory after components are destroyed.

**Symptoms**:
- Memory usage continuously increases
- Application becomes slower over time
- Browser eventually crashes

**Solution**:
```go
// ❌ Bad: No cleanup of signals
type Component struct {
    data reactivity.Signal[string]
    computed reactivity.Memo[int]
}

func (c *Component) destroy() {
    // No cleanup - signals remain in memory!
}

// ✅ Good: Proper signal cleanup
type Component struct {
    data reactivity.Signal[string]
    computed reactivity.Memo[int]
    disposers []func()
}

func (c *Component) destroy() {
    // Clean up all reactive subscriptions
    for _, dispose := range c.disposers {
        dispose()
    }
    c.disposers = nil
    
    // Clear signal references
    c.data = nil
    c.computed = nil
}
```

**Prevention**:
- Implement proper cleanup in component destructors
- Use weak references where appropriate
- Monitor memory usage during development
- Test component lifecycle thoroughly

## Reactivity Issues

### Issue: "Computed values not updating"

**Problem**: Memos or computed values don't update when dependencies change.

**Symptoms**:
- Stale computed values
- UI showing outdated information
- Manual refresh fixes the issue

**Solution**:
```go
// ❌ Bad: Accessing signal outside memo function
baseValue := someSignal.Get() // Captured at creation time
computed := reactivity.CreateMemo(func() int {
    return baseValue * 2 // Won't update when someSignal changes
})

// ✅ Good: Accessing signal inside memo function
computed := reactivity.CreateMemo(func() int {
    return someSignal.Get() * 2 // Will update when someSignal changes
})
```

**Prevention**:
- Always access signals inside memo functions
- Ensure all dependencies are properly tracked
- Test reactivity chains thoroughly

### Issue: "Circular dependencies in reactivity"

**Problem**: Signals depend on each other in a circular manner.

**Symptoms**:
- Stack overflow errors
- Infinite update loops
- Application crashes

**Solution**:
```go
// ❌ Bad: Circular dependency
signalA := reactivity.CreateSignal(0)
signalB := reactivity.CreateSignal(0)

// This creates a circular dependency!
reactivity.CreateEffect(func() {
    signalB.Set(signalA.Get() + 1)
})
reactivity.CreateEffect(func() {
    signalA.Set(signalB.Get() + 1)
})

// ✅ Good: Break circular dependency
signalA := reactivity.CreateSignal(0)
signalB := reactivity.CreateMemo(func() int {
    return signalA.Get() + 1 // One-way dependency
})
```

**Prevention**:
- Design clear data flow patterns
- Avoid bidirectional signal dependencies
- Use state machines for complex state management

## Rendering Problems

### Issue: "Components not re-rendering"

**Problem**: UI doesn't update when it should.

**Symptoms**:
- Stale UI content
- User interactions don't reflect in UI
- Data changes but UI remains the same

**Solution**:
```go
// ❌ Bad: Not using reactive values in render
func renderUser() g.Node {
    user := getCurrentUser() // Static call
    return g.Div(g.Text(user.Name))
}

// ✅ Good: Using reactive signals
func renderUser() g.Node {
    return g.Div(g.Text(currentUser.Get().Name)) // Reactive
}
```

**Prevention**:
- Use reactive signals for all dynamic data
- Ensure render functions access current signal values
- Test UI updates thoroughly

### Issue: "Incorrect conditional rendering"

**Problem**: `Show` or `Switch` components don't behave as expected.

**Symptoms**:
- Content shows when it shouldn't
- Content doesn't show when it should
- Wrong content displayed in switches

**Solution**:
```go
// ❌ Bad: Using non-reactive condition
isVisible := user.IsAdmin // Static value
comps.Show(comps.ShowProps{
    When: reactivity.CreateSignal(isVisible),
    Children: g.Div(g.Text("Admin Panel")),
})

// ✅ Good: Using reactive condition
comps.Show(comps.ShowProps{
    When: reactivity.CreateMemo(func() bool {
        return currentUser.Get().IsAdmin
    }),
    Children: g.Div(g.Text("Admin Panel")),
})
```

**Prevention**:
- Use reactive signals for all conditions
- Test all conditional branches
- Verify switch case matching logic

## Event Handling Issues

### Issue: "Event handlers not firing"

**Problem**: Click or other event handlers don't execute.

**Symptoms**:
- Buttons don't respond to clicks
- Form submissions don't work
- No console errors

**Solution**:
```go
// ❌ Bad: Handler attached to wrong element
g.Div(
    g.Button(g.Text("Click me")),
    dom.OnClick(func() { // Handler on div, not button
        handleClick()
    }),
)

// ✅ Good: Handler on correct element
g.Div(
    g.Button(
        g.Text("Click me"),
        dom.OnClick(func() { // Handler on button
            handleClick()
        }),
    ),
)
```

**Prevention**:
- Attach event handlers to the correct elements
- Test event handling thoroughly
- Use browser dev tools to verify event listeners

### Issue: "Event handler memory leaks"

**Problem**: Event handlers accumulate over time.

**Symptoms**:
- Multiple handlers firing for single events
- Memory usage increases
- Performance degradation

**Solution**:
```go
// ❌ Bad: Adding handlers without cleanup
func updateButton() {
    button.AddEventListener("click", false, handleClick)
    // Old handlers never removed!
}

// ✅ Good: Remove old handlers before adding new ones
func updateButton() {
    if currentHandler != nil {
        currentHandler() // Dispose old handler
    }
    currentHandler = button.AddEventListener("click", false, handleClick)
}
```

**Prevention**:
- Clean up old event handlers before adding new ones
- Use component lifecycle methods for handler management
- Monitor event listener count in development

## Testing Problems

### Issue: "Tests failing due to timing issues"

**Problem**: Tests fail intermittently due to async operations.

**Symptoms**:
- Tests pass sometimes, fail other times
- Tests fail in CI but pass locally
- Timing-related error messages

**Solution**:
```go
// ❌ Bad: Not waiting for async operations
func TestComponent(t *testing.T) {
    component.loadData()
    // Test immediately - data might not be loaded yet!
    assert.Equal(t, expectedData, component.data.Get())
}

// ✅ Good: Waiting for async operations
func TestComponent(t *testing.T) {
    component.loadData()
    
    // Wait for data to load
    eventually(t, func() bool {
        return component.data.Get() != nil
    }, 5*time.Second)
    
    assert.Equal(t, expectedData, component.data.Get())
}
```

**Prevention**:
- Use proper waiting mechanisms in tests
- Mock async operations when possible
- Set appropriate timeouts for test operations

## Debugging Techniques

### Enable Debug Logging

```go
// Add debug logging to track signal updates
func debugSignal[T any](name string, signal reactivity.Signal[T]) {
    reactivity.CreateEffect(func() {
        logutil.Logf("Signal %s updated: %v", name, signal.Get())
    })
}

// Usage
debugSignal("userCount", userCount)
debugSignal("currentUser", currentUser)
```

### Use Browser DevTools

1. **Console Logging**: Add strategic `logutil.Log()` calls
2. **Performance Tab**: Profile rendering performance
3. **Memory Tab**: Monitor memory usage and leaks
4. **Elements Tab**: Inspect DOM structure and updates

### Component State Inspection

```go
// Add debug methods to components
type Component struct {
    // ... component fields
}

func (c *Component) debugState() {
    logutil.Logf("Component state: %+v", map[string]interface{}{
        "data": c.data.Get(),
        "loading": c.loading.Get(),
        "error": c.error.Get(),
    })
}

// Call in event handlers or effects
func (c *Component) handleClick() {
    c.debugState() // Log state before action
    c.performAction()
    c.debugState() // Log state after action
}
```

### Test Reactivity Chains

```go
func TestReactivityChain(t *testing.T) {
    // Create test signals
    input := reactivity.CreateSignal("initial")
    
    // Track all updates
    var updates []string
    reactivity.CreateEffect(func() {
        updates = append(updates, input.Get())
    })
    
    // Test updates
    input.Set("updated")
    
    expected := []string{"initial", "updated"}
    assert.Equal(t, expected, updates)
}
```

### Performance Profiling

```go
// Add timing to expensive operations
func expensiveOperation() {
    start := time.Now()
    defer func() {
        logutil.Logf("Operation took: %v", time.Since(start))
    }()
    
    // ... expensive code
}

// Profile render cycles
func profiledRender() g.Node {
    start := time.Now()
    defer func() {
        logutil.Logf("Render took: %v", time.Since(start))
    }()
    
    return actualRender()
}
```

By following these troubleshooting guidelines and implementing proper debugging techniques, you can quickly identify and resolve issues in your UIwGo applications. Remember to test thoroughly and monitor performance regularly to catch issues early.