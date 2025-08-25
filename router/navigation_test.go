package router

import (
	"testing"
)

// TestNavigation_LinkClickTriggersUpdate is an integration test that verifies
// clicking a link created by router.A updates the router's state.
func TestNavigation_LinkClickTriggersUpdate(t *testing.T) {
	// Create a Router instance with a mock outlet
	router := New([]*RouteDefinition{}, nil)

	// Initial location should be zero value
	initialLocation := router.Location()
	if initialLocation.Pathname != "" {
		t.Errorf("Expected initial pathname to be empty, got %s", initialLocation.Pathname)
	}

	// Create a link using router.A - for now, we'll represent it as a simple struct
	// since gomponents is not yet a dependency
	link := A("/about", "About")

	// Type assert to access the OnClick handler
	linkStruct, ok := link.(struct {
		Href    string
		OnClick func()
	})
	if !ok {
		t.Fatalf("Expected A() to return a struct with Href and OnClick, got %T", link)
	}

	// Verify the Href is correct
	if linkStruct.Href != "/about" {
		t.Errorf("Expected href to be '/about', got %s", linkStruct.Href)
	}

	// Simulate a click by calling the OnClick handler
	linkStruct.OnClick()

	// Verify that the router's LocationState has been updated
	updatedLocation := router.Location()
	if updatedLocation.Pathname != "/about" {
		t.Errorf("Expected pathname to be '/about' after click, got %s", updatedLocation.Pathname)
	}
}

// TestNavigatePushUpdatesHistoryAndState verifies that calling router.Navigate updates the router's state.
func TestNavigatePushUpdatesHistoryAndState(t *testing.T) {
	// Create a Router instance with a mock outlet
	router := New([]*RouteDefinition{}, nil)

	// Initial location should be zero value
	initialLocation := router.Location()
	if initialLocation.Pathname != "" {
		t.Errorf("Expected initial pathname to be empty, got %s", initialLocation.Pathname)
	}

	// Call router.Navigate with a new path
	router.Navigate("/contact", NavigateOptions{})

	// Verify that the router's LocationState has been updated
	updatedLocation := router.Location()
	if updatedLocation.Pathname != "/contact" {
		t.Errorf("Expected pathname to be '/contact' after Navigate, got %s", updatedLocation.Pathname)
	}
}
