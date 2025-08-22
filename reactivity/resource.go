package reactivity

// Resource provides reactive access to an asynchronously loaded value.
// It exposes three reactive getters backed by Signals:
//   - Data(): the loaded value (zero value until first success)
//   - Loading(): whether a fetch is in progress
//   - Error(): last error, or nil
//
// Notes:
// - This implementation is designed for JS/WASM single-threaded runtime.
// - Concurrent (stale) requests are ignored using a monotonically increasing token.
// - On error, the previous Data value is preserved.
//
// Inspired by SolidJS's createResource.

type Resource[T any] interface {
	Data() T
	Loading() bool
	Error() error
}

type resourceImpl[T any] struct {
	data    Signal[T]
	loading Signal[bool]
	err     Signal[error]

	// latestReq increments on each (re)fetch; completions check against it
	latestReq int
}

func (r *resourceImpl[T]) Data() T       { return r.data.Get() }
func (r *resourceImpl[T]) Loading() bool { return r.loading.Get() }
func (r *resourceImpl[T]) Error() error  { return r.err.Get() }

// CreateResource wires an asynchronous fetcher to a source signal.
// Whenever the source value changes, the fetcher is invoked in a goroutine
// and the resulting Data/Loading/Error signals are updated upon completion.
//
// Behavior:
// - Sets Loading(true) and clears Error before invoking fetcher.
// - Runs fetcher on a goroutine to avoid blocking the UI.
// - Only the latest request may update the signals; stale completions are ignored.
// - On error, Data remains as last successful value.
func CreateResource[S any, T any](source Signal[S], fetcher func(S) (T, error)) Resource[T] {
	r := &resourceImpl[T]{
		data:    CreateSignal(*new(T)), // zero T
		loading: CreateSignal(false),
		err:     CreateSignal(error(nil)),
	}

	// Track source changes and trigger fetches
	CreateEffect(func() {
		s := source.Get() // track dependency

		// Prepare for a new request
		r.latestReq++
		reqID := r.latestReq
		r.loading.Set(true)
		r.err.Set(nil)

		// Fire the fetch in a goroutine; ignore results if stale
		go func(val S, id int) {
			data, e := fetcher(val)
			// Only apply if this is the latest request
			if id != r.latestReq {
				return
			}
			if e != nil {
				r.err.Set(e)
			} else {
				r.data.Set(data)
			}
			r.loading.Set(false)
		}(s, reqID)
	})

	return r
}
