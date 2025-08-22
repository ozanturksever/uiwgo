package reactivity

import "testing"

func TestMemoLazyEvaluationAndCaching(t *testing.T) {
	count := CreateSignal(1)
	calls := 0
	memo := CreateMemo(func() int {
		calls++
		return count.Get() * 2
	})

	// Lazy: without Get, calc should not run
	if calls != 0 {
		t.Fatalf("calc calls before Get = %d, want 0", calls)
	}

	// First Get triggers exactly one calculation
	if v := memo.Get(); v != 2 {
		t.Fatalf("first memo.Get() = %d, want 2", v)
	}
	if calls != 1 {
		t.Fatalf("calc calls after first Get = %d, want 1", calls)
	}

	// Multiple Get without change should not re-calc
	_ = memo.Get()
	_ = memo.Get()
	if calls != 1 {
		t.Fatalf("calc calls after repeated Get = %d, want 1", calls)
	}
}

func TestMemoRecomputeOnDepChangeAndEffectPropagation(t *testing.T) {
	s := CreateSignal(2)
	calls := 0
	memo := CreateMemo(func() int {
		calls++
		return s.Get() + 1
	})

	// Hook an effect to memo
	runs := 0
	_ = CreateEffect(func() {
		_ = memo.Get()
		runs++
	})

	if runs != 1 {
		t.Fatalf("initial runs = %d, want 1", runs)
	}
	if calls != 1 {
		t.Fatalf("initial calc calls = %d, want 1", calls)
	}

	// Change dependency -> memo recalculates and effect runs
	s.Set(3)
	if calls != 2 {
		t.Fatalf("calc calls after dep change = %d, want 2", calls)
	}
	if runs != 2 {
		t.Fatalf("effect runs after dep change = %d, want 2", runs)
	}
}

func TestMemoNotAffectedByUnrelatedSignal(t *testing.T) {
	dep := CreateSignal(1)
	unrelated := CreateSignal(100)
	calls := 0
	memo := CreateMemo(func() int {
		calls++
		return dep.Get() * 3
	})

	_ = memo.Get() // initialize
	if calls != 1 {
		t.Fatalf("initial calls = %d, want 1", calls)
	}

	unrelated.Set(200)
	if calls != 1 {
		t.Fatalf("calls after unrelated set = %d, want 1", calls)
	}
}

func TestChainedMemos(t *testing.T) {
	base := CreateSignal(1)
	calls1, calls2 := 0, 0
	m1 := CreateMemo(func() int {
		calls1++
		return base.Get() + 1
	})
	m2 := CreateMemo(func() int {
		calls2++
		return m1.Get() * 2
	})

	// Use m2 in an effect so updates propagate
	runs := 0
	_ = CreateEffect(func() {
		_ = m2.Get()
		runs++
	})

	if runs != 1 || calls1 != 1 || calls2 != 1 {
		t.Fatalf("init runs=%d calls1=%d calls2=%d, want 1,1,1", runs, calls1, calls2)
	}

	base.Set(2)
	if calls1 != 2 {
		t.Fatalf("calls1 after base change = %d, want 2", calls1)
	}
	if calls2 != 2 {
		t.Fatalf("calls2 after base change = %d, want 2", calls2)
	}
	if runs != 2 {
		t.Fatalf("effect runs after base change = %d, want 2", runs)
	}
}
