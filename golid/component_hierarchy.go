// component_hierarchy.go
// Component ownership and hierarchy management with cascade prevention

//go:build !js && !wasm

package golid

import (
	"sync"
	"sync/atomic"
	"time"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🏗️ Component Hierarchy Management
// ------------------------------------

var hierarchyIdCounter uint64

// ComponentHierarchy manages component parent-child relationships
type ComponentHierarchy struct {
	id       uint64
	root     *HierarchyNode
	nodes    map[uint64]*HierarchyNode
	depth    int
	maxDepth int
	mutex    sync.RWMutex
}

// HierarchyNode represents a node in the component hierarchy
type HierarchyNode struct {
	id       uint64
	parent   *HierarchyNode
	children []*HierarchyNode
	owner    *Owner
	mounted  bool
	disposed bool
	mutex    sync.RWMutex
}

// NewComponentHierarchy creates a new component hierarchy
func NewComponentHierarchy(maxDepth int) *ComponentHierarchy {
	return &ComponentHierarchy{
		id:       atomic.AddUint64(&hierarchyIdCounter, 1),
		nodes:    make(map[uint64]*HierarchyNode),
		maxDepth: maxDepth,
	}
}

// ------------------------------------
// 🔗 Hierarchy Operations
// ------------------------------------

// CreateHierarchyNode creates a new hierarchy node
func (h *ComponentHierarchy) CreateHierarchyNode(owner *Owner) *HierarchyNode {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := &HierarchyNode{
		id:       atomic.AddUint64(&hierarchyIdCounter, 1),
		owner:    owner,
		children: make([]*HierarchyNode, 0),
	}

	h.nodes[node.id] = node

	// Set as root if this is the first node
	if h.root == nil {
		h.root = node
	}

	return node
}

// AttachChild attaches a child node to a parent with depth checking
func (h *ComponentHierarchy) AttachChild(parent, child *HierarchyNode) error {
	if parent == nil || child == nil {
		return ErrInvalidHierarchyNode
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check depth limit
	depth := h.calculateDepth(parent)
	if depth >= h.maxDepth {
		return ErrHierarchyDepthExceeded
	}

	// Prevent circular references
	if h.wouldCreateCycle(parent, child) {
		return ErrCircularHierarchy
	}

	parent.mutex.Lock()
	child.mutex.Lock()
	defer parent.mutex.Unlock()
	defer child.mutex.Unlock()

	// Remove child from previous parent if any
	if child.parent != nil {
		h.detachChildUnsafe(child.parent, child)
	}

	// Attach to new parent
	child.parent = parent
	parent.children = append(parent.children, child)

	return nil
}

// DetachChild detaches a child from its parent
func (h *ComponentHierarchy) DetachChild(parent, child *HierarchyNode) error {
	if parent == nil || child == nil {
		return ErrInvalidHierarchyNode
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	parent.mutex.Lock()
	child.mutex.Lock()
	defer parent.mutex.Unlock()
	defer child.mutex.Unlock()

	return h.detachChildUnsafe(parent, child)
}

// detachChildUnsafe detaches a child without locking (internal use)
func (h *ComponentHierarchy) detachChildUnsafe(parent, child *HierarchyNode) error {
	for i, c := range parent.children {
		if c == child {
			parent.children = append(parent.children[:i], parent.children[i+1:]...)
			child.parent = nil
			return nil
		}
	}
	return ErrChildNotFound
}

// ------------------------------------
// 🧹 Cleanup Operations
// ------------------------------------

// DisposeNode disposes a node and all its children
func (h *ComponentHierarchy) DisposeNode(node *HierarchyNode) error {
	if node == nil {
		return ErrInvalidHierarchyNode
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.disposeNodeRecursive(node)
}

// disposeNodeRecursive recursively disposes a node and its children
func (h *ComponentHierarchy) disposeNodeRecursive(node *HierarchyNode) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.disposed {
		return nil
	}

	// Dispose children first
	children := make([]*HierarchyNode, len(node.children))
	copy(children, node.children)

	for _, child := range children {
		h.disposeNodeRecursive(child)
	}

	// Dispose owner if present
	if node.owner != nil {
		node.owner.dispose()
	}

	// Remove from parent
	if node.parent != nil {
		h.detachChildUnsafe(node.parent, node)
	}

	// Mark as disposed
	node.disposed = true
	node.children = nil

	// Remove from hierarchy
	delete(h.nodes, node.id)

	return nil
}

// CleanupDisposed removes all disposed nodes from the hierarchy
func (h *ComponentHierarchy) CleanupDisposed() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	count := 0
	for id, node := range h.nodes {
		node.mutex.RLock()
		disposed := node.disposed
		node.mutex.RUnlock()

		if disposed {
			delete(h.nodes, id)
			count++
		}
	}

	return count
}

// ------------------------------------
// 🔍 Hierarchy Queries
// ------------------------------------

// calculateDepth calculates the depth of a node in the hierarchy
func (h *ComponentHierarchy) calculateDepth(node *HierarchyNode) int {
	depth := 0
	current := node

	for current.parent != nil {
		depth++
		current = current.parent
		if depth > h.maxDepth {
			break // Prevent infinite loops
		}
	}

	return depth
}

// wouldCreateCycle checks if attaching child to parent would create a cycle
func (h *ComponentHierarchy) wouldCreateCycle(parent, child *HierarchyNode) bool {
	current := parent
	for current != nil {
		if current == child {
			return true
		}
		current = current.parent
	}
	return false
}

// GetNodeDepth returns the depth of a specific node
func (h *ComponentHierarchy) GetNodeDepth(node *HierarchyNode) int {
	if node == nil {
		return -1
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.calculateDepth(node)
}

// GetChildren returns a copy of the node's children
func (h *ComponentHierarchy) GetChildren(node *HierarchyNode) []*HierarchyNode {
	if node == nil {
		return nil
	}

	node.mutex.RLock()
	defer node.mutex.RUnlock()

	children := make([]*HierarchyNode, len(node.children))
	copy(children, node.children)
	return children
}

// GetParent returns the parent of a node
func (h *ComponentHierarchy) GetParent(node *HierarchyNode) *HierarchyNode {
	if node == nil {
		return nil
	}

	node.mutex.RLock()
	defer node.mutex.RUnlock()

	return node.parent
}

// ------------------------------------
// 📊 Hierarchy Statistics
// ------------------------------------

// GetHierarchyStats returns statistics about the hierarchy
func (h *ComponentHierarchy) GetHierarchyStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := map[string]interface{}{
		"totalNodes":    len(h.nodes),
		"maxDepth":      h.maxDepth,
		"currentDepth":  0,
		"disposedNodes": 0,
		"mountedNodes":  0,
	}

	maxCurrentDepth := 0
	disposedCount := 0
	mountedCount := 0

	for _, node := range h.nodes {
		node.mutex.RLock()
		if node.disposed {
			disposedCount++
		}
		if node.mounted {
			mountedCount++
		}
		depth := h.calculateDepth(node)
		if depth > maxCurrentDepth {
			maxCurrentDepth = depth
		}
		node.mutex.RUnlock()
	}

	stats["currentDepth"] = maxCurrentDepth
	stats["disposedNodes"] = disposedCount
	stats["mountedNodes"] = mountedCount

	return stats
}

// ------------------------------------
// 🛡️ Cascade Prevention
// ------------------------------------

// CascadePreventionGuard prevents infinite cascades in hierarchy operations
type CascadePreventionGuard struct {
	operations map[uint64]time.Time
	maxOps     int
	timeWindow time.Duration
	mutex      sync.RWMutex
}

// NewCascadePreventionGuard creates a new cascade prevention guard
func NewCascadePreventionGuard(maxOps int, timeWindow time.Duration) *CascadePreventionGuard {
	return &CascadePreventionGuard{
		operations: make(map[uint64]time.Time),
		maxOps:     maxOps,
		timeWindow: timeWindow,
	}
}

// CheckOperation checks if an operation should be allowed
func (g *CascadePreventionGuard) CheckOperation(nodeId uint64) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	now := time.Now()

	// Clean up old operations
	for id, timestamp := range g.operations {
		if now.Sub(timestamp) > g.timeWindow {
			delete(g.operations, id)
		}
	}

	// Count operations for this node
	count := 0
	for id := range g.operations {
		if id == nodeId {
			count++
		}
	}

	// Check if we're over the limit
	if count >= g.maxOps {
		return false
	}

	// Record this operation
	g.operations[nodeId] = now
	return true
}

// Reset resets the guard state
func (g *CascadePreventionGuard) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.operations = make(map[uint64]time.Time)
}

// ------------------------------------
// 🔧 Utility Functions
// ------------------------------------

// CreateComponentWithHierarchy creates a component with hierarchy management
func CreateComponentWithHierarchy(hierarchy *ComponentHierarchy, fn func() gomponents.Node) (*ComponentV2, *HierarchyNode) {
	// Create owner context
	_, cleanup := CreateRoot(func() interface{} { return nil })
	owner := getCurrentOwner()

	// Create hierarchy node
	node := hierarchy.CreateHierarchyNode(owner)

	// Create component
	comp := &ComponentV2{
		id:        atomic.AddUint64(&componentV2IdCounter, 1),
		owner:     owner,
		state:     ComponentUnmountedV2,
		render:    fn,
		resources: NewResourceTracker(),
		guard:     NewLifecycleGuardV2(5),
		children:  make([]*ComponentV2, 0),
	}

	// Register cleanup
	OnCleanup(func() {
		hierarchy.DisposeNode(node)
		cleanup()
	})

	return comp, node
}

// MountComponentWithHierarchy mounts a component within a hierarchy
func MountComponentWithHierarchy(hierarchy *ComponentHierarchy, comp *ComponentV2, node *HierarchyNode, parent *HierarchyNode) error {
	// Attach to hierarchy
	if parent != nil {
		if err := hierarchy.AttachChild(parent, node); err != nil {
			return err
		}
	}

	// Mount component
	node.mutex.Lock()
	node.mounted = true
	node.mutex.Unlock()

	return MountComponentV2(comp, nil)
}

// ------------------------------------
// 🚨 Error Types
// ------------------------------------

var (
	ErrInvalidHierarchyNode   = &HierarchyError{Type: "InvalidNode", Message: "Invalid hierarchy node"}
	ErrHierarchyDepthExceeded = &HierarchyError{Type: "DepthExceeded", Message: "Hierarchy depth limit exceeded"}
	ErrCircularHierarchy      = &HierarchyError{Type: "CircularReference", Message: "Circular hierarchy reference detected"}
	ErrChildNotFound          = &HierarchyError{Type: "ChildNotFound", Message: "Child node not found in parent"}
)

// HierarchyError represents a hierarchy-specific error
type HierarchyError struct {
	Type    string
	Message string
}

func (e *HierarchyError) Error() string {
	return e.Type + ": " + e.Message
}
