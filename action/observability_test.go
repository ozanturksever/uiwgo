package action

import (
	"sync"
	"testing"
	"time"
)

// TestOnError_InvokedOnPanicWithContext tests that the enhanced OnError hook
// is called with the proper context when a panic occurs during dispatch.
func TestOnError_InvokedOnPanicWithContext(t *testing.T) {
	bus := New()

	var capturedCtx Context
	var capturedErr error
	var capturedRecovered any
	var errorHandled bool

	// Set up enhanced error handler
	bus.OnError(func(ctx Context, err error, recovered any) {
		capturedCtx = ctx
		capturedErr = err
		capturedRecovered = recovered
		errorHandled = true
	})

	// Subscribe with a handler that panics
	bus.Subscribe("test.panic", func(action Action[string]) error {
		panic("test panic message")
	})

	// Dispatch action with specific context metadata
	err := bus.Dispatch(Action[string]{
		Type:    "test.panic",
		Payload: "test",
		Meta:    map[string]any{"test": "value"},
		TraceID: "trace-123",
		Source:  "test-source",
		Time:    time.Now(),
	})

	// Should not return error (panic is recovered)
	if err != nil {
		t.Errorf("Expected dispatch to succeed despite panic, got error: %v", err)
	}

	// Allow time for async error handling
	time.Sleep(10 * time.Millisecond)

	// Verify error handler was called
	if !errorHandled {
		t.Fatal("Expected error handler to be called when panic occurs")
	}

	// Verify context is populated
	if capturedCtx.TraceID != "trace-123" {
		t.Errorf("Expected TraceID 'trace-123', got '%s'", capturedCtx.TraceID)
	}
	if capturedCtx.Source != "test-source" {
		t.Errorf("Expected Source 'test-source', got '%s'", capturedCtx.Source)
	}
	if capturedCtx.Meta["test"] != "value" {
		t.Errorf("Expected Meta['test'] = 'value', got %v", capturedCtx.Meta["test"])
	}

	// Verify error details
	if capturedErr == nil {
		t.Fatal("Expected error to be captured")
	}

	// Verify recovered panic value
	if capturedRecovered != "test panic message" {
		t.Errorf("Expected recovered value 'test panic message', got %v", capturedRecovered)
	}
}

// TestDevLogger_DoesNotChangeDelivery tests that the dev logger middleware
// logs actions without affecting delivery behavior or order.
func TestDevLogger_DoesNotChangeDelivery(t *testing.T) {
	bus := New()

	// Enable dev logger
	logEntries := make([]DevLogEntry, 0)
	var logMutex sync.Mutex

	EnableDevLogger(bus, func(entry DevLogEntry) {
		logMutex.Lock()
		defer logMutex.Unlock()
		logEntries = append(logEntries, entry)
	})

	// Track delivery order and count
	var deliveryOrder []string
	var deliveryMutex sync.Mutex

	// Subscribe multiple handlers
	bus.Subscribe("test.order", func(action Action[string]) error {
		deliveryMutex.Lock()
		defer deliveryMutex.Unlock()
		deliveryOrder = append(deliveryOrder, "handler1")
		return nil
	})

	bus.Subscribe("test.order", func(action Action[string]) error {
		deliveryMutex.Lock()
		defer deliveryMutex.Unlock()
		deliveryOrder = append(deliveryOrder, "handler2")
		return nil
	})

	bus.SubscribeAny(func(action any) error {
		deliveryMutex.Lock()
		defer deliveryMutex.Unlock()
		deliveryOrder = append(deliveryOrder, "any-handler")
		return nil
	})

	// Dispatch multiple actions
	for i := 0; i < 3; i++ {
		err := bus.Dispatch(Action[string]{
			Type:    "test.order",
			Payload: "test",
			TraceID: "trace-" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("Dispatch %d failed: %v", i, err)
		}
	}

	// Allow async processing to complete
	time.Sleep(10 * time.Millisecond)

	// Verify delivery order is consistent (should be deterministic)
	expectedOrder := []string{
		"handler1", "handler2", "any-handler", // First dispatch
		"handler1", "handler2", "any-handler", // Second dispatch
		"handler1", "handler2", "any-handler", // Third dispatch
	}

	deliveryMutex.Lock()
	actualOrder := make([]string, len(deliveryOrder))
	copy(actualOrder, deliveryOrder)
	deliveryMutex.Unlock()

	if len(actualOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d deliveries, got %d", len(expectedOrder), len(actualOrder))
	}

	for i, expected := range expectedOrder {
		if actualOrder[i] != expected {
			t.Errorf("Delivery order mismatch at index %d: expected '%s', got '%s'", i, expected, actualOrder[i])
		}
	}

	// Verify dev logger captured entries
	logMutex.Lock()
	logCount := len(logEntries)
	logMutex.Unlock()

	if logCount != 3 {
		t.Errorf("Expected 3 log entries, got %d", logCount)
	}

	// Verify log entries contain expected data
	logMutex.Lock()
	for i, entry := range logEntries {
		if entry.ActionType != "test.order" {
			t.Errorf("Log entry %d: expected ActionType 'test.order', got '%s'", i, entry.ActionType)
		}
		if entry.SubscriberCount != 3 { // 2 specific + 1 any
			t.Errorf("Log entry %d: expected SubscriberCount 3, got %d", i, entry.SubscriberCount)
		}
		if entry.Duration <= 0 {
			t.Errorf("Log entry %d: expected positive Duration, got %v", i, entry.Duration)
		}
		expectedTraceID := "trace-" + string(rune('0'+i))
		if entry.TraceID != expectedTraceID {
			t.Errorf("Log entry %d: expected TraceID '%s', got '%s'", i, expectedTraceID, entry.TraceID)
		}
	}
	logMutex.Unlock()
}

// TestDebugRingBuffer_SizeBounded tests that the debug ring buffer
// maintains a maximum size and overwrites oldest entries.
func TestDebugRingBuffer_SizeBounded(t *testing.T) {
	bus := New()

	// Configure small ring buffer
	bufferSize := 3
	EnableDebugRingBuffer(bus, bufferSize)

	// Dispatch more actions than buffer size
	for i := 0; i < 5; i++ {
		err := bus.Dispatch(Action[string]{
			Type:    "test.buffer",
			Payload: "payload-" + string(rune('0'+i)),
			TraceID: "trace-" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("Dispatch %d failed: %v", i, err)
		}
	}

	// Allow async processing
	time.Sleep(10 * time.Millisecond)

	// Get buffer entries
	entries := GetDebugRingBufferEntries(bus, "test.buffer")

	// Verify buffer size is bounded
	if len(entries) != bufferSize {
		t.Errorf("Expected buffer size %d, got %d entries", bufferSize, len(entries))
	}

	// Verify entries are the most recent ones (last 3 dispatches)
	expectedPayloads := []string{"payload-2", "payload-3", "payload-4"}
	expectedTraceIDs := []string{"trace-2", "trace-3", "trace-4"}

	for i, entry := range entries {
		if entry.Payload != expectedPayloads[i] {
			t.Errorf("Entry %d: expected payload '%s', got '%s'", i, expectedPayloads[i], entry.Payload)
		}
		if entry.TraceID != expectedTraceIDs[i] {
			t.Errorf("Entry %d: expected TraceID '%s', got '%s'", i, expectedTraceIDs[i], entry.TraceID)
		}
	}

	// Test with different action type
	for i := 0; i < 2; i++ {
		err := bus.Dispatch(Action[string]{
			Type:    "test.other",
			Payload: "other-" + string(rune('0'+i)),
			TraceID: "other-trace-" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("Dispatch other %d failed: %v", i, err)
		}
	}

	// Allow async processing
	time.Sleep(10 * time.Millisecond)

	// Verify separate buffers per action type
	originalEntries := GetDebugRingBufferEntries(bus, "test.buffer")
	otherEntries := GetDebugRingBufferEntries(bus, "test.other")

	// Original buffer should be unchanged
	if len(originalEntries) != bufferSize {
		t.Errorf("Original buffer should still have %d entries, got %d", bufferSize, len(originalEntries))
	}

	// Other buffer should have 2 entries
	if len(otherEntries) != 2 {
		t.Errorf("Other buffer should have 2 entries, got %d", len(otherEntries))
	}

	// Test buffer clearing
	ClearDebugRingBuffer(bus, "test.buffer")
	clearedEntries := GetDebugRingBufferEntries(bus, "test.buffer")
	if len(clearedEntries) != 0 {
		t.Errorf("Expected buffer to be cleared, got %d entries", len(clearedEntries))
	}

	// Other buffer should be unaffected
	otherEntriesAfterClear := GetDebugRingBufferEntries(bus, "test.other")
	if len(otherEntriesAfterClear) != 2 {
		t.Errorf("Other buffer should still have 2 entries after clearing different buffer, got %d", len(otherEntriesAfterClear))
	}
}

// TestAnalyticsTap tests the analytics tap helper for observing actions
func TestAnalyticsTap(t *testing.T) {
	bus := New()

	// Track all analytics events
	var analyticsEvents []AnalyticsEvent
	var analyticsMutex sync.Mutex

	// Set up analytics tap with filter
	tap := NewAnalyticsTap(bus, func(event AnalyticsEvent) {
		analyticsMutex.Lock()
		defer analyticsMutex.Unlock()
		analyticsEvents = append(analyticsEvents, event)
	}, WithAnalyticsFilter(func(action any) bool {
		// Only track actions with "analytics" in payload
		if act, ok := action.(Action[string]); ok {
			return act.Payload == "analytics"
		}
		return false
	}))
	defer tap.Dispose()

	// Dispatch various actions
	actions := []Action[string]{
		{Type: "test.event1", Payload: "analytics", TraceID: "trace-1"},
		{Type: "test.event2", Payload: "not-tracked", TraceID: "trace-2"},
		{Type: "test.event3", Payload: "analytics", TraceID: "trace-3"},
	}

	for _, action := range actions {
		err := bus.Dispatch(action)
		if err != nil {
			t.Fatalf("Failed to dispatch action: %v", err)
		}
	}

	// Allow async processing
	time.Sleep(10 * time.Millisecond)

	// Verify only filtered actions were captured
	analyticsMutex.Lock()
	eventCount := len(analyticsEvents)
	analyticsMutex.Unlock()

	if eventCount != 2 {
		t.Errorf("Expected 2 analytics events (filtered), got %d", eventCount)
	}

	// Verify event details
	analyticsMutex.Lock()
	for i, event := range analyticsEvents {
		expectedTraceIDs := []string{"trace-1", "trace-3"}
		if event.TraceID != expectedTraceIDs[i] {
			t.Errorf("Event %d: expected TraceID '%s', got '%s'", i, expectedTraceIDs[i], event.TraceID)
		}
		if event.ActionType != "test.event1" && event.ActionType != "test.event3" {
			t.Errorf("Event %d: unexpected ActionType '%s'", i, event.ActionType)
		}
	}
	analyticsMutex.Unlock()
}
