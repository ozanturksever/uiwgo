package router

import (
	"regexp"
	"strings"
)

// MatcherFunc is a pre-compiled function that determines if a given path
// matches the route's pattern and extracts any dynamic parameters.
type MatcherFunc func(path string) (isMatch bool, params map[string]string)

// RouteDefinition encapsulates all information about a single route.
type RouteDefinition struct {
	Path         string
	Component    func(props ...any) interface{} // Will be more specific with gomponents.Node later
	Children     []*RouteDefinition
	MatchFilters map[string]any // Parameter validation filters (regex or function)

	// Internal pre-compiled matcher for performance.
	matcher MatcherFunc
}

// validateParams checks if captured parameters match their respective filters.
func validateParams(params map[string]string, filters map[string]any) bool {
	for paramName, paramValue := range params {
		if filter, exists := filters[paramName]; exists {
			switch f := filter.(type) {
			case string:
				// Regex filter
				matched, err := regexp.MatchString(f, paramValue)
				if err != nil || !matched {
					return false
				}
			case func(string) bool:
				// Function filter
				if !f(paramValue) {
					return false
				}
			default:
				// Unsupported filter type
				return false
			}
		}
	}
	return true
}

// compileMatcher compiles the route's path pattern into a MatcherFunc.
func compileMatcher(r *RouteDefinition) MatcherFunc {
	// Split the route pattern into segments
	patternSegments := strings.Split(r.Path, "/")

	return func(inputPath string) (bool, map[string]string) {
		// Split the input path into segments
		inputSegments := strings.Split(inputPath, "/")

		params := make(map[string]string)

		// Track pattern and input indices
		patternIndex := 0
		inputIndex := 0

		// Iterate through pattern segments
		for patternIndex < len(patternSegments) && inputIndex < len(inputSegments) {
			patternSegment := patternSegments[patternIndex]
			inputSegment := inputSegments[inputIndex]

			// Skip empty segments (like leading/trailing slashes)
			if patternSegment == "" && inputSegment == "" {
				patternIndex++
				inputIndex++
				continue
			}

			// Handle wildcard segments (starting with *)
			if strings.HasPrefix(patternSegment, "*") {
				// Extract parameter name (remove the asterisk)
				paramName := patternSegment[1:]

				// Wildcard must be the last segment in the pattern
				if patternIndex != len(patternSegments)-1 {
					return false, nil
				}

				// Capture all remaining input segments
				remainingSegments := inputSegments[inputIndex:]
				params[paramName] = strings.Join(remainingSegments, "/")
				// Apply filters after capturing all parameters
				if !validateParams(params, r.MatchFilters) {
					return false, nil
				}
				return true, params
			}

			// Handle optional dynamic segments (ending with ?)
			if strings.HasSuffix(patternSegment, "?") {
				// Extract parameter name (remove the colon and question mark)
				paramName := patternSegment[1 : len(patternSegment)-1]
				// Capture the value from the input segment, even if empty
				params[paramName] = inputSegment
				patternIndex++
				inputIndex++
				continue
			}

			// Handle dynamic segments (starting with :)
			if strings.HasPrefix(patternSegment, ":") {
				// Extract parameter name (remove the colon)
				paramName := patternSegment[1:]
				// Capture the value from the input segment
				params[paramName] = inputSegment
				patternIndex++
				inputIndex++
				continue
			}

			// Handle static segments - exact match required
			if patternSegment != inputSegment {
				return false, nil
			}

			patternIndex++
			inputIndex++
		}

		// Check if we've processed all pattern segments and input segments
		if patternIndex != len(patternSegments) || inputIndex != len(inputSegments) {
			return false, nil
		}

		// Apply filters after capturing all parameters
		if !validateParams(params, r.MatchFilters) {
			return false, nil
		}

		return true, params
	}
}

// NewRouteDefinition creates a new RouteDefinition with a compiled matcher.
func NewRouteDefinition(path string, component func(...any) interface{}) *RouteDefinition {
	rd := &RouteDefinition{
		Path:         path,
		Component:    component,
		Children:     make([]*RouteDefinition, 0),
		MatchFilters: make(map[string]any), // Initialize empty map for filters
	}
	rd.matcher = compileMatcher(rd)
	return rd
}
