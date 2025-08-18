// actions.go
// Pure action functions for business logic operations with separation of concerns

package golid

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 🎯 Core Action Types
// ------------------------------------

var actionIdCounter uint64

// StoreAction represents a pure business logic operation
type StoreAction[T, R any] struct {
	id         uint64
	name       string
	fn         func(T) R
	middleware []ActionMiddleware[T, R]
	owner      *Owner
	metadata   map[string]interface{}
	mutex      sync.RWMutex
}

// AsyncStoreAction represents an asynchronous business logic operation
type AsyncStoreAction[T, R any] struct {
	id         uint64
	name       string
	fn         func(context.Context, T) (R, error)
	middleware []AsyncActionMiddleware[T, R]
	owner      *Owner
	metadata   map[string]interface{}
	mutex      sync.RWMutex
}

// ActionResult represents the result of an action execution
type ActionResult[R any] struct {
	Value     R
	Error     error
	Duration  time.Duration
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// ActionOptions provides configuration for action creation
type ActionOptions[T, R any] struct {
	Name       string
	Owner      *Owner
	Middleware []ActionMiddleware[T, R]
	Metadata   map[string]interface{}
}

// AsyncActionOptions provides configuration for async action creation
type AsyncActionOptions[T, R any] struct {
	Name       string
	Owner      *Owner
	Middleware []AsyncActionMiddleware[T, R]
	Metadata   map[string]interface{}
	Timeout    time.Duration
}

// ------------------------------------
// 🔧 Action Middleware
// ------------------------------------

// ActionMiddleware defines middleware for synchronous actions
type ActionMiddleware[T, R any] interface {
	BeforeAction(action *StoreAction[T, R], payload T) T
	AfterAction(action *StoreAction[T, R], payload T, result R) R
	OnError(action *StoreAction[T, R], payload T, err error) error
}

// AsyncActionMiddleware defines middleware for asynchronous actions
type AsyncActionMiddleware[T, R any] interface {
	BeforeAction(action *AsyncStoreAction[T, R], payload T) T
	AfterAction(action *AsyncStoreAction[T, R], payload T, result R) R
	OnError(action *AsyncStoreAction[T, R], payload T, err error) error
}

// ------------------------------------
// 🎯 Action Creation
// ------------------------------------

// CreateAction creates a new synchronous action
func CreateAction[T, R any](fn func(T) R, options ...ActionOptions[T, R]) *StoreAction[T, R] {
	var opts ActionOptions[T, R]
	if len(options) > 0 {
		opts = options[0]
	}

	action := &StoreAction[T, R]{
		id:         atomic.AddUint64(&actionIdCounter, 1),
		name:       opts.Name,
		fn:         fn,
		middleware: opts.Middleware,
		owner:      opts.Owner,
		metadata:   opts.Metadata,
	}

	if action.metadata == nil {
		action.metadata = make(map[string]interface{})
	}

	if opts.Owner != nil {
		opts.Owner.registerAction(action)
	}

	return action
}

// CreateAsyncAction creates a new asynchronous action
func CreateAsyncAction[T, R any](fn func(context.Context, T) (R, error), options ...AsyncActionOptions[T, R]) *AsyncStoreAction[T, R] {
	var opts AsyncActionOptions[T, R]
	if len(options) > 0 {
		opts = options[0]
	}

	action := &AsyncStoreAction[T, R]{
		id:         atomic.AddUint64(&actionIdCounter, 1),
		name:       opts.Name,
		fn:         fn,
		middleware: opts.Middleware,
		owner:      opts.Owner,
		metadata:   opts.Metadata,
	}

	if action.metadata == nil {
		action.metadata = make(map[string]interface{})
	}

	if opts.Owner != nil {
		opts.Owner.registerAction(action)
	}

	return action
}

// ------------------------------------
// 🚀 Action Execution
// ------------------------------------

// Execute runs the action with the given payload
func (a *StoreAction[T, R]) Execute(payload T) ActionResult[R] {
	start := time.Now()
	result := ActionResult[R]{
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	defer func() {
		result.Duration = time.Since(start)
		if r := recover(); r != nil {
			result.Error = fmt.Errorf("action panic: %v", r)
		}
	}()

	// Apply before middleware
	processedPayload := payload
	for _, middleware := range a.middleware {
		processedPayload = middleware.BeforeAction(a, processedPayload)
	}

	// Execute the action
	value := a.fn(processedPayload)
	result.Value = value

	// Apply after middleware
	processedResult := value
	for _, middleware := range a.middleware {
		processedResult = middleware.AfterAction(a, processedPayload, processedResult)
	}
	result.Value = processedResult

	return result
}

// ExecuteAsync runs the async action with the given payload
func (a *AsyncStoreAction[T, R]) ExecuteAsync(ctx context.Context, payload T) <-chan ActionResult[R] {
	resultChan := make(chan ActionResult[R], 1)

	go func() {
		defer close(resultChan)

		start := time.Now()
		result := ActionResult[R]{
			Timestamp: start,
			Metadata:  make(map[string]interface{}),
		}

		defer func() {
			result.Duration = time.Since(start)
			if r := recover(); r != nil {
				result.Error = fmt.Errorf("async action panic: %v", r)
			}
			resultChan <- result
		}()

		// Apply before middleware
		processedPayload := payload
		for _, middleware := range a.middleware {
			processedPayload = middleware.BeforeAction(a, processedPayload)
		}

		// Execute the async action
		value, err := a.fn(ctx, processedPayload)
		if err != nil {
			// Apply error middleware
			processedError := err
			for _, middleware := range a.middleware {
				processedError = middleware.OnError(a, processedPayload, processedError)
			}
			result.Error = processedError
			return
		}

		result.Value = value

		// Apply after middleware
		processedResult := value
		for _, middleware := range a.middleware {
			processedResult = middleware.AfterAction(a, processedPayload, processedResult)
		}
		result.Value = processedResult
	}()

	return resultChan
}

// Execute runs the async action synchronously (blocks until completion)
func (a *AsyncStoreAction[T, R]) Execute(ctx context.Context, payload T) ActionResult[R] {
	resultChan := a.ExecuteAsync(ctx, payload)
	return <-resultChan
}

// ------------------------------------
// 🔗 Action Composition
// ------------------------------------

// ChainActions creates a new action that chains multiple actions together
func ChainActions[T, U, R any](first *StoreAction[T, U], second *StoreAction[U, R], options ...ActionOptions[T, R]) *StoreAction[T, R] {
	return CreateAction(func(payload T) R {
		intermediate := first.Execute(payload)
		if intermediate.Error != nil {
			panic(intermediate.Error)
		}
		final := second.Execute(intermediate.Value)
		if final.Error != nil {
			panic(final.Error)
		}
		return final.Value
	}, options...)
}

// ComposeActions creates a new action by composing multiple actions
func ComposeActions[T, R any](actions []*StoreAction[T, T], finalAction *StoreAction[T, R], options ...ActionOptions[T, R]) *StoreAction[T, R] {
	return CreateAction(func(payload T) R {
		current := payload
		for _, action := range actions {
			result := action.Execute(current)
			if result.Error != nil {
				panic(result.Error)
			}
			current = result.Value
		}
		final := finalAction.Execute(current)
		if final.Error != nil {
			panic(final.Error)
		}
		return final.Value
	}, options...)
}

// ------------------------------------
// 📦 Action Dispatching
// ------------------------------------

// ActionDispatcher manages action execution and provides centralized dispatching
type ActionDispatcher struct {
	middleware []DispatcherMiddleware
	actions    map[string]interface{}
	mutex      sync.RWMutex
}

// DispatcherMiddleware defines middleware for the action dispatcher
type DispatcherMiddleware interface {
	BeforeDispatch(actionName string, payload interface{}) interface{}
	AfterDispatch(actionName string, payload interface{}, result interface{})
	OnError(actionName string, payload interface{}, err error) error
}

// CreateDispatcher creates a new action dispatcher
func CreateDispatcher(middleware ...DispatcherMiddleware) *ActionDispatcher {
	return &ActionDispatcher{
		middleware: middleware,
		actions:    make(map[string]interface{}),
	}
}

// RegisterAction registers an action with the dispatcher
func (d *ActionDispatcher) RegisterAction(name string, action interface{}) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.actions[name] = action
}

// RegisterAsyncAction registers an async action with the dispatcher
func (d *ActionDispatcher) RegisterAsyncAction(name string, action interface{}) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.actions[name] = action
}

// Dispatch executes an action by name
func (d *ActionDispatcher) Dispatch(actionName string, payload interface{}) (interface{}, error) {
	d.mutex.RLock()
	actionInterface, exists := d.actions[actionName]
	d.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("action '%s' not found", actionName)
	}

	// Apply before middleware
	processedPayload := payload
	for _, middleware := range d.middleware {
		processedPayload = middleware.BeforeDispatch(actionName, processedPayload)
	}

	// Execute action using reflection
	result, err := d.executeAction(actionInterface, processedPayload)
	if err != nil {
		// Apply error middleware
		processedError := err
		for _, middleware := range d.middleware {
			processedError = middleware.OnError(actionName, processedPayload, processedError)
		}
		return nil, processedError
	}

	// Apply after middleware
	for _, middleware := range d.middleware {
		middleware.AfterDispatch(actionName, processedPayload, result)
	}

	return result, nil
}

// executeAction executes an action using reflection
func (d *ActionDispatcher) executeAction(actionInterface interface{}, payload interface{}) (interface{}, error) {
	actionValue := reflect.ValueOf(actionInterface)
	actionType := actionValue.Type()

	// Find Execute method
	executeMethod := actionValue.MethodByName("Execute")
	if !executeMethod.IsValid() {
		return nil, fmt.Errorf("action does not have Execute method")
	}

	// Prepare arguments
	args := []reflect.Value{reflect.ValueOf(payload)}

	// Handle async actions (need context)
	if actionType.String() == "*golid.AsyncAction" {
		args = []reflect.Value{reflect.ValueOf(context.Background()), reflect.ValueOf(payload)}
	}

	// Call Execute method
	results := executeMethod.Call(args)
	if len(results) == 0 {
		return nil, fmt.Errorf("action Execute method returned no results")
	}

	// Extract result
	result := results[0].Interface()

	// Handle ActionResult type using reflection
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Struct {
		// Check if it has Value and Error fields
		valueField := resultValue.FieldByName("Value")
		errorField := resultValue.FieldByName("Error")

		if valueField.IsValid() && errorField.IsValid() {
			// Check if Error field is not nil
			if !errorField.IsNil() {
				if err, ok := errorField.Interface().(error); ok {
					return nil, err
				}
			}
			return valueField.Interface(), nil
		}
	}

	return result, nil
}

// ------------------------------------
// 🧹 Action Cleanup
// ------------------------------------

// registerAction registers an action with an owner for cleanup
func (o *Owner) registerAction(action interface{}) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.disposed {
		return
	}

	// Actions don't need explicit cleanup, but we track them for debugging
	o.addCleanup(func() {
		// Action cleanup logic if needed
	})
}

// ------------------------------------
// 📊 Action Utilities
// ------------------------------------

// GetActionInfo returns information about the action
func (a *StoreAction[T, R]) GetActionInfo() map[string]interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return map[string]interface{}{
		"id":              a.id,
		"name":            a.name,
		"middlewareCount": len(a.middleware),
		"hasOwner":        a.owner != nil,
		"metadata":        a.metadata,
	}
}

// GetAsyncActionInfo returns information about the async action
func (a *AsyncStoreAction[T, R]) GetAsyncActionInfo() map[string]interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return map[string]interface{}{
		"id":              a.id,
		"name":            a.name,
		"middlewareCount": len(a.middleware),
		"hasOwner":        a.owner != nil,
		"metadata":        a.metadata,
	}
}

// ------------------------------------
// 🧪 Action Testing Utilities
// ------------------------------------

// CreateTestAction creates an action for testing purposes
func CreateTestAction[T, R any](fn func(T) R) (*StoreAction[T, R], func()) {
	root, cleanup := CreateRoot(func() *StoreAction[T, R] {
		return CreateAction(fn, ActionOptions[T, R]{
			Name: "test-action",
		})
	})
	return root, cleanup
}

// CreateTestAsyncAction creates an async action for testing purposes
func CreateTestAsyncAction[T, R any](fn func(context.Context, T) (R, error)) (*AsyncStoreAction[T, R], func()) {
	root, cleanup := CreateRoot(func() *AsyncStoreAction[T, R] {
		return CreateAsyncAction(fn, AsyncActionOptions[T, R]{
			Name: "test-async-action",
		})
	})
	return root, cleanup
}
