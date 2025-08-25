//go:build !js && !wasm

package router

// A creates a declarative navigation link component.
// It returns a simple struct with Href and OnClick fields for testing purposes.
// In a real implementation with gomponents, this would return a proper component.
func A(href string, children ...any) any {
	return struct {
		Href    string
		OnClick func()
	}{
		Href: href,
		OnClick: func() {
			// Use the current router to navigate to the href
			if currentRouter != nil {
				currentRouter.navigate(href, NavigateOptions{})
			}
		},
	}
}
