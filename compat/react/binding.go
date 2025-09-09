//go:build js && wasm

package react

import (
	"fmt"
	"reflect"
	"syscall/js"

	"github.com/ozanturksever/logutil"
	"github.com/ozanturksever/uiwgo/reactivity"
)

// BindingOptions configures how a component binding behaves
type BindingOptions struct {
	// ComponentID is the React component instance ID
	ComponentID string

	// ComponentName is the React component name for registration
	ComponentName string

	// Props is a function that returns the current props based on signal values
	Props func() map[string]interface{}

	// OnError is called when binding operations fail
	OnError func(error)

	// AutoRender determines if the component should auto-render on first bind
	AutoRender bool
}

// ComponentBinding represents a bound React component with reactive updates
type ComponentBinding struct {
	options   BindingOptions
	effect    reactivity.Effect
	isActive  bool
	lastProps map[string]interface{}
}

// Bind creates a reactive binding between Go signals and a React component
// The component will automatically update when any signals used in the Props function change
func Bind(options BindingOptions) (*ComponentBinding, error) {
	if options.ComponentID == "" {
		return nil, NewReactBridgeError("component ID is required", ErrorTypeInvalidArgument, ErrorSeverityError, nil)
	}

	if options.ComponentName == "" {
		return nil, NewReactBridgeError("component name is required", ErrorTypeInvalidArgument, ErrorSeverityError, nil)
	}

	if options.Props == nil {
		return nil, NewReactBridgeError("props function is required", ErrorTypeInvalidArgument, ErrorSeverityError, nil)
	}

	binding := &ComponentBinding{
		options:  options,
		isActive: false,
	}

	// Create an effect that will run whenever signals change
	effect := reactivity.CreateEffect(func() {
		if !binding.isActive {
			return
		}

		// Get current props from the function
		currentProps := options.Props()

		// Check if props have actually changed to avoid unnecessary updates
		if binding.propsEqual(currentProps, binding.lastProps) {
			return
		}

		binding.lastProps = currentProps

		// Update the React component
		err := binding.updateComponent(currentProps)
		if err != nil && options.OnError != nil {
			options.OnError(err)
		}
	})

	binding.effect = effect

	// Auto-render if requested
	if options.AutoRender {
		err := binding.Start()
		if err != nil {
			return nil, err
		}
	}

	return binding, nil
}

// Start activates the binding and renders the component
func (b *ComponentBinding) Start() error {
	if b.isActive {
		return nil
	}

	// Get initial props
	initialProps := b.options.Props()
	b.lastProps = initialProps

	// Render the component
	_, err := Render(b.options.ComponentName, initialProps, nil)
	if err != nil {
		return NewReactBridgeError("failed to render component", ErrorTypeRenderFailure, ErrorSeverityError, map[string]interface{}{
			"componentName": b.options.ComponentName,
			"componentID":   b.options.ComponentID,
			"error":         err.Error(),
		})
	}

	b.isActive = true
	logutil.Logf("Started reactive binding for component %s (ID: %s)", b.options.ComponentName, b.options.ComponentID)

	return nil
}

// Stop deactivates the binding and unmounts the component
func (b *ComponentBinding) Stop() error {
	if !b.isActive {
		return nil
	}

	b.isActive = false

	// Dispose the effect to stop reactive updates
	if b.effect != nil {
		b.effect.Dispose()
	}

	// Unmount the component
	err := Unmount(ComponentID(b.options.ComponentID))
	if err != nil {
		return NewReactBridgeError("failed to unmount component", ErrorTypeUnmountFailure, ErrorSeverityError, map[string]interface{}{
			"componentID": b.options.ComponentID,
			"error":       err.Error(),
		})
	}

	logutil.Logf("Stopped reactive binding for component %s (ID: %s)", b.options.ComponentName, b.options.ComponentID)

	return nil
}

// Update manually triggers a component update with current props
func (b *ComponentBinding) Update() error {
	if !b.isActive {
		return NewReactBridgeError("binding is not active", ErrorTypeInvalidState, ErrorSeverityError, nil)
	}

	currentProps := b.options.Props()
	b.lastProps = currentProps

	return b.updateComponent(currentProps)
}

// IsActive returns whether the binding is currently active
func (b *ComponentBinding) IsActive() bool {
	return b.isActive
}

// GetComponentID returns the component ID
func (b *ComponentBinding) GetComponentID() string {
	return b.options.ComponentID
}

// GetComponentName returns the component name
func (b *ComponentBinding) GetComponentName() string {
	return b.options.ComponentName
}

// updateComponent updates the React component with new props
func (b *ComponentBinding) updateComponent(props map[string]interface{}) error {
	err := Update(ComponentID(b.options.ComponentID), props)
	if err != nil {
		return NewReactBridgeError("failed to update component", ErrorTypeUpdateFailure, ErrorSeverityError, map[string]interface{}{
			"componentID": b.options.ComponentID,
			"error":       err.Error(),
		})
	}

	return nil
}

// propsEqual compares two props maps for equality using type-aware, recursive comparison
func (b *ComponentBinding) propsEqual(a, propsB map[string]interface{}) bool {
	if len(a) != len(propsB) {
		return false
	}
	for key, va := range a {
		vb, ok := propsB[key]
		if !ok {
			return false
		}
		if !valuesEqual(va, vb) {
			return false
		}
	}
	return true
}

// valuesEqual performs deep, type-aware equality for supported types
func valuesEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}
	// js.Value: use Equal()
	if av, ok := a.(js.Value); ok {
		if bv, ok2 := b.(js.Value); ok2 {
			return av.Equal(bv)
		}
		return false
	}
	// Primitive direct comparisons
	switch ax := a.(type) {
	case string:
		bx, ok := b.(string)
		return ok && ax == bx
	case bool:
		bx, ok := b.(bool)
		return ok && ax == bx
	case int:
		switch bx := b.(type) {
		case int:
			return ax == bx
		case float64:
			return float64(ax) == bx
		}
		return false
	case int8:
		bx, ok := b.(int8)
		return ok && ax == bx
	case int16:
		bx, ok := b.(int16)
		return ok && ax == bx
	case int32:
		bx, ok := b.(int32)
		return ok && ax == bx
	case int64:
		bx, ok := b.(int64)
		return ok && ax == bx
	case uint:
		bx, ok := b.(uint)
		return ok && ax == bx
	case uint8:
		bx, ok := b.(uint8)
		return ok && ax == bx
	case uint16:
		bx, ok := b.(uint16)
		return ok && ax == bx
	case uint32:
		bx, ok := b.(uint32)
		return ok && ax == bx
	case uint64:
		bx, ok := b.(uint64)
		return ok && ax == bx
	case float32:
		bx, ok := b.(float32)
		return ok && ax == bx
	case float64:
		switch bx := b.(type) {
		case float64:
			return ax == bx
		case int:
			return ax == float64(bx)
		}
		return false
	}

	// Maps with string keys
	if am, ok := a.(map[string]interface{}); ok {
		bm, ok2 := b.(map[string]interface{})
		if !ok2 || len(am) != len(bm) {
			return false
		}
		for k, va := range am {
			vb, ok := bm[k]
			if !ok || !valuesEqual(va, vb) {
				return false
			}
		}
		return true
	}
	// Arrays/slices
	switch av := a.(type) {
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valuesEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	}
	// Fallback to reflect for typed slices/arrays
	if ra, rb := reflect.ValueOf(a), reflect.ValueOf(b); ra.IsValid() && rb.IsValid() {
		if (ra.Kind() == reflect.Slice || ra.Kind() == reflect.Array) && (rb.Kind() == reflect.Slice || rb.Kind() == reflect.Array) {
			if ra.Len() != rb.Len() {
				return false
			}
			for i := 0; i < ra.Len(); i++ {
				if !valuesEqual(ra.Index(i).Interface(), rb.Index(i).Interface()) {
					return false
				}
			}
			return true
		}
		if ra.Kind() == reflect.Map && rb.Kind() == reflect.Map && ra.Type().Key().Kind() == reflect.String && rb.Type().Key().Kind() == reflect.String {
			if ra.Len() != rb.Len() {
				return false
			}
			iter := ra.MapRange()
			for iter.Next() {
				vb := rb.MapIndex(iter.Key())
				if !vb.IsValid() || !valuesEqual(iter.Value().Interface(), vb.Interface()) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// convertToString converts a value to string for comparison
func convertToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case js.Value:
		if v.Type() == js.TypeString {
			return v.String()
		} else if v.Type() == js.TypeNumber {
			return fmt.Sprintf("%g", v.Float())
		} else if v.Type() == js.TypeBoolean {
			return fmt.Sprintf("%t", v.Bool())
		} else {
			return v.String()
		}
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// BindingManager manages multiple component bindings
type BindingManager struct {
	bindings map[string]*ComponentBinding
}

// NewBindingManager creates a new binding manager
func NewBindingManager() *BindingManager {
	return &BindingManager{
		bindings: make(map[string]*ComponentBinding),
	}
}

// Add adds a binding to the manager
func (m *BindingManager) Add(binding *ComponentBinding) {
	m.bindings[binding.GetComponentID()] = binding
}

// Remove removes a binding from the manager
func (m *BindingManager) Remove(componentID string) error {
	binding, exists := m.bindings[componentID]
	if !exists {
		return NewReactBridgeError("binding not found", ErrorTypeNotFound, ErrorSeverityError, map[string]interface{}{
			"componentID": componentID,
		})
	}

	err := binding.Stop()
	delete(m.bindings, componentID)
	return err
}

// Get retrieves a binding by component ID
func (m *BindingManager) Get(componentID string) (*ComponentBinding, bool) {
	binding, exists := m.bindings[componentID]
	return binding, exists
}

// StopAll stops all bindings
func (m *BindingManager) StopAll() error {
	var lastError error

	for componentID, binding := range m.bindings {
		err := binding.Stop()
		if err != nil {
			lastError = err
			logutil.Logf("Error stopping binding for %s: %v", componentID, err)
		}
	}

	// Clear all bindings
	m.bindings = make(map[string]*ComponentBinding)

	return lastError
}

// Count returns the number of active bindings
func (m *BindingManager) Count() int {
	return len(m.bindings)
}

// Global binding manager instance
var globalBindingManager = NewBindingManager()

// GetGlobalBindingManager returns the global binding manager instance
func GetGlobalBindingManager() *BindingManager {
	return globalBindingManager
}

// ThemeBinding represents a reactive binding between a Go signal and the theme system
type ThemeBinding struct {
	effect      reactivity.Effect
	isActive    bool
	themeSignal reactivity.Signal[string]
}

// BindTheme creates a reactive binding between a Go signal and the theme system
// The theme will automatically update when the signal changes
func BindTheme(themeSignal reactivity.Signal[string]) (*ThemeBinding, error) {
	if themeSignal == nil {
		return nil, NewReactBridgeError("theme signal is required", ErrorTypeInvalidArgument, ErrorSeverityError, nil)
	}

	binding := &ThemeBinding{
		isActive:    false,
		themeSignal: themeSignal,
	}

	// Create an effect that will run whenever the theme signal changes
	// We'll create it when the binding starts to ensure it runs properly

	return binding, nil
}

// Start activates the theme binding
func (tb *ThemeBinding) Start() error {
	if tb.isActive {
		return nil
	}

	tb.isActive = true

	// Create the effect now that we're active
	if tb.themeSignal != nil {
		tb.effect = reactivity.CreateEffect(func() {
			if !tb.isActive {
				return
			}

			// Get current theme from the signal
			currentTheme := tb.themeSignal.Get()

			// Update the theme
			err := SetTheme(currentTheme)
			if err != nil {
				logutil.Logf("Failed to set theme to %s: %v", currentTheme, err)
			}
		})
	}

	return nil
}

// Stop deactivates the theme binding
func (tb *ThemeBinding) Stop() {
	if !tb.isActive {
		return
	}

	tb.isActive = false

	// Clean up the effect
	if tb.effect != nil {
		tb.effect.Dispose()
	}
}

// IsActive returns whether the theme binding is currently active
func (tb *ThemeBinding) IsActive() bool {
	return tb.isActive
}
