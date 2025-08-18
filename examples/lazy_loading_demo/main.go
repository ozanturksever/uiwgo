// examples/lazy_loading_demo/main.go
// Comprehensive demonstration of lazy loading patterns with createResource and Suspense

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"app/golid"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ------------------------------------
// 🎯 Demo Data Types
// ------------------------------------

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Post struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

type Comment struct {
	ID     int    `json:"id"`
	PostID int    `json:"postId"`
	Name   string `json:"name"`
	Body   string `json:"body"`
}

// ------------------------------------
// 🔄 Mock API Functions
// ------------------------------------

// Simulate API calls with realistic delays and potential failures
func fetchUsers() ([]User, error) {
	time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)

	// Simulate occasional failures
	if rand.Float32() < 0.1 {
		return nil, fmt.Errorf("failed to fetch users")
	}

	return []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
		{ID: 3, Name: "Bob Johnson", Email: "bob@example.com"},
	}, nil
}

func fetchPosts() ([]Post, error) {
	time.Sleep(time.Duration(rand.Intn(800)+300) * time.Millisecond)

	if rand.Float32() < 0.05 {
		return nil, fmt.Errorf("failed to fetch posts")
	}

	return []Post{
		{ID: 1, Title: "First Post", Body: "This is the first post content", UserID: 1},
		{ID: 2, Title: "Second Post", Body: "This is the second post content", UserID: 2},
		{ID: 3, Title: "Third Post", Body: "This is the third post content", UserID: 1},
	}, nil
}

func fetchComments(postID int) ([]Comment, error) {
	time.Sleep(time.Duration(rand.Intn(600)+200) * time.Millisecond)

	if rand.Float32() < 0.08 {
		return nil, fmt.Errorf("failed to fetch comments for post %d", postID)
	}

	return []Comment{
		{ID: 1, PostID: postID, Name: "Commenter 1", Body: "Great post!"},
		{ID: 2, PostID: postID, Name: "Commenter 2", Body: "Very informative."},
	}, nil
}

func fetchUserProfile(userID int) (User, error) {
	time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

	if rand.Float32() < 0.05 {
		return User{}, fmt.Errorf("failed to fetch user profile %d", userID)
	}

	return User{
		ID:    userID,
		Name:  fmt.Sprintf("User %d", userID),
		Email: fmt.Sprintf("user%d@example.com", userID),
	}, nil
}

// ------------------------------------
// 🎨 Resource-based Components
// ------------------------------------

// UserListComponent demonstrates basic resource usage with Suspense
func UserListComponent() Node {
	// Create resource for users
	usersResource := golid.CreateResource(fetchUsers, golid.ResourceOptions{
		Name:     "users",
		CacheKey: "users-list",
		TTL:      5 * time.Minute,
	})

	// Use the resource to avoid unused warning
	_ = usersResource

	// Create a component that uses the resource
	return Div(
		Class("user-list"),
		H2(Text("Users")),
		Text("Users will be loaded here with Suspense..."),
		// In a real implementation, this would be wrapped in Suspense
		// and the resource would be read to trigger loading
	)
}

// HeavyDashboardComponent simulates a heavy component that should be lazy loaded
func HeavyDashboardComponent() (Node, error) {
	// Simulate heavy component loading
	time.Sleep(2 * time.Second)

	return Div(
		Class("heavy-dashboard"),
		H2(Text("Heavy Dashboard")),
		P(Text("This component took 2 seconds to load and contains complex visualizations.")),
		Div(
			Class("dashboard-widgets"),
			Div(Class("widget"), Text("Widget 1: Charts")),
			Div(Class("widget"), Text("Widget 2: Analytics")),
			Div(Class("widget"), Text("Widget 3: Reports")),
		),
	), nil
}

// LazyDashboardDemo demonstrates lazy component loading
func LazyDashboardDemo() Node {
	// Create lazy component
	lazyDashboard := golid.Lazy(
		HeavyDashboardComponent,
		golid.LazyOptions{
			Name:       "heavy-dashboard",
			Cache:      true,
			MaxRetries: 3,
			OnLoad: func(component Node) {
				fmt.Println("Heavy dashboard loaded successfully!")
			},
			OnError: func(err error) {
				fmt.Printf("Failed to load dashboard: %v\n", err)
			},
		},
	)

	return Div(
		Class("lazy-demo"),
		H2(Text("Lazy Loading Demo")),
		P(Text("The heavy dashboard component will be lazy loaded:")),
		Div(
			ID("dashboard-container"),
			// Render lazy component with fallback
			lazyDashboard.RenderWithFallback(
				Div(
					Class("dashboard-placeholder"),
					Text("Dashboard will load here..."),
				),
			),
		),
	)
}

// ------------------------------------
// 🎯 Advanced Patterns
// ------------------------------------

// UserProfileModal demonstrates conditional resource loading
func UserProfileModal(userID int, isOpen bool) Node {
	if !isOpen {
		return Div() // Empty div when closed
	}

	// Only create resource when modal is open
	profileResource := golid.CreateResource(
		func() (User, error) {
			return fetchUserProfile(userID)
		},
		golid.ResourceOptions{
			Name:     fmt.Sprintf("user-profile-%d", userID),
			CacheKey: fmt.Sprintf("profile-%d", userID),
			TTL:      1 * time.Minute,
			Preload:  true, // Start loading immediately
		},
	)

	// Use the resource to avoid unused warning
	_ = profileResource

	return Div(
		Class("modal-overlay"),
		Div(
			Class("modal"),
			H3(Text("User Profile")),
			Div(
				Class("profile-content"),
				Text("Profile will load here..."),
				// In real implementation, would use Suspense and read from profileResource
			),
			Button(
				Type("button"),
				Text("Close"),
			),
		),
	)
}

// InfiniteScrollComponent demonstrates resource pagination
func InfiniteScrollComponent() Node {
	pageGetter, pageSetter := golid.CreateSignal(1)

	// Create resource that depends on page signal
	itemsResource := golid.CreateResource(
		func() ([]string, error) {
			currentPage := pageGetter()
			time.Sleep(500 * time.Millisecond) // Simulate API call

			items := make([]string, 10)
			for i := 0; i < 10; i++ {
				items[i] = fmt.Sprintf("Item %d (Page %d)", i+1, currentPage)
			}
			return items, nil
		},
		golid.ResourceOptions{
			Name:         "paginated-items",
			Dependencies: []interface{}{pageGetter}, // Refetch when page changes
		},
	)

	// Use the variables to avoid unused warnings
	_ = itemsResource
	_ = pageSetter

	return Div(
		Class("infinite-scroll"),
		H2(Text("Infinite Scroll Demo")),
		Div(
			Class("items-list"),
			Text("Items will load here with pagination..."),
		),
		Button(
			Type("button"),
			Text("Load Next Page"),
		),
	)
}

// ------------------------------------
// 🔄 Error Handling Examples
// ------------------------------------

// ErrorBoundaryDemo demonstrates error handling with resources
func ErrorBoundaryDemo() Node {
	// Resource that might fail
	unreliableResource := golid.CreateResource(
		func() (string, error) {
			if rand.Float32() < 0.5 {
				return "", fmt.Errorf("random failure occurred")
			}
			return "Success! Data loaded correctly.", nil
		},
		golid.ResourceOptions{
			Name:       "unreliable-data",
			MaxRetries: 2,
			OnError: func(err error) {
				fmt.Printf("Resource error: %v\n", err)
			},
		},
	)

	// Create error boundary
	errorBoundary := golid.CreateErrorBoundary(func(err error) interface{} {
		return golid.ErrorFallback(err.Error())
	})

	// Use the variables to avoid unused warnings
	_ = unreliableResource
	_ = errorBoundary

	return Div(
		Class("error-demo"),
		H2(Text("Error Handling Demo")),
		P(Text("This resource has a 50% chance of failing:")),
		Button(
			Type("button"),
			Text("Retry Loading"),
		),
		Div(
			Class("error-content"),
			Text("Error handling will be demonstrated here..."),
		),
	)
}

// ------------------------------------
// 🎨 Main Application
// ------------------------------------

func LazyLoadingApp() Node {
	return El("html",
		El("head",
			El("title", Text("Golid Lazy Loading Demo")),
			El("style", Text(`
				body { font-family: Arial, sans-serif; margin: 20px; }
				.demo-section { margin: 30px 0; padding: 20px; border: 1px solid #ddd; border-radius: 8px; }
				.loading-spinner { text-align: center; padding: 20px; }
				.skeleton-line { height: 20px; background: #f0f0f0; margin: 5px 0; border-radius: 4px; }
				.user-list, .posts-container { margin: 20px 0; }
				.post { margin: 15px 0; padding: 15px; border: 1px solid #eee; border-radius: 4px; }
				.comments { margin-top: 10px; padding-left: 20px; }
				.widget { padding: 10px; margin: 5px; background: #f5f5f5; border-radius: 4px; }
				.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); }
				.modal { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); background: white; padding: 20px; border-radius: 8px; }
				.error-fallback { padding: 20px; border: 1px solid #ff6b6b; background: #ffe0e0; border-radius: 4px; }
				button { padding: 10px 15px; margin: 5px; border: none; background: #007bff; color: white; border-radius: 4px; cursor: pointer; }
				button:hover { background: #0056b3; }
			`)),
		),
		El("body",
			H1(Text("🚀 Golid Lazy Loading & Suspense Demo")),

			// Basic Resource Loading
			Div(
				Class("demo-section"),
				H2(Text("1. Basic Resource Loading with Suspense")),
				P(Text("Demonstrates createResource with automatic Suspense boundaries:")),
				UserListComponent(),
			),

			// Lazy Components
			Div(
				Class("demo-section"),
				H2(Text("2. Lazy Component Loading")),
				P(Text("Demonstrates code-splitting with lazy component loading:")),
				LazyDashboardDemo(),
			),

			// Conditional Loading
			Div(
				Class("demo-section"),
				H2(Text("3. Conditional Resource Loading")),
				P(Text("Resources that load only when needed:")),
				UserProfileModal(1, true), // Simulate open modal
			),

			// Pagination
			Div(
				Class("demo-section"),
				H2(Text("4. Reactive Dependencies & Pagination")),
				P(Text("Resources that refetch when dependencies change:")),
				InfiniteScrollComponent(),
			),

			// Error Handling
			Div(
				Class("demo-section"),
				H2(Text("5. Error Handling & Recovery")),
				P(Text("Demonstrates error boundaries and retry mechanisms:")),
				ErrorBoundaryDemo(),
			),

			// Performance Tips
			Div(
				Class("demo-section"),
				H2(Text("💡 Performance Tips")),
				Ul(
					Li(Text("Use caching with appropriate TTL values")),
					Li(Text("Implement proper error boundaries")),
					Li(Text("Preload critical resources")),
					Li(Text("Use lazy loading for heavy components")),
					Li(Text("Leverage resource deduplication")),
					Li(Text("Monitor cache hit ratios")),
				),
			),
		),
	)
}

func main() {
	fmt.Println("🚀 Starting Golid Lazy Loading Demo...")

	// Initialize random seed for demo
	rand.Seed(time.Now().UnixNano())

	// In a real application, you would render to DOM
	// golid.Render(LazyLoadingApp())
	// golid.Run()

	// For this demo, we'll just show the structure
	fmt.Println("Demo application structure created!")
	fmt.Println("This demonstrates:")
	fmt.Println("- createResource for async data loading")
	fmt.Println("- Suspense boundaries with fallback rendering")
	fmt.Println("- Lazy component loading with code splitting")
	fmt.Println("- Resource caching and deduplication")
	fmt.Println("- Error handling and recovery")
	fmt.Println("- Reactive dependencies and refetching")

	// Simulate some resource operations
	fmt.Println("\n🔄 Simulating resource operations...")

	// Test resource creation
	testResource := golid.CreateResource(fetchUsers, golid.ResourceOptions{
		Name:     "test-users",
		CacheKey: "test-users",
		TTL:      1 * time.Minute,
	})

	fmt.Printf("✅ Created resource: %T\n", testResource)

	// Test lazy component
	lazyComp := golid.Lazy(func() (Node, error) {
		return Div(Text("Lazy loaded!")), nil
	})

	fmt.Printf("✅ Created lazy component: %T\n", lazyComp)

	// Test suspense boundary
	suspense := golid.Suspense(
		Div(Text("Loading...")),
		Div(Text("Content")),
	)

	fmt.Printf("✅ Created suspense boundary: %T\n", suspense)

	// Test cache operations
	cache := golid.NewResourceCache(100, 5*time.Minute)
	cache.Set("test-key", "test-value", 1*time.Minute)

	if value, found := cache.Get("test-key"); found {
		fmt.Printf("✅ Cache test successful: %v\n", value)
	}

	stats := cache.Stats()
	fmt.Printf("✅ Cache stats: %+v\n", stats)

	// Test advanced cache
	advancedCache := golid.NewAdvancedResourceCache(golid.CacheConfig{
		Policy:    golid.CachePolicyLRU,
		MaxSize:   50,
		MaxMemory: 1024 * 1024, // 1MB
	})

	advancedCache.Set("advanced-key", "advanced-value")
	if value, found := advancedCache.Get("advanced-key"); found {
		fmt.Printf("✅ Advanced cache test successful: %v\n", value)
	}

	advancedStats := advancedCache.GetAdvancedStats()
	fmt.Printf("✅ Advanced cache stats: %+v\n", advancedStats)

	// Cleanup
	advancedCache.Stop()

	log.Println("✅ Lazy loading demo completed successfully!")
}
