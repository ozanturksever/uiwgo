package router

import (
	"sync"
)

// Location holds the parsed components of a URL, mirroring the browser's location object.
type Location struct {
	Pathname string
	Search   string
	Hash     string
	State    any // Represents data passed via history.pushState
}

// Subscriber defines the function signature for any callback that wishes
// to be notified of location changes.
type Subscriber func(newLocation Location)

// LocationState is the central reactive store for the router's state.
// It is the single source of truth for the current location.
type LocationState struct {
	mu          sync.RWMutex
	current     Location
	subscribers []Subscriber
}

// NewLocationState creates and returns a new LocationState instance
// with zero-value initial state.
func NewLocationState() *LocationState {
	return &LocationState{
		current:     Location{},
		subscribers: make([]Subscriber, 0),
	}
}

// Get returns the current Location in a thread-safe manner.
func (ls *LocationState) Get() Location {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.current
}

// Subscribe adds a subscriber function to be notified of location changes.
func (ls *LocationState) Subscribe(s Subscriber) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.subscribers = append(ls.subscribers, s)
}

// Set updates the current location and notifies all subscribers synchronously.
func (ls *LocationState) Set(newLocation Location) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.current = newLocation
	for _, subscriber := range ls.subscribers {
		subscriber(newLocation)
	}
}
