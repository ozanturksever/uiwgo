package router

import (
	"testing"
)

// TestLocationState_InitialGetReturnsZeroValue verifies that a newly created LocationState
// returns a zero-value Location when Get() is called.
func TestLocationState_InitialGetReturnsZeroValue(t *testing.T) {
	state := NewLocationState()
	location := state.Get()

	// Check that all fields are zero values
	if location.Pathname != "" {
		t.Errorf("Expected empty Pathname, got %q", location.Pathname)
	}
	if location.Search != "" {
		t.Errorf("Expected empty Search, got %q", location.Search)
	}
	if location.Hash != "" {
		t.Errorf("Expected empty Hash, got %q", location.Hash)
	}
	if location.State != nil {
		t.Errorf("Expected nil State, got %v", location.State)
	}
}

// TestLocationState_SubscribeAddsSubscriber verifies that subscribers can be added
// to the LocationState and are stored correctly.
func TestLocationState_SubscribeAddsSubscriber(t *testing.T) {
	state := NewLocationState()

	// Create a subscriber that increments a counter when called
	var callCount int
	subscriber := func(loc Location) {
		callCount++
	}

	// Subscribe the function
	state.Subscribe(subscriber)

	// Verify the subscriber was added by triggering a notification
	testLocation := Location{Pathname: "/test"}
	state.Set(testLocation)

	// The subscriber should have been called once
	if callCount != 1 {
		t.Errorf("Expected subscriber to be called once, got %d calls", callCount)
	}
}

// TestLocationState_SetNotifiesSubscribers verifies that all subscribers are notified
// when the location state is updated.
func TestLocationState_SetNotifiesSubscribers(t *testing.T) {
	state := NewLocationState()

	// Create multiple subscribers
	var sub1Called, sub2Called, sub3Called bool
	subscriber1 := func(loc Location) {
		sub1Called = true
	}
	subscriber2 := func(loc Location) {
		sub2Called = true
	}
	subscriber3 := func(loc Location) {
		sub3Called = true
	}

	// Subscribe all functions
	state.Subscribe(subscriber1)
	state.Subscribe(subscriber2)
	state.Subscribe(subscriber3)

	// Update the state to trigger notifications
	testLocation := Location{Pathname: "/notify-test"}
	state.Set(testLocation)

	// All subscribers should have been called
	if !sub1Called {
		t.Error("Subscriber1 was not called")
	}
	if !sub2Called {
		t.Error("Subscriber2 was not called")
	}
	if !sub3Called {
		t.Error("Subscriber3 was not called")
	}
}

// TestLocationState_GetReturnsUpdatedValueAfterSet verifies that Get() returns
// the updated value after Set() is called.
func TestLocationState_GetReturnsUpdatedValueAfterSet(t *testing.T) {
	state := NewLocationState()

	// Initial state should be zero value
	initial := state.Get()
	if initial.Pathname != "" {
		t.Errorf("Expected initial Pathname to be empty, got %q", initial.Pathname)
	}

	// Set a new location
	newLocation := Location{
		Pathname: "/updated-path",
		Search:   "?query=test",
		Hash:     "#section",
		State:    "custom state",
	}
	state.Set(newLocation)

	// Get should return the updated value
	updated := state.Get()
	if updated.Pathname != newLocation.Pathname {
		t.Errorf("Expected Pathname %q, got %q", newLocation.Pathname, updated.Pathname)
	}
	if updated.Search != newLocation.Search {
		t.Errorf("Expected Search %q, got %q", newLocation.Search, updated.Search)
	}
	if updated.Hash != newLocation.Hash {
		t.Errorf("Expected Hash %q, got %q", newLocation.Hash, updated.Hash)
	}
	if updated.State != newLocation.State {
		t.Errorf("Expected State %v, got %v", newLocation.State, updated.State)
	}
}
