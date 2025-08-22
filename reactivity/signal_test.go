package reactivity

import (
	"testing"
)

func TestCreateSignalInitialAndSet(t *testing.T) {
	s := CreateSignal(1)
	if got := s.Get(); got != 1 {
		t.Fatalf("initial value = %d, want 1", got)
	}

	// Track runs for effect
	runs := 0
	_ = CreateEffect(func() {
		_ = s.Get()
		runs++
	})
	if runs != 1 {
		t.Fatalf("effect initial runs = %d, want 1", runs)
	}

	s.Set(2)
	if got := s.Get(); got != 2 {
		t.Fatalf("after set value = %d, want 2", got)
	}
	if runs != 2 {
		t.Fatalf("effect runs after set = %d, want 2", runs)
	}
}

func TestNoTriggerOnSameValue(t *testing.T) {
	s := CreateSignal(0)
	runs := 0
	_ = CreateEffect(func() {
		_ = s.Get()
		runs++
	})
	if runs != 1 {
		t.Fatalf("initial runs = %d, want 1", runs)
	}

	s.Set(0) // same value
	if runs != 1 {
		t.Fatalf("runs after same value set = %d, want 1", runs)
	}
}

func TestUnrelatedSignalDoesNotTrigger(t *testing.T) {
	s1 := CreateSignal(1)
	s2 := CreateSignal(10)
	runs := 0
	_ = CreateEffect(func() {
		_ = s1.Get()
		runs++
	})

	s2.Set(20)
	if runs != 1 {
		t.Fatalf("runs after unrelated signal set = %d, want 1", runs)
	}
}
