// resource_registry.go
// Centralized resource registry for comprehensive resource management

package golid

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------
// 📋 Resource Registry Core
// ------------------------------------

var resourceRegistryIdCounter uint64

// ResourceRegistry provides centralized registration and tracking of all reactive resources
type ResourceRegistry struct {
	id            uint64
	memoryManager *MemoryManager

	// Resource tracking
	allocations  map[uint64]*Allocation
	dependencies map[uint64][]uint64 // allocation ID -> dependent allocation IDs
	owners       map[uint64]*Owner   // owner ID -> owner

	// Indexing for fast lookups
	allocationsByType  map[AllocationType]map[uint64]*Allocation
	allocationsByOwner map[uint64]map[uint64]*Allocation

	// Configuration
	config RegistryConfig

	// State
	enabled  bool
	indexing bool

	// Statistics
	stats ResourceRegistryStats

	// Background maintenance
	ctx               context.Context
	cancel            context.CancelFunc
	maintenanceTicker *time.Ticker
	stopMaintenance   chan bool

	// Synchronization
	mutex sync.RWMutex
}

// RegistryConfig provides configuration for resource registry
type RegistryConfig struct {
	MaintenanceInterval      time.Duration `json:"maintenance_interval"`
	EnableDependencyTracking bool          `json:"enable_dependency_tracking"`
	EnableIndexing           bool          `json:"enable_indexing"`
	EnableStatistics         bool          `json:"enable_statistics"`
	MaxHistorySize           int           `json:"max_history_size"`
	CleanupThreshold         int           `json:"cleanup_threshold"`
}

// ResourceRegistryStats tracks resource registry statistics
type ResourceRegistryStats struct {
	TotalRegistrations     uint64        `json:"total_registrations"`
	ActiveAllocations      uint64        `json:"active_allocations"`
	DisposedAllocations    uint64        `json:"disposed_allocations"`
	DependencyCount        uint64        `json:"dependency_count"`
	OwnerCount             uint64        `json:"owner_count"`
	IndexRebuildCount      uint64        `json:"index_rebuild_count"`
	MaintenanceRuns        uint64        `json:"maintenance_runs"`
	AverageMaintenanceTime time.Duration `json:"average_maintenance_time"`
	LastMaintenanceTime    time.Time     `json:"last_maintenance_time"`
	StartTime              time.Time     `json:"start_time"`
	mutex                  sync.RWMutex
}

// ------------------------------------
// 🏗️ Resource Registry Creation
// ------------------------------------

// NewResourceRegistry creates a new resource registry
func NewResourceRegistry(memoryManager *MemoryManager) *ResourceRegistry {
	ctx, cancel := context.WithCancel(context.Background())

	config := DefaultRegistryConfig()

	rr := &ResourceRegistry{
		id:                 atomic.AddUint64(&resourceRegistryIdCounter, 1),
		memoryManager:      memoryManager,
		allocations:        make(map[uint64]*Allocation),
		dependencies:       make(map[uint64][]uint64),
		owners:             make(map[uint64]*Owner),
		allocationsByType:  make(map[AllocationType]map[uint64]*Allocation),
		allocationsByOwner: make(map[uint64]map[uint64]*Allocation),
		config:             config,
		enabled:            true,
		indexing:           config.EnableIndexing,
		ctx:                ctx,
		cancel:             cancel,
		maintenanceTicker:  time.NewTicker(config.MaintenanceInterval),
		stopMaintenance:    make(chan bool, 1),
		stats: ResourceRegistryStats{
			StartTime: time.Now(),
		},
	}

	// Initialize type indexes
	for allocType := AllocSignal; allocType <= AllocCustom; allocType++ {
		rr.allocationsByType[allocType] = make(map[uint64]*Allocation)
	}

	// Start background maintenance
	go rr.backgroundMaintenance()

	return rr
}

// DefaultRegistryConfig returns default registry configuration
func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		MaintenanceInterval:      30 * time.Second,
		EnableDependencyTracking: true,
		EnableIndexing:           true,
		EnableStatistics:         true,
		MaxHistorySize:           10000,
		CleanupThreshold:         1000,
	}
}

// ------------------------------------
// 📝 Resource Registration
// ------------------------------------

// RegisterAllocation registers a new allocation in the registry
func (rr *ResourceRegistry) RegisterAllocation(alloc *Allocation) {
	if !rr.enabled || alloc == nil {
		return
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	// Register allocation
	rr.allocations[alloc.ID] = alloc

	// Update indexes if enabled
	if rr.indexing {
		rr.updateIndexes(alloc, true)
	}

	// Register owner if not already registered
	if alloc.Owner != nil {
		rr.registerOwner(alloc.Owner)
	}

	// Update statistics
	rr.updateStats(func(stats *ResourceRegistryStats) {
		atomic.AddUint64(&stats.TotalRegistrations, 1)
		atomic.AddUint64(&stats.ActiveAllocations, 1)
	})
}

// UnregisterAllocation removes an allocation from the registry
func (rr *ResourceRegistry) UnregisterAllocation(allocID uint64) bool {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	alloc, exists := rr.allocations[allocID]
	if !exists {
		return false
	}

	// Update indexes if enabled
	if rr.indexing {
		rr.updateIndexes(alloc, false)
	}

	// Remove dependencies
	if rr.config.EnableDependencyTracking {
		rr.removeDependencies(allocID)
	}

	// Remove from main registry
	delete(rr.allocations, allocID)

	// Update statistics
	rr.updateStats(func(stats *ResourceRegistryStats) {
		atomic.AddUint64(&stats.ActiveAllocations, ^uint64(0)) // Decrement
		atomic.AddUint64(&stats.DisposedAllocations, 1)
	})

	return true
}

// registerOwner registers an owner in the registry
func (rr *ResourceRegistry) registerOwner(owner *Owner) {
	if owner == nil {
		return
	}

	if _, exists := rr.owners[owner.id]; !exists {
		rr.owners[owner.id] = owner
		rr.allocationsByOwner[owner.id] = make(map[uint64]*Allocation)

		rr.updateStats(func(stats *ResourceRegistryStats) {
			atomic.AddUint64(&stats.OwnerCount, 1)
		})
	}
}

// updateIndexes updates the registry indexes
func (rr *ResourceRegistry) updateIndexes(alloc *Allocation, add bool) {
	if add {
		// Add to type index
		rr.allocationsByType[alloc.Type][alloc.ID] = alloc

		// Add to owner index
		if alloc.Owner != nil {
			if ownerAllocs, exists := rr.allocationsByOwner[alloc.Owner.id]; exists {
				ownerAllocs[alloc.ID] = alloc
			}
		}
	} else {
		// Remove from type index
		delete(rr.allocationsByType[alloc.Type], alloc.ID)

		// Remove from owner index
		if alloc.Owner != nil {
			if ownerAllocs, exists := rr.allocationsByOwner[alloc.Owner.id]; exists {
				delete(ownerAllocs, alloc.ID)
			}
		}
	}
}

// ------------------------------------
// 🔗 Dependency Tracking
// ------------------------------------

// AddDependency adds a dependency relationship between allocations
func (rr *ResourceRegistry) AddDependency(allocID, dependentID uint64) {
	if !rr.config.EnableDependencyTracking {
		return
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	// Check if both allocations exist
	if _, exists := rr.allocations[allocID]; !exists {
		return
	}
	if _, exists := rr.allocations[dependentID]; !exists {
		return
	}

	// Add dependency
	if deps, exists := rr.dependencies[allocID]; exists {
		// Check if dependency already exists
		for _, dep := range deps {
			if dep == dependentID {
				return
			}
		}
		rr.dependencies[allocID] = append(deps, dependentID)
	} else {
		rr.dependencies[allocID] = []uint64{dependentID}
	}

	rr.updateStats(func(stats *ResourceRegistryStats) {
		atomic.AddUint64(&stats.DependencyCount, 1)
	})
}

// RemoveDependency removes a dependency relationship
func (rr *ResourceRegistry) RemoveDependency(allocID, dependentID uint64) {
	if !rr.config.EnableDependencyTracking {
		return
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if deps, exists := rr.dependencies[allocID]; exists {
		for i, dep := range deps {
			if dep == dependentID {
				rr.dependencies[allocID] = append(deps[:i], deps[i+1:]...)
				rr.updateStats(func(stats *ResourceRegistryStats) {
					atomic.AddUint64(&stats.DependencyCount, ^uint64(0)) // Decrement
				})
				break
			}
		}

		// Remove empty dependency list
		if len(rr.dependencies[allocID]) == 0 {
			delete(rr.dependencies, allocID)
		}
	}
}

// removeDependencies removes all dependencies for an allocation
func (rr *ResourceRegistry) removeDependencies(allocID uint64) {
	if deps, exists := rr.dependencies[allocID]; exists {
		count := len(deps)
		delete(rr.dependencies, allocID)

		rr.updateStats(func(stats *ResourceRegistryStats) {
			atomic.AddUint64(&stats.DependencyCount, ^uint64(count-1)) // Subtract count
		})
	}

	// Also remove this allocation as a dependent of others
	for id, deps := range rr.dependencies {
		for i, dep := range deps {
			if dep == allocID {
				rr.dependencies[id] = append(deps[:i], deps[i+1:]...)
				rr.updateStats(func(stats *ResourceRegistryStats) {
					atomic.AddUint64(&stats.DependencyCount, ^uint64(0)) // Decrement
				})
				break
			}
		}

		// Remove empty dependency list
		if len(rr.dependencies[id]) == 0 {
			delete(rr.dependencies, id)
		}
	}
}

// GetDependencies returns all dependencies for an allocation
func (rr *ResourceRegistry) GetDependencies(allocID uint64) []uint64 {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	if deps, exists := rr.dependencies[allocID]; exists {
		result := make([]uint64, len(deps))
		copy(result, deps)
		return result
	}

	return nil
}

// GetDependents returns all allocations that depend on the given allocation
func (rr *ResourceRegistry) GetDependents(allocID uint64) []uint64 {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	var dependents []uint64

	for id, deps := range rr.dependencies {
		for _, dep := range deps {
			if dep == allocID {
				dependents = append(dependents, id)
				break
			}
		}
	}

	return dependents
}

// ------------------------------------
// 🔍 Resource Lookup and Querying
// ------------------------------------

// GetAllocation retrieves an allocation by ID
func (rr *ResourceRegistry) GetAllocation(allocID uint64) (*Allocation, bool) {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	alloc, exists := rr.allocations[allocID]
	return alloc, exists
}

// GetAllocationsByType returns all allocations of a specific type
func (rr *ResourceRegistry) GetAllocationsByType(allocType AllocationType) []*Allocation {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	var result []*Allocation

	if rr.indexing {
		// Use index for fast lookup
		if typeAllocs, exists := rr.allocationsByType[allocType]; exists {
			for _, alloc := range typeAllocs {
				if !alloc.Disposed {
					result = append(result, alloc)
				}
			}
		}
	} else {
		// Scan all allocations
		for _, alloc := range rr.allocations {
			if alloc.Type == allocType && !alloc.Disposed {
				result = append(result, alloc)
			}
		}
	}

	return result
}

// GetAllocationsByOwner returns all allocations owned by a specific owner
func (rr *ResourceRegistry) GetAllocationsByOwner(ownerID uint64) []*Allocation {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	var result []*Allocation

	if rr.indexing {
		// Use index for fast lookup
		if ownerAllocs, exists := rr.allocationsByOwner[ownerID]; exists {
			for _, alloc := range ownerAllocs {
				if !alloc.Disposed {
					result = append(result, alloc)
				}
			}
		}
	} else {
		// Scan all allocations
		for _, alloc := range rr.allocations {
			if alloc.Owner != nil && alloc.Owner.id == ownerID && !alloc.Disposed {
				result = append(result, alloc)
			}
		}
	}

	return result
}

// GetAllAllocations returns all active allocations
func (rr *ResourceRegistry) GetAllAllocations() []*Allocation {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	var result []*Allocation
	for _, alloc := range rr.allocations {
		if !alloc.Disposed {
			result = append(result, alloc)
		}
	}

	return result
}

// GetOwner retrieves an owner by ID
func (rr *ResourceRegistry) GetOwner(ownerID uint64) (*Owner, bool) {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	owner, exists := rr.owners[ownerID]
	return owner, exists
}

// GetAllOwners returns all registered owners
func (rr *ResourceRegistry) GetAllOwners() []*Owner {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	var result []*Owner
	for _, owner := range rr.owners {
		result = append(result, owner)
	}

	return result
}

// ------------------------------------
// 🔄 Background Maintenance
// ------------------------------------

// backgroundMaintenance runs continuous registry maintenance
func (rr *ResourceRegistry) backgroundMaintenance() {
	for {
		select {
		case <-rr.ctx.Done():
			return
		case <-rr.stopMaintenance:
			return
		case <-rr.maintenanceTicker.C:
			rr.performMaintenance()
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (rr *ResourceRegistry) performMaintenance() {
	if !rr.enabled {
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		rr.updateStats(func(stats *ResourceRegistryStats) {
			atomic.AddUint64(&stats.MaintenanceRuns, 1)
			stats.LastMaintenanceTime = time.Now()

			// Update average maintenance time
			runs := atomic.LoadUint64(&stats.MaintenanceRuns)
			if runs > 0 {
				stats.AverageMaintenanceTime = time.Duration(
					(int64(stats.AverageMaintenanceTime)*int64(runs-1) + int64(duration)) / int64(runs),
				)
			}
		})
	}()

	// Clean up disposed allocations
	rr.cleanupDisposedAllocations()

	// Rebuild indexes if needed
	if rr.indexing {
		rr.rebuildIndexesIfNeeded()
	}

	// Clean up orphaned dependencies
	if rr.config.EnableDependencyTracking {
		rr.cleanupOrphanedDependencies()
	}

	// Clean up disposed owners
	rr.cleanupDisposedOwners()
}

// cleanupDisposedAllocations removes disposed allocations from the registry
func (rr *ResourceRegistry) cleanupDisposedAllocations() {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	var toRemove []uint64

	for id, alloc := range rr.allocations {
		if alloc.Disposed {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		alloc := rr.allocations[id]

		// Update indexes
		if rr.indexing {
			rr.updateIndexes(alloc, false)
		}

		// Remove dependencies
		if rr.config.EnableDependencyTracking {
			rr.removeDependencies(id)
		}

		delete(rr.allocations, id)
	}
}

// rebuildIndexesIfNeeded rebuilds indexes if they become inconsistent
func (rr *ResourceRegistry) rebuildIndexesIfNeeded() {
	// Simple heuristic: rebuild if index size differs significantly from allocation count
	rr.mutex.RLock()
	totalIndexed := 0
	for _, typeIndex := range rr.allocationsByType {
		totalIndexed += len(typeIndex)
	}
	totalAllocations := len(rr.allocations)
	rr.mutex.RUnlock()

	// Rebuild if difference is more than 10%
	if abs(totalIndexed-totalAllocations) > totalAllocations/10 {
		rr.rebuildIndexes()
	}
}

// rebuildIndexes rebuilds all indexes from scratch
func (rr *ResourceRegistry) rebuildIndexes() {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	// Clear existing indexes
	for allocType := range rr.allocationsByType {
		rr.allocationsByType[allocType] = make(map[uint64]*Allocation)
	}

	for ownerID := range rr.allocationsByOwner {
		rr.allocationsByOwner[ownerID] = make(map[uint64]*Allocation)
	}

	// Rebuild indexes
	for _, alloc := range rr.allocations {
		if !alloc.Disposed {
			rr.updateIndexes(alloc, true)
		}
	}

	rr.updateStats(func(stats *ResourceRegistryStats) {
		atomic.AddUint64(&stats.IndexRebuildCount, 1)
	})
}

// cleanupOrphanedDependencies removes dependencies to non-existent allocations
func (rr *ResourceRegistry) cleanupOrphanedDependencies() {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	for allocID, deps := range rr.dependencies {
		var validDeps []uint64

		for _, depID := range deps {
			if _, exists := rr.allocations[depID]; exists {
				validDeps = append(validDeps, depID)
			}
		}

		if len(validDeps) != len(deps) {
			if len(validDeps) == 0 {
				delete(rr.dependencies, allocID)
			} else {
				rr.dependencies[allocID] = validDeps
			}
		}
	}
}

// cleanupDisposedOwners removes disposed owners from the registry
func (rr *ResourceRegistry) cleanupDisposedOwners() {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	var toRemove []uint64

	for id, owner := range rr.owners {
		owner.mutex.RLock()
		disposed := owner.disposed
		owner.mutex.RUnlock()

		if disposed {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		delete(rr.owners, id)
		delete(rr.allocationsByOwner, id)

		rr.updateStats(func(stats *ResourceRegistryStats) {
			atomic.AddUint64(&stats.OwnerCount, ^uint64(0)) // Decrement
		})
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ------------------------------------
// 📊 Statistics and Reporting
// ------------------------------------

// GetStats returns current registry statistics
func (rr *ResourceRegistry) GetStats() ResourceRegistryStats {
	rr.stats.mutex.RLock()
	defer rr.stats.mutex.RUnlock()
	return rr.stats
}

// GetReport returns a comprehensive registry report
func (rr *ResourceRegistry) GetReport() map[string]interface{} {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	report := map[string]interface{}{
		"registry_id":        rr.id,
		"enabled":            rr.enabled,
		"indexing":           rr.indexing,
		"config":             rr.config,
		"stats":              rr.GetStats(),
		"total_allocations":  len(rr.allocations),
		"total_owners":       len(rr.owners),
		"total_dependencies": len(rr.dependencies),
	}

	// Add allocation breakdown by type
	typeBreakdown := make(map[string]int)
	for allocType, typeIndex := range rr.allocationsByType {
		typeBreakdown[allocType.String()] = len(typeIndex)
	}
	report["allocations_by_type"] = typeBreakdown

	// Add owner breakdown
	ownerBreakdown := make(map[string]int)
	for ownerID, ownerAllocs := range rr.allocationsByOwner {
		ownerBreakdown[string(rune(ownerID))] = len(ownerAllocs)
	}
	report["allocations_by_owner"] = ownerBreakdown

	return report
}

// updateStats safely updates registry statistics
func (rr *ResourceRegistry) updateStats(fn func(*ResourceRegistryStats)) {
	rr.stats.mutex.Lock()
	defer rr.stats.mutex.Unlock()
	fn(&rr.stats)
}

// ------------------------------------
// 🛠️ Configuration and Control
// ------------------------------------

// SetEnabled enables or disables the registry
func (rr *ResourceRegistry) SetEnabled(enabled bool) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()
	rr.enabled = enabled
}

// SetIndexing enables or disables indexing
func (rr *ResourceRegistry) SetIndexing(indexing bool) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if indexing && !rr.indexing {
		// Enabling indexing - rebuild indexes
		rr.indexing = true
		rr.rebuildIndexes()
	} else if !indexing && rr.indexing {
		// Disabling indexing - clear indexes
		rr.indexing = false
		for allocType := range rr.allocationsByType {
			rr.allocationsByType[allocType] = make(map[uint64]*Allocation)
		}
		for ownerID := range rr.allocationsByOwner {
			rr.allocationsByOwner[ownerID] = make(map[uint64]*Allocation)
		}
	}
}

// UpdateConfig updates the registry configuration
func (rr *ResourceRegistry) UpdateConfig(config RegistryConfig) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	rr.config = config

	// Update indexing if changed
	if config.EnableIndexing != rr.indexing {
		rr.SetIndexing(config.EnableIndexing)
	}

	// Update ticker interval if changed
	if rr.maintenanceTicker != nil {
		rr.maintenanceTicker.Stop()
		rr.maintenanceTicker = time.NewTicker(config.MaintenanceInterval)
	}
}

// ------------------------------------
// 🧹 Disposal
// ------------------------------------

// Dispose cleans up the resource registry and all its resources
func (rr *ResourceRegistry) Dispose() {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	// Stop background maintenance
	if rr.cancel != nil {
		rr.cancel()
	}

	if rr.maintenanceTicker != nil {
		rr.maintenanceTicker.Stop()
	}

	select {
	case rr.stopMaintenance <- true:
	default:
	}

	// Clear all data
	rr.allocations = make(map[uint64]*Allocation)
	rr.dependencies = make(map[uint64][]uint64)
	rr.owners = make(map[uint64]*Owner)

	// Clear indexes
	for allocType := range rr.allocationsByType {
		rr.allocationsByType[allocType] = make(map[uint64]*Allocation)
	}
	for ownerID := range rr.allocationsByOwner {
		rr.allocationsByOwner[ownerID] = make(map[uint64]*Allocation)
	}

	rr.enabled = false
	rr.indexing = false
}
