package action

import (
	"time"
)

// DispatchOption configures how an action is dispatched.
type DispatchOption interface {
	applyDispatch(*dispatchOptions)
}

// dispatchOptions holds the internal configuration for dispatch operations.
type dispatchOptions struct {
	timeout    time.Duration
	context    Context
	priority   int
	persistent bool
	async      bool
}

// WithTimeout sets a timeout for the dispatch operation.
func WithTimeout(timeout time.Duration) DispatchOption {
	return timeoutOption{timeout: timeout}
}

type timeoutOption struct {
	timeout time.Duration
}

func (o timeoutOption) applyDispatch(opts *dispatchOptions) {
	opts.timeout = o.timeout
}

// WithContext sets the context for the dispatch operation.
func WithContext(ctx Context) DispatchOption {
	return contextOption{context: ctx}
}

type contextOption struct {
	context Context
}

func (o contextOption) applyDispatch(opts *dispatchOptions) {
	opts.context = o.context
}

// WithMeta adds metadata to the dispatch operation.
func WithMeta(meta map[string]any) DispatchOption {
	return metaOption{meta: meta}
}

type metaOption struct {
	meta map[string]any
}

func (o metaOption) applyDispatch(opts *dispatchOptions) {
	if opts.context.Meta == nil {
		opts.context.Meta = make(map[string]any)
	}
	for k, v := range o.meta {
		opts.context.Meta[k] = v
	}
}

// WithTrace sets the trace ID for the dispatch operation.
func WithTrace(traceID string) DispatchOption {
	return traceOption{traceID: traceID}
}

type traceOption struct {
	traceID string
}

func (o traceOption) applyDispatch(opts *dispatchOptions) {
	opts.context.TraceID = o.traceID
}

// WithSource sets the source identifier for the dispatch operation.
func WithSource(source string) DispatchOption {
	return sourceOption{source: source}
}

type sourceOption struct {
	source string
}

func (o sourceOption) applyDispatch(opts *dispatchOptions) {
	opts.context.Source = o.source
}

// WithAsync schedules the dispatch to run asynchronously (microtask-like).
func WithAsync() DispatchOption {
	return asyncOption{}
}

type asyncOption struct{}

func (o asyncOption) applyDispatch(opts *dispatchOptions) {
	opts.async = true
}

// SubOption configures how a subscription is created.
type SubOption interface {
	applySub(*subOptions)
}

// subOptions holds the internal configuration for subscription operations.
type subOptions struct {
	priority             int
	once                 bool
	filter               func(any) bool
	whenSignal           interface{}
	distinctUntilChanged bool
	distinctEqualityFunc func(a, b any) bool
}

// SubWithPriority sets the priority for the subscription.
func SubWithPriority(priority int) SubOption {
	return priorityOption{priority: priority}
}

type priorityOption struct {
	priority int
}

func (o priorityOption) applySub(opts *subOptions) {
	opts.priority = o.priority
}

// SubOnce makes the subscription fire only once.
func SubOnce() SubOption {
	return onceOption{}
}

type onceOption struct{}

func (o onceOption) applySub(opts *subOptions) {
	opts.once = true
}

// SubFilter sets a filter function for the subscription.
func SubFilter(filter func(any) bool) SubOption {
	return filterOption{filter: filter}
}

type filterOption struct {
	filter func(any) bool
}

func (o filterOption) applySub(opts *subOptions) {
	opts.filter = o.filter
}

// SubWhen gates delivery based on a reactive signal.
func SubWhen(signal interface{}) SubOption {
	return whenOption{signal: signal}
}

type whenOption struct {
	signal interface{}
}

func (o whenOption) applySub(opts *subOptions) {
	opts.whenSignal = o.signal
}

// SubDistinctUntilChanged suppresses delivery when payload equals last delivered.
// If equalityFunc is nil, uses reflect.DeepEqual for comparison.
func SubDistinctUntilChanged(equalityFunc func(a, b any) bool) SubOption {
	return distinctUntilChangedOption{equalityFunc: equalityFunc}
}

type distinctUntilChangedOption struct {
	equalityFunc func(a, b any) bool
}

func (o distinctUntilChangedOption) applySub(opts *subOptions) {
	opts.distinctUntilChanged = true
	opts.distinctEqualityFunc = o.equalityFunc
}

// BridgeOption configures how actions are bridged between buses.
type BridgeOption interface {
	applyBridge(*bridgeOptions)
}

// bridgeOptions holds the internal configuration for bridge operations.
type bridgeOptions struct {
	filter               func(any) bool
	transform            func(any) any
	initialValue         any
	distinctUntilChanged bool
	distinctEqualityFunc func(a, b any) bool
	bufferSize           int
	dropPolicy           DropPolicy
}

// BridgeWithFilter sets a filter function for the bridge.
func BridgeWithFilter(filter func(any) bool) BridgeOption {
	return bridgeFilterOption{filter: filter}
}

type bridgeFilterOption struct {
	filter func(any) bool
}

func (o bridgeFilterOption) applyBridge(opts *bridgeOptions) {
	opts.filter = o.filter
}

// BridgeWithTransform sets a transform function for the bridge.
func BridgeWithTransform(transform func(any) any) BridgeOption {
	return bridgeTransformOption{transform: transform}
}

type bridgeTransformOption struct {
	transform func(any) any
}

func (o bridgeTransformOption) applyBridge(opts *bridgeOptions) {
	opts.transform = o.transform
}

// BridgeWithInitialValue sets the initial value for signal bridges.
func BridgeWithInitialValue(value any) BridgeOption {
	return bridgeInitialValueOption{value: value}
}

type bridgeInitialValueOption struct {
	value any
}

func (o bridgeInitialValueOption) applyBridge(opts *bridgeOptions) {
	opts.initialValue = o.value
}

// BridgeWithDistinctUntilChanged enables distinct until changed filtering for signal bridges.
func BridgeWithDistinctUntilChanged(equalityFunc func(a, b any) bool) BridgeOption {
	return bridgeDistinctOption{equalityFunc: equalityFunc}
}

type bridgeDistinctOption struct {
	equalityFunc func(a, b any) bool
}

func (o bridgeDistinctOption) applyBridge(opts *bridgeOptions) {
	opts.distinctUntilChanged = true
	opts.distinctEqualityFunc = o.equalityFunc
}

// BridgeWithBufferSize sets the buffer size for stream bridges.
func BridgeWithBufferSize(size int) BridgeOption {
	return bridgeBufferSizeOption{size: size}
}

type bridgeBufferSizeOption struct {
	size int
}

func (o bridgeBufferSizeOption) applyBridge(opts *bridgeOptions) {
	opts.bufferSize = o.size
}

// BridgeWithDropPolicy sets the drop policy for stream bridges when buffer is full.
func BridgeWithDropPolicy(policy DropPolicy) BridgeOption {
	return bridgeDropPolicyOption{policy: policy}
}

type bridgeDropPolicyOption struct {
	policy DropPolicy
}

func (o bridgeDropPolicyOption) applyBridge(opts *bridgeOptions) {
	opts.dropPolicy = o.policy
}

// ConcurrencyPolicy defines how concurrent queries to the same handler are handled.
type ConcurrencyPolicy int

const (
	// ConcurrencyOne allows only one query at a time; reject new queries while one is processing.
	ConcurrencyOne ConcurrencyPolicy = iota
	// ConcurrencyLatest cancels previous query when new one arrives.
	ConcurrencyLatest
	// ConcurrencyQueue processes queries in FIFO order.
	ConcurrencyQueue
)

// QueryOption configures how a query is handled.
type QueryOption interface {
	applyQuery(*queryOptions)
}

// queryOptions holds the internal configuration for query operations.
type queryOptions struct {
	timeout           time.Duration
	priority          int
	concurrencyPolicy ConcurrencyPolicy
}

// QueryWithTimeout sets a timeout for the query operation.
func QueryWithTimeout(timeout time.Duration) QueryOption {
	return queryTimeoutOption{timeout: timeout}
}

type queryTimeoutOption struct {
	timeout time.Duration
}

func (o queryTimeoutOption) applyQuery(opts *queryOptions) {
	opts.timeout = o.timeout
}

// QueryWithPriority sets the priority for the query handler.
func QueryWithPriority(priority int) QueryOption {
	return queryPriorityOption{priority: priority}
}

type queryPriorityOption struct {
	priority int
}

func (o queryPriorityOption) applyQuery(opts *queryOptions) {
	opts.priority = o.priority
}

// QueryWithConcurrencyPolicy sets the concurrency policy for the query handler.
func QueryWithConcurrencyPolicy(policy ConcurrencyPolicy) QueryOption {
	return queryConcurrencyPolicyOption{policy: policy}
}

type queryConcurrencyPolicyOption struct {
	policy ConcurrencyPolicy
}

func (o queryConcurrencyPolicyOption) applyQuery(opts *queryOptions) {
	opts.concurrencyPolicy = o.policy
}

// AskOption configures how a query is asked.
type AskOption interface {
	applyAsk(*askOptions)
}

// askOptions holds the internal configuration for ask operations.
type askOptions struct {
	timeout  time.Duration
	context  Context
	priority int
	traceID  string
	meta     map[string]any
	source   string
}

// AskWithTimeout sets a timeout for the ask operation.
func AskWithTimeout(timeout time.Duration) AskOption {
	return askTimeoutOption{timeout: timeout}
}

type askTimeoutOption struct {
	timeout time.Duration
}

func (o askTimeoutOption) applyAsk(opts *askOptions) {
	opts.timeout = o.timeout
}

// AskWithContext sets the context for the ask operation.
func AskWithContext(ctx Context) AskOption {
	return askContextOption{context: ctx}
}

type askContextOption struct {
	context Context
}

func (o askContextOption) applyAsk(opts *askOptions) {
	opts.context = o.context
}

// AskWithPriority sets the priority for the ask operation.
func AskWithPriority(priority int) AskOption {
	return askPriorityOption{priority: priority}
}

type askPriorityOption struct {
	priority int
}

func (o askPriorityOption) applyAsk(opts *askOptions) {
	opts.priority = o.priority
}

// AskWithTraceID sets the trace ID for the ask operation.
func AskWithTraceID(traceID string) AskOption {
	return askTraceIDOption{traceID: traceID}
}

type askTraceIDOption struct {
	traceID string
}

func (o askTraceIDOption) applyAsk(opts *askOptions) {
	opts.traceID = o.traceID
}

// AskWithMeta sets metadata for the ask operation.
func AskWithMeta(meta map[string]any) AskOption {
	return askMetaOption{meta: meta}
}

type askMetaOption struct {
	meta map[string]any
}

func (o askMetaOption) applyAsk(opts *askOptions) {
	opts.meta = o.meta
}

// AskWithSource sets the source identifier for the ask operation.
func AskWithSource(source string) AskOption {
	return askSourceOption{source: source}
}

type askSourceOption struct {
	source string
}

func (o askSourceOption) applyAsk(opts *askOptions) {
	opts.source = o.source
}
