package main

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestTaskDashboard_InitialRender(t *testing.T) {
	server := testhelpers.NewViteServer("task_dashboard", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	config := testhelpers.DefaultConfig()
	config.Timeout = 60 * time.Second
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	var title string
	var headerText string
	var viewTabs []string

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		testhelpers.Actions.WaitForWASMInit(".task-dashboard", 3*time.Second),
		chromedp.WaitVisible(".dashboard-header", chromedp.ByQuery),
		chromedp.Title(&title),
		chromedp.Text(".dashboard-header h1", &headerText, chromedp.ByQuery),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.view-tab')).map(el => el.textContent)`, &viewTabs),
	)

	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if title != "Task Dashboard" {
		t.Errorf("Expected title 'Task Dashboard', got '%s'", title)
	}

	if headerText != "Task Dashboard" {
		t.Errorf("Expected header 'Task Dashboard', got '%s'", headerText)
	}

	expectedTabs := []string{"Board", "List", "Calendar", "Analytics"}
	if len(viewTabs) != len(expectedTabs) {
		t.Errorf("Expected %d view tabs, got %d", len(expectedTabs), len(viewTabs))
	}

	for i, expected := range expectedTabs {
		if i < len(viewTabs) && viewTabs[i] != expected {
			t.Errorf("Expected tab %d to be '%s', got '%s'", i, expected, viewTabs[i])
		}
	}
}

func TestTaskDashboard_KanbanBoard(t *testing.T) {
	server := testhelpers.NewViteServer("task_dashboard", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var kanbanColumns []string
	var taskCards int

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(".dashboard-header"),
		chromedp.Sleep(2*time.Second),
		// Wait for kanban board to render
		chromedp.WaitVisible(".kanban-board"),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.column-header')).map(el => el.textContent.split('(')[0].trim())`, &kanbanColumns),
		chromedp.Evaluate(`document.querySelectorAll('.task-card').length`, &taskCards),
	)

	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	expectedColumns := []string{"Todo", "In Progress", "Done"}
	if len(kanbanColumns) != len(expectedColumns) {
		t.Errorf("Expected %d kanban columns, got %d", len(expectedColumns), len(kanbanColumns))
	}

	for i, expected := range expectedColumns {
		if i < len(kanbanColumns) && kanbanColumns[i] != expected {
			t.Errorf("Expected column %d to be '%s', got '%s'", i, expected, kanbanColumns[i])
		}
	}

	if taskCards == 0 {
		t.Error("Expected to see task cards in kanban board")
	}
}

// TODO: Fix timing issues with view switching test
// func TestTaskDashboard_ViewSwitching(t *testing.T) {
// 	server := testhelpers.NewViteServer("task_dashboard", "localhost:0")
// 	if err := server.Start(); err != nil {
// 		t.Fatalf("Failed to start dev server: %v", err)
// 	}
// 	defer server.Stop()
//
// 	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
// 	defer chromedpCtx.Cancel()
//
// 	err := chromedp.Run(chromedpCtx.Ctx,
// 		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
// 		testhelpers.Actions.WaitForWASMInit(".task-dashboard", 3*time.Second),
// 		// Wait for dashboard header to appear
// 		chromedp.WaitVisible(".dashboard-header", chromedp.ByQuery),
// 		// Wait for initial kanban board
// 		chromedp.WaitVisible(".kanban-board", chromedp.ByQuery),
// 		// Wait for view tabs to be rendered and event handlers bound
// 		chromedp.WaitVisible("#view-tab-list", chromedp.ByID),
// 		chromedp.Sleep(2*time.Second), // Give time for event handlers to bind
// 		// Click List view tab
// 		chromedp.Click("#view-tab-list", chromedp.ByID),
// 		chromedp.Sleep(1*time.Second),
// 		// Wait for list view to appear
// 		chromedp.WaitVisible(".task-list", chromedp.ByQuery),
// 	)
//
// 	if err != nil {
// 		t.Fatalf("View switching test failed: %v", err)
// 	}
// }

func TestTaskDashboard_SearchFilter(t *testing.T) {
	server := testhelpers.NewViteServer("task_dashboard", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var initialTaskCount int
	var filteredTaskCount int

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(".dashboard-header"),
		chromedp.Sleep(2*time.Second),
		// Wait for kanban board to render
		chromedp.WaitVisible(".kanban-board"),
		// Count initial tasks
		chromedp.Evaluate(`document.querySelectorAll('.task-card').length`, &initialTaskCount),
		// Type in search box
		chromedp.SendKeys(`.filters input[type="text"]`, "design"),
		chromedp.Sleep(500*time.Millisecond),
		// Count filtered tasks
		chromedp.Evaluate(`document.querySelectorAll('.task-card').length`, &filteredTaskCount),
	)

	if err != nil {
		t.Fatalf("Search filter test failed: %v", err)
	}

	if initialTaskCount == 0 {
		t.Error("Expected to see initial tasks")
	}

	if filteredTaskCount >= initialTaskCount {
		t.Errorf("Expected filtered task count (%d) to be less than initial count (%d)", filteredTaskCount, initialTaskCount)
	}
}

func TestTaskDashboard_ShowCompletedFilter(t *testing.T) {
	server := testhelpers.NewViteServer("task_dashboard", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var initialTaskCount int
	var withCompletedTaskCount int

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(".dashboard-header"),
		chromedp.Sleep(2*time.Second),
		// Wait for kanban board to render
		chromedp.WaitVisible(".kanban-board"),
		// Count initial tasks (should exclude completed by default)
		chromedp.Evaluate(`document.querySelectorAll('.task-card').length`, &initialTaskCount),
		// Click show completed checkbox
		chromedp.Click(`.filters input[type="checkbox"]`),
		chromedp.Sleep(500*time.Millisecond),
		// Count tasks with completed shown
		chromedp.Evaluate(`document.querySelectorAll('.task-card').length`, &withCompletedTaskCount),
	)

	if err != nil {
		t.Fatalf("Show completed filter test failed: %v", err)
	}

	if initialTaskCount == 0 {
		t.Error("Expected to see initial tasks")
	}

	if withCompletedTaskCount <= initialTaskCount {
		t.Errorf("Expected task count with completed (%d) to be greater than initial count (%d)", withCompletedTaskCount, initialTaskCount)
	}
}

func TestTaskDashboard_TaskCardContent(t *testing.T) {
	server := testhelpers.NewViteServer("task_dashboard", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var hasTaskTitles bool
	var hasTaskDescriptions bool
	var hasTaskTags bool
	var hasTaskMeta bool

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(".dashboard-header"),
		chromedp.Sleep(2*time.Second),
		// Wait for kanban board to render
		chromedp.WaitVisible(".kanban-board"),
		// Check for task card elements
		chromedp.Evaluate(`document.querySelectorAll('.task-title').length > 0`, &hasTaskTitles),
		chromedp.Evaluate(`document.querySelectorAll('.task-description').length > 0`, &hasTaskDescriptions),
		chromedp.Evaluate(`document.querySelectorAll('.task-tags').length > 0`, &hasTaskTags),
		chromedp.Evaluate(`document.querySelectorAll('.task-meta').length > 0`, &hasTaskMeta),
	)

	if err != nil {
		t.Fatalf("Task card content test failed: %v", err)
	}

	if !hasTaskTitles {
		t.Error("Expected to see task titles")
	}

	if !hasTaskDescriptions {
		t.Error("Expected to see task descriptions")
	}

	if !hasTaskTags {
		t.Error("Expected to see task tags")
	}

	if !hasTaskMeta {
		t.Error("Expected to see task meta information")
	}
}

func TestTaskDashboard_AnalyticsView(t *testing.T) {
	server := testhelpers.NewViteServer("task_dashboard", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var statCards int
	var totalTasksText string

	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(".dashboard-header"),
		chromedp.Sleep(2*time.Second),
		// Click Analytics view tab
		chromedp.Click(".view-tab:nth-child(4)"),
		chromedp.Sleep(500*time.Millisecond),
		// Wait for analytics view to appear
		chromedp.WaitVisible(".analytics-view"),
		// Count stat cards
		chromedp.Evaluate(`document.querySelectorAll('.stat-card').length`, &statCards),
		// Get total tasks text
		chromedp.Text(".stat-card:first-child p", &totalTasksText),
	)

	if err != nil {
		t.Fatalf("Analytics view test failed: %v", err)
	}

	if statCards == 0 {
		t.Error("Expected to see stat cards in analytics view")
	}

	if totalTasksText == "" || totalTasksText == "0" {
		t.Errorf("Expected to see total tasks count, got '%s'", totalTasksText)
	}
}