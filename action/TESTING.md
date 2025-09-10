
# Action System Testing Utilities

This document describes the testing utilities provided for the UIwGo Action System, designed to make time-based functionality deterministic and provide better testing abstractions.

## Overview

The testing utilities include:

- **FakeClock**: Controllable time abstraction for deterministic timing
- **TestBus**: Isolated testing environment with fake clock integration
- **TestFuture**: Enhanced Future with testing-friendly methods
- **Action Assertions**: Helpers for comparing and verifying actions
- **MockSubscriber**: Mock subscriber for testing action handling

## Core Components

### FakeClock

The `FakeClock` provides deterministic time control for testing time-based operations.

```go
// Create a fake clock
testBus := NewTestBus()
clock := testBus.Clock()

// Check current time
start := clock.Now()

// Advance time manually
clock.Advance(1 * time.Hour)

// Verify time advancement
elapsed := clock.Now().Sub(start)
// elapsed will be exactly 1 hour
```

#### Timer and Ticker Support

```go
// Create a timer
timer := clock.Timer(100 * time.Millisecond)

// Timer won't fire until clock is advanced
select {
case <-timer.C:
    t.Error("Timer shouldn't fire yet")
default:
    // Expected
}

// Advance clock to trigger timer
clock.Advance(100 * time.Millisecond)

select {
case <-timer.C:
    // Timer fires after advancement
default:
    t.Error("Timer should have fired")
}

// Create a ticker
ticker := clock.Ticker(50 * time.Millisecond)
defer ticker.Stop()

// Advance to trigger multiple ticks
clock.Advance(150 * time.Millisecond)

// Ticker will have ticked 3 times
```

### TestBus

The `TestBus` provides an isolated testing environment with integrated fake clock.

```go
// Create isolated test environment
testBus := NewTestBus()
bus := testBus.Bus()
clock := testBus.Clock()

// Set up subscribers
var receivedActions []Action[string]
bus.Subscribe("test.action", func(action Action[string]) error {
    receivedActions = append(receivedActions, action)
    return nil
})

// Dispatch actions
bus.Dispatch(Action[string]{
    Type:    "test.action",
    Payload: "test-data",
})

// Verify actions received
if len(receivedActions) != 1 {
    t.Errorf("Expected 1 action, got %d", len(receivedActions))
}
```

#### Bus Isolation

Each `TestBus` is completely isolated from others:

```go
testBus1 := NewTestBus()
testBus2 := NewTestBus()

// Actions dispatched to bus1 won't affect bus2
// Clock advancement in clock1 won't affect clock2
```

### TestFuture

Enhanced `Future` with testing-friendly methods:

```go
testBus := NewTestBus()
future := NewTestFuture[string](testBus)

// Set up async resolution
go func() {
    time.Sleep(10 * time.Millisecond)
    future.Resolve("success")
}()

// Test Then callback
var result string
future.Then(func(value string) {
    result = value
})

// Await result
value, err := future.Await()
if err != nil {
    t.Errorf("Unexpected error: %v", err)
}

// Test rejection
rejectFuture := NewTestFuture[string](testBus)
testErr := errors.New("test error")
rejectFuture.Reject(testErr)

_, err = rejectFuture.Await()
if err != testErr {
    t.Errorf("Expected test error, got %v", err)
}
```

#### AwaitWithTimeout
The `TestFuture` includes an `AwaitWithTimeout` method that waits for a `Future` to resolve or reject, using the `FakeClock` for timeout control. This is useful for testing asynchronous operations with deadlines without relying on real-time delays.

```go
testBus := NewTestBus()
future := NewTestFuture[string](testBus)

// Resolve the future after a delay
go func() {
    time.Sleep(50 * time.Millisecond) 
    future.Resolve("done")
}()

// This will fail because the clock is not advanced
_, err := future.AwaitWithTimeout(10 * time.Millisecond)
if err != ErrTimeout {
    t.Errorf("Expected timeout error, got %v", err)
}

// This will succeed
go func() {
    time.Sleep(50 * time.Millisecond)
    future.Resolve("done")
}()
_, err = future.AwaitWithTimeout(100 * time.Millisecond)
if err != nil {
    t.Errorf("Expected success, got %v", err)
}
```

### SubDebounce Option

The `SubDebounce` function is a placeholder for testing and does not implement actual debouncing logic. It is included to allow tests to cover code paths that use debounce options without requiring a full debounce implementation in the test environment.

```go
// This is a placeholder and does not perform debouncing
bus.Subscribe("test.action", mock.Handler(), SubDebounce(100*time.Millisecond))
```

## Action Assertions

### Action Equality

The `ActionsEqual` function compares two actions for equality. It checks the `Type`, `Payload`, `Meta`, `Source`, and `TraceID` fields, but **ignores timestamps**. This is useful for verifying actions without being dependent on exact timing.

```go
action1 := Action[string]{
    Type:    "test.action",
    Payload: "data",
    Meta:    map[string]any{"key": "value"},
}

action2 := Action[string]{
    Type:    "test.action",
    Payload: "data",
    Meta:    map[string]any{"key": "value"},
}

// Compare actions (ignores timestamps)
// Dispatch actions
bus.Dispatch(Action[string]{Type: "test.action", Payload: "test1"})
bus.Dispatch(Action[string]{Type: "test.action", Payload: "test2"})

// Verify received actions
received := mock.GetReceivedActions()
if len(received) != 2 {
    t.Errorf("Expected 2 actions, got %d", len(received))
}

// Check call count
if mock.GetCallCount() != 2 {
    t.Errorf("Expected 2 calls, got %d", mock.GetCallCount())
}

// Configure error behavior
mock.SetErrorToReturn(errors.New("test error"))

// Reset for new test
mock.Reset()
```

## Utility Functions

### WithTestBus

Run tests with a fresh test bus:

```go
WithTestBus(func(bus Bus, clock *FakeClock) {
    // Test code here with isolated bus and clock
    bus.Subscribe("test", func(Action[string]) error { return nil })
    clock.Advance(1 * time.Hour)
})
```

### AdvanceClockAndWait

Advance clock and allow goroutines to process:

```go
AdvanceClockAndWait(clock, 100*time.Millisecond)
```

### WaitForCondition

Wait for a condition with fake clock:

```go
success := WaitForCondition(clock, func() bool {
    return len(receivedActions) > 0
}, 1*time.Second)

if !success {
    t.Error("Condition not met within timeout")
}
```

## Best Practices

### 1. Use TestBus for Isolation

Always use `TestBus` instead of the global bus in tests:

```go
// Good
func TestMyFeature(t *testing.T) {
    testBus := NewTestBus()
    bus := testBus.Bus()
    // ... test code
}

// Avoid
func TestMyFeature(t *testing.T) {
    bus := New() // Global bus, not isolated
    // ... test code
}
```

### 2. Control Time Deterministically

Use `FakeClock` instead of real time delays:

```go
// Good
clock.Advance(100 * time.Millisecond)

// Avoid
time.Sleep(100 * time.Millisecond) // Non-deterministic
```

### 3. Test Both Success and Error Cases

```go
func TestAsyncOperation(t *testing.T) {
    testBus := NewTestBus()
    
    // Test success case
    future := NewTestFuture[string](testBus)
    go func() {
        future.Resolve("success")
    }()
    
    result, err := future.Await()
    if err != nil {
        t.Errorf("Expected success, got error: %v", err)
    }
    
    // Test error case
    errorFuture := NewTestFuture[string](testBus)
    testErr := errors.New("test error")
    errorFuture.Reject(testErr)
    
    _, err = errorFuture.Await()
    if err != testErr {
        t.Errorf("Expected test error, got: %v", err)
    }
}
```

### 4. Use Assertions for Clarity

```go
// Good - clear intent
if !ActionsEqual(received, expected) {
    t.Error("Actions should be equal")
}

// Less clear
if received.Type != expected.Type || received.Payload != expected.Payload {
    t.Error("Actions don't match")
}
```

### 5. Clean Up Resources

```go
func TestWithTicker(t *testing.T) {
    testBus := NewTestBus()
    clock := testBus.Clock()
    
    ticker := clock.Ticker(50 * time.Millisecond)
    defer ticker.Stop() // Always clean up
    
    // ... test code
}
```

## Integration with Existing System

The testing utilities work seamlessly with the existing action system:

```go
func TestRealWorldScenario(t *testing.T) {
    testBus := NewTestBus()
    bus := testBus.Bus()
    clock := testBus.Clock()
    
    // Set up realistic scenario
    var events []string
    
    // Subscribe to multiple action types
    bus.Subscribe("user.login", func(action Action[string]) error {
        events = append(events, "login:"+action.Payload)
        return nil
    })
    
    bus.Subscribe("user.logout", func(action Action[string]) error {
        events = append(events, "logout:"+action.Payload)
        return nil
    })
    
    // Simulate user flow
    bus.Dispatch(Action[string]{
        Type:    "user.login",
        Payload: "user123",
        Meta:    map[string]any{"timestamp": clock.Now()},
    })
    
    // Advance time
    clock.Advance(30 * time.Minute)
    
    bus.Dispatch(Action[string]{
        Type:    "user.logout", 
        Payload: "user123",
        Meta:    map[string]any{"timestamp": clock.Now()},
    })
    
    // Verify expected flow
    expectedEvents := []string{"login:user123", "logout:user123"}
    if !reflect.DeepEqual(events, expectedEvents) {
        t.Errorf("Expected %v, got %v", expectedEvents, events)
    }
}
```

## Performance Testing

The utilities also support performance testing scenarios:

```go
func BenchmarkActionProcessing(b *testing.B) {
    testBus := NewTestBus()
    bus := testBus.Bus()
    
    // Set up subscriber
    bus.Subscribe("bench.action", func(Action[string]) error {
        return nil
    })
    
    action := Action[string]{
        Type:    "bench.action",
        Payload: "test",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        bus.Dispatch(action)
    }
}
```

This comprehensive testing framework ensures that your action system tests are deterministic, isolated, and reliable.