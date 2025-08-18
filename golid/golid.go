// golid.go
// A minimal reactive UI toolkit for Go+WASM using gomponents

package golid

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/google/uuid"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ---------------------------
// 🔧 Core Types & Interfaces
// ---------------------------

type JsCallback func(this js.Value, args []js.Value) interface{}

var doc = js.Global().Get("document")

// ---------------------------
// 🔍 Global MutationObserver System
// ---------------------------

type ElementCallback func()

type ObserverManager struct {
	observer    js.Value
	callbacks   map[string]ElementCallback
	isObserving bool
}

var globalObserver *ObserverManager

func init() {
	globalObserver = &ObserverManager{
		callbacks: make(map[string]ElementCallback),
	}
}

// RegisterElement registers an element ID with a callback to be executed when the element is found
func (om *ObserverManager) RegisterElement(id string, callback ElementCallback) {
	// Check if element already exists
	if NodeFromID(id).Truthy() {
		callback()
		return
	}

	om.callbacks[id] = callback

	// Start observing if not already observing
	if !om.isObserving {
		om.startObserving()
	}
}

// UnregisterElement removes an element from tracking
func (om *ObserverManager) UnregisterElement(id string) {
	delete(om.callbacks, id)

	// Stop observing if no more callbacks
	if len(om.callbacks) == 0 && om.isObserving {
		om.stopObserving()
	}
}

// startObserving initializes the MutationObserver
func (om *ObserverManager) startObserving() {
	if om.isObserving {
		return
	}

	observerCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		mutations := args[0]
		mutationsLength := mutations.Get("length").Int()

		for i := 0; i < mutationsLength; i++ {
			mutation := mutations.Index(i)
			if mutation.Get("type").String() == "childList" {
				addedNodes := mutation.Get("addedNodes")
				addedNodesLength := addedNodes.Get("length").Int()

				for j := 0; j < addedNodesLength; j++ {
					node := addedNodes.Index(j)
					om.checkNodeForTargets(node)
				}
			}
		}
		return nil
	})

	om.observer = js.Global().Get("MutationObserver").New(observerCallback)

	config := js.Global().Get("Object").New()
	config.Set("childList", true)
	config.Set("subtree", true)

	om.observer.Call("observe", doc.Get("body"), config)
	om.isObserving = true
}

// stopObserving disconnects the MutationObserver
func (om *ObserverManager) stopObserving() {
	if om.isObserving && om.observer.Truthy() {
		om.observer.Call("disconnect")
		om.isObserving = false
	}
}

// checkNodeForTargets checks if any registered elements are found in the added node
func (om *ObserverManager) checkNodeForTargets(node js.Value) {
	if node.Get("nodeType").Int() != 1 { // Not an element node
		return
	}

	var foundIDs []string

	for id, callback := range om.callbacks {
		if node.Get("id").String() == id {
			foundIDs = append(foundIDs, id)
			callback()
		} else {
			// Check descendants using getElementById instead of querySelector
			// to avoid CSS selector syntax issues with UUIDs starting with digits
			found := doc.Call("getElementById", id)
			if found.Truthy() {
				// Verify that the found element is actually a descendant of the node
				if isDescendantOf(found, node) {
					foundIDs = append(foundIDs, id)
					callback()
				}
			}
		}
	}

	// Remove found elements from callbacks
	for _, id := range foundIDs {
		om.UnregisterElement(id)
	}
}

// Helper function to check if element is a descendant of node
func isDescendantOf(element js.Value, ancestor js.Value) bool {
	current := element.Get("parentNode")
	for current.Truthy() {
		if current.Equal(ancestor) {
			return true
		}
		current = current.Get("parentNode")
	}
	return false
}

// ------------------------------------
// 📦 Reactive Signals (State Handling)
// ------------------------------------

type effect struct {
	fn   func()
	deps map[any]struct{}
}

var currentEffect *effect

type hasWatchers interface {
	removeWatcher(*effect)
}

type Signal[T any] struct {
	value    T
	watchers map[*effect]struct{}
}

func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{
		value:    initial,
		watchers: make(map[*effect]struct{}),
	}
}

func (s *Signal[T]) Get() T {
	if currentEffect != nil {
		s.watchers[currentEffect] = struct{}{}
		currentEffect.deps[s] = struct{}{}
	}
	return s.value
}

func (s *Signal[T]) Set(val T) {
	s.value = val
	for e := range s.watchers {
		go runEffect(e)
	}
}

func (s *Signal[T]) removeWatcher(e *effect) {
	delete(s.watchers, e)
}

func Watch(fn func()) {
	e := &effect{
		fn:   fn,
		deps: make(map[any]struct{}),
	}
	runEffect(e)
}

func runEffect(e *effect) {
	for dep := range e.deps {
		if s, ok := dep.(hasWatchers); ok {
			s.removeWatcher(e)
		}
	}
	e.deps = make(map[any]struct{})
	currentEffect = e
	e.fn()
	currentEffect = nil
}

// ------------------------
// 🖼  Reactive DOM Binding
// ------------------------

func Bind(fn func() Node) Node {
	id := GenID()
	placeholder := Span(Attr("id", id))

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			Watch(func() {
				html := RenderHTML(Div(Attr("id", id), fn()))
				elem := NodeFromID(id)
				if elem.Truthy() {
					elem.Set("outerHTML", html)
				}
			})
		}
	})

	return placeholder
}

func BindText(fn func() string) Node {
	id := GenID()
	span := Span(Attr("id", id), Text(fn()))

	// Register with global observer instead of individual MutationObserver
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Set up reactive Watch effect
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					currentVal := elem.Get("textContent").String()
					newVal := fn()
					if currentVal != newVal {
						elem.Set("textContent", newVal)
					}
				}
			})
		}
	})

	return span
}

// -------------------------
// 🧩 Event Binding Helpers
// -------------------------

func OnClick(f func()) Node {
	id := GenID()
	go func() {
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			elem := NodeFromID(id)
			if elem.Truthy() {
				elem.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					f()
					return nil
				}))
			}
			return nil
		}), 0)
	}()

	return Attr("id", id)
}

// ------------------
// 🧱 DOM Utilities
// ------------------

func GenID() string {
	return fmt.Sprintf("e_%s", uuid.NewString())
}

func Append(html string, Element js.Value) {
	Element.Call("insertAdjacentHTML", "beforeend", html)
}

func NodeFromID(id string) js.Value {
	return doc.Call("getElementById", id)
}

func BodyElement() js.Value {
	return doc.Get("body")
}

// ----------------------
// 🧪 Rendering Utilities
// ----------------------

func RenderHTML(n gomponents.Node) string {
	var b strings.Builder
	err := n.Render(&b)
	if err != nil {
		return "<div>render error</div>"
	}
	return b.String()
}

func Render(n Node) {
	Append(RenderHTML(n), BodyElement())
}

// ------------------
// 🛠 Callback Helper
// ------------------

func Callback(f func()) JsCallback {
	return func(this js.Value, args []js.Value) interface{} {
		f()
		return nil
	}
}

// --------------
// 🧭 App Entrypoint
// --------------

func Run() {
	select {}
}

// ------------------
// 🪵 Debugging
// ------------------

func Log(args ...interface{}) {
	js.Global().Get("console").Call("log", args...)
}

func Logf(format string, args ...interface{}) {
	js.Global().Get("console").Call("log", fmt.Sprintf(format, args...))
}

// ------------------
// Lists (Foreach())
// -------------------

func ForEach[T any](items []T, render func(T) Node) Node {
	var children []Node
	for _, item := range items {
		children = append(children, render(item))
	}
	return Group(children)
}

func ForEachSignal[T any](sig *Signal[[]T], render func(T) Node) Node {
	return Bind(func() Node {
		items := sig.Get()
		var children []Node
		for _, item := range items {
			children = append(children, render(item))
		}
		return Group(children)
	})
}

// -----------
// text inputs
// -----------

func OnInput(handler func(string)) Node {
	id := GenID()
	go func() {
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			elem := NodeFromID(id)
			if elem.Truthy() {
				elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					value := this.Get("value").String()
					handler(value)
					return nil
				}))
			}
			return nil
		}), 0)
	}()
	return Attr("id", id)
}

func BindInput(sig *Signal[string], placeholder string) Node {
	id := GenID()
	isComposing := false

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// attach listeners
			elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				if !isComposing {
					val := this.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
				}
				return nil
			}))

			elem.Call("addEventListener", "compositionstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = true
				return nil
			}))

			elem.Call("addEventListener", "compositionend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = false
				val := this.Get("value").String()
				if val != sig.Get() {
					sig.Set(val)
				}
				return nil
			}))

			elem.Call("addEventListener", "paste", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
					return nil
				}), 0)
				return nil
			}))

			// Now set up the Watch effect after the element is ready
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					signalVal := sig.Get()
					elemVal := elem.Get("value").String()
					if elemVal != signalVal {
						selectionStart := elem.Get("selectionStart")
						selectionEnd := elem.Get("selectionEnd")

						elem.Set("value", signalVal)

						if doc.Get("activeElement").Equal(elem) {
							if selectionStart.Truthy() && selectionEnd.Truthy() {
								start := selectionStart.Int()
								end := selectionEnd.Int()
								maxPos := len(signalVal)

								if start > maxPos {
									start = maxPos
								}
								if end > maxPos {
									end = maxPos
								}

								elem.Call("setSelectionRange", start, end)
							}
						}
					}
				}
			})
		}
	})

	return Input(
		Attr("id", id),
		Type("text"),
		Placeholder(placeholder),
		Value(sig.Get()), // initial value
	)
}

// Enhanced input binding with type support
func BindInputWithType(sig *Signal[string], inputType, placeholder string) Node {
	id := GenID()
	isComposing := false

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// attach listeners
			elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				if !isComposing {
					val := this.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
				}
				return nil
			}))

			elem.Call("addEventListener", "compositionstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = true
				return nil
			}))

			elem.Call("addEventListener", "compositionend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = false
				val := this.Get("value").String()
				if val != sig.Get() {
					sig.Set(val)
				}
				return nil
			}))

			elem.Call("addEventListener", "paste", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
					return nil
				}), 0)
				return nil
			}))

			// Now set up the Watch effect after the element is ready
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					signalVal := sig.Get()
					elemVal := elem.Get("value").String()
					if elemVal != signalVal {
						selectionStart := elem.Get("selectionStart")
						selectionEnd := elem.Get("selectionEnd")

						elem.Set("value", signalVal)

						if doc.Get("activeElement").Equal(elem) {
							if selectionStart.Truthy() && selectionEnd.Truthy() {
								start := selectionStart.Int()
								end := selectionEnd.Int()
								maxPos := len(signalVal)

								if start > maxPos {
									start = maxPos
								}
								if end > maxPos {
									end = maxPos
								}

								elem.Call("setSelectionRange", start, end)
							}
						}
					}
				}
			})
		}
	})

	return Input(
		Attr("id", id),
		Type(inputType),
		Placeholder(placeholder),
		Value(sig.Get()),
	)
}

// Bind input with focus state tracking
func BindInputWithFocus(sig *Signal[string], focusSig *Signal[bool], placeholder string) Node {
	id := GenID()
	isComposing := false

	// Register with global observer instead of polling
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			// Input handling
			elem.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				if !isComposing {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
				}
				return nil
			}))

			// Composition handling
			elem.Call("addEventListener", "compositionstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = true
				return nil
			}))

			elem.Call("addEventListener", "compositionend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				isComposing = false
				val := elem.Get("value").String()
				if val != sig.Get() {
					sig.Set(val)
				}
				return nil
			}))

			// Focus/blur handling
			elem.Call("addEventListener", "focus", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				focusSig.Set(true)
				return nil
			}))

			elem.Call("addEventListener", "blur", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				focusSig.Set(false)
				return nil
			}))

			// Paste handling
			elem.Call("addEventListener", "paste", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := elem.Get("value").String()
					if val != sig.Get() {
						sig.Set(val)
					}
					return nil
				}), 0)
				return nil
			}))

			// Now set up the Watch effect after the element is ready
			Watch(func() {
				elem := NodeFromID(id)
				if elem.Truthy() {
					signalVal := sig.Get()
					elemVal := elem.Get("value").String()
					if elemVal != signalVal {
						selectionStart := elem.Get("selectionStart")
						selectionEnd := elem.Get("selectionEnd")

						elem.Set("value", signalVal)

						if doc.Get("activeElement").Equal(elem) {
							if selectionStart.Truthy() && selectionEnd.Truthy() {
								start := selectionStart.Int()
								end := selectionEnd.Int()
								maxPos := len(signalVal)

								if start > maxPos {
									start = maxPos
								}
								if end > maxPos {
									end = maxPos
								}

								elem.Call("setSelectionRange", start, end)
							}
						}
					}
				}
			})
		}
	})

	return Input(
		Attr("id", id),
		Type("text"),
		Placeholder(placeholder),
		Value(sig.Get()),
	)
}

// -------------------------
// 🧭 Router System
// -------------------------

// RouteParams holds parameters extracted from the URL path
type RouteParams map[string]string

// RouteMatch represents a matched route
type RouteMatch struct {
	Path    string
	Params  RouteParams
	Query   map[string]string
	Handler func(RouteParams) Node
}

// Route defines a single route configuration
type Route struct {
	Path     string
	Handler  func(RouteParams) Node
	Guard    func(RouteParams) bool // Optional route guard
	Children []Route                // For nested routing
}

// Router manages the routing system
type Router struct {
	routes        []Route
	currentPath   *Signal[string]
	currentParams *Signal[RouteParams]
	currentQuery  *Signal[map[string]string]
	basePath      string
	history       js.Value
}

// Global router instance
var globalRouter *Router

// NewRouter creates a new router instance
func NewRouter() *Router {
	r := &Router{
		routes:        make([]Route, 0),
		currentPath:   NewSignal(""),
		currentParams: NewSignal(make(RouteParams)),
		currentQuery:  NewSignal(make(map[string]string)),
		basePath:      "",
		history:       js.Global().Get("window").Get("history"),
	}

	// Set up popstate event listener
	js.Global().Get("window").Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		r.handleURLChange()
		return nil
	}))

	// Initialize with current URL
	r.handleURLChange()

	return r
}

// SetGlobalRouter sets the global router instance
func SetGlobalRouter(r *Router) {
	globalRouter = r
}

// GetCurrentPath returns the current path signal
func (r *Router) GetCurrentPath() *Signal[string] {
	return r.currentPath
}

// GetCurrentParams returns the current params signal
func (r *Router) GetCurrentParams() *Signal[RouteParams] {
	return r.currentParams
}

// GetCurrentQuery returns the current query signal
func (r *Router) GetCurrentQuery() *Signal[map[string]string] {
	return r.currentQuery
}

// AddRoute adds a route to the router
func (r *Router) AddRoute(path string, handler func(RouteParams) Node) {
	r.routes = append(r.routes, Route{
		Path:    path,
		Handler: handler,
	})
}

// AddRouteWithGuard adds a route with a guard function
func (r *Router) AddRouteWithGuard(path string, handler func(RouteParams) Node, guard func(RouteParams) bool) {
	r.routes = append(r.routes, Route{
		Path:    path,
		Handler: handler,
		Guard:   guard,
	})
}

// Navigate programmatically navigates to a path
func (r *Router) Navigate(path string) {
	fullPath := r.basePath + path
	r.history.Call("pushState", nil, "", fullPath)
	r.handleURLChange()
}

// Replace replaces current history entry
func (r *Router) Replace(path string) {
	fullPath := r.basePath + path
	r.history.Call("replaceState", nil, "", fullPath)
	r.handleURLChange()
}

// handleURLChange processes URL changes
func (r *Router) handleURLChange() {
	location := js.Global().Get("window").Get("location")
	pathname := location.Get("pathname").String()
	search := location.Get("search").String()

	// Remove base path if present
	if r.basePath != "" && strings.HasPrefix(pathname, r.basePath) {
		pathname = strings.TrimPrefix(pathname, r.basePath)
	}

	// Parse query parameters
	query := parseQueryString(search)

	// Find matching route
	match := r.matchRoute(pathname)

	// Update signals
	r.currentPath.Set(pathname)
	r.currentQuery.Set(query)

	if match != nil {
		r.currentParams.Set(match.Params)
	} else {
		r.currentParams.Set(make(RouteParams))
	}
}

// matchRoute finds a matching route for the given path
func (r *Router) matchRoute(path string) *RouteMatch {
	for _, route := range r.routes {
		if params := r.matchPath(route.Path, path); params != nil {
			// Check guard if present
			if route.Guard != nil && !route.Guard(*params) {
				continue
			}

			return &RouteMatch{
				Path:    path,
				Params:  *params,
				Handler: route.Handler,
			}
		}
	}
	return nil
}

// matchPath checks if a path matches a route pattern and extracts parameters
func (r *Router) matchPath(pattern, path string) *RouteParams {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle root path
	if pattern == "/" && path == "/" {
		params := make(RouteParams)
		return &params
	}

	if len(patternParts) != len(pathParts) {
		return nil
	}

	params := make(RouteParams)

	for i, patternPart := range patternParts {
		if patternPart == "" && pathParts[i] == "" {
			continue // Both empty, skip
		}

		if strings.HasPrefix(patternPart, ":") {
			// Parameter
			paramName := strings.TrimPrefix(patternPart, ":")
			params[paramName] = pathParts[i]
		} else if patternPart != pathParts[i] {
			// Literal match failed
			return nil
		}
	}

	return &params
}

// parseQueryString parses URL query string into a map
func parseQueryString(query string) map[string]string {
	params := make(map[string]string)
	if query == "" {
		return params
	}

	// Remove leading ?
	query = strings.TrimPrefix(query, "?")

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}

		params[key] = value
	}

	return params
}

// RouterOutlet renders the current route
func RouterOutlet() Node {
	if globalRouter == nil {
		return Div(Text("Router not initialized"))
	}

	return Bind(func() Node {
		path := globalRouter.currentPath.Get()
		params := globalRouter.currentParams.Get()

		match := globalRouter.matchRoute(path)
		if match != nil {
			return match.Handler(params)
		}

		// Default 404 page
		return Div(
			H1(Text("404 - Page Not Found")),
			P(Text("The page you're looking for doesn't exist.")),
			A(Href("/"), Text("Go Home")),
		)
	})
}

// RouterLink creates a navigation link that uses the router
func RouterLink(href string, children ...Node) Node {
	id := GenID()

	// Register click handler with global observer
	globalObserver.RegisterElement(id, func() {
		elem := NodeFromID(id)
		if elem.Truthy() {
			elem.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				event := args[0]
				// Prevent default browser navigation
				event.Call("preventDefault")

				// Use router for client-side navigation
				if globalRouter != nil {
					globalRouter.Navigate(href)
				}
				return nil
			}))
		}
	})

	return A(
		Attr("id", id),
		Href(href), // Keep href for accessibility/SEO
		Style("text-decoration: none; color: inherit; cursor: pointer;"),
		Group(children),
	)
}

// UseParams returns the current route parameters
func UseParams() *Signal[RouteParams] {
	if globalRouter == nil {
		return NewSignal(make(RouteParams))
	}
	return globalRouter.currentParams
}

// UseQuery returns the current query parameters
func UseQuery() *Signal[map[string]string] {
	if globalRouter == nil {
		return NewSignal(make(map[string]string))
	}
	return globalRouter.currentQuery
}

// UseLocation returns the current path
func UseLocation() *Signal[string] {
	if globalRouter == nil {
		return NewSignal("")
	}
	return globalRouter.currentPath
}
