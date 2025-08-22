package reactivity

import "testing"

func TestOnCleanupCalledOnRerun(t *testing.T) {
	s := CreateSignal(0)
	cleanupCalls := 0
	runs := 0
	_ = CreateEffect(func() {
		OnCleanup(func() { cleanupCalls++ })
		_ = s.Get()
		runs++
	})

	if runs != 1 || cleanupCalls != 0 {
		t.Fatalf("init runs=%d cleanup=%d, want 1 and 0", runs, cleanupCalls)
	}

	s.Set(1)
	if runs != 2 {
		t.Fatalf("runs after set=%d, want 2", runs)
	}
	if cleanupCalls != 1 {
		t.Fatalf("cleanup after rerun=%d, want 1", cleanupCalls)
	}
}

func TestOnCleanupCalledOnDispose(t *testing.T) {
	cleanupCalls := 0
	e := CreateEffect(func() {
		OnCleanup(func() { cleanupCalls++ })
	})
	if cleanupCalls != 0 {
		t.Fatalf("cleanup before dispose=%d, want 0", cleanupCalls)
	}
	e.Dispose()
	if cleanupCalls != 1 {
		t.Fatalf("cleanup after dispose=%d, want 1", cleanupCalls)
	}
}
