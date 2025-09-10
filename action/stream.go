package action

import (
	"errors"
	"sync"
)

// Stream represents a typed buffered channel-like abstraction for receiving action payloads.
// It provides backpressure handling through bounded buffers and drop policies.
type Stream[T any] interface {
	// Recv blocks until a value is available or the stream is disposed, returning the value and true.
	// Returns zero value and false if the stream is disposed.
	Recv() (T, bool)

	// TryRecv attempts to receive a value without blocking.
	// Returns the value and true if available, zero value and false if not available or disposed.
	TryRecv() (T, bool)

	// Dispose closes the stream and releases associated resources.
	// After disposal, Recv and TryRecv will return false.
	Dispose()

	// IsDisposed returns true if the stream has been disposed.
	IsDisposed() bool
}

// DropPolicy defines how the stream handles overflow when the buffer is full.
type DropPolicy int

const (
	// DropOldest removes the oldest item in the buffer when full.
	DropOldest DropPolicy = iota
	// DropNewest ignores new items when the buffer is full.
	DropNewest
	// DropAll clears the entire buffer when full (keeps only the new item).
	DropAll
)

// streamImpl is the internal implementation of Stream[T].
type streamImpl[T any] struct {
	mu           sync.Mutex
	buffer       []T
	capacity     int
	dropPolicy   DropPolicy
	disposed     bool
	recvCond     *sync.Cond
	subscription Subscription
}

// NewStream creates a new Stream with the specified buffer capacity and drop policy.
func NewStreamWithSubscription[T any](capacity int, dropPolicy DropPolicy, subscription Subscription) Stream[T] {
	if capacity <= 0 {
		capacity = 1 // Minimum capacity of 1
	}

	s := &streamImpl[T]{
		buffer:       make([]T, 0, capacity),
		capacity:     capacity,
		dropPolicy:   dropPolicy,
		subscription: subscription,
	}
	s.recvCond = sync.NewCond(&s.mu)
	return s
}

// NewStream creates a new Stream with the specified buffer capacity and drop policy.
func NewStream[T any](capacity int, dropPolicy DropPolicy) Stream[T] {
	return NewStreamWithSubscription[T](capacity, dropPolicy, nil)
}

// Recv blocks until a value is available or the stream is disposed.
func (s *streamImpl[T]) Recv() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for len(s.buffer) == 0 && !s.disposed {
		s.recvCond.Wait()
	}

	if s.disposed {
		var zero T
		return zero, false
	}

	// Remove from front of buffer (FIFO)
	value := s.buffer[0]
	s.buffer = s.buffer[1:]
	return value, true
}

// TryRecv attempts to receive a value without blocking.
func (s *streamImpl[T]) TryRecv() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.buffer) == 0 || s.disposed {
		var zero T
		return zero, false
	}

	// Remove from front of buffer (FIFO)
	value := s.buffer[0]
	s.buffer = s.buffer[1:]
	return value, true
}

// Dispose closes the stream and releases associated resources.
func (s *streamImpl[T]) Dispose() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.disposed {
		return
	}

	s.disposed = true
	s.buffer = nil         // Help GC
	s.recvCond.Broadcast() // Wake up any waiting receivers

	// Dispose the subscription if it exists
	if s.subscription != nil {
		s.subscription.Dispose()
	}
}

// IsDisposed returns true if the stream has been disposed.
func (s *streamImpl[T]) IsDisposed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disposed
}

// push adds a value to the stream, applying the drop policy if the buffer is full.
// This method is used internally by the bridge implementations.
func (s *streamImpl[T]) push(value T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.disposed {
		return false
	}

	// If buffer is full, apply drop policy
	if len(s.buffer) >= s.capacity {
		switch s.dropPolicy {
		case DropOldest:
			// Remove oldest item and add new one
			if len(s.buffer) > 0 {
				s.buffer = s.buffer[1:]
			}
			s.buffer = append(s.buffer, value)
		case DropNewest:
			// Ignore new item
			return false
		case DropAll:
			// Clear buffer and add only the new item
			s.buffer = s.buffer[:0]
			s.buffer = append(s.buffer, value)
		}
	} else {
		// Buffer has space, just append
		s.buffer = append(s.buffer, value)
	}

	// Signal that a value is available
	s.recvCond.Signal()
	return true
}

// ErrStreamDisposed is returned when operations are attempted on a disposed stream.
var ErrStreamDisposed = errors.New("stream has been disposed")

// ErrStreamFull is returned when the stream buffer is full and the drop policy prevents adding new items.
var ErrStreamFull = errors.New("stream buffer is full")
