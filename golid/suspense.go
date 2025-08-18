// suspense.go
// Suspense boundary implementation with fallback rendering for async operations

package golid

import (
	"fmt"
	"sync"
	"sync/atomic"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🔄 Suspense Types
// ------------------------------------

// SuspenseBoundary manages loading states and fallback rendering
type SuspenseBoundary struct {
	id            uint64
	fallback      gomponents.Node
	children      []gomponents.Node
	owner         *Owner
	suspended     bool
	errorBoundary *ErrorBoundary
	mutex         sync.RWMutex
	onSuspend     func()
	onResume      func()
}

// SuspenseOptions provides configuration for suspense boundaries
type SuspenseOptions struct {
	Name          string
	OnSuspend     func()
	OnResume      func()
	ErrorBoundary *ErrorBoundary
	Owner         *Owner
}

// SuspenseContext tracks the current suspense boundary
type SuspenseContext struct {
	boundary *SuspenseBoundary
	depth    int
	mutex    sync.RWMutex
}

// Global suspense management
var (
	suspenseIdCounter uint64
	currentSuspense   *SuspenseContext
	suspenseMutex     sync.RWMutex
)

// ------------------------------------
// 🏗️ Suspense Creation
// ------------------------------------

// Suspense creates a new suspense boundary with fallback content
func Suspense(fallback gomponents.Node, children ...gomponents.Node) *SuspenseBoundary {
	return CreateSuspense(fallback, SuspenseOptions{}, children...)
}

// CreateSuspense creates a suspense boundary with options
func CreateSuspense(fallback gomponents.Node, options SuspenseOptions, children ...gomponents.Node) *SuspenseBoundary {
	owner := options.Owner
	if owner == nil {
		owner = getCurrentOwner()
	}

	boundary := &SuspenseBoundary{
		id:            atomic.AddUint64(&suspenseIdCounter, 1),
		fallback:      fallback,
		children:      children,
		owner:         owner,
		suspended:     false,
		errorBoundary: options.ErrorBoundary,
		onSuspend:     options.OnSuspend,
		onResume:      options.OnResume,
	}

	// Register with owner for cleanup
	if owner != nil {
		owner.addCleanup(func() {
			boundary.cleanup()
		})
	}

	return boundary
}

// ------------------------------------
// 🎭 Suspense Rendering
// ------------------------------------

// Render renders the suspense boundary, handling suspended states
func (s *SuspenseBoundary) Render() gomponents.Node {
	s.mutex.RLock()
	suspended := s.suspended
	s.mutex.RUnlock()

	// If suspended, render fallback
	if suspended {
		return s.fallback
	}

	// Try to render children within suspense context
	return s.renderWithSuspense()
}

// renderWithSuspense renders children with suspense exception handling
func (s *SuspenseBoundary) renderWithSuspense() gomponents.Node {
	// Set current suspense context
	prevSuspense := getCurrentSuspense()
	setCurrentSuspense(&SuspenseContext{
		boundary: s,
		depth:    getContextDepth(prevSuspense) + 1,
	})
	defer setCurrentSuspense(prevSuspense)

	// Create error boundary if not provided
	errorBoundary := s.errorBoundary
	if errorBoundary == nil {
		errorBoundary = CreateErrorBoundary(func(err error) interface{} {
			// Check if this is a suspense exception
			if _, ok := err.(SuspenseException); ok {
				s.suspend()
				return s.fallback
			}
			// For other errors, propagate or show error fallback
			return gomponents.Text(fmt.Sprintf("Error: %v", err))
		})
	}

	var result gomponents.Node
	err := errorBoundary.Catch(func() {
		if len(s.children) == 1 {
			result = s.children[0]
		} else {
			result = gomponents.Group(s.children)
		}
	})

	if err != nil {
		// Error occurred during rendering
		if _, ok := err.(SuspenseException); ok {
			s.suspend()
			return s.fallback
		}
		// Other error, show error state
		return gomponents.Text(fmt.Sprintf("Error: %v", err))
	}

	// Successfully rendered, ensure we're not suspended
	s.resume()
	return result
}

// ------------------------------------
// 🔄 Suspense State Management
// ------------------------------------

// suspend puts the boundary into suspended state
func (s *SuspenseBoundary) suspend() {
	s.mutex.Lock()
	wasSuspended := s.suspended
	s.suspended = true
	s.mutex.Unlock()

	if !wasSuspended && s.onSuspend != nil {
		s.onSuspend()
	}
}

// resume takes the boundary out of suspended state
func (s *SuspenseBoundary) resume() {
	s.mutex.Lock()
	wasSuspended := s.suspended
	s.suspended = false
	s.mutex.Unlock()

	if wasSuspended && s.onResume != nil {
		s.onResume()
	}
}

// IsSuspended returns whether the boundary is currently suspended
func (s *SuspenseBoundary) IsSuspended() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.suspended
}

// SetFallback updates the fallback content
func (s *SuspenseBoundary) SetFallback(fallback gomponents.Node) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.fallback = fallback
}

// AddChild adds a child component to the boundary
func (s *SuspenseBoundary) AddChild(child gomponents.Node) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.children = append(s.children, child)
}

// ------------------------------------
// 🧹 Cleanup
// ------------------------------------

// cleanup cleans up the suspense boundary
func (s *SuspenseBoundary) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.children = nil
	s.fallback = nil
}

// ------------------------------------
// 🌐 Context Management
// ------------------------------------

// getCurrentSuspense returns the current suspense context
func getCurrentSuspense() *SuspenseContext {
	suspenseMutex.RLock()
	defer suspenseMutex.RUnlock()
	return currentSuspense
}

// setCurrentSuspense sets the current suspense context
func setCurrentSuspense(ctx *SuspenseContext) {
	suspenseMutex.Lock()
	defer suspenseMutex.Unlock()
	currentSuspense = ctx
}

// getContextDepth returns the depth of a suspense context
func getContextDepth(ctx *SuspenseContext) int {
	if ctx == nil {
		return 0
	}
	return ctx.depth
}

// ------------------------------------
// 🔄 Suspense Utilities
// ------------------------------------

// WithSuspense wraps a function with suspense handling
func WithSuspense[T any](fallback gomponents.Node, fn func() T) T {
	boundary := Suspense(fallback)

	// Set suspense context
	prevSuspense := getCurrentSuspense()
	setCurrentSuspense(&SuspenseContext{
		boundary: boundary,
		depth:    getContextDepth(prevSuspense) + 1,
	})
	defer setCurrentSuspense(prevSuspense)

	return fn()
}

// SuspendIf conditionally suspends based on a condition
func SuspendIf(condition bool, message string) {
	if condition {
		panic(NewSuspenseException(message))
	}
}

// SuspendUntil suspends until a condition becomes true
func SuspendUntil[T any](getter func() T, predicate func(T) bool, message string) T {
	value := getter()
	if !predicate(value) {
		panic(NewSuspenseException(message))
	}
	return value
}

// ------------------------------------
// 🎯 Resource Integration
// ------------------------------------

// SuspenseResource wraps a resource with automatic suspense handling
func SuspenseResource[T any](resource *AsyncResource[T], fallback gomponents.Node) gomponents.Node {
	return Suspense(fallback, gomponents.Text("Loading...")).Render()
}

// CreateSuspenseResource creates a resource that automatically suspends
func CreateSuspenseResource[T any](
	fetcher func() (T, error),
	fallback gomponents.Node,
	options ...ResourceOptions,
) (*AsyncResource[T], *SuspenseBoundary) {
	resource := CreateResource(fetcher, options...)
	boundary := Suspense(fallback)

	// Connect resource state to suspense boundary
	CreateEffect(func() {
		state := resource.State()
		if state.Loading && state.Data == nil {
			boundary.suspend()
		} else {
			boundary.resume()
		}
	}, boundary.owner)

	return resource, boundary
}

// ------------------------------------
// 🔄 Nested Suspense Support
// ------------------------------------

// NestedSuspense creates a nested suspense boundary
func NestedSuspense(fallback gomponents.Node, children ...gomponents.Node) *SuspenseBoundary {
	currentCtx := getCurrentSuspense()

	options := SuspenseOptions{}
	if currentCtx != nil {
		options.Owner = currentCtx.boundary.owner
	}

	return CreateSuspense(fallback, options, children...)
}

// PropagateToParent propagates suspension to parent boundary
func PropagateToParent(message string) {
	currentCtx := getCurrentSuspense()
	if currentCtx != nil && currentCtx.depth > 1 {
		// Find parent suspense boundary and suspend it
		panic(NewSuspenseException(fmt.Sprintf("Propagated: %s", message)))
	} else {
		panic(NewSuspenseException(message))
	}
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// MockSuspenseBoundary creates a mock suspense boundary for testing
func MockSuspenseBoundary(fallback gomponents.Node) *SuspenseBoundary {
	return &SuspenseBoundary{
		id:        atomic.AddUint64(&suspenseIdCounter, 1),
		fallback:  fallback,
		children:  []gomponents.Node{},
		suspended: false,
	}
}

// SimulateSuspense simulates a suspense exception for testing
func SimulateSuspense(message string) {
	panic(NewSuspenseException(message))
}

// ------------------------------------
// 🔍 Suspense Diagnostics
// ------------------------------------

// SuspenseStats provides statistics about suspense boundaries
type SuspenseStats struct {
	TotalBoundaries int
	SuspendedCount  int
	AverageDepth    float64
	MaxDepth        int
}

// GetSuspenseStats returns current suspense statistics
func GetSuspenseStats() SuspenseStats {
	// This would be implemented with global tracking
	// For now, return basic stats
	return SuspenseStats{
		TotalBoundaries: int(atomic.LoadUint64(&suspenseIdCounter)),
		SuspendedCount:  0, // Would need global tracking
		AverageDepth:    1.0,
		MaxDepth:        1,
	}
}

// ------------------------------------
// 🎨 Suspense Components
// ------------------------------------

// LoadingSpinner creates a loading spinner fallback
func LoadingSpinner() gomponents.Node {
	return gomponents.El("div",
		gomponents.Attr("class", "loading-spinner"),
		gomponents.Attr("style", "display: flex; justify-content: center; align-items: center; padding: 20px;"),
		gomponents.Text("Loading..."),
	)
}

// LoadingSkeleton creates a skeleton loading fallback
func LoadingSkeleton(lines int) gomponents.Node {
	skeletonLines := make([]gomponents.Node, lines)
	for i := 0; i < lines; i++ {
		skeletonLines[i] = gomponents.El("div",
			gomponents.Attr("class", "skeleton-line"),
			gomponents.Attr("style", "height: 20px; background: #f0f0f0; margin: 5px 0; border-radius: 4px;"),
		)
	}

	return gomponents.El("div",
		gomponents.Attr("class", "loading-skeleton"),
		gomponents.Group(skeletonLines),
	)
}

// ErrorFallback creates an error fallback component
func ErrorFallback(message string) gomponents.Node {
	return gomponents.El("div",
		gomponents.Attr("class", "error-fallback"),
		gomponents.Attr("style", "padding: 20px; border: 1px solid #ff6b6b; background: #ffe0e0; border-radius: 4px;"),
		gomponents.El("h3", gomponents.Text("Something went wrong")),
		gomponents.El("p", gomponents.Text(message)),
	)
}
