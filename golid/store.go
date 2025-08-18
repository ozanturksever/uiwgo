// store.go
// Reactive store system for centralized state management with separation of concerns

package golid

import (
	"sync"
	"sync/atomic"
)

// ------------------------------------
// 🏪 Core Store Types
// ------------------------------------

var storeIdCounter uint64

// Store represents a reactive store with centralized state management
type Store[T any] struct {
	id          uint64
	signal      *ReactiveSignal[T]
	getter      func() T
	setter      func(T)
	subscribers map[uint64]*StoreSubscription
	middleware  []StoreMiddleware[T]
	owner       *Owner
	name        string
	mutex       sync.RWMutex
}

// StoreSubscription represents a subscription to store changes
type StoreSubscription struct {
	id       uint64
	callback func(interface{})
	cleanup  func()
	owner    *Owner
}

// StoreOptions provides configuration for store creation
type StoreOptions[T any] struct {
	Name       string
	Owner      *Owner
	Middleware []StoreMiddleware[T]
	Equals     func(prev, next T) bool
}

// StoreMiddleware defines middleware for store operations
type StoreMiddleware[T any] interface {
	BeforeUpdate(store *Store[T], oldValue, newValue T) T
	AfterUpdate(store *Store[T], oldValue, newValue T)
}

// ------------------------------------
// 🏪 Store Creation
// ------------------------------------

// CreateStore creates a new reactive store with centralized state management
func CreateStore[T any](initialState T, options ...StoreOptions[T]) *Store[T] {
	var opts StoreOptions[T]
	if len(options) > 0 {
		opts = options[0]
	}

	// Create underlying reactive signal
	signalOpts := SignalOptions[T]{
		Name:   opts.Name,
		Owner:  opts.Owner,
		Equals: opts.Equals,
	}
	getter, setter := CreateSignal(initialState, signalOpts)

	store := &Store[T]{
		id:          atomic.AddUint64(&storeIdCounter, 1),
		getter:      getter,
		setter:      setter,
		subscribers: make(map[uint64]*StoreSubscription),
		middleware:  opts.Middleware,
		owner:       opts.Owner,
		name:        opts.Name,
	}

	// Get the underlying signal for direct access
	if opts.Owner != nil {
		opts.Owner.registerStore(store)
	}

	return store
}

// ------------------------------------
// 🔄 Store Operations
// ------------------------------------

// Get retrieves the current store value
func (s *Store[T]) Get() T {
	return s.getter()
}

// Set updates the store value with middleware processing
func (s *Store[T]) Set(value T) {
	oldValue := s.getter()

	// Apply before middleware
	processedValue := value
	for _, middleware := range s.middleware {
		processedValue = middleware.BeforeUpdate(s, oldValue, processedValue)
	}

	// Update the underlying signal
	s.setter(processedValue)

	// Apply after middleware
	for _, middleware := range s.middleware {
		middleware.AfterUpdate(s, oldValue, processedValue)
	}

	// Notify subscribers
	s.notifySubscribers(processedValue)
}

// Update applies a function to the current store value
func (s *Store[T]) Update(fn func(T) T) {
	current := s.Get()
	s.Set(fn(current))
}

// Subscribe adds a subscription to store changes
func (s *Store[T]) Subscribe(callback func(T)) func() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subscriptionId := atomic.AddUint64(&storeIdCounter, 1)

	// Wrap callback to handle type conversion
	wrappedCallback := func(value interface{}) {
		if typedValue, ok := value.(T); ok {
			callback(typedValue)
		}
	}

	subscription := &StoreSubscription{
		id:       subscriptionId,
		callback: wrappedCallback,
		owner:    getCurrentOwner(),
	}

	// Create cleanup function
	cleanup := func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		delete(s.subscribers, subscriptionId)
	}
	subscription.cleanup = cleanup

	s.subscribers[subscriptionId] = subscription

	// Register cleanup with owner if available
	if subscription.owner != nil {
		subscription.owner.addCleanup(cleanup)
	}

	return cleanup
}

// notifySubscribers notifies all subscribers of store changes
func (s *Store[T]) notifySubscribers(value T) {
	s.mutex.RLock()
	subscribers := make([]*StoreSubscription, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		subscribers = append(subscribers, sub)
	}
	s.mutex.RUnlock()

	// Notify subscribers outside of lock to prevent deadlocks
	for _, sub := range subscribers {
		sub.callback(value)
	}
}

// ------------------------------------
// 🔗 Derived Stores
// ------------------------------------

// CreateDerivedStore creates a store derived from other stores
func CreateDerivedStore[T any](deriveFn func() T, options ...StoreOptions[T]) *Store[T] {
	var opts StoreOptions[T]
	if len(options) > 0 {
		opts = options[0]
	}

	// Create memo for derived computation
	memo := CreateMemo(deriveFn, opts.Owner)

	// Create store with memo as getter
	store := &Store[T]{
		id:          atomic.AddUint64(&storeIdCounter, 1),
		getter:      memo,
		setter:      func(T) { panic("Cannot set value on derived store") },
		subscribers: make(map[uint64]*StoreSubscription),
		middleware:  opts.Middleware,
		owner:       opts.Owner,
		name:        opts.Name,
	}

	// Track changes to derived value
	CreateEffect(func() {
		value := memo()
		store.notifySubscribers(value)
	}, opts.Owner)

	if opts.Owner != nil {
		opts.Owner.registerStore(store)
	}

	return store
}

// ------------------------------------
// 🔄 Store Composition
// ------------------------------------

// CombineStores combines multiple stores into a single store
func CombineStores[T any](combiner func() T, options ...StoreOptions[T]) *Store[T] {
	return CreateDerivedStore(combiner, options...)
}

// MapStore creates a new store by mapping values from an existing store
func MapStore[T, U any](store *Store[T], mapper func(T) U, options ...StoreOptions[U]) *Store[U] {
	return CreateDerivedStore(func() U {
		return mapper(store.Get())
	}, options...)
}

// FilterStore creates a new store that only updates when the filter condition is met
func FilterStore[T any](store *Store[T], filter func(T) bool, options ...StoreOptions[T]) *Store[T] {
	var lastValue T
	var hasValue bool

	return CreateDerivedStore(func() T {
		current := store.Get()
		if filter(current) {
			lastValue = current
			hasValue = true
		}
		if !hasValue {
			panic("FilterStore accessed before any value passed filter")
		}
		return lastValue
	}, options...)
}

// ------------------------------------
// 🔧 Store Utilities
// ------------------------------------

// GetStoreInfo returns information about the store
func (s *Store[T]) GetStoreInfo() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"id":              s.id,
		"name":            s.name,
		"subscriberCount": len(s.subscribers),
		"middlewareCount": len(s.middleware),
		"hasOwner":        s.owner != nil,
	}
}

// Clone creates a copy of the store with the same value
func (s *Store[T]) Clone(options ...StoreOptions[T]) *Store[T] {
	return CreateStore(s.Get(), options...)
}

// ------------------------------------
// 🧹 Store Cleanup
// ------------------------------------

// Dispose cleans up the store and all its subscriptions
func (s *Store[T]) Dispose() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Cleanup all subscriptions
	for _, sub := range s.subscribers {
		if sub.cleanup != nil {
			sub.cleanup()
		}
	}
	s.subscribers = make(map[uint64]*StoreSubscription)
}

// ------------------------------------
// 🏠 Owner Integration
// ------------------------------------

// registerStore registers a store with an owner for cleanup
func (o *Owner) registerStore(store interface{}) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.disposed {
		return
	}

	// Add cleanup for store disposal
	o.addCleanup(func() {
		if disposable, ok := store.(interface{ Dispose() }); ok {
			disposable.Dispose()
		}
	})
}

// ------------------------------------
// 🧪 Store Testing Utilities
// ------------------------------------

// CreateTestStore creates a store for testing purposes
func CreateTestStore[T any](initialState T) (*Store[T], func()) {
	root, cleanup := CreateRoot(func() *Store[T] {
		return CreateStore(initialState, StoreOptions[T]{
			Name: "test-store",
		})
	})
	return root, cleanup
}

// GetStoreValue is a testing utility to get store value without tracking
func GetStoreValue[T any](store *Store[T]) T {
	return Track(func() T {
		return store.Get()
	})
}

// ------------------------------------
// 📊 Store Statistics
// ------------------------------------

// StoreStats provides statistics about store usage
type StoreStats struct {
	TotalStores        int
	ActiveStores       int
	TotalSubscriptions int
}

var globalStoreRegistry = struct {
	stores map[uint64]*Store[interface{}]
	mutex  sync.RWMutex
}{
	stores: make(map[uint64]*Store[interface{}]),
}

// GetGlobalStoreStats returns global store statistics
func GetGlobalStoreStats() StoreStats {
	globalStoreRegistry.mutex.RLock()
	defer globalStoreRegistry.mutex.RUnlock()

	totalSubscriptions := 0
	activeStores := 0

	for _, store := range globalStoreRegistry.stores {
		if store != nil {
			activeStores++
			info := store.GetStoreInfo()
			if count, ok := info["subscriberCount"].(int); ok {
				totalSubscriptions += count
			}
		}
	}

	return StoreStats{
		TotalStores:        len(globalStoreRegistry.stores),
		ActiveStores:       activeStores,
		TotalSubscriptions: totalSubscriptions,
	}
}
