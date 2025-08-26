package reactivity

import "sync"

type Location struct {
	Pathname string
	Search   string
	Hash     string
}

type LocationState struct {
	mu          sync.RWMutex
	location    Location
	subscribers []func(Location)
}

func NewLocationState() *LocationState {
	return &LocationState{
		location: Location{}, // Zero value initialization
	}
}

func (s *LocationState) Get() Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.location
}

func (s *LocationState) Subscribe(subscriber func(Location)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers = append(s.subscribers, subscriber)
}

func (s *LocationState) Set(newLoc Location) {
	s.mu.Lock()
	s.location = newLoc
	subs := make([]func(Location), len(s.subscribers))
	copy(subs, s.subscribers)
	s.mu.Unlock()

	for _, sub := range subs {
		sub(newLoc)
	}
}