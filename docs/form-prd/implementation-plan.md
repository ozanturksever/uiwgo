# Form Library Implementation Plan

**Author:** Ozan Turksever
**Date:** September 11, 2025
**Status:** Proposed

This document outlines the phased implementation plan for the `uiwgo` form library, as specified in the [Form Library PRD](forms-prd.md).

## Phase 1: Core Data Structures and State Management

**Goal:** Establish the foundational data structures and reactive state management for the form library.

*   **Task 1.1: Define Core Types**
    *   Create a new `form` package.
    *   Define the `form.FieldDef` struct with fields: `Name`, `Label`, `InitialValue`, `Validators`, `Widget`, `WidgetAttrs`.
    *   Define the `form.Validator` and `form.CrossFieldValidator` function types.
    *   Define the `form.Widget` function type.

*   **Task 1.2: Implement Form State**
    *   Create the `form.State` struct.
    *   It will hold the form schema, cross-field validators, and reactive signals for field values and errors.
    *   Use a `map[string]*reactivity.Signal[any]` for field values.
    *   Use a `map[string]*reactivity.Signal[error]` for field errors.
    *   Add a `*reactivity.Signal[error]` for global form errors.

*   **Task 1.3: Implement `NewFromSchema` Constructor**
    *   Create the `form.NewFromSchema` function.
    *   It should accept `[]form.FieldDef` and optional `[]form.CrossFieldValidator`.
    *   Inside the constructor, initialize the signals for each field based on the schema.
    *   Populate the signals with the `InitialValue` from each `FieldDef`.

## Phase 2: Rendering and Layout

**Goal:** Implement the components necessary to render form fields, supporting both default and custom layouts.

*   **Task 2.1: Create the `widgets` Sub-package**
    *   Create a `widgets` sub-package inside the `form` package.
    *   Implement a basic `TextInput` widget. It should:
        *   Accept `*form.State` and `*form.FieldDef`.
        *   Render an `<input>` element.
        *   Bind the input's value to the corresponding signal in `form.State`.
        *   Display the field's error signal.

*   **Task 2.2: Implement `form.Field` Component**
    *   Create the `form.Field(state, "fieldName")` component.
    *   It should look up the `FieldDef` from the state's schema.
    *   It will call the `Widget` function defined in the `FieldDef` to render the field.

*   **Task 2.3: Implement `form.Render` Component**
    *   Create the `form.Render(state)` component.
    *   It will iterate over the schema in `form.State`.
    *   For each `FieldDef`, it will render the field using its defined `Widget`, creating a simple vertical layout.

## Phase 3: Validation System

**Goal:** Implement a robust validation system for both per-field and cross-field validation rules.

*   **Task 3.1: Create the `validators` Sub-package**
    *   Create a `validators` sub-package inside the `form` package.
    *   Implement the `Required` validator.
    *   Implement the `MinLength(length int)` validator.
    *   Implement the `Email` validator using a regex.

*   **Task 3.2: Integrate Per-Field Validation**
    *   Add a `Validate()` method to `form.State`.
    *   This method will iterate through each field, run its validators, and update the field's error signal if validation fails.

*   **Task 3.3: Implement Cross-Field Validation**
    *   Extend the `Validate()` method to run cross-field validators after all per-field validations have passed.
    *   If a cross-field validator returns an error, update the global error signal.

*   **Task 3.4: Implement Error Display Components**
    *   Ensure widgets display per-field errors correctly.
    *   Create the `form.GlobalErrors(state)` component to display the global error signal.

## Phase 4: Form Submission

**Goal:** Implement the form submission process, tying together state, rendering, and validation.

*   **Task 4.1: Implement `form.For` Component**
    *   Create the `form.For(state, handleSubmit, ...children)` component.
    *   It will render an HTML `<form>` tag.
    *   It will attach an `onsubmit` event handler.

*   **Task 4.2: Implement Submission Logic**
    *   The `onsubmit` handler will:
        1.  Prevent the default form submission event.
        2.  Call the `form.State.Validate()` method.
        3.  If validation is successful, call the `handleSubmit` function.
        4.  If validation fails, the error signals will be updated, and the UI will reactively display the errors.

## Phase 5: Polish, Documentation, and Examples

**Goal:** Finalize the library with additional features, comprehensive documentation, and practical examples.

*   **Task 5.1: Add More Widgets and Validators**
    *   Implement `Textarea`, `PasswordInput`, `Select`, and `Checkbox` widgets.
    *   Implement a `Pattern(regex string)` validator.

*   **Task 5.2: Write Documentation**
    *   Write API documentation for all public types and functions.
    *   Update the `forms-usage.md` guide with detailed examples for all features.

*   **Task 5.3: Create a Demo Example**
    *   Create a new example in the `examples/` directory to showcase the form library.
    *   The example should include:
        *   A simple form using `form.Render`.
        *   A complex form with a custom layout using `form.Field`.
        *   A form demonstrating cross-field validation.
