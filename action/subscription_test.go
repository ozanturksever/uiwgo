package action

import (
	"testing"
)

func TestNoOpSubscription(t *testing.T) {
	// Test NoOpSubscription creation
	sub := NewNoOpSubscription()

	// Test that subscription is active by default
	if !sub.IsActive() {
		t.Error("Expected NoOpSubscription to be active by default")
	}

	// Test Dispose method
	err := sub.Dispose()
	if err != nil {
		t.Errorf("Expected Dispose to return nil error, got %v", err)
	}

	// Test that subscription is no longer active after disposal
	if sub.IsActive() {
		t.Error("Expected NoOpSubscription to be inactive after Dispose")
	}

	// Test that multiple Dispose calls are safe
	err = sub.Dispose()
	if err != nil {
		t.Errorf("Expected multiple Dispose calls to be safe, got error: %v", err)
	}
}

func TestNoOpSubscriptionMultipleInstances(t *testing.T) {
	// Test that multiple NoOpSubscription instances are independent
	sub1 := NewNoOpSubscription()
	sub2 := NewNoOpSubscription()

	// Both should be active initially
	if !sub1.IsActive() || !sub2.IsActive() {
		t.Error("Expected both subscriptions to be active initially")
	}

	// Dispose one
	err := sub1.Dispose()
	if err != nil {
		t.Errorf("Expected Dispose to return nil error, got %v", err)
	}

	// First should be inactive, second should still be active
	if sub1.IsActive() {
		t.Error("Expected first subscription to be inactive after Dispose")
	}
	if !sub2.IsActive() {
		t.Error("Expected second subscription to remain active")
	}
}

func TestSubscriptionInterface(t *testing.T) {
	// Test that NoOpSubscription implements the Subscription interface
	var sub Subscription = NewNoOpSubscription()

	// Test interface methods
	if !sub.IsActive() {
		t.Error("Expected subscription to be active via interface")
	}

	err := sub.Dispose()
	if err != nil {
		t.Errorf("Expected Dispose to work via interface, got %v", err)
	}

	if sub.IsActive() {
		t.Error("Expected subscription to be inactive via interface after Dispose")
	}
}
