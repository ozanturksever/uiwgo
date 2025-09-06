package appmanager

import (
	"context"
	"strings"
	"testing"
)

func TestAppManager_Creation(t *testing.T) {
	config := &AppConfig{
		AppID:          "test-app",
		MountElementID: "app",
		InitialState: AppState{
			User: map[string]any{"name": "test"},
			UI:   UIState{Theme: "light"},
		},
	}

	manager := NewAppManager(config)

	if manager == nil {
		t.Fatal("NewAppManager returned nil")
	}
	if manager.config.AppID != "test-app" {
		t.Errorf("Expected AppID 'test-app', got '%s'", manager.config.AppID)
	}
	if manager.initialized.Get() {
		t.Error("Expected initialized to be false")
	}
	if manager.running.Get() {
		t.Error("Expected running to be false")
	}

	// Test reactive state
	state := manager.store.Get()
	name := state.User.(map[string]any)["name"]
	if name != "test" {
		t.Error("Initial state not properly set")
	}
}

func TestAppManager_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful_initialization",
			config: &AppConfig{
				AppID:          "test-app",
				MountElementID: "app",
				InitialState:   AppState{},
			},
			expectError: false,
		},
		{
			name: "double_initialization",
			config: &AppConfig{
				AppID:          "test-app",
				MountElementID: "app",
				InitialState:   AppState{},
			},
			expectError: true,
			errorMsg:    "already initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mockdom for testing (sets a mock bridge manager)
			
			manager := NewAppManager(tt.config)

			// First initialization
			ctx := context.Background()
			err := manager.Initialize(ctx)

			if tt.name == "double_initialization" {
				// Try to initialize again
				err = manager.Initialize(ctx)
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !manager.initialized.Get() {
					t.Error("Expected manager to be initialized")
				}
				if manager.lifecycle.GetState() != LifecycleStateInitialized {
					t.Error("Expected lifecycle state to be initialized")
				}
			}
		})
	}
}
