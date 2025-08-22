package reactivity

// OnCleanup registers a function to be called before the current effect
// re-executes and when it is disposed. If called outside of an effect,
// it is ignored in this MVP implementation.
func OnCleanup(fn func()) {
	if currentEffect == nil || currentEffect.disposed {
		return
	}
	currentEffect.cleanups = append(currentEffect.cleanups, fn)
}
