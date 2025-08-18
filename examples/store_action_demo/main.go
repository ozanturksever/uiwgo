// examples/store_action_demo/main.go
// Comprehensive demonstration of store and action patterns

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"app/golid"
)

// ------------------------------------
// 📊 User Management Example
// ------------------------------------

// User represents a user in our application
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// UserState represents the state of our user management
type UserState struct {
	Users       []User `json:"users"`
	CurrentUser *User  `json:"current_user"`
	Loading     bool   `json:"loading"`
	Error       string `json:"error"`
}

// ------------------------------------
// 🎯 User Actions
// ------------------------------------

// CreateUserPayload represents the payload for creating a user
type CreateUserPayload struct {
	Name  string
	Email string
}

// UpdateUserPayload represents the payload for updating a user
type UpdateUserPayload struct {
	ID    int
	Name  string
	Email string
}

// ------------------------------------
// 🏪 Store Setup
// ------------------------------------

func setupUserStore() (*golid.Store[UserState], *golid.ActionDispatcher) {
	// Create initial state
	initialState := UserState{
		Users:       []User{},
		CurrentUser: nil,
		Loading:     false,
		Error:       "",
	}

	// Create store with middleware
	loggingMiddleware := golid.NewLoggingMiddleware[UserState]("UserStore")
	validationMiddleware := golid.NewValidationMiddleware("UserStore", func(state UserState) error {
		if len(state.Users) > 100 {
			return fmt.Errorf("too many users")
		}
		return nil
	})

	store := golid.CreateStore(initialState, golid.StoreOptions[UserState]{
		Name: "UserStore",
		Middleware: []golid.StoreMiddleware[UserState]{
			loggingMiddleware,
			validationMiddleware,
		},
	})

	// Create action dispatcher
	dispatcher := golid.CreateDispatcher(
		golid.NewDispatcherLoggingMiddleware(),
		golid.NewDispatcherPerformanceMiddleware(),
	)

	// Register actions
	setupUserActions(store, dispatcher)

	return store, dispatcher
}

func setupUserActions(store *golid.Store[UserState], dispatcher *golid.ActionDispatcher) {
	// Create User Action
	createUserAction := golid.CreateAction(func(payload CreateUserPayload) UserState {
		currentState := store.Get()

		// Simulate user creation
		newUser := User{
			ID:       len(currentState.Users) + 1,
			Name:     payload.Name,
			Email:    payload.Email,
			IsActive: true,
		}

		return UserState{
			Users:       append(currentState.Users, newUser),
			CurrentUser: currentState.CurrentUser,
			Loading:     false,
			Error:       "",
		}
	}, golid.ActionOptions[CreateUserPayload, UserState]{
		Name: "CreateUser",
		Middleware: []golid.ActionMiddleware[CreateUserPayload, UserState]{
			golid.NewActionLoggingMiddleware[CreateUserPayload, UserState]("CreateUser"),
			golid.NewActionValidationMiddleware[CreateUserPayload, UserState]("CreateUser",
				func(payload CreateUserPayload) error {
					if payload.Name == "" {
						return fmt.Errorf("name is required")
					}
					if payload.Email == "" {
						return fmt.Errorf("email is required")
					}
					return nil
				}, nil),
		},
	})

	// Update User Action
	updateUserAction := golid.CreateAction(func(payload UpdateUserPayload) UserState {
		currentState := store.Get()
		users := make([]User, len(currentState.Users))
		copy(users, currentState.Users)

		// Find and update user
		for i, user := range users {
			if user.ID == payload.ID {
				users[i].Name = payload.Name
				users[i].Email = payload.Email
				break
			}
		}

		return UserState{
			Users:       users,
			CurrentUser: currentState.CurrentUser,
			Loading:     false,
			Error:       "",
		}
	}, golid.ActionOptions[UpdateUserPayload, UserState]{
		Name: "UpdateUser",
	})

	// Delete User Action
	deleteUserAction := golid.CreateAction(func(userID int) UserState {
		currentState := store.Get()
		users := make([]User, 0)

		// Filter out deleted user
		for _, user := range currentState.Users {
			if user.ID != userID {
				users = append(users, user)
			}
		}

		return UserState{
			Users:       users,
			CurrentUser: currentState.CurrentUser,
			Loading:     false,
			Error:       "",
		}
	}, golid.ActionOptions[int, UserState]{
		Name: "DeleteUser",
	})

	// Set Loading Action
	setLoadingAction := golid.CreateAction(func(loading bool) UserState {
		currentState := store.Get()
		return UserState{
			Users:       currentState.Users,
			CurrentUser: currentState.CurrentUser,
			Loading:     loading,
			Error:       currentState.Error,
		}
	}, golid.ActionOptions[bool, UserState]{
		Name: "SetLoading",
	})

	// Async Fetch Users Action
	fetchUsersAction := golid.CreateAsyncAction(func(ctx context.Context, _ struct{}) (UserState, error) {
		// Simulate API call
		time.Sleep(100 * time.Millisecond)

		// Mock users
		users := []User{
			{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true},
			{ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: true},
			{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", IsActive: false},
		}

		currentState := store.Get()
		return UserState{
			Users:       users,
			CurrentUser: currentState.CurrentUser,
			Loading:     false,
			Error:       "",
		}, nil
	}, golid.AsyncActionOptions[struct{}, UserState]{
		Name: "FetchUsers",
		Middleware: []golid.AsyncActionMiddleware[struct{}, UserState]{
			golid.NewAsyncActionLoggingMiddleware[struct{}, UserState]("FetchUsers"),
		},
	})

	// Register actions with dispatcher
	dispatcher.RegisterAction("createUser", createUserAction)
	dispatcher.RegisterAction("updateUser", updateUserAction)
	dispatcher.RegisterAction("deleteUser", deleteUserAction)
	dispatcher.RegisterAction("setLoading", setLoadingAction)
	dispatcher.RegisterAsyncAction("fetchUsers", fetchUsersAction)
}

// ------------------------------------
// 🔗 Derived Stores Example
// ------------------------------------

func setupDerivedStores(userStore *golid.Store[UserState]) {
	// Active users store
	activeUsersStore := golid.CreateDerivedStore(func() []User {
		state := userStore.Get()
		activeUsers := make([]User, 0)
		for _, user := range state.Users {
			if user.IsActive {
				activeUsers = append(activeUsers, user)
			}
		}
		return activeUsers
	}, golid.StoreOptions[[]User]{
		Name: "ActiveUsersStore",
	})

	// User count store
	userCountStore := golid.CreateDerivedStore(func() int {
		state := userStore.Get()
		return len(state.Users)
	}, golid.StoreOptions[int]{
		Name: "UserCountStore",
	})

	// Loading state store
	loadingStore := golid.CreateDerivedStore(func() bool {
		state := userStore.Get()
		return state.Loading
	}, golid.StoreOptions[bool]{
		Name: "LoadingStore",
	})

	// Subscribe to derived stores
	activeUsersStore.Subscribe(func(users []User) {
		fmt.Printf("Active users updated: %d users\n", len(users))
	})

	userCountStore.Subscribe(func(count int) {
		fmt.Printf("User count updated: %d\n", count)
	})

	loadingStore.Subscribe(func(loading bool) {
		fmt.Printf("Loading state: %t\n", loading)
	})
}

// ------------------------------------
// 💾 Persistence Example
// ------------------------------------

func setupPersistence(store *golid.Store[UserState]) (*golid.PersistentStore[UserState], error) {
	// Create persistent store with auto-save
	persistentStore, err := golid.PersistStore(store, golid.PersistenceOptions{
		Key:      "user_state",
		AutoSave: true,
		Throttle: 500 * time.Millisecond,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create persistent store: %w", err)
	}

	fmt.Println("Persistence enabled with auto-save")
	return persistentStore, nil
}

// ------------------------------------
// 🧪 Store Composition Example
// ------------------------------------

func setupStoreComposition(userStore *golid.Store[UserState]) {
	// Create a notification store
	notificationStore := golid.CreateStore([]string{}, golid.StoreOptions[[]string]{
		Name: "NotificationStore",
	})

	// Create a combined app state store
	appStateStore := golid.CombineStores(func() map[string]interface{} {
		userState := userStore.Get()
		notifications := notificationStore.Get()

		return map[string]interface{}{
			"users":         userState,
			"notifications": notifications,
			"timestamp":     time.Now(),
		}
	}, golid.StoreOptions[map[string]interface{}]{
		Name: "AppStateStore",
	})

	// Subscribe to combined state
	appStateStore.Subscribe(func(state map[string]interface{}) {
		fmt.Printf("App state updated at: %v\n", state["timestamp"])
	})

	// Add some notifications
	notificationStore.Set([]string{"Welcome!", "New user created"})
}

// ------------------------------------
// 📊 Performance Monitoring
// ------------------------------------

func monitorPerformance(dispatcher *golid.ActionDispatcher) {
	// Get performance middleware
	for _, middleware := range []golid.DispatcherMiddleware{} {
		if perfMiddleware, ok := middleware.(*golid.DispatcherPerformanceMiddleware); ok {
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				for range ticker.C {
					stats := perfMiddleware.GetStats()
					fmt.Println("\n=== Performance Stats ===")
					for actionName, stat := range stats {
						fmt.Printf("%s: %d ops, avg: %v, max: %v\n",
							actionName, stat.TotalOperations, stat.AverageDuration, stat.MaxDuration)
					}
					fmt.Println("========================\n")
				}
			}()
			break
		}
	}
}

// ------------------------------------
// 🎮 Demo Execution
// ------------------------------------

func main() {
	fmt.Println("🏪 Golid Store & Action Demo")
	fmt.Println("============================")

	// Create root context for proper cleanup
	_, cleanup := golid.CreateRoot(func() interface{} {
		// Setup user store and actions
		userStore, dispatcher := setupUserStore()

		// Setup derived stores
		setupDerivedStores(userStore)

		// Setup persistence
		persistentStore, err := setupPersistence(userStore)
		if err != nil {
			log.Printf("Persistence setup failed: %v", err)
		}

		// Setup store composition
		setupStoreComposition(userStore)

		// Monitor performance
		monitorPerformance(dispatcher)

		// Subscribe to main store changes
		userStore.Subscribe(func(state UserState) {
			fmt.Printf("User store updated: %d users, loading: %t\n",
				len(state.Users), state.Loading)
		})

		// Demo actions
		fmt.Println("\n🎯 Executing Actions...")

		// Set loading
		dispatcher.Dispatch("setLoading", true)

		// Fetch users (async)
		fmt.Println("Fetching users...")
		result, err := dispatcher.Dispatch("fetchUsers", struct{}{})
		if err != nil {
			fmt.Printf("Fetch users failed: %v\n", err)
		} else {
			userStore.Set(result.(UserState))
		}

		// Create new user
		fmt.Println("Creating new user...")
		result, err = dispatcher.Dispatch("createUser", CreateUserPayload{
			Name:  "Alice Cooper",
			Email: "alice@example.com",
		})
		if err != nil {
			fmt.Printf("Create user failed: %v\n", err)
		} else {
			userStore.Set(result.(UserState))
		}

		// Update user
		fmt.Println("Updating user...")
		result, err = dispatcher.Dispatch("updateUser", UpdateUserPayload{
			ID:    1,
			Name:  "John Updated",
			Email: "john.updated@example.com",
		})
		if err != nil {
			fmt.Printf("Update user failed: %v\n", err)
		} else {
			userStore.Set(result.(UserState))
		}

		// Delete user
		fmt.Println("Deleting user...")
		result, err = dispatcher.Dispatch("deleteUser", 2)
		if err != nil {
			fmt.Printf("Delete user failed: %v\n", err)
		} else {
			userStore.Set(result.(UserState))
		}

		// Show final state
		finalState := userStore.Get()
		fmt.Printf("\n📊 Final State: %d users\n", len(finalState.Users))
		for _, user := range finalState.Users {
			fmt.Printf("  - %s (%s) [Active: %t]\n", user.Name, user.Email, user.IsActive)
		}

		// Show store info
		fmt.Printf("\n📈 Store Info: %+v\n", userStore.GetStoreInfo())

		// Show persistence info
		if persistentStore != nil {
			fmt.Printf("💾 Last Save: %v\n", persistentStore.GetLastSaveTime())
		}

		return nil
	})

	// Wait a bit to see async operations
	time.Sleep(2 * time.Second)

	// Cleanup
	cleanup()
	fmt.Println("\n✅ Demo completed successfully!")
}
