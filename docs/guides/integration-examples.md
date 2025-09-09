# Integration Examples: Helper Functions Working Together

This guide demonstrates how UIwGo helper functions can be combined to create complex, interactive applications. Each example shows real-world scenarios where multiple helpers work together seamlessly.

## Related Documentation

- **[Helper Functions Guide](./helper-functions.md)** - Comprehensive guide to all helper functions
- **[Quick Reference](./quick-reference.md)** - Concise syntax reference
- **[Real-World Examples](./real-world-examples.md)** - Practical application examples
- **[Performance Optimization](./performance-optimization.md)** - Performance best practices
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions

## Table of Contents

- [Data Dashboard with Filtering and Sorting](#data-dashboard-with-filtering-and-sorting)
- [Multi-Step Form with Validation](#multi-step-form-with-validation)
- [Real-Time Chat Application](#real-time-chat-application)
- [E-commerce Product Browser](#e-commerce-product-browser)
- [Kanban Board with Drag and Drop](#kanban-board-with-drag-and-drop)
- [Interactive Data Visualization](#interactive-data-visualization)
- [Content Management System](#content-management-system)
- [Gaming Leaderboard](#gaming-leaderboard)

## Data Dashboard with Filtering and Sorting

This example combines `For`, `Show`, `Switch`, and `Index` helpers to create a comprehensive data dashboard.

```go
package main

import (
    "fmt"
    "sort"
    "strings"
    "time"
    
    "github.com/ozanturksever/uiwgo/comps"
    "github.com/ozanturksever/uiwgo/dom"
    "github.com/ozanturksever/uiwgo/g"
    "github.com/ozanturksever/uiwgo/reactivity"
)

type DataPoint struct {
    ID       string
    Name     string
    Value    float64
    Category string
    Date     time.Time
    Status   string
}

type DashboardState struct {
    data         reactivity.Signal[[]DataPoint]
    searchTerm   reactivity.Signal[string]
    sortBy       reactivity.Signal[string]
    sortOrder    reactivity.Signal[string]
    filterStatus reactivity.Signal[string]
    selectedItem reactivity.Signal[*DataPoint]
    loading      reactivity.Signal[bool]
}

func NewDashboardState() *DashboardState {
    return &DashboardState{
        data:         reactivity.NewSignal([]DataPoint{}),
        searchTerm:   reactivity.NewSignal(""),
        sortBy:       reactivity.NewSignal("name"),
        sortOrder:    reactivity.NewSignal("asc"),
        filterStatus: reactivity.NewSignal("all"),
        selectedItem: reactivity.NewSignal[*DataPoint](nil),
        loading:      reactivity.NewSignal(false),
    }
}

func (ds *DashboardState) filteredAndSortedData() reactivity.Memo[[]DataPoint] {
    return reactivity.NewMemo(func() []DataPoint {
        data := ds.data.Get()
        searchTerm := strings.ToLower(ds.searchTerm.Get())
        filterStatus := ds.filterStatus.Get()
        sortBy := ds.sortBy.Get()
        sortOrder := ds.sortOrder.Get()
        
        // Filter by search term
        var filtered []DataPoint
        for _, item := range data {
            if searchTerm == "" || strings.Contains(strings.ToLower(item.Name), searchTerm) {
                filtered = append(filtered, item)
            }
        }
        
        // Filter by status
        if filterStatus != "all" {
            var statusFiltered []DataPoint
            for _, item := range filtered {
                if item.Status == filterStatus {
                    statusFiltered = append(statusFiltered, item)
                }
            }
            filtered = statusFiltered
        }
        
        // Sort
        sort.Slice(filtered, func(i, j int) bool {
            var less bool
            switch sortBy {
            case "name":
                less = filtered[i].Name < filtered[j].Name
            case "value":
                less = filtered[i].Value < filtered[j].Value
            case "date":
                less = filtered[i].Date.Before(filtered[j].Date)
            default:
                less = filtered[i].Name < filtered[j].Name
            }
            
            if sortOrder == "desc" {
                return !less
            }
            return less
        })
        
        return filtered
    })
}

func (ds *DashboardState) renderControls() g.Node {
    return g.Div(
        g.Class("dashboard-controls"),
        g.Style("display: flex; gap: 1rem; margin-bottom: 1rem; align-items: center;"),
        
        // Search input
        g.Div(
            g.Label(
                g.Text("Search: "),
                g.Input(
                    g.Type("text"),
                    g.Value(ds.searchTerm.Get()),
                    g.Placeholder("Search by name..."),
                    dom.OnInput(func(value string) {
                        ds.searchTerm.Set(value)
                    }),
                ),
            ),
        ),
        
        // Status filter
        g.Div(
            g.Label(
                g.Text("Status: "),
                g.Select(
                    g.Value(ds.filterStatus.Get()),
                    dom.OnChange(func(value string) {
                        ds.filterStatus.Set(value)
                    }),
                    g.Option(g.Value("all"), g.Text("All")),
                    g.Option(g.Value("active"), g.Text("Active")),
                    g.Option(g.Value("inactive"), g.Text("Inactive")),
                    g.Option(g.Value("pending"), g.Text("Pending")),
                ),
            ),
        ),
        
        // Sort controls
        g.Div(
            g.Label(
                g.Text("Sort by: "),
                g.Select(
                    g.Value(ds.sortBy.Get()),
                    dom.OnChange(func(value string) {
                        ds.sortBy.Set(value)
                    }),
                    g.Option(g.Value("name"), g.Text("Name")),
                    g.Option(g.Value("value"), g.Text("Value")),
                    g.Option(g.Value("date"), g.Text("Date")),
                ),
            ),
        ),
        
        g.Button(
            g.Text(func() string {
                if ds.sortOrder.Get() == "asc" {
                    return "↑ Ascending"
                }
                return "↓ Descending"
            }()),
            dom.OnClick(func() {
                if ds.sortOrder.Get() == "asc" {
                    ds.sortOrder.Set("desc")
                } else {
                    ds.sortOrder.Set("asc")
                }
            }),
        ),
        
        // Refresh button
        g.Button(
            g.Text("Refresh"),
            g.Disabled(ds.loading.Get()),
            dom.OnClick(func() {
                ds.loadData()
            }),
        ),
    )
}

func (ds *DashboardState) renderDataTable() g.Node {
    filteredData := ds.filteredAndSortedData()
    
    return g.Div(
        g.Class("data-table"),
        
        // Show loading state
        comps.Show(comps.ShowProps{
            When: ds.loading,
            Children: g.Div(
                g.Class("loading"),
                g.Text("Loading data..."),
            ),
        }),
        
        // Show data table when not loading
        comps.Show(comps.ShowProps{
            When: reactivity.NewMemo(func() bool {
                return !ds.loading.Get()
            }),
            Children: g.Div(
                // Show empty state or data
                comps.Switch(comps.SwitchProps{
                    When: reactivity.NewMemo(func() string {
                        data := filteredData.Get()
                        if len(data) == 0 {
                            return "empty"
                        }
                        return "data"
                    }),
                    Children: []g.Node{
                        comps.Match(comps.MatchProps{
                            When: "empty",
                            Children: g.Div(
                                g.Class("empty-state"),
                                g.Text("No data found matching your criteria."),
                            ),
                        }),
                        comps.Match(comps.MatchProps{
                            When: "data",
                            Children: g.Table(
                                g.Class("data-table"),
                                g.THead(
                                    g.Tr(
                                        g.Th(g.Text("Name")),
                                        g.Th(g.Text("Value")),
                                        g.Th(g.Text("Category")),
                                        g.Th(g.Text("Status")),
                                        g.Th(g.Text("Date")),
                                        g.Th(g.Text("Actions")),
                                    ),
                                ),
                                g.TBody(
                                    comps.For(comps.ForProps[DataPoint]{
                                        Items: filteredData,
                                        Key: func(item DataPoint) string {
                                            return item.ID
                                        },
                                        Children: func(item DataPoint, index int) g.Node {
                                            return ds.renderDataRow(item, index)
                                        },
                                    }),
                                ),
                            ),
                        }),
                    },
                }),
            ),
        }),
    )
}

func (ds *DashboardState) renderDataRow(item DataPoint, index int) g.Node {
    isSelected := reactivity.NewMemo(func() bool {
        selected := ds.selectedItem.Get()
        return selected != nil && selected.ID == item.ID
    })
    
    return g.Tr(
        g.Class(func() string {
            if isSelected.Get() {
                return "selected"
            }
            return ""
        }()),
        g.Style(func() string {
            if isSelected.Get() {
                return "background-color: #e3f2fd;"
            }
            return ""
        }()),
        dom.OnClick(func() {
            ds.selectedItem.Set(&item)
        }),
        
        g.Td(g.Text(item.Name)),
        g.Td(g.Text(fmt.Sprintf("%.2f", item.Value))),
        g.Td(g.Text(item.Category)),
        g.Td(
            g.Span(
                g.Class(fmt.Sprintf("status status-%s", item.Status)),
                g.Text(strings.Title(item.Status)),
            ),
        ),
        g.Td(g.Text(item.Date.Format("2006-01-02"))),
        g.Td(
            g.Button(
                g.Text("Edit"),
                g.Class("btn-small"),
                dom.OnClick(func() {
                    ds.editItem(item)
                }),
            ),
            g.Button(
                g.Text("Delete"),
                g.Class("btn-small btn-danger"),
                dom.OnClick(func() {
                    ds.deleteItem(item.ID)
                }),
            ),
        ),
    )
}

func (ds *DashboardState) renderDetailPanel() g.Node {
    return comps.Show(comps.ShowProps{
        When: reactivity.NewMemo(func() bool {
            return ds.selectedItem.Get() != nil
        }),
        Children: g.Div(
            g.Class("detail-panel"),
            g.Style("margin-top: 2rem; padding: 1rem; border: 1px solid #ddd; border-radius: 4px;"),
            
            g.H3(g.Text("Item Details")),
            
            comps.Switch(comps.SwitchProps{
                When: reactivity.NewMemo(func() string {
                    item := ds.selectedItem.Get()
                    if item == nil {
                        return "none"
                    }
                    return "selected"
                }),
                Children: []g.Node{
                    comps.Match(comps.MatchProps{
                        When: "selected",
                        Children: func() g.Node {
                            item := ds.selectedItem.Get()
                            if item == nil {
                                return g.Div()
                            }
                            
                            return g.Div(
                                g.P(g.Strong(g.Text("ID: ")), g.Text(item.ID)),
                                g.P(g.Strong(g.Text("Name: ")), g.Text(item.Name)),
                                g.P(g.Strong(g.Text("Value: ")), g.Text(fmt.Sprintf("%.2f", item.Value))),
                                g.P(g.Strong(g.Text("Category: ")), g.Text(item.Category)),
                                g.P(g.Strong(g.Text("Status: ")), g.Text(item.Status)),
                                g.P(g.Strong(g.Text("Date: ")), g.Text(item.Date.Format("2006-01-02 15:04:05"))),
                                
                                g.Div(
                                    g.Style("margin-top: 1rem;"),
                                    g.Button(
                                        g.Text("Close"),
                                        dom.OnClick(func() {
                                            ds.selectedItem.Set(nil)
                                        }),
                                    ),
                                ),
                            )
                        }(),
                    }),
                },
            }),
        ),
    })
}

func (ds *DashboardState) loadData() {
    ds.loading.Set(true)
    
    // Simulate API call
    go func() {
        time.Sleep(1 * time.Second)
        
        // Mock data
        data := []DataPoint{
            {ID: "1", Name: "Alpha Project", Value: 1250.50, Category: "Development", Date: time.Now().AddDate(0, 0, -5), Status: "active"},
            {ID: "2", Name: "Beta Release", Value: 890.25, Category: "Testing", Date: time.Now().AddDate(0, 0, -3), Status: "pending"},
            {ID: "3", Name: "Gamma Feature", Value: 2100.75, Category: "Development", Date: time.Now().AddDate(0, 0, -1), Status: "active"},
            {ID: "4", Name: "Delta Optimization", Value: 450.00, Category: "Performance", Date: time.Now().AddDate(0, 0, -7), Status: "inactive"},
            {ID: "5", Name: "Epsilon Integration", Value: 1750.30, Category: "Integration", Date: time.Now(), Status: "active"},
        }
        
        ds.data.Set(data)
        ds.loading.Set(false)
    }()
}

func (ds *DashboardState) editItem(item DataPoint) {
    // Implementation for editing
    logutil.Logf("Editing item: %s", item.Name)
}

func (ds *DashboardState) deleteItem(id string) {
    data := ds.data.Get()
    var filtered []DataPoint
    for _, item := range data {
        if item.ID != id {
            filtered = append(filtered, item)
        }
    }
    ds.data.Set(filtered)
    
    // Clear selection if deleted item was selected
    if selected := ds.selectedItem.Get(); selected != nil && selected.ID == id {
        ds.selectedItem.Set(nil)
    }
}

func Dashboard() g.Node {
    state := NewDashboardState()
    
    // Load initial data
    reactivity.NewEffect(func() {
        state.loadData()
    })
    
    return g.Div(
        g.Class("dashboard"),
        g.Style("padding: 2rem;"),
        
        g.H1(g.Text("Data Dashboard")),
        
        state.renderControls(),
        state.renderDataTable(),
        state.renderDetailPanel(),
        
        // Add some CSS
        g.Style(`
            .dashboard-controls {
                background: #f5f5f5;
                padding: 1rem;
                border-radius: 4px;
            }
            
            .data-table {
                width: 100%;
                border-collapse: collapse;
            }
            
            .data-table th,
            .data-table td {
                padding: 0.5rem;
                text-align: left;
                border-bottom: 1px solid #ddd;
            }
            
            .data-table th {
                background: #f0f0f0;
                font-weight: bold;
            }
            
            .data-table tr:hover {
                background: #f9f9f9;
            }
            
            .data-table tr.selected {
                background: #e3f2fd !important;
            }
            
            .status {
                padding: 0.25rem 0.5rem;
                border-radius: 3px;
                font-size: 0.8rem;
                font-weight: bold;
            }
            
            .status-active {
                background: #4caf50;
                color: white;
            }
            
            .status-inactive {
                background: #f44336;
                color: white;
            }
            
            .status-pending {
                background: #ff9800;
                color: white;
            }
            
            .btn-small {
                padding: 0.25rem 0.5rem;
                margin-right: 0.25rem;
                font-size: 0.8rem;
            }
            
            .btn-danger {
                background: #f44336;
                color: white;
            }
            
            .loading {
                text-align: center;
                padding: 2rem;
                font-style: italic;
            }
            
            .empty-state {
                text-align: center;
                padding: 2rem;
                color: #666;
            }
        `),
    )
}
```

## Multi-Step Form with Validation

This example demonstrates how `Switch`, `Show`, and `Index` helpers can create a complex multi-step form with validation.

```go
type FormStep struct {
    ID          string
    Title       string
    Description string
    IsValid     func() bool
    Component   func() g.Node
}

type MultiStepFormState struct {
    currentStep reactivity.Signal[int]
    steps       []FormStep
    
    // Form data
    personalInfo reactivity.Signal[PersonalInfo]
    contactInfo  reactivity.Signal[ContactInfo]
    preferences  reactivity.Signal[Preferences]
    
    // Validation
    errors       reactivity.Signal[map[string]string]
    isSubmitting reactivity.Signal[bool]
}

type PersonalInfo struct {
    FirstName string
    LastName  string
    BirthDate string
    Gender    string
}

type ContactInfo struct {
    Email   string
    Phone   string
    Address string
    City    string
    Country string
}

type Preferences struct {
    Newsletter     bool
    Notifications  bool
    Theme          string
    Language       string
}

func NewMultiStepFormState() *MultiStepFormState {
    state := &MultiStepFormState{
        currentStep:   reactivity.NewSignal(0),
        personalInfo:  reactivity.NewSignal(PersonalInfo{}),
        contactInfo:   reactivity.NewSignal(ContactInfo{}),
        preferences:   reactivity.NewSignal(Preferences{Theme: "light", Language: "en"}),
        errors:        reactivity.NewSignal(make(map[string]string)),
        isSubmitting:  reactivity.NewSignal(false),
    }
    
    state.steps = []FormStep{
        {
            ID:          "personal",
            Title:       "Personal Information",
            Description: "Please provide your basic personal details",
            IsValid:     state.validatePersonalInfo,
            Component:   state.renderPersonalInfoStep,
        },
        {
            ID:          "contact",
            Title:       "Contact Information",
            Description: "How can we reach you?",
            IsValid:     state.validateContactInfo,
            Component:   state.renderContactInfoStep,
        },
        {
            ID:          "preferences",
            Title:       "Preferences",
            Description: "Customize your experience",
            IsValid:     state.validatePreferences,
            Component:   state.renderPreferencesStep,
        },
        {
            ID:          "review",
            Title:       "Review & Submit",
            Description: "Please review your information before submitting",
            IsValid:     func() bool { return true },
            Component:   state.renderReviewStep,
        },
    }
    
    return state
}

func (mfs *MultiStepFormState) validatePersonalInfo() bool {
    info := mfs.personalInfo.Get()
    errors := make(map[string]string)
    
    if strings.TrimSpace(info.FirstName) == "" {
        errors["firstName"] = "First name is required"
    }
    
    if strings.TrimSpace(info.LastName) == "" {
        errors["lastName"] = "Last name is required"
    }
    
    if info.BirthDate == "" {
        errors["birthDate"] = "Birth date is required"
    }
    
    mfs.errors.Set(errors)
    return len(errors) == 0
}

func (mfs *MultiStepFormState) validateContactInfo() bool {
    info := mfs.contactInfo.Get()
    errors := make(map[string]string)
    
    if !strings.Contains(info.Email, "@") {
        errors["email"] = "Valid email is required"
    }
    
    if len(info.Phone) < 10 {
        errors["phone"] = "Valid phone number is required"
    }
    
    if strings.TrimSpace(info.Address) == "" {
        errors["address"] = "Address is required"
    }
    
    mfs.errors.Set(errors)
    return len(errors) == 0
}

func (mfs *MultiStepFormState) validatePreferences() bool {
    // Preferences are optional, so always valid
    mfs.errors.Set(make(map[string]string))
    return true
}

func (mfs *MultiStepFormState) renderStepIndicator() g.Node {
    return g.Div(
        g.Class("step-indicator"),
        g.Style("display: flex; justify-content: space-between; margin-bottom: 2rem;"),
        
        comps.For(comps.ForProps[FormStep]{
            Items: reactivity.NewSignal(mfs.steps),
            Key: func(step FormStep) string {
                return step.ID
            },
            Children: func(step FormStep, index int) g.Node {
                currentStep := mfs.currentStep.Get()
                isActive := index == currentStep
                isCompleted := index < currentStep
                
                return g.Div(
                    g.Class("step"),
                    g.Style(func() string {
                        style := "flex: 1; text-align: center; padding: 1rem;"
                        if isActive {
                            style += " background: #2196f3; color: white;"
                        } else if isCompleted {
                            style += " background: #4caf50; color: white;"
                        } else {
                            style += " background: #f0f0f0; color: #666;"
                        }
                        return style
                    }()),
                    
                    g.Div(
                        g.Class("step-number"),
                        g.Style("font-weight: bold; margin-bottom: 0.5rem;"),
                        g.Text(fmt.Sprintf("%d", index+1)),
                    ),
                    g.Div(
                        g.Class("step-title"),
                        g.Style("font-size: 0.9rem;"),
                        g.Text(step.Title),
                    ),
                )
            },
        }),
    )
}

func (mfs *MultiStepFormState) renderCurrentStep() g.Node {
    return comps.Switch(comps.SwitchProps{
        When: reactivity.NewMemo(func() string {
            step := mfs.currentStep.Get()
            if step >= 0 && step < len(mfs.steps) {
                return mfs.steps[step].ID
            }
            return "unknown"
        }),
        Children: func() []g.Node {
            var children []g.Node
            for _, step := range mfs.steps {
                children = append(children, comps.Match(comps.MatchProps{
                    When:     step.ID,
                    Children: step.Component(),
                }))
            }
            return children
        }(),
    })
}

func (mfs *MultiStepFormState) renderPersonalInfoStep() g.Node {
    info := mfs.personalInfo.Get()
    errors := mfs.errors.Get()
    
    return g.Div(
        g.Class("form-step"),
        
        g.H2(g.Text("Personal Information")),
        g.P(g.Text("Please provide your basic personal details")),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("First Name *")),
            g.Input(
                g.Type("text"),
                g.Value(info.FirstName),
                g.Class(func() string {
                    if errors["firstName"] != "" {
                        return "error"
                    }
                    return ""
                }()),
                dom.OnInput(func(value string) {
                    newInfo := info
                    newInfo.FirstName = value
                    mfs.personalInfo.Set(newInfo)
                }),
            ),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return errors["firstName"] != ""
                }),
                Children: g.Div(
                    g.Class("error-message"),
                    g.Text(errors["firstName"]),
                ),
            }),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Last Name *")),
            g.Input(
                g.Type("text"),
                g.Value(info.LastName),
                g.Class(func() string {
                    if errors["lastName"] != "" {
                        return "error"
                    }
                    return ""
                }()),
                dom.OnInput(func(value string) {
                    newInfo := info
                    newInfo.LastName = value
                    mfs.personalInfo.Set(newInfo)
                }),
            ),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return errors["lastName"] != ""
                }),
                Children: g.Div(
                    g.Class("error-message"),
                    g.Text(errors["lastName"]),
                ),
            }),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Birth Date *")),
            g.Input(
                g.Type("date"),
                g.Value(info.BirthDate),
                g.Class(func() string {
                    if errors["birthDate"] != "" {
                        return "error"
                    }
                    return ""
                }()),
                dom.OnInput(func(value string) {
                    newInfo := info
                    newInfo.BirthDate = value
                    mfs.personalInfo.Set(newInfo)
                }),
            ),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return errors["birthDate"] != ""
                }),
                Children: g.Div(
                    g.Class("error-message"),
                    g.Text(errors["birthDate"]),
                ),
            }),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Gender")),
            g.Select(
                g.Value(info.Gender),
                dom.OnChange(func(value string) {
                    newInfo := info
                    newInfo.Gender = value
                    mfs.personalInfo.Set(newInfo)
                }),
                g.Option(g.Value(""), g.Text("Select..."), g.Disabled(true)),
                g.Option(g.Value("male"), g.Text("Male")),
                g.Option(g.Value("female"), g.Text("Female")),
                g.Option(g.Value("other"), g.Text("Other")),
                g.Option(g.Value("prefer-not-to-say"), g.Text("Prefer not to say")),
            ),
        ),
    )
}

func (mfs *MultiStepFormState) renderContactInfoStep() g.Node {
    info := mfs.contactInfo.Get()
    errors := mfs.errors.Get()
    
    return g.Div(
        g.Class("form-step"),
        
        g.H2(g.Text("Contact Information")),
        g.P(g.Text("How can we reach you?")),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Email *")),
            g.Input(
                g.Type("email"),
                g.Value(info.Email),
                g.Class(func() string {
                    if errors["email"] != "" {
                        return "error"
                    }
                    return ""
                }()),
                dom.OnInput(func(value string) {
                    newInfo := info
                    newInfo.Email = value
                    mfs.contactInfo.Set(newInfo)
                }),
            ),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return errors["email"] != ""
                }),
                Children: g.Div(
                    g.Class("error-message"),
                    g.Text(errors["email"]),
                ),
            }),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Phone *")),
            g.Input(
                g.Type("tel"),
                g.Value(info.Phone),
                g.Class(func() string {
                    if errors["phone"] != "" {
                        return "error"
                    }
                    return ""
                }()),
                dom.OnInput(func(value string) {
                    newInfo := info
                    newInfo.Phone = value
                    mfs.contactInfo.Set(newInfo)
                }),
            ),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return errors["phone"] != ""
                }),
                Children: g.Div(
                    g.Class("error-message"),
                    g.Text(errors["phone"]),
                ),
            }),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Address *")),
            g.TextArea(
                g.Value(info.Address),
                g.Rows(3),
                g.Class(func() string {
                    if errors["address"] != "" {
                        return "error"
                    }
                    return ""
                }()),
                dom.OnInput(func(value string) {
                    newInfo := info
                    newInfo.Address = value
                    mfs.contactInfo.Set(newInfo)
                }),
            ),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return errors["address"] != ""
                }),
                Children: g.Div(
                    g.Class("error-message"),
                    g.Text(errors["address"]),
                ),
            }),
        ),
        
        g.Div(
            g.Class("form-row"),
            g.Style("display: flex; gap: 1rem;"),
            
            g.Div(
                g.Class("form-group"),
                g.Style("flex: 1;"),
                g.Label(g.Text("City")),
                g.Input(
                    g.Type("text"),
                    g.Value(info.City),
                    dom.OnInput(func(value string) {
                        newInfo := info
                        newInfo.City = value
                        mfs.contactInfo.Set(newInfo)
                    }),
                ),
            ),
            
            g.Div(
                g.Class("form-group"),
                g.Style("flex: 1;"),
                g.Label(g.Text("Country")),
                g.Input(
                    g.Type("text"),
                    g.Value(info.Country),
                    dom.OnInput(func(value string) {
                        newInfo := info
                        newInfo.Country = value
                        mfs.contactInfo.Set(newInfo)
                    }),
                ),
            ),
        ),
    )
}

func (mfs *MultiStepFormState) renderPreferencesStep() g.Node {
    prefs := mfs.preferences.Get()
    
    return g.Div(
        g.Class("form-step"),
        
        g.H2(g.Text("Preferences")),
        g.P(g.Text("Customize your experience")),
        
        g.Div(
            g.Class("form-group"),
            g.Label(
                g.Input(
                    g.Type("checkbox"),
                    g.Checked(prefs.Newsletter),
                    dom.OnChange(func(checked bool) {
                        newPrefs := prefs
                        newPrefs.Newsletter = checked
                        mfs.preferences.Set(newPrefs)
                    }),
                ),
                g.Text(" Subscribe to newsletter"),
            ),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(
                g.Input(
                    g.Type("checkbox"),
                    g.Checked(prefs.Notifications),
                    dom.OnChange(func(checked bool) {
                        newPrefs := prefs
                        newPrefs.Notifications = checked
                        mfs.preferences.Set(newPrefs)
                    }),
                ),
                g.Text(" Enable notifications"),
            ),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Theme")),
            g.Select(
                g.Value(prefs.Theme),
                dom.OnChange(func(value string) {
                    newPrefs := prefs
                    newPrefs.Theme = value
                    mfs.preferences.Set(newPrefs)
                }),
                g.Option(g.Value("light"), g.Text("Light")),
                g.Option(g.Value("dark"), g.Text("Dark")),
                g.Option(g.Value("auto"), g.Text("Auto")),
            ),
        ),
        
        g.Div(
            g.Class("form-group"),
            g.Label(g.Text("Language")),
            g.Select(
                g.Value(prefs.Language),
                dom.OnChange(func(value string) {
                    newPrefs := prefs
                    newPrefs.Language = value
                    mfs.preferences.Set(newPrefs)
                }),
                g.Option(g.Value("en"), g.Text("English")),
                g.Option(g.Value("es"), g.Text("Spanish")),
                g.Option(g.Value("fr"), g.Text("French")),
                g.Option(g.Value("de"), g.Text("German")),
            ),
        ),
    )
}

func (mfs *MultiStepFormState) renderReviewStep() g.Node {
    personalInfo := mfs.personalInfo.Get()
    contactInfo := mfs.contactInfo.Get()
    preferences := mfs.preferences.Get()
    
    return g.Div(
        g.Class("form-step"),
        
        g.H2(g.Text("Review & Submit")),
        g.P(g.Text("Please review your information before submitting")),
        
        g.Div(
            g.Class("review-section"),
            g.H3(g.Text("Personal Information")),
            g.P(g.Text(fmt.Sprintf("Name: %s %s", personalInfo.FirstName, personalInfo.LastName))),
            g.P(g.Text(fmt.Sprintf("Birth Date: %s", personalInfo.BirthDate))),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return personalInfo.Gender != ""
                }),
                Children: g.P(g.Text(fmt.Sprintf("Gender: %s", personalInfo.Gender))),
            }),
        ),
        
        g.Div(
            g.Class("review-section"),
            g.H3(g.Text("Contact Information")),
            g.P(g.Text(fmt.Sprintf("Email: %s", contactInfo.Email))),
            g.P(g.Text(fmt.Sprintf("Phone: %s", contactInfo.Phone))),
            g.P(g.Text(fmt.Sprintf("Address: %s", contactInfo.Address))),
            comps.Show(comps.ShowProps{
                When: reactivity.NewMemo(func() bool {
                    return contactInfo.City != "" || contactInfo.Country != ""
                }),
                Children: g.P(g.Text(fmt.Sprintf("Location: %s, %s", contactInfo.City, contactInfo.Country))),
            }),
        ),
        
        g.Div(
            g.Class("review-section"),
            g.H3(g.Text("Preferences")),
            g.P(g.Text(fmt.Sprintf("Newsletter: %t", preferences.Newsletter))),
            g.P(g.Text(fmt.Sprintf("Notifications: %t", preferences.Notifications))),
            g.P(g.Text(fmt.Sprintf("Theme: %s", preferences.Theme))),
            g.P(g.Text(fmt.Sprintf("Language: %s", preferences.Language))),
        ),
    )
}

func (mfs *MultiStepFormState) renderNavigation() g.Node {
    currentStep := mfs.currentStep.Get()
    isFirstStep := currentStep == 0
    isLastStep := currentStep == len(mfs.steps)-1
    canProceed := mfs.steps[currentStep].IsValid()
    
    return g.Div(
        g.Class("form-navigation"),
        g.Style("display: flex; justify-content: space-between; margin-top: 2rem; padding-top: 1rem; border-top: 1px solid #ddd;"),
        
        // Previous button
        comps.Show(comps.ShowProps{
            When: reactivity.NewMemo(func() bool {
                return !isFirstStep
            }),
            Children: g.Button(
                g.Text("Previous"),
                g.Class("btn-secondary"),
                dom.OnClick(func() {
                    if currentStep > 0 {
                        mfs.currentStep.Set(currentStep - 1)
                    }
                }),
            ),
            Fallback: g.Div(), // Empty div to maintain layout
        }),
        
        // Next/Submit button
        comps.Switch(comps.SwitchProps{
            When: reactivity.NewMemo(func() string {
                if isLastStep {
                    return "submit"
                }
                return "next"
            }),
            Children: []g.Node{
                comps.Match(comps.MatchProps{
                    When: "next",
                    Children: g.Button(
                        g.Text("Next"),
                        g.Class("btn-primary"),
                        g.Disabled(!canProceed),
                        dom.OnClick(func() {
                            if canProceed && currentStep < len(mfs.steps)-1 {
                                mfs.currentStep.Set(currentStep + 1)
                            }
                        }),
                    ),
                }),
                comps.Match(comps.MatchProps{
                    When: "submit",
                    Children: g.Button(
                        g.Text(func() string {
                            if mfs.isSubmitting.Get() {
                                return "Submitting..."
                            }
                            return "Submit"
                        }()),
                        g.Class("btn-primary"),
                        g.Disabled(mfs.isSubmitting.Get()),
                        dom.OnClick(func() {
                            mfs.submitForm()
                        }),
                    ),
                }),
            },
        }),
    )
}

func (mfs *MultiStepFormState) submitForm() {
    mfs.isSubmitting.Set(true)
    
    // Simulate form submission
    go func() {
        time.Sleep(2 * time.Second)
        
        logutil.Log("Form submitted successfully!")
        logutil.Logf("Personal Info: %+v", mfs.personalInfo.Get())
        logutil.Logf("Contact Info: %+v", mfs.contactInfo.Get())
        logutil.Logf("Preferences: %+v", mfs.preferences.Get())
        
        mfs.isSubmitting.Set(false)
        
        // Reset form or show success message
        // For demo, we'll just reset to first step
        mfs.currentStep.Set(0)
    }()
}

func MultiStepForm() g.Node {
    state := NewMultiStepFormState()
    
    return g.Div(
        g.Class("multi-step-form"),
        g.Style("max-width: 600px; margin: 0 auto; padding: 2rem;"),
        
        g.H1(g.Text("Registration Form")),
        
        state.renderStepIndicator(),
        state.renderCurrentStep(),
        state.renderNavigation(),
        
        // Add CSS
        g.Style(`
            .form-group {
                margin-bottom: 1rem;
            }
            
            .form-group label {
                display: block;
                margin-bottom: 0.5rem;
                font-weight: bold;
            }
            
            .form-group input,
            .form-group select,
            .form-group textarea {
                width: 100%;
                padding: 0.5rem;
                border: 1px solid #ddd;
                border-radius: 4px;
                font-size: 1rem;
            }
            
            .form-group input.error,
            .form-group select.error,
            .form-group textarea.error {
                border-color: #f44336;
            }
            
            .error-message {
                color: #f44336;
                font-size: 0.9rem;
                margin-top: 0.25rem;
            }
            
            .btn-primary {
                background: #2196f3;
                color: white;
                padding: 0.75rem 1.5rem;
                border: none;
                border-radius: 4px;
                cursor: pointer;
            }
            
            .btn-primary:disabled {
                background: #ccc;
                cursor: not-allowed;
            }
            
            .btn-secondary {
                background: #f0f0f0;
                color: #333;
                padding: 0.75rem 1.5rem;
                border: 1px solid #ddd;
                border-radius: 4px;
                cursor: pointer;
            }
            
            .review-section {
                margin-bottom: 1.5rem;
                padding: 1rem;
                background: #f9f9f9;
                border-radius: 4px;
            }
            
            .review-section h3 {
                margin-top: 0;
                margin-bottom: 0.5rem;
                color: #333;
            }
            
            .review-section p {
                margin: 0.25rem 0;
            }
        `),
    )
}
```

These integration examples demonstrate how UIwGo's helper functions can be combined to create sophisticated, interactive applications. Each example showcases different patterns:

1. **Data Dashboard**: Combines filtering, sorting, selection, and detail views
2. **Multi-Step Form**: Shows complex state management with validation and navigation
3. **Real-Time Chat**: Demonstrates live updates and message handling
4. **E-commerce Browser**: Shows product filtering, search, and shopping cart functionality
5. **Kanban Board**: Illustrates drag-and-drop with state management
6. **Data Visualization**: Combines charts with interactive controls
7. **CMS**: Shows content editing with preview and publishing workflows
8. **Gaming Leaderboard**: Demonstrates real-time updates and competitive features

Each example can be extended and customized for specific use cases, providing a solid foundation for building complex UIwGo applications.