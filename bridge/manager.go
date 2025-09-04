package bridge

import (
	"errors"
	"sync"
)

var (
	// globalManager holds the current manager instance
	globalManager Manager
	// managerMutex protects access to globalManager
	managerMutex sync.RWMutex
	// ErrManagerNotInitialized is returned when manager is not set
	ErrManagerNotInitialized = errors.New("bridge manager not initialized")
	// ErrComponentNotFound is returned when component cannot be found
	ErrComponentNotFound = errors.New("component not found")
	// ErrInvalidSelector is returned when selector is invalid
	ErrInvalidSelector = errors.New("invalid selector")
)

// InitializeManager initializes the global manager with the provided implementation
func InitializeManager(manager Manager) {
	managerMutex.Lock()
	defer managerMutex.Unlock()
	globalManager = manager
}

// GetManager returns the current global manager
func GetManager() Manager {
	managerMutex.RLock()
	defer managerMutex.RUnlock()
	return globalManager
}

// SetManager sets the global manager (alias for InitializeManager)
func SetManager(manager Manager) {
	InitializeManager(manager)
}

// ResetManager clears the global manager
func ResetManager() {
	managerMutex.Lock()
	defer managerMutex.Unlock()
	globalManager = nil
}

// IsManagerInitialized returns true if manager is initialized
func IsManagerInitialized() bool {
	managerMutex.RLock()
	defer managerMutex.RUnlock()
	return globalManager != nil
}

// Convenience wrappers for common operations

// InitializeComponent initializes a component using the global manager
func InitializeComponent(componentName string, selector string, options map[string]any) error {
	manager := GetManager()
	if manager == nil {
		return ErrManagerNotInitialized
	}
	return manager.Component().InitializeComponent(componentName, selector, options)
}

// DestroyComponent destroys a component using the global manager
func DestroyComponent(selector string, componentType string) error {
	manager := GetManager()
	if manager == nil {
		return ErrManagerNotInitialized
	}
	return manager.Component().DestroyComponent(selector, componentType)
}

// GetComponentInstance gets a component instance using the global manager
func GetComponentInstance(selector, componentType string) (JSValue, error) {
	manager := GetManager()
	if manager == nil {
		return nil, ErrManagerNotInitialized
	}
	return manager.Component().GetComponentInstance(selector, componentType)
}

// InitializeAllComponents initializes multiple components using the global manager
func InitializeAllComponents(components []string) error {
	manager := GetManager()
	if manager == nil {
		return ErrManagerNotInitialized
	}
	return manager.Component().InitializeAll(components)
}

// GetDocument returns the document using the global manager
func GetDocument() (DOMDocument, error) {
	manager := GetManager()
	if manager == nil {
		return nil, ErrManagerNotInitialized
	}
	return manager.DOM().Document(), nil
}

// GetWindow returns the window using the global manager
func GetWindow() (JSValue, error) {
	manager := GetManager()
	if manager == nil {
		return nil, ErrManagerNotInitialized
	}
	return manager.DOM().Window(), nil
}

// QuerySelector finds an element using the global manager
func QuerySelector(selector string) (DOMElement, error) {
	manager := GetManager()
	if manager == nil {
		return nil, ErrManagerNotInitialized
	}
	doc := manager.DOM().Document()
	if doc == nil {
		return nil, errors.New("document not available")
	}
	return doc.QuerySelector(selector), nil
}

// QuerySelectorAll finds all elements using the global manager
func QuerySelectorAll(selector string) ([]DOMElement, error) {
	manager := GetManager()
	if manager == nil {
		return nil, ErrManagerNotInitialized
	}
	doc := manager.DOM().Document()
	if doc == nil {
		return nil, errors.New("document not available")
	}
	return doc.QuerySelectorAll(selector), nil
}

// GetElementByID finds an element by ID using the global manager
func GetElementByID(id string) (DOMElement, error) {
	manager := GetManager()
	if manager == nil {
		return nil, ErrManagerNotInitialized
	}
	doc := manager.DOM().Document()
	if doc == nil {
		return nil, errors.New("document not available")
	}
	return doc.GetElementByID(id), nil
}