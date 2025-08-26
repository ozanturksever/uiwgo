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
// For nested routes, it allows partial matches when the route has children.
func compileMatcher(r *RouteDefinition) MatcherFunc {
	// Split the route pattern into segments, filtering out empty ones
	// Parse pattern segments with optional markers
patternSegments := make([]string, 0)
optionalSegments := make(map[int]bool)

for i, s := range filterEmptySegments(strings.Split(r.Path, "/")) {
	if strings.HasSuffix(s, "?") {
		patternSegments = append(patternSegments, strings.TrimSuffix(s, "?"))
		optionalSegments[i] = true
	} else {
		patternSegments = append(patternSegments, s)
	}
}

	return func(inputPath string) (bool, map[string]string) {
		// Split the input path into segments
		// For optional parameters, we need to preserve empty segments
		inputSegments := strings.Split(inputPath, "/")
		if len(inputSegments) > 0 && inputSegments[0] == "" {
			inputSegments = inputSegments[1:] // Remove leading empty segment from root slash
		}

		params := make(map[string]string)
		
		// Handle root path case
		if len(patternSegments) == 0 {
			// Pattern is root path, check if input is also root-like
			if len(inputSegments) == 0 || (len(inputSegments) == 1 && inputSegments[0] == "") {
				return true, params
			}
			return false, nil
		}

		// Track pattern and input indices
		patternIndex := 0
		inputIndex := 0

		// Iterate through pattern segments
		for patternIndex < len(patternSegments) {
			patternSegment := patternSegments[patternIndex]

			if inputIndex >= len(inputSegments) {
				// Check if current pattern segment is optional
				if optionalSegments[patternIndex] {
					patternIndex++
					continue
				}
				return false, nil
			}

			inputSegment := inputSegments[inputIndex]

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
				
				// Check if the input segment is empty (optional parameter absent)
				if inputSegment == "" {
					// Optional parameter is absent, set empty value
					params[paramName] = ""
					patternIndex++
					inputIndex++
					continue
				}
				
				// Capture the value from the input segment
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
				// Immediately validate against filter if present
				if filter, exists := r.MatchFilters[paramName]; exists {
					switch f := filter.(type) {
					case string:
						if matched, _ := regexp.MatchString(f, inputSegment); !matched {
							return false, nil
						}
					case func(string) bool:
						if !f(inputSegment) {
							return false, nil
						}
					default:
						return false, nil
					}
				}
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

		// Check if we've processed all pattern segments
		if patternIndex != len(patternSegments) {
			// Not all pattern segments were matched
			return false, nil
		}

		// For routes with children, allow partial matches (more input segments remaining)
		// For routes without children, require exact match (all input segments consumed)
		if len(r.Children) == 0 && inputIndex != len(inputSegments) {
			// Route has no children but there are remaining input segments
			return false, nil
		}

		// Apply filters after capturing all parameters
		if !validateParams(params, r.MatchFilters) {
			return false, nil
		}

		return true, params
	}
}

// filterEmptySegments removes empty strings from a slice of segments
func filterEmptySegments(segments []string) []string {
	filtered := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment != "" {
			filtered = append(filtered, segment)
		}
	}
	return filtered
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
