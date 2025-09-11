# uiwgo Form Helpers: Usage Guide & Best Practices

This document provides practical examples and guidance for using the form package in your uiwgo applications. It is intended to be a companion to the official Product Requirements Document (PRD).

## 1. Introduction

The uiwgo form library is designed to be declarative, reactive, and flexible. The core philosophy is to define your form's structure and logic as data (a schema) and then use components to render that structure. This guide will walk you through the most common patterns you'll encounter and highlight some potential pitfalls to avoid.

## 2. Common Use Cases (The "Do's")

These examples demonstrate the idiomatic way to build forms with uiwgo.

### Use Case 1: The Simple, Stacked Form

For basic forms like a contact or feedback form, the default layout renderer is the fastest way to get started.

**Scenario**: A simple contact form with Name, Email, and Message fields.

```go
// 1. Define the schema
var contactSchema = []form.FieldDef{
   { Name: "name", Label: "Name", Widget: widgets.TextInput, Validators: []form.Validator{validators.Required()} },
   { Name: "email", Label: "Email", Widget: widgets.TextInput, Validators: []form.Validator{validators.Required(), validators.Email()} },
   { Name: "message", Label: "Message", Widget: widgets.Textarea, Validators: []form.Validator{validators.Required()} },
}

// 2. Create state and render the form
func ContactForm() gomponents.Node {
   formState := form.NewFromSchema(contactSchema)

   return html.Form(
       form.For(formState, handleContactSubmit,
           // form.Render does all the layout work for you!
           form.Render(formState),
           html.Button(gomponents.Text("Send Message")),
       ),
   )
}
```

**Key Takeaway**: For standard, single-column layouts, `form.Render()` is all you need.

### Use Case 2: The Multi-Column Custom Layout

When the design requires more complex layouts, use `form.Field()` to place each field individually within your own grid or flexbox structure.

**Scenario**: A user profile page where "First Name" and "Last Name" appear side-by-side.

```go
var profileSchema = []form.FieldDef{
   { Name: "first_name", Label: "First Name", Widget: widgets.TextInput },
   { Name: "last_name", Label: "Last Name", Widget: widgets.TextInput },
   { Name: "bio", Label: "Biography", Widget: widgets.Textarea },
}

func ProfileForm() gomponents.Node {
   formState := form.NewFromSchema(profileSchema)

   return html.Form(
       form.For(formState, handleProfileUpdate,
           // Custom layout using Tailwind CSS classes
           html.Div(
               html.Class("grid grid-cols-2 gap-x-4"),
               
               // Use form.Field() to render a single field by name
               form.Field(formState, "first_name"), // Column 1
               form.Field(formState, "last_name"),  // Column 2
           ),
           
           // A full-width field
           form.Field(formState, "bio"),

           html.Button(gomponents.Text("Update Profile")),
       ),
   )
}
```

**Key Takeaway**: `form.Field(state, "name")` is the essential tool for full layout control.

### Use Case 3: Form with Cross-Field Validation

For validation that depends on multiple fields, pass `CrossFieldValidator` functions when creating the form state.

**Scenario**: A user registration form that requires password confirmation.

```go
// 1. Define the validator
func passwordsMustMatch() form.CrossFieldValidator {
   return func(values map[string]any) error {
       if values["password"] != values["confirm_password"] {
           return errors.New("Passwords do not match.")
       }
       return nil
   }
}

// 2. Define the schema
var signupSchema = []form.FieldDef{
   { Name: "email", Label: "Email", Widget: widgets.TextInput, ... },
   { Name: "password", Label: "Password", Widget: widgets.PasswordInput, ... },
   { Name: "confirm_password", Label: "Confirm Password", Widget: widgets.PasswordInput, ... },
}

// 3. Create state with the cross-field validator
func SignupForm() gomponents.Node {
   formState := form.NewFromSchema(
       signupSchema,
       passwordsMustMatch(), // Add the validator here
   )

   return html.Form(
       form.For(formState, handleSignup,
           form.Render(formState),
           // GlobalErrors will display the "Passwords do not match" message
           form.GlobalErrors(formState),
           html.Button(gomponents.Text("Create Account")),
       ),
   )
}
```

**Key Takeaway**: Complex form-wide rules are handled at the state creation level, keeping the UI and schema clean.

## 3. Common Pitfalls & Anti-Patterns (The "Don'ts")

Avoiding these common mistakes will help you write cleaner, more maintainable code.

### Anti-Pattern 1: Bypassing the Form State

**Wrong**: Trying to get a value directly from the DOM within your submission handler. This breaks the reactive model.

```go
// WRONG
func handleSubmit(fs *form.State) {
   // This is brittle and doesn't use the validated state.
   email := js.Global().Get("document").Call("getElementById", "email-input-id").Get("value").String()
   fmt.Println("Email from DOM:", email)
}
```

**Correct**: Always get data from the validated form state.

```go
// CORRECT
func handleSubmit(fs *form.State) {
   // This uses the clean, validated data from the form's reactive state.
   values := fs.Values()
   email, _ := values["email"].(string)
   fmt.Println("Email from state:", email)
}
```

### Anti-Pattern 2: Mixing `form.Render` and `form.Field`

**Wrong**: Using `form.Render()` when you also need to place some fields manually. The `Render` component is designed to render *all* fields from the schema, which will lead to duplication.

```go
// WRONG - This will render "first_name" and "last_name" twice!
html.Div(
   html.Class("grid grid-cols-2"),
   form.Field(formState, "first_name"),
   form.Field(formState, "last_name"),
)
form.Render(formState) // This will render ALL fields again
```

**Correct**: Choose one strategy. If you need any custom layout, do not use `form.Render`. Build the entire layout with `form.Field`.

```go
// CORRECT
html.Div(
   html.Class("grid grid-cols-2"),
   form.Field(formState, "first_name"),
   form.Field(formState, "last_name"),
)
// Render the rest of the fields individually as well
form.Field(formState, "email") 
```

### Anti-Pattern 3: Placing Validation Logic in the Submit Handler

**Wrong**: Writing validation checks inside your `handleSubmit` function. This makes the logic non-reusable and mixes concerns.

```go
// WRONG
func handleSubmit(fs *form.State) {
   values := fs.Values()
   email, _ := values["email"].(string)
   
   // Manual validation logic belongs in the schema, not here.
   if !strings.Contains(email, "@") {
       // How do you even show an error to the user from here?
       log.Println("Invalid email!")
       return
   }
   
   // ... proceed with submission
}
```

**Correct**: Define all validation rules declaratively in the schema. The framework handles the rest.

```go
// CORRECT
var mySchema = []form.FieldDef{
   // The validation rule is part of the definition.
   { Name: "email", ..., Validators: []form.Validator{validators.Required(), validators.Email()} },
}

func handleSubmit(fs *form.State) {
   // By the time this handler is called, you know the data is valid.
   // No need to check for '@' here.
   log.Println("Submitting valid data:", fs.Values())
}
```
