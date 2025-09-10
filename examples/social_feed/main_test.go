//go:build !js && !wasm

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestSocialFeed(t *testing.T) {
	server := testhelpers.NewViteServer("social_feed", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to social feed: %v", err)
	}

	// Test that the feed loads correctly
	var title string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text("h1", &title, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get title: %v", err)
	}

	if title != "Social Feed" {
		t.Errorf("Expected title 'Social Feed', got '%s'", title)
	}

	// Test that post composer is visible
	var composerPlaceholder string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.AttributeValue("textarea", "placeholder", &composerPlaceholder, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get composer placeholder: %v", err)
	}

	if composerPlaceholder != "What's on your mind?" {
		t.Errorf("Expected composer placeholder 'What's on your mind?', got '%s'", composerPlaceholder)
	}
}

func TestSocialFeedPostCreation(t *testing.T) {
	server := testhelpers.NewViteServer("social_feed", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to social feed: %v", err)
	}

	// Create a new post
	postContent := "This is a test post from the browser test!"
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.SendKeys("textarea", postContent, chromedp.ByQuery),
		chromedp.Click("button[type='submit']", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Wait for post to appear
	time.Sleep(1 * time.Second)

	// Verify post content appears in feed
	var postText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".post-content", &postText, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get post text: %v", err)
	}

	if !strings.Contains(postText, postContent) {
		t.Errorf("Expected post to contain '%s', got '%s'", postContent, postText)
	}
}

func TestSocialFeedFiltering(t *testing.T) {
	server := testhelpers.NewViteServer("social_feed", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to social feed: %v", err)
	}

	// Test filter tabs are visible
	var filterTabText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".filter-tab", &filterTabText, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get filter tab text: %v", err)
	}

	if filterTabText != "All" {
		t.Errorf("Expected first filter tab to be 'All', got '%s'", filterTabText)
	}

	// Click on a filter tab
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click("//button[contains(text(), 'Image')]", chromedp.BySearch),
	)
	if err != nil {
		t.Fatalf("Failed to click Image filter: %v", err)
	}

	// Wait for filter to apply
	time.Sleep(500 * time.Millisecond)

	// Verify filter is active
	var activeFilter string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".filter-tab.active", &activeFilter, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get active filter: %v", err)
	}

	if activeFilter != "Image" {
		t.Errorf("Expected active filter to be 'Image', got '%s'", activeFilter)
	}
}

func TestSocialFeedPostInteraction(t *testing.T) {
	server := testhelpers.NewViteServer("social_feed", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to social feed: %v", err)
	}

	// Test like button interaction
	var likeCountBefore string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".like-button", &likeCountBefore, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get initial like count: %v", err)
	}

	// Click like button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(".like-button", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to click like button: %v", err)
	}

	// Wait for like to register
	time.Sleep(500 * time.Millisecond)

	// Verify like count changed
	var likeCountAfter string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".like-button", &likeCountAfter, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get updated like count: %v", err)
	}

	if likeCountBefore == likeCountAfter {
		t.Error("Expected like count to change after clicking like button")
	}

	// Test comment button interaction
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(".comment-button", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to click comment button: %v", err)
	}

	// Wait for comments to show
	time.Sleep(500 * time.Millisecond)

	// Verify comments section is visible
	var commentsVisible bool
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.WaitVisible(".comments-section", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Comments section should be visible after clicking comment button: %v", err)
	}

	_ = commentsVisible // Avoid unused variable error
}

func TestSocialFeedInfiniteScroll(t *testing.T) {
	server := testhelpers.NewViteServer("social_feed", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	err := chromedp.Run(chromedpCtx.Ctx,
		testhelpers.Actions.NavigateAndWaitForLoad(server.URL(), "body"),
		// Wait for WASM to initialize
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to social feed: %v", err)
	}

	// Check if load more button is present
	var loadMoreText string
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Text(".load-more", &loadMoreText, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to find load more button: %v", err)
	}

	if loadMoreText != "Load More" {
		t.Errorf("Expected load more button text 'Load More', got '%s'", loadMoreText)
	}

	// Click load more button
	err = chromedp.Run(chromedpCtx.Ctx,
		chromedp.Click(".load-more", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to click load more button: %v", err)
	}

	// Wait for more posts to load
	time.Sleep(1 * time.Second)

	// Verify more posts are loaded (this would need more specific selectors in a real implementation)
	t.Log("Infinite scroll test completed - more posts should be loaded")
}
