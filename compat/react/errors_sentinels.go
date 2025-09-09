//go:build js && wasm

package react

import "errors"

// Sentinel errors used for common failure cases. These are referenced in tests
// and allow callers to use errors.Is for precise matching.
var (
    ErrBridgeNotInitialized = errors.New("react bridge not initialized")
    ErrComponentIDRequired  = errors.New("component ID is required")
    ErrComponentNameRequired = errors.New("component name is required")
    ErrPropsFuncRequired    = errors.New("props function is required")
)
