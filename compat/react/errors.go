//go:build js && wasm

package react

import (
	"fmt"
	"time"
)

// ErrorType represents different categories of errors
type ErrorType string

// Error types for categorization
const (
	ErrorTypeComponentRender     ErrorType = "component_render"
	ErrorTypeComponentUpdate     ErrorType = "component_update"
	ErrorTypeComponentUnmount    ErrorType = "component_unmount"
	ErrorTypeCallbackInvocation  ErrorType = "callback_invocation"
	ErrorTypeCallbackNotFound    ErrorType = "callback_not_found"
	ErrorTypeEventEmission       ErrorType = "event_emission"
	ErrorTypePropsSerializer     ErrorType = "props_serialization"
	ErrorTypeBridgeInit          ErrorType = "bridge_initialization"
	ErrorTypeDOMManipulation     ErrorType = "dom_manipulation"
	ErrorTypeUnknown             ErrorType = "unknown"
	ErrorTypeInvalidArgument     ErrorType = "invalid_argument"
	ErrorTypeNotFound            ErrorType = "not_found"
	ErrorTypeRenderFailure       ErrorType = "render_failure"
	ErrorTypeUpdateFailure       ErrorType = "update_failure"
	ErrorTypeUnmountFailure      ErrorType = "unmount_failure"
	ErrorTypeInvalidState        ErrorType = "invalid_state"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

// Error severity levels
const (
	ErrorSeverityInfo    ErrorSeverity = "info"
	ErrorSeverityWarning ErrorSeverity = "warning"
	ErrorSeverityError   ErrorSeverity = "error"
	ErrorSeverityFatal   ErrorSeverity = "fatal"
)

// ReactBridgeError represents an enhanced error with additional context
type ReactBridgeError struct {
	Message   string                 `json:"message"`
	Type      ErrorType              `json:"type"`
	Severity  ErrorSeverity          `json:"severity"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
	Stack     string                 `json:"stack,omitempty"`
}

// Error implements the error interface
func (e *ReactBridgeError) Error() string {
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Severity, e.Message)
}

// NewReactBridgeError creates a new ReactBridgeError
func NewReactBridgeError(message string, errorType ErrorType, severity ErrorSeverity, context map[string]interface{}) *ReactBridgeError {
	if context == nil {
		context = make(map[string]interface{})
	}
	
	return &ReactBridgeError{
		Message:   message,
		Type:      errorType,
		Severity:  severity,
		Context:   context,
		Timestamp: time.Now(),
	}
}

// WithStack adds stack trace information to the error
func (e *ReactBridgeError) WithStack(stack string) *ReactBridgeError {
	e.Stack = stack
	return e
}

// WithContext adds additional context to the error
func (e *ReactBridgeError) WithContext(key string, value interface{}) *ReactBridgeError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}