package reactivity

import (
	"fmt"
	"reflect"
)

// Store is a fine-grained reactive state container for nested data.
//
// Notes on design:
// - Get() returns a non-reactive snapshot of the entire state. Using the
//   returned struct's fields will NOT register fine-grained dependencies due to
//   Go's lack of field access interception.
// - For fine-grained reactivity, use Select(path...).Get() to read a specific
//   nested property. Select returns a Signal for that property, so effects only
//   subscribe to exactly what they read.
//
// Path rules: strings address struct fields by name; ints address slice indices.
// Example: Select[bool]("Todos", 0, "Completed")
//
// SetState uses the same path rules with the last argument being the new value:
//   setState("Todos", 0, "Completed", true)
// or replace the entire slice:
//   setState("Todos", newTodos)
//
// This mirrors SolidJS createStore behavior within Go's constraints.
//
// T should generally be a struct whose exported fields represent state.
// Slices of structs are supported; maps are not supported in this MVP.
// Pointers are dereferenced on build; nested pointers should be avoided.
// Unexported fields are ignored.
//
// Zero values and DeepEqual checks in signals prevent redundant effect runs.

type Store[T any] interface {
	// Get returns a snapshot of the entire state (non-reactive).
	Get() T
	// Select returns a Signal[any] for the nested property addressed by path.
	// Use Adapt[V] to cast it to a typed signal.
	Select(path ...any) Signal[any]
	// SelectLen returns a Signal[int] representing the length of the slice/array at the given path.
	SelectLen(path ...any) Signal[int]
}

type store[T any] struct {
	root *storeNode
	typ  reflect.Type
}

// SelectLen returns a Signal[int] for the length of the slice/array at path.
func (s *store[T]) SelectLen(path ...any) Signal[int] {
	n := s.root
	for i, p := range path {
		switch key := p.(type) {
		case string:
			if n.fields == nil {
				panic(fmt.Sprintf("SelectLen: segment %d ('%v') does not point to a struct", i, key))
			}
			nn := n.fields[key]
			if nn == nil {
				// Create empty slice node with length 0 if accessed early
				nn = &storeNode{slen: CreateSignal(0)}
				n.fields[key] = nn
			}
			n = nn
		case int:
			if n.elems == nil {
				panic(fmt.Sprintf("SelectLen: segment %d (%v) does not point to a slice/array", i, key))
			}
			n = n.elems[key]
		default:
			panic(fmt.Sprintf("SelectLen: unsupported path segment %T", p))
		}
	}
	if n.slen == nil {
		n.slen = CreateSignal(len(n.elems))
	}
	return n.slen
}

type storeNode struct {
	// kind/type of the node's data
	typ reflect.Type
	// leaf is non-nil for leaf nodes (non-struct, non-slice)
	leaf Signal[any]
	// struct fields
	fields map[string]*storeNode
	// slice elements
	elems []*storeNode
	// slice length signal (only for slice/array nodes)
	slen Signal[int]
}

// CreateStore builds a reactive store out of the initial state.
// It returns the store and a setState function. The setState accepts a path
// (strings for fields, ints for indices) followed by the new value as the
// last argument: setState("Todos", 0, "Completed", true).
func CreateStore[T any](initialState T) (Store[T], func(...any)) {
	val := reflect.ValueOf(initialState)
	typ := reflect.TypeOf(initialState)
	root := buildNode(val)
	st := &store[T]{root: root, typ: typ}

	setter := func(args ...any) {
		if len(args) == 0 {
			panic("setState requires at least a value")
		}
		newVal := args[len(args)-1]
		path := args[:len(args)-1]
		if len(path) == 0 {
			// Replace entire root
			st.assignNodeValue(st.root, reflect.ValueOf(newVal))
			return
		}
		n := st.root
		for i, p := range path {
			switch key := p.(type) {
			case string:
				if n.fields == nil {
					panic(fmt.Sprintf("path at segment %d ('%v') does not point to a struct", i, key))
				}
				nn, ok := n.fields[key]
				if !ok {
					// If missing (e.g., setting new field on struct), create a node on demand based on the incoming value type.
					nn = buildNode(reflect.ValueOf(newVal))
					n.fields[key] = nn
				}
				n = nn
			case int:
				if n.elems == nil {
					panic(fmt.Sprintf("path at segment %d (%v) does not point to a slice/array", i, key))
				}
				idx := key
				if idx < 0 {
					panic("negative index in setState path")
				}
				// Expand elems if necessary
				for len(n.elems) <= idx {
					// Create properly typed element nodes based on the slice element type
					var child *storeNode
					if n.typ != nil && (n.typ.Kind() == reflect.Slice || n.typ.Kind() == reflect.Array) {
						et := n.typ.Elem()
						child = buildNode(reflect.Zero(et))
					} else {
						child = &storeNode{leaf: CreateSignal(any(nil))}
					}
					n.elems = append(n.elems, child)
				}
				if n.slen != nil {
					n.slen.Set(len(n.elems))
				}
				n = n.elems[idx]
			default:
				panic(fmt.Sprintf("unsupported path segment type %T; use string (field) or int (index)", p))
			}
		}
		st.assignNodeValue(n, reflect.ValueOf(newVal))
	}

	return st, setter
}

func buildNode(v reflect.Value) *storeNode {
	// Dereference pointers
	for v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// Create a leaf nil signal
			return &storeNode{typ: nil, leaf: CreateSignal(any(nil))}
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return &storeNode{typ: nil, leaf: CreateSignal(any(nil))}
	}
	t := v.Type()
	switch v.Kind() {
	case reflect.Struct:
		n := &storeNode{typ: t, fields: make(map[string]*storeNode)}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			// Only exported fields
			if f.PkgPath != "" {
				continue
			}
			fv := v.Field(i)
			n.fields[f.Name] = buildNode(fv)
		}
		return n
	case reflect.Slice, reflect.Array:
		l := v.Len()
		elems := make([]*storeNode, l)
		for i := 0; i < l; i++ {
			elems[i] = buildNode(v.Index(i))
		}
		return &storeNode{typ: t, elems: elems, slen: CreateSignal(l)}
	default:
		// leaf
		return &storeNode{typ: t, leaf: CreateSignal(any(v.Interface()))}
	}
}

// assignNodeValue updates the node's signals to match the new value. It tries
// to reuse existing child nodes where possible and only updates leaf signals
// when values actually change (delegated to Signal.Set DeepEqual).
func (s *store[T]) assignNodeValue(n *storeNode, val reflect.Value) {
	// Normalize val
	for val.IsValid() && val.Kind() == reflect.Ptr {
		if val.IsNil() {
			// nil pointer -> set leaf to nil
			if n.leaf == nil {
				n.leaf = CreateSignal(any(nil))
			} else {
				n.leaf.Set(any(nil))
			}
			return
		}
		val = val.Elem()
	}
	if !val.IsValid() {
		if n.leaf == nil {
			n.leaf = CreateSignal(any(nil))
		} else {
			n.leaf.Set(any(nil))
		}
		return
	}
	switch val.Kind() {
	case reflect.Struct:
		if n.fields == nil {
			n.fields = make(map[string]*storeNode)
			n.leaf = nil
			n.elems = nil
		}
		n.typ = val.Type()
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" { // unexported
				continue
			}
			fv := val.Field(i)
			child, ok := n.fields[f.Name]
			if !ok {
				child = buildNode(fv)
				n.fields[f.Name] = child
			} else {
				s.assignNodeValue(child, fv)
			}
		}
		// Remove fields that no longer exist in type? Not necessary for static structs.
	case reflect.Slice, reflect.Array:
		l := val.Len()
		if n.elems == nil {
			n.elems = make([]*storeNode, 0, l)
			n.fields = nil
			n.leaf = nil
			if n.slen == nil {
				n.slen = CreateSignal(0)
			}
		}
		n.typ = val.Type()
		// Adjust length
		if len(n.elems) > l {
			n.elems = n.elems[:l]
		}
		for i := 0; i < l; i++ {
			if i < len(n.elems) && n.elems[i] != nil {
				s.assignNodeValue(n.elems[i], val.Index(i))
				continue
			}
			n.elems = append(n.elems, buildNode(val.Index(i)))
		}
		if n.slen == nil {
			n.slen = CreateSignal(l)
		} else {
			n.slen.Set(l)
		}
	default:
		// Leaf
		if n.leaf == nil {
			n.leaf = CreateSignal(any(val.Interface()))
			return
		}
		n.leaf.Set(any(val.Interface()))
	}
}

// Get builds a snapshot value of the entire tree.
func (s *store[T]) Get() T {
	out := reflect.New(s.typ)
	buildSnapshot(s.root, out.Elem())
	return out.Elem().Interface().(T)
}

func buildSnapshot(n *storeNode, dst reflect.Value) {
	// Ensure dst is addressable
	switch dst.Kind() {
	case reflect.Struct:
		for i := 0; i < dst.NumField(); i++ {
			f := dst.Type().Field(i)
			if f.PkgPath != "" {
				continue
			}
			child := n.fields[f.Name]
			if child == nil {
				continue
			}
			buildSnapshot(child, dst.Field(i))
		}
	case reflect.Slice:
		// For slices, construct slice of the same length as node.elems
		l := len(n.elems)
		dst.Set(reflect.MakeSlice(dst.Type(), l, l))
		for i := 0; i < l; i++ {
			buildSnapshot(n.elems[i], dst.Index(i))
		}
	default:
		// Leaf
		if n.leaf == nil {
			// leave zero
			return
		}
		v := n.leaf.Get() // Register dependency on this leaf if snapshot is read in an effect
		// However, note: building full snapshot touches many leaves; prefer Select for fine-grained.
		if v == nil {
			// keep zero value
			return
		}
		val := reflect.ValueOf(v)
		if val.Type().AssignableTo(dst.Type()) {
			dst.Set(val)
			return
		}
		// Try convert if possible
		if val.Type().ConvertibleTo(dst.Type()) {
			dst.Set(val.Convert(dst.Type()))
		}
	}
}

// Select returns a Signal[any] for a nested property.
func (s *store[T]) Select(path ...any) Signal[any] {
	n := s.root
	for i, p := range path {
		switch key := p.(type) {
		case string:
			if n.fields == nil {
				panic(fmt.Sprintf("Select: segment %d ('%v') does not point to a struct", i, key))
			}
			nn := n.fields[key]
			if nn == nil {
				// Lazily create a typed field node based on the parent struct type if available
				if n.typ != nil && n.typ.Kind() == reflect.Struct {
					if f, ok := n.typ.FieldByName(key); ok && f.PkgPath == "" {
						nn = buildNode(reflect.Zero(f.Type))
					}
				}
				if nn == nil {
					// Fallback to an empty leaf to allow late population
					nn = &storeNode{leaf: CreateSignal(any(nil))}
				}
				n.fields[key] = nn
			}
			n = nn
		case int:
			if n.elems == nil {
				panic(fmt.Sprintf("Select: segment %d (%v) does not point to a slice/array", i, key))
			}
			idx := key
			if idx < 0 {
				panic("Select: negative index")
			}
			for len(n.elems) <= idx {
				// Create properly typed element nodes based on the slice element type
				var child *storeNode
				if n.typ != nil && (n.typ.Kind() == reflect.Slice || n.typ.Kind() == reflect.Array) {
					et := n.typ.Elem()
					child = buildNode(reflect.Zero(et))
				} else {
					child = &storeNode{leaf: CreateSignal(any(nil))}
				}
				n.elems = append(n.elems, child)
			}
			if n.slen != nil {
				n.slen.Set(len(n.elems))
			}
			n = n.elems[idx]
		default:
			panic(fmt.Sprintf("Select: unsupported path segment type %T", p))
		}
	}
	// If n is non-leaf (struct/slice), we provide a memo that snapshots it.
	if n.leaf == nil {
		return CreateMemo(func() any {
			if n.typ == nil {
				return nil
			}
			dst := reflect.New(n.typ).Elem()
			buildSnapshot(n, dst)
			return dst.Interface()
		})
	}
	return n.leaf
}

// Adapt wraps a Signal[any] into a typed Signal[V].
type adapter[V any] struct {
	inner Signal[any]
}

// Adapt converts a generic any-based signal to a typed one.
func Adapt[V any](s Signal[any]) Signal[V] { return &adapter[V]{inner: s} }

func (t *adapter[V]) Get() V {
	v := t.inner.Get()
	if v == nil {
		var zero V
		return zero
	}
	vv, ok := v.(V)
	if ok {
		return vv
	}
	// Attempt reflect convert
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf((*V)(nil)).Elem()
	if rv.IsValid() && rv.Type().ConvertibleTo(rt) {
		out := rv.Convert(rt)
		return out.Interface().(V)
	}
	var zero V
	return zero
}

func (t *adapter[V]) Set(v V) { t.inner.Set(any(v)) }
