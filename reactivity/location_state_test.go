package reactivity

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLocationState_InitialGetReturnsZeroValue(t *testing.T) {
	state := NewLocationState()
	initial := state.Get()

	assert.Equal(t, "", initial.Pathname, "Expected empty Pathname")
	assert.Equal(t, "", initial.Search, "Expected empty Search")
	assert.Equal(t, "", initial.Hash, "Expected empty Hash")
}

func TestLocationState_SubscribeAddsSubscriber(t *testing.T) {
	state := NewLocationState()
	called := false

	state.Subscribe(func(Location) {
		called = true
	})

	state.Set(Location{Pathname: "/test"})
	assert.True(t, called, "Subscriber should be called on state change")
}