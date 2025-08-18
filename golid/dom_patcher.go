// dom_patcher.go
// Fine-grained DOM patching system for batched updates

//go:build js && wasm

package golid

import (
	"sync"
	"syscall/js"
)

// ------------------------------------
// 🔧 DOM Operation Types
// ------------------------------------

// OperationType defines the type of DOM operation
type OperationType int

const (
	SetText OperationType = iota
	SetAttribute
	RemoveAttribute
	SetProperty
	InsertNode
	RemoveNode
	ReplaceNode
)

// DOMOperation represents a single DOM manipulation operation
type DOMOperation struct {
	type_    OperationType
	target   js.Value
	property string
	value    interface{}
	oldValue interface{}
	newNode  js.Value
	refNode  js.Value
}

// ------------------------------------
// 🔄 DOM Patcher Implementation
// ------------------------------------

// DOMPatcher manages batched DOM operations for optimal performance
type DOMPatcher struct {
	operations []DOMOperation
	batching   bool
	mutex      sync.RWMutex
}

// newDOMPatcher creates a new DOM patcher instance
func newDOMPatcher() *DOMPatcher {
	return &DOMPatcher{
		operations: make([]DOMOperation, 0),
		batching:   false,
	}
}

// QueueOperation adds a DOM operation to the batch queue
func (p *DOMPatcher) QueueOperation(op DOMOperation) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !op.target.Truthy() {
		return // Skip invalid targets
	}

	p.operations = append(p.operations, op)

	// Always execute immediately for WASM to ensure updates are applied
	p.flushOperations()
}

// BatchUpdates runs a function within a batched update context
func (p *DOMPatcher) BatchUpdates(fn func()) {
	p.mutex.Lock()
	wasBatching := p.batching
	p.batching = true
	p.mutex.Unlock()

	defer func() {
		p.mutex.Lock()
		p.batching = wasBatching
		shouldFlush := !wasBatching && len(p.operations) > 0
		p.mutex.Unlock()

		if shouldFlush {
			p.Flush()
		}
	}()

	fn()
}

// Flush executes all queued DOM operations
func (p *DOMPatcher) Flush() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.flushOperations()
}

// flushOperations executes all queued operations (must be called with mutex held)
func (p *DOMPatcher) flushOperations() {
	if len(p.operations) == 0 {
		return
	}

	// Group operations by type for optimal execution
	textOps := make([]DOMOperation, 0)
	attrOps := make([]DOMOperation, 0)
	nodeOps := make([]DOMOperation, 0)

	for _, op := range p.operations {
		switch op.type_ {
		case SetText:
			textOps = append(textOps, op)
		case SetAttribute, RemoveAttribute, SetProperty:
			attrOps = append(attrOps, op)
		case InsertNode, RemoveNode, ReplaceNode:
			nodeOps = append(nodeOps, op)
		}
	}

	// Execute operations in optimal order
	p.executeTextOperations(textOps)
	p.executeAttributeOperations(attrOps)
	p.executeNodeOperations(nodeOps)

	// Clear operations
	p.operations = p.operations[:0]
}

// executeTextOperations executes text content operations
func (p *DOMPatcher) executeTextOperations(ops []DOMOperation) {
	for _, op := range ops {
		if !op.target.Truthy() {
			continue
		}

		switch op.type_ {
		case SetText:
			if str, ok := op.value.(string); ok {
				// Only update if value actually changed
				current := op.target.Get("textContent").String()
				if current != str {
					op.target.Set("textContent", str)
				}
			}
		}
	}
}

// executeAttributeOperations executes attribute and property operations
func (p *DOMPatcher) executeAttributeOperations(ops []DOMOperation) {
	for _, op := range ops {
		if !op.target.Truthy() {
			continue
		}

		switch op.type_ {
		case SetAttribute:
			if str, ok := op.value.(string); ok {
				op.target.Call("setAttribute", op.property, str)
			}
		case RemoveAttribute:
			op.target.Call("removeAttribute", op.property)
		case SetProperty:
			op.target.Set(op.property, op.value)
		}
	}
}

// executeNodeOperations executes node manipulation operations
func (p *DOMPatcher) executeNodeOperations(ops []DOMOperation) {
	for _, op := range ops {
		if !op.target.Truthy() {
			continue
		}

		switch op.type_ {
		case InsertNode:
			if op.newNode.Truthy() {
				if op.refNode.Truthy() {
					op.target.Call("insertBefore", op.newNode, op.refNode)
				} else {
					op.target.Call("appendChild", op.newNode)
				}
			}
		case RemoveNode:
			if op.newNode.Truthy() {
				op.target.Call("removeChild", op.newNode)
			}
		case ReplaceNode:
			if op.newNode.Truthy() && op.refNode.Truthy() {
				op.target.Call("replaceChild", op.newNode, op.refNode)
			}
		}
	}
}

// ------------------------------------
// 📊 Performance Optimizations
// ------------------------------------

// OptimizeOperations removes redundant operations and merges compatible ones
func (p *DOMPatcher) OptimizeOperations() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.operations) <= 1 {
		return
	}

	// Track latest operation per target+property
	latest := make(map[string]int)
	optimized := make([]DOMOperation, 0, len(p.operations))

	for _, op := range p.operations {
		key := p.getOperationKey(op)

		// For text and attribute operations, only keep the latest
		if op.type_ == SetText || op.type_ == SetAttribute || op.type_ == SetProperty {
			if prevIndex, exists := latest[key]; exists {
				// Mark previous operation as redundant by setting invalid index
				if prevIndex < len(optimized) {
					optimized[prevIndex].target = js.Undefined()
				}
			}
			latest[key] = len(optimized)
		}

		optimized = append(optimized, op)
	}

	// Remove redundant operations
	filtered := make([]DOMOperation, 0, len(optimized))
	for _, op := range optimized {
		if op.target.Truthy() {
			filtered = append(filtered, op)
		}
	}

	p.operations = filtered
}

// getOperationKey generates a unique key for operation deduplication
func (p *DOMPatcher) getOperationKey(op DOMOperation) string {
	// Use element reference and property as key
	elementId := op.target.Get("id").String()
	if elementId == "" {
		// Fallback to object reference string
		elementId = op.target.String()
	}
	return elementId + ":" + op.property
}

// ------------------------------------
// 📊 Statistics & Debugging
// ------------------------------------

// PatcherStats provides statistics about the DOM patcher
type PatcherStats struct {
	QueuedOperations int
	IsBatching       bool
}

// GetStats returns current patcher statistics
func (p *DOMPatcher) GetStats() PatcherStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return PatcherStats{
		QueuedOperations: len(p.operations),
		IsBatching:       p.batching,
	}
}

// GetQueuedOperationCount returns the number of queued operations
func (p *DOMPatcher) GetQueuedOperationCount() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.operations)
}

// ClearQueue clears all queued operations (for testing)
func (p *DOMPatcher) ClearQueue() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.operations = p.operations[:0]
}
