# Helper Functions Quick Reference

A concise reference for UIwGo helper functions with syntax examples and common patterns.

## Table of Contents

- [Conditional Rendering](#conditional-rendering)
- [List Rendering](#list-rendering)
- [Dynamic Components](#dynamic-components)
- [Text Binding](#text-binding)
- [Common Patterns](#common-patterns)
- [Performance Tips](#performance-tips)

## Conditional Rendering

### Show

Display content conditionally based on a boolean signal.

```go
// Basic usage
comps.Show(comps.ShowProps{
    When: isVisible,
    Children: g.Div(g.Text("Visible content")),
})

// With fallback
comps.Show(comps.ShowProps{
    When: hasData,
    Children: g.Div(g.Text("Data loaded")),
    Fallback: g.Div(g.Text("Loading...")),
})

// With memo for complex conditions
comps.Show(comps.ShowProps{
    When: reactivity.NewMemo(func() bool {
        return user.Get().IsAdmin && feature.Get().Enabled
    }),
    Children: g.Button(g.Text("Admin Panel")),
})
```

### Switch/Match

Conditional rendering with multiple branches.

```go
// Basic switch
comps.Switch(comps.SwitchProps{
    When: status,
    Children: []g.Node{
        comps.Match(comps.MatchProps{
            When: "loading",
            Children: g.Div(g.Text("Loading...")),
        }),
        comps.Match(comps.MatchProps{
            When: "success",
            Children: g.Div(g.Text("Success!")),
        }),
        comps.Match(comps.MatchProps{
            When: "error",
            Children: g.Div(g.Text("Error occurred")),
        }),
    },
})

// With computed values
comps.Switch(comps.SwitchProps{
    When: reactivity.NewMemo(func() string {
        count := counter.Get()
        if count == 0 {
            return "empty"
        } else if count < 10 {
            return "few"
        }
        return "many"
    }),
    Children: []g.Node{
        comps.Match(comps.MatchProps{
            When: "empty",
            Children: g.P(g.Text("No items")),
        }),
        comps.Match(comps.MatchProps{
            When: "few",
            Children: g.P(g.Text("A few items")),
        }),
        comps.Match(comps.MatchProps{
            When: "many",
            Children: g.P(g.Text("Many items")),
        }),
    },
})
```

## List Rendering

### For

Render lists with automatic reconciliation.

```go
// Basic list
comps.For(comps.ForProps[string]{
    Items: items,
    Key: func(item string) string {
        return item
    },
    Children: func(item string, index int) g.Node {
        return g.Li(g.Text(item))
    },
})

// Complex objects
comps.For(comps.ForProps[User]{
    Items: users,
    Key: func(user User) string {
        return user.ID
    },
    Children: func(user User, index int) g.Node {
        return g.Div(
            g.Class("user-card"),
            g.H3(g.Text(user.Name)),
            g.P(g.Text(user.Email)),
        )
    },
})

// With filtered/sorted data
filteredUsers := reactivity.NewMemo(func() []User {
    users := allUsers.Get()
    filter := searchFilter.Get()
    
    var filtered []User
    for _, user := range users {
        if strings.Contains(user.Name, filter) {
            filtered = append(filtered, user)
        }
    }
    return filtered
})

comps.For(comps.ForProps[User]{
    Items: filteredUsers,
    Key: func(user User) string {
        return user.ID
    },
    Children: func(user User, index int) g.Node {
        return renderUserCard(user, index)
    },
})
```

### Index

Render lists where index matters more than item identity.

```go
// When order/position is important
comps.Index(comps.IndexProps[string]{
    Items: items,
    Children: func(item reactivity.Memo[string], index int) g.Node {
        return g.Div(
            g.Text(fmt.Sprintf("%d: %s", index, item.Get())),
        )
    },
})

// For arrays or position-dependent rendering
comps.Index(comps.IndexProps[GridCell]{
    Items: gridData,
    Children: func(cell reactivity.Memo[GridCell], index int) g.Node {
        row := index / gridWidth
        col := index % gridWidth
        
        return g.Div(
            g.Class("grid-cell"),
            g.Style(fmt.Sprintf("grid-row: %d; grid-column: %d;", row+1, col+1)),
            g.Text(cell.Get().Value),
        )
    },
})
```

## Dynamic Components

### Dynamic

Render different components based on runtime conditions.

```go
// Component switching
comps.Dynamic(comps.DynamicProps{
    Component: componentType,
    Props: componentProps,
})

// With computed component
comps.Dynamic(comps.DynamicProps{
    Component: reactivity.NewMemo(func() string {
        mode := viewMode.Get()
        switch mode {
        case "list":
            return "ListView"
        case "grid":
            return "GridView"
        case "table":
            return "TableView"
        default:
            return "ListView"
        }
    }),
    Props: reactivity.NewMemo(func() map[string]any {
        return map[string]any{
            "data": data.Get(),
            "options": viewOptions.Get(),
        }
    }),
})
```

## Text Binding

### BindText

Bind reactive text content to DOM elements.

```go
// Simple text binding
g.Span(
    comps.BindText(userName),
)

// Computed text
g.P(
    comps.BindText(reactivity.NewMemo(func() string {
        count := itemCount.Get()
        if count == 1 {
            return "1 item"
        }
        return fmt.Sprintf("%d items", count)
    })),
)

// Formatted text
g.Div(
    comps.BindText(reactivity.NewMemo(func() string {
        user := currentUser.Get()
        return fmt.Sprintf("Welcome, %s!", user.Name)
    })),
)
```

## Common Patterns

### Loading States

```go
// Simple loading
comps.Show(comps.ShowProps{
    When: isLoading,
    Children: g.Div(g.Text("Loading...")),
    Fallback: renderContent(),
})

// Loading with different states
comps.Switch(comps.SwitchProps{
    When: loadingState,
    Children: []g.Node{
        comps.Match(comps.MatchProps{
            When: "idle",
            Children: g.Button(g.Text("Load Data")),
        }),
        comps.Match(comps.MatchProps{
            When: "loading",
            Children: g.Div(g.Text("Loading...")),
        }),
        comps.Match(comps.MatchProps{
            When: "success",
            Children: renderData(),
        }),
        comps.Match(comps.MatchProps{
            When: "error",
            Children: g.Div(g.Text("Error loading data")),
        }),
    },
})
```

### Empty States

```go
// List with empty state
comps.Show(comps.ShowProps{
    When: reactivity.NewMemo(func() bool {
        return len(items.Get()) > 0
    }),
    Children: g.Ul(
        comps.For(comps.ForProps[Item]{
            Items: items,
            Key: func(item Item) string { return item.ID },
            Children: func(item Item, index int) g.Node {
                return g.Li(g.Text(item.Name))
            },
        }),
    ),
    Fallback: g.Div(
        g.Class("empty-state"),
        g.Text("No items found"),
    ),
})
```

### Form Validation

```go
// Input with validation
g.Div(
    g.Input(
        g.Value(email.Get()),
        g.Class(func() string {
            if isValidEmail.Get() {
                return "valid"
            }
            return "invalid"
        }()),
        dom.OnInput(func(value string) {
            email.Set(value)
        }),
    ),
    comps.Show(comps.ShowProps{
        When: reactivity.NewMemo(func() bool {
            return !isValidEmail.Get() && email.Get() != ""
        }),
        Children: g.Div(
            g.Class("error"),
            g.Text("Please enter a valid email"),
        ),
    }),
)
```

### Conditional Styling

```go
// Dynamic classes
g.Div(
    g.Class(func() string {
        classes := []string{"base-class"}
        if isActive.Get() {
            classes = append(classes, "active")
        }
        if isSelected.Get() {
            classes = append(classes, "selected")
        }
        return strings.Join(classes, " ")
    }()),
    g.Text("Content"),
)

// Dynamic styles
g.Div(
    g.Style(func() string {
        opacity := "1"
        if isHidden.Get() {
            opacity = "0"
        }
        return fmt.Sprintf("opacity: %s; transition: opacity 0.3s;", opacity)
    }()),
    g.Text("Fading content"),
)
```

### Nested Conditions

```go
// Multiple levels of conditions
comps.Show(comps.ShowProps{
    When: isLoggedIn,
    Children: comps.Show(comps.ShowProps{
        When: hasPermission,
        Children: g.Div(
            g.Text("Protected content"),
        ),
        Fallback: g.Div(
            g.Text("Access denied"),
        ),
    }),
    Fallback: g.Div(
        g.Text("Please log in"),
    ),
})
```

## Performance Tips

### Key Functions

```go
// Good: Stable, unique keys
Key: func(user User) string {
    return user.ID
}

// Bad: Non-unique or unstable keys
Key: func(user User) string {
    return user.Name // Names might not be unique
}

// Bad: Index as key for dynamic lists
Key: func(user User) string {
    return fmt.Sprintf("%d", index) // Index changes when list reorders
}
```

### Memo Usage

```go
// Good: Memo for expensive computations
filteredItems := reactivity.NewMemo(func() []Item {
    items := allItems.Get()
    filter := searchFilter.Get()
    
    // Expensive filtering logic
    return expensiveFilter(items, filter)
})

// Good: Memo for complex conditions
isVisible := reactivity.NewMemo(func() bool {
    return user.Get().IsAdmin && 
           feature.Get().Enabled && 
           !maintenance.Get().Active
})
```

### Avoid Inline Functions

```go
// Good: Define render functions outside
func renderUserCard(user User, index int) g.Node {
    return g.Div(
        g.Class("user-card"),
        g.H3(g.Text(user.Name)),
        g.P(g.Text(user.Email)),
    )
}

comps.For(comps.ForProps[User]{
    Items: users,
    Key: func(user User) string { return user.ID },
    Children: renderUserCard,
})

// Avoid: Inline complex render functions
comps.For(comps.ForProps[User]{
    Items: users,
    Key: func(user User) string { return user.ID },
    Children: func(user User, index int) g.Node {
        // Complex rendering logic here
        // This creates new functions on every render
    },
})
```

### Signal Dependencies

```go
// Good: Minimal dependencies
fullName := reactivity.NewMemo(func() string {
    user := currentUser.Get()
    return user.FirstName + " " + user.LastName
})

// Avoid: Unnecessary dependencies
badMemo := reactivity.NewMemo(func() string {
    user := currentUser.Get()
    settings := appSettings.Get() // Unnecessary if not used
    return user.FirstName + " " + user.LastName
})
```

## Quick Syntax Reference

| Helper | Purpose | Key Props |
|--------|---------|----------|
| `Show` | Conditional rendering | `When`, `Children`, `Fallback` |
| `Switch` | Multi-branch conditions | `When`, `Children` |
| `Match` | Switch branch | `When`, `Children` |
| `For` | List rendering with keys | `Items`, `Key`, `Children` |
| `Index` | Position-based lists | `Items`, `Children` |
| `Dynamic` | Component switching | `Component`, `Props` |
| `BindText` | Reactive text | Signal or Memo |

## Common Gotchas

1. **Key Stability**: Always use stable, unique keys for `For` helper
2. **Memo Dependencies**: Only access signals you actually need in memos
3. **Function Recreation**: Avoid creating new functions in render loops
4. **Signal Updates**: Don't update signals during render (use effects instead)
5. **Memory Leaks**: Clean up effects and subscriptions when components unmount

For more detailed information, see the [Helper Functions Guide](./helper-functions.md) and [Performance Optimization](./performance-optimization.md).


## Alpine-inspired inline helpers

A concise list of new inline event and lifecycle helpers inspired by Alpine.js. Each returns a gomponents attribute you can attach to elements directly.

- OnInitInline(fn)
  - Run once after element connects to DOM
  - Example:
    ```go
    g.Div(dom.OnInitInline(func(el dom.Element) { /* setup */ }))
    ```

- OnDestroyInline(fn)
  - Run when element is removed from DOM
  - Example:
    ```go
    g.Div(dom.OnDestroyInline(func(el dom.Element) { /* cleanup */ }))
    ```

- OnVisibleInline(fn)
  - Fire once when element enters viewport (IntersectionObserver)
  - Example:
    ```go
    g.Div(dom.OnVisibleInline(func(el dom.Element) { /* lazy-load */ }))
    ```

- OnResizeInline(fn)
  - React to element resize events (ResizeObserver)
  - Example:
    ```go
    g.Div(dom.OnResizeInline(func(el dom.Element) { /* size changed */ }))
    ```

- OnClickOnceInline(fn)
  - Click handler that runs only once and then auto-unregisters
  - Example:
    ```go
    g.Button(g.Text("Once"), dom.OnClickOnceInline(func(el dom.Element) { /* do */ }))
    ```

See also: Design notes and options in [Alpine-inspired inline events](../alpine_inline_events.md).
