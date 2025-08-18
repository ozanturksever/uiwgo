// memory_integration.go
// Integration of memory management system with existing Owner contexts and reactive systems

package golid

import (
	"reflect"
	"sync"
	"time"
)

// ------------------------------------
// 🔗 Memory Management Integration
// ------------------------------------

var (
	globalMemoryIntegration *MemoryIntegration
	integrationOnce         sync.Once
)

// MemoryIntegration provides integration between memory management and reactive systems
type MemoryIntegration struct {
	memoryManager *MemoryManager
	enabled       bool
	mutex         sync.RWMutex
}

// GetGlobalMemoryIntegration returns the global memory integration instance
func GetGlobalMemoryIntegration() *MemoryIntegration {
	integrationOnce.Do(func() {
		globalMemoryIntegration = &MemoryIntegration{
			memoryManager: GetGlobalMemoryManager(),
			enabled:       true,
		}
	})
	return globalMemoryIntegration
}

// ------------------------------------
// 🎯 Signal Integration
// ------------------------------------

// Enhanced CreateSignal with memory tracking
func CreateSignalWithMemoryTracking[T any](initial T, options ...SignalOptions[T]) (func() T, func(T)) {
	integration := GetGlobalMemoryIntegration()

	// Create the signal normally
	getter, setter := CreateSignal(initial, options...)

	// Track the signal allocation if memory management is enabled
	if integration.enabled && integration.memoryManager != nil {
		owner := getCurrentOwner()

		// Estimate signal size (simplified)
		var sizeEstimate int64 = 64 // Base signal overhead

		// Track the allocation
		alloc := integration.memoryManager.TrackAllocation(
			AllocSignal,
			sizeEstimate,
			owner,
			map[string]interface{}{
				"type":       "signal",
				"value_type": getTypeName[T](),
				"has_owner":  owner != nil,
			},
		)

		// Set up cleanup when owner is disposed
		if owner != nil && alloc != nil {
			owner.addCleanup(func() {
				integration.memoryManager.ReleaseAllocation(alloc.ID)
			})
		}
	}

	return getter, setter
}

// Enhanced CreateEffect with memory tracking
func CreateEffectWithMemoryTracking(fn func(), owner *Owner) *Computation {
	integration := GetGlobalMemoryIntegration()

	// Create the effect normally
	comp := CreateEffect(fn, owner)

	// Track the effect allocation if memory management is enabled
	if integration.enabled && integration.memoryManager != nil {
		if owner == nil {
			owner = getCurrentOwner()
		}

		// Track the allocation
		alloc := integration.memoryManager.TrackAllocation(
			AllocEffect,
			128, // Estimated effect overhead
			owner,
			map[string]interface{}{
				"type":           "effect",
				"computation_id": comp.id,
				"has_owner":      owner != nil,
			},
		)

		// Set up cleanup
		if owner != nil && alloc != nil {
			owner.addCleanup(func() {
				integration.memoryManager.ReleaseAllocation(alloc.ID)
			})
		}

		// Also track in computation cleanup
		if comp != nil && alloc != nil {
			comp.onCleanup(func() {
				integration.memoryManager.ReleaseAllocation(alloc.ID)
			})
		}
	}

	return comp
}

// Enhanced CreateMemo with memory tracking
func CreateMemoWithMemoryTracking[T any](fn func() T, owner *Owner) func() T {
	integration := GetGlobalMemoryIntegration()

	// Create the memo normally
	memoFn := CreateMemo(fn, owner)

	// Track the memo allocation if memory management is enabled
	if integration.enabled && integration.memoryManager != nil {
		if owner == nil {
			owner = getCurrentOwner()
		}

		// Track the allocation
		alloc := integration.memoryManager.TrackAllocation(
			AllocMemo,
			96, // Estimated memo overhead
			owner,
			map[string]interface{}{
				"type":       "memo",
				"value_type": getTypeName[T](),
				"has_owner":  owner != nil,
			},
		)

		// Set up cleanup
		if owner != nil && alloc != nil {
			owner.addCleanup(func() {
				integration.memoryManager.ReleaseAllocation(alloc.ID)
			})
		}
	}

	return memoFn
}

// ------------------------------------
// 🏠 Owner Integration
// ------------------------------------

// Enhanced CreateRoot with memory tracking
func CreateRootWithMemoryTracking[T any](fn func() T) (T, func()) {
	integration := GetGlobalMemoryIntegration()

	// Create the root normally
	result, cleanup := CreateRoot(fn)

	// Track the owner allocation if memory management is enabled
	if integration.enabled && integration.memoryManager != nil {
		owner := getCurrentOwner()

		if owner != nil {
			// Track the owner allocation
			alloc := integration.memoryManager.TrackAllocation(
				AllocComponent,
				256,          // Estimated owner overhead
				owner.parent, // Parent owner
				map[string]interface{}{
					"type":       "owner",
					"owner_id":   owner.id,
					"has_parent": owner.parent != nil,
				},
			)

			// Enhanced cleanup function
			enhancedCleanup := func() {
				// Perform memory leak check before cleanup
				if integration.memoryManager.leakDetector != nil {
					ownerAllocs := integration.memoryManager.GetAllocationsByOwner(owner)
					if len(ownerAllocs) > 0 {
						// Log potential leaks
						for _, ownerAlloc := range ownerAllocs {
							if !ownerAlloc.Disposed {
								// Mark as suspicious
								integration.memoryManager.leakDetector.CheckAllocation(ownerAlloc)
							}
						}
					}
				}

				// Original cleanup
				cleanup()

				// Release owner allocation
				if alloc != nil {
					integration.memoryManager.ReleaseAllocation(alloc.ID)
				}
			}

			return result, enhancedCleanup
		}
	}

	return result, cleanup
}

// ------------------------------------
// 📦 Resource Integration
// ------------------------------------

// Enhanced CreateResource with memory tracking
func CreateResourceWithMemoryTracking[T any](fetcher func() (T, error), options ...ResourceOptions) *AsyncResource[T] {
	integration := GetGlobalMemoryIntegration()

	// Create the resource normally
	resource := CreateResource(fetcher, options...)

	// Track the resource allocation if memory management is enabled
	if integration.enabled && integration.memoryManager != nil {
		owner := getCurrentOwner()

		// Track the allocation
		alloc := integration.memoryManager.TrackAllocation(
			AllocResource,
			512, // Estimated resource overhead
			owner,
			map[string]interface{}{
				"type":        "async_resource",
				"resource_id": resource.id,
				"has_owner":   owner != nil,
				"has_cache":   resource.cache != nil,
			},
		)

		// Set up cleanup
		if owner != nil && alloc != nil {
			owner.addCleanup(func() {
				integration.memoryManager.ReleaseAllocation(alloc.ID)
			})
		}
	}

	return resource
}

// ------------------------------------
// 🔧 Utility Functions
// ------------------------------------

// getTypeName returns the type name for a generic type
func getTypeName[T any]() string {
	var zero T
	return reflect.TypeOf(zero).String()
}

// TrackDOMBinding tracks DOM binding allocations
func TrackDOMBinding(elementID string, eventType string, owner *Owner) uint64 {
	integration := GetGlobalMemoryIntegration()

	if !integration.enabled || integration.memoryManager == nil {
		return 0
	}

	// Track the DOM binding allocation
	alloc := integration.memoryManager.TrackAllocation(
		AllocDOMBinding,
		128, // Estimated DOM binding overhead
		owner,
		map[string]interface{}{
			"type":       "dom_binding",
			"element_id": elementID,
			"event_type": eventType,
			"has_owner":  owner != nil,
		},
	)

	if alloc == nil {
		return 0
	}

	// Set up cleanup
	if owner != nil {
		owner.addCleanup(func() {
			integration.memoryManager.ReleaseAllocation(alloc.ID)
		})
	}

	return alloc.ID
}

// TrackEventSubscription tracks event subscription allocations
func TrackEventSubscription(eventName string, handlerID string, owner *Owner) uint64 {
	integration := GetGlobalMemoryIntegration()

	if !integration.enabled || integration.memoryManager == nil {
		return 0
	}

	// Track the event subscription allocation
	alloc := integration.memoryManager.TrackAllocation(
		AllocEventSubscription,
		96, // Estimated event subscription overhead
		owner,
		map[string]interface{}{
			"type":       "event_subscription",
			"event_name": eventName,
			"handler_id": handlerID,
			"has_owner":  owner != nil,
		},
	)

	if alloc == nil {
		return 0
	}

	// Set up cleanup
	if owner != nil {
		owner.addCleanup(func() {
			integration.memoryManager.ReleaseAllocation(alloc.ID)
		})
	}

	return alloc.ID
}

// ------------------------------------
// 📊 Integration Monitoring
// ------------------------------------

// GetIntegrationStats returns statistics about memory integration
func (mi *MemoryIntegration) GetIntegrationStats() map[string]interface{} {
	mi.mutex.RLock()
	defer mi.mutex.RUnlock()

	stats := map[string]interface{}{
		"enabled": mi.enabled,
	}

	if mi.memoryManager != nil {
		memStats := mi.memoryManager.GetMemoryStats()
		stats["memory_stats"] = memStats

		// Add allocation breakdown
		allocBreakdown := make(map[string]uint64)
		for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
			allocs := mi.memoryManager.GetAllocationsByType(allocType)
			allocBreakdown[allocType.String()] = uint64(len(allocs))
		}
		stats["allocation_breakdown"] = allocBreakdown

		// Add leak detection info
		if mi.memoryManager.leakDetector != nil {
			violations := mi.memoryManager.leakDetector.GetViolations()
			suspicious := mi.memoryManager.leakDetector.GetSuspiciousAllocations()

			stats["leak_detection"] = map[string]interface{}{
				"violations_count": len(violations),
				"suspicious_count": len(suspicious),
			}
		}

		// Add cleanup info
		if mi.memoryManager.cleanupTracker != nil {
			cleanupStats := mi.memoryManager.cleanupTracker.GetStats()
			stats["cleanup_stats"] = cleanupStats
		}
	}

	return stats
}

// PerformMemoryHealthCheck performs a comprehensive memory health check
func (mi *MemoryIntegration) PerformMemoryHealthCheck() map[string]interface{} {
	mi.mutex.RLock()
	defer mi.mutex.RUnlock()

	healthCheck := map[string]interface{}{
		"timestamp": getCurrentTimestamp(),
		"status":    "healthy",
		"issues":    make([]string, 0),
		"warnings":  make([]string, 0),
	}

	if !mi.enabled || mi.memoryManager == nil {
		healthCheck["status"] = "disabled"
		return healthCheck
	}

	issues := make([]string, 0)
	warnings := make([]string, 0)

	// Check memory usage
	memStats := mi.memoryManager.GetMemoryStats()

	// Check for high memory usage
	if memStats.ActiveMemoryUsage > 100*1024*1024 { // 100MB
		issues = append(issues, "High memory usage detected")
	} else if memStats.ActiveMemoryUsage > 50*1024*1024 { // 50MB
		warnings = append(warnings, "Elevated memory usage")
	}

	// Check for leaks
	if memStats.LeaksDetected > 0 {
		issues = append(issues, "Memory leaks detected")
	}

	// Check cleanup success rate
	if memStats.CleanupOperations > 0 {
		successRate := float64(memStats.SuccessfulCleanups) / float64(memStats.CleanupOperations) * 100
		if successRate < 90 {
			issues = append(issues, "Low cleanup success rate")
		} else if successRate < 95 {
			warnings = append(warnings, "Cleanup success rate could be improved")
		}
	}

	// Check allocation distribution
	totalAllocs := uint64(0)
	maxTypeCount := uint64(0)

	for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
		allocs := mi.memoryManager.GetAllocationsByType(allocType)
		count := uint64(len(allocs))
		totalAllocs += count
		if count > maxTypeCount {
			maxTypeCount = count
		}
	}

	if totalAllocs > 0 && maxTypeCount > totalAllocs*7/10 { // 70% dominated by one type
		warnings = append(warnings, "Allocation pattern dominated by single type")
	}

	// Determine overall status
	if len(issues) > 0 {
		healthCheck["status"] = "unhealthy"
	} else if len(warnings) > 0 {
		healthCheck["status"] = "warning"
	}

	healthCheck["issues"] = issues
	healthCheck["warnings"] = warnings
	healthCheck["memory_stats"] = memStats

	return healthCheck
}

// getCurrentTimestamp returns current timestamp in a standard format
func getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02T15:04:05.000Z")
}

// SetEnabled enables or disables memory integration
func (mi *MemoryIntegration) SetEnabled(enabled bool) {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()
	mi.enabled = enabled
}

// IsEnabled returns whether memory integration is enabled
func (mi *MemoryIntegration) IsEnabled() bool {
	mi.mutex.RLock()
	defer mi.mutex.RUnlock()
	return mi.enabled
}

// GetMemoryManager returns the underlying memory manager
func (mi *MemoryIntegration) GetMemoryManager() *MemoryManager {
	mi.mutex.RLock()
	defer mi.mutex.RUnlock()
	return mi.memoryManager
}

// ------------------------------------
// 🔄 Reactive System Hooks
// ------------------------------------

// HookIntoReactiveSystem integrates memory management into the reactive system
func HookIntoReactiveSystem() {
	integration := GetGlobalMemoryIntegration()

	if !integration.enabled {
		return
	}

	// This would be called during framework initialization to set up hooks
	// In a real implementation, this would modify the reactive system to use
	// the memory-tracked versions of CreateSignal, CreateEffect, etc.

	// For now, this serves as a placeholder for the integration point
}

// ------------------------------------
// 🧪 Development Utilities
// ------------------------------------

// EnableMemoryDebugging enables detailed memory debugging
func EnableMemoryDebugging() {
	integration := GetGlobalMemoryIntegration()

	if integration.memoryManager != nil {
		integration.memoryManager.SetDebugMode(true)

		// Enable detailed tracking
		config := integration.memoryManager.config
		config.EnableStackTraces = true
		config.EnableDetailedLogging = true
		integration.memoryManager.UpdateConfig(config)
	}
}

// DisableMemoryDebugging disables detailed memory debugging
func DisableMemoryDebugging() {
	integration := GetGlobalMemoryIntegration()

	if integration.memoryManager != nil {
		integration.memoryManager.SetDebugMode(false)

		// Disable detailed tracking
		config := integration.memoryManager.config
		config.EnableStackTraces = false
		config.EnableDetailedLogging = false
		integration.memoryManager.UpdateConfig(config)
	}
}

// GetMemoryReport generates a comprehensive memory report
func GetMemoryReport() map[string]interface{} {
	integration := GetGlobalMemoryIntegration()

	if !integration.enabled || integration.memoryManager == nil {
		return map[string]interface{}{
			"status": "disabled",
		}
	}

	return integration.memoryManager.GetDetailedReport()
}

// TriggerMemoryCleanup triggers immediate memory cleanup
func TriggerMemoryCleanup() {
	integration := GetGlobalMemoryIntegration()

	if integration.enabled && integration.memoryManager != nil {
		// Force GC
		integration.memoryManager.ForceGC()

		// Clean up stale allocations
		integration.memoryManager.CleanupStaleAllocations(5 * time.Minute)

		// Trigger leak detection
		if integration.memoryManager.leakDetector != nil {
			integration.memoryManager.leakDetector.PerformLeakCheck()
		}
	}
}
