# Router Implementation Todo List (TDD Roadmap)

This file provides an execution-ready roadmap for implementing the SolidJS-inspired Router in Go+Wasm. It enforces Test-Driven Development (TDD) strictly — **all tests precede implementation**. Each todo represents a small, atomic task achievable in isolation.

When needed read full design doc at `@/spec/ROUTERDESIGN.txt`
Make sure you follow repo guidelines at `@AGENTS.md`
---

## Part 1: Core Reactive State Manager (LocationState)

- [ ] **Test:** Create failing unit test `TestLocationState_InitialGetReturnsZeroValue`.
- [ ] **Impl:** Implement `Location` struct and `NewLocationState()` constructor.

- [ ] **Test:** Write test `TestLocationState_SubscribeAddsSubscriber`.
- [ ] **Impl:** Implement `Subscribe(s Subscriber)` method.

- [ ] **Test:** Write test `TestLocationState_SetNotifiesSubscribers`.
- [ ] **Impl:** Implement `Set(newLocation Location)` with synchronous notifications.

- [ ] **Test:** Write test `TestLocationState_GetReturnsUpdatedValueAfterSet`.
- [ ] **Impl:** Implement thread-safe `Get()` using `RWMutex`.

✅ Criteria: All LocationState tests must pass with deterministic, synchronous updates.

---

## Part 2: Route Matching Engine

- [ ] **Test:** Write failing unit test `TestRouteMatcher_StaticPathMatches`.
- [ ] **Impl:** Implement `RouteDefinition` struct with `Path` and matcher closure.

- [ ] **Test:** Add test `TestRouteMatcher_DynamicSegmentCapturesParam`.
- [ ] **Impl:** Implement matcher function handling `:param`.

- [ ] **Test:** Add test `TestRouteMatcher_OptionalSegment`.
- [ ] **Impl:** Implement `:param?` optional segment logic.

- [ ] **Test:** Add test `TestRouteMatcher_WildcardSegment`.
- [ ] **Impl:** Implement greedy `*param` matching.

- [ ] **Test:** Add test `TestRouteMatcher_PrecedenceOrderMatters`.
- [ ] **Impl:** Implement first-match-wins traversal order logic.

- [ ] **Test:** Add test `TestRouteMatcher_MatchFiltersRegexpAndFunc`.
- [ ] **Impl:** Integrate `MatchFilters` validation.

✅ Criteria: All route matcher features covered by table-driven tests.

---

## Part 3: Router Public API Construction

- [ ] **Test:** Write test `TestRouterNewInitializesWithRoutesAndOutlet`.
- [ ] **Impl:** Implement `router.New(routes []RouteDefinition, outlet dom.Element)`.

- [ ] **Test:** Write test `TestRouteBuilderCreatesDefinition`.
- [ ] **Impl:** Implement `router.Route(path string, component func() gomponents.Node, ...)`.

- [ ] **Test:** Write test `TestRouterLocationReturnsCurrentState`.
- [ ] **Impl:** Implement `router.Location()` accessor.

- [ ] **Test:** Write test `TestRouterParamsReturnsMatchedParams`.
- [ ] **Impl:** Implement `router.Params()` API.

✅ Criteria: Public API accessible, corresponds to SolidJS Router mapping.

---

## Part 4: Navigation System

- [ ] **Test:** Write integration test `TestNavigation_LinkClickTriggersUpdate`.
- [ ] **Impl:** Implement `router.A(href string, children ...Node)` with post-render event binding.

- [ ] **Test:** Write unit test `TestNavigatePushUpdatesHistoryAndState`.
- [ ] **Impl:** Implement `router.Navigate(path string, opts NavigateOptions)`.

- [ ] **Test:** Write browser test `TestNavigation_BrowserHistoryPopstate`.
- [ ] **Impl:** Implement `popstate` event listener in `router.New()`.

✅ Criteria: Declarative links, programmatic navigation, and browser history all converge on `LocationState.Set()`.

---

## Part 5: View Rendering with gomponents

- [ ] **Test:** Write test `TestRouterInitialRenderMountsComponent`.
- [ ] **Impl:** Implement master render function with "destructive-and-replace" strategy.

- [ ] **Test:** Write test `TestRouterUpdatesViewOnRouteChange`.
- [ ] **Impl:** Implement reactive subscription between `LocationState` and render loop.

- [ ] **Test:** Write test `TestNestedRoutesRenderWithLayout`.
- [ ] **Impl:** Implement nested render composition passing child node to parent layout.

✅ Criteria: Rendering loop integrates with gomponents properly, produces correct DOM.

---

## Part 6: Browser Integration Tests (chromedp)

- [ ] **Test:** Write chromedp test `TestNavigation_LinkClickRendersAboutPage`.
- [ ] **Impl:** Ensure compiled wasm mounts A links and triggers DOM updates.

- [ ] **Test:** Write chromedp test `TestNavigation_BrowserForwardAndBack`.
- [ ] **Impl:** Ensure popstate updates DOM correctly.

- [ ] **Test:** Write chromedp test `TestRouterParamsAreVisibleInDOM`.
- [ ] **Impl:** Ensure dynamic parameters feed into component props and render in DOM.

✅ Criteria: Full router navigation cycle validated in live browser tests.

---

## Part 7: Logging & Guidelines Enforcement

- [ ] **Test:** Add linter-like test `TestNoFmtOrConsoleLoggingUsed`.
- [ ] **Impl:** Replace all debug prints with `logutil.Log/Logf`.

✅ Criteria: All logging goes through `logutil`, no fmt.Println or syscall/js console usage.

---

## Part 8: Documentation & Examples

- [ ] **Doc:** Create README with API reference and usage.
- [ ] **Example:** Build `examples/router_demo` app with multiple routes.
- [ ] **Test:** chromedp validate example app (`make test-example router_demo`).

✅ Criteria: Example works end-to-end, included in full test suite.

---

# Summary

This roadmap enforces a strict **TDD-first, incremental implementation** of the router:
1. **Write failing unit/integration/browser tests first.**
2. **Implement minimal code to satisfy tests.**
3. **Validate with both Go unit tests and chromedp browser automation.**
4. **Confirm logutil-only logging and DOM safety with honnef.co/go/js/dom/v2.**
