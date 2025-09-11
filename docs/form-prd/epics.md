# Form Library Epics

**Author:** Ozan Turksever
**Date:** September 11, 2025
**Status:** Proposed

This document defines the high-level epics for the `uiwgo` form library project, based on the [Form Library PRD](forms-prd.md).

---

### Epic 1: Form Schema & State Foundation

**Goal:** To establish the core data structures and reactive state management that will underpin the entire form library. This epic focuses on the "brains" of the operation, without worrying about the UI.

**User Stories / Features:**

*   As a developer, I want to define my form fields and validation rules in a single Go data structure (`form.FieldDef`).
*   The system needs a central state manager (`form.State`) that holds the value and error state for every field.
*   I want to create a form's state manager from a schema definition (`form.NewFromSchema`).
*   The state of each field (its value and error) must be a reactive signal, so that changes can be automatically tracked.

---

### Epic 2: Form Rendering & Layout Engine

**Goal:** To provide developers with flexible tools to render forms, supporting everything from rapid prototyping with default layouts to complex, custom-designed UIs.

**User Stories / Features:**

*   As a developer, I want to render a complete form with a standard vertical layout using a single component (`form.Render`).
*   As a developer, I want to arrange fields individually within a custom layout, like a multi-column grid (`form.Field`).
*   The library must provide a set of standard input widgets (e.g., `TextInput`, `Textarea`, `PasswordInput`).
*   Widgets must be able to bind to the form's reactive state to display values and receive user input.

---

### Epic 3: Comprehensive Validation & Error Handling

**Goal:** To implement a robust and declarative validation system that gives developers fine-grained control and provides clear, immediate feedback to users.

**User Stories / Features:**

*   As a developer, I want to define per-field validation rules (e.g., `Required`, `Email`, `MinLength`) directly in my form schema.
*   As a developer, I want to implement cross-field validation logic, such as password confirmation.
*   When validation fails, form errors must automatically appear in the UI next to the relevant field.
*   The system should support a dedicated area for displaying global form errors (errors that don't belong to a single field).

---

### Epic 4: Form Submission Workflow

**Goal:** To create a seamless and secure form submission process that integrates validation and prevents invalid data from being submitted.

**User Stories / Features:**

*   As a developer, I want a simple way to wrap my form components and handle submission (`form.For`).
*   The system must automatically trigger all validations when the user tries to submit the form.
*   The submission handler function should only be called if all validation rules pass.
*   If validation fails on submission, the UI should automatically update to show all relevant errors.
