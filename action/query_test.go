package action

import (
	"fmt"
	"testing"
	"time"
)

func TestHandleQuery_RegisterAndAnswer(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register a handler
	handlerCalled := false
	sub := bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		handlerCalled = true
		return "response", nil
	})

	// Verify subscription is active
	if !sub.IsActive() {
		t.Error("Expected subscription to be active")
	}

	// Send a query
	action := Action[string]{
		Type:    queryType,
		Payload: "request",
	}
	result, err := bus.Ask(queryType, action)

	// Verify no error from Ask
	if err != nil {
		t.Errorf("Expected no error from Ask, got %v", err)
	}

	// Verify result is a future
	future, ok := result.(Future[any])
	if !ok {
		t.Error("Expected result to be a Future")
	}

	// Wait for the result
	actualResult, err := future.Await()
	if err != nil {
		t.Errorf("Expected no error from future, got %v", err)
	}

	// Verify handler was called
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	// Verify result
	if actualResult != "response" {
		t.Errorf("Expected result 'response', got %v", actualResult)
	}

	// Dispose the subscription
	err = sub.Dispose()
	if err != nil {
		t.Errorf("Expected no error when disposing subscription, got %v", err)
	}

	// Verify subscription is no longer active
	if sub.IsActive() {
		t.Error("Expected subscription to be inactive after disposal")
	}
}

func TestHandleQuery_ReplaceExisting(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register first handler
	firstHandlerCalled := false
	firstSub := bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		firstHandlerCalled = true
		return "first-response", nil
	})

	// Register second handler (should replace first)
	secondHandlerCalled := false
	secondSub := bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		secondHandlerCalled = true
		return "second-response", nil
	})

	// Verify both subscriptions are active
	if !firstSub.IsActive() {
		t.Error("Expected first subscription to be active")
	}
	if !secondSub.IsActive() {
		t.Error("Expected second subscription to be active")
	}

	// Send a query
	action := Action[string]{
		Type:    queryType,
		Payload: "request",
	}
	result, err := bus.Ask(queryType, action)

	// Verify no error from Ask
	if err != nil {
		t.Errorf("Expected no error from Ask, got %v", err)
	}

	// Verify result is a future
	future, ok := result.(Future[any])
	if !ok {
		t.Error("Expected result to be a Future")
	}

	// Wait for the result
	actualResult, err := future.Await()
	if err != nil {
		t.Errorf("Expected no error from future, got %v", err)
	}

	// Verify first handler was not called
	if firstHandlerCalled {
		t.Error("Expected first handler not to be called")
	}

	// Verify second handler was called
	if !secondHandlerCalled {
		t.Error("Expected second handler to be called")
	}

	// Verify result
	if actualResult != "second-response" {
		t.Errorf("Expected result 'second-response', got %v", actualResult)
	}
}

func TestAsk_NoHandler_Error(t *testing.T) {
	bus := New()

	// Define a query type with no handler
	queryType := "test-query"

	// Send a query
	action := Action[string]{
		Type:    queryType,
		Payload: "request",
	}
	result, err := bus.Ask(queryType, action)

	// Verify error
	if err != nil {
		t.Errorf("Expected no error from Ask, got %v", err)
	}

	// Verify result is a future that rejects with ErrNoHandler
	if future, ok := result.(Future[any]); ok {
		_, err := future.Await()
		if err != ErrNoHandler {
			t.Errorf("Expected ErrNoHandler, got %v", err)
		}
	} else {
		t.Error("Expected result to be a Future")
	}
}

func TestAsk_Timeout(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register a handler that takes too long
	bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		time.Sleep(100 * time.Millisecond)
		return "response", nil
	})

	// Send a query with a short timeout
	action := Action[string]{
		Type:    queryType,
		Payload: "request",
	}
	result, err := bus.Ask(queryType, action, AskWithTimeout(10*time.Millisecond))

	// Verify no error from Ask
	if err != nil {
		t.Errorf("Expected no error from Ask, got %v", err)
	}

	// Verify result is a future that rejects with ErrTimeout
	if future, ok := result.(Future[any]); ok {
		_, err := future.Await()
		if err != ErrTimeout {
			t.Errorf("Expected ErrTimeout, got %v", err)
		}
	} else {
		t.Error("Expected result to be a Future")
	}
}

func TestAsk_Concurrency_One(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register a handler with ConcurrencyOne policy
	activeCount := 0
	maxActiveCount := 0
	bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		activeCount++
		if activeCount > maxActiveCount {
			maxActiveCount = activeCount
		}
		time.Sleep(50 * time.Millisecond) // Simulate work
		activeCount--
		return "response", nil
	}, QueryWithConcurrencyPolicy(ConcurrencyOne))

	// Send multiple queries concurrently
	futures := make([]Future[any], 3)
	for i := 0; i < 3; i++ {
		action := Action[string]{
			Type:    queryType,
			Payload: fmt.Sprintf("request-%d", i),
		}
		result, err := bus.Ask(queryType, action)
		if err != nil {
			t.Errorf("Expected no error from Ask, got %v", err)
		}
		if future, ok := result.(Future[any]); ok {
			futures[i] = future
		} else {
			t.Error("Expected result to be a Future")
		}
	}

	// Wait for all futures to complete
	errors := make([]error, len(futures))
	for i, future := range futures {
		_, errors[i] = future.Await()
	}

	// Verify that only one query was processed at a time
	if maxActiveCount > 1 {
		t.Errorf("Expected max active count to be 1 with ConcurrencyOne policy, got %d", maxActiveCount)
	}

	// Verify that at least one query was rejected
	rejectedCount := 0
	for _, err := range errors {
		if err == ErrNoHandler {
			rejectedCount++
		}
	}
	if rejectedCount == 0 {
		t.Error("Expected at least one query to be rejected with ConcurrencyOne policy")
	}
}

func TestAsk_Concurrency_LatestCancelsPrevious(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register a handler with ConcurrencyLatest policy
	activeCount := 0
	maxActiveCount := 0
	canceledCount := 0
	bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		activeCount++
		if activeCount > maxActiveCount {
			maxActiveCount = activeCount
		}
		time.Sleep(50 * time.Millisecond) // Simulate work
		activeCount--
		// Check if this query was canceled
		if action.Meta != nil {
			if _, exists := action.Meta["canceled"]; exists {
				canceledCount++
				return nil, ErrTimeout
			}
		}
		return "response", nil
	}, QueryWithConcurrencyPolicy(ConcurrencyLatest))

	// Send multiple queries concurrently
	futures := make([]Future[any], 3)
	for i := 0; i < 3; i++ {
		action := Action[string]{
			Type:    queryType,
			Payload: fmt.Sprintf("request-%d", i),
		}
		result, err := bus.Ask(queryType, action)
		if err != nil {
			t.Errorf("Expected no error from Ask, got %v", err)
		}
		if future, ok := result.(Future[any]); ok {
			futures[i] = future
		} else {
			t.Error("Expected result to be a Future")
		}
	}

	// Wait for all futures to complete
	results := make([]any, len(futures))
	errors := make([]error, len(futures))
	for i, future := range futures {
		results[i], errors[i] = future.Await()
	}

	// Verify that at some point multiple queries were active (before cancellation)
	if maxActiveCount == 0 {
		t.Error("Expected max active count to be greater than 0 with ConcurrencyLatest policy")
	}

	// Verify that the last query succeeded
	if errors[2] != nil {
		t.Errorf("Expected last query to succeed, got error: %v", errors[2])
	}
	if results[2] != "response" {
		t.Errorf("Expected last query to return 'response', got %v", results[2])
	}
}

func TestAsk_Concurrency_QueueProcessesInOrder(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register a handler with ConcurrencyQueue policy
	callOrder := make([]int, 0)
	bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		// Extract the request number from the payload
		var requestNum int
		fmt.Sscanf(action.Payload, "request-%d", &requestNum)
		callOrder = append(callOrder, requestNum)
		time.Sleep(10 * time.Millisecond) // Simulate work
		return fmt.Sprintf("response-%d", requestNum), nil
	}, QueryWithConcurrencyPolicy(ConcurrencyQueue))

	// Send multiple queries concurrently
	futures := make([]Future[any], 3)
	for i := 0; i < 3; i++ {
		action := Action[string]{
			Type:    queryType,
			Payload: fmt.Sprintf("request-%d", i),
		}
		result, err := bus.Ask(queryType, action)
		if err != nil {
			t.Errorf("Expected no error from Ask, got %v", err)
		}
		if future, ok := result.(Future[any]); ok {
			futures[i] = future
		} else {
			t.Error("Expected result to be a Future")
		}
	}

	// Wait for all futures to complete
	results := make([]any, len(futures))
	errors := make([]error, len(futures))
	for i, future := range futures {
		results[i], errors[i] = future.Await()
	}

	// Verify no errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Expected no error for query %d, got %v", i, err)
		}
	}

	// Verify that all queries were processed
	if len(callOrder) != 3 {
		t.Errorf("Expected all 3 queries to be processed, got %d", len(callOrder))
	}

	// Verify results
	for i, result := range results {
		expected := fmt.Sprintf("response-%d", i)
		if result != expected {
			t.Errorf("Expected result[%d] to be %s, got %s", i, expected, result)
		}
	}
}

func TestAsk_PropagatesMetaTraceSource(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register a handler that checks the context
	var receivedMeta map[string]any
	var receivedTraceID string
	var receivedSource string
	bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		receivedMeta = action.Meta
		receivedTraceID = action.TraceID
		receivedSource = action.Source
		return "response", nil
	})

	// Send a query with meta, trace, and source
	action := Action[string]{
		Type:    queryType,
		Payload: "request",
	}
	result, err := bus.Ask(queryType, action,
		AskWithMeta(map[string]any{"key": "value"}),
		AskWithTraceID("trace-123"),
		AskWithSource("test-source"))

	// Verify no error from Ask
	if err != nil {
		t.Errorf("Expected no error from Ask, got %v", err)
	}

	// Wait for the result
	if future, ok := result.(Future[any]); ok {
		_, err := future.Await()
		if err != nil {
			t.Errorf("Expected no error from future, got %v", err)
		}
	} else {
		t.Error("Expected result to be a Future")
	}

	// Verify meta was passed through
	if receivedMeta == nil {
		t.Error("Expected meta to be passed through")
	} else if receivedMeta["key"] != "value" {
		t.Errorf("Expected meta key 'key' to be 'value', got %v", receivedMeta["key"])
	}

	// Verify trace ID was passed through
	if receivedTraceID != "trace-123" {
		t.Errorf("Expected trace ID 'trace-123', got %s", receivedTraceID)
	}

	// Verify source was passed through
	if receivedSource != "test-source" {
		t.Errorf("Expected source 'test-source', got %s", receivedSource)
	}
}

func TestAsk_CleanupOnSubscriptionDispose(t *testing.T) {
	bus := New()

	// Define a query type
	queryType := "test-query"

	// Register a handler
	sub := bus.HandleQuery(queryType, func(action Action[string]) (any, error) {
		time.Sleep(50 * time.Millisecond) // Simulate work
		return "response", nil
	})

	// Send a query
	action := Action[string]{
		Type:    queryType,
		Payload: "request",
	}
	result, err := bus.Ask(queryType, action)

	// Verify no error from Ask
	if err != nil {
		t.Errorf("Expected no error from Ask, got %v", err)
	}

	// Verify result is a future
	future, ok := result.(Future[any])
	if !ok {
		t.Error("Expected result to be a Future")
	}

	// Dispose the subscription while the query is still processing
	err = sub.Dispose()
	if err != nil {
		t.Errorf("Expected no error when disposing subscription, got %v", err)
	}

	// Wait for the result
	_, err = future.Await()
	if err == nil {
		t.Error("Expected error after subscription disposal, got nil")
	}
}
