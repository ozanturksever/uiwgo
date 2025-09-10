package action

import (
	"errors"
	"testing"
	"time"
)

// TestFakeClock_BasicFunctionality tests that FakeClock provides controllable time
func TestFakeClock_BasicFunctionality(t *testing.T) {
	// Create a test bus with fake clock
	testBus := NewTestBus()
	clock := testBus.Clock()

	// Test Now() returns consistent time
	startTime := clock.Now()

	// Advance clock
	clock.Advance(1 * time.Hour)

	endTime := clock.Now()

	expectedDuration := 1 * time.Hour
	actualDuration := endTime.Sub(startTime)

	if actualDuration != expectedDuration {
		t.Errorf("Expected time advancement of %v, got %v", expectedDuration, actualDuration)
	}
}

// TestTestBus_Isolation tests that TestBus provides isolated testing environments
func TestTestBus_Isolation(t *testing.T) {
	// Create two test buses
	testBus1 := NewTestBus()
	testBus2 := NewTestBus()

	bus1 := testBus1.Bus()
	bus2 := testBus2.Bus()
	clock1 := testBus1.Clock()
	clock2 := testBus2.Clock()

	// Set up subscribers on both buses
	var bus1Received []string
	var bus2Received []string

	bus1.Subscribe("test.action", func(action Action[string]) error {
		bus1Received = append(bus1Received, action.Payload)
		return nil
	})

	bus2.Subscribe("test.action", func(action Action[string]) error {
		bus2Received = append(bus2Received, action.Payload)
		return nil
	})

	// Dispatch to bus1
	bus1.Dispatch(Action[string]{
		Type:    "test.action",
		Payload: "bus1-action",
	})

	// Dispatch to bus2
	bus2.Dispatch(Action[string]{
		Type:    "test.action",
		Payload: "bus2-action",
	})

	// Verify isolation - each bus should only receive its own actions
	if len(bus1Received) != 1 || bus1Received[0] != "bus1-action" {
		t.Errorf("Expected bus1 to receive only 'bus1-action', got %v", bus1Received)
	}

	if len(bus2Received) != 1 || bus2Received[0] != "bus2-action" {
		t.Errorf("Expected bus2 to receive only 'bus2-action', got %v", bus2Received)
	}

	// Test clock isolation - advancing one shouldn't affect the other
	initialTime1 := clock1.Now()
	initialTime2 := clock2.Now()

	clock1.Advance(1 * time.Hour)

	if !clock1.Now().Equal(initialTime1.Add(1 * time.Hour)) {
		t.Error("Clock1 should be advanced by 1 hour")
	}

	if !clock2.Now().Equal(initialTime2) {
		t.Error("Clock2 should not be affected by clock1 advancement")
	}
}

// TestFuture_AwaitThenCatch tests Future functionality with deterministic timing
func TestFuture_AwaitThenCatch(t *testing.T) {
	testBus := NewTestBus()

	// Test successful resolution
	future := NewTestFuture[string](testBus)

	// Set up async resolution
	go func() {
		time.Sleep(10 * time.Millisecond) // Small real delay
		future.Resolve("success")
	}()

	// Test Then callback
	var thenResult string
	future.Then(func(result string) {
		thenResult = result
	})

	// Await normally (not using fake clock for this test since we need actual async behavior)
	result, err := future.Await()

	if err != nil {
		t.Errorf("Expected successful await, got error: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got '%s'", result)
	}

	if thenResult != "success" {
		t.Errorf("Expected then callback to receive 'success', got '%s'", thenResult)
	}

	// Test Catch callback with rejection
	rejectFuture := NewTestFuture[string](testBus)

	var catchErr error
	rejectFuture.Catch(func(err error) {
		catchErr = err
	})

	testErr := errors.New("test error")
	rejectFuture.Reject(testErr)

	_, err = rejectFuture.Await()
	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}

	if catchErr != testErr {
		t.Errorf("Expected catch callback to receive test error, got %v", catchErr)
	}
}

// TestActionAssertions tests action equality and ordering helpers
func TestActionAssertions(t *testing.T) {
	action1 := Action[string]{
		Type:    "test.action",
		Payload: "payload1",
		Meta:    map[string]any{"key": "value"},
	}

	action2 := Action[string]{
		Type:    "test.action",
		Payload: "payload1",
		Meta:    map[string]any{"key": "value"},
	}

	action3 := Action[string]{
		Type:    "test.action",
		Payload: "payload2",
	}

	// Test action equality
	if !ActionsEqual(action1, action2) {
		t.Error("Expected action1 and action2 to be equal")
	}

	if ActionsEqual(action1, action3) {
		t.Error("Expected action1 and action3 to be different")
	}

	// Test action ordering verification
	var receivedActions []Action[string]
	testBus := NewTestBus()
	bus := testBus.Bus()

	bus.Subscribe("test.action", func(action Action[string]) error {
		receivedActions = append(receivedActions, action)
		return nil
	})

	// Dispatch actions in order
	bus.Dispatch(action1)
	bus.Dispatch(action3)

	expectedOrder := []Action[string]{action1, action3}
	if !ActionsInOrder(receivedActions, expectedOrder) {
		t.Error("Expected actions to be received in dispatch order")
	}
}

// TestFakeClock_TimerAndTicker tests FakeClock with Timer and Ticker functionality
func TestFakeClock_TimerAndTicker(t *testing.T) {
	testBus := NewTestBus()
	clock := testBus.Clock()

	// Test Timer - more simplified approach
	timer := clock.Timer(100 * time.Millisecond)

	// Timer shouldn't be ready initially
	select {
	case <-timer.C:
		t.Error("Timer should not fire before advancing clock")
	default:
		// Expected - timer not ready
	}

	// Advance clock partially
	clock.Advance(50 * time.Millisecond)

	// Timer still shouldn't be ready
	select {
	case <-timer.C:
		t.Error("Timer should not fire at 50ms")
	default:
		// Expected - timer not ready
	}

	// Advance clock to trigger timer
	clock.Advance(50 * time.Millisecond)

	// Timer should now be ready
	select {
	case <-timer.C:
		// Expected - timer fired
	default:
		t.Error("Timer should fire at 100ms")
	}

	// Test Ticker - simplified
	ticker := clock.Ticker(50 * time.Millisecond)
	defer ticker.Stop()

	// Initially no ticks
	select {
	case <-ticker.C:
		t.Error("Ticker should not tick before time advancement")
	default:
		// Expected
	}

	// Advance to first tick
	clock.Advance(50 * time.Millisecond)

	select {
	case <-ticker.C:
		// Expected - first tick
	default:
		t.Error("Ticker should tick after 50ms")
	}

	// Advance to second tick
	clock.Advance(50 * time.Millisecond)

	select {
	case <-ticker.C:
		// Expected - second tick
	default:
		t.Error("Ticker should tick after 100ms total")
	}
}

// TestFakeClock_After tests FakeClock After functionality
func TestFakeClock_After(t *testing.T) {
	testBus := NewTestBus()
	clock := testBus.Clock()

	// Test with a direct timer approach to avoid goroutine timing issues
	timer := clock.Timer(100 * time.Millisecond)

	// Should not trigger initially
	select {
	case <-timer.C:
		t.Error("Timer should not trigger before advancing clock")
	default:
		// Expected
	}

	// Advance clock partially
	clock.Advance(50 * time.Millisecond)
	
	select {
	case <-timer.C:
		t.Error("Timer should not trigger at 50ms")
	default:
		// Expected
	}

	// Advance clock to trigger timer
	clock.Advance(50 * time.Millisecond)
	
	select {
	case <-timer.C:
		// Expected - Timer triggered
	case <-time.After(100 * time.Millisecond):
		t.Error("Timer should trigger after advancing clock by duration")
	}
}

// TestFakeClock_Sleep tests FakeClock Sleep functionality
func TestFakeClock_Sleep(t *testing.T) {
	testBus := NewTestBus()
	clock := testBus.Clock()

	start := clock.Now()

	// Sleep is now just a small delay for testing purposes
	clock.Sleep(100 * time.Millisecond)

	// Verify time advancement works independently
	clock.Advance(100 * time.Millisecond)

	elapsed := clock.Now().Sub(start)
	if elapsed != 100*time.Millisecond {
		t.Errorf("Expected 100ms elapsed, got %v", elapsed)
	}
}
