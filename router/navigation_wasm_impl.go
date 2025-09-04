//go:build js && wasm

package router

import (
	"syscall/js"
	"strings"

	"github.com/ozanturksever/uiwgo/logutil"
	dom "honnef.co/go/js/dom/v2"
)

// navigateWASMImpl handles navigation in WASM builds with proper history API integration.
func (r *Router) navigateWASMImpl(path string, options NavigateOptions) {
	window := dom.GetWindow()
	if window == nil {
		// Fallback for test environments where DOM is not available
		newLocation := Location{
			Pathname: path,
			Search:   "",
			Hash:     "",
			State:    options.State,
		}
		r.locationState.Set(newLocation)
		return
	}

	history := window.History()
	if history == nil {
		// Fallback for test environments where History API is not available
		newLocation := Location{
			Pathname: path,
			Search:   "",
			Hash:     "",
			State:    options.State,
		}
		r.locationState.Set(newLocation)
		return
	}

	// Create a new Location with the given path
	newLocation := Location{
		Pathname: path,
		Search:   "", // TODO: Parse search from path if needed
		Hash:     "", // TODO: Parse hash from path if needed
		State:    options.State,
	}

	// Push the new state to browser history
	var stateValue js.Value
	if options.State != nil {
		// Convert Go state to JS value
		stateValue = js.ValueOf(options.State)
	} else {
		stateValue = js.Null()
	}

	// Use pushState to add the new location to browser history
	history.PushState(stateValue, "", path)

	// Update the router's location state (this will trigger rendering)
	r.locationState.Set(newLocation)
}

// setupNavigationWASM sets up WASM-specific navigation functionality.
func setupNavigationWASM(router *Router) {
	// Add click event delegation for navigation links
	setupLinkClickHandling(router)
}

// setupLinkClickHandling sets up event delegation for navigation links.
func setupLinkClickHandling(router *Router) {
	document := dom.GetWindow().Document()

	// Add click event listener to document for event delegation
	document.AddEventListener("click", true, func(event dom.Event) {
		// Find closest ancestor <a>
		target := event.Target()
		if target == nil {
			return
		}

		// Only intercept unmodified left-clicks
		if me, ok := event.(dom.MouseEvent); ok {
			if me.Button() != 0 || me.AltKey() || me.CtrlKey() || me.MetaKey() || me.ShiftKey() {
				return
			}
		}

		// Ascend DOM to find an anchor element
		var curr dom.Node
		if tn, ok := target.(dom.Node); ok {
			curr = tn
		} else {
			return
		}

		var anchor dom.Element
		for curr != nil {
			if el, ok := curr.(dom.Element); ok {
				if el.TagName() == "A" {
					anchor = el
					break
				}
			}
			curr = curr.ParentNode()
		}

		if anchor == nil {
			return
		}

		// Respect explicit opts that should not be intercepted
		targetAttr := strings.ToLower(anchor.GetAttribute("target"))
		if targetAttr == "_blank" || anchor.HasAttribute("download") {
			return
		}

		// Get href and decide if we should intercept
		href := anchor.GetAttribute("href")
		if href == "" {
			return
		}
		if strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") || strings.HasPrefix(href, "javascript:") {
			return
		}

		// If the anchor opts into router explicitly, always intercept
		explicitRouter := anchor.HasAttribute("data-router-link") || anchor.GetAttribute("data-router-link") == "true"

		shouldIntercept := explicitRouter
		if !shouldIntercept {
			// Internal navigation heuristics similar to SolidStart
			if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "./") || strings.HasPrefix(href, "../") {
				shouldIntercept = true
			} else if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
				// Check same-origin via URL API
				base := dom.GetWindow().Location().Href()
				u := js.Global().Get("URL").New(href, base)
				if u.Truthy() {
					currOrigin := js.Global().Get("location").Get("origin").String()
					destOrigin := u.Get("origin").String()
					if destOrigin == currOrigin {
						shouldIntercept = true
					}
				}
			} else {
				// Relative without scheme, treat as internal
				shouldIntercept = true
			}
		}

		if !shouldIntercept {
			return
		}

		// Prevent default browser navigation and route internally
		event.PreventDefault()
		event.StopPropagation()

		// Normalize path to include search/hash if provided
		path := href
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			base := dom.GetWindow().Location().Href()
			u := js.Global().Get("URL").New(href, base)
			if u.Truthy() {
				path = u.Get("pathname").String() + u.Get("search").String() + u.Get("hash").String()
			}
		}

		logutil.Logf("Intercepted navigation to %s", path)
		router.Navigate(path, NavigateOptions{})
	})
}