package reactivity

import (
	"testing"
)

type testNested struct {
	A int
	B string
}

type testItem struct {
	ID        int
	Completed bool
}

type testApp struct {
	Items []testItem
}

func TestStore_FieldSpecificEffects(t *testing.T) {
	store, setState := CreateStore(testNested{A: 1, B: "x"})

	var runsA, runsB int
	_ = CreateEffect(func() {
		_ = Adapt[int](store.Select("A")).Get()
		runsA++
	})
	_ = CreateEffect(func() {
		_ = Adapt[string](store.Select("B")).Get()
		runsB++
	})

	if runsA != 1 || runsB != 1 {
		t.Fatalf("initial runsA=%d runsB=%d; want 1,1", runsA, runsB)
	}

	setState("A", 2)
	if runsA != 2 {
		t.Fatalf("runsA after A change = %d; want 2", runsA)
	}
	if runsB != 1 {
		t.Fatalf("runsB after A change = %d; want 1 (unchanged)", runsB)
	}

	// Set same value for B -> no rerun
	setState("B", "x")
	if runsB != 1 {
		t.Fatalf("runsB after setting same value = %d; want 1", runsB)
	}

	// Change B -> only B reruns
	setState("B", "y")
	if runsB != 2 {
		t.Fatalf("runsB after B change = %d; want 2", runsB)
	}
	if runsA != 2 {
		t.Fatalf("runsA after B change = %d; want 2 (unchanged)", runsA)
	}
}

func TestStore_SliceFineGrained(t *testing.T) {
	store, setState := CreateStore(testApp{Items: []testItem{{ID: 1}, {ID: 2}}})

	var runs0, runs1, runsLen int
	_ = CreateEffect(func() {
		_ = store.SelectLen("Items").Get()
		runsLen++
	})
	_ = CreateEffect(func() {
		_ = Adapt[bool](store.Select("Items", 0, "Completed")).Get()
		runs0++
	})
	_ = CreateEffect(func() {
		_ = Adapt[bool](store.Select("Items", 1, "Completed")).Get()
		runs1++
	})

	if runs0 != 1 || runs1 != 1 || runsLen != 1 {
		t.Fatalf("initial runs0=%d runs1=%d runsLen=%d; want 1,1,1", runs0, runs1, runsLen)
	}

	// Toggle only item 0 -> only its effect should rerun
	setState("Items", 0, "Completed", true)
	if runs0 != 2 {
		t.Fatalf("runs0 after item0 toggle = %d; want 2", runs0)
	}
	if runs1 != 1 {
		t.Fatalf("runs1 after item0 toggle = %d; want 1 (unchanged)", runs1)
	}
	if runsLen != 1 {
		t.Fatalf("runsLen after item0 toggle = %d; want 1 (unchanged)", runsLen)
	}

	// Append a new item by replacing the slice; len should change, per-item unchanged
	cur := store.Get().Items
	newList := append(append([]testItem{}, cur...), testItem{ID: 3})
	setState("Items", newList)
	if runsLen != 2 {
		t.Fatalf("runsLen after append = %d; want 2", runsLen)
	}
	if runs0 != 2 || runs1 != 1 {
		t.Fatalf("per-item runs after append: runs0=%d runs1=%d; want 2,1", runs0, runs1)
	}

	// Remove middle item (index 1). Length changes; per-existing items unchanged
	cur = store.Get().Items
	if len(cur) < 3 {
		t.Fatalf("expected at least 3 items before removal; got %d", len(cur))
	}
	list := make([]testItem, 0, len(cur)-1)
	for i, it := range cur {
		if i != 1 {
			list = append(list, it)
		}
	}
	setState("Items", list)
	if runsLen != 3 {
		t.Fatalf("runsLen after remove = %d; want 3", runsLen)
	}
	if runs0 != 2 || runs1 != 1 { // index 1 effect still points to the second position; values stayed equal (false)
		t.Fatalf("per-item runs after remove: runs0=%d runs1=%d; want 2,1", runs0, runs1)
	}
}

func TestStore_AdaptSetMutates(t *testing.T) {
	store, _ := CreateStore(testNested{A: 1, B: "x"})
	sa := Adapt[int](store.Select("A"))

	runs := 0
	_ = CreateEffect(func() {
		_ = sa.Get()
		runs++
	})
	if runs != 1 {
		t.Fatalf("initial runs = %d; want 1", runs)
	}

	sa.Set(5)
	if got := sa.Get(); got != 5 {
		t.Fatalf("A after set via Adapt = %d; want 5", got)
	}
	if runs != 2 {
		t.Fatalf("runs after set via Adapt = %d; want 2", runs)
	}
}

func TestStore_ExpandThenSelectField(t *testing.T) {
	store, setState := CreateStore(testApp{Items: []testItem{}})

	// Selecting before any items exist should not panic and should create typed nodes
	runs := 0
	_ = CreateEffect(func() {
		_ = Adapt[bool](store.Select("Items", 0, "Completed")).Get()
		runs++
	})
	if runs != 1 {
		t.Fatalf("initial runs = %d; want 1", runs)
	}

	// Now set nested field via expansion in setState; should rerun once
	setState("Items", 0, "Completed", true)
	if runs != 2 {
		t.Fatalf("runs after setting nested field via expansion = %d; want 2", runs)
	}

	// Setting same value should not rerun
	setState("Items", 0, "Completed", true)
	if runs != 2 {
		t.Fatalf("runs after setting same value = %d; want 2", runs)
	}
}
