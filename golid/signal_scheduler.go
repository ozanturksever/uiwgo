// signal_scheduler.go
// Batched update scheduling and cascade prevention

package golid

import (
	"container/heap"
	"sync"
	"time"
)

// ------------------------------------
// 📅 Scheduler Implementation
// ------------------------------------

var (
	globalScheduler    *Scheduler
	schedulerOnce      sync.Once
	schedulerInitMutex sync.Mutex
)

// Scheduler manages batched updates and prevents infinite cascades
type Scheduler struct {
	queue      *PriorityQueue
	microtask  chan *ScheduledTask
	running    bool
	batchDepth int
	mutex      sync.Mutex
	stopChan   chan bool
	maxDepth   int
}

// ScheduledTask represents a task to be executed by the scheduler
type ScheduledTask struct {
	priority    Priority
	computation *Computation
	timestamp   int64
	index       int // for heap implementation
}

// PriorityQueue implements a priority queue for scheduled tasks
type PriorityQueue []*ScheduledTask

// getScheduler returns the global scheduler instance
func getScheduler() *Scheduler {
	schedulerOnce.Do(func() {
		globalScheduler = &Scheduler{
			queue:     &PriorityQueue{},
			microtask: make(chan *ScheduledTask, 1000),
			maxDepth:  50, // Prevent infinite cascades
			stopChan:  make(chan bool, 1),
		}
		heap.Init(globalScheduler.queue)
		go globalScheduler.run()
	})
	return globalScheduler
}

// ------------------------------------
// 🔄 Priority Queue Implementation
// ------------------------------------

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Higher priority (lower number) comes first
	if pq[i].priority != pq[j].priority {
		return pq[i].priority < pq[j].priority
	}
	// If same priority, earlier timestamp comes first
	return pq[i].timestamp < pq[j].timestamp
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	task := x.(*ScheduledTask)
	task.index = n
	*pq = append(*pq, task)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	task := old[n-1]
	old[n-1] = nil
	task.index = -1
	*pq = old[0 : n-1]
	return task
}

// ------------------------------------
// 📅 Scheduler Methods
// ------------------------------------

// schedule adds a task to the scheduler queue
func (s *Scheduler) schedule(task *ScheduledTask) {
	if task.computation == nil {
		return
	}

	task.timestamp = time.Now().UnixNano()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if we're at max cascade depth
	if s.batchDepth >= s.maxDepth {
		return
	}

	// Add to microtask queue for immediate processing
	select {
	case s.microtask <- task:
	default:
		// If microtask queue is full, add to priority queue
		heap.Push(s.queue, task)
	}
}

// batch runs a function within a batched update context
func (s *Scheduler) batch(fn func()) {
	s.mutex.Lock()
	s.batchDepth++
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		s.batchDepth--
		shouldFlush := s.batchDepth == 0
		s.mutex.Unlock()

		if shouldFlush {
			s.flush()
		}
	}()

	fn()
}

// flush processes all pending tasks
func (s *Scheduler) flush() {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return
	}
	s.running = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		s.running = false
		s.mutex.Unlock()
	}()

	// Process microtasks first
	for {
		select {
		case task := <-s.microtask:
			s.executeTask(task)
		default:
			goto processQueue
		}
	}

processQueue:
	// Process priority queue
	s.mutex.Lock()
	for s.queue.Len() > 0 {
		task := heap.Pop(s.queue).(*ScheduledTask)
		s.mutex.Unlock()
		s.executeTask(task)
		s.mutex.Lock()
	}
	s.mutex.Unlock()
}

// executeTask executes a single scheduled task
func (s *Scheduler) executeTask(task *ScheduledTask) {
	if task.computation == nil {
		return
	}

	// Check if computation is still dirty (might have been cleaned by another task)
	task.computation.mutex.RLock()
	state := task.computation.state
	task.computation.mutex.RUnlock()

	if state == Dirty {
		task.computation.run()
	}
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	// Use a more efficient timer for WASM environment
	ticker := time.NewTicker(4 * time.Millisecond) // Higher frequency for better responsiveness
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Only flush if there's work to do
			s.mutex.Lock()
			hasWork := s.queue.Len() > 0 || len(s.microtask) > 0
			s.mutex.Unlock()

			if hasWork {
				s.flush()
			}
		case <-s.stopChan:
			return
		}
	}
}

// stop stops the scheduler
func (s *Scheduler) stop() {
	select {
	case s.stopChan <- true:
	default:
	}
}

// ------------------------------------
// 🛡️ Cascade Prevention
// ------------------------------------

// CascadeGuard prevents infinite update cascades
type CascadeGuard struct {
	depth    int
	maxDepth int
	visited  map[uint64]bool
	mutex    sync.RWMutex
}

var globalCascadeGuard = &CascadeGuard{
	maxDepth: 50,
	visited:  make(map[uint64]bool),
}

// Enter attempts to enter a computation, returns false if cascade limit reached
func (g *CascadeGuard) Enter(computationId uint64) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.depth >= g.maxDepth {
		return false
	}

	if g.visited[computationId] {
		return false // Already processing this computation
	}

	g.depth++
	g.visited[computationId] = true
	return true
}

// Exit exits a computation
func (g *CascadeGuard) Exit(computationId uint64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.depth--
	delete(g.visited, computationId)
}

// Reset resets the cascade guard (for testing)
func (g *CascadeGuard) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.depth = 0
	g.visited = make(map[uint64]bool)
}

// ------------------------------------
// 📊 Scheduler Statistics
// ------------------------------------

// SchedulerStats provides statistics about the scheduler
type SchedulerStats struct {
	QueueSize     int
	MicrotaskSize int
	BatchDepth    int
	Running       bool
}

// GetStats returns current scheduler statistics
func (s *Scheduler) GetStats() SchedulerStats {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return SchedulerStats{
		QueueSize:     s.queue.Len(),
		MicrotaskSize: len(s.microtask),
		BatchDepth:    s.batchDepth,
		Running:       s.running,
	}
}

// ------------------------------------
// 🧪 Testing Utilities
// ------------------------------------

// ResetScheduler resets the global scheduler (for testing)
func ResetScheduler() {
	schedulerInitMutex.Lock()
	defer schedulerInitMutex.Unlock()

	if globalScheduler != nil {
		globalScheduler.stop()
	}

	// Reset the once to allow re-initialization
	schedulerOnce = sync.Once{}
	globalScheduler = nil
	globalCascadeGuard.Reset()
}

// FlushScheduler forces immediate execution of all pending tasks
func FlushScheduler() {
	if globalScheduler != nil {
		globalScheduler.flush()
	}
}
