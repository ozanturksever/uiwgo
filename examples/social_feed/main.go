//go:build js && wasm

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	"github.com/ozanturksever/uiwgo/wasm"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type PostType string

const (
	PostTypeText  PostType = "text"
	PostTypeImage PostType = "image"
	PostTypeVideo PostType = "video"
	PostTypeLink  PostType = "link"
)

type Post struct {
	ID        string
	Author    User
	Content   string
	Type      PostType
	MediaURL  string
	Timestamp time.Time
	Likes     int
	Comments  []Comment
	Shares    int
	IsLiked   bool
	IsShared  bool
}

type Comment struct {
	ID        string
	Author    User
	Content   string
	Timestamp time.Time
	Likes     int
	IsLiked   bool
}

type User struct {
	ID        string
	Username  string
	Name      string
	AvatarURL string
	Verified  bool
}

type SocialFeed struct {
	posts          reactivity.Signal[[]Post]
	loading        reactivity.Signal[bool]
	hasMore        reactivity.Signal[bool]
	selectedFilter reactivity.Signal[PostType]
	showComments   reactivity.Signal[map[string]bool]
	newPostContent reactivity.Signal[string]
	currentUser    reactivity.Signal[User]
}

func NewSocialFeed() *SocialFeed {
	return &SocialFeed{
		posts:          reactivity.CreateSignal([]Post{}),
		loading:        reactivity.CreateSignal(false),
		hasMore:        reactivity.CreateSignal(true),
		selectedFilter: reactivity.CreateSignal(PostType("")),
		showComments:   reactivity.CreateSignal(make(map[string]bool)),
		newPostContent: reactivity.CreateSignal(""),
		currentUser:    reactivity.CreateSignal(User{}),
	}
}

func (sf *SocialFeed) loadSampleData() {
	// Sample users
	users := []User{
		{ID: "1", Username: "johndoe", Name: "John Doe", AvatarURL: "https://via.placeholder.com/40x40?text=JD", Verified: true},
		{ID: "2", Username: "janesmith", Name: "Jane Smith", AvatarURL: "https://via.placeholder.com/40x40?text=JS", Verified: false},
		{ID: "3", Username: "mikejohnson", Name: "Mike Johnson", AvatarURL: "https://via.placeholder.com/40x40?text=MJ", Verified: true},
		{ID: "4", Username: "sarahwilson", Name: "Sarah Wilson", AvatarURL: "https://via.placeholder.com/40x40x40?text=SW", Verified: false},
	}

	// Sample posts
	samplePosts := []Post{
		{
			ID:        "1",
			Author:    users[0],
			Content:   "Just launched my new project! Excited to share it with the community. 🚀",
			Type:      PostTypeText,
			Timestamp: time.Now().Add(-2 * time.Hour),
			Likes:     42,
			Comments: []Comment{
				{
					ID:        "1",
					Author:    users[1],
					Content:   "Congratulations! This looks amazing!",
					Timestamp: time.Now().Add(-1 * time.Hour),
					Likes:     5,
					IsLiked:   false,
				},
			},
			Shares:   8,
			IsLiked:  false,
			IsShared: false,
		},
		{
			ID:        "2",
			Author:    users[1],
			Content:   "Beautiful sunset at the beach today! Nature never fails to amaze me. 🌅",
			Type:      PostTypeImage,
			MediaURL:  "https://via.placeholder.com/600x400?text=Sunset",
			Timestamp: time.Now().Add(-4 * time.Hour),
			Likes:     128,
			Comments: []Comment{
				{
					ID:        "2",
					Author:    users[2],
					Content:   "Stunning photo! Where was this taken?",
					Timestamp: time.Now().Add(-3 * time.Hour),
					Likes:     3,
					IsLiked:   true,
				},
				{
					ID:        "3",
					Author:    users[3],
					Content:   "Absolutely breathtaking! 😍",
					Timestamp: time.Now().Add(-2 * time.Hour),
					Likes:     7,
					IsLiked:   false,
				},
			},
			Shares:   15,
			IsLiked:  true,
			IsShared: false,
		},
		{
			ID:        "3",
			Author:    users[2],
			Content:   "Check out this tutorial on building reactive UIs with Go and WebAssembly!",
			Type:      PostTypeLink,
			MediaURL:  "https://via.placeholder.com/600x400?text=Tutorial",
			Timestamp: time.Now().Add(-6 * time.Hour),
			Likes:     89,
			Comments:  []Comment{},
			Shares:    23,
			IsLiked:   false,
			IsShared:  true,
		},
		{
			ID:        "4",
			Author:    users[3],
			Content:   "New video: My journey learning Go programming",
			Type:      PostTypeVideo,
			MediaURL:  "https://via.placeholder.com/600x400?text=Video",
			Timestamp: time.Now().Add(-8 * time.Hour),
			Likes:     67,
			Comments: []Comment{
				{
					ID:        "4",
					Author:    users[0],
					Content:   "Great video! Very helpful for beginners.",
					Timestamp: time.Now().Add(-7 * time.Hour),
					Likes:     12,
					IsLiked:   false,
				},
			},
			Shares:   11,
			IsLiked:  false,
			IsShared: false,
		},
	}

	sf.posts.Set(samplePosts)
	sf.currentUser.Set(users[0]) // Set current user to John Doe
}

func (sf *SocialFeed) render() g.Node {
	// Filter posts based on selected type
	filteredPosts := reactivity.CreateMemo(func() []Post {
		posts := sf.posts.Get()
		filter := sf.selectedFilter.Get()

		if filter == "" {
			return posts
		}

		var filtered []Post
		for _, post := range posts {
			if post.Type == filter {
				filtered = append(filtered, post)
			}
		}
		return filtered
	})

	return h.Div(
		h.Class("social-feed"),

		// Header with post composer
		h.Header(
			h.Class("feed-header"),
			sf.renderPostComposer(),

			// Filter tabs
			h.Nav(
				h.Class("post-filters"),
				h.Button(
					h.Class("filter-tab"),
					g.If(sf.selectedFilter.Get() == "", g.Attr("class", "filter-tab active")),
					g.Text("All"),
					dom.OnClickInline(func(el dom.Element) {
						sf.selectedFilter.Set("")
					}),
				),
				comps.For(comps.ForProps[PostType]{
					Items: reactivity.CreateSignal([]PostType{
						PostTypeText, PostTypeImage, PostTypeVideo, PostTypeLink,
					}),
					Key: func(pType PostType) string { return string(pType) },
					Children: func(pType PostType, index int) g.Node {
						isActive := reactivity.CreateMemo(func() bool {
							return sf.selectedFilter.Get() == pType
						})

						return h.Button(
							h.Class("filter-tab"),
							g.If(isActive.Get(), h.Class("active")),
							g.Text(strings.Title(string(pType))),
							dom.OnClickInline(func(el dom.Element) {
								sf.selectedFilter.Set(pType)
							}),
						)
					},
				}),
			),
		),

		// Posts feed
		h.Main(
			h.Class("feed-content"),
			comps.For(comps.ForProps[Post]{
				Items: filteredPosts,
				Key:   func(post Post) string { return post.ID },
				Children: func(post Post, index int) g.Node {
					return sf.renderPost(post)
				},
			}),

			// Loading indicator
			comps.Show(comps.ShowProps{
				When: sf.loading,
				Children: h.Div(
					h.Class("loading-indicator"),
					g.Text("Loading more posts..."),
				),
			}),

			// Load more button
			comps.Show(comps.ShowProps{
				When: reactivity.CreateMemo(func() bool {
					return sf.hasMore.Get() && !sf.loading.Get()
				}),
				Children: h.Button(
					h.Class("load-more"),
					g.Text("Load More"),
					dom.OnClickInline(func(el dom.Element) {
						sf.loadMorePosts()
					}),
				),
			}),
		),
	)
}

func (sf *SocialFeed) renderPost(post Post) g.Node {
	commentsVisible := reactivity.CreateMemo(func() bool {
		return sf.showComments.Get()[post.ID]
	})

	return h.Article(
		h.Class("post"),

		// Post header
		h.Header(
			h.Class("post-header"),
			h.Img(
				h.Class("avatar"),
				h.Src(post.Author.AvatarURL),
				h.Alt(post.Author.Name),
			),
			h.Div(
				h.Class("author-info"),
				h.H3(
					h.Class("author-name"),
					g.Text(post.Author.Name),
					comps.Show(comps.ShowProps{
						When: reactivity.CreateSignal(post.Author.Verified),
						Children: h.Span(
							h.Class("verified-badge"),
							g.Text("✓"),
						),
					}),
				),
				h.P(
					h.Class("username"),
					g.Text("@"+post.Author.Username),
				),
				h.Time(
					h.Class("timestamp"),
					g.Text(formatTimeAgo(post.Timestamp)),
				),
			),
		),

		// Post content
		h.Div(
			h.Class("post-content"),
			h.P(g.Text(post.Content)),

			// Media content based on post type
			comps.Switch(comps.SwitchProps{
				When: reactivity.CreateSignal(post.Type),
				Children: []g.Node{
					comps.Match(comps.MatchProps{
						When: PostTypeImage,
						Children: comps.Show(comps.ShowProps{
							When: reactivity.CreateSignal(post.MediaURL != ""),
							Children: h.Img(
								h.Class("post-image"),
								h.Src(post.MediaURL),
								h.Alt("Post image"),
							),
						}),
					}),
					comps.Match(comps.MatchProps{
						When: PostTypeVideo,
						Children: comps.Show(comps.ShowProps{
							When: reactivity.CreateSignal(post.MediaURL != ""),
							Children: h.Video(
								h.Class("post-video"),
								g.Attr("controls", ""),
								g.Attr("src", post.MediaURL),
							),
						}),
					}),
				},
			}),
		),

		// Post actions
		h.Footer(
			h.Class("post-actions"),
			h.Button(
				h.Class("action-button like-button"),
				g.If(post.IsLiked, h.Class("liked")),
				g.Text(fmt.Sprintf("♥ %d", post.Likes)),
				dom.OnClickInline(func(el dom.Element) {
					sf.toggleLike(post.ID)
				}),
			),

			h.Button(
				h.Class("action-button comment-button"),
				g.Text(fmt.Sprintf("💬 %d", len(post.Comments))),
				dom.OnClickInline(func(el dom.Element) {
					sf.toggleComments(post.ID)
				}),
			),

			h.Button(
				h.Class("action-button share-button"),
				g.If(post.IsShared, h.Class("shared")),
				g.Text(fmt.Sprintf("🔄 %d", post.Shares)),
				dom.OnClickInline(func(el dom.Element) {
					sf.sharePost(post.ID)
				}),
			),
		),

		// Comments section
		comps.Show(comps.ShowProps{
			When: commentsVisible,
			Children: h.Div(
				h.Class("comments-section"),
				comps.For(comps.ForProps[Comment]{
					Items: reactivity.CreateSignal(post.Comments),
					Key:   func(comment Comment) string { return comment.ID },
					Children: func(comment Comment, index int) g.Node {
						return sf.renderComment(comment)
					},
				}),

				// Comment composer
				h.Div(
					h.Class("comment-composer"),
					h.Textarea(
						h.Placeholder("Write a comment..."),
						h.Rows("2"),
					),
					h.Button(
						g.Text("Post Comment"),
						dom.OnClickInline(func(el dom.Element) {
							// Handle comment submission
						}),
					),
				),
			),
		}),
	)
}

func (sf *SocialFeed) renderComment(comment Comment) g.Node {
	return h.Div(
		h.Class("comment"),
		h.Img(
			h.Class("comment-avatar"),
			h.Src(comment.Author.AvatarURL),
			h.Alt(comment.Author.Name),
		),
		h.Div(
			h.Class("comment-content"),
			h.H4(
				h.Class("comment-author"),
				g.Text(comment.Author.Name),
			),
			h.P(g.Text(comment.Content)),
			h.Div(
				h.Class("comment-actions"),
				h.Time(
					h.Class("comment-time"),
					g.Text(formatTimeAgo(comment.Timestamp)),
				),
				h.Button(
					h.Class("comment-like"),
					g.If(comment.IsLiked, h.Class("liked")),
					g.Text(fmt.Sprintf("♥ %d", comment.Likes)),
				),
			),
		),
	)
}

func (sf *SocialFeed) renderPostComposer() g.Node {
	return h.Div(
		h.Class("post-composer"),
		h.Textarea(
			h.Class("composer-input"),
			h.Placeholder("What's on your mind?"),
			h.Value(sf.newPostContent.Get()),
			dom.OnInputInline(func(el dom.Element) {
				sf.newPostContent.Set(el.Underlying().Get("value").String())
			}),
		),
		h.Div(
			h.Class("composer-actions"),
			h.Button(
				h.Class("media-button"),
				g.Text("📷 Photo"),
			),
			h.Button(
				h.Class("media-button"),
				g.Text("🎥 Video"),
			),
			h.Button(
				h.Class("post-button"),
				g.Attr("disabled", func() string {
					if sf.newPostContent.Get() == "" {
						return "true"
					}
					return "false"
				}()),
				g.Text("Post"),
				dom.OnClickInline(func(el dom.Element) {
					sf.createPost()
				}),
			),
		),
	)
}

// Helper methods
func (sf *SocialFeed) toggleLike(postID string) {
	posts := sf.posts.Get()
	for i, post := range posts {
		if post.ID == postID {
			posts[i].IsLiked = !post.IsLiked
			if post.IsLiked {
				posts[i].Likes++
			} else {
				posts[i].Likes--
			}
			break
		}
	}
	sf.posts.Set(posts)
}

func (sf *SocialFeed) toggleComments(postID string) {
	current := sf.showComments.Get()
	current[postID] = !current[postID]
	sf.showComments.Set(current)
}

func (sf *SocialFeed) sharePost(postID string) {
	posts := sf.posts.Get()
	for i, post := range posts {
		if post.ID == postID {
			posts[i].IsShared = !post.IsShared
			if post.IsShared {
				posts[i].Shares++
			} else {
				posts[i].Shares--
			}
			break
		}
	}
	sf.posts.Set(posts)
}

func (sf *SocialFeed) loadMorePosts() {
	sf.loading.Set(true)

	// Simulate API call
	go func() {
		time.Sleep(1 * time.Second)

		// Add more sample posts
		newPosts := []Post{
			{
				ID:        fmt.Sprintf("new-%d", time.Now().Unix()),
				Author:    User{ID: "2", Username: "janesmith", Name: "Jane Smith", AvatarURL: "https://via.placeholder.com/40x40?text=JS", Verified: false},
				Content:   "Just discovered this amazing new framework! The future of web development is here. 🌟",
				Type:      PostTypeText,
				Timestamp: time.Now().Add(-1 * time.Hour),
				Likes:     34,
				Comments:  []Comment{},
				Shares:    5,
				IsLiked:   false,
				IsShared:  false,
			},
		}

		currentPosts := sf.posts.Get()
		sf.posts.Set(append(currentPosts, newPosts...))
		sf.hasMore.Set(false) // No more posts for demo
		sf.loading.Set(false)
	}()
}

func (sf *SocialFeed) createPost() {
	content := sf.newPostContent.Get()
	if content == "" {
		return
	}

	currentUser := sf.currentUser.Get()
	newPost := Post{
		ID:        fmt.Sprintf("post-%d", time.Now().Unix()),
		Author:    currentUser,
		Content:   content,
		Type:      PostTypeText,
		Timestamp: time.Now(),
		Likes:     0,
		Comments:  []Comment{},
		Shares:    0,
		IsLiked:   false,
		IsShared:  false,
	}

	posts := sf.posts.Get()
	sf.posts.Set(append([]Post{newPost}, posts...)) // Add to beginning
	sf.newPostContent.Set("")
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}

func main() {
	wasm.Initialize(wasm.InitConfig{})

	feed := NewSocialFeed()

	// Load sample data on startup
	reactivity.CreateEffect(func() {
		feed.loadSampleData()
	})

	comps.Mount("app", feed.render)

	select {}
}
