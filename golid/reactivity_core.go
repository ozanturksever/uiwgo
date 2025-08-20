// reactivity_core.go
// Core signal primitives for SolidJS-inspired fine-grained reactivity system

package golid

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

// ------------------------------------
// 🔧 Core Reactivity Types
// ------------------------------------

var (
	signalIdCounter      uint64
	computationIdCounter uint64
	ownerIdCounter       uint64
)

// ComputationState represents the state of a computation
type ComputationState int

const (
	Clean ComputationState = iota // Computation is up to date
	Check                         // Computation needs to check dependencies
	Dirty                         // Computation needs to be re-run
)

// Priority defines the priority of scheduled tasks
type Priority int

const (
	UserBlocking Priority = iota // Immediate (user input)
	Normal                       // Default priority
	Idle                         // Low priority
)

// ------------------------------------
// 🎯 Signal Primitive
// ------------------------------------

// ReactiveSignal represents a reactive value with automatic dependency tracking
type ReactiveSignal[T any] struct {
	id          uint64
	value       T
	subscribers map[uint64]*Computation
	owner       *Owner
	compareFn   func(prev, next T) bool
	mutex       sync.RWMutex
}

// SignalOptions provides configuration for signal creation
type SignalOptions[T any] struct {
	Name   string
	Equals func(prev, next T) bool
	Owner  *Owner
}

// safeEqual safely compares two values, handling uncomparable types like slices
func safeEqual[T any](a, b T) bool {
	// Fast path: try direct comparison first
	defer func() {
		if recover() != nil {
			// Comparison panicked, values are uncomparable
			// This means they're definitely different if we got here
		}
	}()

	// For most types, this will work and be very fast
	if any(a) == any(b) {
		return true
	}

	// Only use reflect.DeepEqual for complex types that failed direct comparison
	// This is much faster than always using reflect.DeepEqual
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Quick type check - if types differ, values are different
	if aVal.Type() != bVal.Type() {
		return false
	}

	// For slices, maps, and other complex types, use DeepEqual
	switch aVal.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array, reflect.Struct, reflect.Ptr, reflect.Interface:
		return reflect.DeepEqual(a, b)
	default:
		// For basic types that failed == comparison, they're different
		return false
	}
}

// CreateSignal creates a new reactive signal with automatic dependency tracking
func CreateSignal[T any](initial T, options ...SignalOptions[T]) (func() T, func(T)) {
	s := &ReactiveSignal[T]{
		id:          atomic.AddUint64(&signalIdCounter, 1),
		value:       initial,
		subscribers: make(map[uint64]*Computation, 2), // Pre-allocate small capacity
		owner:       getCurrentOwner(),
	}

	// Apply options
	if len(options) > 0 {
		opt := options[0]
		if opt.Equals != nil {
			s.compareFn = opt.Equals
		}
		if opt.Owner != nil {
			s.owner = opt.Owner
		}
	}

	// Register with owner for cleanup
	if s.owner != nil {
		s.owner.registerSignal(s)
	}

	// Return getter and setter functions
	getter := func() T {
		return s.Get()
	}

	setter := func(value T) {
		s.Set(value)
	}

	return getter, setter
}

// Get retrieves the current value and tracks dependency if in computation context
func (s *ReactiveSignal[T]) Get() T {
	// Track dependency if we're in a computation context
	if comp := getCurrentComputation(); comp != nil {
		s.subscribe(comp)
		comp.addDependency(s)
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.value
}

// Set updates the signal value and schedules dependent computations
func (s *ReactiveSignal[T]) Set(value T) {
	s.mutex.Lock()

	// Check if value actually changed using custom compareFn or default equality
	if s.compareFn != nil {
		if s.compareFn(s.value, value) {
			s.mutex.Unlock()
			return
		}
	} else {
		// Use safe comparison that handles uncomparable types like slices
		if safeEqual(s.value, value) {
			s.mutex.Unlock()
			return
		}
	}

	s.value = value

	// Optimize: Only collect subscribers if there are any
	if len(s.subscribers) == 0 {
		s.mutex.Unlock()
		return
	}

	subscribers := make([]*Computation, 0, len(s.subscribers))
	for _, comp := range s.subscribers {
		subscribers = append(subscribers, comp)
	}
	s.mutex.Unlock()

	// Schedule updates in batch to prevent cascades
	getScheduler().batch(func() {
		for _, comp := range subscribers {
			comp.markDirty()
		}
	})
}

// Update applies a function to the current value
func (s *ReactiveSignal[T]) Update(fn func(T) T) {
	s.mutex.RLock()
	current := s.value
	s.mutex.RUnlock()
	s.Set(fn(current))
}

// subscribe adds a computation as a subscriber
func (s *ReactiveSignal[T]) subscribe(comp *Computation) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.subscribers[comp.id] = comp
}

// unsubscribe removes a computation from subscribers
func (s *ReactiveSignal[T]) unsubscribe(comp *Computation) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.subscribers, comp.id)
}

// ------------------------------------
// 🔄 Computation Primitive (Effect)
// ------------------------------------

// Computation represents a reactive computation (effect)
type Computation struct {
	id           uint64
	fn           func()
	dependencies map[uint64]Dependency
	owner        *Owner
	state        ComputationState
	cleanups     []func()
	context      map[string]interface{}
	mutex        sync.RWMutex
}

// Dependency interface for objects that can be subscribed to
type Dependency interface {
	Subscribe(*Computation)
	Unsubscribe(*Computation)
}

// CreateEffect creates a reactive effect that runs when dependencies change
func CreateEffect(fn func(), owner *Owner) *Computation {
	if owner == nil {
		owner = getCurrentOwner()
	}

	comp := &Computation{
		id:           atomic.AddUint64(&computationIdCounter, 1),
		fn:           fn,
		dependencies: make(map[uint64]Dependency),
		owner:        owner,
		state:        Dirty,
		context:      make(map[string]interface{}),
	}

	if owner != nil {
		owner.registerComputation(comp)
	}

	// Run immediately
	comp.run()

	return comp
}

// CreateMemo creates a memoized computation that returns a value
func CreateMemo[T any](fn func() T, owner *Owner) func() T {
	var memoValue T
	var hasValue bool
	var memoMutex sync.RWMutex

	if owner == nil {
		owner = getCurrentOwner()
	}

	comp := &Computation{
		id:           atomic.AddUint64(&computationIdCounter, 1),
		dependencies: make(map[uint64]Dependency),
		owner:        owner,
		state:        Dirty,
		context:      make(map[string]interface{}),
	}

	// Custom computation function that stores the memo value
	comp.fn = func() {
		newValue := fn()
		memoMutex.Lock()
		memoValue = newValue
		hasValue = true
		memoMutex.Unlock()
	}

	if owner != nil {
		owner.registerComputation(comp)
	}

	// Run immediately to compute initial value
	comp.run()

	return func() T {
		// Track dependency if we're in a computation context
		if currentComp := getCurrentComputation(); currentComp != nil {
			currentComp.addDependency(comp)
			comp.subscribe(currentComp)
		}

		// Ensure computation is up to date
		comp.mutex.RLock()
		state := comp.state
		comp.mutex.RUnlock()

		if state != Clean {
			comp.run()
		}

		memoMutex.RLock()
		defer memoMutex.RUnlock()

		if !hasValue {
			panic("Memo accessed before initialization")
		}
		return memoValue
	}
}

// run executes the computation and tracks dependencies
func (c *Computation) run() {
	c.mutex.Lock()
	if c.state == Clean {
		c.mutex.Unlock()
		return
	}
	c.mutex.Unlock()

	// Cleanup old dependencies
	c.cleanup()

	// Set as current computation for dependency tracking
	prevComputation := getCurrentComputation()
	setCurrentComputation(c)
	defer func() {
		setCurrentComputation(prevComputation)
		c.mutex.Lock()
		c.state = Clean
		c.mutex.Unlock()
	}()

	// Run the computation function
	c.fn()
}

// markDirty marks the computation as needing re-execution
func (c *Computation) markDirty() {
	c.mutex.Lock()
	if c.state != Dirty {
		c.state = Dirty
		c.mutex.Unlock()

		// Schedule for execution
		getScheduler().schedule(&ScheduledTask{
			priority:    Normal,
			computation: c,
		})
	} else {
		c.mutex.Unlock()
	}
}

// addDependency adds a dependency to this computation
func (c *Computation) addDependency(dep Dependency) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Use reflection to get the ID field if it exists
	var id uint64
	if hasIDField(dep) {
		id = getIDFromDependency(dep)
	} else {
		// Fallback to pointer-based ID for unknown types
		id = uint64(uintptr(unsafe.Pointer(&dep)))
	}
	c.dependencies[id] = dep
}

// hasIDField checks if a dependency has an ID field
func hasIDField(dep Dependency) bool {
	val := reflect.ValueOf(dep)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false
	}
	return val.FieldByName("id").IsValid()
}

// getIDFromDependency extracts the ID from a dependency using reflection
func getIDFromDependency(dep Dependency) uint64 {
	val := reflect.ValueOf(dep)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		idField := val.FieldByName("id")
		if idField.IsValid() && idField.Kind() == reflect.Uint64 {
			return idField.Uint()
		}
	}
	return 0
}

// cleanup removes all dependencies and runs cleanup functions
func (c *Computation) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Unsubscribe from all dependencies
	for _, dep := range c.dependencies {
		dep.Unsubscribe(c)
	}
	c.dependencies = make(map[uint64]Dependency)

	// Run cleanup functions
	for _, cleanup := range c.cleanups {
		cleanup()
	}
	c.cleanups = nil
}

// onCleanup adds a cleanup function to be called when computation is disposed
func (c *Computation) onCleanup(fn func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cleanups = append(c.cleanups, fn)
}

// Subscribe implements the Dependency interface for Computation
func (c *Computation) Subscribe(comp *Computation) {
	// For memos, we need to track when other computations depend on us
	// This is a simplified implementation
}

// Unsubscribe implements the Dependency interface for Computation
func (c *Computation) Unsubscribe(comp *Computation) {
	// For memos, we need to untrack when other computations no longer depend on us
	// This is a simplified implementation
}

// subscribe adds a computation as a subscriber (for memo dependencies)
func (c *Computation) subscribe(comp *Computation) {
	// This would track subscribers in a full implementation
	// For now, we'll rely on the scheduler to handle updates
}

// ------------------------------------
// 🏠 Owner Context (Scope Management)
// ------------------------------------

// Owner represents a scope for managing signal and computation lifecycles
type Owner struct {
	id           uint64
	parent       *Owner
	children     []*Owner
	computations []*Computation
	signals      []SignalRef
	cleanups     []func()
	context      map[string]interface{}
	disposed     bool
	mutex        sync.RWMutex
}

// SignalRef holds a weak reference to a signal for cleanup
type SignalRef struct {
	id    uint64
	ptr   unsafe.Pointer
	type_ string
}

// CreateRoot creates a new owner context and runs the function within it
func CreateRoot[T any](fn func() T) (T, func()) {
	owner := &Owner{
		id:           atomic.AddUint64(&ownerIdCounter, 1),
		parent:       getCurrentOwner(),
		children:     make([]*Owner, 0),
		computations: make([]*Computation, 0),
		signals:      make([]SignalRef, 0),
		cleanups:     make([]func(), 0),
		context:      make(map[string]interface{}),
	}

	// Add to parent's children if we have a parent
	if owner.parent != nil {
		owner.parent.addChild(owner)
	}

	// Run function with this owner as current
	prevOwner := getCurrentOwner()
	setCurrentOwner(owner)
	defer setCurrentOwner(prevOwner)

	result := fn()

	// Return result and cleanup function
	cleanup := func() {
		owner.dispose()
	}

	return result, cleanup
}

// CreateOwner creates a new ownership scope for automatic resource management
func CreateOwner(fn func()) *Owner {
	owner := &Owner{
		id:           atomic.AddUint64(&ownerIdCounter, 1),
		parent:       getCurrentOwner(),
		children:     make([]*Owner, 0),
		computations: make([]*Computation, 0),
		signals:      make([]SignalRef, 0),
		cleanups:     make([]func(), 0),
		context:      make(map[string]interface{}),
	}

	// Add to parent's children if we have a parent
	if owner.parent != nil {
		owner.parent.addChild(owner)
	}

	// Run function with this owner as current
	prevOwner := getCurrentOwner()
	setCurrentOwner(owner)
	defer setCurrentOwner(prevOwner)

	fn()

	return owner
}

// RunWithOwner runs a function with the specified owner as current
func RunWithOwner[T any](owner *Owner, fn func() T) T {
	prevOwner := getCurrentOwner()
	setCurrentOwner(owner)
	defer setCurrentOwner(prevOwner)
	return fn()
}

// OnCleanup registers a cleanup function with the current owner
func OnCleanup(fn func()) {
	if owner := getCurrentOwner(); owner != nil {
		owner.addCleanup(fn)
	} else if comp := getCurrentComputation(); comp != nil {
		comp.onCleanup(fn)
	}
}

// registerSignal registers a signal with this owner for cleanup
func (o *Owner) registerSignal(signal interface{}) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.disposed {
		return
	}

	ref := SignalRef{
		ptr:   unsafe.Pointer(&signal),
		type_: "signal",
	}
	o.signals = append(o.signals, ref)
}

// registerComputation registers a computation with this owner
func (o *Owner) registerComputation(comp *Computation) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.disposed {
		return
	}

	o.computations = append(o.computations, comp)
}

// addChild adds a child owner
func (o *Owner) addChild(child *Owner) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.disposed {
		return
	}

	o.children = append(o.children, child)
}

// addCleanup adds a cleanup function
func (o *Owner) addCleanup(fn func()) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.disposed {
		return
	}

	o.cleanups = append(o.cleanups, fn)
}

// Dispose manually disposes the owner and cleans up all resources
func (o *Owner) Dispose() {
	o.dispose()
}

// dispose cleans up all resources owned by this owner
func (o *Owner) dispose() {
	o.mutex.Lock()
	if o.disposed {
		o.mutex.Unlock()
		return
	}
	o.disposed = true

	// Copy references for cleanup
	computations := make([]*Computation, len(o.computations))
	copy(computations, o.computations)

	children := make([]*Owner, len(o.children))
	copy(children, o.children)

	cleanups := make([]func(), len(o.cleanups))
	copy(cleanups, o.cleanups)

	o.mutex.Unlock()

	// Dispose children first
	for _, child := range children {
		child.dispose()
	}

	// Cleanup computations
	for _, comp := range computations {
		comp.cleanup()
	}

	// Run cleanup functions
	for _, cleanup := range cleanups {
		cleanup()
	}

	// Remove from parent
	if o.parent != nil {
		o.parent.removeChild(o)
	}
}

// removeChild removes a child owner
func (o *Owner) removeChild(child *Owner) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	for i, c := range o.children {
		if c == child {
			o.children = append(o.children[:i], o.children[i+1:]...)
			break
		}
	}
}

// Implement Dependency interface for ReactiveSignal
func (s *ReactiveSignal[T]) Subscribe(comp *Computation) {
	s.subscribe(comp)
}

func (s *ReactiveSignal[T]) Unsubscribe(comp *Computation) {
	s.unsubscribe(comp)
}
