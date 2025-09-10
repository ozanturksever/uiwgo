package action

import (
	"sync"
	"time"
)

// Future represents a promise-like abstraction for asynchronous operations.
// It supports Then, Catch, Await, and Done methods for handling results.
type Future[T any] interface {
	// Then registers a callback to be called when the future resolves successfully.
	// Returns a new Future for chaining.
	Then(func(T)) Future[T]

	// Catch registers a callback to be called when the future rejects with an error.
	// Returns a new Future for chaining.
	Catch(func(error)) Future[T]

	// Await blocks until the future resolves or rejects, returning the result and error.
	Await() (T, error)

	// Done returns true if the future has resolved or rejected.
	Done() bool
}

// futureImpl is the internal implementation of Future[T].
type futureImpl[T any] struct {
	mu        sync.Mutex
	result    T
	err       error
	done      bool
	thenCb    func(T)
	catchCb   func(error)
	next      *futureImpl[T]
	createdAt time.Time
}

// Then registers a callback to be called when the future resolves successfully.
func (f *futureImpl[T]) Then(cb func(T)) Future[T] {
	f.mu.Lock()
	defer f.mu.Unlock()

	next := &futureImpl[T]{}
	f.next = next

	if f.done && f.err == nil {
		// Already resolved, call callback immediately
		cb(f.result)
	} else if !f.done {
		// Not yet resolved, store callback
		f.thenCb = cb
	} else {
		// Already rejected, don't call callback
	}

	return next
}

// Catch registers a callback to be called when the future rejects with an error.
func (f *futureImpl[T]) Catch(cb func(error)) Future[T] {
	f.mu.Lock()
	defer f.mu.Unlock()

	next := &futureImpl[T]{}
	f.next = next

	if f.done && f.err != nil {
		// Already rejected, call callback immediately
		cb(f.err)
	} else if !f.done {
		// Not yet resolved, store callback
		f.catchCb = cb
	} else {
		// Already resolved, don't call callback
	}

	return next
}

// Await blocks until the future resolves or rejects, returning the result and error.
func (f *futureImpl[T]) Await() (T, error) {
	// Simple implementation that blocks until done
	// In a more sophisticated implementation, this could use channels
	for !f.Done() {
		time.Sleep(1 * time.Millisecond)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	return f.result, f.err
}

// Done returns true if the future has resolved or rejected.
func (f *futureImpl[T]) Done() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.done
}

// resolve marks the future as resolved with a result.
func (f *futureImpl[T]) resolve(result T) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.done {
		return // Already resolved or rejected
	}

	f.result = result
	f.done = true

	// Call then callback if registered
	if f.thenCb != nil {
		f.thenCb(result)
	}

	// Propagate to next future if it exists
	if f.next != nil {
		f.next.resolve(result)
	}
}

// reject marks the future as rejected with an error.
func (f *futureImpl[T]) reject(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.done {
		return // Already resolved or rejected
	}

	f.err = err
	f.done = true

	// Call catch callback if registered
	if f.catchCb != nil {
		f.catchCb(err)
	}

	// Propagate to next future if it exists
	if f.next != nil {
		f.next.reject(err)
	}
}

// NewFuture creates a new Future[T].
func NewFuture[T any]() Future[T] {
	return &futureImpl[T]{
		createdAt: time.Now(),
	}
}
