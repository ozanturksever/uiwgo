package action

import "sync"

// Subscription represents a subscription to actions or events.
// It can be disposed to stop receiving notifications.
type Subscription interface {
	// Dispose stops the subscription and releases resources.
	// Returns an error if the subscription is already disposed.
	Dispose() error

	// IsActive returns true if the subscription is active and can receive events.
	IsActive() bool
}

// NoOpSubscription is a no-operation subscription that does nothing.
// It's used as a safe return value when a real subscription cannot be created.
type NoOpSubscription struct {
	once   sync.Once
	active bool
}

// NewNoOpSubscription creates a new NoOpSubscription.
func NewNoOpSubscription() *NoOpSubscription {
	return &NoOpSubscription{
		active: true,
	}
}

// Dispose stops the subscription (no-op implementation).
func (n *NoOpSubscription) Dispose() error {
	n.once.Do(func() {
		n.active = false
	})
	return nil
}

// IsActive returns true if the subscription is active.
func (n *NoOpSubscription) IsActive() bool {
	return n.active
}
