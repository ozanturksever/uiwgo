//go:build js && wasm

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ozanturksever/logutil"
	comps "github.com/ozanturksever/uiwgo/comps"
	dom "github.com/ozanturksever/uiwgo/dom"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

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
	tasks          reactivity.Signal[[]Task]
	currentView    reactivity.Signal[DashboardView]
	selectedTags   reactivity.Signal[[]string]
	assigneeFilter reactivity.Signal[string]
	showCompleted  reactivity.Signal[bool]
	searchTerm     reactivity.Signal[string]
}

func NewTaskDashboard() *TaskDashboard {
	td := &TaskDashboard{
		tasks:          reactivity.CreateSignal([]Task{}),
		currentView:    reactivity.CreateSignal(ViewBoard),
		selectedTags:   reactivity.CreateSignal([]string{}),
		assigneeFilter: reactivity.CreateSignal(""),
		showCompleted:  reactivity.CreateSignal(false),
		searchTerm:     reactivity.CreateSignal(""),
	}
	td.loadSampleData()
	return td
}

func (td *TaskDashboard) loadSampleData() {
	sampleTasks := []Task{
		{
			ID:          "1",
			Title:       "Design new homepage",
			Description: "Create wireframes and mockups for the new homepage design",
			Status:      TaskStatusTodo,
			Priority:    TaskPriorityHigh,
			Assignee:    "Alice",
			DueDate:     time.Now().AddDate(0, 0, 7),
			CreatedAt:   time.Now().AddDate(0, 0, -2),
			Tags:        []string{"design", "frontend"},
		},
		{
			ID:          "2",
			Title:       "Implement user authentication",
			Description: "Add login and registration functionality",
			Status:      TaskStatusInProgress,
			Priority:    TaskPriorityHigh,
			Assignee:    "Bob",
			DueDate:     time.Now().AddDate(0, 0, 5),
			CreatedAt:   time.Now().AddDate(0, 0, -5),
			Tags:        []string{"backend", "security"},
		},
		{
			ID:          "3",
			Title:       "Write unit tests",
			Description: "Add comprehensive test coverage for core modules",
			Status:      TaskStatusTodo,
			Priority:    TaskPriorityMedium,
			Assignee:    "Charlie",
			DueDate:     time.Now().AddDate(0, 0, 10),
			CreatedAt:   time.Now().AddDate(0, 0, -1),
			Tags:        []string{"testing", "quality"},
		},
		{
			ID:          "4",
			Title:       "Setup CI/CD pipeline",
			Description: "Configure automated testing and deployment",
			Status:      TaskStatusDone,
			Priority:    TaskPriorityMedium,
			Assignee:    "David",
			DueDate:     time.Now().AddDate(0, 0, -3),
			CreatedAt:   time.Now().AddDate(0, 0, -10),
			Tags:        []string{"devops", "automation"},
		},
		{
			ID:          "5",
			Title:       "Database optimization",
			Description: "Optimize database queries and add indexes",
			Status:      TaskStatusInProgress,
			Priority:    TaskPriorityLow,
			Assignee:    "Eve",
			DueDate:     time.Now().AddDate(0, 0, 14),
			CreatedAt:   time.Now().AddDate(0, 0, -3),
			Tags:        []string{"database", "performance"},
		},
	}
	td.tasks.Set(sampleTasks)
}

func (td *TaskDashboard) render() Node {
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

	// Setup DOM event handlers after mount
	comps.OnMount(func() {
		// View tab click handlers
		views := []DashboardView{ViewBoard, ViewList, ViewCalendar, ViewAnalytics}
		for _, view := range views {
			if btn := dom.GetElementByID(fmt.Sprintf("view-tab-%s", view)); btn != nil {
				capturedView := view // Capture for closure
				dom.BindClickToCallback(btn, func() {
					td.currentView.Set(capturedView)
				})
			}
		}

		// Search input handler
		if searchInput := dom.GetElementByID("search-input"); searchInput != nil {
			dom.BindInputToSignal(searchInput, td.searchTerm)
		}

		// Show completed checkbox handler
		if checkbox := dom.GetElementByID("show-completed-checkbox"); checkbox != nil {
			dom.BindClickToCallback(checkbox, func() {
				td.showCompleted.Set(!td.showCompleted.Get())
			})
		}
	})

	return Div(
		Class("task-dashboard"),

		// Header with navigation and filters
		Header(
			Class("dashboard-header"),
			H1(Text("Task Dashboard")),

			// View switcher
			Nav(
				Class("view-switcher"),
				comps.For(comps.ForProps[DashboardView]{
					Items: reactivity.CreateSignal([]DashboardView{
						ViewBoard, ViewList, ViewCalendar, ViewAnalytics,
					}),
					Key: func(view DashboardView) string { return string(view) },
					Children: func(view DashboardView, index int) Node {
						isActive := reactivity.CreateMemo(func() bool {
							return td.currentView.Get() == view
						})

						return Button(
						ID(fmt.Sprintf("view-tab-%s", view)),
						Class("view-tab"),
						If(isActive.Get(), Class("active")),
						Text(strings.Title(string(view))),
					)
					},
				}),
			),

			// Filters
			Div(
				Class("filters"),
				Input(
					ID("search-input"),
					Type("text"),
					Attr("placeholder", "Search tasks..."),
					Attr("value", td.searchTerm.Get()),
				),

				Label(
					Input(
						ID("show-completed-checkbox"),
						Type("checkbox"),
						If(td.showCompleted.Get(), Attr("checked", "")),
					),
					Text("Show completed"),
				),
			),
		),

		// Main content area
		Main(
			Class("dashboard-content"),
			comps.Switch(comps.SwitchProps{
				When: td.currentView,
				Children: []Node{
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

func (td *TaskDashboard) renderKanbanBoard(tasks reactivity.Signal[[]Task]) Node {
	// Group tasks by status
	tasksByStatus := reactivity.CreateMemo(func() map[TaskStatus][]Task {
		groups := make(map[TaskStatus][]Task)
		for _, task := range tasks.Get() {
			groups[task.Status] = append(groups[task.Status], task)
		}
		return groups
	})

	statuses := []TaskStatus{TaskStatusTodo, TaskStatusInProgress, TaskStatusDone}

	return Div(
		Class("kanban-board"),
		comps.For(comps.ForProps[TaskStatus]{
				Items: reactivity.CreateSignal(statuses),
			Key:   func(status TaskStatus) string { return string(status) },
			Children: func(status TaskStatus, index int) Node {
					columnTasks := reactivity.CreateMemo(func() []Task {
						return tasksByStatus.Get()[status]
					})

				return Div(
					Class("kanban-column"),
					H3(
						Class("column-header"),
						Text(strings.Title(strings.ReplaceAll(string(status), "_", " "))),
						Span(
							Class("task-count"),
							Text(fmt.Sprintf("(%d)", len(columnTasks.Get()))),
						),
					),

					Div(
						Class("column-tasks"),
						comps.For(comps.ForProps[Task]{
							Items: columnTasks,
							Key:   func(task Task) string { return task.ID },
							Children: func(task Task, index int) Node {
								return td.renderTaskCard(task)
							},
						}),
					),
				)
			},
		}),
	)
}

func (td *TaskDashboard) renderTaskCard(task Task) Node {
	return Div(
		Class("task-card"),
		Class(fmt.Sprintf("priority-%s", task.Priority)),

		H4(
			Class("task-title"),
			Text(task.Title),
		),

		comps.Show(comps.ShowProps{
			When: reactivity.CreateSignal(task.Description != ""),
			Children: P(
				Class("task-description"),
				Text(task.Description),
			),
		}),

		// Tags
		comps.Show(comps.ShowProps{
			When: reactivity.CreateSignal(len(task.Tags) > 0),
			Children: Div(
				Class("task-tags"),
				comps.For(comps.ForProps[string]{
					Items: reactivity.CreateSignal(task.Tags),
					Key:   func(tag string) string { return tag },
					Children: func(tag string, index int) Node {
						return Span(
							Class("tag"),
							Text(tag),
						)
					},
				}),
			),
		}),

		// Task meta
		Div(
			Class("task-meta"),
			comps.Show(comps.ShowProps{
				When: reactivity.CreateSignal(task.Assignee != ""),
				Children: Span(
				Class("assignee"),
				Text(task.Assignee),
			),
		}),

		comps.Show(comps.ShowProps{
			When: reactivity.CreateSignal(!task.DueDate.IsZero()),
			Children: Span(
				Class("due-date"),
				Text(task.DueDate.Format("Jan 2")),
			),
		}),
		),
	)
}

func (td *TaskDashboard) renderTaskList(tasks reactivity.Signal[[]Task]) Node {
	return Div(
		Class("task-list"),
		comps.For(comps.ForProps[Task]{
			Items: tasks,
			Key:   func(task Task) string { return task.ID },
			Children: func(task Task, index int) Node {
				return Div(
					Class("task-list-item"),
					Class(fmt.Sprintf("priority-%s", task.Priority)),
					Class(fmt.Sprintf("status-%s", task.Status)),

					Div(
						Class("task-content"),
						H4(Text(task.Title)),
						comps.Show(comps.ShowProps{
						When: reactivity.CreateSignal(task.Description != ""),
							Children: P(Text(task.Description)),
						}),
					),

					Div(
						Class("task-info"),
						Span(
						Class("status"),
						Text(strings.Title(strings.ReplaceAll(string(task.Status), "_", " "))),
					),
					Span(
						Class("priority"),
						Text(strings.Title(string(task.Priority))),
					),
					comps.Show(comps.ShowProps{
					When: reactivity.CreateSignal(task.Assignee != ""),
						Children: Span(
							Class("assignee"),
							Text(task.Assignee),
						),
					}),
					),
				)
			},
		}),
	)
}

func (td *TaskDashboard) renderCalendarView(tasks reactivity.Signal[[]Task]) Node {
	return Div(
		Class("calendar-view"),
		H3(Text("Calendar View")),
		P(Text("Calendar view implementation would go here")),
		// Simplified calendar view for demo
		comps.For(comps.ForProps[Task]{
			Items: tasks,
			Key:   func(task Task) string { return task.ID },
			Children: func(task Task, index int) Node {
				if task.DueDate.IsZero() {
					return Text("")
				}
				return Div(
					Class("calendar-task"),
					Text(fmt.Sprintf("%s - %s", task.DueDate.Format("Jan 2"), task.Title)),
				)
			},
		}),
	)
}

func (td *TaskDashboard) renderAnalyticsView(tasks reactivity.Signal[[]Task]) Node {
	// Calculate analytics
	analytics := reactivity.CreateMemo(func() map[string]int {
		stats := make(map[string]int)
		for _, task := range tasks.Get() {
			stats["total"]++
			stats[string(task.Status)]++
			stats[string(task.Priority)+"_priority"]++
		}
		return stats
	})

	return Div(
		Class("analytics-view"),
		H3(Text("Analytics")),

		Div(
			Class("analytics-grid"),

			Div(
				Class("stat-card"),
				H4(Text("Total Tasks")),
				P(Text(fmt.Sprintf("%d", analytics.Get()["total"]))),
			),

			Div(
				Class("stat-card"),
				H4(Text("Todo")),
				P(Text(fmt.Sprintf("%d", analytics.Get()["todo"]))),
			),

			Div(
				Class("stat-card"),
				H4(Text("In Progress")),
				P(Text(fmt.Sprintf("%d", analytics.Get()["in_progress"]))),
			),

			Div(
				Class("stat-card"),
				H4(Text("Done")),
				P(Text(fmt.Sprintf("%d", analytics.Get()["done"]))),
			),

			Div(
				Class("stat-card"),
				H4(Text("High Priority")),
				P(Text(fmt.Sprintf("%d", analytics.Get()["high_priority"]))),
			),

			Div(
				Class("stat-card"),
				H4(Text("Medium Priority")),
				P(Text(fmt.Sprintf("%d", analytics.Get()["medium_priority"]))),
			),
		),
	)
}

func main() {
	logutil.Log("Task Dashboard starting...")

	dashboard := NewTaskDashboard()

	// Mount the app and get a disposer function
	disposer := comps.Mount("app", func() Node { return dashboard.render() })
	_ = disposer // We don't use it in this example since the app runs indefinitely

	logutil.Log("Task Dashboard initialized")

	// Prevent exit
	select {}
}