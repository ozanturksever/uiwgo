package action

import (
	"testing"
)

func TestDefineActionReturnsStableName(t *testing.T) {
	// Test that DefineAction returns a stable name
	actionType := DefineAction[string]("test-action")

	if actionType.Name != "test-action" {
		t.Errorf("Expected action name 'test-action', got '%s'", actionType.Name)
	}

	// Test that calling it again returns the same name
	actionType2 := DefineAction[string]("test-action")
	if actionType.Name != actionType2.Name {
		t.Errorf("Expected same action name, got '%s' and '%s'", actionType.Name, actionType2.Name)
	}
}

func TestActionMetadataDefaults(t *testing.T) {
	// Test that Action has proper default values
	actionType := DefineAction[string]("test-action")
	action := Action[string]{
		Type:    actionType.Name,
		Payload: "test-payload",
	}

	// Meta should be nil by default (not initialized)
	if action.Meta != nil {
		t.Errorf("Expected Meta to be nil by default, got %v", action.Meta)
	}

	// Test with initialized Meta
	action.Meta = make(map[string]any)
	action.Meta["key"] = "value"

	if action.Meta["key"] != "value" {
		t.Errorf("Expected Meta to contain 'key': 'value', got %v", action.Meta)
	}

	// Test other fields
	if action.Type != "test-action" {
		t.Errorf("Expected Type to be 'test-action', got '%s'", action.Type)
	}

	if action.Payload != "test-payload" {
		t.Errorf("Expected Payload to be 'test-payload', got '%s'", action.Payload)
	}

	// Time should be zero by default
	if !action.Time.IsZero() {
		t.Errorf("Expected Time to be zero by default, got %v", action.Time)
	}

	// Source should be empty by default
	if action.Source != "" {
		t.Errorf("Expected Source to be empty by default, got '%s'", action.Source)
	}

	// TraceID should be empty by default
	if action.TraceID != "" {
		t.Errorf("Expected TraceID to be empty by default, got '%s'", action.TraceID)
	}
}

func TestContextMetaAccessors(t *testing.T) {
	// Test Context creation and Meta accessors
	ctx := Context{
		Scope: "test-scope",
		Meta:  make(map[string]any),
	}

	// Test MetaWith method
	ctx.Meta["existing"] = "value"
	newCtx := ctx.MetaWith("new-key", "new-value")

	// Original context should be unchanged
	if ctx.Meta["new-key"] != nil {
		t.Errorf("Expected original context to be unchanged, but found new-key: %v", ctx.Meta["new-key"])
	}

	// New context should have both keys
	if newCtx.Meta["existing"] != "value" {
		t.Errorf("Expected new context to preserve existing meta, got %v", newCtx.Meta["existing"])
	}

	if newCtx.Meta["new-key"] != "new-value" {
		t.Errorf("Expected new context to have new-key: 'new-value', got %v", newCtx.Meta["new-key"])
	}

	// Test MetaValue method
	value, exists := newCtx.MetaValue("new-key")
	if !exists || value != "new-value" {
		t.Errorf("Expected MetaValue to return 'new-value' and true, got %v and %v", value, exists)
	}

	// Test non-existent key
	value, exists = newCtx.MetaValue("non-existent")
	if exists || value != nil {
		t.Errorf("Expected MetaValue to return nil and false for non-existent key, got %v and %v", value, exists)
	}

	// Test default values
	if newCtx.Scope != "test-scope" {
		t.Errorf("Expected Scope to be preserved, got '%s'", newCtx.Scope)
	}

	if !newCtx.Time.IsZero() {
		t.Errorf("Expected Time to be zero by default, got %v", newCtx.Time)
	}

	if newCtx.Source != "" {
		t.Errorf("Expected Source to be empty by default, got '%s'", newCtx.Source)
	}

	if newCtx.TraceID != "" {
		t.Errorf("Expected TraceID to be empty by default, got '%s'", newCtx.TraceID)
	}
}
