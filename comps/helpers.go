//go:build js && wasm

package comps

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"syscall/js"

	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

var (
	idCounter    uint64
	textRegistry = map[string]textBinder{}
	htmlRegistry = map[string]htmlBinder{}
	showRegistry = map[string]showBinder{}
	forRegistry  = map[string]forBinder{}
	indexRegistry = map[string]indexBinder{}
	switchRegistry = map[string]switchBinder{}
	dynamicRegistry = map[string]dynamicBinder{}
	currentMountContainer string // tracks the current mount container during binding
)

// getCurrentMountContainer returns the current mount container ID
func getCurrentMountContainer() string {
	return currentMountContainer
}

// setCurrentMountContainer sets the current mount container ID
func setCurrentMountContainer(containerID string) {
	currentMountContainer = containerID
}

// cleanupRegistriesForContainer removes all registry entries for a specific container
func cleanupRegistriesForContainer(containerID string) {
	// Clean up text registry
	for id, binder := range textRegistry {
		if binder.container == containerID {
			if binder.effect != nil {
				binder.effect.Dispose()
			}
			delete(textRegistry, id)
		}
	}
	
	// Clean up html registry
	for id, binder := range htmlRegistry {
		if binder.container == containerID {
			if binder.effect != nil {
				binder.effect.Dispose()
			}
			delete(htmlRegistry, id)
		}
	}
	
	// Clean up show registry
	for id, binder := range showRegistry {
		if binder.container == containerID {
			delete(showRegistry, id)
		}
	}
	
	// Clean up for registry
	for id, binder := range forRegistry {
		if binder.mountContainer == containerID {
			// Dispose the effect if it exists
			if binder.effect != nil {
				binder.effect.Dispose()
			}
			// Clean up child records
			for _, record := range binder.childRecords {
				if record.cleanup != nil {
					record.cleanup()
				}
			}
			delete(forRegistry, id)
		}
	}
	
	// Clean up index registry
	for id, binder := range indexRegistry {
		if binder.mountContainer == containerID {
			// Dispose the effect if it exists
			if binder.effect != nil {
				binder.effect.Dispose()
			}
			// Clean up child records
			for _, record := range binder.childRecords {
				if record.cleanup != nil {
					record.cleanup()
				}
			}
			delete(indexRegistry, id)
		}
	}
	
	// Clean up switch registry
	for id, binder := range switchRegistry {
		if binder.mountContainer == containerID {
			// Dispose the effect if it exists
			if binder.effect != nil {
				binder.effect.Dispose()
			}
			// Clean up current component
			if binder.currentCleanup != nil {
				binder.currentCleanup()
			}
			delete(switchRegistry, id)
		}
	}
	
	// Clean up dynamic registry
	for id, binder := range dynamicRegistry {
		if binder.mountContainer == containerID {
			// Dispose the effect if it exists
			if binder.effect != nil {
				binder.effect.Dispose()
			}
			// Clean up current component
			if binder.currentCleanup != nil {
				binder.currentCleanup()
			}
			delete(dynamicRegistry, id)
		}
	}
}

type textBinder struct {
	fn        func() string
	container string // elementID of the mounted container
	effect    reactivity.Effect // effect for reactive updates
}

type htmlBinder struct {
	fn        func() g.Node
	container string // elementID of the mounted container
	effect    reactivity.Effect // effect for reactive updates
}

type showBinder struct {
	when      reactivity.Signal[bool]
	html      string
	container string // elementID of the mounted container
}

type forBinder struct {
	items        any // reactivity.Signal[[]T] or func() []T
	keyFn        any // func(T) string
	childrenFn   any // func(item T, index int) g.Node
	childRecords map[string]*childRecord
	container    js.Value
	effect       reactivity.Effect
	mountContainer string // elementID of the mounted container
}

type indexBinder struct {
	items          any // reactivity.Signal[[]T] or func() []T
	childrenFn     any // func(getItem func() T, index int) g.Node
	childRecords   []*childRecord
	container      js.Value
	effect         reactivity.Effect
	mountContainer string // elementID of the mounted container
}

type childRecord struct {
	key       string
	index     int
	element   js.Value
	cleanup   func()
}

type switchBinder struct {
	whenFn         any // reactivity.Signal[any] or func() any
	cases          []matchCase
	fallback       g.Node
	container      js.Value
	effect         reactivity.Effect
	currentCleanup func()
	mountContainer string // elementID of the mounted container
}

type matchCase struct {
	when     any // value or func() bool
	children g.Node
}

type dynamicBinder struct {
	component      any // reactivity.Signal[ComponentFunc] or func() ComponentFunc
	container      js.Value
	effect         reactivity.Effect
	currentCleanup func()
	mountContainer string // elementID of the mounted container
}

func nextID(prefix string) string {
	id := atomic.AddUint64(&idCounter, 1)
	return prefix + strconv.FormatUint(id, 36)
}

// OnMount schedules a function to run after Mount has attached the DOM.
func OnMount(fn func()) g.Node {
	// We return a no-op node so it can be used in gomponents trees.
	enqueueOnMount(fn)
	return g.Group([]g.Node{})
}

// OnCleanup is re-exported from reactivity.
var OnCleanup = reactivity.OnCleanup

// BindText creates a reactive text node placeholder.
// It outputs a <span data-uiwgo-txt="id">initial</span> and registers
// the computation for post-mount reactive updates.
func BindText(fn func() string) g.Node {
	id := nextID("t")
	// Get current mount container from context if available
	containerID := getCurrentMountContainer()
	textRegistry[id] = textBinder{fn: fn, container: containerID}
	// Compute initial text without tracking
	initial := fn()
	return g.El("span", g.Attr("data-uiwgo-txt", id), g.Text(initial))
}

// BindHTML creates a reactive HTML container whose innerHTML is re-rendered from a
// gomponents Node-producing function whenever its dependencies change.
// It uses a <div> wrapper as the container.
func BindHTML(fn func() g.Node) g.Node {
	id := nextID("h")
	// Get current mount container from context if available
	containerID := getCurrentMountContainer()
	htmlRegistry[id] = htmlBinder{fn: fn, container: containerID}
	// Render initial content
	var buf bytes.Buffer
	_ = fn().Render(&buf)
	return g.El("div", g.Attr("data-uiwgo-html", id), g.Raw(buf.String()))
}

// BindHTMLAs is like BindHTML but uses the provided tag name as the container element.
// This is useful to keep valid HTML structure (e.g., <li> inside <ul>).
func BindHTMLAs(tag string, fn func() g.Node, attrs ...g.Node) g.Node {
	id := nextID("h")
	htmlRegistry[id] = htmlBinder{
		fn:        fn,
		container: getCurrentMountContainer(),
	}
	var buf bytes.Buffer
	_ = fn().Render(&buf)
	// Place attrs before the initial HTML content
	nodes := append([]g.Node{g.Attr("data-uiwgo-html", id)}, attrs...)
	nodes = append(nodes, g.Raw(buf.String()))
	return g.El(tag, nodes...)
}

// attachBinders scans the mounted DOM (or a subtree) and attaches reactive behaviors.
func attachBinders(root js.Value) {
	attachTextBindersIn(root)
	attachHTMLBindersIn(root)
	attachShowBindersIn(root)
	attachForBindersIn(root)
	attachIndexBindersIn(root)
	attachSwitchBindersIn(root)
	attachDynamicBindersIn(root)
}

func attachTextBindersIn(root js.Value) {
	nodes := root.Call("querySelectorAll", "[data-uiwgo-txt]")
	ln := nodes.Get("length").Int()
	for i := 0; i < ln; i++ {
		el := nodes.Call("item", i)
		// avoid duplicate attachment
		if el.Call("hasAttribute", "data-uiwgo-bound-text").Bool() {
			continue
		}
		el.Call("setAttribute", "data-uiwgo-bound-text", "1")

		id := el.Call("getAttribute", "data-uiwgo-txt").String()
		if binder, ok := textRegistry[id]; ok {
			// Create a reactive effect that updates textContent
			effect := reactivity.CreateEffect(func() {
				newText := binder.fn()
				el.Set("textContent", newText)
			})
			// Store the effect in the binder for cleanup
			binder.effect = effect
			textRegistry[id] = binder
			
			// Register element with current scope for mutation observer tracking
			if currentScope := reactivity.GetCurrentCleanupScope(); currentScope != nil {
				dom.RegisterElementScope(el, currentScope)
			}
		}
	}
}

// ShowProps configures the Show control flow.
type ShowProps struct {
	When     reactivity.Signal[bool]
	Children g.Node
}

// ForProps configures the For control flow for keyed list rendering.
type ForProps[T any] struct {
	Items    any // reactivity.Signal[[]T] or func() []T
	Key      func(T) string
	Children func(item T, index int) g.Node
}

// IndexProps configures the Index control flow for index-based rendering.
type IndexProps[T any] struct {
	Items    any // reactivity.Signal[[]T] or func() []T
	Children func(getItem func() T, index int) g.Node
}

// SwitchProps configures the Switch control flow for branch selection.
type SwitchProps struct {
	When     any // reactivity.Signal[any] or func() any
	Fallback g.Node
	Children []g.Node // Array of Match nodes
}

// MatchProps configures a Match case within a Switch.
type MatchProps struct {
	When     any // value or func() bool for matching
	Children g.Node
}

// DynamicProps configures the Dynamic control flow for reactive component rendering.
type DynamicProps struct {
	Component any // reactivity.Signal[ComponentFunc] or func() ComponentFunc
}

// Show renders its children only when the When signal is true.
// It outputs a <span data-uiwgo-show="id">[initial child html]</span>
// and attaches a reactive toggle after mount.
func Show(p ShowProps) g.Node {
	id := nextID("s")
	// Pre-render children to HTML for quick toggle
	var buf bytes.Buffer
	_ = p.Children.Render(&buf)
	html := buf.String()
	containerID := getCurrentMountContainer()
	showRegistry[id] = showBinder{when: p.When, html: html, container: containerID}

	if p.When.Get() {
		return g.El("span", g.Attr("data-uiwgo-show", id), g.Raw(html))
	}
	return g.El("span", g.Attr("data-uiwgo-show", id))
}

// For renders a list of items with keyed reconciliation.
// It outputs a <div data-uiwgo-for="id"></div> container and manages
// efficient insertion/removal/move operations based on keys.
func For[T any](p ForProps[T]) g.Node {
	id := nextID("f")
	containerID := getCurrentMountContainer()
	forRegistry[id] = forBinder{
		items:          p.Items,
		keyFn:          p.Key,
		childrenFn:     p.Children,
		childRecords:   make(map[string]*childRecord),
		mountContainer: containerID,
	}
	return g.El("div", g.Attr("data-uiwgo-for", id))
}

// Index renders a list of items with index-based reconciliation.
// It outputs a <div data-uiwgo-index="id"></div> container and maintains
// stable child components that reactively track their items by index.
func Index[T any](p IndexProps[T]) g.Node {
	id := nextID("i")
	containerID := getCurrentMountContainer()
	indexRegistry[id] = indexBinder{
		items:          p.Items,
		childrenFn:     p.Children,
		childRecords:   make([]*childRecord, 0),
		mountContainer: containerID,
	}
	return g.El("div", g.Attr("data-uiwgo-index", id))
}

// Switch renders one of its Match children based on the When condition.
// It outputs a <div data-uiwgo-switch="id"></div> container and reactively
// switches between branches with proper cleanup.
func Switch(p SwitchProps) g.Node {
	id := nextID("sw")
	// Extract match cases from children - will be populated by extractMatchCases
	cases := make([]matchCase, 0)
	containerID := getCurrentMountContainer()
	switchRegistry[id] = switchBinder{
		whenFn:         p.When,
		cases:          cases,
		fallback:       p.Fallback,
		mountContainer: containerID,
	}
	// Include the Match children as templates inside the switch container
	children := []g.Node{g.Attr("data-uiwgo-switch", id)}
	children = append(children, p.Children...)
	return g.El("div", children...)
}

// Match creates a case for use within a Switch component.
// Note: This is a simplified implementation. In practice, you'd want
// a more sophisticated way to collect Match cases within Switch.
func Match(p MatchProps) g.Node {
	// Store match data in a data attribute for the switch binder to read
	id := nextID("m")
	return g.El("template", 
		g.Attr("data-uiwgo-match", id),
		g.Attr("data-match-when", fmt.Sprintf("%v", p.When)),
		p.Children,
	)
}

// Dynamic renders a component reactively based on a signal or function.
// It outputs a <div data-uiwgo-dynamic="id"></div> container and reactively
// switches between different components with proper cleanup.
func Dynamic(p DynamicProps) g.Node {
	id := nextID("dyn")
	containerID := getCurrentMountContainer()
	dynamicRegistry[id] = dynamicBinder{
		component:      p.Component,
		mountContainer: containerID,
	}
	return g.El("div", g.Attr("data-uiwgo-dynamic", id))
}

func attachShowBindersIn(root js.Value) {
	nodes := root.Call("querySelectorAll", "[data-uiwgo-show]")
	ln := nodes.Get("length").Int()
	for i := 0; i < ln; i++ {
		el := nodes.Call("item", i)
		// avoid duplicate attachment
		if el.Call("hasAttribute", "data-uiwgo-bound-show").Bool() {
			continue
		}
		el.Call("setAttribute", "data-uiwgo-bound-show", "1")

		id := el.Call("getAttribute", "data-uiwgo-show").String()
		if b, ok := showRegistry[id]; ok {
			// Track visibility and update innerHTML
			var visible bool
			reactivity.CreateEffect(func() {
				v := b.when.Get()
				if v && !visible {
					el.Set("innerHTML", b.html)
					// new content may contain binders
					attachBinders(el)
					visible = true
				} else if !v && visible {
					el.Set("innerHTML", "")
					visible = false
				}
			})
		}
	}
}

func attachHTMLBindersIn(root js.Value) {
	nodes := root.Call("querySelectorAll", "[data-uiwgo-html]")
	ln := nodes.Get("length").Int()
	for i := 0; i < ln; i++ {
		el := nodes.Call("item", i)
		// avoid duplicate attachment
		if el.Call("hasAttribute", "data-uiwgo-bound-html").Bool() {
			continue
		}
		el.Call("setAttribute", "data-uiwgo-bound-html", "1")

		id := el.Call("getAttribute", "data-uiwgo-html").String()
		if binder, ok := htmlRegistry[id]; ok {
			effect := reactivity.CreateEffect(func() {
				var buf bytes.Buffer
				_ = binder.fn().Render(&buf)
				el.Set("innerHTML", buf.String())
				// bind nested newly-rendered content
				attachBinders(el)
			})
			// Store the effect in the binder for cleanup
			binder.effect = effect
			htmlRegistry[id] = binder
			
			// Register element with current scope for mutation observer tracking
			if currentScope := reactivity.GetCurrentCleanupScope(); currentScope != nil {
				dom.RegisterElementScope(el, currentScope)
			}
		}
	}
}

func attachForBindersIn(root js.Value) {
	nodes := root.Call("querySelectorAll", "[data-uiwgo-for]")
	ln := nodes.Get("length").Int()
	for i := 0; i < ln; i++ {
		el := nodes.Call("item", i)
		// avoid duplicate attachment
		if el.Call("hasAttribute", "data-uiwgo-bound-for").Bool() {
			continue
		}
		el.Call("setAttribute", "data-uiwgo-bound-for", "1")

		id := el.Call("getAttribute", "data-uiwgo-for").String()
		if binder, ok := forRegistry[id]; ok {
			binder.container = el
			forRegistry[id] = binder
			// Create reactive effect for list reconciliation
			effect := reactivity.CreateEffect(func() {
				reconcileForList(id)
			})
			binder.effect = effect
			forRegistry[id] = binder
			// Register cleanup
			reactivity.OnCleanup(func() {
				if b, exists := forRegistry[id]; exists {
					for _, record := range b.childRecords {
						if record.cleanup != nil {
							record.cleanup()
						}
					}
					if b.effect != nil {
						b.effect.Dispose()
					}
					delete(forRegistry, id)
				}
			})
		}
	}
}

func attachIndexBindersIn(root js.Value) {
	nodes := root.Call("querySelectorAll", "[data-uiwgo-index]")
	ln := nodes.Get("length").Int()
	for i := 0; i < ln; i++ {
		el := nodes.Call("item", i)
		// avoid duplicate attachment
		if el.Call("hasAttribute", "data-uiwgo-bound-index").Bool() {
			continue
		}
		el.Call("setAttribute", "data-uiwgo-bound-index", "1")

		id := el.Call("getAttribute", "data-uiwgo-index").String()
		if binder, ok := indexRegistry[id]; ok {
			binder.container = el
			indexRegistry[id] = binder
			// Create reactive effect for list reconciliation
			effect := reactivity.CreateEffect(func() {
				if b, exists := indexRegistry[id]; exists {
					reconcileIndexList(&b)
				}
			})
			binder.effect = effect
			indexRegistry[id] = binder
			// Register cleanup
			reactivity.OnCleanup(func() {
				if b, exists := indexRegistry[id]; exists {
					for _, record := range b.childRecords {
						if record != nil && record.cleanup != nil {
							record.cleanup()
						}
					}
					if b.effect != nil {
						b.effect.Dispose()
					}
					delete(indexRegistry, id)
				}
			})
		}
	}
}

func attachSwitchBindersIn(root js.Value) {
	nodes := root.Call("querySelectorAll", "[data-uiwgo-switch]")
	ln := nodes.Get("length").Int()
	for i := 0; i < ln; i++ {
		el := nodes.Call("item", i)
		// avoid duplicate attachment
		if el.Call("hasAttribute", "data-uiwgo-bound-switch").Bool() {
			continue
		}
		el.Call("setAttribute", "data-uiwgo-bound-switch", "1")

		id := el.Call("getAttribute", "data-uiwgo-switch").String()
		if binder, ok := switchRegistry[id]; ok {
			// Extract match cases from template children
			binder.cases = extractMatchCases(el)
			binder.container = el
			switchRegistry[id] = binder
			// Create reactive effect for branch switching
			effect := reactivity.CreateEffect(func() {
				reconcileSwitchBranch(id)
			})
			binder.effect = effect
			switchRegistry[id] = binder
			// Register cleanup
			reactivity.OnCleanup(func() {
				if b, exists := switchRegistry[id]; exists {
					// Cleanup current branch
					if b.currentCleanup != nil {
						b.currentCleanup()
					}
					// Dispose effect
					if b.effect != nil {
						b.effect.Dispose()
					}
					// Remove from registry
					delete(switchRegistry, id)
				}
			})
		}
	}
}

func attachDynamicBindersIn(root js.Value) {
	nodes := root.Call("querySelectorAll", "[data-uiwgo-dynamic]")
	for i := 0; i < nodes.Get("length").Int(); i++ {
		node := nodes.Call("item", i)
		id := node.Call("getAttribute", "data-uiwgo-dynamic").String()
		binder, exists := dynamicRegistry[id]
		if !exists {
			continue
		}
		// Prevent duplicate attachment
		if binder.effect != nil {
			continue
		}
		binder.container = node
		binder.effect = reactivity.CreateEffect(func() {
			reconcileDynamicComponent(&binder)
		})
		// Register cleanup
		reactivity.OnCleanup(func() {
			// Cleanup current component
			if binder.currentCleanup != nil {
				binder.currentCleanup()
			}
			// Dispose effect
			if binder.effect != nil {
				binder.effect.Dispose()
			}
			// Remove from registry
			delete(dynamicRegistry, id)
		})
		dynamicRegistry[id] = binder
	}
}

// reconcileForList implements keyed diffing for For components
func reconcileForList(id string) {
	binder, ok := forRegistry[id]
	if !ok {
		return
	}

	// Get current items using reflection to handle both Signal and func types
	items := getItemsFromSource(binder.items)
	if items == nil {
		return
	}



	// Build new keys
	newKeys := make([]string, len(items))
	for i, item := range items {
		key := callKeyFunc(binder.keyFn, item)
		newKeys[i] = key
	}

	// Track which keys are new, removed, or moved
	oldRecords := binder.childRecords
	newRecords := make(map[string]*childRecord)

	// Remove items that are no longer present
	for key, record := range oldRecords {
		found := false
		for _, newKey := range newKeys {
			if newKey == key {
				found = true
				break
			}
		}
		if !found {
			// Remove from DOM and cleanup
			if !record.element.IsUndefined() {
				record.element.Call("remove")
			}
			if record.cleanup != nil {
				record.cleanup()
			}
		}
	}

	// Process new items and reorder
	for i, key := range newKeys {
		if record, exists := oldRecords[key]; exists {
			// Reuse existing record
			record.index = i
			newRecords[key] = record
		} else {
			// Create new item
			item := items[i]
			element, cleanup := createItemElement(binder.childrenFn, item, i)
			newRecords[key] = &childRecord{
				key:     key,
				index:   i,
				element: element,
				cleanup: cleanup,
			}
		}
	}

	// Reorder DOM elements to match new order
	container := binder.container
	
	// Clear container first to avoid duplication
	for container.Get("firstChild").Truthy() {
		container.Call("removeChild", container.Get("firstChild"))
	}
	
	// Append elements in correct order
	for _, key := range newKeys {
		record := newRecords[key]
		container.Call("appendChild", record.element)
	}

	// Update registry
	binder.childRecords = newRecords
	forRegistry[id] = binder
}

// extractMatchCases extracts match cases from template children within a switch container
func extractMatchCases(container js.Value) []matchCase {
	cases := make([]matchCase, 0)
	templates := container.Call("querySelectorAll", "template[data-uiwgo-match]")
	for i := 0; i < templates.Get("length").Int(); i++ {
		template := templates.Call("item", i)
		whenAttr := template.Call("getAttribute", "data-match-when").String()
		// For simplicity, we'll store the when value as a string
		// In a real implementation, you'd want more sophisticated matching
		cases = append(cases, matchCase{
			when:     whenAttr,
			children: nil, // We'll render from template content when needed
		})
	}
	return cases
}

// reconcileSwitchBranch handles reactive branch switching for Switch components
func reconcileSwitchBranch(id string) {
	binder, exists := switchRegistry[id]
	if !exists {
		return
	}

	// Get current when value
	currentWhen := getValueFromSource(binder.whenFn)



	// Clean up previous branch
	if binder.currentCleanup != nil {
		binder.currentCleanup()
		binder.currentCleanup = nil
	}

	// Clear container but preserve templates
	// Remove all non-template children
	children := binder.container.Get("children")
	for i := children.Get("length").Int() - 1; i >= 0; i-- {
		child := children.Call("item", i)
		if child.Get("tagName").String() != "TEMPLATE" {
			binder.container.Call("removeChild", child)
		}
	}

	// Find matching case
	for _, matchCase := range binder.cases {
		// Convert both to strings for comparison
		currentWhenStr := fmt.Sprintf("%v", currentWhen)
		matchWhenStr := fmt.Sprintf("%v", matchCase.when)
		if currentWhenStr == matchWhenStr {
			
			// Found a match - get the template content
			templates := binder.container.Call("querySelectorAll", 
				fmt.Sprintf("template[data-match-when='%s']", matchWhenStr))
			if templates.Get("length").Int() > 0 {
				template := templates.Call("item", 0)
				// Clone template content
				content := template.Get("content").Call("cloneNode", true)
				binder.container.Call("appendChild", content)
				// Attach binders to new content
				attachBinders(binder.container)
				return
			}
			break
		}
	}

	// No match found, use fallback
	if binder.fallback != nil {
		var buf bytes.Buffer
		binder.fallback.Render(&buf)
		binder.container.Set("innerHTML", buf.String())
		// Attach binders to fallback content
		attachBinders(binder.container)
	}

	// Update registry
	switchRegistry[id] = binder
}

// getValueFromSource extracts value from Signal or function
func getValueFromSource(source any) any {
	if source == nil {
		return nil
	}
	// Check if it's a Signal using reflection to handle generic types
	val := reflect.ValueOf(source)
	if val.IsValid() {
		// Check if it has a Get() method
		getMethod := val.MethodByName("Get")
		if getMethod.IsValid() && getMethod.Type().NumIn() == 0 && getMethod.Type().NumOut() == 1 {
			results := getMethod.Call(nil)
			if len(results) > 0 {
				return results[0].Interface()
			}
		}
		// Check if it's a function
		if val.Kind() == reflect.Func {
			results := val.Call(nil)
			if len(results) > 0 {
				return results[0].Interface()
			}
		}
	}
	return source
}

// reconcileDynamicComponent handles reactive component switching for Dynamic components
func reconcileDynamicComponent(binder *dynamicBinder) {
	// Get current component function
	currentComponent := getComponentFromSource(binder.component)

	// Clean up previous component
	if binder.currentCleanup != nil {
		binder.currentCleanup()
		binder.currentCleanup = nil
	}

	// Clear container
	binder.container.Set("innerHTML", "")

	// Render new component if available
	if currentComponent != nil {
		// Create a temporary container for the component
		tempDiv := js.Global().Get("document").Call("createElement", "div")
		
		// Render the component to HTML
		var buf bytes.Buffer
		_ = currentComponent().Render(&buf)
		tempDiv.Set("innerHTML", buf.String())
		
		// Move all children from temp div to actual container
		for tempDiv.Get("firstChild").Truthy() {
			child := tempDiv.Get("firstChild")
			binder.container.Call("appendChild", child)
		}
		
		// Attach binders to the new content (excluding dynamic to prevent recursion)
		attachTextBindersIn(binder.container)
		attachHTMLBindersIn(binder.container)
		attachShowBindersIn(binder.container)
		attachForBindersIn(binder.container)
		attachIndexBindersIn(binder.container)
		attachSwitchBindersIn(binder.container)
	}
}

// getComponentFromSource extracts ComponentFunc from Signal or function
func getComponentFromSource(source any) func() g.Node {
	if source == nil {
		return nil
	}
	
	// Check if it's a Signal[func() g.Node] specifically
	if signal, ok := source.(reactivity.Signal[func() g.Node]); ok {
		return signal.Get()
	}
	
	// Try to get the value using reflection
	v := reflect.ValueOf(source)
	if !v.IsValid() {
		return nil
	}
	
	// If it's a function, call it
	if v.Kind() == reflect.Func {
		result := v.Call(nil)
		if len(result) > 0 && result[0].IsValid() {
			if fn, ok := result[0].Interface().(func() g.Node); ok {
				return fn
			}
		}
	}
	
	// If it's already a ComponentFunc
	if fn, ok := source.(func() g.Node); ok {
		return fn
	}
	
	return nil
}

// reconcileIndexList implements index-based reconciliation for Index components
func reconcileIndexList(binder *indexBinder) {
	// Get current items
	items := getItemsFromSource(binder.items)
	if items == nil {
		return
	}

	// Adjust child records to match new length
	oldLen := len(binder.childRecords)
	newLen := len(items)

	// Remove excess records
	for i := newLen; i < oldLen; i++ {
		if binder.childRecords[i] != nil {
			if !binder.childRecords[i].element.IsUndefined() {
				binder.childRecords[i].element.Call("remove")
			}
			if binder.childRecords[i].cleanup != nil {
				binder.childRecords[i].cleanup()
			}
		}
	}

	// Resize slice
	if newLen < oldLen {
		binder.childRecords = binder.childRecords[:newLen]
	} else {
		// Extend slice with nil entries
		for i := oldLen; i < newLen; i++ {
			binder.childRecords = append(binder.childRecords, nil)
		}
	}

	// Create or update records
	for i := 0; i < newLen; i++ {
		if binder.childRecords[i] == nil {
			// Create new record with getter function
			getItem := createItemGetter(binder.items, i)
			element, cleanup := createIndexItemElement(binder.childrenFn, getItem, i)
			binder.childRecords[i] = &childRecord{
				key:     strconv.Itoa(i),
				index:   i,
				element: element,
				cleanup: cleanup,
			}
		}
	}

	// Clear container and append elements in correct order
	for binder.container.Get("firstChild").Truthy() {
		binder.container.Call("removeChild", binder.container.Get("firstChild"))
	}
	
	// Append elements in correct order
	for _, record := range binder.childRecords {
		if record != nil {
			binder.container.Call("appendChild", record.element)
		}
	}
}

// getItemsFromSource extracts items from either a Signal or a function
func getItemsFromSource(source any) []any {
	if source == nil {
		return nil
	}

	v := reflect.ValueOf(source)
	if v.Kind() == reflect.Func {
		// Call the function to get items
		results := v.Call(nil)
		if len(results) == 0 {
			return nil
		}
		sliceVal := results[0]
		return interfaceSliceFromReflect(sliceVal)
	}

	// Assume it's a Signal - call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return nil
	}
	results := getMethod.Call(nil)
	if len(results) == 0 {
		return nil
	}
	sliceVal := results[0]
	return interfaceSliceFromReflect(sliceVal)
}

// interfaceSliceFromReflect converts a reflect.Value slice to []any
func interfaceSliceFromReflect(sliceVal reflect.Value) []any {
	if sliceVal.Kind() != reflect.Slice {
		return nil
	}
	len := sliceVal.Len()
	result := make([]any, len)
	for i := 0; i < len; i++ {
		result[i] = sliceVal.Index(i).Interface()
	}
	return result
}

// callKeyFunc calls the key function with the given item
func callKeyFunc(keyFn any, item any) string {
	if keyFn == nil {
		return ""
	}
	v := reflect.ValueOf(keyFn)
	if v.Kind() != reflect.Func {
		return ""
	}
	args := []reflect.Value{reflect.ValueOf(item)}
	results := v.Call(args)
	if len(results) == 0 {
		return ""
	}
	return results[0].String()
}

// createItemElement creates a DOM element for a For item
func createItemElement(childrenFn any, item any, index int) (js.Value, func()) {
	if childrenFn == nil {
		return js.Undefined(), nil
	}

	v := reflect.ValueOf(childrenFn)
	if v.Kind() != reflect.Func {
		return js.Undefined(), nil
	}

	// Call childrenFn(item, index)
	args := []reflect.Value{
		reflect.ValueOf(item),
		reflect.ValueOf(index),
	}
	results := v.Call(args)
	if len(results) == 0 {
		return js.Undefined(), nil
	}

	// Render the Node to HTML
	node := results[0].Interface().(g.Node)
	var buf bytes.Buffer
	_ = node.Render(&buf)
	html := buf.String()

	// Create a wrapper div and set innerHTML
	doc := js.Global().Get("document")
	wrapper := doc.Call("createElement", "div")
	wrapper.Set("innerHTML", html)

	// Extract the first child as the actual element
	var element js.Value
	if wrapper.Get("firstElementChild").Truthy() {
		element = wrapper.Get("firstElementChild")
	} else {
		element = wrapper
	}

	// Attach only text binders to the new element to avoid re-attaching Index binders
	// Note: Only attach if not already attached to prevent duplication
	if !element.Call("hasAttribute", "data-uiwgo-attached").Bool() {
		element.Call("setAttribute", "data-uiwgo-attached", "1")
		attachTextBindersIn(element)
		attachHTMLBindersIn(element)
		attachShowBindersIn(element)
		// Skip attachIndexBindersIn to prevent recursive attachment
		attachSwitchBindersIn(element)
		attachDynamicBindersIn(element)
	}

	// Create cleanup function
	cleanup := func() {
		// Cleanup will be handled by the reactive system
		// when effects are disposed
	}

	return element, cleanup
}

// createIndexItemElement creates a DOM element for an Index item
func createIndexItemElement(childrenFn any, getItem func() any, index int) (js.Value, func()) {
	if childrenFn == nil {
		return js.Undefined(), nil
	}

	v := reflect.ValueOf(childrenFn)
	if v.Kind() != reflect.Func {
		return js.Undefined(), nil
	}

	// Create a typed wrapper function that matches the expected signature
	funcType := v.Type()
	if funcType.NumIn() != 2 {
		return js.Undefined(), nil
	}

	// Get the expected type of the first parameter (getItem function)
	getItemType := funcType.In(0)
	if getItemType.Kind() != reflect.Func {
		return js.Undefined(), nil
	}

	// Create a typed wrapper function
	wrapperFn := reflect.MakeFunc(getItemType, func(args []reflect.Value) []reflect.Value {
		item := getItem()
		if item == nil {
			// Return zero value of the expected return type
			returnType := getItemType.Out(0)
			return []reflect.Value{reflect.Zero(returnType)}
		}
		return []reflect.Value{reflect.ValueOf(item)}
	})

	// Call childrenFn(typedGetItem, index)
	args := []reflect.Value{
		wrapperFn,
		reflect.ValueOf(index),
	}
	results := v.Call(args)
	if len(results) == 0 {
		return js.Undefined(), nil
	}

	// Render the Node to HTML
	node := results[0].Interface().(g.Node)
	var buf bytes.Buffer
	_ = node.Render(&buf)
	html := buf.String()

	// Create a wrapper div and set innerHTML
	doc := js.Global().Get("document")
	wrapper := doc.Call("createElement", "div")
	wrapper.Set("innerHTML", html)

	// Extract the first child as the actual element
	var element js.Value
	if wrapper.Get("firstElementChild").Truthy() {
		element = wrapper.Get("firstElementChild")
	} else {
		element = wrapper
	}

	// Attach binders to the new element
	// Note: Only attach if not already attached to prevent duplication
	if !element.Call("hasAttribute", "data-uiwgo-attached").Bool() {
		element.Call("setAttribute", "data-uiwgo-attached", "1")
		// Skip attachTextBindersIn to prevent duplicate text elements
		attachHTMLBindersIn(element)
		attachShowBindersIn(element)
		attachForBindersIn(element)
		// Skip attachIndexBindersIn to prevent recursive attachment
		attachSwitchBindersIn(element)
		attachDynamicBindersIn(element)
	}

	// Create cleanup function
	cleanup := func() {
		// Cleanup will be handled by the reactive system
	}

	return element, cleanup
}

// createItemGetter creates a getter function for Index components
func createItemGetter(source any, index int) func() any {
	return func() any {
		items := getItemsFromSource(source)
		if items == nil || index >= len(items) {
			return nil
		}
		return items[index]
	}
}
