# Real-World Helper Function Examples

This guide demonstrates practical usage patterns for UIwGo helper functions through complete, real-world application examples. Each example shows how to combine multiple helpers effectively.

## Related Documentation

- **[Helper Functions Guide](./helper-functions.md)** - Comprehensive guide to all helper functions
- **[Quick Reference](./quick-reference.md)** - Concise syntax reference
- **[Integration Examples](./integration-examples.md)** - Complex multi-helper scenarios
- **[Performance Optimization](./performance-optimization.md)** - Performance best practices
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions

## Table of Contents

- [E-commerce Product Catalog](#e-commerce-product-catalog)
- [Task Management Dashboard](#task-management-dashboard)
- [Social Media Feed](#social-media-feed)
- [Data Analytics Dashboard](#data-analytics-dashboard)
- [User Settings Panel](#user-settings-panel)
- [Chat Application](#chat-application)
- [File Manager](#file-manager)
- [Shopping Cart](#shopping-cart)

## E-commerce Product Catalog

A complete product catalog with filtering, sorting, and different view modes.

```go
package main

import (
    "fmt"
    "sort"
    "strings"
    
    "github.com/ozanturksever/uiwgo/comps"
    "github.com/ozanturksever/uiwgo/dom"
    "github.com/ozanturksever/uiwgo/reactivity"
    g "maragu.dev/gomponents"
)

type Product struct {
    ID          string
    Name        string
    Price       float64
    Category    string
    ImageURL    string
    Description string
    InStock     bool
    Rating      float64
}

type ViewMode string

const (
    ViewModeGrid ViewMode = "grid"
    ViewModeList ViewMode = "list"
)

type SortBy string

const (
    SortByName     SortBy = "name"
    SortByPrice    SortBy = "price"
    SortByRating   SortBy = "rating"
    SortByCategory SortBy = "category"
)

type ProductCatalog struct {
    products     reactivity.Signal[[]Product]
    searchTerm   reactivity.Signal[string]
    selectedCategory reactivity.Signal[string]
    viewMode     reactivity.Signal[ViewMode]
    sortBy       reactivity.Signal[SortBy]
    sortAsc      reactivity.Signal[bool]
    showOutOfStock reactivity.Signal[bool]
    loading      reactivity.Signal[bool]
}

func NewProductCatalog() *ProductCatalog {
    return &ProductCatalog{
        products:       reactivity.CreateSignal([]Product{}),
        searchTerm:     reactivity.CreateSignal(""),
        selectedCategory: reactivity.CreateSignal(""),
        viewMode:       reactivity.CreateSignal(ViewModeGrid),
        sortBy:         reactivity.CreateSignal(SortByName),
        sortAsc:        reactivity.CreateSignal(true),
        showOutOfStock: reactivity.CreateSignal(true),
        loading:        reactivity.CreateSignal(false),
    }
}

func (pc *ProductCatalog) render() g.Node {
    // Computed filtered and sorted products
    filteredProducts := reactivity.CreateMemo(func() []Product {
        products := pc.products.Get()
        search := strings.ToLower(pc.searchTerm.Get())
        category := pc.selectedCategory.Get()
        showOOS := pc.showOutOfStock.Get()
        
        var filtered []Product
        for _, p := range products {
            // Category filter
            if category != "" && p.Category != category {
                continue
            }
            
            // Search filter
            if search != "" {
                if !strings.Contains(strings.ToLower(p.Name), search) &&
                   !strings.Contains(strings.ToLower(p.Description), search) {
                    continue
                }
            }
            
            // Stock filter
            if !showOOS && !p.InStock {
                continue
            }
            
            filtered = append(filtered, p)
        }
        
        // Sort products
        sortBy := pc.sortBy.Get()
        asc := pc.sortAsc.Get()
        
        sort.Slice(filtered, func(i, j int) bool {
            var less bool
            switch sortBy {
            case SortByName:
                less = filtered[i].Name < filtered[j].Name
            case SortByPrice:
                less = filtered[i].Price < filtered[j].Price
            case SortByRating:
                less = filtered[i].Rating < filtered[j].Rating
            case SortByCategory:
                less = filtered[i].Category < filtered[j].Category
            }
            
            if asc {
                return less
            }
            return !less
        })
        
        return filtered
    })
    
    // Get unique categories
    categories := reactivity.CreateMemo(func() []string {
        products := pc.products.Get()
        catMap := make(map[string]bool)
        for _, p := range products {
            catMap[p.Category] = true
        }
        
        var cats []string
        for cat := range catMap {
            cats = append(cats, cat)
        }
        sort.Strings(cats)
        return cats
    })
    
    return g.Div(
        g.Class("product-catalog"),
        
        // Header with filters and controls
        g.Div(
            g.Class("catalog-header"),
            g.H1(g.Text("Product Catalog")),
            
            // Search bar
            g.Div(
                g.Class("search-bar"),
                g.Input(
                    g.Type("text"),
                    g.Placeholder("Search products..."),
                    g.Value(pc.searchTerm.Get()),
                    dom.OnInput(func(value string) {
                        pc.searchTerm.Set(value)
                    }),
                ),
            ),
            
            // Filters row
            g.Div(
                g.Class("filters-row"),
                
                // Category filter
                g.Select(
                    g.Value(pc.selectedCategory.Get()),
                    dom.OnChange(func(value string) {
                        pc.selectedCategory.Set(value)
                    }),
                    g.Option(g.Value(""), g.Text("All Categories")),
                    comps.For(comps.ForProps[string]{
                        Items: categories,
                        Key: func(cat string) string { return cat },
                        Children: func(cat string, index int) g.Node {
                            return g.Option(
                                g.Value(cat),
                                g.Text(cat),
                            )
                        },
                    }),
                ),
                
                // Sort controls
                g.Select(
                    g.Value(string(pc.sortBy.Get())),
                    dom.OnChange(func(value string) {
                        pc.sortBy.Set(SortBy(value))
                    }),
                    g.Option(g.Value(string(SortByName)), g.Text("Name")),
                    g.Option(g.Value(string(SortByPrice)), g.Text("Price")),
                    g.Option(g.Value(string(SortByRating)), g.Text("Rating")),
                    g.Option(g.Value(string(SortByCategory)), g.Text("Category")),
                ),
                
                g.Button(
                    g.Text(func() string {
                        if pc.sortAsc.Get() {
                            return "↑ Asc"
                        }
                        return "↓ Desc"
                    }()),
                    dom.OnClick(func() {
                        pc.sortAsc.Set(!pc.sortAsc.Get())
                    }),
                ),
                
                // View mode toggle
                g.Div(
                    g.Class("view-toggle"),
                    g.Button(
                        g.Class(func() string {
                            if pc.viewMode.Get() == ViewModeGrid {
                                return "active"
                            }
                            return ""
                        }()),
                        g.Text("Grid"),
                        dom.OnClick(func() {
                            pc.viewMode.Set(ViewModeGrid)
                        }),
                    ),
                    g.Button(
                        g.Class(func() string {
                            if pc.viewMode.Get() == ViewModeList {
                                return "active"
                            }
                            return ""
                        }()),
                        g.Text("List"),
                        dom.OnClick(func() {
                            pc.viewMode.Set(ViewModeList)
                        }),
                    ),
                ),
                
                // Show out of stock toggle
                g.Label(
                    g.Input(
                        g.Type("checkbox"),
                        g.Checked(pc.showOutOfStock.Get()),
                        dom.OnChange(func() {
                            pc.showOutOfStock.Set(!pc.showOutOfStock.Get())
                        }),
                    ),
                    g.Text("Show out of stock"),
                ),
            ),
        ),
        
        // Loading state
        comps.Show(comps.ShowProps{
            When: pc.loading,
            Children: g.Div(
                g.Class("loading-state"),
                g.Text("Loading products..."),
            ),
        }),
        
        // Products display
        comps.Show(comps.ShowProps{
            When: reactivity.CreateMemo(func() bool {
                return !pc.loading.Get()
            }),
            Children: comps.Switch(comps.SwitchProps{
                When: pc.viewMode,
                Children: []g.Node{
                    comps.Match(comps.MatchProps{
                        When: ViewModeGrid,
                        Children: pc.renderProductGrid(filteredProducts),
                    }),
                    comps.Match(comps.MatchProps{
                        When: ViewModeList,
                        Children: pc.renderProductList(filteredProducts),
                    }),
                },
            }),
        }),
        
        // Empty state
        comps.Show(comps.ShowProps{
            When: reactivity.CreateMemo(func() bool {
                return !pc.loading.Get() && len(filteredProducts.Get()) == 0
            }),
            Children: g.Div(
                g.Class("empty-state"),
                g.H3(g.Text("No products found")),
                g.P(g.Text("Try adjusting your filters or search terms.")),
            ),
        }),
    )
}

func (pc *ProductCatalog) renderProductGrid(products reactivity.Signal[[]Product]) g.Node {
    return g.Div(
        g.Class("product-grid"),
        comps.For(comps.ForProps[Product]{
            Items: products,
            Key: func(p Product) string { return p.ID },
            Children: func(p Product, index int) g.Node {
                return g.Div(
                    g.Class("product-card"),
                    g.If(!p.InStock, g.Class("out-of-stock")),
                    
                    g.Img(
                        g.Src(p.ImageURL),
                        g.Alt(p.Name),
                        g.Class("product-image"),
                    ),
                    
                    g.Div(
                        g.Class("product-info"),
                        g.H3(g.Text(p.Name)),
                        g.P(
                            g.Class("category"),
                            g.Text(p.Category),
                        ),
                        g.P(
                            g.Class("price"),
                            g.Text(fmt.Sprintf("$%.2f", p.Price)),
                        ),
                        g.Div(
                            g.Class("rating"),
                            g.Text(fmt.Sprintf("★ %.1f", p.Rating)),
                        ),
                        comps.Show(comps.ShowProps{
                            When: reactivity.CreateSignal(!p.InStock),
                            Children: g.Span(
                                g.Class("stock-status"),
                                g.Text("Out of Stock"),
                            ),
                        }),
                    ),
                )
            },
        }),
    )
}

func (pc *ProductCatalog) renderProductList(products reactivity.Signal[[]Product]) g.Node {
    return g.Div(
        g.Class("product-list"),
        comps.For(comps.ForProps[Product]{
            Items: products,
            Key: func(p Product) string { return p.ID },
            Children: func(p Product, index int) g.Node {
                return g.Div(
                    g.Class("product-row"),
                    g.If(!p.InStock, g.Class("out-of-stock")),
                    
                    g.Img(
                        g.Src(p.ImageURL),
                        g.Alt(p.Name),
                        g.Class("product-thumbnail"),
                    ),
                    
                    g.Div(
                        g.Class("product-details"),
                        g.H3(g.Text(p.Name)),
                        g.P(
                            g.Class("description"),
                            g.Text(p.Description),
                        ),
                        g.Div(
                            g.Class("product-meta"),
                            g.Span(
                                g.Class("category"),
                                g.Text(p.Category),
                            ),
                            g.Span(
                                g.Class("price"),
                                g.Text(fmt.Sprintf("$%.2f", p.Price)),
                            ),
                            g.Span(
                                g.Class("rating"),
                                g.Text(fmt.Sprintf("★ %.1f", p.Rating)),
                            ),
                            comps.Show(comps.ShowProps{
                                When: reactivity.CreateSignal(!p.InStock),
                                Children: g.Span(
                                    g.Class("stock-status"),
                                    g.Text("Out of Stock"),
                                ),
                            }),
                        ),
                    ),
                )
            },
        }),
    )
}
```

## Task Management Dashboard

A comprehensive task management interface with multiple views and real-time updates.

```go
type TaskStatus string

const (
    TaskStatusTodo       TaskStatus = "todo"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
    TaskPriorityLow    TaskPriority = "low"
    TaskPriorityMedium TaskPriority = "medium"
    TaskPriorityHigh   TaskPriority = "high"
)

type Task struct {
    ID          string
    Title       string
    Description string
    Status      TaskStatus
    Priority    TaskPriority
    Assignee    string
    DueDate     time.Time
    CreatedAt   time.Time
    Tags        []string
}

type DashboardView string

const (
    ViewBoard     DashboardView = "board"
    ViewList      DashboardView = "list"
    ViewCalendar  DashboardView = "calendar"
    ViewAnalytics DashboardView = "analytics"
)

type TaskDashboard struct {
    tasks        reactivity.Signal[[]Task]
    currentView  reactivity.Signal[DashboardView]
    selectedTags reactivity.Signal[[]string]
    assigneeFilter reactivity.Signal[string]
    showCompleted  reactivity.Signal[bool]
    searchTerm     reactivity.Signal[string]
}

func NewTaskDashboard() *TaskDashboard {
    return &TaskDashboard{
        tasks:          reactivity.CreateSignal([]Task{}),
        currentView:    reactivity.CreateSignal(ViewBoard),
        selectedTags:   reactivity.CreateSignal([]string{}),
        assigneeFilter: reactivity.CreateSignal(""),
        showCompleted:  reactivity.CreateSignal(false),
        searchTerm:     reactivity.CreateSignal(""),
    }
}

func (td *TaskDashboard) render() g.Node {
    // Computed filtered tasks
    filteredTasks := reactivity.CreateMemo(func() []Task {
        tasks := td.tasks.Get()
        search := strings.ToLower(td.searchTerm.Get())
        assignee := td.assigneeFilter.Get()
        selectedTags := td.selectedTags.Get()
        showCompleted := td.showCompleted.Get()
        
        var filtered []Task
        for _, task := range tasks {
            // Status filter
            if !showCompleted && task.Status == TaskStatusDone {
                continue
            }
            
            // Assignee filter
            if assignee != "" && task.Assignee != assignee {
                continue
            }
            
            // Search filter
            if search != "" {
                if !strings.Contains(strings.ToLower(task.Title), search) &&
                   !strings.Contains(strings.ToLower(task.Description), search) {
                    continue
                }
            }
            
            // Tags filter
            if len(selectedTags) > 0 {
                hasTag := false
                for _, selectedTag := range selectedTags {
                    for _, taskTag := range task.Tags {
                        if taskTag == selectedTag {
                            hasTag = true
                            break
                        }
                    }
                    if hasTag {
                        break
                    }
                }
                if !hasTag {
                    continue
                }
            }
            
            filtered = append(filtered, task)
        }
        
        return filtered
    })
    
    return g.Div(
        g.Class("task-dashboard"),
        
        // Header with navigation and filters
        g.Header(
            g.Class("dashboard-header"),
            g.H1(g.Text("Task Dashboard")),
            
            // View switcher
            g.Nav(
                g.Class("view-switcher"),
                comps.For(comps.ForProps[DashboardView]{
                    Items: reactivity.CreateSignal([]DashboardView{
                        ViewBoard, ViewList, ViewCalendar, ViewAnalytics,
                    }),
                    Key: func(view DashboardView) string { return string(view) },
                    Children: func(view DashboardView, index int) g.Node {
                        isActive := reactivity.CreateMemo(func() bool {
                            return td.currentView.Get() == view
                        })
                        
                        return g.Button(
                            g.Class("view-tab"),
                            g.If(isActive.Get(), g.Class("active")),
                            g.Text(strings.Title(string(view))),
                            dom.OnClick(func() {
                                td.currentView.Set(view)
                            }),
                        )
                    },
                }),
            ),
            
            // Filters
            g.Div(
                g.Class("filters"),
                g.Input(
                    g.Type("text"),
                    g.Placeholder("Search tasks..."),
                    g.Value(td.searchTerm.Get()),
                    dom.OnInput(func(value string) {
                        td.searchTerm.Set(value)
                    }),
                ),
                
                g.Label(
                    g.Input(
                        g.Type("checkbox"),
                        g.Checked(td.showCompleted.Get()),
                        dom.OnChange(func() {
                            td.showCompleted.Set(!td.showCompleted.Get())
                        }),
                    ),
                    g.Text("Show completed"),
                ),
            ),
        ),
        
        // Main content area
        g.Main(
            g.Class("dashboard-content"),
            comps.Switch(comps.SwitchProps{
                When: td.currentView,
                Children: []g.Node{
                    comps.Match(comps.MatchProps{
                        When: ViewBoard,
                        Children: td.renderKanbanBoard(filteredTasks),
                    }),
                    comps.Match(comps.MatchProps{
                        When: ViewList,
                        Children: td.renderTaskList(filteredTasks),
                    }),
                    comps.Match(comps.MatchProps{
                        When: ViewCalendar,
                        Children: td.renderCalendarView(filteredTasks),
                    }),
                    comps.Match(comps.MatchProps{
                        When: ViewAnalytics,
                        Children: td.renderAnalyticsView(filteredTasks),
                    }),
                },
            }),
        ),
    )
}

func (td *TaskDashboard) renderKanbanBoard(tasks reactivity.Signal[[]Task]) g.Node {
    // Group tasks by status
    tasksByStatus := reactivity.CreateMemo(func() map[TaskStatus][]Task {
        groups := make(map[TaskStatus][]Task)
        for _, task := range tasks.Get() {
            groups[task.Status] = append(groups[task.Status], task)
        }
        return groups
    })
    
    statuses := []TaskStatus{TaskStatusTodo, TaskStatusInProgress, TaskStatusDone}
    
    return g.Div(
        g.Class("kanban-board"),
        comps.For(comps.ForProps[TaskStatus]{
            Items: reactivity.CreateSignal(statuses),
            Key: func(status TaskStatus) string { return string(status) },
            Children: func(status TaskStatus, index int) g.Node {
                columnTasks := reactivity.CreateMemo(func() []Task {
                    return tasksByStatus.Get()[status]
                })
                
                return g.Div(
                    g.Class("kanban-column"),
                    g.H3(
                        g.Class("column-header"),
                        g.Text(strings.Title(strings.ReplaceAll(string(status), "_", " "))),
                        g.Span(
                            g.Class("task-count"),
                            g.Text(fmt.Sprintf("(%d)", len(columnTasks.Get()))),
                        ),
                    ),
                    
                    g.Div(
                        g.Class("column-tasks"),
                        comps.For(comps.ForProps[Task]{
                            Items: columnTasks,
                            Key: func(task Task) string { return task.ID },
                            Children: func(task Task, index int) g.Node {
                                return td.renderTaskCard(task)
                            },
                        }),
                    ),
                )
            },
        }),
    )
}

func (td *TaskDashboard) renderTaskCard(task Task) g.Node {
    return g.Div(
        g.Class("task-card"),
        g.Class(fmt.Sprintf("priority-%s", task.Priority)),
        
        g.H4(
            g.Class("task-title"),
            g.Text(task.Title),
        ),
        
        comps.Show(comps.ShowProps{
            When: reactivity.CreateSignal(task.Description != ""),
            Children: g.P(
                g.Class("task-description"),
                g.Text(task.Description),
            ),
        }),
        
        // Tags
        comps.Show(comps.ShowProps{
            When: reactivity.CreateSignal(len(task.Tags) > 0),
            Children: g.Div(
                g.Class("task-tags"),
                comps.For(comps.ForProps[string]{
                    Items: reactivity.CreateSignal(task.Tags),
                    Key: func(tag string) string { return tag },
                    Children: func(tag string, index int) g.Node {
                        return g.Span(
                            g.Class("tag"),
                            g.Text(tag),
                        )
                    },
                }),
            ),
        }),
        
        // Task meta
        g.Div(
            g.Class("task-meta"),
            comps.Show(comps.ShowProps{
                When: reactivity.CreateSignal(task.Assignee != ""),
                Children: g.Span(
                    g.Class("assignee"),
                    g.Text(task.Assignee),
                ),
            }),
            
            comps.Show(comps.ShowProps{
                When: reactivity.CreateSignal(!task.DueDate.IsZero()),
                Children: g.Span(
                    g.Class("due-date"),
                    g.Text(task.DueDate.Format("Jan 2")),
                ),
            }),
        ),
    )
}
```

## Social Media Feed

A dynamic social media feed with infinite scroll, real-time updates, and interactive features.

```go
type PostType string

const (
    PostTypeText  PostType = "text"
    PostTypeImage PostType = "image"
    PostTypeVideo PostType = "video"
    PostTypeLink  PostType = "link"
)

type Post struct {
    ID        string
    Author    User
    Content   string
    Type      PostType
    MediaURL  string
    Timestamp time.Time
    Likes     int
    Comments  []Comment
    Shares    int
    IsLiked   bool
    IsShared  bool
}

type Comment struct {
    ID        string
    Author    User
    Content   string
    Timestamp time.Time
    Likes     int
    IsLiked   bool
}

type User struct {
    ID        string
    Username  string
    Name      string
    AvatarURL string
    Verified  bool
}

type SocialFeed struct {
    posts           reactivity.Signal[[]Post]
    loading         reactivity.Signal[bool]
    hasMore         reactivity.Signal[bool]
    selectedFilter  reactivity.Signal[PostType]
    showComments    reactivity.Signal[map[string]bool]
    newPostContent  reactivity.Signal[string]
    currentUser     reactivity.Signal[User]
}

func NewSocialFeed() *SocialFeed {
    return &SocialFeed{
        posts:          reactivity.CreateSignal([]Post{}),
        loading:        reactivity.CreateSignal(false),
        hasMore:        reactivity.CreateSignal(true),
        selectedFilter: reactivity.CreateSignal(""),
        showComments:   reactivity.CreateSignal(make(map[string]bool)),
        newPostContent: reactivity.CreateSignal(""),
        currentUser:    reactivity.CreateSignal(User{}),
    }
}

func (sf *SocialFeed) render() g.Node {
    // Filter posts based on selected type
    filteredPosts := reactivity.CreateMemo(func() []Post {
        posts := sf.posts.Get()
        filter := sf.selectedFilter.Get()
        
        if filter == "" {
            return posts
        }
        
        var filtered []Post
        for _, post := range posts {
            if post.Type == filter {
                filtered = append(filtered, post)
            }
        }
        return filtered
    })
    
    return g.Div(
        g.Class("social-feed"),
        
        // Header with post composer
        g.Header(
            g.Class("feed-header"),
            sf.renderPostComposer(),
            
            // Filter tabs
            g.Nav(
                g.Class("post-filters"),
                g.Button(
                    g.Class(func() string {
                        if sf.selectedFilter.Get() == "" {
                            return "filter-tab active"
                        }
                        return "filter-tab"
                    }()),
                    g.Text("All"),
                    dom.OnClick(func() {
                        sf.selectedFilter.Set("")
                    }),
                ),
                comps.For(comps.ForProps[PostType]{
                    Items: reactivity.CreateSignal([]PostType{
                        PostTypeText, PostTypeImage, PostTypeVideo, PostTypeLink,
                    }),
                    Key: func(pType PostType) string { return string(pType) },
                    Children: func(pType PostType, index int) g.Node {
                        isActive := reactivity.CreateMemo(func() bool {
                            return sf.selectedFilter.Get() == pType
                        })
                        
                        return g.Button(
                            g.Class("filter-tab"),
                            g.If(isActive.Get(), g.Class("active")),
                            g.Text(strings.Title(string(pType))),
                            dom.OnClick(func() {
                                sf.selectedFilter.Set(pType)
                            }),
                        )
                    },
                }),
            ),
        ),
        
        // Posts feed
        g.Main(
            g.Class("feed-content"),
            comps.For(comps.ForProps[Post]{
                Items: filteredPosts,
                Key: func(post Post) string { return post.ID },
                Children: func(post Post, index int) g.Node {
                    return sf.renderPost(post)
                },
            }),
            
            // Loading indicator
            comps.Show(comps.ShowProps{
                When: sf.loading,
                Children: g.Div(
                    g.Class("loading-indicator"),
                    g.Text("Loading more posts..."),
                ),
            }),
            
            // Load more button
            comps.Show(comps.ShowProps{
                When: reactivity.CreateMemo(func() bool {
                    return sf.hasMore.Get() && !sf.loading.Get()
                }),
                Children: g.Button(
                    g.Class("load-more"),
                    g.Text("Load More"),
                    dom.OnClick(func() {
                        sf.loadMorePosts()
                    }),
                ),
            }),
        ),
    )
}

func (sf *SocialFeed) renderPost(post Post) g.Node {
    commentsVisible := reactivity.CreateMemo(func() bool {
        return sf.showComments.Get()[post.ID]
    })
    
    return g.Article(
        g.Class("post"),
        
        // Post header
        g.Header(
            g.Class("post-header"),
            g.Img(
                g.Class("avatar"),
                g.Src(post.Author.AvatarURL),
                g.Alt(post.Author.Name),
            ),
            g.Div(
                g.Class("author-info"),
                g.H3(
                    g.Class("author-name"),
                    g.Text(post.Author.Name),
                    comps.Show(comps.ShowProps{
                        When: reactivity.CreateSignal(post.Author.Verified),
                        Children: g.Span(
                            g.Class("verified-badge"),
                            g.Text("✓"),
                        ),
                    }),
                ),
                g.P(
                    g.Class("username"),
                    g.Text("@" + post.Author.Username),
                ),
                g.Time(
                    g.Class("timestamp"),
                    g.Text(formatTimeAgo(post.Timestamp)),
                ),
            ),
        ),
        
        // Post content
        g.Div(
            g.Class("post-content"),
            g.P(g.Text(post.Content)),
            
            // Media content based on post type
            comps.Switch(comps.SwitchProps{
                When: reactivity.CreateSignal(post.Type),
                Children: []g.Node{
                    comps.Match(comps.MatchProps{
                        When: PostTypeImage,
                        Children: comps.Show(comps.ShowProps{
                            When: reactivity.CreateSignal(post.MediaURL != ""),
                            Children: g.Img(
                                g.Class("post-image"),
                                g.Src(post.MediaURL),
                                g.Alt("Post image"),
                            ),
                        }),
                    }),
                    comps.Match(comps.MatchProps{
                        When: PostTypeVideo,
                        Children: comps.Show(comps.ShowProps{
                            When: reactivity.CreateSignal(post.MediaURL != ""),
                            Children: g.Video(
                                g.Class("post-video"),
                                g.Attr("controls", ""),
                                g.Attr("src", post.MediaURL),
                            ),
                        }),
                    }),
                },
            }),
        ),
        
        // Post actions
        g.Footer(
            g.Class("post-actions"),
            g.Button(
                g.Class("action-button like-button"),
                g.If(post.IsLiked, g.Class("liked")),
                g.Text(fmt.Sprintf("♥ %d", post.Likes)),
                dom.OnClick(func() {
                    sf.toggleLike(post.ID)
                }),
            ),
            
            g.Button(
                g.Class("action-button comment-button"),
                g.Text(fmt.Sprintf("💬 %d", len(post.Comments))),
                dom.OnClick(func() {
                    sf.toggleComments(post.ID)
                }),
            ),
            
            g.Button(
                g.Class("action-button share-button"),
                g.If(post.IsShared, g.Class("shared")),
                g.Text(fmt.Sprintf("🔄 %d", post.Shares)),
                dom.OnClick(func() {
                    sf.sharePost(post.ID)
                }),
            ),
        ),
        
        // Comments section
        comps.Show(comps.ShowProps{
            When: commentsVisible,
            Children: g.Div(
                g.Class("comments-section"),
                comps.For(comps.ForProps[Comment]{
                    Items: reactivity.CreateSignal(post.Comments),
                    Key: func(comment Comment) string { return comment.ID },
                    Children: func(comment Comment, index int) g.Node {
                        return sf.renderComment(comment)
                    },
                }),
                
                // Comment composer
                g.Div(
                    g.Class("comment-composer"),
                    g.Textarea(
                        g.Placeholder("Write a comment..."),
                        g.Rows("2"),
                    ),
                    g.Button(
                        g.Text("Post Comment"),
                        dom.OnClick(func() {
                            // Handle comment submission
                        }),
                    ),
                ),
            ),
        }),
    )
}

func (sf *SocialFeed) renderComment(comment Comment) g.Node {
    return g.Div(
        g.Class("comment"),
        g.Img(
            g.Class("comment-avatar"),
            g.Src(comment.Author.AvatarURL),
            g.Alt(comment.Author.Name),
        ),
        g.Div(
            g.Class("comment-content"),
            g.H4(
                g.Class("comment-author"),
                g.Text(comment.Author.Name),
            ),
            g.P(g.Text(comment.Content)),
            g.Div(
                g.Class("comment-actions"),
                g.Time(
                    g.Class("comment-time"),
                    g.Text(formatTimeAgo(comment.Timestamp)),
                ),
                g.Button(
                    g.Class("comment-like"),
                    g.If(comment.IsLiked, g.Class("liked")),
                    g.Text(fmt.Sprintf("♥ %d", comment.Likes)),
                ),
            ),
        ),
    )
}

func (sf *SocialFeed) renderPostComposer() g.Node {
    return g.Div(
        g.Class("post-composer"),
        g.Textarea(
            g.Class("composer-input"),
            g.Placeholder("What's on your mind?"),
            g.Value(sf.newPostContent.Get()),
            dom.OnInput(func(value string) {
                sf.newPostContent.Set(value)
            }),
        ),
        g.Div(
            g.Class("composer-actions"),
            g.Button(
                g.Class("media-button"),
                g.Text("📷 Photo"),
            ),
            g.Button(
                g.Class("media-button"),
                g.Text("🎥 Video"),
            ),
            g.Button(
                g.Class("post-button"),
                g.Disabled(sf.newPostContent.Get() == ""),
                g.Text("Post"),
                dom.OnClick(func() {
                    sf.createPost()
                }),
            ),
        ),
    )
}

// Helper methods
func (sf *SocialFeed) toggleLike(postID string) {
    // Implementation for toggling like
}

func (sf *SocialFeed) toggleComments(postID string) {
    current := sf.showComments.Get()
    current[postID] = !current[postID]
    sf.showComments.Set(current)
}

func (sf *SocialFeed) sharePost(postID string) {
    // Implementation for sharing post
}

func (sf *SocialFeed) loadMorePosts() {
    // Implementation for loading more posts
}

func (sf *SocialFeed) createPost() {
    // Implementation for creating new post
    sf.newPostContent.Set("")
}

func formatTimeAgo(t time.Time) string {
    // Implementation for formatting time ago
    return "2h ago"
}
```

These examples demonstrate how UIwGo's helper functions work together to create complex, interactive applications. Each example shows:

1. **State Management**: Using reactive signals for application state
2. **Conditional Rendering**: Using `Show` and `Switch/Match` for dynamic content
3. **List Rendering**: Using `For` with proper key functions for efficient updates
4. **Component Composition**: Breaking down complex UIs into manageable components
5. **Event Handling**: Responding to user interactions with proper state updates
6. **Performance Optimization**: Using memos for expensive computations
7. **Error Handling**: Graceful handling of edge cases and empty states

These patterns can be adapted and combined to build virtually any type of web application with UIwGo.