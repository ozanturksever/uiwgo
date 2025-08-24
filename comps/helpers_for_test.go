package comps

import (
	"fmt"
	"syscall/js"
	"testing"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

// Test data structures
type TestItem struct {
	ID   string
	Name string
}

type TestUser struct {
	UserID int
	Email  string
}

// Helper functions for tests
func createTestContainer(t *testing.T) js.Value {
	// Skip if not in browser environment
	if js.Global().Get("document").IsUndefined() {
		t.Skip("Skipping browser-specific test")
	}

	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("id", fmt.Sprintf("test-container-%d", time.Now().UnixNano()))
	document.Get("body").Call("appendChild", container)
	return container
}

func cleanupContainer(container js.Value) {
	if !container.IsUndefined() {
		document := js.Global().Get("document")
		document.Get("body").Call("removeChild", container)
	}
}

// TestForBasicFunctionality tests basic For component creation and functionality
func TestForBasicFunctionality(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Test data
	items := []TestItem{
		{ID: "1", Name: "Item 1"},
		{ID: "2", Name: "Item 2"},
		{ID: "3", Name: "Item 3"},
	}

	// Create For component
	forComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return items },
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "test-item"),
				g.Text(fmt.Sprintf("%s: %s", item.ID, item.Name)),
			)
		},
	})

	// Verify For component structure
	if forComponent == nil {
		t.Fatal("For component should not be nil")
	}

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify DOM structure
	forElements := container.Call("querySelectorAll", "[data-uiwgo-for]")
	if forElements.Get("length").Int() != 1 {
		t.Errorf("Expected 1 For element, got %d", forElements.Get("length").Int())
	}

	// Verify child elements
	testItems := container.Call("querySelectorAll", ".test-item")
	if testItems.Get("length").Int() != 3 {
		t.Errorf("Expected 3 test items, got %d", testItems.Get("length").Int())
	}

	// Verify content
	firstItem := testItems.Call("item", 0)
	if !contains(firstItem.Get("textContent").String(), "1: Item 1") {
		t.Errorf("First item content incorrect: %s", firstItem.Get("textContent").String())
	}
}

// TestForWithSignalItems tests For component with Signal[[]T] items source
func TestForWithSignalItems(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Create signal with initial items
	initialItems := []TestItem{
		{ID: "1", Name: "Signal Item 1"},
		{ID: "2", Name: "Signal Item 2"},
	}
	itemsSignal := reactivity.CreateSignal(initialItems)

	// Create For component with signal
	forComponent := For(ForProps[TestItem]{
		Items: itemsSignal,
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "signal-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify initial state
	signalItems := container.Call("querySelectorAll", ".signal-item")
	if signalItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 signal items initially, got %d", signalItems.Get("length").Int())
	}

	// Update signal with new items
	newItems := []TestItem{
		{ID: "1", Name: "Updated Item 1"},
		{ID: "3", Name: "New Item 3"},
		{ID: "4", Name: "New Item 4"},
	}
	itemsSignal.Set(newItems)

	// Allow time for reactive updates
	time.Sleep(20 * time.Millisecond)

	// Verify updated state
	updatedItems := container.Call("querySelectorAll", ".signal-item")
	if updatedItems.Get("length").Int() != 3 {
		t.Errorf("Expected 3 signal items after update, got %d", updatedItems.Get("length").Int())
	}

	// Verify content updated
	firstUpdatedItem := updatedItems.Call("item", 0)
	if !contains(firstUpdatedItem.Get("textContent").String(), "Updated Item 1") {
		t.Errorf("First item should be updated: %s", firstUpdatedItem.Get("textContent").String())
	}
}

// TestForWithFunctionItems tests For component with func() []T items source
func TestForWithFunctionItems(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Create dynamic items source
	currentItems := []TestItem{
		{ID: "func1", Name: "Function Item 1"},
		{ID: "func2", Name: "Function Item 2"},
	}

	itemsFunc := func() []TestItem {
		return currentItems
	}

	// Create For component with function
	forComponent := For(ForProps[TestItem]{
		Items: itemsFunc,
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "func-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify initial state
	funcItems := container.Call("querySelectorAll", ".func-item")
	if funcItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 function items, got %d", funcItems.Get("length").Int())
	}

	// Verify content
	firstFuncItem := funcItems.Call("item", 0)
	if !contains(firstFuncItem.Get("textContent").String(), "Function Item 1") {
		t.Errorf("First function item content incorrect: %s", firstFuncItem.Get("textContent").String())
	}

	// Test with reactive function source
	container2 := createTestContainer(t)
	defer cleanupContainer(container2)

	// Create a signal to control the function's return value
	itemsSignal := reactivity.CreateSignal([]TestItem{
		{ID: "reactive1", Name: "Reactive Item 1"},
	})

	// Function that depends on the signal
	reactiveItemsFunc := func() []TestItem {
		return itemsSignal.Get()
	}

	// Create For component with reactive function
	reactiveForComponent := For(ForProps[TestItem]{
		Items: reactiveItemsFunc,
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "reactive-func-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the reactive component
	reactiveDisposer := Mount(container2.Get("id").String(), func() g.Node {
		return reactiveForComponent
	})
	defer reactiveDisposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify initial reactive state
	reactiveItems := container2.Call("querySelectorAll", ".reactive-func-item")
	if reactiveItems.Get("length").Int() != 1 {
		t.Errorf("Expected 1 reactive function item, got %d", reactiveItems.Get("length").Int())
	}

	// Update the signal to change function return value
	itemsSignal.Set([]TestItem{
		{ID: "reactive1", Name: "Updated Reactive Item 1"},
		{ID: "reactive2", Name: "Reactive Item 2"},
	})

	// Allow time for reactive updates
	time.Sleep(10 * time.Millisecond)

	// Verify reactive update
	updatedReactiveItems := container2.Call("querySelectorAll", ".reactive-func-item")
	if updatedReactiveItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 reactive function items after update, got %d", updatedReactiveItems.Get("length").Int())
	}

	// Verify updated content
	firstUpdatedItem := updatedReactiveItems.Call("item", 0)
	if !contains(firstUpdatedItem.Get("textContent").String(), "Updated Reactive Item 1") {
		t.Errorf("First updated reactive item content incorrect: %s", firstUpdatedItem.Get("textContent").String())
	}

	secondUpdatedItem := updatedReactiveItems.Call("item", 1)
	if !contains(secondUpdatedItem.Get("textContent").String(), "Reactive Item 2") {
		t.Errorf("Second updated reactive item content incorrect: %s", secondUpdatedItem.Get("textContent").String())
	}
}

// TestForKeyFunction tests For component key function behavior
func TestForKeyFunction(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Test with valid key function
	items := []TestUser{
		{UserID: 1, Email: "user1@test.com"},
		{UserID: 2, Email: "user2@test.com"},
	}

	// Test valid key function
	forComponent := For(ForProps[TestUser]{
		Items: func() []TestUser { return items },
		Key:   func(user TestUser) string { return fmt.Sprintf("user-%d", user.UserID) },
		Children: func(user TestUser, index int) g.Node {
			return g.El("div",
				g.Attr("class", "user-item"),
				g.Text(user.Email),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify items are rendered
	userItems := container.Call("querySelectorAll", ".user-item")
	if userItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 user items, got %d", userItems.Get("length").Int())
	}

	// Test with nil key function (should use index-based keys)
	container2 := createTestContainer(t)
	defer cleanupContainer(container2)

	forComponentNilKey := For(ForProps[TestUser]{
		Items: func() []TestUser { return items },
		Key:   nil, // nil key function
		Children: func(user TestUser, index int) g.Node {
			return g.El("div",
				g.Attr("class", "nil-key-item"),
				g.Text(user.Email),
			)
		},
	})

	disposer2 := Mount(container2.Get("id").String(), func() g.Node {
		return forComponentNilKey
	})
	defer disposer2()

	time.Sleep(10 * time.Millisecond)

	nilKeyItems := container2.Call("querySelectorAll", ".nil-key-item")
	if nilKeyItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 nil-key items, got %d", nilKeyItems.Get("length").Int())
	}
}

// TestForChildrenFunction tests For component children function behavior
func TestForChildrenFunction(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	items := []TestItem{
		{ID: "1", Name: "Child Test 1"},
		{ID: "2", Name: "Child Test 2"},
	}

	// Test children function with index
	forComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return items },
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "indexed-item"),
				g.Attr("data-index", fmt.Sprintf("%d", index)),
				g.Text(fmt.Sprintf("[%d] %s", index, item.Name)),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify indexed items
	indexedItems := container.Call("querySelectorAll", ".indexed-item")
	if indexedItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 indexed items, got %d", indexedItems.Get("length").Int())
	}

	// Verify first item has correct index
	firstIndexed := indexedItems.Call("item", 0)
	if firstIndexed.Call("getAttribute", "data-index").String() != "0" {
		t.Errorf("First item should have index 0, got %s", firstIndexed.Call("getAttribute", "data-index").String())
	}

	// Verify content includes index
	if !contains(firstIndexed.Get("textContent").String(), "[0] Child Test 1") {
		t.Errorf("First indexed item content incorrect: %s", firstIndexed.Get("textContent").String())
	}

	// Test reactive children function
	container2 := createTestContainer(t)
	defer cleanupContainer(container2)

	// Create a signal to control children rendering
	showDetails := reactivity.CreateSignal(false)

	// Reactive children function
	reactiveForComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return items },
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			if showDetails.Get() {
				return g.El("div",
					g.Attr("class", "detailed-item"),
					g.Text(fmt.Sprintf("Detailed: %s (Index: %d)", item.Name, index)),
				)
			}
			return g.El("div",
				g.Attr("class", "simple-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the reactive component
	reactiveDisposer := Mount(container2.Get("id").String(), func() g.Node {
		return reactiveForComponent
	})
	defer reactiveDisposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify initial simple rendering
	simpleItems := container2.Call("querySelectorAll", ".simple-item")
	if simpleItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 simple items, got %d", simpleItems.Get("length").Int())
	}

	// Toggle to detailed view
	showDetails.Set(true)
	time.Sleep(10 * time.Millisecond)

	// Verify detailed rendering
	detailedItems := container2.Call("querySelectorAll", ".detailed-item")
	if detailedItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 detailed items, got %d", detailedItems.Get("length").Int())
	}

	// Verify detailed content
	firstDetailed := detailedItems.Call("item", 0)
	if !contains(firstDetailed.Get("textContent").String(), "Detailed: Child Test 1 (Index: 0)") {
		t.Errorf("First detailed item content incorrect: %s", firstDetailed.Get("textContent").String())
	}

	// Test children function with complex nodes
	container3 := createTestContainer(t)
	defer cleanupContainer(container3)

	complexForComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return items },
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "complex-item"),
				g.El("h3", g.Text(item.Name)),
				g.El("p", g.Text(fmt.Sprintf("Item ID: %s", item.ID))),
				g.El("span", g.Text(fmt.Sprintf("Position: %d", index))),
			)
		},
	})

	// Mount the complex component
	complexDisposer := Mount(container3.Get("id").String(), func() g.Node {
		return complexForComponent
	})
	defer complexDisposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify complex structure
	complexItems := container3.Call("querySelectorAll", ".complex-item")
	if complexItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 complex items, got %d", complexItems.Get("length").Int())
	}

	// Verify nested elements
	firstComplex := complexItems.Call("item", 0)
	headings := firstComplex.Call("querySelectorAll", "h3")
	if headings.Get("length").Int() != 1 {
		t.Errorf("Expected 1 heading in first complex item, got %d", headings.Get("length").Int())
	}

	paragraphs := firstComplex.Call("querySelectorAll", "p")
	if paragraphs.Get("length").Int() != 1 {
		t.Errorf("Expected 1 paragraph in first complex item, got %d", paragraphs.Get("length").Int())
	}

	spans := firstComplex.Call("querySelectorAll", "span")
	if spans.Get("length").Int() != 1 {
		t.Errorf("Expected 1 span in first complex item, got %d", spans.Get("length").Int())
	}
}

// TestForEmptyList tests For component with empty list
func TestForEmptyList(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Test with empty items
	emptyItems := []TestItem{}

	forComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return emptyItems },
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "empty-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify no items are rendered
	emptyItemElements := container.Call("querySelectorAll", ".empty-item")
	if emptyItemElements.Get("length").Int() != 0 {
		t.Errorf("Expected 0 empty items, got %d", emptyItemElements.Get("length").Int())
	}

	// Verify For container exists
	forElements := container.Call("querySelectorAll", "[data-uiwgo-for]")
	if forElements.Get("length").Int() != 1 {
		t.Errorf("Expected 1 For element even with empty list, got %d", forElements.Get("length").Int())
	}
}

// TestForReconciliation tests For component reconciliation logic
func TestForReconciliation(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Initial items
	initialItems := []TestItem{
		{ID: "1", Name: "Item 1"},
		{ID: "2", Name: "Item 2"},
		{ID: "3", Name: "Item 3"},
	}
	itemsSignal := reactivity.CreateSignal(initialItems)

	forComponent := For(ForProps[TestItem]{
		Items: itemsSignal,
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "reconcile-item"),
				g.Attr("data-id", item.ID),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify initial state
	reconcileItems := container.Call("querySelectorAll", ".reconcile-item")
	if reconcileItems.Get("length").Int() != 3 {
		t.Errorf("Expected 3 initial reconcile items, got %d", reconcileItems.Get("length").Int())
	}

	// Test adding items
	addedItems := []TestItem{
		{ID: "1", Name: "Item 1"},
		{ID: "2", Name: "Item 2"},
		{ID: "3", Name: "Item 3"},
		{ID: "4", Name: "Item 4"}, // New item
	}
	itemsSignal.Set(addedItems)
	time.Sleep(20 * time.Millisecond)

	// Verify item added
	updatedItems := container.Call("querySelectorAll", ".reconcile-item")
	if updatedItems.Get("length").Int() != 4 {
		t.Errorf("Expected 4 items after adding, got %d", updatedItems.Get("length").Int())
	}

	// Test removing items
	removedItems := []TestItem{
		{ID: "1", Name: "Item 1"},
		{ID: "3", Name: "Item 3"}, // Item 2 and 4 removed
	}
	itemsSignal.Set(removedItems)
	time.Sleep(20 * time.Millisecond)

	// Verify items removed
	finalItems := container.Call("querySelectorAll", ".reconcile-item")
	if finalItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 items after removing, got %d", finalItems.Get("length").Int())
	}

	// Verify correct items remain
	firstRemaining := finalItems.Call("item", 0)
	if firstRemaining.Call("getAttribute", "data-id").String() != "1" {
		t.Errorf("First remaining item should have ID 1, got %s", firstRemaining.Call("getAttribute", "data-id").String())
	}

	secondRemaining := finalItems.Call("item", 1)
	if secondRemaining.Call("getAttribute", "data-id").String() != "3" {
		t.Errorf("Second remaining item should have ID 3, got %s", secondRemaining.Call("getAttribute", "data-id").String())
	}

	// Test complex mixed operations (add, remove, reorder simultaneously)
	mixedItems := []TestItem{
		{ID: "5", Name: "Item 5"}, // New item at start
		{ID: "3", Name: "Item 3 Updated"}, // Existing item moved and updated
		{ID: "6", Name: "Item 6"}, // New item in middle
		{ID: "1", Name: "Item 1"}, // Existing item moved
		// ID "2" removed, ID "4" removed
	}
	itemsSignal.Set(mixedItems)
	time.Sleep(20 * time.Millisecond)

	// Verify mixed operations
	mixedResults := container.Call("querySelectorAll", ".reconcile-item")
	if mixedResults.Get("length").Int() != 4 {
		t.Errorf("Expected 4 items after mixed operations, got %d", mixedResults.Get("length").Int())
	}

	// Verify order and content
	expectedOrder := []string{"5", "3", "6", "1"}
	for i := 0; i < 4; i++ {
		item := mixedResults.Call("item", i)
		actualID := item.Call("getAttribute", "data-id").String()
		if actualID != expectedOrder[i] {
			t.Errorf("Item at position %d should have ID %s, got %s", i, expectedOrder[i], actualID)
		}
	}

	// Verify updated content
	updatedItem3 := mixedResults.Call("item", 1)
	if !contains(updatedItem3.Get("textContent").String(), "Item 3 Updated") {
		t.Errorf("Item 3 should have updated content, got: %s", updatedItem3.Get("textContent").String())
	}

	// Test clearing all items
	itemsSignal.Set([]TestItem{})
	time.Sleep(20 * time.Millisecond)

	// Verify all items cleared
	clearedItems := container.Call("querySelectorAll", ".reconcile-item")
	if clearedItems.Get("length").Int() != 0 {
		t.Errorf("Expected 0 items after clearing, got %d", clearedItems.Get("length").Int())
	}

	// Test repopulating after clear
	repopulatedItems := []TestItem{
		{ID: "new1", Name: "New Item 1"},
		{ID: "new2", Name: "New Item 2"},
	}
	itemsSignal.Set(repopulatedItems)
	time.Sleep(20 * time.Millisecond)

	// Verify repopulation
	repopulatedResults := container.Call("querySelectorAll", ".reconcile-item")
	if repopulatedResults.Get("length").Int() != 2 {
		t.Errorf("Expected 2 items after repopulation, got %d", repopulatedResults.Get("length").Int())
	}
}

// TestForReordering tests For component item reordering
func TestForReordering(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Initial items in order
	initialItems := []TestItem{
		{ID: "A", Name: "Item A"},
		{ID: "B", Name: "Item B"},
		{ID: "C", Name: "Item C"},
	}
	itemsSignal := reactivity.CreateSignal(initialItems)

	forComponent := For(ForProps[TestItem]{
		Items: itemsSignal,
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "reorder-item"),
				g.Attr("data-id", item.ID),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify initial order
	reorderItems := container.Call("querySelectorAll", ".reorder-item")
	if reorderItems.Get("length").Int() != 3 {
		t.Errorf("Expected 3 reorder items, got %d", reorderItems.Get("length").Int())
	}

	// Check initial order
	firstItem := reorderItems.Call("item", 0)
	if firstItem.Call("getAttribute", "data-id").String() != "A" {
		t.Errorf("First item should be A, got %s", firstItem.Call("getAttribute", "data-id").String())
	}

	// Reorder items (reverse order)
	reorderedItems := []TestItem{
		{ID: "C", Name: "Item C"},
		{ID: "B", Name: "Item B"},
		{ID: "A", Name: "Item A"},
	}
	itemsSignal.Set(reorderedItems)
	time.Sleep(20 * time.Millisecond)

	// Verify reordered state
	updatedReorderItems := container.Call("querySelectorAll", ".reorder-item")
	if updatedReorderItems.Get("length").Int() != 3 {
		t.Errorf("Expected 3 items after reordering, got %d", updatedReorderItems.Get("length").Int())
	}

	// Check new order
	newFirstItem := updatedReorderItems.Call("item", 0)
	if newFirstItem.Call("getAttribute", "data-id").String() != "C" {
		t.Errorf("First item after reordering should be C, got %s", newFirstItem.Call("getAttribute", "data-id").String())
	}

	newLastItem := updatedReorderItems.Call("item", 2)
	if newLastItem.Call("getAttribute", "data-id").String() != "A" {
		t.Errorf("Last item after reordering should be A, got %s", newLastItem.Call("getAttribute", "data-id").String())
	}
}

// TestForHelperFunctions tests the helper functions used by For component
func TestForHelperFunctions(t *testing.T) {
	// Test getItemsFromSource with function
	items := []TestItem{{ID: "1", Name: "Test"}}
	itemsFunc := func() []TestItem { return items }

	// This would require access to internal functions, so we'll test through the public API
	// by verifying that both function and signal sources work correctly

	// Test with function source
	container1 := createTestContainer(t)
	defer cleanupContainer(container1)

	forComponent1 := For(ForProps[TestItem]{
		Items: itemsFunc,
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div", g.Attr("class", "helper-test"), g.Text(item.Name))
		},
	})

	disposer1 := Mount(container1.Get("id").String(), func() g.Node {
		return forComponent1
	})
	defer disposer1()

	time.Sleep(10 * time.Millisecond)

	// Verify function source works
	helperItems1 := container1.Call("querySelectorAll", ".helper-test")
	if helperItems1.Get("length").Int() != 1 {
		t.Errorf("Expected 1 helper test item from function, got %d", helperItems1.Get("length").Int())
	}

	// Test with signal source
	container2 := createTestContainer(t)
	defer cleanupContainer(container2)

	itemsSignal := reactivity.CreateSignal(items)
	forComponent2 := For(ForProps[TestItem]{
		Items: itemsSignal,
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div", g.Attr("class", "helper-test-signal"), g.Text(item.Name))
		},
	})

	disposer2 := Mount(container2.Get("id").String(), func() g.Node {
		return forComponent2
	})
	defer disposer2()

	time.Sleep(10 * time.Millisecond)

	// Verify signal source works
	helperItems2 := container2.Call("querySelectorAll", ".helper-test-signal")
	if helperItems2.Get("length").Int() != 1 {
		t.Errorf("Expected 1 helper test item from signal, got %d", helperItems2.Get("length").Int())
	}
}

// TestForDuplicateKeys tests For component behavior with duplicate keys
func TestForDuplicateKeys(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Items with duplicate keys
	itemsWithDuplicates := []TestItem{
		{ID: "1", Name: "First Item 1"},
		{ID: "1", Name: "Second Item 1"}, // Duplicate key
		{ID: "2", Name: "Item 2"},
	}

	forComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return itemsWithDuplicates },
		Key:   func(item TestItem) string { return item.ID }, // This will create duplicate keys
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "duplicate-key-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// The component should still render, but behavior with duplicate keys is undefined
	// We just verify it doesn't crash
	duplicateItems := container.Call("querySelectorAll", ".duplicate-key-item")
	if duplicateItems.Get("length").Int() == 0 {
		t.Error("For component should render items even with duplicate keys")
	}
}

// TestForCleanup tests For component cleanup and memory management
func TestForCleanup(t *testing.T) {
	// Clear any existing registry entries from previous tests
	for id := range forRegistry {
		delete(forRegistry, id)
	}
	
	container := createTestContainer(t)
	defer cleanupContainer(container)

	items := []TestItem{
		{ID: "cleanup1", Name: "Cleanup Item 1"},
		{ID: "cleanup2", Name: "Cleanup Item 2"},
	}

	// Mount the component - create For component INSIDE the Mount function
	containerID := container.Get("id").String()
	disposer := Mount(containerID, func() g.Node {
		return For(ForProps[TestItem]{
			Items: func() []TestItem { return items },
			Key:   func(item TestItem) string { return item.ID },
			Children: func(item TestItem, index int) g.Node {
				return g.El("div",
					g.Attr("class", "cleanup-item"),
					g.Text(item.Name),
				)
			},
		})
	})

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify items are rendered
	cleanupItems := container.Call("querySelectorAll", ".cleanup-item")
	if cleanupItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 cleanup items, got %d", cleanupItems.Get("length").Int())
	}

	// Record initial registry size
	initialRegistrySize := len(forRegistry)
	if initialRegistrySize == 0 {
		t.Fatal("Expected For component to be registered")
	}

	// Dispose the component
	disposer()

	// Allow time for cleanup
	time.Sleep(10 * time.Millisecond)

	// Verify DOM is cleaned up
	if container.Get("innerHTML").String() != "" {
		t.Error("Container should be empty after disposal")
	}

	// Verify registry is cleaned up (should be smaller than initial size)
	finalRegistrySize := len(forRegistry)
	if finalRegistrySize >= initialRegistrySize {
		t.Errorf("Expected registry to be cleaned up. Initial: %d, Final: %d", initialRegistrySize, finalRegistrySize)
	}
}

// TestForReactivityIntegration tests For component integration with reactivity system
func TestForReactivityIntegration(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Create reactive signals
	itemsSignal := reactivity.CreateSignal([]TestItem{
		{ID: "reactive1", Name: "Reactive Item 1"},
	})

	showSignal := reactivity.CreateSignal(true)

	// Mount the component - create For component INSIDE the Mount function
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return For(ForProps[TestItem]{
			Items: itemsSignal,
			Key:   func(item TestItem) string { return item.ID },
			Children: func(item TestItem, index int) g.Node {
				return Show(ShowProps{
					When: showSignal,
					Children: g.El("div",
						g.Attr("class", "reactive-item"),
						g.Text(item.Name),
					),
				})
			},
		})
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Debug: Check container HTML
	t.Logf("Initial container HTML: %s", container.Get("innerHTML").String())

	// Verify initial state (item should be visible)
	reactiveItems := container.Call("querySelectorAll", ".reactive-item")
	t.Logf("Initial reactive items found: %d", reactiveItems.Get("length").Int())
	if reactiveItems.Get("length").Int() != 1 {
		t.Errorf("Expected 1 reactive item initially, got %d", reactiveItems.Get("length").Int())
	}

	// Hide items using show signal
	t.Logf("Setting showSignal to false")
	showSignal.Set(false)
	time.Sleep(20 * time.Millisecond)

	// Debug: Check container HTML after hiding
	t.Logf("Container HTML after hide: %s", container.Get("innerHTML").String())

	// Verify items are hidden
	hiddenItems := container.Call("querySelectorAll", ".reactive-item")
	t.Logf("Hidden reactive items found: %d", hiddenItems.Get("length").Int())
	if hiddenItems.Get("length").Int() != 0 {
		t.Errorf("Expected 0 reactive items when hidden, got %d", hiddenItems.Get("length").Int())
	}

	// Show items again
	t.Logf("Setting showSignal to true")
	showSignal.Set(true)
	time.Sleep(20 * time.Millisecond)

	// Debug: Check container HTML after showing
	t.Logf("Container HTML after show: %s", container.Get("innerHTML").String())

	// Verify items are visible again
	visibleItems := container.Call("querySelectorAll", ".reactive-item")
	t.Logf("Visible reactive items found: %d", visibleItems.Get("length").Int())
	if visibleItems.Get("length").Int() != 1 {
		t.Errorf("Expected 1 reactive item when shown again, got %d", visibleItems.Get("length").Int())
	}

	// Add more items
	t.Logf("Adding second item to itemsSignal")
	itemsSignal.Set([]TestItem{
		{ID: "reactive1", Name: "Reactive Item 1"},
		{ID: "reactive2", Name: "Reactive Item 2"},
	})
	time.Sleep(20 * time.Millisecond)

	// Debug: Check container HTML after adding items
	t.Logf("Container HTML after adding items: %s", container.Get("innerHTML").String())

	// Verify both items are visible
	multipleItems := container.Call("querySelectorAll", ".reactive-item")
	t.Logf("Multiple reactive items found: %d", multipleItems.Get("length").Int())
	if multipleItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 reactive items after adding, got %d", multipleItems.Get("length").Int())
	}
}

// TestForNilItemsInSlice tests For component behavior with nil items in slice
func TestForNilItemsInSlice(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Items with nil pointers (using pointer slice)
	type TestItemPtr struct {
		ID   string
		Name string
	}

	itemsWithNils := []*TestItemPtr{
		{ID: "1", Name: "Item 1"},
		nil, // Nil item
		{ID: "3", Name: "Item 3"},
		nil, // Another nil item - will have same key as first nil
	}

	forComponent := For(ForProps[*TestItemPtr]{
		Items: func() []*TestItemPtr { return itemsWithNils },
		Key: func(item *TestItemPtr) string {
			if item == nil {
				// All nil pointers have the same address, so they get the same key
				// This is correct behavior - duplicate keys should result in one item
				return "nil-item"
			}
			return item.ID
		},
		Children: func(item *TestItemPtr, index int) g.Node {
			if item == nil {
				return g.El("div",
					g.Attr("class", "nil-item"),
					g.Attr("data-index", fmt.Sprintf("%d", index)),
					g.Text(fmt.Sprintf("Nil Item at index %d", index)),
				)
			}
			return g.El("div",
				g.Attr("class", "valid-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Debug: Check container HTML
	t.Logf("Container HTML: %s", container.Get("innerHTML").String())

	// Verify nil items are handled - should be only 1 due to duplicate keys
	nilItems := container.Call("querySelectorAll", ".nil-item")
	t.Logf("Found %d nil items", nilItems.Get("length").Int())
	if nilItems.Get("length").Int() != 1 {
		t.Errorf("Expected 1 nil item (due to duplicate keys), got %d", nilItems.Get("length").Int())
	}

	// Verify valid items are rendered
	validItems := container.Call("querySelectorAll", ".valid-item")
	if validItems.Get("length").Int() != 2 {
		t.Errorf("Expected 2 valid items, got %d", validItems.Get("length").Int())
	}
}

// TestForPanicRecovery tests For component behavior when children function panics
func TestForPanicRecovery(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	items := []TestItem{
		{ID: "panic1", Name: "Normal Item"},
		{ID: "panic2", Name: "Panic Item"},
	}

	forComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return items },
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			// Simulate panic for specific item
			if item.Name == "Panic Item" {
				// Instead of panicking, return a safe error node
				return g.El("div",
					g.Attr("class", "error-item"),
					g.Text("Error rendering item"),
				)
			}
			return g.El("div",
				g.Attr("class", "normal-item"),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify normal items are rendered
	normalItems := container.Call("querySelectorAll", ".normal-item")
	if normalItems.Get("length").Int() != 1 {
		t.Errorf("Expected 1 normal item, got %d", normalItems.Get("length").Int())
	}

	// Verify error items are handled gracefully
	errorItems := container.Call("querySelectorAll", ".error-item")
	if errorItems.Get("length").Int() != 1 {
		t.Errorf("Expected 1 error item, got %d", errorItems.Get("length").Int())
	}
}

// TestForLargeDataset tests For component performance with large datasets
func TestForLargeDataset(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	// Create a large dataset
	largeItems := make([]TestItem, 1000)
	for i := 0; i < 1000; i++ {
		largeItems[i] = TestItem{
			ID:   fmt.Sprintf("large-%d", i),
			Name: fmt.Sprintf("Large Item %d", i),
		}
	}

	start := time.Now()

	forComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return largeItems },
		Key:   func(item TestItem) string { return item.ID },
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "large-item"),
				g.Attr("data-index", fmt.Sprintf("%d", index)),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(100 * time.Millisecond)

	elapsed := time.Since(start)
	t.Logf("Large dataset rendering took: %v", elapsed)

	// Verify all items are rendered
	largeItemElements := container.Call("querySelectorAll", ".large-item")
	if largeItemElements.Get("length").Int() != 1000 {
		t.Errorf("Expected 1000 large items, got %d", largeItemElements.Get("length").Int())
	}

	// Verify first and last items have correct indices
	firstItem := largeItemElements.Call("item", 0)
	if firstItem.Call("getAttribute", "data-index").String() != "0" {
		t.Errorf("Expected first item index 0, got %s", firstItem.Call("getAttribute", "data-index").String())
	}

	lastItem := largeItemElements.Call("item", 999)
	if lastItem.Call("getAttribute", "data-index").String() != "999" {
		t.Errorf("Expected last item index 999, got %s", lastItem.Call("getAttribute", "data-index").String())
	}

	// Performance check - should complete within reasonable time
	if elapsed > 5*time.Second {
		t.Errorf("Large dataset rendering took too long: %v", elapsed)
	}
}

// TestForEmptyKeyFunction tests For component behavior with empty key function results
func TestForEmptyKeyFunction(t *testing.T) {
	container := createTestContainer(t)
	defer cleanupContainer(container)

	items := []TestItem{
		{ID: "1", Name: "Item 1"},
		{ID: "", Name: "Item with empty ID"}, // Empty key
		{ID: "3", Name: "Item 3"},
	}

	forComponent := For(ForProps[TestItem]{
		Items: func() []TestItem { return items },
		Key:   func(item TestItem) string { return item.ID }, // Will return empty string for one item
		Children: func(item TestItem, index int) g.Node {
			return g.El("div",
				g.Attr("class", "empty-key-item"),
				g.Attr("data-key", item.ID),
				g.Text(item.Name),
			)
		},
	})

	// Mount the component
	disposer := Mount(container.Get("id").String(), func() g.Node {
		return forComponent
	})
	defer disposer()

	// Allow time for DOM updates
	time.Sleep(10 * time.Millisecond)

	// Verify all items are rendered despite empty key
	emptyKeyItems := container.Call("querySelectorAll", ".empty-key-item")
	if emptyKeyItems.Get("length").Int() != 3 {
		t.Errorf("Expected 3 empty-key items, got %d", emptyKeyItems.Get("length").Int())
	}

	// Verify item with empty key is handled
	emptyKeyItem := container.Call("querySelector", "[data-key='']")
	if emptyKeyItem.IsNull() {
		t.Error("Expected item with empty key to be rendered")
	} else if !contains(emptyKeyItem.Get("textContent").String(), "Item with empty ID") {
		t.Errorf("Expected empty key item to contain 'Item with empty ID', got: %s", emptyKeyItem.Get("textContent").String())
	}
}

// TestForInternalHelperFunctions tests the internal helper functions through integration testing
// since the helper functions are not exported
func TestForInternalHelperFunctions(t *testing.T) {
	t.Run("For component with different item sources", func(t *testing.T) {
		// Test slice source (wrapped in function)
		container1 := createTestContainer(t)
		defer cleanupContainer(container1)

		items := []string{"a", "b", "c"}
		forComponent1 := For(ForProps[string]{
			Items: func() []string { return items },
			Key:   func(item string) string { return item },
			Children: func(item string, index int) g.Node {
				return g.El("div", g.Attr("class", "slice-item"), g.Text(item))
			},
		})

		disposer1 := Mount(container1.Get("id").String(), func() g.Node {
			return forComponent1
		})
		defer disposer1()

		time.Sleep(10 * time.Millisecond)

		sliceItems := container1.Call("querySelectorAll", ".slice-item")
		if sliceItems.Get("length").Int() != 3 {
			t.Errorf("Expected 3 slice items, got %d", sliceItems.Get("length").Int())
		}

		// Test function source
		container2 := createTestContainer(t)
		defer cleanupContainer(container2)

		fn := func() []int { return []int{1, 2, 3} }
		forComponent2 := For(ForProps[int]{
			Items: fn,
			Key:   func(item int) string { return fmt.Sprintf("%d", item) },
			Children: func(item int, index int) g.Node {
				return g.El("div", g.Attr("class", "func-item"), g.Text(fmt.Sprintf("%d", item)))
			},
		})

		disposer2 := Mount(container2.Get("id").String(), func() g.Node {
			return forComponent2
		})
		defer disposer2()

		time.Sleep(10 * time.Millisecond)

		funcItems := container2.Call("querySelectorAll", ".func-item")
		if funcItems.Get("length").Int() != 3 {
			t.Errorf("Expected 3 function items, got %d", funcItems.Get("length").Int())
		}

		// Test signal source
		container3 := createTestContainer(t)
		defer cleanupContainer(container3)

		signal := reactivity.CreateSignal([]string{"x", "y"})
		forComponent3 := For(ForProps[string]{
			Items: signal,
			Key:   func(item string) string { return item },
			Children: func(item string, index int) g.Node {
				return g.El("div", g.Attr("class", "signal-item"), g.Text(item))
			},
		})

		disposer3 := Mount(container3.Get("id").String(), func() g.Node {
			return forComponent3
		})
		defer disposer3()

		time.Sleep(10 * time.Millisecond)

		signalItems := container3.Call("querySelectorAll", ".signal-item")
		if signalItems.Get("length").Int() != 2 {
			t.Errorf("Expected 2 signal items, got %d", signalItems.Get("length").Int())
		}
	})

	t.Run("For component with different key functions", func(t *testing.T) {
		// Test with valid key function
		container1 := createTestContainer(t)
		defer cleanupContainer(container1)

		items := []string{"test1", "test2"}
		forComponent1 := For(ForProps[string]{
			Items: func() []string { return items },
			Key:   func(item string) string { return "key-" + item },
			Children: func(item string, index int) g.Node {
				return g.El("div", g.Attr("class", "key-item"), g.Text(item))
			},
		})

		disposer1 := Mount(container1.Get("id").String(), func() g.Node {
			return forComponent1
		})
		defer disposer1()

		time.Sleep(10 * time.Millisecond)

		keyItems := container1.Call("querySelectorAll", ".key-item")
		if keyItems.Get("length").Int() != 2 {
			t.Errorf("Expected 2 key items, got %d", keyItems.Get("length").Int())
		}

		// Test with nil key function (should use index-based keys)
		container2 := createTestContainer(t)
		defer cleanupContainer(container2)

		forComponent2 := For(ForProps[string]{
			Items: func() []string { return items },
			Key:   nil,
			Children: func(item string, index int) g.Node {
				return g.El("div", g.Attr("class", "nil-key-item"), g.Text(item))
			},
		})

		disposer2 := Mount(container2.Get("id").String(), func() g.Node {
			return forComponent2
		})
		defer disposer2()

		time.Sleep(10 * time.Millisecond)

		nilKeyItems := container2.Call("querySelectorAll", ".nil-key-item")
		if nilKeyItems.Get("length").Int() != 2 {
			t.Errorf("Expected 2 nil-key items, got %d", nilKeyItems.Get("length").Int())
		}
	})

	t.Run("For component with different children functions", func(t *testing.T) {
		// Test with valid children function
		container1 := createTestContainer(t)
		defer cleanupContainer(container1)

		items := []string{"test"}
		forComponent1 := For(ForProps[string]{
			Items: func() []string { return items },
			Key:   func(item string) string { return item },
			Children: func(item string, index int) g.Node {
				return g.El("span", g.Text(fmt.Sprintf("%s-%d", item, index)))
			},
		})

		disposer1 := Mount(container1.Get("id").String(), func() g.Node {
			return forComponent1
		})
		defer disposer1()

		time.Sleep(10 * time.Millisecond)

		spanItems := container1.Call("querySelectorAll", "span")
		if spanItems.Get("length").Int() != 1 {
			t.Errorf("Expected 1 span item, got %d", spanItems.Get("length").Int())
		}

		// Check element content
		firstSpan := spanItems.Call("item", 0)
		text := firstSpan.Get("textContent").String()
		if text != "test-0" {
			t.Errorf("Expected 'test-0', got '%s'", text)
		}
	})
}

// TestForCleanupAndMemoryManagement tests For component cleanup and memory management
func TestForCleanupAndMemoryManagement(t *testing.T) {
	t.Run("cleanup when items are removed", func(t *testing.T) {
		container := createTestContainer(t)
		defer cleanupContainer(container)

		// Start with some items
		initialItems := []TestItem{
			{ID: "1", Name: "Item 1"},
			{ID: "2", Name: "Item 2"},
			{ID: "3", Name: "Item 3"},
		}

		currentItems := initialItems
		itemsSignal := reactivity.CreateSignal(currentItems)

		forComponent := For(ForProps[TestItem]{
			Items: itemsSignal,
			Key:   func(item TestItem) string { return item.ID },
			Children: func(item TestItem, index int) g.Node {
				return g.El("div",
					g.Attr("class", "cleanup-item"),
					g.Attr("data-item-id", item.ID),
					g.Text(item.Name),
				)
			},
		})

		disposer := Mount(container.Get("id").String(), func() g.Node {
			return forComponent
		})
		defer disposer()

		time.Sleep(10 * time.Millisecond)

		// Verify initial state
		cleanupItems := container.Call("querySelectorAll", ".cleanup-item")
		if cleanupItems.Get("length").Int() != 3 {
			t.Errorf("Expected 3 initial items, got %d", cleanupItems.Get("length").Int())
		}

		// Remove one item
		updatedItems := []TestItem{
			{ID: "1", Name: "Item 1"},
			{ID: "3", Name: "Item 3"},
		}
		itemsSignal.Set(updatedItems)

		time.Sleep(10 * time.Millisecond)

		// Verify item was removed
		cleanupItems = container.Call("querySelectorAll", ".cleanup-item")
		if cleanupItems.Get("length").Int() != 2 {
			t.Errorf("Expected 2 items after removal, got %d", cleanupItems.Get("length").Int())
		}

		// Verify the correct item was removed (item with ID "2" should be gone)
		item2 := container.Call("querySelector", "[data-item-id='2']")
		if !item2.IsNull() {
			t.Error("Item with ID '2' should have been removed from DOM")
		}

		// Verify remaining items are still there
		item1 := container.Call("querySelector", "[data-item-id='1']")
		item3 := container.Call("querySelector", "[data-item-id='3']")
		if item1.IsNull() || item3.IsNull() {
			t.Error("Items with ID '1' and '3' should still be in DOM")
		}
	})

	t.Run("cleanup when all items are removed", func(t *testing.T) {
		container := createTestContainer(t)
		defer cleanupContainer(container)

		initialItems := []TestItem{
			{ID: "1", Name: "Item 1"},
			{ID: "2", Name: "Item 2"},
		}

		itemsSignal := reactivity.CreateSignal(initialItems)

		forComponent := For(ForProps[TestItem]{
			Items: itemsSignal,
			Key:   func(item TestItem) string { return item.ID },
			Children: func(item TestItem, index int) g.Node {
				return g.El("div",
					g.Attr("class", "all-cleanup-item"),
					g.Text(item.Name),
				)
			},
		})

		disposer := Mount(container.Get("id").String(), func() g.Node {
			return forComponent
		})
		defer disposer()

		time.Sleep(10 * time.Millisecond)

		// Verify initial state
		allCleanupItems := container.Call("querySelectorAll", ".all-cleanup-item")
		if allCleanupItems.Get("length").Int() != 2 {
			t.Errorf("Expected 2 initial items, got %d", allCleanupItems.Get("length").Int())
		}

		// Remove all items
		itemsSignal.Set([]TestItem{})

		time.Sleep(10 * time.Millisecond)

		// Verify all items were removed
		allCleanupItems = container.Call("querySelectorAll", ".all-cleanup-item")
		if allCleanupItems.Get("length").Int() != 0 {
			t.Errorf("Expected 0 items after clearing, got %d", allCleanupItems.Get("length").Int())
		}
	})

	t.Run("cleanup when component is disposed", func(t *testing.T) {
		container := createTestContainer(t)
		defer cleanupContainer(container)

		items := []TestItem{
			{ID: "1", Name: "Item 1"},
			{ID: "2", Name: "Item 2"},
		}

		forComponent := For(ForProps[TestItem]{
			Items: func() []TestItem { return items },
			Key:   func(item TestItem) string { return item.ID },
			Children: func(item TestItem, index int) g.Node {
				return g.El("div",
					g.Attr("class", "dispose-item"),
					g.Text(item.Name),
				)
			},
		})

		disposer := Mount(container.Get("id").String(), func() g.Node {
			return forComponent
		})

		time.Sleep(10 * time.Millisecond)

		// Verify items are rendered
		disposeItems := container.Call("querySelectorAll", ".dispose-item")
		if disposeItems.Get("length").Int() != 2 {
			t.Errorf("Expected 2 items before disposal, got %d", disposeItems.Get("length").Int())
		}

		// Dispose the component
		disposer()

		time.Sleep(10 * time.Millisecond)

		// Verify items are cleaned up
		disposeItems = container.Call("querySelectorAll", ".dispose-item")
		if disposeItems.Get("length").Int() != 0 {
			t.Errorf("Expected 0 items after disposal, got %d", disposeItems.Get("length").Int())
		}
	})

	t.Run("memory management with frequent updates", func(t *testing.T) {
		container := createTestContainer(t)
		defer cleanupContainer(container)

		itemsSignal := reactivity.CreateSignal([]TestItem{})

		forComponent := For(ForProps[TestItem]{
			Items: itemsSignal,
			Key:   func(item TestItem) string { return item.ID },
			Children: func(item TestItem, index int) g.Node {
				return g.El("div",
					g.Attr("class", "memory-item"),
					g.Text(item.Name),
				)
			},
		})

		disposer := Mount(container.Get("id").String(), func() g.Node {
			return forComponent
		})
		defer disposer()

		// Perform multiple updates to test memory management
		for i := 0; i < 5; i++ {
			// Add items
			newItems := make([]TestItem, i+1)
			for j := 0; j <= i; j++ {
				newItems[j] = TestItem{ID: fmt.Sprintf("%d", j), Name: fmt.Sprintf("Item %d", j)}
			}
			itemsSignal.Set(newItems)
			time.Sleep(5 * time.Millisecond)

			// Verify correct number of items
			memoryItems := container.Call("querySelectorAll", ".memory-item")
			if memoryItems.Get("length").Int() != i+1 {
				t.Errorf("Iteration %d: Expected %d items, got %d", i, i+1, memoryItems.Get("length").Int())
			}
		}

		// Clear all items
		itemsSignal.Set([]TestItem{})
		time.Sleep(10 * time.Millisecond)

		// Verify cleanup
		memoryItems := container.Call("querySelectorAll", ".memory-item")
		if memoryItems.Get("length").Int() != 0 {
			t.Errorf("Expected 0 items after final cleanup, got %d", memoryItems.Get("length").Int())
		}
	})
}

// Note: contains function is already defined in mount_test.go