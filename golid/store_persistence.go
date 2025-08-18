//go:build js && wasm
// +build js,wasm

// store_persistence.go
// Store persistence and hydration utilities for state management

package golid

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"syscall/js"
	"time"
)

// ------------------------------------
// 🗄️ Persistence Types
// ------------------------------------

// PersistenceAdapter defines the interface for persistence backends
type PersistenceAdapter interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
	Delete(key string) error
	Exists(key string) bool
	Clear() error
}

// PersistenceOptions configures store persistence
type PersistenceOptions struct {
	Key         string
	Adapter     PersistenceAdapter
	Serializer  Serializer
	Throttle    time.Duration
	AutoSave    bool
	Versioning  bool
	Compression bool
}

// Serializer defines the interface for data serialization
type Serializer interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte, target interface{}) error
}

// PersistentStore wraps a store with persistence capabilities
type PersistentStore[T any] struct {
	store       *Store[T]
	options     PersistenceOptions
	lastSave    time.Time
	saveTimer   *time.Timer
	mutex       sync.RWMutex
	initialized bool
}

// ------------------------------------
// 🌐 Browser Storage Adapters
// ------------------------------------

// LocalStorageAdapter implements persistence using browser localStorage
type LocalStorageAdapter struct {
	prefix string
}

// NewLocalStorageAdapter creates a new localStorage adapter
func NewLocalStorageAdapter(prefix string) *LocalStorageAdapter {
	return &LocalStorageAdapter{prefix: prefix}
}

// Save stores data in localStorage
func (a *LocalStorageAdapter) Save(key string, data []byte) error {
	if !js.Global().Get("localStorage").Truthy() {
		return fmt.Errorf("localStorage not available")
	}

	fullKey := a.prefix + key
	js.Global().Get("localStorage").Call("setItem", fullKey, string(data))
	return nil
}

// Load retrieves data from localStorage
func (a *LocalStorageAdapter) Load(key string) ([]byte, error) {
	if !js.Global().Get("localStorage").Truthy() {
		return nil, fmt.Errorf("localStorage not available")
	}

	fullKey := a.prefix + key
	value := js.Global().Get("localStorage").Call("getItem", fullKey)

	if value.IsNull() {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return []byte(value.String()), nil
}

// Delete removes data from localStorage
func (a *LocalStorageAdapter) Delete(key string) error {
	if !js.Global().Get("localStorage").Truthy() {
		return fmt.Errorf("localStorage not available")
	}

	fullKey := a.prefix + key
	js.Global().Get("localStorage").Call("removeItem", fullKey)
	return nil
}

// Exists checks if key exists in localStorage
func (a *LocalStorageAdapter) Exists(key string) bool {
	if !js.Global().Get("localStorage").Truthy() {
		return false
	}

	fullKey := a.prefix + key
	value := js.Global().Get("localStorage").Call("getItem", fullKey)
	return !value.IsNull()
}

// Clear removes all items with the prefix
func (a *LocalStorageAdapter) Clear() error {
	if !js.Global().Get("localStorage").Truthy() {
		return fmt.Errorf("localStorage not available")
	}

	localStorage := js.Global().Get("localStorage")
	length := localStorage.Get("length").Int()

	keysToRemove := make([]string, 0)
	for i := 0; i < length; i++ {
		key := localStorage.Call("key", i).String()
		if len(key) >= len(a.prefix) && key[:len(a.prefix)] == a.prefix {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		localStorage.Call("removeItem", key)
	}

	return nil
}

// ------------------------------------
// 📱 SessionStorage Adapter
// ------------------------------------

// SessionStorageAdapter implements persistence using browser sessionStorage
type SessionStorageAdapter struct {
	prefix string
}

// NewSessionStorageAdapter creates a new sessionStorage adapter
func NewSessionStorageAdapter(prefix string) *SessionStorageAdapter {
	return &SessionStorageAdapter{prefix: prefix}
}

// Save stores data in sessionStorage
func (a *SessionStorageAdapter) Save(key string, data []byte) error {
	if !js.Global().Get("sessionStorage").Truthy() {
		return fmt.Errorf("sessionStorage not available")
	}

	fullKey := a.prefix + key
	js.Global().Get("sessionStorage").Call("setItem", fullKey, string(data))
	return nil
}

// Load retrieves data from sessionStorage
func (a *SessionStorageAdapter) Load(key string) ([]byte, error) {
	if !js.Global().Get("sessionStorage").Truthy() {
		return nil, fmt.Errorf("sessionStorage not available")
	}

	fullKey := a.prefix + key
	value := js.Global().Get("sessionStorage").Call("getItem", fullKey)

	if value.IsNull() {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return []byte(value.String()), nil
}

// Delete removes data from sessionStorage
func (a *SessionStorageAdapter) Delete(key string) error {
	if !js.Global().Get("sessionStorage").Truthy() {
		return fmt.Errorf("sessionStorage not available")
	}

	fullKey := a.prefix + key
	js.Global().Get("sessionStorage").Call("removeItem", fullKey)
	return nil
}

// Exists checks if key exists in sessionStorage
func (a *SessionStorageAdapter) Exists(key string) bool {
	if !js.Global().Get("sessionStorage").Truthy() {
		return false
	}

	fullKey := a.prefix + key
	value := js.Global().Get("sessionStorage").Call("getItem", fullKey)
	return !value.IsNull()
}

// Clear removes all items with the prefix
func (a *SessionStorageAdapter) Clear() error {
	if !js.Global().Get("sessionStorage").Truthy() {
		return fmt.Errorf("sessionStorage not available")
	}

	sessionStorage := js.Global().Get("sessionStorage")
	length := sessionStorage.Get("length").Int()

	keysToRemove := make([]string, 0)
	for i := 0; i < length; i++ {
		key := sessionStorage.Call("key", i).String()
		if len(key) >= len(a.prefix) && key[:len(a.prefix)] == a.prefix {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		sessionStorage.Call("removeItem", key)
	}

	return nil
}

// ------------------------------------
// 💾 Memory Adapter (for testing)
// ------------------------------------

// MemoryAdapter implements in-memory persistence for testing
type MemoryAdapter struct {
	data  map[string][]byte
	mutex sync.RWMutex
}

// NewMemoryAdapter creates a new memory adapter
func NewMemoryAdapter() *MemoryAdapter {
	return &MemoryAdapter{
		data: make(map[string][]byte),
	}
}

// Save stores data in memory
func (a *MemoryAdapter) Save(key string, data []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Make a copy to avoid data races
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	a.data[key] = dataCopy
	return nil
}

// Load retrieves data from memory
func (a *MemoryAdapter) Load(key string) ([]byte, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	data, exists := a.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Return a copy to avoid data races
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return dataCopy, nil
}

// Delete removes data from memory
func (a *MemoryAdapter) Delete(key string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	delete(a.data, key)
	return nil
}

// Exists checks if key exists in memory
func (a *MemoryAdapter) Exists(key string) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	_, exists := a.data[key]
	return exists
}

// Clear removes all data from memory
func (a *MemoryAdapter) Clear() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.data = make(map[string][]byte)
	return nil
}

// ------------------------------------
// 🔄 JSON Serializer
// ------------------------------------

// JSONSerializer implements JSON serialization
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Serialize converts data to JSON bytes
func (s *JSONSerializer) Serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// Deserialize converts JSON bytes to data
func (s *JSONSerializer) Deserialize(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

// ------------------------------------
// 🏪 Persistent Store Implementation
// ------------------------------------

// PersistStore creates a persistent store wrapper
func PersistStore[T any](store *Store[T], options PersistenceOptions) (*PersistentStore[T], error) {
	// Set defaults
	if options.Adapter == nil {
		options.Adapter = NewLocalStorageAdapter("golid_store_")
	}
	if options.Serializer == nil {
		options.Serializer = NewJSONSerializer()
	}
	if options.Throttle == 0 {
		options.Throttle = 100 * time.Millisecond
	}

	ps := &PersistentStore[T]{
		store:   store,
		options: options,
	}

	// Load initial state if it exists
	if err := ps.Load(); err != nil {
		// If loading fails, that's okay - we'll use the current store state
	}

	// Set up auto-save if enabled
	if options.AutoSave {
		ps.setupAutoSave()
	}

	ps.initialized = true
	return ps, nil
}

// Load restores store state from persistence
func (ps *PersistentStore[T]) Load() error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if !ps.options.Adapter.Exists(ps.options.Key) {
		return fmt.Errorf("no persisted state found for key: %s", ps.options.Key)
	}

	data, err := ps.options.Adapter.Load(ps.options.Key)
	if err != nil {
		return fmt.Errorf("failed to load persisted state: %w", err)
	}

	var value T
	if err := ps.options.Serializer.Deserialize(data, &value); err != nil {
		return fmt.Errorf("failed to deserialize persisted state: %w", err)
	}

	// Update store with loaded value
	ps.store.Set(value)
	return nil
}

// Save persists current store state
func (ps *PersistentStore[T]) Save() error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	value := ps.store.Get()
	data, err := ps.options.Serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize store state: %w", err)
	}

	if err := ps.options.Adapter.Save(ps.options.Key, data); err != nil {
		return fmt.Errorf("failed to save store state: %w", err)
	}

	ps.lastSave = time.Now()
	return nil
}

// Delete removes persisted state
func (ps *PersistentStore[T]) Delete() error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	return ps.options.Adapter.Delete(ps.options.Key)
}

// setupAutoSave configures automatic saving
func (ps *PersistentStore[T]) setupAutoSave() {
	// Subscribe to store changes
	ps.store.Subscribe(func(value T) {
		if !ps.initialized {
			return
		}

		ps.mutex.Lock()
		defer ps.mutex.Unlock()

		// Cancel existing timer
		if ps.saveTimer != nil {
			ps.saveTimer.Stop()
		}

		// Set up throttled save
		ps.saveTimer = time.AfterFunc(ps.options.Throttle, func() {
			if err := ps.Save(); err != nil {
				// Handle save error (could emit to error handler)
				fmt.Printf("Auto-save failed: %v\n", err)
			}
		})
	})
}

// GetStore returns the underlying store
func (ps *PersistentStore[T]) GetStore() *Store[T] {
	return ps.store
}

// GetLastSaveTime returns the last save timestamp
func (ps *PersistentStore[T]) GetLastSaveTime() time.Time {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.lastSave
}

// ------------------------------------
// 🔄 Store Hydration Utilities
// ------------------------------------

// HydrationData represents data for store hydration
type HydrationData struct {
	Stores    map[string]interface{} `json:"stores"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
}

// StoreHydrator manages hydration of multiple stores
type StoreHydrator struct {
	stores     map[string]interface{}
	serializer Serializer
	mutex      sync.RWMutex
}

// NewStoreHydrator creates a new store hydrator
func NewStoreHydrator() *StoreHydrator {
	return &StoreHydrator{
		stores:     make(map[string]interface{}),
		serializer: NewJSONSerializer(),
	}
}

// RegisterStore registers a store for hydration
func (h *StoreHydrator) RegisterStore(name string, store interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.stores[name] = store
}

// Dehydrate extracts state from all registered stores
func (h *StoreHydrator) Dehydrate() (*HydrationData, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	data := &HydrationData{
		Stores:    make(map[string]interface{}),
		Timestamp: time.Now(),
		Version:   "1.0",
	}

	for name, storeInterface := range h.stores {
		// Use reflection to get store value
		storeValue := reflect.ValueOf(storeInterface)
		if storeValue.Kind() == reflect.Ptr {
			storeValue = storeValue.Elem()
		}

		// Look for Get method
		getMethod := storeValue.MethodByName("Get")
		if !getMethod.IsValid() {
			continue
		}

		// Call Get method
		results := getMethod.Call(nil)
		if len(results) > 0 {
			data.Stores[name] = results[0].Interface()
		}
	}

	return data, nil
}

// Hydrate restores state to all registered stores
func (h *StoreHydrator) Hydrate(data *HydrationData) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for name, value := range data.Stores {
		storeInterface, exists := h.stores[name]
		if !exists {
			continue
		}

		// Use reflection to set store value
		storeValue := reflect.ValueOf(storeInterface)
		if storeValue.Kind() == reflect.Ptr {
			storeValue = storeValue.Elem()
		}

		// Look for Set method
		setMethod := storeValue.MethodByName("Set")
		if !setMethod.IsValid() {
			continue
		}

		// Call Set method with value
		args := []reflect.Value{reflect.ValueOf(value)}
		setMethod.Call(args)
	}

	return nil
}

// SerializeHydrationData converts hydration data to bytes
func (h *StoreHydrator) SerializeHydrationData(data *HydrationData) ([]byte, error) {
	return h.serializer.Serialize(data)
}

// DeserializeHydrationData converts bytes to hydration data
func (h *StoreHydrator) DeserializeHydrationData(bytes []byte) (*HydrationData, error) {
	var data HydrationData
	err := h.serializer.Deserialize(bytes, &data)
	return &data, err
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// CreateTestPersistentStore creates a persistent store for testing
func CreateTestPersistentStore[T any](initialState T, key string) (*PersistentStore[T], func()) {
	store := CreateStore(initialState)
	adapter := NewMemoryAdapter()

	options := PersistenceOptions{
		Key:      key,
		Adapter:  adapter,
		AutoSave: true,
		Throttle: 10 * time.Millisecond,
	}

	persistentStore, _ := PersistStore(store, options)

	cleanup := func() {
		adapter.Clear()
		store.Dispose()
	}

	return persistentStore, cleanup
}
