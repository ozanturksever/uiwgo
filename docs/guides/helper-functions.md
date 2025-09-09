# Helper Functions Guide

This guide covers UIwGo's powerful helper functions for building dynamic, reactive user interfaces. Learn advanced patterns, best practices, and real-world usage examples.

## Table of Contents

- [Overview](#overview)
- [Conditional Rendering](#conditional-rendering)
- [List Rendering](#list-rendering)
- [Dynamic Components](#dynamic-components)
- [Performance Optimization](#performance-optimization)
- [Error Handling](#error-handling)
- [Advanced Patterns](#advanced-patterns)
- [Best Practices](#best-practices)
- [Common Pitfalls](#common-pitfalls)

## Overview

UIwGo provides a comprehensive set of helper functions that enable reactive, efficient UI development:

- **Show**: Conditional rendering based on boolean signals
- **For**: Keyed list rendering with efficient reconciliation
- **Index**: Index-based list rendering for stable references
- **Switch/Match**: Multi-branch conditional rendering
- **Dynamic**: Dynamic component switching
- **Memo**: Memoized components for performance
- **Lazy**: Lazy-loaded components
- **ErrorBoundary**: Error handling and fallbacks
- **Portal**: Render content outside component tree
- **Fragment**: Group elements without wrapper

### Mental Model

```go
// UIwGo helpers are reactive by design
visible := reactivity.NewSignal(true)
items := reactivity.NewSignal([]string{"a", "b", "c"})

// Helpers automatically update when signals change
comps.Show(comps.ShowProps{
    When: visible,
    Children: g.Text("I'm visible!"),
})

comps.For(comps.ForProps[string]{
    Items: items,
    Key: func(item string) string { return item },
    Children: func(item string, index int) g.Node {
        return g.Li(g.Text(item))
    },
})
```

## Conditional Rendering

### Basic Show Usage

```go
type AppState struct {
    showMessage reactivity.Signal[bool]
    message     reactivity.Signal[string]
}

func (s *AppState) render() g.Node {
    return g.Div(
        g.Button(
            g.Text("Toggle"),
            dom.OnClick(func() {
                s.showMessage.Set(!s.showMessage.Get())
            }),
        ),
        comps.Show(comps.ShowProps{
            When: s.showMessage,
            Children: g.P(
                g.Text(s.message.Get()),
            ),
        }),
    )
}
```

### Advanced Conditional Patterns

#### Computed Conditions

```go
type UserState struct {
    user     reactivity.Signal[*User]
    isAdmin  reactivity.Signal[bool]
}

func (s *UserState) render() g.Node {
    // Computed condition based on multiple signals
    canEdit := reactivity.NewMemo(func() bool {
        user := s.user.Get()
        return user != nil && (s.isAdmin.Get() || user.ID == getCurrentUserID())
    })
    
    return comps.Show(comps.ShowProps{
        When: canEdit,
        Children: g.Button(
            g.Text("Edit"),
            dom.OnClick(s.handleEdit),
        ),
    })
}
```

#### Nested Conditions

```go
func renderUserProfile(user reactivity.Signal[*User]) g.Node {
    return comps.Show(comps.ShowProps{
        When: reactivity.NewMemo(func() bool {
            return user.Get() != nil
        }),
        Children: g.Div(
            g.H2(g.Text("User Profile")),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    u := user.Get()
                    return u != nil && u.Avatar != ""
                }),
                Children: g.Img(g.Attr("src", user.Get().Avatar)),
            }),
            g.P(g.Text(user.Get().Name)),
        ),
    })
}
```

## List Rendering

### For vs Index: When to Use Each

#### Use For when:
- Items have stable, unique identifiers
- You need efficient reconciliation
- Items can be reordered, added, or removed

```go
type TodoItem struct {
    ID   string
    Text string
    Done bool
}

func renderTodoList(todos reactivity.Signal[[]TodoItem]) g.Node {
    return g.Ul(
        comps.For(comps.ForProps[TodoItem]{
            Items: todos,
            Key: func(item TodoItem) string {
                return item.ID // Stable unique key
            },
            Children: func(item TodoItem, index int) g.Node {
                return g.Li(
                    g.Input(
                        g.Type("checkbox"),
                        g.If(item.Done, g.Checked()),
                        dom.OnChange(func() {
                            // Update specific item
                            updateTodo(item.ID, !item.Done)
                        }),
                    ),
                    g.Text(item.Text),
                )
            },
        }),
    )
}
```

#### Use Index when:
- Items don't have stable identifiers
- You need reactive access to current item state
- List order is stable

```go
func renderCounters(counts reactivity.Signal[[]int]) g.Node {
    return g.Div(
        comps.Index(comps.IndexProps[int]{
            Items: counts,
            Children: func(getItem func() int, index int) g.Node {
                // getItem() always returns current value
                return g.Div(
                    g.Text(fmt.Sprintf("Counter %d: %d", index, getItem())),
                    g.Button(
                        g.Text("+"),
                        dom.OnClick(func() {
                            current := counts.Get()
                            current[index]++
                            counts.Set(current)
                        }),
                    ),
                )
            },
        }),
    )
}
```

### Advanced List Patterns

#### Filtered Lists

```go
type ProductFilter struct {
    products   reactivity.Signal[[]Product]
    searchTerm reactivity.Signal[string]
    category   reactivity.Signal[string]
}

func (f *ProductFilter) render() g.Node {
    filteredProducts := reactivity.NewMemo(func() []Product {
        products := f.products.Get()
        search := strings.ToLower(f.searchTerm.Get())
        cat := f.category.Get()
        
        var filtered []Product
        for _, p := range products {
            if cat != "" && p.Category != cat {
                continue
            }
            if search != "" && !strings.Contains(strings.ToLower(p.Name), search) {
                continue
            }
            filtered = append(filtered, p)
        }
        return filtered
    })
    
    return comps.For(comps.ForProps[Product]{
        Items: filteredProducts,
        Key: func(p Product) string { return p.ID },
        Children: func(p Product, index int) g.Node {
            return renderProductCard(p)
        },
    })
}
```

#### Grouped Lists

```go
func renderGroupedItems(items reactivity.Signal[[]Item]) g.Node {
    groupedItems := reactivity.NewMemo(func() map[string][]Item {
        groups := make(map[string][]Item)
        for _, item := range items.Get() {
            groups[item.Category] = append(groups[item.Category], item)
        }
        return groups
    })
    
    categories := reactivity.NewMemo(func() []string {
        var cats []string
        for cat := range groupedItems.Get() {
            cats = append(cats, cat)
        }
        sort.Strings(cats)
        return cats
    })
    
    return comps.For(comps.ForProps[string]{
        Items: categories,
        Key: func(cat string) string { return cat },
        Children: func(category string, index int) g.Node {
            categoryItems := reactivity.NewMemo(func() []Item {
                return groupedItems.Get()[category]
            })
            
            return g.Div(
                g.H3(g.Text(category)),
                comps.For(comps.ForProps[Item]{
                    Items: categoryItems,
                    Key: func(item Item) string { return item.ID },
                    Children: func(item Item, index int) g.Node {
                        return renderItem(item)
                    },
                }),
            )
        },
    })
}
```

## Dynamic Components

### Switch/Match for Multi-Branch Logic

```go
type ViewMode string

const (
    ViewModeList ViewMode = "list"
    ViewModeGrid ViewMode = "grid"
    ViewModeCard ViewMode = "card"
)

func renderProductView(products reactivity.Signal[[]Product], mode reactivity.Signal[ViewMode]) g.Node {
    return comps.Switch(comps.SwitchProps{
        When: mode,
        Children: []g.Node{
            comps.Match(comps.MatchProps{
                When: ViewModeList,
                Children: renderProductList(products),
            }),
            comps.Match(comps.MatchProps{
                When: ViewModeGrid,
                Children: renderProductGrid(products),
            }),
            comps.Match(comps.MatchProps{
                When: ViewModeCard,
                Children: renderProductCards(products),
            }),
        },
        Fallback: g.Div(g.Text("Unknown view mode")),
    })
}
```

### Dynamic Component Loading

```go
type ComponentType string

const (
    ComponentChart ComponentType = "chart"
    ComponentTable ComponentType = "table"
    ComponentMap   ComponentType = "map"
)

func renderDynamicWidget(componentType reactivity.Signal[ComponentType], data reactivity.Signal[any]) g.Node {
    component := reactivity.NewMemo(func() func() g.Node {
        switch componentType.Get() {
        case ComponentChart:
            return func() g.Node { return renderChart(data) }
        case ComponentTable:
            return func() g.Node { return renderTable(data) }
        case ComponentMap:
            return func() g.Node { return renderMap(data) }
        default:
            return func() g.Node { return g.Div(g.Text("Unknown component")) }
        }
    })
    
    return comps.Dynamic(comps.DynamicProps{
        Component: component,
    })
}
```

## Performance Optimization

### Memo for Expensive Computations

```go
func renderExpensiveChart(data reactivity.Signal[[]DataPoint]) g.Node {
    // Memoize expensive chart rendering
    return comps.Memo(func() g.Node {
        points := data.Get()
        
        // Expensive computation
        processedData := processChartData(points)
        chartConfig := generateChartConfig(processedData)
        
        return g.Div(
            g.Class("chart-container"),
            renderSVGChart(chartConfig),
        )
    }, data) // Re-compute only when data changes
}
```

### Lazy Loading

```go
func renderLazySection() g.Node {
    return comps.Lazy(func() func() g.Node {
        // This function is called only when the component is first rendered
        heavyData := loadHeavyData()
        
        return func() g.Node {
            return g.Div(
                g.H2(g.Text("Heavy Section")),
                renderHeavyContent(heavyData),
            )
        }
    })
}
```

### Optimizing List Updates

```go
// Good: Stable keys for efficient reconciliation
comps.For(comps.ForProps[User]{
    Items: users,
    Key: func(user User) string {
        return user.ID // Stable, unique identifier
    },
    Children: func(user User, index int) g.Node {
        return renderUserCard(user)
    },
})

// Avoid: Using index as key (causes unnecessary re-renders)
comps.For(comps.ForProps[User]{
    Items: users,
    Key: func(user User) string {
        return strconv.Itoa(index) // Bad: index changes when list reorders
    },
    Children: func(user User, index int) g.Node {
        return renderUserCard(user)
    },
})
```

## Error Handling

### ErrorBoundary for Graceful Failures

```go
func renderUserDashboard(userID string) g.Node {
    return comps.ErrorBoundary(comps.ErrorBoundaryProps{
        Fallback: func(err error) g.Node {
            logutil.Logf("Dashboard error: %v", err)
            return g.Div(
                g.Class("error-boundary"),
                g.H3(g.Text("Something went wrong")),
                g.P(g.Text("We're sorry, but the dashboard failed to load.")),
                g.Button(
                    g.Text("Retry"),
                    dom.OnClick(func() {
                        // Trigger re-render or reload
                        window.Location().Reload()
                    }),
                ),
            )
        },
        Children: g.Div(
            renderUserProfile(userID),
            renderUserStats(userID),
            renderUserActivity(userID),
        ),
    })
}
```

## Advanced Patterns

### Portal for Modals and Overlays

```go
func renderModal(isOpen reactivity.Signal[bool], content g.Node) g.Node {
    return comps.Show(comps.ShowProps{
        When: isOpen,
        Children: comps.Portal("modal-root", g.Div(
            g.Class("modal-overlay"),
            dom.OnClick(func() {
                isOpen.Set(false)
            }),
            g.Div(
                g.Class("modal-content"),
                dom.OnClick(func(e js.Value) {
                    e.Call("stopPropagation") // Prevent closing when clicking content
                }),
                content,
            ),
        )),
    })
}
```

### Compound Components

```go
type TabsState struct {
    activeTab reactivity.Signal[string]
    tabs      []TabConfig
}

type TabConfig struct {
    ID      string
    Label   string
    Content g.Node
}

func (s *TabsState) render() g.Node {
    return g.Div(
        g.Class("tabs"),
        // Tab headers
        g.Div(
            g.Class("tab-headers"),
            comps.For(comps.ForProps[TabConfig]{
                Items: reactivity.NewSignal(s.tabs),
                Key: func(tab TabConfig) string { return tab.ID },
                Children: func(tab TabConfig, index int) g.Node {
                    isActive := reactivity.NewMemo(func() bool {
                        return s.activeTab.Get() == tab.ID
                    })
                    
                    return g.Button(
                        g.Class("tab-header"),
                        g.If(isActive.Get(), g.Class("active")),
                        g.Text(tab.Label),
                        dom.OnClick(func() {
                            s.activeTab.Set(tab.ID)
                        }),
                    )
                },
            }),
        ),
        // Tab content
        g.Div(
            g.Class("tab-content"),
            comps.Switch(comps.SwitchProps{
                When: s.activeTab,
                Children: func() []g.Node {
                    var matches []g.Node
                    for _, tab := range s.tabs {
                        matches = append(matches, comps.Match(comps.MatchProps{
                            When: tab.ID,
                            Children: tab.Content,
                        }))
                    }
                    return matches
                }(),
            }),
        ),
    )
}
```

### Higher-Order Components

```go
// WithLoading HOC
func WithLoading(isLoading reactivity.Signal[bool], component g.Node) g.Node {
    return comps.Switch(comps.SwitchProps{
        When: isLoading,
        Children: []g.Node{
            comps.Match(comps.MatchProps{
                When: true,
                Children: g.Div(
                    g.Class("loading-spinner"),
                    g.Text("Loading..."),
                ),
            }),
            comps.Match(comps.MatchProps{
                When: false,
                Children: component,
            }),
        },
    })
}

// WithPermissions HOC
func WithPermissions(hasPermission reactivity.Signal[bool], component g.Node, fallback g.Node) g.Node {
    return comps.Show(comps.ShowProps{
        When: hasPermission,
        Children: component,
    })
    // Note: For fallback, you'd need a more complex pattern or use Switch
}
```

## Best Practices

### 1. Signal Management

```go
// Good: Centralized state management
type AppState struct {
    user     reactivity.Signal[*User]
    theme    reactivity.Signal[Theme]
    settings reactivity.Signal[Settings]
}

// Good: Computed signals for derived state
func (s *AppState) isDarkMode() reactivity.Signal[bool] {
    return reactivity.NewMemo(func() bool {
        return s.theme.Get() == ThemeDark
    })
}

// Avoid: Creating signals in render functions
func badRender() g.Node {
    // Bad: Creates new signal on every render
    counter := reactivity.NewSignal(0)
    return g.Div(g.Text(strconv.Itoa(counter.Get())))
}
```

### 2. Key Functions

```go
// Good: Stable, unique keys
Key: func(item Product) string {
    return item.ID
}

// Good: Composite keys when needed
Key: func(item OrderItem) string {
    return fmt.Sprintf("%s-%s", item.OrderID, item.ProductID)
}

// Avoid: Non-unique or unstable keys
Key: func(item Product) string {
    return item.Name // Bad: names might not be unique
}

Key: func(item Product) string {
    return strconv.Itoa(rand.Int()) // Bad: changes every time
}
```

### 3. Component Composition

```go
// Good: Small, focused components
func renderUserAvatar(user *User) g.Node {
    return g.Img(
        g.Class("avatar"),
        g.Src(user.AvatarURL),
        g.Alt(user.Name),
    )
}

func renderUserCard(user *User) g.Node {
    return g.Div(
        g.Class("user-card"),
        renderUserAvatar(user),
        g.H3(g.Text(user.Name)),
        g.P(g.Text(user.Email)),
    )
}

// Good: Reusable helper components
func renderLoadingState(message string) g.Node {
    return g.Div(
        g.Class("loading"),
        g.Text(message),
    )
}
```

### 4. Error Handling

```go
// Good: Wrap potentially failing components
func renderDataVisualization(data reactivity.Signal[[]DataPoint]) g.Node {
    return comps.ErrorBoundary(comps.ErrorBoundaryProps{
        Fallback: func(err error) g.Node {
            return g.Div(
                g.Class("error-state"),
                g.Text("Failed to render chart"),
            )
        },
        Children: comps.Memo(func() g.Node {
            return renderComplexChart(data.Get())
        }, data),
    })
}
```

## Common Pitfalls

### 1. Signal Creation in Render

```go
// Bad: Creates new signal on every render
func badComponent() g.Node {
    counter := reactivity.NewSignal(0) // New signal every time!
    return g.Button(
        g.Text(strconv.Itoa(counter.Get())),
        dom.OnClick(func() {
            counter.Set(counter.Get() + 1)
        }),
    )
}

// Good: Signal created once and reused
type CounterComponent struct {
    counter reactivity.Signal[int]
}

func NewCounterComponent() *CounterComponent {
    return &CounterComponent{
        counter: reactivity.NewSignal(0),
    }
}

func (c *CounterComponent) render() g.Node {
    return g.Button(
        g.Text(strconv.Itoa(c.counter.Get())),
        dom.OnClick(func() {
            c.counter.Set(c.counter.Get() + 1)
        }),
    )
}
```

### 2. Inefficient List Keys

```go
// Bad: Using array index as key
comps.For(comps.ForProps[Item]{
    Items: items,
    Key: func(item Item) string {
        return strconv.Itoa(index) // Index not available here anyway!
    },
    Children: func(item Item, index int) g.Node {
        return renderItem(item)
    },
})

// Good: Using stable item property
comps.For(comps.ForProps[Item]{
    Items: items,
    Key: func(item Item) string {
        return item.ID
    },
    Children: func(item Item, index int) g.Node {
        return renderItem(item)
    },
})
```

### 3. Memory Leaks with Effects

```go
// Bad: Effect not cleaned up
func badComponent() g.Node {
    reactivity.NewEffect(func() {
        // This effect will never be disposed!
        logutil.Log("Effect running")
    })
    return g.Div()
}

// Good: Effect cleaned up properly
func goodComponent() g.Node {
    return comps.OnMount(func() {
        effect := reactivity.NewEffect(func() {
            logutil.Log("Effect running")
        })
        
        comps.OnCleanup(func() {
            effect.Dispose()
        })
    })
}
```

### 4. Overusing Memo

```go
// Bad: Memoizing simple operations
func badMemo(name reactivity.Signal[string]) g.Node {
    return comps.Memo(func() g.Node {
        return g.Text(name.Get()) // Too simple to memo
    }, name)
}

// Good: Memoizing expensive operations
func goodMemo(data reactivity.Signal[[]DataPoint]) g.Node {
    return comps.Memo(func() g.Node {
        // Expensive chart rendering
        return renderComplexChart(data.Get())
    }, data)
}
```

## Related Documentation

- **[Quick Reference](./quick-reference.md)** - Concise syntax reference for all helper functions
- **[Real-World Examples](./real-world-examples.md)** - Practical applications and usage patterns
- **[Integration Examples](./integration-examples.md)** - Complex scenarios with multiple helpers working together
- **[Performance Optimization](./performance-optimization.md)** - Performance best practices for helper functions
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions
- **[Control Flow](./control-flow.md)** - Core concepts of reactive control flow

This guide provides a comprehensive foundation for using UIwGo's helper functions effectively. Remember to always consider performance, maintainability, and user experience when choosing patterns and implementing components.

For more detailed examples and patterns, see the [Real-World Examples](./real-world-examples.md) guide.


## Alpine-inspired inline helpers

These helpers bring Alpine.js-like ergonomics to inline events and lifecycle, while staying idiomatic to UIwGo and gomponents. They all return gomponents attributes so you can attach them directly to element builders. See the design notes for more options and rationale in [Alpine-inspired inline events and lifecycle hooks](../alpine_inline_events.md).

- OnInitInline(handler func(el dom.Element))
  - Run once shortly after the element is connected to the DOM (x-init style)
  - Schedules on a microtask to avoid layout thrash
  - Example:
    ```go
    g.Div(
      dom.OnInitInline(func(el dom.Element) {
        // do one-time setup
      }),
      g.Text("Ready")
    )
    ```

- OnDestroyInline(handler func(el dom.Element))
  - Run when the element is removed from the DOM
  - Uses MutationObserver under the hood and auto-cleans
  - Example:
    ```go
    g.Div(
      dom.OnDestroyInline(func(el dom.Element) {
        // cleanup timers/listeners
      }),
    )
    ```

- OnVisibleInline(handler func(el dom.Element))
  - Fire once when the element first becomes visible in the viewport
  - Uses IntersectionObserver; automatically unobserves after first fire
  - Example:
    ```go
    g.Div(
      dom.OnVisibleInline(func(el dom.Element) {
        // lazy-load, start animations, etc.
      }),
    )
    ```

- OnResizeInline(handler func(el dom.Element))
  - Listen to element size changes
  - Uses ResizeObserver; cleans up with scope
  - Example:
    ```go
    g.Div(
      dom.OnResizeInline(func(el dom.Element) {
        // react to size changes
      }),
    )
    ```

- OnClickOnceInline(handler func(el dom.Element))
  - Convenience for a click handler that runs only once per element
  - Automatically removes its attribute and handler after running
  - Example:
    ```go
    g.Button(
      g.Text("Once"),
      dom.OnClickOnceInline(func(el dom.Element) {
        // runs a single time
      }),
    )
    ```

Notes
- All helpers follow our JS/DOM interop guideline: prefer honnef.co/go/js/dom/v2 typed elements (re-exported as dom.Element)
- Handlers are tied to the current cleanup scope and are disposed automatically
- For additional planned modifiers (debounce, throttle, outside, capture, etc.), see the design doc linked above
