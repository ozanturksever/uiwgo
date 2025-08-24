//go:build js && wasm

package main

import (
	"fmt"
	"math/rand"
	"syscall/js"
	"time"

	comps "github.com/ozanturksever/uiwgo/comps"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type User struct {
	ID   int
	Name string
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	// Mount the app and get a disposer function
	disposer := comps.Mount("app", func() Node { return UserProfileApp() })
	_ = disposer // We don't use it in this example since the app runs indefinitely
	
	// Prevent exit
	select {}
}
func UserProfileApp() Node {
	// Source signal: current user ID to load
	userID := reactivity.CreateSignal(1)

	// Resource: fetch user by id (simulated API with delay and possible error)
	userRes := reactivity.CreateResource(userID, fetchUser)

	// Expose actions for buttons
	setUser1 := js.FuncOf(func(this js.Value, args []js.Value) any { userID.Set(1); return nil })
	setUser2 := js.FuncOf(func(this js.Value, args []js.Value) any { userID.Set(2); return nil })
	randomUser := js.FuncOf(func(this js.Value, args []js.Value) any { userID.Set(1 + rand.Intn(3)); return nil })
	js.Global().Set("setUser1", setUser1)
	js.Global().Set("setUser2", setUser2)
	js.Global().Set("randomUser", randomUser)

	return Div(
		Style("font-family: Arial, sans-serif; max-width: 700px; margin: 40px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);"),
			H1(Text("Resource Example")),
			P(Text("Demonstrates CreateResource: loading, error handling, and re-fetch on source change")),

			// Controls
			Div(
				Style("display:flex; gap:10px; margin: 10px 0;"),
				Button(Text("Load User 1"), ID("load-user1-btn"), Attr("onclick", "window.setUser1()"), Style("padding:8px 12px")),
				Button(Text("Load User 2 (errors)"), ID("load-user2-btn"), Attr("onclick", "window.setUser2()"), Style("padding:8px 12px")),
				Button(Text("Random User"), ID("random-user-btn"), Attr("onclick", "window.randomUser()"), Style("padding:8px 12px")),
			),

			// Status and data rendering
			Div(
				ID("user-display"),
				Style("margin-top: 16px; padding: 16px; border: 1px solid #eee; border-radius: 8px; background:#fafafa;"),
				comps.BindHTML(func() Node {
					if userRes.Loading() {
						return P(Style("color:#555"), Text("Loading..."))
					}
					if err := userRes.Error(); err != nil {
						return P(Style("color:#c00"), Text(fmt.Sprintf("Error: %v", err)))
					}
					u := userRes.Data()
					// If no data yet (e.g., initial zero), show a hint
					if u.ID == 0 && u.Name == "" {
						return P(Style("color:#777; font-style:italic;"), Text("No user loaded yet"))
					}
					return Div(
						H2(Text(fmt.Sprintf("User #%d", u.ID))),
						P(Text(fmt.Sprintf("Name: %s", u.Name))),
					)
				}),
			),
		),
	)
}

// fetchUser simulates an asynchronous API call with latency and possible error for ID=2.
func fetchUser(id int) (User, error) {
	// Simulate network delay
	time.Sleep(900 * time.Millisecond)

	// Simulate an error for certain IDs
	if id == 2 {
		return User{}, fmt.Errorf("user %d not found", id)
	}
	// Simulate a successful response
	return User{ID: id, Name: fmt.Sprintf("User-%d", id)}, nil
}
