package router

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRouteMatcher_StaticPathMatches(t *testing.T) {
	route := NewRouteDefinition("/about", nil)
	
	isMatch, params := route.matcher("/about")
	assert.True(t, isMatch, "Static path should match exactly")
	assert.Empty(t, params, "No params expected for static match")
}

func TestRouteMatcher_DynamicSegmentCapturesParam(t *testing.T) {
	route := NewRouteDefinition("/users/:id", nil)

	// Valid match with parameter
	isMatch, params := route.matcher("/users/123")
	assert.True(t, isMatch, "Should match dynamic segment")
	assert.Equal(t, "123", params["id"], "Should capture :id parameter")

	// Too few segments
	isMatch, params = route.matcher("/users")
	assert.False(t, isMatch, "Should not match with missing segment")

	// Too many segments
	isMatch, params = route.matcher("/users/123/posts")
	assert.False(t, isMatch, "Should not match with extra segments")
}

func TestRouteMatcher_OptionalSegmentMatches(t *testing.T) {
	route := NewRouteDefinition("/archive/:year?/:month?", nil)

	tests := []struct {
		path     string
		matches  bool
		params   map[string]string
	}{
		{"/archive/2023", true, map[string]string{"year": "2023"}},
		{"/archive/2023/08", true, map[string]string{"year": "2023", "month": "08"}},
		{"/archive", true, map[string]string{}},
		{"/archive/2023/08/15", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			isMatch, params := route.matcher(tt.path)
			assert.Equal(t, tt.matches, isMatch, "Path: %s", tt.path)
			if tt.matches {
				assert.Equal(t, tt.params, params, "Params mismatch for path: %s", tt.path)
			}
		})
	}
}
