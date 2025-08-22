package reactivity

import "testing"

func TestEffectDependencyAndDispose(t *testing.T) {
	s1 := CreateSignal(1)
	s2 := CreateSignal(10)

	runs := 0
	e := CreateEffect(func() {
		_ = s1.Get()
		runs++
	})

	if runs != 1 {
		t.Fatalf("initial runs = %d, want 1", runs)
	}

	// Changing unrelated signal should not re-run
	s2.Set(20)
	if runs != 1 {
		t.Fatalf("runs after unrelated signal = %d, want 1", runs)
	}

	// Changing related signal re-runs
	s1.Set(2)
	if runs != 2 {
		t.Fatalf("runs after related signal = %d, want 2", runs)
	}

	// Dispose prevents further runs
	e.Dispose()
	s1.Set(3)
	if runs != 2 {
		t.Fatalf("runs after dispose = %d, want 2", runs)
	}
}
