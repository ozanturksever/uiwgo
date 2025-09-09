//go:build js && wasm

package react

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/ozanturksever/logutil"
)

// ComponentID represents a unique identifier for a React component instance
type ComponentID string

// Props represents the properties passed to a React component
type Props map[string]interface{}

// RenderOptions contains options for rendering a React component
type RenderOptions struct {
	ContainerID string `json:"containerId,omitempty"`
	Replace     bool   `json:"replace,omitempty"`
}

// ReactBridge provides the Go interface to the JavaScript React bridge
type ReactBridge struct {
	bridge js.Value
}

// NewReactBridge creates a new React bridge instance
func NewReactBridge() (*ReactBridge, error) {
	if !js.Global().Get("window").Truthy() {
		return nil, fmt.Errorf("window object not available")
	}

	reactCompat := js.Global().Get("ReactCompat")
	if !reactCompat.Truthy() {
		return nil, fmt.Errorf("ReactCompat not found on window object")
	}

	return &ReactBridge{
		bridge: reactCompat,
	}, nil
}

// Render renders a React component and returns its component ID
func (rb *ReactBridge) Render(componentName string, props Props, options *RenderOptions) (ComponentID, error) {
	logutil.Logf("Rendering React component: %s", componentName)

	// Validate props for unsupported types to preserve predictable interop
	if err := validateProps(props); err != nil {
		return "", fmt.Errorf("failed to serialize props: %w", err)
	}

	// Convert props directly to a JS object to avoid JSON marshalling
	propsJS := MapToJSObject(props)

	// Prepare options as JS object (or undefined)
	var optionsJS js.Value
	if options != nil {
		optMap := map[string]interface{}{}
		if options.ContainerID != "" {
			optMap["containerId"] = options.ContainerID
		}
		if options.Replace {
			optMap["replace"] = true
		}
		optionsJS = MapToJSObject(optMap)
	} else {
		optionsJS = js.Undefined()
	}

	// Call the JavaScript bridge
	result := rb.bridge.Call("renderComponent", componentName, propsJS, optionsJS)

	if !result.Truthy() {
		return "", fmt.Errorf("failed to render component %s", componentName)
	}

	componentID := ComponentID(result.String())
	logutil.Logf("Component %s rendered with ID: %s", componentName, componentID)

	return componentID, nil
}

// Update updates the props of an existing React component
func (rb *ReactBridge) Update(componentID ComponentID, props Props) error {
	logutil.Logf("Updating React component: %s", componentID)

	// Validate props
	if err := validateProps(props); err != nil {
		return fmt.Errorf("failed to serialize props: %w", err)
	}

	// Convert props to JS object directly
	propsJS := MapToJSObject(props)

	// Call the JavaScript bridge
	result := rb.bridge.Call("updateComponent", string(componentID), propsJS)

	if !result.Truthy() {
		return fmt.Errorf("failed to update component %s", componentID)
	}

	logutil.Logf("Component %s updated successfully", componentID)
	return nil
}

// Unmount unmounts a React component
func (rb *ReactBridge) Unmount(componentID ComponentID) error {
	logutil.Logf("Unmounting React component: %s", componentID)

	// Call the JavaScript bridge
	result := rb.bridge.Call("unmountComponent", string(componentID))

	if !result.Truthy() {
		return fmt.Errorf("failed to unmount component %s", componentID)
	}

	logutil.Logf("Component %s unmounted successfully", componentID)
	return nil
}

// Register registers a React component with the bridge
func (rb *ReactBridge) Register(componentName string, component js.Value) error {
	logutil.Logf("Registering React component: %s", componentName)

	// Call the JavaScript bridge
	result := rb.bridge.Call("register", componentName, component)

	if !result.Truthy() {
		return fmt.Errorf("failed to register component %s", componentName)
	}

	logutil.Logf("Component %s registered successfully", componentName)
	return nil
}

// Resolve resolves a registered React component
func (rb *ReactBridge) Resolve(componentName string) (js.Value, error) {
	logutil.Logf("Resolving React component: %s", componentName)

	// Call the JavaScript bridge
	result := rb.bridge.Call("resolve", componentName)

	if !result.Truthy() {
		return js.Undefined(), fmt.Errorf("component %s not found", componentName)
	}

	logutil.Logf("Component %s resolved successfully", componentName)
	return result, nil
}

// GetDiagnostics returns diagnostic information from the React bridge
func (rb *ReactBridge) GetDiagnostics() (map[string]interface{}, error) {
	logutil.Log("Getting React bridge diagnostics")

	// Call the JavaScript bridge
	result := rb.bridge.Call("getDiagnostics")

	if !result.Truthy() {
		return nil, fmt.Errorf("failed to get diagnostics")
	}

	// Convert JS object to Go map
	diagnosticsJSON := js.Global().Get("JSON").Call("stringify", result).String()
	var diagnostics map[string]interface{}
	if err := json.Unmarshal([]byte(diagnosticsJSON), &diagnostics); err != nil {
		return nil, fmt.Errorf("failed to parse diagnostics: %w", err)
	}

	return diagnostics, nil
}

// Global bridge instance
var globalBridge *ReactBridge

// InitializeBridge initializes the global React bridge
func InitializeBridge() error {
	// Always (re)initialize to ensure tests that mutate the JS bridge get a fresh instance
	bridge, err := NewReactBridge()
	if err != nil {
		return fmt.Errorf("failed to initialize React bridge: %w", err)
	}

	globalBridge = bridge
	logutil.Log("React bridge initialized successfully")
	return nil
}

// GetBridge returns the global React bridge instance
func GetBridge() (*ReactBridge, error) {
	if globalBridge == nil {
		return nil, ErrBridgeNotInitialized
	}
	return globalBridge, nil
}

// Convenience functions that use the global bridge

// Render renders a React component using the global bridge
func Render(componentName string, props Props, options *RenderOptions) (ComponentID, error) {
	bridge, err := GetBridge()
	if err != nil {
		return "", err
	}
	return bridge.Render(componentName, props, options)
}

// Update updates a React component using the global bridge
func Update(componentID ComponentID, props Props) error {
	bridge, err := GetBridge()
	if err != nil {
		return err
	}
	return bridge.Update(componentID, props)
}

// Unmount unmounts a React component using the global bridge
func Unmount(componentID ComponentID) error {
	bridge, err := GetBridge()
	if err != nil {
		return err
	}
	return bridge.Unmount(componentID)
}

// Register registers a React component using the global bridge
func Register(componentName string, component js.Value) error {
	bridge, err := GetBridge()
	if err != nil {
		return err
	}
	return bridge.Register(componentName, component)
}

// Resolve resolves a React component using the global bridge
func Resolve(componentName string) (js.Value, error) {
	bridge, err := GetBridge()
	if err != nil {
		return js.Undefined(), err
	}
	return bridge.Resolve(componentName)
}

// SetTheme sets the theme by calling the JavaScript setTheme function
func (rb *ReactBridge) SetTheme(theme string) error {
	logutil.Logf("Setting theme: %s", theme)

	setThemeFunc := rb.bridge.Get("setTheme")
	if !setThemeFunc.Truthy() {
		return fmt.Errorf("setTheme function not found on ReactCompat bridge")
	}

	// Call the JavaScript setTheme function
	result := setThemeFunc.Invoke(theme)
	if !result.IsUndefined() && !result.IsNull() {
		// Check if there was an error
		if result.Type() == js.TypeString {
			errorMsg := result.String()
			if errorMsg != "" {
				return fmt.Errorf("setTheme failed: %s", errorMsg)
			}
		}
	}

	return nil
}

// SetTheme sets the theme using the global bridge
func SetTheme(theme string) error {
	bridge, err := GetBridge()
	if err != nil {
		return err
	}
	return bridge.SetTheme(theme)
}

// GetDiagnostics returns diagnostic information using the global bridge
func GetDiagnostics() (map[string]interface{}, error) {
	bridge, err := GetBridge()
	if err != nil {
		return nil, err
	}
	return bridge.GetDiagnostics()
}
