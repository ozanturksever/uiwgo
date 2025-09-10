package action

import (
	"errors"
)

// Predefined errors for the action system
var (
	// ErrNoHandler is returned when no handler is found for an action or query
	ErrNoHandler = errors.New("no handler found for action")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")

	// ErrDisposed is returned when trying to use a disposed subscription or resource
	ErrDisposed = errors.New("resource has been disposed")
)

// dispatchError represents an error that occurred during dispatch
type dispatchError struct {
	context   Context
	err       error
	recovered any
}

func (e *dispatchError) Error() string {
	return e.err.Error()
}

// panicError represents a panic that was recovered
type panicError struct {
	value any
}

func (e *panicError) Error() string {
	return "panic recovered"
}
