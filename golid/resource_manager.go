// resource_manager.go
// Resource tracking and cleanup automation for component lifecycle management

package golid

import (
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 🗂️ Resource Management System
// ------------------------------------

var resourceIdCounter uint64

// ResourceType represents the type of resource being tracked
type ResourceType int

const (
	ResourceTimer ResourceType = iota
	ResourceInterval
	ResourceEventListener
	ResourceSignal
	ResourceEffect
	ResourceCustom
)

// String returns the string representation of ResourceType
func (r ResourceType) String() string {
	switch r {
	case ResourceTimer:
		return "Timer"
	case ResourceInterval:
		return "Interval"
	case ResourceEventListener:
		return "EventListener"
	case ResourceSignal:
		return "Signal"
	case ResourceEffect:
		return "Effect"
	case ResourceCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// Resource represents a tracked resource
type Resource struct {
	id       uint64
	type_    ResourceType
	name     string
	cleanup  func()
	created  time.Time
	disposed bool
	mutex    sync.RWMutex
}

// ResourceTracker manages component resources and their cleanup
type ResourceTracker struct {
	resources map[uint64]*Resource
	mutex     sync.RWMutex
	disposed  bool
}

// NewResourceTracker creates a new resource tracker
func NewResourceTracker() *ResourceTracker {
	return &ResourceTracker{
		resources: make(map[uint64]*Resource),
	}
}

// ------------------------------------
// 📝 Resource Registration
// ------------------------------------

// TrackResource registers a resource for automatic cleanup
func (rt *ResourceTracker) TrackResource(type_ ResourceType, name string, cleanup func()) uint64 {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if rt.disposed {
		// If tracker is disposed, execute cleanup immediately
		if cleanup != nil {
			cleanup()
		}
		return 0
	}

	resource := &Resource{
		id:      atomic.AddUint64(&resourceIdCounter, 1),
		type_:   type_,
		name:    name,
		cleanup: cleanup,
		created: time.Now(),
	}

	rt.resources[resource.id] = resource
	return resource.id
}

// TrackTimer tracks a timer for automatic cleanup
func (rt *ResourceTracker) TrackTimer(name string, cleanup func()) uint64 {
	return rt.TrackResource(ResourceTimer, name, cleanup)
}

// TrackInterval tracks an interval for automatic cleanup
func (rt *ResourceTracker) TrackInterval(name string, cleanup func()) uint64 {
	return rt.TrackResource(ResourceInterval, name, cleanup)
}

// TrackEventListener tracks an event listener for automatic cleanup
func (rt *ResourceTracker) TrackEventListener(name string, cleanup func()) uint64 {
	return rt.TrackResource(ResourceEventListener, name, cleanup)
}

// TrackSignal tracks a signal for automatic cleanup
func (rt *ResourceTracker) TrackSignal(name string, cleanup func()) uint64 {
	return rt.TrackResource(ResourceSignal, name, cleanup)
}

// TrackEffect tracks an effect for automatic cleanup
func (rt *ResourceTracker) TrackEffect(name string, cleanup func()) uint64 {
	return rt.TrackResource(ResourceEffect, name, cleanup)
}

// ------------------------------------
// 🧹 Resource Cleanup
// ------------------------------------

// CleanupResource cleans up a specific resource
func (rt *ResourceTracker) CleanupResource(id uint64) bool {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	resource, exists := rt.resources[id]
	if !exists {
		return false
	}

	resource.mutex.Lock()
	if !resource.disposed && resource.cleanup != nil {
		resource.cleanup()
		resource.disposed = true
	}
	resource.mutex.Unlock()

	delete(rt.resources, id)
	return true
}

// CleanupResourcesByType cleans up all resources of a specific type
func (rt *ResourceTracker) CleanupResourcesByType(type_ ResourceType) int {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	count := 0
	for id, resource := range rt.resources {
		if resource.type_ == type_ {
			resource.mutex.Lock()
			if !resource.disposed && resource.cleanup != nil {
				resource.cleanup()
				resource.disposed = true
				count++
			}
			resource.mutex.Unlock()
			delete(rt.resources, id)
		}
	}

	return count
}

// CleanupAll cleans up all tracked resources
func (rt *ResourceTracker) CleanupAll() int {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if rt.disposed {
		return 0
	}

	count := 0
	for id, resource := range rt.resources {
		resource.mutex.Lock()
		if !resource.disposed && resource.cleanup != nil {
			resource.cleanup()
			resource.disposed = true
			count++
		}
		resource.mutex.Unlock()
		delete(rt.resources, id)
	}

	rt.disposed = true
	return count
}

// ------------------------------------
// 📊 Resource Statistics
// ------------------------------------

// GetResourceCount returns the number of tracked resources
func (rt *ResourceTracker) GetResourceCount() int {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()
	return len(rt.resources)
}

// GetResourcesByType returns resources grouped by type
func (rt *ResourceTracker) GetResourcesByType() map[ResourceType]int {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	counts := make(map[ResourceType]int)
	for _, resource := range rt.resources {
		counts[resource.type_]++
	}

	return counts
}

// GetResourceStats returns detailed resource statistics
func (rt *ResourceTracker) GetResourceStats() map[string]interface{} {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	stats := map[string]interface{}{
		"totalResources": len(rt.resources),
		"disposed":       rt.disposed,
		"byType":         make(map[string]int),
		"oldestResource": time.Time{},
	}

	byType := make(map[string]int)
	var oldest time.Time

	for _, resource := range rt.resources {
		byType[resource.type_.String()]++
		if oldest.IsZero() || resource.created.Before(oldest) {
			oldest = resource.created
		}
	}

	stats["byType"] = byType
	if !oldest.IsZero() {
		stats["oldestResource"] = oldest
	}

	return stats
}

// ------------------------------------
// 🛡️ Legacy Leak Detection Support
// ------------------------------------

// LegacyLeakViolation represents a detected memory leak (legacy compatibility)
type LegacyLeakViolation struct {
	ResourceType ResourceType
	Count        int
	Threshold    int
	DetectedAt   time.Time
}

// CheckForLeaks provides legacy compatibility for existing tests
func CheckForLeaks(tracker *ResourceTracker, threshold int) []LegacyLeakViolation {
	violations := make([]LegacyLeakViolation, 0)

	resourcesByType := tracker.GetResourcesByType()
	for resourceType, count := range resourcesByType {
		if count > threshold {
			violation := LegacyLeakViolation{
				ResourceType: resourceType,
				Count:        count,
				Threshold:    threshold,
				DetectedAt:   time.Now(),
			}
			violations = append(violations, violation)
		}
	}

	return violations
}

// ------------------------------------
// 🔧 Utility Functions
// ------------------------------------

// WithResourceCleanup wraps a function with automatic resource cleanup
func WithResourceCleanup(tracker *ResourceTracker, fn func(*ResourceTracker)) {
	defer tracker.CleanupAll()
	fn(tracker)
}

// AutoCleanupTimer creates a timer with automatic cleanup
func AutoCleanupTimer(tracker *ResourceTracker, duration time.Duration, fn func()) {
	timer := time.AfterFunc(duration, fn)
	tracker.TrackTimer("auto-timer", func() {
		timer.Stop()
	})
}

// AutoCleanupInterval creates an interval with automatic cleanup
func AutoCleanupInterval(tracker *ResourceTracker, duration time.Duration, fn func()) {
	ticker := time.NewTicker(duration)
	tracker.TrackInterval("auto-interval", func() {
		ticker.Stop()
	})

	go func() {
		for range ticker.C {
			fn()
		}
	}()
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// MockResource creates a mock resource for testing
func MockResource(tracker *ResourceTracker, type_ ResourceType, name string) uint64 {
	var cleaned bool
	return tracker.TrackResource(type_, name, func() {
		cleaned = true
		_ = cleaned // Prevent unused variable warning
	})
}

// GetGlobalResourceStats returns global resource statistics
func GetGlobalResourceStats() map[string]interface{} {
	// This would be implemented to track global resource usage
	// across all components and trackers
	return map[string]interface{}{
		"totalTrackers":  0, // Would be implemented with global tracking
		"totalResources": 0,
	}
}
