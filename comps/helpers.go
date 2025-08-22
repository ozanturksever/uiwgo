//go:build js && wasm

package comps

import (
	"bytes"
	"strconv"
	"sync/atomic"
	"syscall/js"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
	g "maragu.dev/gomponents"
)

var (
	idCounter    uint64
	textRegistry = map[string]func() string{}
	htmlRegistry = map[string]func() g.Node{}
	showRegistry = map[string]showBinder{}
)

type showBinder struct {
	when reactivity.Signal[bool]
	html string
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
	textRegistry[id] = fn
	// Compute initial text without tracking
	initial := fn()
	return g.El("span", g.Attr("data-uiwgo-txt", id), g.Text(initial))
}

// BindHTML creates a reactive HTML container whose innerHTML is re-rendered from a
// gomponents Node-producing function whenever its dependencies change.
// It uses a <div> wrapper as the container.
func BindHTML(fn func() g.Node) g.Node {
	id := nextID("h")
	htmlRegistry[id] = fn
	// Render initial content
	var buf bytes.Buffer
	_ = fn().Render(&buf)
	return g.El("div", g.Attr("data-uiwgo-html", id), g.Raw(buf.String()))
}

// BindHTMLAs is like BindHTML but uses the provided tag name as the container element.
// This is useful to keep valid HTML structure (e.g., <li> inside <ul>).
func BindHTMLAs(tag string, fn func() g.Node, attrs ...g.Node) g.Node {
	id := nextID("h")
	htmlRegistry[id] = fn
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
		if fn, ok := textRegistry[id]; ok {
			// Create a reactive effect that updates textContent
			reactivity.CreateEffect(func() {
				el.Set("textContent", fn())
			})
		}
	}
}

// ShowProps configures the Show control flow.
type ShowProps struct {
	When     reactivity.Signal[bool]
	Children g.Node
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
	showRegistry[id] = showBinder{when: p.When, html: html}

	if p.When.Get() {
		return g.El("span", g.Attr("data-uiwgo-show", id), g.Raw(html))
	}
	return g.El("span", g.Attr("data-uiwgo-show", id))
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
		if fn, ok := htmlRegistry[id]; ok {
			reactivity.CreateEffect(func() {
				var buf bytes.Buffer
				_ = fn().Render(&buf)
				el.Set("innerHTML", buf.String())
				// bind nested newly-rendered content
				attachBinders(el)
			})
		}
	}
}
