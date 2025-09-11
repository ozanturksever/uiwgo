# PRD: Declarative & Reactive Form Helpers for uiwgo

**Author:** Ozan Turksever
**Owner:** Ozan Turksever, Logsign B.V.
**Date:** September 11, 2025
**Status:** Proposed

## 1. Introduction & Executive Summary

This document outlines the requirements for a new form package within the uiwgo framework. The goal is to provide a declarative, reactive, and schema-driven system for building, validating, and rendering HTML forms. This feature will be deeply integrated with the existing reactivity and gomponents patterns, offering a powerful, Django-like experience for handling forms in Go WASM applications. By allowing developers to define forms as data structures, we will significantly reduce boilerplate, improve code clarity, and enable dynamic form generation from sources like Protobuf or JSON schemas.

## 2. Problem Statement

Building forms in web applications involves several repetitive and error-prone tasks: managing field state, handling user input, validating data, and displaying error messages. Currently, uiwgo developers must handle this logic manually using individual signals and components, leading to verbose code that is difficult to maintain and reuse. There is no standardized way to handle complex validation logic or to create sophisticated visual layouts (e.g., multi-column) without significant manual effort. This increases development time and the likelihood of bugs.

## 3. Goals & Objectives

*   **Goal**: To make building robust, secure, and user-friendly forms in uiwgo simple and elegant.
*   **Objectives**:
    *   **Reduce Boilerplate**: Abstract away the repetitive logic of state management and event handling for forms.
    *   **Improve Declarativeness**: Allow developers to define an entire form's structure and rules as a single, clear data structure (a schema).
    *   **Enable Custom Layouts**: Provide tools for developers to easily create complex form layouts, such as multi-column grids, while reusing the schema definition.
    *   **Enable Dynamic Generation**: The schema-driven approach must be flexible enough to allow forms to be generated from external sources.
    *   **Provide Robust Validation**: Offer built-in support for both per-field and cross-field validation rules.
    *   **Integrate Seamlessly**: Ensure the new library feels like a natural extension of uiwgo's existing reactivity and component model.

## 4. Target Audience

This feature is for all Go developers using the uiwgo framework to build web frontends. It will be especially valuable for those creating applications with data-entry requirements, such as user registration, settings pages, or complex business applications.

## 5. User Stories

*   As a developer, I want to define my form fields and validation rules in a single Go data structure so that my form logic is centralized and easy to read.
*   As a developer, I want to render a complete form with a standard vertical layout using a single component so that I can quickly build simple forms.
*   As a developer, I want to arrange 'First Name' and 'Last Name' side-by-side in a two-column grid so that my user registration form is more compact and intuitive.
*   As a developer, I want to implement password confirmation logic by defining a single cross-field validation rule so that I can ensure data integrity without complex manual checks.
*   As a developer, I want form errors to automatically appear in the UI when validation fails so that I can provide immediate feedback to the user with minimal effort.

## 6. Functional Requirements

### 6.1. Schema-Driven Form Definition

The system will provide a `form.FieldDef` struct to define a single form field. This struct will include:

*   `Name` (string): The programmatic name of the field (e.g., "user_email").
*   `Label` (string): The user-visible label.
*   `InitialValue` (any): The default value for the field.
*   `Validators` (`[]form.Validator`): A slice of per-field validation functions.
*   `Widget` (`form.Widget`): A function responsible for rendering the field's UI.
*   `WidgetAttrs` (`gomponents.Node`): Optional attributes to pass to the widget (e.g., `html.Placeholder`).

### 6.2. Reactive State Management

*   A `form.NewFromSchema` constructor will accept a schema (`[]form.FieldDef`) and an optional list of `CrossFieldValidators`.
*   It will return a `*form.State` object that manages the entire form's state using `reactivity.Signal`.
*   The `form.State` will expose signals for each field's value and errors, as well as a signal for global form errors.

### 6.3. Widget-Based Rendering

*   A `widgets` sub-package will provide standard widget functions (`TextInput`, `Textarea`, `PasswordInput`, `Select`, `Checkbox`).
*   A widget is a function that takes the `form.State` and a `*form.FieldDef` and returns a `gomponents.Node`.
*   The library will provide two primary rendering components:
    *   `form.Render(state)`: A default layout component that automatically iterates through the form's schema and renders all fields sequentially in a single column.
    *   `form.Field(state, "fieldName")`: A component that renders a single field by its name. This is the primary tool for building custom form layouts.

### 6.4. Validation

*   **Per-Field Validation**: A `validators` sub-package will provide common validator functions (`Required`, `MinLength`, `Email`, `Pattern`, etc.). These are attached to individual fields in the schema.
*   **Cross-Field Validation**: The system will support `CrossFieldValidator` functions that receive a map of all form values and return a single error. These are passed during form creation.
*   **Error Display**:
    *   `widgets` will automatically include a mechanism to show per-field errors.
    *   A `form.GlobalErrors` component will be provided to display errors from cross-field validation.

### 6.5. Form Submission

*   A `form.For(state, handleSubmit, ...children)` component will wrap the form.
*   On submission, it will trigger all validations (per-field first, then cross-field).
*   If validation succeeds, it will call the `handleSubmit` function.
*   If validation fails, it will update the relevant error signals, causing the UI to display the errors automatically.

### 6.6. Layout Customization

*   **Standard Layout (`form.Render`)**: The `form.Render(state)` component provides a standard, single-column vertical layout. It is designed for rapid development and simple forms where fields are rendered sequentially.
*   **Custom Layouts (`form.Field`)**: For complete control over the form's appearance, the system will provide the `form.Field(state, "fieldName")` component. This allows developers to place individual fields within any custom `gomponents` layout structure (e.g., CSS Grid, Flexbox), enabling multi-column designs and visually grouped fields.
*   **Example of a Custom Two-Column Layout**:
    ```go
    html.Div(
       html.Class("grid grid-cols-2 gap-4"), // Tailwind CSS for a 2-col grid

       // Column 1
       html.Div(
           form.Field(formState, "first_name"),
           form.Field(formState, "email"),
       ),

       // Column 2
       html.Div(
           form.Field(formState, "last_name"),
           form.Field(formState, "phone_number"),
       ),
    )
    ```

## 7. Non-Functional Requirements

*   **Performance**: The reactivity system should be efficient, ensuring no unnecessary re-renders. DOM updates should be minimal.
*   **Usability**: The API should be intuitive for developers familiar with Go and `gomponents`.
*   **Documentation**: The feature must be accompanied by comprehensive documentation, including a conceptual overview, API reference, and practical examples for both simple and custom layouts.

## 8. Out of Scope (for v1.0)

*   **Dynamic Schema Updates**: Modifying the form's schema after it has been created.
*   **File Uploads**: A dedicated `FileField` widget and state management for file objects.
*   **Form Sets**: Managing a collection of identical forms on one page.
*   **Automatic Schema Generation**: Tools to generate `FieldDef` schemas directly from Protobuf or other sources (though the design explicitly enables this to be built separately).

## 9. Success Metrics

*   **Adoption**: The new form package is used in new feature development within Logsign's products.
*   **Code Reduction**: A measurable decrease in the lines of code required to create new forms compared to the old manual method.
*   **Developer Feedback**: Positive feedback from the development team regarding ease of use, clarity, and layout flexibility.
