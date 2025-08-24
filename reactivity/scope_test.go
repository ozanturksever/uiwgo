package reactivity

import (
	"testing"
)

func TestCleanupScopeCreation(t *testing.T) {
	scope := NewCleanupScope(nil)
	if scope == nil {
		t.Fatal("NewCleanupScope should not return nil")
	}
	if scope.disposed {
		t.Error("New scope should not be disposed")
	}
	if len(scope.disposers) != 0 {
		t.Error("New scope should have empty disposers")
	}
	if scope.parent != nil {
		t.Error("Root scope should have nil parent")
	}
}

func TestCleanupScopeWithParent(t *testing.T) {
	parent := NewCleanupScope(nil)
	child := NewCleanupScope(parent)
	
	if child.parent != parent {
		t.Error("Child scope should have correct parent")
	}
	if len(parent.children) != 1 {
		t.Error("Parent should have one child")
	}
	if parent.children[0] != child {
		t.Error("Parent should contain child reference")
	}
}

func TestCleanupScopeRegisterDisposer(t *testing.T) {
	scope := NewCleanupScope(nil)
	called := false
	
	scope.RegisterDisposer(func() {
		called = true
	})
	
	if len(scope.disposers) != 1 {
		t.Error("Scope should have one disposer")
	}
	
	scope.Dispose()
	
	if !called {
		t.Error("Disposer should have been called")
	}
	if !scope.disposed {
		t.Error("Scope should be marked as disposed")
	}
}

func TestCleanupScopeMultipleDisposers(t *testing.T) {
	scope := NewCleanupScope(nil)
	callCount := 0
	
	scope.RegisterDisposer(func() { callCount++ })
	scope.RegisterDisposer(func() { callCount++ })
	scope.RegisterDisposer(func() { callCount++ })
	
	scope.Dispose()
	
	if callCount != 3 {
		t.Errorf("Expected 3 disposer calls, got %d", callCount)
	}
}

func TestCleanupScopeDisposeOnlyOnce(t *testing.T) {
	scope := NewCleanupScope(nil)
	callCount := 0
	
	scope.RegisterDisposer(func() { callCount++ })
	
	scope.Dispose()
	scope.Dispose() // Second call should be ignored
	
	if callCount != 1 {
		t.Errorf("Expected 1 disposer call, got %d", callCount)
	}
}

func TestCleanupScopeRegisterAfterDispose(t *testing.T) {
	scope := NewCleanupScope(nil)
	scope.Dispose()
	
	called := false
	scope.RegisterDisposer(func() {
		called = true
	})
	
	if called {
		t.Error("Disposer should not be called when registered after disposal")
	}
	if len(scope.disposers) != 0 {
		t.Error("Disposed scope should not accept new disposers")
	}
}

func TestCleanupScopeChildDisposal(t *testing.T) {
	parent := NewCleanupScope(nil)
	child := NewCleanupScope(parent)
	
	childDisposed := false
	child.RegisterDisposer(func() {
		childDisposed = true
	})
	
	parent.Dispose()
	
	if !childDisposed {
		t.Error("Child disposer should be called when parent is disposed")
	}
	if !child.disposed {
		t.Error("Child should be marked as disposed when parent is disposed")
	}
	if !parent.disposed {
		t.Error("Parent should be marked as disposed")
	}
}

func TestCleanupScopeNestedChildren(t *testing.T) {
	root := NewCleanupScope(nil)
	child1 := NewCleanupScope(root)
	child2 := NewCleanupScope(root)
	grandchild := NewCleanupScope(child1)
	
	disposalOrder := []string{}
	
	root.RegisterDisposer(func() { disposalOrder = append(disposalOrder, "root") })
	child1.RegisterDisposer(func() { disposalOrder = append(disposalOrder, "child1") })
	child2.RegisterDisposer(func() { disposalOrder = append(disposalOrder, "child2") })
	grandchild.RegisterDisposer(func() { disposalOrder = append(disposalOrder, "grandchild") })
	
	root.Dispose()
	
	if len(disposalOrder) != 4 {
		t.Errorf("Expected 4 disposals, got %d", len(disposalOrder))
	}
	
	// All should be disposed
	if !root.disposed || !child1.disposed || !child2.disposed || !grandchild.disposed {
		t.Error("All scopes should be disposed")
	}
}

func TestCleanupScopeCurrentScope(t *testing.T) {
	original := GetCurrentCleanupScope()
	
	scope := NewCleanupScope(nil)
	SetCurrentCleanupScope(scope)
	
	if GetCurrentCleanupScope() != scope {
		t.Error("Current scope should be set correctly")
	}
	
	// Restore original
	SetCurrentCleanupScope(original)
	
	if GetCurrentCleanupScope() != original {
		t.Error("Original scope should be restored")
	}
}

func TestCleanupScopeWithCurrentScope(t *testing.T) {
	parent := NewCleanupScope(nil)
	SetCurrentCleanupScope(parent)
	
	called := false
	RegisterCleanup(func() {
		called = true
	})
	
	parent.Dispose()
	
	if !called {
		t.Error("RegisterCleanup should register with current scope")
	}
	
	// Clean up
	SetCurrentCleanupScope(nil)
}

func TestCleanupScopeRegisterCleanupWithoutCurrentScope(t *testing.T) {
	SetCurrentCleanupScope(nil)
	
	called := false
	RegisterCleanup(func() {
		called = true
	})
	
	// Should not panic and should not be called
	if called {
		t.Error("RegisterCleanup without current scope should be ignored")
	}
}

func TestCleanupScopeWithEffect(t *testing.T) {
	scope := NewCleanupScope(nil)
	SetCurrentCleanupScope(scope)
	
	s := CreateSignal(0)
	effectRuns := 0
	cleanupCalls := 0
	
	CreateEffect(func() {
		_ = s.Get()
		effectRuns++
		OnCleanup(func() {
			cleanupCalls++
		})
	})
	
	if effectRuns != 1 {
		t.Errorf("Expected 1 effect run, got %d", effectRuns)
	}
	
	// Dispose the scope should also dispose the effect
	scope.Dispose()
	
	if cleanupCalls != 1 {
		t.Errorf("Expected 1 cleanup call after scope disposal, got %d", cleanupCalls)
	}
	
	// Try to trigger the effect again - it should not run since it's disposed
	s.Set(1)
	if effectRuns != 1 {
		t.Errorf("Effect should not run after disposal, got %d runs", effectRuns)
	}
	
	// Clean up
	SetCurrentCleanupScope(nil)
}

func TestCleanupScopeWithMultipleEffects(t *testing.T) {
	scope := NewCleanupScope(nil)
	SetCurrentCleanupScope(scope)
	
	s1 := CreateSignal(0)
	s2 := CreateSignal(0)
	effect1Runs := 0
	effect2Runs := 0
	cleanup1Calls := 0
	cleanup2Calls := 0
	
	CreateEffect(func() {
		_ = s1.Get()
		effect1Runs++
		OnCleanup(func() {
			cleanup1Calls++
		})
	})
	
	CreateEffect(func() {
		_ = s2.Get()
		effect2Runs++
		OnCleanup(func() {
			cleanup2Calls++
		})
	})
	
	if effect1Runs != 1 || effect2Runs != 1 {
		t.Errorf("Expected 1 run each, got %d and %d", effect1Runs, effect2Runs)
	}
	
	// Dispose the scope should dispose both effects
	scope.Dispose()
	
	if cleanup1Calls != 1 || cleanup2Calls != 1 {
		t.Errorf("Expected 1 cleanup call each, got %d and %d", cleanup1Calls, cleanup2Calls)
	}
	
	// Try to trigger the effects again - they should not run since they're disposed
	s1.Set(1)
	s2.Set(1)
	if effect1Runs != 1 || effect2Runs != 1 {
		t.Errorf("Effects should not run after disposal, got %d and %d runs", effect1Runs, effect2Runs)
	}
	
	// Clean up
	SetCurrentCleanupScope(nil)
}

func TestCleanupScopeWithCleanupScope(t *testing.T) {
	parent := NewCleanupScope(nil)
	effectRuns := 0
	cleanupCalls := 0
	
	WithCleanupScope(parent, func(scope *CleanupScope) {
		s := CreateSignal(0)
		CreateEffect(func() {
			_ = s.Get()
			effectRuns++
			OnCleanup(func() {
				cleanupCalls++
			})
		})
	})
	
	if effectRuns != 1 {
		t.Errorf("Expected 1 effect run, got %d", effectRuns)
	}
	
	if cleanupCalls != 1 {
		t.Errorf("Expected 1 cleanup call after WithCleanupScope, got %d", cleanupCalls)
	}
}