# Control Flow Guide

This guide covers how to implement conditional rendering, loops, dynamic content, and other control flow patterns in UIwGo. Learn to build dynamic UIs that respond to state changes and user interactions.

## Table of Contents

- [Overview](#overview)
- [Conditional Rendering](#conditional-rendering)
- [Lists and Loops](#lists-and-loops)
- [Dynamic Content](#dynamic-content)
- [Nested Components](#nested-components)
- [Advanced Patterns](#advanced-patterns)
- [Performance Considerations](#performance-considerations)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)

## Overview

UIwGo handles control flow through reactive patterns rather than traditional templating. Instead of template directives like `v-if` or `*ngFor`, you use:

- **Reactive signals** to control visibility and content
- **Memos** to compute dynamic HTML
- **Effects** to update DOM based on state changes
- **Component composition** for complex structures

### Mental Model

```go
// Traditional templating (conceptual)
// <div v-if="showMessage">{{ message }}</div>
// <li v-for="item in items">{{ item.name }}</li>

// UIwGo approach
showMessage := reactivity.CreateSignal(true)
message := reactivity.CreateSignal("Hello")
items := reactivity.CreateSignal([]Item{})

// Reactive HTML generation using g.If
return g.If(showMessage.Get(),
    g.Div(comps.BindText(message)),
    g.Empty(),
)
```

## Conditional Rendering

### Basic Conditional Display

Show or hide elements based on state:

```go
func ConditionalComponent() g.Node {
    isVisible := reactivity.CreateSignal(false)
    message := reactivity.CreateSignal("Hello, World!")
    
    return g.Div(
        g.Class("conditional-container"),
        g.Button(
            g.Text("Toggle Message"),
            dom.OnClickInline(func(el dom.Element) {
                isVisible.Set(!isVisible.Get())
            }),
        ),
        g.If(isVisible.Get(),
            g.P(
                g.Class("message"),
                comps.BindText(message),
            ),
            g.P(
                g.Class("placeholder"),
                g.Text("Nothing to show"),
            ),
        ),
    )
}
```

### Multiple Conditions

Handle multiple states with switch-like logic:

```go
type Status int

const (
    StatusLoading Status = iota
    StatusSuccess
    StatusError
    StatusEmpty
)

func StatusComponent() g.Node {
    status := reactivity.CreateSignal(StatusLoading)
    data := reactivity.CreateSignal([]string{})
    error := reactivity.CreateSignal[error](nil)
    
    content := reactivity.CreateMemo(func() g.Node {
        switch status.Get() {
        case StatusLoading:
            return g.Div(
                g.Class("loading"),
                g.Text("Loading..."),
            )
            
        case StatusError:
            err := error.Get()
            if err != nil {
                return g.Div(
                    g.Class("error"),
                    g.Text(fmt.Sprintf("Error: %s", err.Error())),
                )
            }
            return g.Div(
                g.Class("error"),
                g.Text("Unknown error occurred"),
            )
            
        case StatusEmpty:
            return g.Div(
                g.Class("empty"),
                g.Text("No data available"),
            )
            
        case StatusSuccess:
            items := data.Get()
            if len(items) == 0 {
                return g.Div(
                    g.Class("empty"),
                    g.Text("No items found"),
                )
            }
            
            listItems := make([]g.Node, len(items))
            for i, item := range items {
                listItems[i] = g.Li(g.Text(item))
            }
            return g.Ul(
                g.Class("data-list"),
                listItems...,
            )
            
        default:
            return g.Div(
                g.Class("unknown"),
                g.Text("Unknown status"),
            )
        }
    })
    
    loadData := func() {
        status.Set(StatusLoading)
        // Simulate async data loading
        go func() {
            time.Sleep(1 * time.Second)
            data.Set([]string{"Item 1", "Item 2", "Item 3"})
            status.Set(StatusSuccess)
        }()
    }
    
    clearData := func() {
        data.Set([]string{})
        status.Set(StatusEmpty)
    }
    
    triggerError := func() {
        error.Set(fmt.Errorf("simulated error"))
        status.Set(StatusError)
    }
    
    return g.Div(
        g.Class("status-container"),
        g.Div(
            g.Class("controls"),
            g.Button(
                g.Text("Load Data"),
                dom.OnClickInline(func(el dom.Element) { loadData() }),
            ),
            g.Button(
                g.Text("Clear"),
                dom.OnClickInline(func(el dom.Element) { clearData() }),
            ),
            g.Button(
                g.Text("Trigger Error"),
                dom.OnClickInline(func(el dom.Element) { triggerError() }),
            ),
        ),
        comps.BindNode(content),
    )
}


```

### Conditional Classes and Attributes

Dynamically apply CSS classes and attributes:

```go
func ButtonComponent() g.Node {
    isActive := reactivity.CreateSignal(false)
    isLoading := reactivity.CreateSignal(false)
    text := reactivity.CreateSignal("Click me")
    
    buttonClass := reactivity.CreateMemo(func() string {
        classes := []string{"btn"}
        
        if isActive.Get() {
            classes = append(classes, "btn-active")
        }
        
        if isLoading.Get() {
            classes = append(classes, "btn-loading")
        }
        
        return strings.Join(classes, " ")
    })
    
    handleClick := func() {
        if isLoading.Get() {
            return // Don't handle clicks when loading
        }
        
        isActive.Set(!isActive.Get())
        
        // Simulate async operation
        if isActive.Get() {
            isLoading.Set(true)
            text.Set("Loading...")
            
            go func() {
                time.Sleep(2 * time.Second)
                isLoading.Set(false)
                text.Set("Completed!")
            }()
        } else {
            text.Set("Click me")
        }
    }
    
    return g.Button(
        comps.BindClass(buttonClass),
        g.If(isLoading.Get(), g.Attr("disabled", "disabled")),
        comps.BindText(text),
        dom.OnClickInline(func(el dom.Element) {
            handleClick()
        }),
    )
}


```

## Lists and Loops

### Basic List Rendering

Render dynamic lists from arrays:

```go
type Item struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Done bool   `json:"done"`
}

func TodoList() g.Node {
    items := reactivity.CreateSignal([]Item{})
    newItem := reactivity.CreateSignal("")
    nextID := reactivity.CreateSignal(1)
    
    listContent := reactivity.CreateMemo(func() g.Node {
        itemList := items.Get()
        
        if len(itemList) == 0 {
            return g.P(
                g.Class("empty"),
                g.Text("No items yet. Add one above!"),
            )
        }
        
        listItems := make([]g.Node, len(itemList))
        for i, item := range itemList {
            listItems[i] = g.Li(
                g.Class("todo-item"),
                g.Attr("data-id", fmt.Sprintf("%d", item.ID)),
                g.Input(
                    g.Attr("type", "checkbox"),
                    g.If(item.Done, g.Attr("checked", "checked")),
                    dom.OnChangeInline(func(el dom.Element) {
                        toggleItem(item.ID)
                    }),
                ),
                g.Span(
                    g.Class(map[bool]string{true: "done", false: "pending"}[item.Done]),
                    g.Text(item.Name),
                ),
                g.Button(
                    g.Text("Delete"),
                    dom.OnClickInline(func(el dom.Element) {
                        deleteItem(item.ID)
                    }),
                ),
            )
        }
        
        return g.Ul(
            g.Class("todo-list"),
            listItems...,
        )
    })
    
    addItem := func() {
        text := strings.TrimSpace(newItem.Get())
        if text == "" {
            return
        }
        
        currentItems := items.Get()
        newItemObj := Item{
            ID:   nextID.Get(),
            Name: text,
            Done: false,
        }
        
        items.Set(append(currentItems, newItemObj))
        nextID.Set(nextID.Get() + 1)
        newItem.Set("")
    }
    
    toggleItem := func(id int) {
        currentItems := items.Get()
        for i, item := range currentItems {
            if item.ID == id {
                currentItems[i].Done = !currentItems[i].Done
                break
            }
        }
        items.Set(currentItems)
    }
    
    deleteItem := func(id int) {
        currentItems := items.Get()
        newItems := make([]Item, 0, len(currentItems))
        for _, item := range currentItems {
            if item.ID != id {
                newItems = append(newItems, item)
            }
        }
        items.Set(newItems)
    }
    
    return g.Div(
        g.Class("todo-container"),
        g.Div(
            g.Class("add-item"),
            g.Input(
                g.Attr("type", "text"),
                g.Attr("placeholder", "Add new item..."),
                comps.BindValue(newItem),
                dom.OnKeyDownInline(func(el dom.Element, event dom.Event) {
                    if event.(*dom.KeyboardEvent).Key == "Enter" {
                        addItem()
                    }
                }),
            ),
            g.Button(
                g.Text("Add"),
                dom.OnClickInline(func(el dom.Element) {
                    addItem()
                }),
            ),
        ),
        comps.BindNode(listContent),
    )
}

```

### Filtered and Sorted Lists

Implement filtering and sorting:

```go
func FilteredList() g.Node {
    allItems := reactivity.CreateSignal([]Item{})
    filter := reactivity.CreateSignal("")
    sortBy := reactivity.CreateSignal("name")
    showDoneOnly := reactivity.CreateSignal(false)
    
    filteredItems := reactivity.CreateMemo(func() []Item {
        items := allItems.Get()
        filterText := strings.ToLower(filter.Get())
        sortByValue := sortBy.Get()
        showDoneOnlyValue := showDoneOnly.Get()
        
        // Filter items
        var filtered []Item
        for _, item := range items {
            // Text filter
            if filterText != "" && !strings.Contains(strings.ToLower(item.Name), filterText) {
                continue
            }
            
            // Done filter
            if showDoneOnlyValue && !item.Done {
                continue
            }
            
            filtered = append(filtered, item)
        }
        
        // Sort items
        switch sortByValue {
        case "name":
            sort.Slice(filtered, func(i, j int) bool {
                return filtered[i].Name < filtered[j].Name
            })
        case "status":
            sort.Slice(filtered, func(i, j int) bool {
                if filtered[i].Done != filtered[j].Done {
                    return !filtered[i].Done // Pending items first
                }
                return filtered[i].Name < filtered[j].Name
            })
        }
        
        return filtered
    })
    
    listContent := reactivity.CreateMemo(func() g.Node {
        items := filteredItems.Get()
        
        if len(items) == 0 {
            return g.P(
                g.Class("empty"),
                g.Text("No items match the current filter."),
            )
        }
        
        listItems := make([]g.Node, len(items))
        for i, item := range items {
            status := "pending"
            if item.Done {
                status = "done"
            }
            
            listItems[i] = g.Li(
                g.Class("item "+status),
                g.Span(
                    g.Class("name"),
                    g.Text(item.Name),
                ),
                g.Span(
                    g.Class("status"),
                    g.Text(status),
                ),
            )
        }
        
        return g.Ul(
            g.Class("filtered-list"),
            listItems...,
        )
    })
    
    return g.Div(
        g.Class("filtered-list-container"),
        g.Div(
            g.Class("filters"),
            g.Input(
                g.Attr("type", "text"),
                g.Attr("placeholder", "Filter items..."),
                comps.BindValue(filter),
            ),
            g.Select(
                comps.BindValue(sortBy),
                g.Option(
                    g.Attr("value", "name"),
                    g.Text("Sort by Name"),
                ),
                g.Option(
                    g.Attr("value", "status"),
                    g.Text("Sort by Status"),
                ),
            ),
            g.Label(
                g.Input(
                    g.Attr("type", "checkbox"),
                    comps.BindChecked(showDoneOnly),
                ),
                g.Text(" Show completed only"),
            ),
        ),
        comps.BindNode(listContent),
    )
}


```

## Dynamic Content

### Content Switching

Switch between different content types:

```go
type ContentType string

const (
    ContentText  ContentType = "text"
    ContentImage ContentType = "image"
    ContentVideo ContentType = "video"
    ContentList  ContentType = "list"
)

func DynamicContent() g.Node {
    contentType := reactivity.CreateSignal(ContentText)
    textData := reactivity.CreateSignal("Sample text content")
    imageURL := reactivity.CreateSignal("https://via.placeholder.com/300x200")
    videoURL := reactivity.CreateSignal("https://www.w3schools.com/html/mov_bbb.mp4")
    listData := reactivity.CreateSignal([]string{"Item 1", "Item 2", "Item 3"})
    
    content := reactivity.CreateMemo(func() g.Node {
        switch contentType.Get() {
        case ContentText:
            return g.Div(
                g.Class("text-content"),
                g.H3(g.Text("Text Content")),
                g.P(comps.BindText(textData)),
                g.Textarea(
                    g.Attr("rows", "4"),
                    g.Attr("cols", "50"),
                    comps.BindValue(textData),
                ),
            )
            
        case ContentImage:
            return g.Div(
                g.Class("image-content"),
                g.H3(g.Text("Image Content")),
                g.Img(
                    comps.BindAttr("src", imageURL),
                    g.Attr("alt", "Dynamic image"),
                    g.Attr("style", "max-width: 100%;"),
                ),
                g.Input(
                    g.Attr("type", "url"),
                    g.Attr("placeholder", "Image URL"),
                    comps.BindValue(imageURL),
                ),
            )
            
        case ContentVideo:
            return g.Div(
                g.Class("video-content"),
                g.H3(g.Text("Video Content")),
                g.El("video",
                    g.Attr("controls", "controls"),
                    g.Attr("style", "max-width: 100%;"),
                    g.El("source",
                        comps.BindAttr("src", videoURL),
                        g.Attr("type", "video/mp4"),
                    ),
                    g.Text("Your browser does not support the video tag."),
                ),
                g.Input(
                    g.Attr("type", "url"),
                    g.Attr("placeholder", "Video URL"),
                    comps.BindValue(videoURL),
                ),
            )
            
        case ContentList:
            listItems := listData.Get()
            items := make([]g.Node, len(listItems))
            for i, item := range listItems {
                itemSignal := reactivity.CreateSignal(item)
                items[i] = g.Li(
                    g.Input(
                        g.Attr("type", "text"),
                        comps.BindValue(itemSignal),
                    ),
                    g.Button(
                        g.Text("Remove"),
                        dom.OnClickInline(func(el dom.Element) {
                            removeListItem(i)
                        }),
                    ),
                )
            }
            
            return g.Div(
                g.Class("list-content"),
                g.H3(g.Text("List Content")),
                g.Ul(items...),
                g.Button(
                    g.Text("Add Item"),
                    dom.OnClickInline(func(el dom.Element) {
                        addListItem()
                    }),
                ),
            )
            
        default:
            return g.Div(
                g.Class("unknown-content"),
                g.Text("Unknown content type"),
            )
        }
    })
    
    // Helper functions
    addListItem := func() {
        current := listData.Get()
        listData.Set(append(current, "New item"))
    }
    
    removeListItem := func(index int) {
        current := listData.Get()
        if index >= 0 && index < len(current) {
            newList := make([]string, 0, len(current)-1)
            newList = append(newList, current[:index]...)
            newList = append(newList, current[index+1:]...)
            listData.Set(newList)
        }
    }
    
    return g.Div(
        g.Class("dynamic-container"),
        g.Div(
            g.Class("content-selector"),
            g.H2(g.Text("Dynamic Content Demo")),
            g.Div(
                g.Class("tabs"),
                g.Button(
                    g.Text("Text"),
                    comps.BindClass("active", reactivity.CreateMemo(func() bool {
                        return contentType.Get() == ContentText
                    })),
                    dom.OnClickInline(func(el dom.Element) {
                        contentType.Set(ContentText)
                    }),
                ),
                g.Button(
                    g.Text("Image"),
                    comps.BindClass("active", reactivity.CreateMemo(func() bool {
                        return contentType.Get() == ContentImage
                    })),
                    dom.OnClickInline(func(el dom.Element) {
                        contentType.Set(ContentImage)
                    }),
                ),
                g.Button(
                    g.Text("Video"),
                    comps.BindClass("active", reactivity.CreateMemo(func() bool {
                        return contentType.Get() == ContentVideo
                    })),
                    dom.OnClickInline(func(el dom.Element) {
                        contentType.Set(ContentVideo)
                    }),
                ),
                g.Button(
                    g.Text("List"),
                    comps.BindClass("active", reactivity.CreateMemo(func() bool {
                        return contentType.Get() == ContentList
                    })),
                    dom.OnClickInline(func(el dom.Element) {
                        contentType.Set(ContentList)
                    }),
                ),
            ),
        ),
        g.Div(
            g.Class("content-area"),
            comps.BindNode(content),
        ),
    )
}


```

## Nested Components

### Component Composition

Build complex UIs by composing smaller components:

```go
// Child component
func Card(initialTitle, initialContent string) g.Node {
    title := reactivity.CreateSignal(initialTitle)
    content := reactivity.CreateSignal(initialContent)
    visible := reactivity.CreateSignal(true)
    
    return reactivity.CreateMemo(func() g.Node {
        if !visible.Get() {
            return g.Text("")
        }
        
        return g.Div(
            g.Class("card"),
            g.Div(
                g.Class("card-header"),
                g.H3(comps.BindText(title)),
                g.Button(
                    g.Text("×"),
                    dom.OnClickInline(func(el dom.Element) {
                        visible.Set(false)
                    }),
                ),
            ),
            g.Div(
                g.Class("card-body"),
                comps.BindText(content),
            ),
        )
    })
}

// Parent component that manages multiple cards
func Dashboard() g.Node {
    newTitle := reactivity.CreateSignal("")
    newContent := reactivity.CreateSignal("")
    
    // Initialize with some default cards
    cards := reactivity.CreateSignal([]struct{
        title   string
        content string
        id      int
    }{
        {"Welcome", "Welcome to the dashboard!", 0},
        {"Stats", "Your statistics will appear here.", 1},
        {"News", "Latest news and updates.", 2},
    })
    
    nextID := reactivity.CreateSignal(3)
    
    addCard := func() {
        title := strings.TrimSpace(newTitle.Get())
        content := strings.TrimSpace(newContent.Get())
        
        if title == "" || content == "" {
            return
        }
        
        currentCards := cards.Get()
        id := nextID.Get()
        
        newCard := struct{
            title   string
            content string
            id      int
        }{title, content, id}
        
        cards.Set(append(currentCards, newCard))
        nextID.Set(id + 1)
        
        // Clear form
        newTitle.Set("")
        newContent.Set("")
    }
    
    removeCard := func(id int) {
        currentCards := cards.Get()
        newCards := make([]struct{
            title   string
            content string
            id      int
        }, 0, len(currentCards))
        
        for _, card := range currentCards {
            if card.id != id {
                newCards = append(newCards, card)
            }
        }
        
        cards.Set(newCards)
    }
    
    cardsGrid := reactivity.CreateMemo(func() g.Node {
        currentCards := cards.Get()
        cardNodes := make([]g.Node, len(currentCards))
        
        for i, cardData := range currentCards {
            cardNodes[i] = g.Div(
                g.Class("card-wrapper"),
                Card(cardData.title, cardData.content),
            )
        }
        
        return g.Div(
            g.Class("cards-grid"),
            cardNodes...,
        )
    })
    
    return g.Div(
        g.Class("dashboard"),
        g.Div(
            g.Class("dashboard-header"),
            g.H1(g.Text("Dashboard")),
            g.Div(
                g.Class("add-card-form"),
                g.Input(
                    g.Attr("type", "text"),
                    g.Attr("placeholder", "Card title"),
                    comps.BindValue(newTitle),
                ),
                g.Input(
                    g.Attr("type", "text"),
                    g.Attr("placeholder", "Card content"),
                    comps.BindValue(newContent),
                ),
                g.Button(
                    g.Text("Add Card"),
                    dom.OnClickInline(func(el dom.Element) {
                        addCard()
                    }),
                ),
            ),
        ),
        comps.BindNode(cardsGrid),
    )
```

## Advanced Patterns

### Lazy Loading

Load content only when needed:

```go
func LazySection() g.Node {
    isExpanded := reactivity.CreateSignal(false)
    isLoaded := reactivity.CreateSignal(false)
    data := reactivity.CreateSignal("")
    loading := reactivity.CreateSignal(false)
    
    loadContent := func() {
        if isLoaded.Get() || loading.Get() {
            return
        }
        
        loading.Set(true)
        
        // Simulate async loading
        go func() {
            time.Sleep(2 * time.Second)
            data.Set("This is the lazily loaded content! It contains important information that was fetched asynchronously.")
            loading.Set(false)
            isLoaded.Set(true)
        }()
    }
    
    toggleExpanded := func() {
        expanded := !isExpanded.Get()
        isExpanded.Set(expanded)
        
        if expanded && !isLoaded.Get() {
            loadContent()
        }
    }
    
    content := reactivity.CreateMemo(func() g.Node {
        if !isExpanded.Get() {
            return g.P(g.Text("Click to expand..."))
        }
        
        if loading.Get() {
            return g.P(g.Text("Loading content..."))
        }
        
        if !isLoaded.Get() {
            return g.P(g.Text("Content not loaded yet."))
        }
        
        return g.Div(
            g.Class("loaded-content"),
            g.Text(data.Get()),
        )
    })
    
    return g.Div(
        g.Class("lazy-section"),
        g.Button(
            g.Text("Toggle Section"),
            dom.OnClickInline(func(el dom.Element) {
                toggleExpanded()
            }),
        ),
        comps.BindNode(content),
    )

```

### Virtual Scrolling (Concept)

For very large lists, implement virtual scrolling:

```go
func VirtualList(itemHeight, visibleCount int, items []string) g.Node {
    allItems := reactivity.CreateSignal(items)
    scrollTop := reactivity.CreateSignal(0)
    
    startIndex := reactivity.CreateMemo(func() int {
        scrollTop := scrollTop.Get()
        index := scrollTop / itemHeight
        if index < 0 {
            return 0
        }
        return index
    })
    
    endIndex := reactivity.CreateMemo(func() int {
        start := startIndex.Get()
        end := start + visibleCount
        allItems := allItems.Get()
        if end > len(allItems) {
            return len(allItems)
        }
        return end
    })
    
    visibleItems := reactivity.CreateMemo(func() []g.Node {
        allItems := allItems.Get()
        start := startIndex.Get()
        end := endIndex.Get()
        
        if start >= len(allItems) {
            return []g.Node{}
        }
        
        items := allItems[start:end]
        nodes := make([]g.Node, len(items))
        for i, item := range items {
            nodes[i] = g.Div(
                g.Class("virtual-item"),
                g.Style(fmt.Sprintf("height: %dpx;", itemHeight)),
                g.Text(item),
            )
        }
        return nodes
    })
    
    handleScroll := func(el dom.Element) {
        scrollTop.Set(el.ScrollTop())
    }
    
    return g.Div(
        g.Class("virtual-list"),
        g.Style(fmt.Sprintf("height: %dpx; overflow-y: auto;", visibleCount*itemHeight)),
        dom.OnScrollInline(handleScroll),
        g.Div(
            g.Style(fmt.Sprintf("height: %dpx; position: relative;", len(allItems.Get())*itemHeight)),
            g.Div(
                g.Style(fmt.Sprintf("position: absolute; top: %dpx;", startIndex.Get()*itemHeight)),
                comps.BindNodes(visibleItems),
            ),
        ),
    )
}

// This provides a basic virtual scrolling implementation
// Full production implementation would include additional optimizations
```

## Performance Considerations

### Optimize Memo Dependencies

```go
// BAD: Memo depends on entire large object
expensiveMemo := reactivity.CreateMemo(func() string {
    user := largeUserObject.Get() // Entire object dependency
    return user.Name // Only need name
})

// GOOD: Extract specific fields
userName := reactivity.CreateMemo(func() string {
    user := largeUserObject.Get()
    return user.Name
})

displayName := reactivity.CreateMemo(func() string {
    name := userName.Get() // Depends only on name
    return fmt.Sprintf("Hello, %s!", name)
})
```

### Batch DOM Updates

```go
// BAD: Multiple individual updates
func updateMultipleFields() {
    name.Set("John")
    email.Set("john@example.com")
    age.Set(30)
    // Each triggers separate DOM updates
}

// GOOD: Batch updates
func updateMultipleFields() {
    reactivity.Batch(func() {
        name.Set("John")
        email.Set("john@example.com")
        age.Set(30)
    })
    // Single DOM update after batch
}
```

### Debounce Expensive Operations

```go
func SearchComponent() g.Node {
    query := reactivity.CreateSignal("")
    results := reactivity.CreateSignal([]string{})
    
    // Debounced search effect
    var searchTimer *time.Timer
    reactivity.CreateEffect(func() {
        queryValue := query.Get()
        
        // Cancel previous timer
        if searchTimer != nil {
            searchTimer.Stop()
        }
        
        // Set new timer
        searchTimer = time.AfterFunc(300*time.Millisecond, func() {
            performSearch(queryValue)
        })
    })
    
    performSearch := func(query string) {
        if query == "" {
            results.Set([]string{})
            return
        }
        
        // Simulate API call
        go func() {
            time.Sleep(100 * time.Millisecond)
            
            // Mock search results
            searchResults := []string{
                fmt.Sprintf("Result 1 for '%s'", query),
                fmt.Sprintf("Result 2 for '%s'", query),
                fmt.Sprintf("Result 3 for '%s'", query),
            }
            
            results.Set(searchResults)
        }()
    }
    
    resultsList := reactivity.CreateMemo(func() g.Node {
        items := results.Get()
        if len(items) == 0 {
            return g.P(g.Text("No results"))
        }
        
        nodes := make([]g.Node, len(items))
        for i, item := range items {
            nodes[i] = g.Li(g.Text(item))
        }
        return g.Ul(nodes...)
    })
    
    return g.Div(
        g.Class("search-component"),
        g.Input(
            g.Type("text"),
            g.Placeholder("Search..."),
            comps.BindValue(query),
            dom.OnInputInline(func(el dom.Element) {
                query.Set(el.Value())
            }),
        ),
        comps.BindNode(resultsList),
    )
}
```

## Common Patterns

### Toggle Pattern

```go
func Toggle(initial bool) g.Node {
    isOn := reactivity.CreateSignal(initial)
    
    toggleState := func() {
        isOn.Set(!isOn.Get())
    }
    
    return g.Button(
        g.Class("toggle"),
        comps.BindClass("on", reactivity.CreateMemo(func() bool {
            return isOn.Get()
        })),
        comps.BindText(reactivity.CreateMemo(func() string {
            if isOn.Get() {
                return "ON"
            }
            return "OFF"
        })),
        dom.OnClickInline(func(el dom.Element) {
            toggleState()
        }),
    )
    }
    

```

### Counter Pattern

```go
func Counter(initial, min, max int) g.Node {
    count := reactivity.CreateSignal(initial)
    
    increment := func() {
        if count.Get() < max {
            count.Set(count.Get() + 1)
        }
    }
    
    decrement := func() {
        if count.Get() > min {
            count.Set(count.Get() - 1)
        }
    }
    
    return g.Div(
        g.Class("counter"),
        g.Button(
            g.Text("-"),
            comps.BindDisabled(reactivity.CreateMemo(func() bool {
                return count.Get() <= min
            })),
            dom.OnClickInline(func(el dom.Element) {
                decrement()
            }),
        ),
        g.Span(
            g.Class("count"),
            comps.BindText(reactivity.CreateMemo(func() string {
                return fmt.Sprintf("%d", count.Get())
            })),
        ),
        g.Button(
            g.Text("+"),
            comps.BindDisabled(reactivity.CreateMemo(func() bool {
                return count.Get() >= max
            })),
            dom.OnClickInline(func(el dom.Element) {
                increment()
            }),
        ),
    )
}


```

### Form Validation Pattern

```go
type ValidationRule func(string) string

func ValidatedField(initial string, placeholder string, rules ...ValidationRule) g.Node {
    value := reactivity.CreateSignal(initial)
    
    error := reactivity.CreateMemo(func() string {
        currentValue := value.Get()
        
        for _, rule := range rules {
            if err := rule(currentValue); err != "" {
                return err
            }
        }
        
        return ""
    })
    
    hasError := reactivity.CreateMemo(func() bool {
        return error.Get() != ""
    })
    
    return g.Div(
        g.Class("validated-field"),
        g.Input(
            g.Type("text"),
            g.Placeholder(placeholder),
            comps.BindValue(value),
            comps.BindClass("error", hasError),
            dom.OnInputInline(func(el dom.Element) {
                value.Set(el.Value())
            }),
        ),
        g.If(hasError.Get(),
            g.Div(
                g.Class("error-message"),
                comps.BindText(error),
            ),
        ),
    )
}

// Validation rules
func Required(value string) string {
    if strings.TrimSpace(value) == "" {
        return "This field is required"
    }
    return ""
}

func MinLength(min int) ValidationRule {
    return func(value string) string {
        if len(value) < min {
            return fmt.Sprintf("Must be at least %d characters", min)
        }
        return ""
    }
}

func Email(value string) string {
    if value != "" && !strings.Contains(value, "@") {
        return "Must be a valid email address"
    }
    return ""
}

// Usage
func ContactForm() g.Node {
    return g.Form(
        g.Class("contact-form"),
        ValidatedField("", "Enter your name", Required, MinLength(2)),
        ValidatedField("", "Enter your email", Required, Email),
        g.Button(
            g.Type("submit"),
            g.Text("Submit"),
        ),
    )
}
```

## Best Practices

### 1. Keep Components Focused

```go
// GOOD: Single responsibility
func UserProfile(user *reactivity.Signal[User]) g.Node {
    return g.Div(
        g.Class("user-profile"),
        comps.BindText(reactivity.CreateMemo(func() string {
            return user.Get().Name
        })),
    )
}

func UserSettings(settings *reactivity.Signal[Settings]) g.Node {
    return g.Div(
        g.Class("user-settings"),
        // Settings UI
    )
}

// BAD: Too many responsibilities
func UserEverything() g.Node {
    user := reactivity.CreateSignal(User{})
    settings := reactivity.CreateSignal(Settings{})
    posts := reactivity.CreateSignal([]Post{})
    friends := reactivity.CreateSignal([]User{})
    // ... too much state in one component
    
    return g.Div(
        // Complex UI handling everything
    )
}
```

### 2. Use Memos for Expensive Computations

```go
// GOOD: Expensive computation cached
expensiveResult := reactivity.CreateMemo(func() Result {
    data := largeDataSet.Get()
    return performExpensiveCalculation(data)
})

// BAD: Recomputes every time
reactivity.CreateEffect(func() {
    data := largeDataSet.Get()
    result := performExpensiveCalculation(data) // Runs every time
    display.Set(result.String())
})
```

### 3. Implement Proper Cleanup

```go
func ComponentWithCleanup() g.Node {
    var effects []reactivity.Effect
    var timers []*time.Timer
    
    // Create effects and timers
    effect := reactivity.CreateEffect(func() {
        // Effect logic
    })
    effects = append(effects, effect)
    
    timer := time.AfterFunc(5*time.Second, func() {
        // Timer logic
    })
    timers = append(timers, timer)
    
    // Register cleanup when component unmounts
    reactivity.CreateEffect(func() {
        return func() { // Cleanup function
            for _, effect := range effects {
                effect.Dispose()
            }
            for _, timer := range timers {
                timer.Stop()
            }
        }
    })
    
    return g.Div(
        // Component UI
    )
}
```

### 4. Use Descriptive Names

```go
// GOOD: Clear intent
userDisplayName := reactivity.CreateMemo(func() string {
    user := currentUser.Get()
    return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
})

// BAD: Unclear purpose
data := reactivity.CreateMemo(func() string {
    u := user.Get()
    return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
})
```

### 5. Handle Edge Cases

```go
// Always check for empty states, loading states, and errors
content := reactivity.CreateMemo(func() g.Node {
    if loading.Get() {
        return g.Div(
            g.Class("loading"),
            g.Text("Loading..."),
        )
    }
    
    if err := error.Get(); err != nil {
        return g.Div(
            g.Class("error"),
            g.Text(fmt.Sprintf("Error: %s", err.Error())),
        )
    }
    
    items := data.Get()
    if len(items) == 0 {
        return g.Div(
            g.Class("empty"),
            g.Text("No items found"),
        )
    }
    
    // Render items...
    return renderItems(items)
})
```

Next: Learn about [Forms & Events](./forms-events.md) or explore [API Reference](../api/core-apis.md).