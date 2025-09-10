package action

import (
	"time"
)

// ActionType represents a typed action identifier with a stable name.
type ActionType[T any] struct {
	Name string
}

// DefineAction creates a new ActionType with the given name.
// The name should be unique within the application to avoid conflicts.
func DefineAction[T any](name string) ActionType[T] {
	return ActionType[T]{Name: name}
}

// QueryType represents a typed query identifier with request and response types.
type QueryType[Req, Res any] struct {
	Name string
}

// DefineQuery creates a new QueryType with the given name.
// The name should be unique within the application to avoid conflicts.
func DefineQuery[Req, Res any](name string) QueryType[Req, Res] {
	return QueryType[Req, Res]{Name: name}
}

// Action represents a typed action with payload and metadata.
type Action[T any] struct {
	Type    string         // The action type name
	Payload T              // The typed payload data
	Meta    map[string]any // Optional metadata for the action
	Time    time.Time      // Timestamp when the action was created
	Source  string         // Source identifier of the action creator
	TraceID string         // Trace ID for distributed tracing
}

// Context provides execution context for action processing.
type Context struct {
	Scope   string         // Scope identifier for the context
	Meta    map[string]any // Metadata associated with the context
	Time    time.Time      // Timestamp when the context was created
	TraceID string         // Trace ID for distributed tracing
	Source  string         // Source identifier
}

// MetaWith creates a new Context with additional metadata.
// The original context remains unchanged.
func (c Context) MetaWith(key string, value any) Context {
	// Create a copy of the meta map to avoid modifying the original
	newMeta := make(map[string]any)
	for k, v := range c.Meta {
		newMeta[k] = v
	}
	newMeta[key] = value

	return Context{
		Scope:   c.Scope,
		Meta:    newMeta,
		Time:    c.Time,
		TraceID: c.TraceID,
		Source:  c.Source,
	}
}

// MetaValue retrieves a value from the context's metadata.
// Returns the value and true if found, nil and false otherwise.
func (c Context) MetaValue(key string) (any, bool) {
	if c.Meta == nil {
		return nil, false
	}
	value, exists := c.Meta[key]
	return value, exists
}
