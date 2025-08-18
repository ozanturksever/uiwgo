// graceful_degradation.go
// Graceful degradation system with fallback signals and progressive enhancement

package golid

import (
	"sync"
	"sync/atomic"
	"time"

	"maragu.dev/gomponents"
)

// ------------------------------------
// 🛡️ Fallback Signal System
// ------------------------------------

// FallbackSignal provides automatic fallback when primary signal fails
type FallbackSignal[T any] struct {
	id         uint64
	primary    func() T
	fallback   func() T
	validator  func(T) bool
	state      *Signal[FallbackState]
	lastValue  T
	errorCount int
	maxErrors  int
	timeout    time.Duration
	mutex      sync.RWMutex
}

// FallbackState represents the current state of a fallback signal
type FallbackState int

const (
	FallbackPrimary FallbackState = iota
	FallbackSecondary
	FallbackFailed
)

// String returns the string representation of FallbackState
func (s FallbackState) String() string {
	switch s {
	case FallbackPrimary:
		return "Primary"
	case FallbackSecondary:
		return "Secondary"
	case FallbackFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

var fallbackSignalIdCounter uint64

// CreateFallbackSignal creates a new fallback signal
func CreateFallbackSignal[T any](primary func() T, fallback func() T, options ...FallbackOptions[T]) *FallbackSignal[T] {
	id := atomic.AddUint64(&fallbackSignalIdCounter, 1)

	stateSignal := NewSignal(FallbackPrimary)

	fs := &FallbackSignal[T]{
		id:        id,
		primary:   primary,
		fallback:  fallback,
		state:     stateSignal,
		maxErrors: 3,
		timeout:   5 * time.Second,
	}

	// Apply options
	if len(options) > 0 {
		opt := options[0]
		if opt.Validator != nil {
			fs.validator = opt.Validator
		}
		if opt.MaxErrors > 0 {
			fs.maxErrors = opt.MaxErrors
		}
		if opt.Timeout > 0 {
			fs.timeout = opt.Timeout
		}
	}

	// Initialize with primary value
	fs.tryPrimary()

	return fs
}

// FallbackOptions provides configuration for fallback signals
type FallbackOptions[T any] struct {
	Validator func(T) bool
	MaxErrors int
	Timeout   time.Duration
}

// Get retrieves the current value, using fallback if primary fails
func (fs *FallbackSignal[T]) Get() T {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	currentState := fs.state.Get()

	switch currentState {
	case FallbackPrimary:
		if value, ok := fs.tryPrimary(); ok {
			return value
		}
		// Primary failed, try fallback
		fs.switchToFallback()
		return fs.tryFallback()

	case FallbackSecondary:
		// Try primary again occasionally
		if fs.errorCount > 0 && time.Now().Unix()%10 == 0 {
			if value, ok := fs.tryPrimary(); ok {
				fs.switchToPrimary()
				return value
			}
		}
		return fs.tryFallback()

	case FallbackFailed:
		// Return last known good value
		return fs.lastValue

	default:
		return fs.tryFallback()
	}
}

// tryPrimary attempts to get value from primary source
func (fs *FallbackSignal[T]) tryPrimary() (T, bool) {
	defer func() {
		if r := recover(); r != nil {
			fs.recordError()
		}
	}()

	// Execute with timeout
	resultChan := make(chan T, 1)
	errorChan := make(chan bool, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- true
			}
		}()
		value := fs.primary()
		if fs.validator == nil || fs.validator(value) {
			resultChan <- value
		} else {
			errorChan <- true
		}
	}()

	select {
	case value := <-resultChan:
		fs.lastValue = value
		fs.errorCount = 0
		return value, true
	case <-errorChan:
		fs.recordError()
		return *new(T), false
	case <-time.After(fs.timeout):
		fs.recordError()
		return *new(T), false
	}
}

// tryFallback attempts to get value from fallback source
func (fs *FallbackSignal[T]) tryFallback() T {
	defer func() {
		if r := recover(); r != nil {
			// If fallback also fails, return last known value
		}
	}()

	value := fs.fallback()
	fs.lastValue = value
	return value
}

// recordError records an error and updates state if necessary
func (fs *FallbackSignal[T]) recordError() {
	fs.errorCount++
	if fs.errorCount >= fs.maxErrors {
		fs.switchToFallback()
	}
}

// switchToPrimary switches to primary source
func (fs *FallbackSignal[T]) switchToPrimary() {
	fs.state.Set(FallbackPrimary)
	fs.errorCount = 0
}

// switchToFallback switches to fallback source
func (fs *FallbackSignal[T]) switchToFallback() {
	fs.state.Set(FallbackSecondary)
}

// GetState returns the current fallback state
func (fs *FallbackSignal[T]) GetState() FallbackState {
	return fs.state.Get()
}

// ------------------------------------
// 🎯 Feature Detection System
// ------------------------------------

// FeatureDetector manages feature availability and capabilities
type FeatureDetector struct {
	features map[string]*Feature
	mutex    sync.RWMutex
}

// Feature represents a detectable feature or capability
type Feature struct {
	Name          string
	Available     bool
	Detector      func() bool
	Fallback      func() interface{}
	LastChecked   time.Time
	CheckInterval time.Duration
}

var globalFeatureDetector = &FeatureDetector{
	features: make(map[string]*Feature),
}

// RegisterFeature registers a new feature for detection
func RegisterFeature(name string, detector func() bool, fallback func() interface{}) {
	globalFeatureDetector.mutex.Lock()
	defer globalFeatureDetector.mutex.Unlock()

	feature := &Feature{
		Name:          name,
		Detector:      detector,
		Fallback:      fallback,
		CheckInterval: 30 * time.Second,
	}

	// Initial detection
	feature.Available = detector()
	feature.LastChecked = time.Now()

	globalFeatureDetector.features[name] = feature
}

// IsFeatureAvailable checks if a feature is available
func IsFeatureAvailable(name string) bool {
	globalFeatureDetector.mutex.RLock()
	defer globalFeatureDetector.mutex.RUnlock()

	feature, exists := globalFeatureDetector.features[name]
	if !exists {
		return false
	}

	// Re-check if enough time has passed
	if time.Since(feature.LastChecked) > feature.CheckInterval {
		globalFeatureDetector.mutex.RUnlock()
		globalFeatureDetector.mutex.Lock()
		feature.Available = feature.Detector()
		feature.LastChecked = time.Now()
		globalFeatureDetector.mutex.Unlock()
		globalFeatureDetector.mutex.RLock()
	}

	return feature.Available
}

// GetFeatureFallback returns the fallback for a feature
func GetFeatureFallback(name string) interface{} {
	globalFeatureDetector.mutex.RLock()
	defer globalFeatureDetector.mutex.RUnlock()

	feature, exists := globalFeatureDetector.features[name]
	if !exists || feature.Fallback == nil {
		return nil
	}

	return feature.Fallback()
}

// ------------------------------------
// 🏗️ Progressive Enhancement
// ------------------------------------

// ProgressiveComponent creates a component with progressive enhancement
type ProgressiveComponent struct {
	baseline func() gomponents.Node
	enhanced func() gomponents.Node
	features []string
	fallback func() gomponents.Node
	detector *FeatureDetector
}

// CreateProgressiveComponent creates a new progressive component
func CreateProgressiveComponent(
	baseline func() gomponents.Node,
	enhanced func() gomponents.Node,
	requiredFeatures []string,
) *ProgressiveComponent {
	return &ProgressiveComponent{
		baseline: baseline,
		enhanced: enhanced,
		features: requiredFeatures,
		detector: globalFeatureDetector,
	}
}

// Render renders the component based on feature availability
func (pc *ProgressiveComponent) Render() gomponents.Node {
	// Check if all required features are available
	allAvailable := true
	for _, feature := range pc.features {
		if !IsFeatureAvailable(feature) {
			allAvailable = false
			break
		}
	}

	if allAvailable && pc.enhanced != nil {
		// Try enhanced version with fallback
		defer func() {
			if r := recover(); r != nil {
				// Enhanced version failed, use baseline
			}
		}()
		return pc.enhanced()
	}

	// Use baseline version
	if pc.baseline != nil {
		return pc.baseline()
	}

	// Last resort fallback
	if pc.fallback != nil {
		return pc.fallback()
	}

	return gomponents.Text("Content unavailable")
}

// WithFallback sets a fallback renderer
func (pc *ProgressiveComponent) WithFallback(fallback func() gomponents.Node) *ProgressiveComponent {
	pc.fallback = fallback
	return pc
}

// ------------------------------------
// 🔧 Degraded Mode Operations
// ------------------------------------

// DegradedModeManager manages system-wide degraded mode
type DegradedModeManager struct {
	enabled      bool
	level        DegradationLevel
	features     map[string]bool
	fallbacks    map[string]func() interface{}
	onModeChange func(DegradationLevel)
	mutex        sync.RWMutex
}

// DegradationLevel represents the level of system degradation
type DegradationLevel int

const (
	NormalMode DegradationLevel = iota
	LightDegradation
	ModerateDegradation
	SevereDegradation
	EmergencyMode
)

// String returns the string representation of DegradationLevel
func (d DegradationLevel) String() string {
	switch d {
	case NormalMode:
		return "Normal"
	case LightDegradation:
		return "Light"
	case ModerateDegradation:
		return "Moderate"
	case SevereDegradation:
		return "Severe"
	case EmergencyMode:
		return "Emergency"
	default:
		return "Unknown"
	}
}

var globalDegradedModeManager = &DegradedModeManager{
	features:  make(map[string]bool),
	fallbacks: make(map[string]func() interface{}),
}

// EnableDegradedMode enables degraded mode at the specified level
func EnableDegradedMode(level DegradationLevel) {
	globalDegradedModeManager.mutex.Lock()
	defer globalDegradedModeManager.mutex.Unlock()

	globalDegradedModeManager.enabled = true
	globalDegradedModeManager.level = level

	// Disable features based on degradation level
	switch level {
	case LightDegradation:
		globalDegradedModeManager.disableFeature("animations")
		globalDegradedModeManager.disableFeature("transitions")
	case ModerateDegradation:
		globalDegradedModeManager.disableFeature("animations")
		globalDegradedModeManager.disableFeature("transitions")
		globalDegradedModeManager.disableFeature("lazy-loading")
	case SevereDegradation:
		globalDegradedModeManager.disableFeature("animations")
		globalDegradedModeManager.disableFeature("transitions")
		globalDegradedModeManager.disableFeature("lazy-loading")
		globalDegradedModeManager.disableFeature("reactive-updates")
	case EmergencyMode:
		// Disable all non-essential features
		for feature := range globalDegradedModeManager.features {
			if feature != "core-rendering" {
				globalDegradedModeManager.disableFeature(feature)
			}
		}
	}

	if globalDegradedModeManager.onModeChange != nil {
		go globalDegradedModeManager.onModeChange(level)
	}
}

// DisableDegradedMode disables degraded mode
func DisableDegradedMode() {
	globalDegradedModeManager.mutex.Lock()
	defer globalDegradedModeManager.mutex.Unlock()

	globalDegradedModeManager.enabled = false
	globalDegradedModeManager.level = NormalMode

	// Re-enable all features
	for feature := range globalDegradedModeManager.features {
		globalDegradedModeManager.features[feature] = true
	}

	if globalDegradedModeManager.onModeChange != nil {
		go globalDegradedModeManager.onModeChange(NormalMode)
	}
}

// IsInDegradedMode checks if the system is in degraded mode
func IsInDegradedMode() bool {
	globalDegradedModeManager.mutex.RLock()
	defer globalDegradedModeManager.mutex.RUnlock()
	return globalDegradedModeManager.enabled
}

// GetDegradationLevel returns the current degradation level
func GetDegradationLevel() DegradationLevel {
	globalDegradedModeManager.mutex.RLock()
	defer globalDegradedModeManager.mutex.RUnlock()
	return globalDegradedModeManager.level
}

// disableFeature disables a specific feature
func (dmm *DegradedModeManager) disableFeature(feature string) {
	dmm.features[feature] = false
}

// IsFeatureEnabled checks if a feature is enabled in current mode
func IsFeatureEnabled(feature string) bool {
	globalDegradedModeManager.mutex.RLock()
	defer globalDegradedModeManager.mutex.RUnlock()

	if !globalDegradedModeManager.enabled {
		return true // All features enabled in normal mode
	}

	enabled, exists := globalDegradedModeManager.features[feature]
	if !exists {
		return true // Unknown features are enabled by default
	}

	return enabled
}

// ------------------------------------
// 🎨 Graceful Component Wrappers
// ------------------------------------

// GracefulComponent wraps a component with graceful degradation
func GracefulComponent(
	component func() gomponents.Node,
	fallback func() gomponents.Node,
	requiredFeatures ...string,
) gomponents.Node {
	// Check if we're in degraded mode
	if IsInDegradedMode() {
		// Check if required features are available
		for _, feature := range requiredFeatures {
			if !IsFeatureEnabled(feature) {
				return fallback()
			}
		}
	}

	// Try to render the component with error handling
	defer func() {
		if r := recover(); r != nil {
			// Component failed, use fallback
		}
	}()

	return component()
}

// SafeRender renders a component with automatic fallback on error
func SafeRender(component func() gomponents.Node, fallback gomponents.Node) gomponents.Node {
	defer func() {
		if r := recover(); r != nil {
			// Return fallback on panic
		}
	}()

	result := component()
	if result == nil {
		return fallback
	}

	return result
}

// ConditionalFeatureRender renders different content based on feature availability
func ConditionalFeatureRender(
	feature string,
	enhanced func() gomponents.Node,
	basic func() gomponents.Node,
) gomponents.Node {
	if IsFeatureAvailable(feature) && IsFeatureEnabled(feature) {
		return SafeRender(enhanced, basic())
	}
	return basic()
}
