package action

import (
	"reflect"
	"sync"
	"time"
)

// Clock interface abstracts time operations for testability
type Clock interface {
	Now() time.Time
	Sleep(duration time.Duration)
	After(duration time.Duration) <-chan time.Time
	Timer(duration time.Duration) *FakeTimer
	Ticker(duration time.Duration) *FakeTicker
}

// FakeClock provides deterministic time control for testing
type FakeClock struct {
	mu       sync.RWMutex
	now      time.Time
	timers   []*FakeTimer
	tickers  []*FakeTicker
	sleepers []chan struct{}
}

// NewFakeClock creates a new fake clock starting at the current time
func NewFakeClock() *FakeClock {
	return &FakeClock{
		now:     time.Now(),
		timers:  make([]*FakeTimer, 0),
		tickers: make([]*FakeTicker, 0),
	}
}

// Now returns the current fake time
func (c *FakeClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

// Advance moves the fake clock forward by the given duration
func (c *FakeClock) Advance(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(duration)

	// Trigger timers that should fire
	for _, timer := range c.timers {
		if timer.shouldTrigger(c.now) {
			timer.trigger()
		}
	}

	// Trigger tickers that should tick
	for _, ticker := range c.tickers {
		ticker.checkTicks(c.now)
	}

	// Wake up sleepers
	for _, sleeper := range c.sleepers {
		select {
		case sleeper <- struct{}{}:
		default:
		}
	}
	c.sleepers = c.sleepers[:0] // Clear sleepers
}

// Sleep blocks until the fake clock is advanced by at least the given duration
func (c *FakeClock) Sleep(duration time.Duration) {
	// For testing purposes, just use a small real sleep to allow goroutines to run
	time.Sleep(100 * time.Microsecond)
}

// After returns a channel that receives the current time after the specified duration
func (c *FakeClock) After(duration time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	timer := c.Timer(duration)

	go func() {
		timeValue := <-timer.C
		ch <- timeValue
	}()

	return ch
}

// Timer creates a new fake timer
func (c *FakeClock) Timer(duration time.Duration) *FakeTimer {
	c.mu.Lock()
	defer c.mu.Unlock()

	timer := &FakeTimer{
		C:        make(chan time.Time, 1),
		deadline: c.now.Add(duration),
		clock:    c,
	}

	c.timers = append(c.timers, timer)
	return timer
}

// Ticker creates a new fake ticker
func (c *FakeClock) Ticker(duration time.Duration) *FakeTicker {
	c.mu.Lock()
	defer c.mu.Unlock()

	ticker := &FakeTicker{
		C:        make(chan time.Time, 1),
		interval: duration,
		next:     c.now.Add(duration),
		clock:    c,
		stopped:  false,
	}

	c.tickers = append(c.tickers, ticker)
	return ticker
}

// FakeTimer represents a fake timer for testing
type FakeTimer struct {
	C        chan time.Time
	deadline time.Time
	clock    *FakeClock
	fired    bool
}

// shouldTrigger checks if the timer should fire at the given time
func (t *FakeTimer) shouldTrigger(now time.Time) bool {
	return !t.fired && now.After(t.deadline) || now.Equal(t.deadline)
}

// trigger fires the timer
func (t *FakeTimer) trigger() {
	if !t.fired {
		t.fired = true
		select {
		case t.C <- t.deadline: // Use deadline instead of clock.Now() to avoid mutex deadlock
		default:
		}
	}
}

// Stop prevents the timer from firing
func (t *FakeTimer) Stop() bool {
	if t.fired {
		return false
	}
	t.fired = true
	return true
}

// FakeTicker represents a fake ticker for testing
type FakeTicker struct {
	C        chan time.Time
	interval time.Duration
	next     time.Time
	clock    *FakeClock
	stopped  bool
	mu       sync.Mutex
}

// checkTicks checks if the ticker should tick at the given time
func (t *FakeTicker) checkTicks(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return
	}

	for now.After(t.next) || now.Equal(t.next) {
		select {
		case t.C <- t.next: // Use t.next instead of clock.Now() to avoid mutex deadlock
		default:
		}
		t.next = t.next.Add(t.interval)
	}
}

// Stop stops the ticker
func (t *FakeTicker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopped = true
}

// TestBus provides an isolated testing environment for the action system
type TestBus struct {
	bus   Bus
	clock *FakeClock
}

// NewTestBus creates a new test bus with a fake clock
func NewTestBus() *TestBus {
	clock := NewFakeClock()

	// Create a new bus instance for testing
	bus := New()

	return &TestBus{
		bus:   bus,
		clock: clock,
	}
}

// Bus returns the underlying bus instance
func (tb *TestBus) Bus() Bus {
	return tb.bus
}

// Clock returns the fake clock instance
func (tb *TestBus) Clock() *FakeClock {
	return tb.clock
}

// TestFuture provides testing-friendly Future functionality
type TestFuture[T any] struct {
	*futureImpl[T]
	testBus *TestBus
}

// NewTestFuture creates a new test future
func NewTestFuture[T any](testBus *TestBus) *TestFuture[T] {
	impl := &futureImpl[T]{
		createdAt: testBus.clock.Now(),
	}

	return &TestFuture[T]{
		futureImpl: impl,
		testBus:    testBus,
	}
}

// Resolve resolves the future with a value
func (tf *TestFuture[T]) Resolve(value T) {
	tf.futureImpl.resolve(value)
}

// Reject rejects the future with an error
func (tf *TestFuture[T]) Reject(err error) {
	tf.futureImpl.reject(err)
}

// AwaitWithTimeout waits for the future to resolve or timeout using fake clock
func (tf *TestFuture[T]) AwaitWithTimeout(timeout time.Duration) (T, error) {
	var zero T
	deadline := tf.testBus.clock.Now().Add(timeout)

	for !tf.Done() {
		if tf.testBus.clock.Now().After(deadline) {
			return zero, ErrTimeout
		}
		tf.testBus.clock.Sleep(1 * time.Millisecond)
	}

	return tf.Await()
}

// SubOption for debounce (placeholder implementation)
// Note: This is a simplified implementation for testing purposes
func SubDebounce(duration time.Duration) SubOption {
	return filterOption{
		filter: func(any) bool {
			// Simplified debounce logic for testing
			// In a real implementation, this would integrate with the clock system
			return true
		},
	}
}

// Action assertion helpers

// ActionsEqual compares two actions for equality, ignoring timestamps
func ActionsEqual[T any](a1, a2 Action[T]) bool {
	if a1.Type != a2.Type {
		return false
	}

	if !reflect.DeepEqual(a1.Payload, a2.Payload) {
		return false
	}

	if !reflect.DeepEqual(a1.Meta, a2.Meta) {
		return false
	}

	if a1.Source != a2.Source {
		return false
	}

	if a1.TraceID != a2.TraceID {
		return false
	}

	return true
}

// ActionsInOrder verifies that actions were received in the expected order
func ActionsInOrder[T any](received, expected []Action[T]) bool {
	if len(received) != len(expected) {
		return false
	}

	for i, expectedAction := range expected {
		if !ActionsEqual(received[i], expectedAction) {
			return false
		}
	}

	return true
}

// AssertActionReceived checks if an action was received
func AssertActionReceived[T any](received []Action[T], expected Action[T]) bool {
	for _, action := range received {
		if ActionsEqual(action, expected) {
			return true
		}
	}
	return false
}

// AssertActionCount verifies the number of actions received
func AssertActionCount[T any](received []Action[T], expectedCount int) bool {
	return len(received) == expectedCount
}

// Utility functions for common testing patterns

// WithTestBus runs a test function with a fresh test bus
func WithTestBus(testFunc func(Bus, *FakeClock)) {
	testBus := NewTestBus()
	testFunc(testBus.Bus(), testBus.Clock())
}

// AdvanceClockAndWait advances the clock and allows goroutines to process
func AdvanceClockAndWait(clock *FakeClock, duration time.Duration) {
	clock.Advance(duration)
	time.Sleep(1 * time.Millisecond) // Allow goroutines to process
}

// WaitForCondition waits for a condition to be true with fake clock advancement
func WaitForCondition(clock *FakeClock, condition func() bool, timeout time.Duration) bool {
	deadline := clock.Now().Add(timeout)

	for !condition() {
		if clock.Now().After(deadline) {
			return false
		}
		clock.Advance(1 * time.Millisecond)
		time.Sleep(100 * time.Microsecond) // Small real delay for goroutines
	}

	return true
}

// MockSubscriber creates a mock subscriber for testing
type MockSubscriber[T any] struct {
	mu              sync.Mutex
	receivedActions []Action[T]
	errorToReturn   error
	callCount       int
}

// NewMockSubscriber creates a new mock subscriber
func NewMockSubscriber[T any]() *MockSubscriber[T] {
	return &MockSubscriber[T]{
		receivedActions: make([]Action[T], 0),
	}
}

// Handler returns the handler function for this mock subscriber
func (ms *MockSubscriber[T]) Handler() func(Action[T]) error {
	return func(action Action[T]) error {
		ms.mu.Lock()
		defer ms.mu.Unlock()

		ms.receivedActions = append(ms.receivedActions, action)
		ms.callCount++
		return ms.errorToReturn
	}
}

// GetReceivedActions returns all actions received by this subscriber
func (ms *MockSubscriber[T]) GetReceivedActions() []Action[T] {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result := make([]Action[T], len(ms.receivedActions))
	copy(result, ms.receivedActions)
	return result
}

// GetCallCount returns the number of times the handler was called
func (ms *MockSubscriber[T]) GetCallCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.callCount
}

// SetErrorToReturn sets the error that the handler should return
func (ms *MockSubscriber[T]) SetErrorToReturn(err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.errorToReturn = err
}

// Reset clears all received actions and resets call count
func (ms *MockSubscriber[T]) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.receivedActions = ms.receivedActions[:0]
	ms.callCount = 0
	ms.errorToReturn = nil
}
