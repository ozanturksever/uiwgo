UIwGo Action System Implementation Plan (ASTODOPLAN)

Purpose
- Provide a detailed, testable roadmap to implement the Action System.
- Ensure each deliverable has clear tasks, acceptance criteria, and tests.
- Cover core bus, scoping, reactive bridges, helpers (debounce, throttle, once, etc.), queries, lifecycle, and observability.

Deliverables Overview (Epics)
E1. Core Types and Package Skeleton
E2. Global Bus and Scoping
E3. Dispatch Path and Ordering
E4. Subscription Lifecycle and Options
E5. Reactive Bridges (ToSignal/ToStream)
E6. Query/Ask-Handle Pattern
E7. Lifecycle Integration Helpers
E8. Observability, Errors, and Tracing
E9. Performance and Batching Integration
E10. Developer Utilities and Test Harness
E11. Documentation and Examples
E12. Integration and Regression Test Suite

General Engineering Principles
- Type safety via generics for payloads.
- Sync ordered delivery by default; async available.
- At-most-once per subscriber per dispatch.
- Error isolation; no handler can block others.
- Deterministic tests; abstract time for debounce/throttle/sample.

E1. Core Types and Package Skeleton
Tasks
1. Create package: action.
2. Define ActionType[T] (Name string) and constructor DefineAction[T](name string).
3. Define Action[T] with fields: Type string, Payload T, Meta map[string]any, Time time.Time, Source string, TraceID string.
4. Define Context struct: Scope string, Meta map[string]any, Time time.Time, TraceID string, Source string; With methods for Meta lookup.
5. Define Bus interface and default implementation (unexported struct). Methods: Dispatch, Subscribe, SubscribeAny, Scope, ToSignal, ToStream, HandleQuery, Ask, OnError.
6. Define Subscription interface: Dispose() error, IsActive() bool.
7. Define options types: DispatchOption, SubOption, BridgeOption, QueryOption, AskOption with internal option structs.
8. Provide errors: ErrNoHandler, ErrTimeout, ErrDisposed.
9. Create a NoOpSubscription implementation for safe returns in edge cases.

Acceptance Criteria
- All types compile; no TODOs in public API signatures.
- Basic ctor and getters unit tested.

Tests
- TestDefineActionReturnsStableName
- TestActionMetadataDefaults
- TestContextMetaAccessors

E2. Global Bus and Scoping
Tasks
1. Implement Global() singleton bus with thread-safe lazy init.
2. Implement New() for local bus creation.
3. Implement Scope(name string) *Bus: returns a new bus instance with inherited error handler and a concatenated scope path; maintains independent subscriber registries.
4. Define scope representation (e.g., "root/a/b") and ensure immutability of parent bus.
5. Ensure subscriptions and dispatch are confined to their bus instance.

Acceptance Criteria
- Global bus accessible and unique.
- Scoped bus creates isolated registries; no cross-delivery.
- Scope string is exposed via Context.Scope.

Tests
- TestGlobalBusSingleton
- TestScopedBusIsolation_NoDeliveryAcrossScopes
- TestScopePathComposition_ParentChild
- TestSubscribeAndDispatchWithinScopeOnly

E3. Dispatch Path and Ordering
Tasks
1. Implement Dispatch[T](typ, payload, opts...) with:
   - Build Action[T] + Context.
   - Lookup subscribers by action type name.
   - Order: priority desc, then FIFO subscription order.
   - Synchronous default; WithAsync schedules microtask-like dispatch.
2. Apply DispatchOptions: WithMeta, WithTrace, WithSource, WithAsync.
3. Implement SubscribeAny delivery for Action[any].
4. Error handling: recover panics; call error hook; continue other handlers.

Acceptance Criteria
- Synchronous dispatch executes handlers before return by default.
- Deterministic ordering: priority then subscription order.
- SubscribeAny receives all actions.

Tests
- TestDispatchSyncOrdering_PriorityThenFIFO
- TestDispatchWithAsync_SchedulesLater
- TestDispatchWithMetaTraceSource_PropagatesToContext
- TestSubscribeAny_ReceivesAllActions
- TestHandlerPanic_IsolatedAndErrorHookInvoked

E4. Subscription Lifecycle and Options
Tasks
1. Implement Subscribe[T] returning Subscription with Dispose/IsActive; add internal ID and creation time.
2. Implement SubOption: Once()
   - Auto-dispose after first successful delivery.
3. Implement SubOption: Filter(predicate)
   - Skip delivery when predicate returns false.
4. Implement SubOption: WithPriority(n)
   - Maintain internal sorted list; insert/subscribe accordingly.
5. Implement SubOption: When(sig *Signal[bool])
   - Delivery only when sig is true; no buffering by default.
6. Implement SubOption: Buffered(n int, dropPolicy: oldest|newest)
   - Queue per subscription; used with async delivery loop.
7. Implement SubOption: Replay(n int)
   - Maintain ring buffer per action type in bus; deliver last n on subscribe, respecting Filter/When at delivery time.
8. Implement SubOption: DeliverAsync()
   - Override per-subscription to deliver asynchronously.
9. Implement SubOption: Debounce(d), Throttle(d with leading/trailing), SampleEvery(d)
   - Requires time abstraction (see E10 FakeClock); implement per-subscription scheduler.
10. Implement SubOption: DistinctUntilChanged(optional equality func)
   - Suppress payload delivery when equals to last delivered.

Acceptance Criteria
- All options co-exist predictably; When gating + Debounce/Throttle semantics documented (see tests).
- Dispose prevents further deliveries, including queued ones when feasible.
- Replay replays only buffered actions within same bus.

Tests
- TestOnce_DisposesAfterFirstDelivery
- TestFilter_SkipsWhenFalse
- TestWithPriority_RespectedInOrdering
- TestWhen_GatesDeliveryTrueOnly
- TestBuffered_DropOldestPolicy
- TestBuffered_DropNewestPolicy
- TestReplay_ReplaysLastNOnSubscribe
- TestDeliverAsync_PerSubscriptionAsync
- TestDebounce_Basic
- TestDebounce_MultipleBursts
- TestThrottle_LeadingTrueTrailingFalse
- TestThrottle_LeadingFalseTrailingTrue
- TestSampleEvery_PeriodicDelivery
- TestDistinctUntilChanged_DefaultEquality
- TestDistinctUntilChanged_CustomEquality
- TestDispose_StopsDeliveryImmediately

E5. Reactive Bridges (ToSignal/ToStream)
Tasks
1. Implement ToSignal[T](typ, BridgeOptions...) -> *Signal[T]
   - Initial value support.
   - DistinctUntilChanged per BridgeOption.
   - Optional Filter/Map transformation.
2. Implement ToStream[T](typ, BridgeOptions...) -> Stream[T]
   - Stream abstraction: Recv()/TryRecv()/Dispose(), buffer size in options.
   - Backpressure via bounded buffer and drop policy options.
3. Ensure bridge subscriptions auto-dispose when the signal/stream is disposed.
4. Ensure reactivity batching support:
   - If in a batch, signal updates are coalesced.

Acceptance Criteria
- Signal updates reflect last seen payload respecting distinct/filter/map.
- Stream behaves as a typed buffered channel with disposal.

Tests
- TestToSignal_InitialValue
- TestToSignal_UpdatesOnDispatch
- TestToSignal_DistinctUntilChanged
- TestToSignal_FilterMap
- TestToStream_BasicRecv
- TestToStream_BufferAndDropPolicy
- TestBridge_DisposeCleansSubscription

E6. Query/Ask-Handle Pattern
Tasks
1. Define QueryType[Req, Res] with DefineQuery(name string).
2. Implement HandleQuery(qt, handler, opts...) -> Subscription
   - One active handler per QueryType per bus; replace or error (decide: replace by default with a warning).
3. Implement Ask(qt, req, opts...) -> Future[Res]
   - Future supports Then, Catch, Await, Done.
   - Timeout option; error on ErrNoHandler or ErrTimeout.
4. QueryOption: Concurrency policy (One|Latest|Queue) for handler reentrancy control.
5. AskOption: Trace/Meta/Source pass-through.

Acceptance Criteria
- Exactly one handler effective per QueryType per bus.
- Ask resolves deterministically; respects timeout and concurrency policy.

Tests
- TestHandleQuery_RegisterAndAnswer
- TestHandleQuery_ReplaceExisting
- TestAsk_NoHandler_Error
- TestAsk_Timeout
- TestAsk_Concurrency_One
- TestAsk_Concurrency_LatestCancelsPrevious
- TestAsk_Concurrency_QueueProcessesInOrder
- TestAsk_PropagatesMetaTraceSource

E7. Lifecycle Integration Helpers
Tasks
1. Implement helper: AutoSubscribe(subscribeFn) that returns a Disposable tied to component lifecycle (OnMount/OnUnmount hooks).
2. Implement helper: WithLifecycleDispose(sub Subscription) to register cleanup automatically.
3. Provide When(sig *Signal[bool]) option that internally hooks unmount to detach watchers and avoid leaks.
4. Provide high-level helpers:
   - OnAction[T](bus, type, handler, opts...) that auto-registers and disposes on unmount.
   - UseActionSignal[T](bus, type, opts...) that returns a signal and ensures proper cleanup.

Acceptance Criteria
- Subscriptions created during mount are disposed on unmount without dev effort.
- No leaks under repeated mount/unmount cycles.

Tests
- TestLifecycle_AutoDisposeOnUnmount
- TestLifecycle_ReMountCreatesNewActiveSubscription
- TestWhen_StopsListeningWhenSignalFalseAndResumesTrue
- TestNoLeak_MultipleMountUnmountCycles (memory/GC sensitive; at least verify IsActive toggling)

E8. Observability, Errors, and Tracing
Tasks
1. Implement bus.OnError(h func(ctx Context, err error, recovered any)) to set hook.
2. Integrate tracing fields into Action and Context.
3. Add optional dev logger middleware: Log all dispatches with time, type, subscribers, duration, and errors.
4. Add debug ring buffer per action type (size configurable) for replay and diagnostics.
5. Add “tap” SubscribeAny helper for analytics with safe filters.

Acceptance Criteria
- Recovering panics calls error hook with context populated.
- Dev logger can be toggled without affecting behavior.

Tests
- TestOnError_InvokedOnPanicWithContext
- TestDevLogger_DoesNotChangeDelivery
- TestDebugRingBuffer_SizeBounded

E9. Performance and Batching Integration
Tasks
1. Integrate with reactive batching API to coalesce multiple signal updates during dispatch bursts.
2. Minimize allocations: reuse buffers, pre-size slices, avoid per-dispatch heap where possible.
3. Microtask scheduler for async dispatch with a minimal queue.
4. Benchmarks:
   - DispatchSingleSubscriber
   - DispatchManySubscribers
   - DebounceWithHighFrequency
   - ThrottleScrollingLikePattern
5. Profiling hooks for development builds.

Acceptance Criteria
- No significant regressions across common patterns.
- Benchmarks documented with thresholds.

Benchmarks (Non-blocking CI)
- BenchmarkDispatch_Single_1K
- BenchmarkDispatch_1KSubscribers
- BenchmarkDebounce_10KEvents

E10. Developer Utilities and Test Harness
Tasks
1. FakeClock abstraction to control time; support Sleep, Now, After, Timer/Ticker semantics.
2. TestBus helper to create isolated buses with fake clock and deterministic async scheduler.
3. Future test helper: Await with timeout for async tests.
4. Common matchers/assert helpers for action equality and ordering.

Acceptance Criteria
- All time-based helpers deterministic under FakeClock.
- Tests do not rely on real time.

Tests
- TestFakeClock_AdvanceTriggersDebounceThrottle
- TestTestBus_Isolation
- TestFuture_AwaitThenCatch

E11. Documentation and Examples
Tasks
1. Author Action System quickstart and API reference.
2. Write cookbook for patterns: search debounce, global notifications, request/response queries, scoped chat.
3. Provide minimal examples:
   - Counter with action-driven increment.
   - Search input with debounced query.
   - Modal open/close via actions.
   - Query example fetching data and rendering result.

Acceptance Criteria
- Examples compile and run.
- Docs cover all public API with concise samples.

Tests
- Example compile checks.
- Smoke tests to ensure no panics on basic interactions.

E12. Integration and Regression Test Suite
Tasks
1. Integration tests covering:
   - Multiple components communicating via bus.
   - Scoped buses with overlapping action types.
   - DOM event -> action -> reactive UI updates.
2. Regression tests for previously fixed bugs (as they appear).
3. Fuzz tests for delivery order and handler panics.

Acceptance Criteria
- Integration test matrix passes under CI-like environment.
- No flakiness (use FakeClock and deterministic scheduling).

Tests
- TestIntegration_ComponentToComponent_ActionFlow
- TestIntegration_ScopedBuses_DoNotLeak
- TestIntegration_DOMEventBridgedToActionUpdatesSignal
- FuzzDispatch_PanicRecoveryAndOrdering

Detailed Helper Implementation Notes and Test Cases

Once()
- Implementation: wrapper sets a delivered flag; after first pass, calls Dispose.
- Tests:
  - Delivers exactly once.
  - Does not deliver after resubscribe without Once.

Filter(predicate)
- Implementation: check predicate(Action[T]) before invoking handler.
- Tests:
  - Predicate false suppresses delivery.
  - Predicate true allows delivery.

When(signal)
- Implementation: holds reference to boolean signal; delivery only when Get() == true at delivery time; optional hook to update internal state on signal changes.
- Tests:
  - Toggle signal to enable/disable.
  - Combined with Debounce: only debounced firing when active.

WithPriority(n)
- Implementation: maintain sorted slice by priority desc; stable by creation order.
- Tests:
  - Higher priority runs first.
  - Same priority preserves FIFO.

Buffered(n, policy)
- Implementation: per-subscription ring buffer. On async deliver loop, read and dispatch in order.
- Tests:
  - Overflow with drop-oldest retains most recent n.
  - Overflow with drop-newest retains earliest n.

Replay(n)
- Implementation: per action type ring buffer in bus. On subscribe, copy last n and deliver in-order before new live events.
- Tests:
  - Replays up to n.
  - Respects Filter and When at replay delivery time.

DeliverAsync()
- Implementation: per-subscription scheduling; handler dispatch on next tick; can share scheduler with WithAsync.
- Tests:
  - Delivery occurs after Dispatch returns.
  - Ordering preserved among async subscribers with same tick.

Debounce(d)
- Implementation: schedule timer keyed per subscription; restart on new events; emit last payload on idle.
- Tests:
  - Single emit after burst.
  - New burst resets timer.

Throttle(d, leading/trailing)
- Implementation: track window; leading emit immediate if allowed; trailing captures last and emits on window end if configured.
- Tests:
  - Leading only: first of window emitted, rest suppressed.
  - Trailing only: last of burst emitted after window.
  - Leading+Trailing: both ends emitted, no duplicates.

SampleEvery(d)
- Implementation: periodic ticker; deliver the most recent payload if any since last tick.
- Tests:
  - Emits at sample ticks only.
  - No emits if no new payloads.

DistinctUntilChanged(equal)
- Implementation: store last emitted payload; compare with equality function.
- Tests:
  - Suppresses equal payload deliveries.
  - Emits when changed.

Bridges

ToSignal
- Implementation: Subscribe with optional distinct/filter/map; update typed signal; return signal with a Dispose that disposes underlying subscription.
- Tests:
  - Reflects last action payload.
  - Distinct prevents redundant updates.
  - Cleanup releases subscription.

ToStream
- Implementation: Subscribe with buffered channel-like structure and configured size/policy; TryRecv/Recv operations.
- Tests:
  - Respects buffer.
  - Drop policy works.
  - Dispose closes stream and unsubscribes.

Queries

HandleQuery and Ask
- Implementation: per-QueryType registry with a single active handler; Ask issues a request and resolves a Future; apply timeout and concurrency.
- Tests:
  - Success path returns expected result.
  - No handler errors.
  - Timeout behavior deterministic via FakeClock.
  - Concurrency modes: One (reject/queue), Latest (cancel previous), Queue (FIFO).

Observability

OnError Hook
- Implementation: recover panic; pass context plus recovered value to hook; include stack trace optionally in debug builds.
- Tests:
  - Panic triggers hook.
  - Hook does not interfere with other subscribers.

Dev Logger
- Implementation: middleware-like function on bus to trace dispatches.
- Tests:
  - Non-intrusive; verify it records expected fields.

Performance Notes and Micro-Designs
- Registry: map[string][]subscriber; subscribers hold priority, creation index, options, buffers.
- Iteration: use snapshot copy of slice for stable traversal during dispatch.
- Allocation control: reuse buffers via pools for hot paths (optional).
- Async scheduling: a single goroutine/queue for bus or per-subscription lightweight queues depending on complexity.

Work Breakdown and Timeline (Suggested)
- Week 1: E1, E2 base; E3 sync dispatch; basic tests.
- Week 2: E4 core options (Once, Filter, Priority, When); tests; E8 error hook.
- Week 3: E4 advanced options (Buffered, Replay, Debounce, Throttle, Sample, Distinct); FakeClock (E10); tests.
- Week 4: E5 bridges; E6 queries; tests.
- Week 5: E7 lifecycle helpers; E9 performance/batching; benchmarks; docs (E11).
- Week 6: E12 integration tests; polish; finalize docs and examples.

Definition of Done Checklist
- Public API stable and documented.
- 95%+ coverage for action package (excluding benchmarks).
- All time-dependent tests use FakeClock; no flakes.
- Benchmarks recorded and reviewed.
- Integration tests pass under CI-like environment.
- Examples validated manually and via smoke tests.

Risk and Mitigation
- Time flakiness: use FakeClock and deterministic scheduler.
- Complexity creep in options: keep clear precedence rules and document.
- Memory pressure with buffers: default to small sizes; expose config; test drops.

Appendix: Proposed Public API Summary (Reference)
- Types: Action[T], ActionType[T], QueryType[Req, Res], Bus, Subscription, Context
- Constructors: New(), Global(), DefineAction[T](name), DefineQuery[Req,Res](name)
- Bus Methods:
  - Dispatch[T](typ ActionType[T], payload T, opts ...DispatchOption)
  - Subscribe[T](typ ActionType[T], h func(Context, T), opts ...SubOption) Subscription
  - SubscribeAny(h func(Context, Action[any]), opts ...SubOption) Subscription
  - Scope(name string) *Bus
  - ToSignal[T](typ ActionType[T], opts ...BridgeOption) *Signal[T]
  - ToStream[T](typ ActionType[T], opts ...BridgeOption) Stream[T]
  - HandleQuery[Req, Res](qt QueryType[Req, Res], h func(Context, Req) (Res, error), opts ...QueryOption) Subscription
  - Ask[Req, Res](qt QueryType[Req, Res], req Req, opts ...AskOption) Future[Res]
  - OnError(h func(Context, error, any))
- Options:
  - Dispatch: WithAsync, WithMeta, WithTrace, WithSource
  - Sub: Once, Filter, WithPriority, When, Buffered, Replay, DeliverAsync, Debounce, Throttle, SampleEvery, DistinctUntilChanged
  - Bridge: Initial, Replay, Filter, Map, DistinctUntilChanged, Buffer, DropPolicy
  - Query: Timeout, Concurrency, WithPriority
  - Ask: Timeout, Trace, Meta, Source
