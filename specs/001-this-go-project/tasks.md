# Tasks: React Compatibility Layer

**Input**: Design documents from `/specs/001-this-go-project/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/react-compat.md

## Phase 1: Core Bridge Implementation & Testing

- [ ] **T001**: Create the Go package structure and API stubs in `compat/react/react.go`.
- [ ] **T002**: Write a failing test for rendering a simple component in `compat/react/react_test.go`.
- [ ] **T003**: Implement the JavaScript bridge for rendering in a new file, e.g., `assets/js/bridge.js`. This involves creating the `window.renderComponent` function.

## Phase 2: State Synchronization (Go <-> React)

- [ ] **T004**: Write a failing test for Go-to-React state updates (props) in `compat/react/react_test.go`.
- [ ] **T005**: Implement the `updateComponent` bridge function in `compat/react/react.go` and `assets/js/bridge.js`.
- [ ] **T006**: Write a failing test for React-to-Go state updates (callbacks) in `compat/react/react_test.go`.
- [ ] **T007**: Implement callback handling in the bridge, ensuring `js.Func` props are correctly managed.

## Phase 3: Finalization and Integration

- [ ] **T008**: Write a failing test for component unmounting and resource cleanup in `compat/react/react_test.go`.
- [ ] **T009**: Implement the `unmountComponent` bridge function and associated Go cleanup logic.
- [ ] **T010**: Create the `react_demo` example in `examples/react_demo/main.go` to demonstrate the date picker integration.
- [ ] **T011**: [P] Create the necessary HTML and JS files for the `react_demo` example.
- [ ] **T012**: [P] Update project documentation with a guide on using the React compatibility feature.

## Dependencies
- **T001** must be completed before **T002**.
- **T002** must be completed and failing before **T003**.
- **T004** must be completed and failing before **T005**.
- **T006** must be completed and failing before **T007**.
- **T008** must be completed and failing before **T009**.
- **T010** and **T011** can be worked on in parallel after **T009** is complete.
- **T012** can be worked on in parallel with **T010** and **T011**.
