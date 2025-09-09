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
showMessage := reactivity.NewSignal(true)
message := reactivity.NewSignal("Hello")
items := reactivity.NewSignal([]Item{})

// Reactive HTML generation
content := reactivity.NewMemo(func() string {
    if showMessage.Get() {
        return fmt.Sprintf(`<div>%s</div>`, message.Get())
    }
    return ""
})
```

## Conditional Rendering

### Basic Conditional Display

Show or hide elements based on state:

```go
type ConditionalComponent struct {
    isVisible *reactivity.Signal[bool]
    message   *reactivity.Signal[string]
    
    content *reactivity.Memo[string]
}

func NewConditionalComponent() *ConditionalComponent {
    c := &ConditionalComponent{
        isVisible: reactivity.NewSignal(false),
        message:   reactivity.NewSignal("Hello, World!"),
    }
    
    c.content = reactivity.NewMemo(func() string {
        if c.isVisible.Get() {
            return fmt.Sprintf(`<p class="message">%s</p>`, c.message.Get())
        }
        return `<p class="placeholder">Nothing to show</p>`
    })
    
    return c
}

func (c *ConditionalComponent) Render() g.Node {
    return h.Div(g.Class("conditional-container"),
        h.Button(
            g.Attr("data-click", "toggle"),
            g.Text("Toggle Message"),
        ),
        h.Div(
            g.Attr("data-html", "content"),
            g.Text(c.content.Get()),
        ),
    )
}

func (c *ConditionalComponent) Attach() {
    comps.BindClick("toggle", c.toggle)
    comps.BindHTML("content", c.content)
}

func (c *ConditionalComponent) toggle() {
    c.isVisible.Update(func(visible bool) bool {
        return !visible
    })
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

type StatusComponent struct {
    status *reactivity.Signal[Status]
    data   *reactivity.Signal[[]string]
    error  *reactivity.Signal[error]
    
    content *reactivity.Memo[string]
}

func NewStatusComponent() *StatusComponent {
    c := &StatusComponent{
        status: reactivity.NewSignal(StatusLoading),
        data:   reactivity.NewSignal([]string{}),
        error:  reactivity.NewSignal(nil),
    }
    
    c.content = reactivity.NewMemo(func() string {
        switch c.status.Get() {
        case StatusLoading:
            return `<div class="loading">Loading...</div>`
            
        case StatusError:
            err := c.error.Get()
            if err != nil {
                return fmt.Sprintf(`<div class="error">Error: %s</div>`, err.Error())
            }
            return `<div class="error">Unknown error occurred</div>`
            
        case StatusEmpty:
            return `<div class="empty">No data available</div>`
            
        case StatusSuccess:
            data := c.data.Get()
            if len(data) == 0 {
                return `<div class="empty">No items found</div>`
            }
            
            var items strings.Builder
            for _, item := range data {
                items.WriteString(fmt.Sprintf(`<li>%s</li>`, item))
            }
            return fmt.Sprintf(`<ul class="data-list">%s</ul>`, items.String())
            
        default:
            return `<div class="unknown">Unknown status</div>`
        }
    })
    
    return c
}

func (c *StatusComponent) Render() g.Node {
    return h.Div(g.Class("status-container"),
        h.Div(g.Class("controls"),
            h.Button(
                g.Attr("data-click", "load"),
                g.Text("Load Data"),
            ),
            h.Button(
                g.Attr("data-click", "clear"),
                g.Text("Clear"),
            ),
            h.Button(
                g.Attr("data-click", "error"),
                g.Text("Trigger Error"),
            ),
        ),
        h.Div(
            g.Attr("data-html", "content"),
            g.Text(c.content.Get()),
        ),
    )
}

func (c *StatusComponent) Attach() {
    comps.BindClick("load", c.loadData)
    comps.BindClick("clear", c.clearData)
    comps.BindClick("error", c.triggerError)
    comps.BindHTML("content", c.content)
}

func (c *StatusComponent) loadData() {
    c.status.Set(StatusLoading)
    
    // Simulate async data loading
    go func() {
        time.Sleep(1 * time.Second)
        c.data.Set([]string{"Item 1", "Item 2", "Item 3"})
        c.status.Set(StatusSuccess)
    }()
}

func (c *StatusComponent) clearData() {
    c.data.Set([]string{})
    c.status.Set(StatusEmpty)
}

func (c *StatusComponent) triggerError() {
    c.error.Set(fmt.Errorf("simulated error"))
    c.status.Set(StatusError)
}
```

### Conditional Classes and Attributes

Dynamically apply CSS classes and attributes:

```go
type ButtonComponent struct {
    isActive  *reactivity.Signal[bool]
    isLoading *reactivity.Signal[bool]
    text      *reactivity.Signal[string]
    
    buttonClass *reactivity.Memo[string]
    buttonAttrs *reactivity.Memo[string]
}

func NewButtonComponent() *ButtonComponent {
    c := &ButtonComponent{
        isActive:  reactivity.NewSignal(false),
        isLoading: reactivity.NewSignal(false),
        text:      reactivity.NewSignal("Click me"),
    }
    
    c.buttonClass = reactivity.NewMemo(func() string {
        classes := []string{"btn"}
        
        if c.isActive.Get() {
            classes = append(classes, "btn-active")
        }
        
        if c.isLoading.Get() {
            classes = append(classes, "btn-loading")
        }
        
        return strings.Join(classes, " ")
    })
    
    c.buttonAttrs = reactivity.NewMemo(func() string {
        attrs := []string{}
        
        if c.isLoading.Get() {
            attrs = append(attrs, `disabled="disabled"`)
        }
        
        return strings.Join(attrs, " ")
    })
    
    return c
}

func (c *ButtonComponent) Render() g.Node {
    return h.Button(
        g.Class(c.buttonClass.Get()),
        g.Attr("data-click", "handleClick"),
        g.If(c.isLoading.Get(), g.Attr("disabled", "disabled")),
        g.Text(c.text.Get()),
    )
}

func (c *ButtonComponent) Attach() {
    comps.BindClick("handleClick", c.handleClick)
    
    // Update button when state changes
    reactivity.NewEffect(func() {
        // Re-render button when class or attributes change
        _ = c.buttonClass.Get()
        _ = c.buttonAttrs.Get()
        _ = c.text.Get()
        
        // Update the button element
        c.updateButton()
    })
}

func (c *ButtonComponent) updateButton() {
    button := dom.QuerySelector(`[data-click="handleClick"]`)
    if button != nil {
        button.SetClassName(c.buttonClass.Get())
        button.SetTextContent(c.text.Get())
        
        if c.isLoading.Get() {
            button.SetAttribute("disabled", "disabled")
        } else {
            button.RemoveAttribute("disabled")
        }
    }
}

func (c *ButtonComponent) handleClick() {
    if c.isLoading.Get() {
        return // Ignore clicks while loading
    }
    
    c.isLoading.Set(true)
    c.text.Set("Loading...")
    
    // Simulate async operation
    go func() {
        time.Sleep(2 * time.Second)
        c.isLoading.Set(false)
        c.isActive.Update(func(active bool) bool { return !active })
        
        if c.isActive.Get() {
            c.text.Set("Active")
        } else {
            c.text.Set("Inactive")
        }
    }()
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

type TodoList struct {
    items    *reactivity.Signal[[]Item]
    newItem  *reactivity.Signal[string]
    
    listHTML *reactivity.Memo[string]
}

func NewTodoList() *TodoList {
    tl := &TodoList{
        items:   reactivity.NewSignal([]Item{}),
        newItem: reactivity.NewSignal(""),
    }
    
    tl.listHTML = reactivity.NewMemo(func() string {
        items := tl.items.Get()
        
        if len(items) == 0 {
            return `<p class="empty">No items yet. Add one above!</p>`
        }
        
        var html strings.Builder
        html.WriteString(`<ul class="todo-list">`)
        
        for _, item := range items {
            checkedAttr := ""
            if item.Done {
                checkedAttr = "checked"
            }
            
            html.WriteString(fmt.Sprintf(`
                <li class="todo-item" data-id="%d">
                    <input type="checkbox" %s data-change="toggle-%d">
                    <span class="%s">%s</span>
                    <button data-click="delete-%d">Delete</button>
                </li>
            `, item.ID, checkedAttr, item.ID, 
               map[bool]string{true: "done", false: "pending"}[item.Done],
               item.Name, item.ID))
        }
        
        html.WriteString(`</ul>`)
        return html.String()
    })
    
    return tl
}

func (tl *TodoList) Render() string {
    return fmt.Sprintf(`
        <div class="todo-container">
            <div class="add-item">
                <input type="text" data-input="newItem" 
                       placeholder="Add new item..." value="%s">
                <button data-click="add">Add</button>
            </div>
            <div data-html="list">%s</div>
        </div>
    `, tl.newItem.Get(), tl.listHTML.Get())
}

func (tl *TodoList) Attach() {
    tl.BindInput("newItem", tl.newItem)
    tl.BindClick("add", tl.addItem)
    tl.BindHTML("list", tl.listHTML)
    
    // Bind dynamic handlers for each item
    reactivity.NewEffect(func() {
        items := tl.items.Get()
        
        // Re-bind handlers when list changes
        for _, item := range items {
            tl.bindItemHandlers(item)
        }
    })
}

func (tl *TodoList) bindItemHandlers(item Item) {
    // Bind toggle handler
    toggleSelector := fmt.Sprintf(`[data-change="toggle-%d"]`, item.ID)
    if element := dom.QuerySelector(toggleSelector); element != nil {
        element.AddEventListener("change", false, func(event dom.Event) {
            tl.toggleItem(item.ID)
        })
    }
    
    // Bind delete handler
    deleteSelector := fmt.Sprintf(`[data-click="delete-%d"]`, item.ID)
    if element := dom.QuerySelector(deleteSelector); element != nil {
        element.AddEventListener("click", false, func(event dom.Event) {
            tl.deleteItem(item.ID)
        })
    }
}

func (tl *TodoList) addItem() {
    text := strings.TrimSpace(tl.newItem.Get())
    if text == "" {
        return
    }
    
    tl.items.Update(func(items []Item) []Item {
        newID := len(items) + 1
        return append(items, Item{
            ID:   newID,
            Name: text,
            Done: false,
        })
    })
    
    tl.newItem.Set("")
}

func (tl *TodoList) toggleItem(id int) {
    tl.items.Update(func(items []Item) []Item {
        for i, item := range items {
            if item.ID == id {
                items[i].Done = !items[i].Done
                break
            }
        }
        return items
    })
}

func (tl *TodoList) deleteItem(id int) {
    tl.items.Update(func(items []Item) []Item {
        filtered := make([]Item, 0, len(items))
        for _, item := range items {
            if item.ID != id {
                filtered = append(filtered, item)
            }
        }
        return filtered
    })
}
```

### Filtered and Sorted Lists

Implement filtering and sorting:

```go
type FilteredList struct {
    allItems     *reactivity.Signal[[]Item]
    filter       *reactivity.Signal[string]
    sortBy       *reactivity.Signal[string]
    showDoneOnly *reactivity.Signal[bool]
    
    filteredItems *reactivity.Memo[[]Item]
    listHTML      *reactivity.Memo[string]
}

func NewFilteredList() *FilteredList {
    fl := &FilteredList{
        allItems:     reactivity.NewSignal([]Item{}),
        filter:       reactivity.NewSignal(""),
        sortBy:       reactivity.NewSignal("name"),
        showDoneOnly: reactivity.NewSignal(false),
    }
    
    fl.filteredItems = reactivity.NewMemo(func() []Item {
        items := fl.allItems.Get()
        filter := strings.ToLower(fl.filter.Get())
        sortBy := fl.sortBy.Get()
        showDoneOnly := fl.showDoneOnly.Get()
        
        // Filter items
        var filtered []Item
        for _, item := range items {
            // Text filter
            if filter != "" && !strings.Contains(strings.ToLower(item.Name), filter) {
                continue
            }
            
            // Done filter
            if showDoneOnly && !item.Done {
                continue
            }
            
            filtered = append(filtered, item)
        }
        
        // Sort items
        switch sortBy {
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
    
    fl.listHTML = reactivity.NewMemo(func() string {
        items := fl.filteredItems.Get()
        
        if len(items) == 0 {
            return `<p class="empty">No items match the current filter.</p>`
        }
        
        var html strings.Builder
        html.WriteString(`<ul class="filtered-list">`)
        
        for _, item := range items {
            status := "pending"
            if item.Done {
                status = "done"
            }
            
            html.WriteString(fmt.Sprintf(`
                <li class="item %s">
                    <span class="name">%s</span>
                    <span class="status">%s</span>
                </li>
            `, status, item.Name, status))
        }
        
        html.WriteString(`</ul>`)
        return html.String()
    })
    
    return fl
}

func (fl *FilteredList) Render() string {
    return fmt.Sprintf(`
        <div class="filtered-container">
            <div class="controls">
                <input type="text" data-input="filter" 
                       placeholder="Filter items..." value="%s">
                
                <select data-change="sortBy">
                    <option value="name" %s>Sort by Name</option>
                    <option value="status" %s>Sort by Status</option>
                </select>
                
                <label>
                    <input type="checkbox" data-change="showDoneOnly" %s>
                    Show completed only
                </label>
            </div>
            
            <div class="stats">
                Showing %d of %d items
            </div>
            
            <div data-html="list">%s</div>
        </div>
    `, 
        fl.filter.Get(),
        map[bool]string{true: "selected", false: ""}[fl.sortBy.Get() == "name"],
        map[bool]string{true: "selected", false: ""}[fl.sortBy.Get() == "status"],
        map[bool]string{true: "checked", false: ""}[fl.showDoneOnly.Get()],
        len(fl.filteredItems.Get()),
        len(fl.allItems.Get()),
        fl.listHTML.Get())
}

func (fl *FilteredList) Attach() {
    fl.BindInput("filter", fl.filter)
    fl.BindChange("sortBy", func(value string) {
        fl.sortBy.Set(value)
    })
    fl.BindChange("showDoneOnly", func(checked bool) {
        fl.showDoneOnly.Set(checked)
    })
    fl.BindHTML("list", fl.listHTML)
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

type DynamicContent struct {
    contentType *reactivity.Signal[ContentType]
    textData    *reactivity.Signal[string]
    imageURL    *reactivity.Signal[string]
    videoURL    *reactivity.Signal[string]
    listData    *reactivity.Signal[[]string]
    
    content *reactivity.Memo[string]
}

func NewDynamicContent() *DynamicContent {
    dc := &DynamicContent{
        contentType: reactivity.NewSignal(ContentText),
        textData:    reactivity.NewSignal("Sample text content"),
        imageURL:    reactivity.NewSignal("https://via.placeholder.com/300x200"),
        videoURL:    reactivity.NewSignal("https://www.w3schools.com/html/mov_bbb.mp4"),
        listData:    reactivity.NewSignal([]string{"Item 1", "Item 2", "Item 3"}),
    }
    
    dc.content = reactivity.NewMemo(func() string {
        switch dc.contentType.Get() {
        case ContentText:
            return fmt.Sprintf(`
                <div class="text-content">
                    <h3>Text Content</h3>
                    <p>%s</p>
                    <textarea data-input="textData" rows="4" cols="50">%s</textarea>
                </div>
            `, dc.textData.Get(), dc.textData.Get())
            
        case ContentImage:
            return fmt.Sprintf(`
                <div class="image-content">
                    <h3>Image Content</h3>
                    <img src="%s" alt="Dynamic image" style="max-width: 100%%;">
                    <input type="url" data-input="imageURL" 
                           placeholder="Image URL" value="%s">
                </div>
            `, dc.imageURL.Get(), dc.imageURL.Get())
            
        case ContentVideo:
            return fmt.Sprintf(`
                <div class="video-content">
                    <h3>Video Content</h3>
                    <video controls style="max-width: 100%%;">
                        <source src="%s" type="video/mp4">
                        Your browser does not support the video tag.
                    </video>
                    <input type="url" data-input="videoURL" 
                           placeholder="Video URL" value="%s">
                </div>
            `, dc.videoURL.Get(), dc.videoURL.Get())
            
        case ContentList:
            listItems := dc.listData.Get()
            var items strings.Builder
            for i, item := range listItems {
                items.WriteString(fmt.Sprintf(`
                    <li>
                        <input type="text" data-input="listItem-%d" value="%s">
                        <button data-click="removeItem-%d">Remove</button>
                    </li>
                `, i, item, i))
            }
            
            return fmt.Sprintf(`
                <div class="list-content">
                    <h3>List Content</h3>
                    <ul>%s</ul>
                    <button data-click="addListItem">Add Item</button>
                </div>
            `, items.String())
            
        default:
            return `<div class="unknown-content">Unknown content type</div>`
        }
    })
    
    return dc
}

func (dc *DynamicContent) Render() string {
    return fmt.Sprintf(`
        <div class="dynamic-container">
            <div class="content-selector">
                <h2>Dynamic Content Demo</h2>
                <div class="tabs">
                    <button class="%s" data-click="selectText">Text</button>
                    <button class="%s" data-click="selectImage">Image</button>
                    <button class="%s" data-click="selectVideo">Video</button>
                    <button class="%s" data-click="selectList">List</button>
                </div>
            </div>
            
            <div class="content-area" data-html="content">
                %s
            </div>
        </div>
    `,
        map[bool]string{true: "active", false: ""}[dc.contentType.Get() == ContentText],
        map[bool]string{true: "active", false: ""}[dc.contentType.Get() == ContentImage],
        map[bool]string{true: "active", false: ""}[dc.contentType.Get() == ContentVideo],
        map[bool]string{true: "active", false: ""}[dc.contentType.Get() == ContentList],
        dc.content.Get())
}

func (dc *DynamicContent) Attach() {
    dc.BindClick("selectText", func() { dc.contentType.Set(ContentText) })
    dc.BindClick("selectImage", func() { dc.contentType.Set(ContentImage) })
    dc.BindClick("selectVideo", func() { dc.contentType.Set(ContentVideo) })
    dc.BindClick("selectList", func() { dc.contentType.Set(ContentList) })
    
    dc.BindHTML("content", dc.content)
    
    // Bind dynamic inputs based on content type
    reactivity.NewEffect(func() {
        contentType := dc.contentType.Get()
        
        switch contentType {
        case ContentText:
            dc.BindInput("textData", dc.textData)
            
        case ContentImage:
            dc.BindInput("imageURL", dc.imageURL)
            
        case ContentVideo:
            dc.BindInput("videoURL", dc.videoURL)
            
        case ContentList:
            dc.bindListHandlers()
        }
    })
}

func (dc *DynamicContent) bindListHandlers() {
    dc.BindClick("addListItem", dc.addListItem)
    
    // Bind handlers for existing list items
    listData := dc.listData.Get()
    for i := range listData {
        dc.bindListItemHandlers(i)
    }
}

func (dc *DynamicContent) bindListItemHandlers(index int) {
    // Bind input for list item
    inputSelector := fmt.Sprintf("listItem-%d", index)
    if element := dom.QuerySelector(fmt.Sprintf(`[data-input="%s"]`, inputSelector)); element != nil {
        element.AddEventListener("input", false, func(event dom.Event) {
            value := element.(*dom.HTMLInputElement).Value
            dc.updateListItem(index, value)
        })
    }
    
    // Bind remove button
    removeSelector := fmt.Sprintf("removeItem-%d", index)
    dc.BindClick(removeSelector, func() {
        dc.removeListItem(index)
    })
}

func (dc *DynamicContent) addListItem() {
    dc.listData.Update(func(items []string) []string {
        return append(items, "New item")
    })
}

func (dc *DynamicContent) updateListItem(index int, value string) {
    dc.listData.Update(func(items []string) []string {
        if index >= 0 && index < len(items) {
            items[index] = value
        }
        return items
    })
}

func (dc *DynamicContent) removeListItem(index int) {
    dc.listData.Update(func(items []string) []string {
        if index >= 0 && index < len(items) {
            return append(items[:index], items[index+1:]...)
        }
        return items
    })
}
```

## Nested Components

### Component Composition

Build complex UIs by composing smaller components:

```go
// Child component
type Card struct {
    title   *reactivity.Signal[string]
    content *reactivity.Signal[string]
    visible *reactivity.Signal[bool]
}

func NewCard(title, content string) *Card {
    return &Card{
        title:   reactivity.NewSignal(title),
        content: reactivity.NewSignal(content),
        visible: reactivity.NewSignal(true),
    }
}

func (c *Card) Render() string {
    if !c.visible.Get() {
        return ""
    }
    
    return fmt.Sprintf(`
        <div class="card">
            <div class="card-header">
                <h3 data-text="title">%s</h3>
                <button data-click="close">×</button>
            </div>
            <div class="card-body" data-text="content">
                %s
            </div>
        </div>
    `, c.title.Get(), c.content.Get())
}

func (c *Card) Attach() {
    c.BindText("title", c.title)
    c.BindText("content", c.content)
    c.BindClick("close", func() {
        c.visible.Set(false)
    })
}

// Parent component that manages multiple cards
type Dashboard struct {
    cards    []*Card
    newTitle *reactivity.Signal[string]
    newContent *reactivity.Signal[string]
    
    cardsHTML *reactivity.Memo[string]
}

func NewDashboard() *Dashboard {
    d := &Dashboard{
        cards: []*Card{
            NewCard("Welcome", "Welcome to the dashboard!"),
            NewCard("Stats", "Your statistics will appear here."),
            NewCard("News", "Latest news and updates."),
        },
        newTitle:   reactivity.NewSignal(""),
        newContent: reactivity.NewSignal(""),
    }
    
    d.cardsHTML = reactivity.NewMemo(func() string {
        var html strings.Builder
        
        for i, card := range d.cards {
            if card.visible.Get() {
                html.WriteString(fmt.Sprintf(`
                    <div class="card-wrapper" data-card="%d">
                        %s
                    </div>
                `, i, card.Render()))
            }
        }
        
        return html.String()
    })
    
    return d
}

func (d *Dashboard) Render() string {
    return fmt.Sprintf(`
        <div class="dashboard">
            <div class="dashboard-header">
                <h1>Dashboard</h1>
                <div class="add-card-form">
                    <input type="text" data-input="newTitle" 
                           placeholder="Card title" value="%s">
                    <input type="text" data-input="newContent" 
                           placeholder="Card content" value="%s">
                    <button data-click="addCard">Add Card</button>
                </div>
            </div>
            
            <div class="cards-grid" data-html="cards">
                %s
            </div>
        </div>
    `, d.newTitle.Get(), d.newContent.Get(), d.cardsHTML.Get())
}

func (d *Dashboard) Attach() {
    d.BindInput("newTitle", d.newTitle)
    d.BindInput("newContent", d.newContent)
    d.BindClick("addCard", d.addCard)
    d.BindHTML("cards", d.cardsHTML)
    
    // Attach all card components
    for _, card := range d.cards {
        card.Attach()
    }
    
    // Re-attach cards when HTML changes
    reactivity.NewEffect(func() {
        _ = d.cardsHTML.Get() // Track changes
        
        // Re-attach all visible cards
        for _, card := range d.cards {
            if card.visible.Get() {
                card.Attach()
            }
        }
    })
}

func (d *Dashboard) addCard() {
    title := strings.TrimSpace(d.newTitle.Get())
    content := strings.TrimSpace(d.newContent.Get())
    
    if title == "" || content == "" {
        return
    }
    
    newCard := NewCard(title, content)
    d.cards = append(d.cards, newCard)
    
    // Clear form
    d.newTitle.Set("")
    d.newContent.Set("")
    
    // Trigger re-render
    d.cardsHTML.Invalidate()
}
```

## Advanced Patterns

### Lazy Loading

Load content only when needed:

```go
type LazySection struct {
    isExpanded *reactivity.Signal[bool]
    isLoaded   *reactivity.Signal[bool]
    data       *reactivity.Signal[string]
    loading    *reactivity.Signal[bool]
    
    content *reactivity.Memo[string]
}

func NewLazySection() *LazySection {
    ls := &LazySection{
        isExpanded: reactivity.NewSignal(false),
        isLoaded:   reactivity.NewSignal(false),
        data:       reactivity.NewSignal(""),
        loading:    reactivity.NewSignal(false),
    }
    
    ls.content = reactivity.NewMemo(func() string {
        if !ls.isExpanded.Get() {
            return `<p>Click to expand...</p>`
        }
        
        if ls.loading.Get() {
            return `<p>Loading content...</p>`
        }
        
        if !ls.isLoaded.Get() {
            return `<p>Content not loaded yet.</p>`
        }
        
        return fmt.Sprintf(`<div class="loaded-content">%s</div>`, ls.data.Get())
    })
    
    return ls
}

func (ls *LazySection) Render() string {
    return fmt.Sprintf(`
        <div class="lazy-section">
            <button data-click="toggle">
                %s
            </button>
            <div class="section-content" data-html="content">
                %s
            </div>
        </div>
    `, 
        map[bool]string{true: "Collapse", false: "Expand"}[ls.isExpanded.Get()],
        ls.content.Get())
}

func (ls *LazySection) Attach() {
    ls.BindClick("toggle", ls.toggle)
    ls.BindHTML("content", ls.content)
}

func (ls *LazySection) toggle() {
    ls.isExpanded.Update(func(expanded bool) bool {
        newExpanded := !expanded
        
        // Load data when expanding for the first time
        if newExpanded && !ls.isLoaded.Get() {
            ls.loadData()
        }
        
        return newExpanded
    })
}

func (ls *LazySection) loadData() {
    ls.loading.Set(true)
    
    go func() {
        // Simulate API call
        time.Sleep(2 * time.Second)
        
        ls.data.Set("This is the lazily loaded content! It contains important information that was fetched from the server.")
        ls.isLoaded.Set(true)
        ls.loading.Set(false)
    }()
}
```

### Virtual Scrolling (Concept)

For very large lists, implement virtual scrolling:

```go
type VirtualList struct {
    allItems     *reactivity.Signal[[]string]
    scrollTop    *reactivity.Signal[int]
    itemHeight   int
    visibleCount int
    
    visibleItems *reactivity.Memo[[]string]
    startIndex   *reactivity.Memo[int]
    endIndex     *reactivity.Memo[int]
}

func NewVirtualList(itemHeight, visibleCount int) *VirtualList {
    vl := &VirtualList{
        allItems:     reactivity.NewSignal([]string{}),
        scrollTop:    reactivity.NewSignal(0),
        itemHeight:   itemHeight,
        visibleCount: visibleCount,
    }
    
    vl.startIndex = reactivity.NewMemo(func() int {
        scrollTop := vl.scrollTop.Get()
        index := scrollTop / vl.itemHeight
        if index < 0 {
            return 0
        }
        return index
    })
    
    vl.endIndex = reactivity.NewMemo(func() int {
        start := vl.startIndex.Get()
        end := start + vl.visibleCount
        allItems := vl.allItems.Get()
        if end > len(allItems) {
            return len(allItems)
        }
        return end
    })
    
    vl.visibleItems = reactivity.NewMemo(func() []string {
        allItems := vl.allItems.Get()
        start := vl.startIndex.Get()
        end := vl.endIndex.Get()
        
        if start >= len(allItems) {
            return []string{}
        }
        
        return allItems[start:end]
    })
    
    return vl
}

// Implementation would include scroll handling and DOM manipulation
// This is a simplified concept - full implementation would be more complex
```

## Performance Considerations

### Optimize Memo Dependencies

```go
// BAD: Memo depends on entire large object
expensiveMemo := reactivity.NewMemo(func() string {
    user := largeUserObject.Get() // Entire object dependency
    return user.Name // Only need name
})

// GOOD: Extract specific fields
userName := reactivity.NewMemo(func() string {
    user := largeUserObject.Get()
    return user.Name
})

displayName := reactivity.NewMemo(func() string {
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
type SearchComponent struct {
    query   *reactivity.Signal[string]
    results *reactivity.Signal[[]string]
    
    debouncedSearch *reactivity.Effect
}

func NewSearchComponent() *SearchComponent {
    sc := &SearchComponent{
        query:   reactivity.NewSignal(""),
        results: reactivity.NewSignal([]string{}),
    }
    
    // Debounced search effect
    var searchTimer *time.Timer
    sc.debouncedSearch = reactivity.NewEffect(func() {
        query := sc.query.Get()
        
        // Cancel previous timer
        if searchTimer != nil {
            searchTimer.Stop()
        }
        
        // Set new timer
        searchTimer = time.AfterFunc(300*time.Millisecond, func() {
            sc.performSearch(query)
        })
    })
    
    return sc
}

func (sc *SearchComponent) performSearch(query string) {
    if query == "" {
        sc.results.Set([]string{})
        return
    }
    
    // Simulate API call
    go func() {
        time.Sleep(100 * time.Millisecond)
        
        // Mock search results
        results := []string{
            fmt.Sprintf("Result 1 for '%s'", query),
            fmt.Sprintf("Result 2 for '%s'", query),
            fmt.Sprintf("Result 3 for '%s'", query),
        }
        
        sc.results.Set(results)
    }()
}
```

## Common Patterns

### Toggle Pattern

```go
type Toggle struct {
    isOn *reactivity.Signal[bool]
}

func NewToggle(initial bool) *Toggle {
    return &Toggle{
        isOn: reactivity.NewSignal(initial),
    }
}

func (t *Toggle) Render() string {
    state := "off"
    if t.isOn.Get() {
        state = "on"
    }
    
    return fmt.Sprintf(`
        <button class="toggle %s" data-click="toggle">
            %s
        </button>
    `, state, strings.ToUpper(state))
}

func (t *Toggle) Attach() {
    t.BindClick("toggle", func() {
        t.isOn.Update(func(on bool) bool { return !on })
    })
}
```

### Counter Pattern

```go
type Counter struct {
    count *reactivity.Signal[int]
    min   int
    max   int
}

func NewCounter(initial, min, max int) *Counter {
    return &Counter{
        count: reactivity.NewSignal(initial),
        min:   min,
        max:   max,
    }
}

func (c *Counter) Render() string {
    count := c.count.Get()
    
    return fmt.Sprintf(`
        <div class="counter">
            <button data-click="decrement" %s>-</button>
            <span class="count" data-text="count">%d</span>
            <button data-click="increment" %s>+</button>
        </div>
    `,
        map[bool]string{true: "disabled", false: ""}[count <= c.min],
        count,
        map[bool]string{true: "disabled", false: ""}[count >= c.max])
}

func (c *Counter) Attach() {
    c.BindText("count", c.count)
    c.BindClick("increment", c.increment)
    c.BindClick("decrement", c.decrement)
}

func (c *Counter) increment() {
    c.count.Update(func(n int) int {
        if n < c.max {
            return n + 1
        }
        return n
    })
}

func (c *Counter) decrement() {
    c.count.Update(func(n int) int {
        if n > c.min {
            return n - 1
        }
        return n
    })
}
```

### Form Validation Pattern

```go
type ValidationRule func(string) string

type ValidatedField struct {
    value *reactivity.Signal[string]
    error *reactivity.Memo[string]
    rules []ValidationRule
}

func NewValidatedField(initial string, rules ...ValidationRule) *ValidatedField {
    vf := &ValidatedField{
        value: reactivity.NewSignal(initial),
        rules: rules,
    }
    
    vf.error = reactivity.NewMemo(func() string {
        value := vf.value.Get()
        
        for _, rule := range vf.rules {
            if err := rule(value); err != "" {
                return err
            }
        }
        
        return ""
    })
    
    return vf
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
func NewContactForm() *ContactForm {
    return &ContactForm{
        name:  NewValidatedField("", Required, MinLength(2)),
        email: NewValidatedField("", Required, Email),
    }
}
```

## Best Practices

### 1. Keep Components Focused

```go
// GOOD: Single responsibility
type UserProfile struct {
    user *reactivity.Signal[User]
}

type UserSettings struct {
    settings *reactivity.Signal[Settings]
}

// BAD: Too many responsibilities
type UserEverything struct {
    user     *reactivity.Signal[User]
    settings *reactivity.Signal[Settings]
    posts    *reactivity.Signal[[]Post]
    friends  *reactivity.Signal[[]User]
    // ... too much
}
```

### 2. Use Memos for Expensive Computations

```go
// GOOD: Expensive computation cached
expensiveResult := reactivity.NewMemo(func() Result {
    data := largeDataSet.Get()
    return performExpensiveCalculation(data)
})

// BAD: Recomputes every time
reactivity.NewEffect(func() {
    data := largeDataSet.Get()
    result := performExpensiveCalculation(data) // Runs every time
    display.Set(result.String())
})
```

### 3. Implement Proper Cleanup

```go
type Component struct {
    effects []reactivity.Effect
    timers  []*time.Timer
}

func (c *Component) Cleanup() {
    // Dispose effects
    for _, effect := range c.effects {
        effect.Dispose()
    }
    
    // Stop timers
    for _, timer := range c.timers {
        timer.Stop()
    }
}
```

### 4. Use Descriptive Names

```go
// GOOD: Clear intent
userDisplayName := reactivity.NewMemo(func() string {
    user := currentUser.Get()
    return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
})

// BAD: Unclear purpose
data := reactivity.NewMemo(func() string {
    u := user.Get()
    return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
})
```

### 5. Handle Edge Cases

```go
// Always check for empty states, loading states, and errors
content := reactivity.NewMemo(func() string {
    if loading.Get() {
        return `<div class="loading">Loading...</div>`
    }
    
    if err := error.Get(); err != nil {
        return fmt.Sprintf(`<div class="error">Error: %s</div>`, err.Error())
    }
    
    items := data.Get()
    if len(items) == 0 {
        return `<div class="empty">No items found</div>`
    }
    
    // Render items...
    return renderItems(items)
})
```

Next: Learn about [Forms & Events](./forms-events.md) or explore [API Reference](../api/core-apis.md).