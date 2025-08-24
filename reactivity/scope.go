package reactivity

import (
	"fmt"
	"syscall/js"
)

// CleanupScope represents a container-scoped cleanup context that manages
// disposers for effects, event listeners, and other resources tied to a subtree.
type CleanupScope struct {
	parent    *CleanupScope
	children  []*CleanupScope
	disposers []func()
	disposed  bool
}

// currentCleanupScope holds the currently active cleanup scope
var currentCleanupScope *CleanupScope

// NewCleanupScope creates a new cleanup scope with an optional parent.
// If parent is provided, this scope will be automatically disposed when the parent is disposed.
func NewCleanupScope(parent *CleanupScope) *CleanupScope {
	scope := &CleanupScope{
		parent:    parent,
		children:  make([]*CleanupScope, 0),
		disposers: make([]func(), 0),
		disposed:  false,
	}
	
	// Register with parent if provided
	if parent != nil {
		parent.children = append(parent.children, scope)
	}
	
	return scope
}

// RegisterDisposer registers a function to be called when this scope is disposed.
// If the scope is already disposed, the disposer is ignored.
func (s *CleanupScope) RegisterDisposer(disposer func()) {
	if s.disposed {
		return
	}
	s.disposers = append(s.disposers, disposer)
}

// GetParent returns the parent scope of this cleanup scope.
func (s *CleanupScope) GetParent() *CleanupScope {
	return s.parent
}

// SetParent sets the parent scope of this cleanup scope.
// If this scope already has a parent, it will be removed from the old parent's children.
// If newParent is provided, this scope will be added to the new parent's children.
func (s *CleanupScope) SetParent(newParent *CleanupScope) {
	// Remove from old parent's children if we have one
	if s.parent != nil {
		for i, child := range s.parent.children {
			if child == s {
				// Remove this scope from parent's children slice
				s.parent.children = append(s.parent.children[:i], s.parent.children[i+1:]...)
				break
			}
		}
	}
	
	// Set new parent
	s.parent = newParent
	
	// Add to new parent's children if provided
	if newParent != nil {
		newParent.children = append(newParent.children, s)
	}
}

// Dispose runs all registered disposers and recursively disposes all child scopes.
// This method is idempotent - calling it multiple times has no additional effect.
func (s *CleanupScope) Dispose() {
	if s.disposed {
		return
	}
	
	// Debug: Log scope disposal
	if js.Global().Truthy() {
		js.Global().Get("console").Call("log", fmt.Sprintf("[CleanupScope] Disposing scope with %d disposers", len(s.disposers)))
	}
	
	s.disposed = true
	
	// Dispose all children first
	for _, child := range s.children {
		child.Dispose()
	}
	
	// Run all disposers
	for _, disposer := range s.disposers {
		disposer()
	}
	
	// Clear references to help GC
	s.disposers = nil
	s.children = nil
}

// GetCurrentCleanupScope returns the currently active cleanup scope.
func GetCurrentCleanupScope() *CleanupScope {
	return currentCleanupScope
}

// SetCurrentCleanupScope sets the currently active cleanup scope.
func SetCurrentCleanupScope(scope *CleanupScope) {
	currentCleanupScope = scope
}

// RegisterCleanup registers a disposer with the current cleanup scope.
// If no current scope is set, the registration is ignored.
func RegisterCleanup(disposer func()) {
	if currentCleanupScope != nil {
		currentCleanupScope.RegisterDisposer(disposer)
	}
}

// WithCleanupScope executes a function within a new cleanup scope context.
// The scope is automatically disposed after the function completes.
func WithCleanupScope(parent *CleanupScope, fn func(*CleanupScope)) {
	scope := NewCleanupScope(parent)
	previous := currentCleanupScope
	SetCurrentCleanupScope(scope)
	
	defer func() {
		SetCurrentCleanupScope(previous)
		scope.Dispose()
	}()
	
	fn(scope)
}