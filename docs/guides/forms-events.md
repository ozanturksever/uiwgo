# Forms & Events Guide

This guide covers form handling, event management, and user interaction patterns in UIwGo. Learn to build interactive forms with validation, handle various events, and create responsive user interfaces.

## Table of Contents

- [Overview](#overview)
- [Form Basics](#form-basics)
- [Input Handling](#input-handling)
- [Form Validation](#form-validation)
- [Event Management](#event-management)
- [Advanced Form Patterns](#advanced-form-patterns)
- [File Uploads](#file-uploads)
- [Real-time Validation](#real-time-validation)
- [Form State Management](#form-state-management)
- [Best Practices](#best-practices)

## Overview

UIwGo handles forms and events through reactive bindings and signal-based state management. Instead of traditional form libraries, you use:

- **Data attributes** for event binding (`data-click`, `data-input`, `data-change`)
- **Reactive signals** for form state
- **Memos** for computed validation and formatting
- **Effects** for side effects and API calls

### Mental Model

```go
// Traditional form handling
// <form onsubmit="handleSubmit(event)">
//   <input name="email" onchange="validateEmail(event)">
// </form>

// UIwGo approach
email := reactivity.NewSignal("")
emailError := reactivity.NewMemo(func() string {
    return validateEmail(email.Get())
})

// Reactive form rendering
formHTML := reactivity.NewMemo(func() string {
    return fmt.Sprintf(`
        <form data-submit="handleSubmit">
            <input data-input="email" value="%s">
            <span class="error">%s</span>
        </form>
    `, email.Get(), emailError.Get())
})
```

## Form Basics

### Simple Contact Form

A basic form with text inputs and submission:

```go
type ContactForm struct {
    name    *reactivity.Signal[string]
    email   *reactivity.Signal[string]
    message *reactivity.Signal[string]
    
    isSubmitting *reactivity.Signal[bool]
    submitResult *reactivity.Signal[string]
    
    formHTML *reactivity.Memo[string]
}

func NewContactForm() *ContactForm {
    cf := &ContactForm{
        name:         reactivity.NewSignal(""),
        email:        reactivity.NewSignal(""),
        message:      reactivity.NewSignal(""),
        isSubmitting: reactivity.NewSignal(false),
        submitResult: reactivity.NewSignal(""),
    }
    
    cf.formHTML = reactivity.NewMemo(func() string {
        submitText := "Send Message"
        if cf.isSubmitting.Get() {
            submitText = "Sending..."
        }
        
        result := cf.submitResult.Get()
        resultHTML := ""
        if result != "" {
            resultHTML = fmt.Sprintf(`<div class="result">%s</div>`, result)
        }
        
        return fmt.Sprintf(`
            <form class="contact-form" data-submit="handleSubmit">
                <div class="field">
                    <label for="name">Name:</label>
                    <input type="text" id="name" data-input="name" 
                           value="%s" required>
                </div>
                
                <div class="field">
                    <label for="email">Email:</label>
                    <input type="email" id="email" data-input="email" 
                           value="%s" required>
                </div>
                
                <div class="field">
                    <label for="message">Message:</label>
                    <textarea id="message" data-input="message" 
                              rows="4" required>%s</textarea>
                </div>
                
                <button type="submit" %s>%s</button>
                
                %s
            </form>
        `, 
            cf.name.Get(),
            cf.email.Get(), 
            cf.message.Get(),
            map[bool]string{true: "disabled", false: ""}[cf.isSubmitting.Get()],
            submitText,
            resultHTML)
    })
    
    return cf
}

func (cf *ContactForm) Render() g.Node {
    return h.Div(g.Class("contact-container"),
        h.H2(g.Text("Contact Us")),
        h.Div(
            g.Attr("data-html", "form"),
            g.Text(cf.formHTML.Get()),
        ),
    )
}

func (cf *ContactForm) Attach() {
    comps.BindInput("name", cf.name)
    comps.BindInput("email", cf.email)
    comps.BindInput("message", cf.message)
    comps.BindSubmit("handleSubmit", cf.handleSubmit)
    comps.BindHTML("form", cf.formHTML)
}

func (cf *ContactForm) handleSubmit() {
    // Prevent submission if already submitting
    if cf.isSubmitting.Get() {
        return
    }
    
    // Basic validation
    if cf.name.Get() == "" || cf.email.Get() == "" || cf.message.Get() == "" {
        cf.submitResult.Set("Please fill in all fields.")
        return
    }
    
    cf.isSubmitting.Set(true)
    cf.submitResult.Set("")
    
    // Simulate API call
    go func() {
        time.Sleep(2 * time.Second)
        
        // Simulate success/failure
        if strings.Contains(cf.email.Get(), "@") {
            cf.submitResult.Set("Message sent successfully!")
            
            // Clear form on success
            cf.name.Set("")
            cf.email.Set("")
            cf.message.Set("")
        } else {
            cf.submitResult.Set("Failed to send message. Please check your email.")
        }
        
        cf.isSubmitting.Set(false)
    }()
}
```

## Input Handling

### Different Input Types

Handle various HTML input types:

```go
type InputDemo struct {
    textValue     *reactivity.Signal[string]
    numberValue   *reactivity.Signal[int]
    dateValue     *reactivity.Signal[string]
    timeValue     *reactivity.Signal[string]
    colorValue    *reactivity.Signal[string]
    rangeValue    *reactivity.Signal[int]
    checkboxValue *reactivity.Signal[bool]
    radioValue    *reactivity.Signal[string]
    selectValue   *reactivity.Signal[string]
    
    summary *reactivity.Memo[string]
}

func NewInputDemo() *InputDemo {
    id := &InputDemo{
        textValue:     reactivity.NewSignal("Hello"),
        numberValue:   reactivity.NewSignal(42),
        dateValue:     reactivity.NewSignal("2024-01-01"),
        timeValue:     reactivity.NewSignal("12:00"),
        colorValue:    reactivity.NewSignal("#ff0000"),
        rangeValue:    reactivity.NewSignal(50),
        checkboxValue: reactivity.NewSignal(true),
        radioValue:    reactivity.NewSignal("option1"),
        selectValue:   reactivity.NewSignal("apple"),
    }
    
    id.summary = reactivity.NewMemo(func() string {
        return fmt.Sprintf(`
            <div class="summary">
                <h3>Current Values:</h3>
                <ul>
                    <li>Text: %s</li>
                    <li>Number: %d</li>
                    <li>Date: %s</li>
                    <li>Time: %s</li>
                    <li>Color: %s</li>
                    <li>Range: %d</li>
                    <li>Checkbox: %t</li>
                    <li>Radio: %s</li>
                    <li>Select: %s</li>
                </ul>
            </div>
        `,
            id.textValue.Get(),
            id.numberValue.Get(),
            id.dateValue.Get(),
            id.timeValue.Get(),
            id.colorValue.Get(),
            id.rangeValue.Get(),
            id.checkboxValue.Get(),
            id.radioValue.Get(),
            id.selectValue.Get())
    })
    
    return id
}

func (id *InputDemo) Render() string {
    return fmt.Sprintf(`
        <div class="input-demo">
            <h2>Input Types Demo</h2>
            
            <form class="demo-form">
                <div class="field">
                    <label>Text Input:</label>
                    <input type="text" data-input="text" value="%s">
                </div>
                
                <div class="field">
                    <label>Number Input:</label>
                    <input type="number" data-input="number" value="%d" min="0" max="100">
                </div>
                
                <div class="field">
                    <label>Date Input:</label>
                    <input type="date" data-input="date" value="%s">
                </div>
                
                <div class="field">
                    <label>Time Input:</label>
                    <input type="time" data-input="time" value="%s">
                </div>
                
                <div class="field">
                    <label>Color Input:</label>
                    <input type="color" data-input="color" value="%s">
                </div>
                
                <div class="field">
                    <label>Range Input (0-100):</label>
                    <input type="range" data-input="range" value="%d" min="0" max="100">
                    <span>%d</span>
                </div>
                
                <div class="field">
                    <label>
                        <input type="checkbox" data-change="checkbox" %s>
                        Checkbox Option
                    </label>
                </div>
                
                <div class="field">
                    <label>Radio Options:</label>
                    <label><input type="radio" name="radio" value="option1" data-change="radio" %s> Option 1</label>
                    <label><input type="radio" name="radio" value="option2" data-change="radio" %s> Option 2</label>
                    <label><input type="radio" name="radio" value="option3" data-change="radio" %s> Option 3</label>
                </div>
                
                <div class="field">
                    <label>Select Dropdown:</label>
                    <select data-change="select">
                        <option value="apple" %s>Apple</option>
                        <option value="banana" %s>Banana</option>
                        <option value="orange" %s>Orange</option>
                    </select>
                </div>
            </form>
            
            <div data-html="summary">%s</div>
        </div>
    `,
        id.textValue.Get(),
        id.numberValue.Get(),
        id.dateValue.Get(),
        id.timeValue.Get(),
        id.colorValue.Get(),
        id.rangeValue.Get(),
        id.rangeValue.Get(),
        map[bool]string{true: "checked", false: ""}[id.checkboxValue.Get()],
        map[bool]string{true: "checked", false: ""}[id.radioValue.Get() == "option1"],
        map[bool]string{true: "checked", false: ""}[id.radioValue.Get() == "option2"],
        map[bool]string{true: "checked", false: ""}[id.radioValue.Get() == "option3"],
        map[bool]string{true: "selected", false: ""}[id.selectValue.Get() == "apple"],
        map[bool]string{true: "selected", false: ""}[id.selectValue.Get() == "banana"],
        map[bool]string{true: "selected", false: ""}[id.selectValue.Get() == "orange"],
        id.summary.Get())
}

func (id *InputDemo) Attach() {
    id.BindInput("text", id.textValue)
    
    id.BindInput("number", func(value string) {
        if num, err := strconv.Atoi(value); err == nil {
            id.numberValue.Set(num)
        }
    })
    
    id.BindInput("date", id.dateValue)
    id.BindInput("time", id.timeValue)
    id.BindInput("color", id.colorValue)
    
    id.BindInput("range", func(value string) {
        if num, err := strconv.Atoi(value); err == nil {
            id.rangeValue.Set(num)
        }
    })
    
    id.BindChange("checkbox", id.checkboxValue)
    id.BindChange("radio", id.radioValue)
    id.BindChange("select", id.selectValue)
    
    id.BindHTML("summary", id.summary)
}
```

## Form Validation

### Real-time Validation

Implement validation that updates as users type:

```go
type ValidationRule func(string) string

type ValidatedField struct {
    value    *reactivity.Signal[string]
    error    *reactivity.Memo[string]
    isValid  *reactivity.Memo[bool]
    rules    []ValidationRule
    touched  *reactivity.Signal[bool]
}

func NewValidatedField(initial string, rules ...ValidationRule) *ValidatedField {
    vf := &ValidatedField{
        value:   reactivity.NewSignal(initial),
        rules:   rules,
        touched: reactivity.NewSignal(false),
    }
    
    vf.error = reactivity.NewMemo(func() string {
        // Only show errors after field has been touched
        if !vf.touched.Get() {
            return ""
        }
        
        value := vf.value.Get()
        for _, rule := range vf.rules {
            if err := rule(value); err != "" {
                return err
            }
        }
        return ""
    })
    
    vf.isValid = reactivity.NewMemo(func() bool {
        value := vf.value.Get()
        for _, rule := range vf.rules {
            if err := rule(value); err != "" {
                return false
            }
        }
        return true
    })
    
    return vf
}

func (vf *ValidatedField) SetTouched() {
    vf.touched.Set(true)
}

// Validation rules
func Required(value string) string {
    if strings.TrimSpace(value) == "" {
        return "This field is required"
    }
    return ""
}

func MinLength(min int) ValidationRule {
    return func(value string) string {
        if len(value) < min {
            return fmt.Sprintf("Must be at least %d characters", min)
        }
        return ""
    }
}

func MaxLength(max int) ValidationRule {
    return func(value string) string {
        if len(value) > max {
            return fmt.Sprintf("Must be no more than %d characters", max)
        }
        return ""
    }
}

func Email(value string) string {
    if value == "" {
        return ""
    }
    
    // Simple email validation
    if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
        return "Must be a valid email address"
    }
    
    return ""
}

func Pattern(pattern string, message string) ValidationRule {
    regex := regexp.MustCompile(pattern)
    return func(value string) string {
        if value != "" && !regex.MatchString(value) {
            return message
        }
        return ""
    }
}

// Registration form with validation
type RegistrationForm struct {
    username        *ValidatedField
    email          *ValidatedField
    password       *ValidatedField
    confirmPassword *ValidatedField
    
    isSubmitting *reactivity.Signal[bool]
    submitResult *reactivity.Signal[string]
    
    isFormValid *reactivity.Memo[bool]
    formHTML    *reactivity.Memo[string]
}

func NewRegistrationForm() *RegistrationForm {
    rf := &RegistrationForm{
        username: NewValidatedField("", 
            Required, 
            MinLength(3), 
            MaxLength(20),
            Pattern(`^[a-zA-Z0-9_]+$`, "Username can only contain letters, numbers, and underscores")),
        email: NewValidatedField("", Required, Email),
        password: NewValidatedField("", Required, MinLength(8)),
        isSubmitting: reactivity.NewSignal(false),
        submitResult: reactivity.NewSignal(""),
    }
    
    // Confirm password validation depends on password field
    rf.confirmPassword = NewValidatedField("", Required, func(value string) string {
        if value != rf.password.value.Get() {
            return "Passwords do not match"
        }
        return ""
    })
    
    rf.isFormValid = reactivity.NewMemo(func() bool {
        return rf.username.isValid.Get() &&
               rf.email.isValid.Get() &&
               rf.password.isValid.Get() &&
               rf.confirmPassword.isValid.Get()
    })
    
    rf.formHTML = reactivity.NewMemo(func() string {
        submitText := "Register"
        if rf.isSubmitting.Get() {
            submitText = "Registering..."
        }
        
        result := rf.submitResult.Get()
        resultHTML := ""
        if result != "" {
            resultHTML = fmt.Sprintf(`<div class="result">%s</div>`, result)
        }
        
        return fmt.Sprintf(`
            <form class="registration-form" data-submit="handleSubmit">
                %s
                %s
                %s
                %s
                
                <button type="submit" %s %s>%s</button>
                
                %s
            </form>
        `,
            rf.renderField("username", "Username", "text", rf.username),
            rf.renderField("email", "Email", "email", rf.email),
            rf.renderField("password", "Password", "password", rf.password),
            rf.renderField("confirmPassword", "Confirm Password", "password", rf.confirmPassword),
            map[bool]string{true: "disabled", false: ""}[rf.isSubmitting.Get() || !rf.isFormValid.Get()],
            map[bool]string{true: "class=\"btn-disabled\"", false: "class=\"btn-primary\""}[!rf.isFormValid.Get()],
            submitText,
            resultHTML)
    })
    
    return rf
}

func (rf *RegistrationForm) renderField(name, label, inputType string, field *ValidatedField) string {
    errorClass := ""
    errorHTML := ""
    
    if field.error.Get() != "" {
        errorClass = "error"
        errorHTML = fmt.Sprintf(`<span class="error-message">%s</span>`, field.error.Get())
    }
    
    return fmt.Sprintf(`
        <div class="field %s">
            <label for="%s">%s:</label>
            <input type="%s" id="%s" data-input="%s" data-blur="%s-blur" 
                   value="%s" autocomplete="off">
            %s
        </div>
    `, errorClass, name, label, inputType, name, name, name, field.value.Get(), errorHTML)
}

func (rf *RegistrationForm) Render() string {
    return fmt.Sprintf(`
        <div class="registration-container">
            <h2>Create Account</h2>
            <div data-html="form">%s</div>
        </div>
    `, rf.formHTML.Get())
}

func (rf *RegistrationForm) Attach() {
    rf.BindInput("username", rf.username.value)
    rf.BindInput("email", rf.email.value)
    rf.BindInput("password", rf.password.value)
    rf.BindInput("confirmPassword", rf.confirmPassword.value)
    
    // Mark fields as touched on blur
    rf.BindBlur("username-blur", rf.username.SetTouched)
    rf.BindBlur("email-blur", rf.email.SetTouched)
    rf.BindBlur("password-blur", rf.password.SetTouched)
    rf.BindBlur("confirmPassword-blur", rf.confirmPassword.SetTouched)
    
    rf.BindSubmit("handleSubmit", rf.handleSubmit)
    rf.BindHTML("form", rf.formHTML)
}

func (rf *RegistrationForm) handleSubmit() {
    // Mark all fields as touched to show any validation errors
    rf.username.SetTouched()
    rf.email.SetTouched()
    rf.password.SetTouched()
    rf.confirmPassword.SetTouched()
    
    if !rf.isFormValid.Get() {
        rf.submitResult.Set("Please fix the errors above.")
        return
    }
    
    if rf.isSubmitting.Get() {
        return
    }
    
    rf.isSubmitting.Set(true)
    rf.submitResult.Set("")
    
    // Simulate API call
    go func() {
        time.Sleep(2 * time.Second)
        
        // Simulate success/failure
        username := rf.username.value.Get()
        if username == "admin" {
            rf.submitResult.Set("Username 'admin' is not available.")
        } else {
            rf.submitResult.Set("Account created successfully!")
            
            // Clear form on success
            rf.username.value.Set("")
            rf.email.value.Set("")
            rf.password.value.Set("")
            rf.confirmPassword.value.Set("")
            
            // Reset touched state
            rf.username.touched.Set(false)
            rf.email.touched.Set(false)
            rf.password.touched.Set(false)
            rf.confirmPassword.touched.Set(false)
        }
        
        rf.isSubmitting.Set(false)
    }()
}
```

## Event Management

### Custom Event Handlers

Handle various DOM events beyond basic clicks and inputs:

```go
type EventDemo struct {
    mousePosition *reactivity.Signal[string]
    keyPressed    *reactivity.Signal[string]
    scrollPosition *reactivity.Signal[int]
    windowSize    *reactivity.Signal[string]
    
    eventLog *reactivity.Signal[[]string]
    
    displayHTML *reactivity.Memo[string]
}

func NewEventDemo() *EventDemo {
    ed := &EventDemo{
        mousePosition:  reactivity.NewSignal("0, 0"),
        keyPressed:     reactivity.NewSignal("None"),
        scrollPosition: reactivity.NewSignal(0),
        windowSize:     reactivity.NewSignal("Unknown"),
        eventLog:       reactivity.NewSignal([]string{}),
    }
    
    ed.displayHTML = reactivity.NewMemo(func() string {
        log := ed.eventLog.Get()
        logHTML := ""
        
        if len(log) > 0 {
            var logItems strings.Builder
            // Show last 10 events
            start := 0
            if len(log) > 10 {
                start = len(log) - 10
            }
            
            for i := start; i < len(log); i++ {
                logItems.WriteString(fmt.Sprintf(`<li>%s</li>`, log[i]))
            }
            
            logHTML = fmt.Sprintf(`
                <div class="event-log">
                    <h3>Recent Events:</h3>
                    <ul>%s</ul>
                </div>
            `, logItems.String())
        }
        
        return fmt.Sprintf(`
            <div class="event-display">
                <div class="status-grid">
                    <div class="status-item">
                        <strong>Mouse Position:</strong> %s
                    </div>
                    <div class="status-item">
                        <strong>Last Key:</strong> %s
                    </div>
                    <div class="status-item">
                        <strong>Scroll Position:</strong> %dpx
                    </div>
                    <div class="status-item">
                        <strong>Window Size:</strong> %s
                    </div>
                </div>
                
                %s
            </div>
        `,
            ed.mousePosition.Get(),
            ed.keyPressed.Get(),
            ed.scrollPosition.Get(),
            ed.windowSize.Get(),
            logHTML)
    })
    
    return ed
}

func (ed *EventDemo) Render() string {
    return fmt.Sprintf(`
        <div class="event-demo">
            <h2>Event Handling Demo</h2>
            
            <div class="interactive-area" 
                 data-mousemove="handleMouseMove"
                 data-keydown="handleKeyDown"
                 data-scroll="handleScroll"
                 data-resize="handleResize"
                 tabindex="0"
                 style="height: 300px; border: 2px solid #ccc; padding: 20px; overflow-y: auto;">
                
                <p>This area captures various events:</p>
                <ul>
                    <li>Move your mouse around</li>
                    <li>Press keys (click here first to focus)</li>
                    <li>Scroll within this area</li>
                    <li>Resize the window</li>
                </ul>
                
                <div style="height: 500px; background: linear-gradient(to bottom, #f0f0f0, #e0e0e0);">
                    <p>Scroll content...</p>
                </div>
            </div>
            
            <div data-html="display">%s</div>
            
            <button data-click="clearLog">Clear Event Log</button>
        </div>
    `, ed.displayHTML.Get())
}

func (ed *EventDemo) Attach() {
    ed.BindHTML("display", ed.displayHTML)
    ed.BindClick("clearLog", ed.clearLog)
    
    // Custom event handlers
    ed.bindCustomEvents()
}

func (ed *EventDemo) bindCustomEvents() {
    area := dom.QuerySelector(".interactive-area")
    if area == nil {
        return
    }
    
    // Mouse move handler
    area.AddEventListener("mousemove", false, func(event dom.Event) {
        mouseEvent := event.(*dom.MouseEvent)
        rect := area.GetBoundingClientRect()
        x := int(mouseEvent.ClientX) - int(rect.Left)
        y := int(mouseEvent.ClientY) - int(rect.Top)
        
        position := fmt.Sprintf("%d, %d", x, y)
        ed.mousePosition.Set(position)
        ed.logEvent(fmt.Sprintf("Mouse moved to %s", position))
    })
    
    // Key down handler
    area.AddEventListener("keydown", false, func(event dom.Event) {
        keyEvent := event.(*dom.KeyboardEvent)
        key := keyEvent.Key
        
        ed.keyPressed.Set(key)
        ed.logEvent(fmt.Sprintf("Key pressed: %s", key))
        
        // Prevent default for some keys
        if key == "Tab" {
            event.PreventDefault()
        }
    })
    
    // Scroll handler
    area.AddEventListener("scroll", false, func(event dom.Event) {
        scrollTop := area.ScrollTop()
        ed.scrollPosition.Set(scrollTop)
        ed.logEvent(fmt.Sprintf("Scrolled to %dpx", scrollTop))
    })
    
    // Window resize handler
    window := dom.GetWindow()
    window.AddEventListener("resize", false, func(event dom.Event) {
        width := window.InnerWidth()
        height := window.InnerHeight()
        size := fmt.Sprintf("%dx%d", width, height)
        
        ed.windowSize.Set(size)
        ed.logEvent(fmt.Sprintf("Window resized to %s", size))
    })
    
    // Initialize window size
    width := window.InnerWidth()
    height := window.InnerHeight()
    ed.windowSize.Set(fmt.Sprintf("%dx%d", width, height))
}

func (ed *EventDemo) logEvent(message string) {
    timestamp := time.Now().Format("15:04:05")
    logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
    
    ed.eventLog.Update(func(log []string) []string {
        return append(log, logEntry)
    })
}

func (ed *EventDemo) clearLog() {
    ed.eventLog.Set([]string{})
}
```

## Advanced Form Patterns

### Dynamic Form Builder

Create forms dynamically based on configuration:

```go
type FieldType string

const (
    FieldText     FieldType = "text"
    FieldEmail    FieldType = "email"
    FieldNumber   FieldType = "number"
    FieldTextarea FieldType = "textarea"
    FieldSelect   FieldType = "select"
    FieldCheckbox FieldType = "checkbox"
    FieldRadio    FieldType = "radio"
)

type FieldConfig struct {
    Name        string            `json:"name"`
    Label       string            `json:"label"`
    Type        FieldType         `json:"type"`
    Required    bool              `json:"required"`
    Placeholder string            `json:"placeholder"`
    Options     []string          `json:"options,omitempty"`
    Validation  []ValidationRule  `json:"-"`
}

type DynamicForm struct {
    config    []FieldConfig
    values    map[string]*reactivity.Signal[string]
    errors    map[string]*reactivity.Memo[string]
    touched   map[string]*reactivity.Signal[bool]
    
    isSubmitting *reactivity.Signal[bool]
    submitResult *reactivity.Signal[string]
    
    formHTML *reactivity.Memo[string]
}

func NewDynamicForm(config []FieldConfig) *DynamicForm {
    df := &DynamicForm{
        config:       config,
        values:       make(map[string]*reactivity.Signal[string]),
        errors:       make(map[string]*reactivity.Memo[string]),
        touched:      make(map[string]*reactivity.Signal[bool]),
        isSubmitting: reactivity.NewSignal(false),
        submitResult: reactivity.NewSignal(""),
    }
    
    // Initialize fields
    for _, field := range config {
        df.values[field.Name] = reactivity.NewSignal("")
        df.touched[field.Name] = reactivity.NewSignal(false)
        
        // Create validation memo
        df.errors[field.Name] = reactivity.NewMemo(func() string {
            if !df.touched[field.Name].Get() {
                return ""
            }
            
            value := df.values[field.Name].Get()
            
            // Required validation
            if field.Required && strings.TrimSpace(value) == "" {
                return fmt.Sprintf("%s is required", field.Label)
            }
            
            // Custom validation rules
            for _, rule := range field.Validation {
                if err := rule(value); err != "" {
                    return err
                }
            }
            
            return ""
        })
    }
    
    df.formHTML = reactivity.NewMemo(func() string {
        var fields strings.Builder
        
        for _, field := range df.config {
            fields.WriteString(df.renderField(field))
        }
        
        submitText := "Submit"
        if df.isSubmitting.Get() {
            submitText = "Submitting..."
        }
        
        result := df.submitResult.Get()
        resultHTML := ""
        if result != "" {
            resultHTML = fmt.Sprintf(`<div class="result">%s</div>`, result)
        }
        
        return fmt.Sprintf(`
            <form class="dynamic-form" data-submit="handleSubmit">
                %s
                <button type="submit" %s>%s</button>
                %s
            </form>
        `, fields.String(),
            map[bool]string{true: "disabled", false: ""}[df.isSubmitting.Get()],
            submitText,
            resultHTML)
    })
    
    return df
}

func (df *DynamicForm) renderField(field FieldConfig) string {
    value := df.values[field.Name].Get()
    error := df.errors[field.Name].Get()
    
    errorClass := ""
    errorHTML := ""
    if error != "" {
        errorClass = "error"
        errorHTML = fmt.Sprintf(`<span class="error-message">%s</span>`, error)
    }
    
    var input string
    
    switch field.Type {
    case FieldText, FieldEmail, FieldNumber:
        input = fmt.Sprintf(`
            <input type="%s" id="%s" data-input="%s" data-blur="%s-blur"
                   value="%s" placeholder="%s" %s>
        `, field.Type, field.Name, field.Name, field.Name, value, field.Placeholder,
            map[bool]string{true: "required", false: ""}[field.Required])
            
    case FieldTextarea:
        input = fmt.Sprintf(`
            <textarea id="%s" data-input="%s" data-blur="%s-blur"
                      placeholder="%s" rows="4" %s>%s</textarea>
        `, field.Name, field.Name, field.Name, field.Placeholder,
            map[bool]string{true: "required", false: ""}[field.Required], value)
            
    case FieldSelect:
        var options strings.Builder
        options.WriteString(`<option value="">Select...</option>`)
        for _, option := range field.Options {
            selected := ""
            if value == option {
                selected = "selected"
            }
            options.WriteString(fmt.Sprintf(`<option value="%s" %s>%s</option>`, option, selected, option))
        }
        
        input = fmt.Sprintf(`
            <select id="%s" data-change="%s" data-blur="%s-blur" %s>
                %s
            </select>
        `, field.Name, field.Name, field.Name,
            map[bool]string{true: "required", false: ""}[field.Required], options.String())
            
    case FieldCheckbox:
        checked := ""
        if value == "true" {
            checked = "checked"
        }
        input = fmt.Sprintf(`
            <label>
                <input type="checkbox" data-change="%s" %s>
                %s
            </label>
        `, field.Name, checked, field.Label)
        
        // For checkbox, return early with different structure
        return fmt.Sprintf(`
            <div class="field checkbox-field %s">
                %s
                %s
            </div>
        `, errorClass, input, errorHTML)
        
    case FieldRadio:
        var radios strings.Builder
        for _, option := range field.Options {
            checked := ""
            if value == option {
                checked = "checked"
            }
            radios.WriteString(fmt.Sprintf(`
                <label>
                    <input type="radio" name="%s" value="%s" data-change="%s" %s>
                    %s
                </label>
            `, field.Name, option, field.Name, checked, option))
        }
        input = radios.String()
    }
    
    return fmt.Sprintf(`
        <div class="field %s">
            <label for="%s">%s:</label>
            %s
            %s
        </div>
    `, errorClass, field.Name, field.Label, input, errorHTML)
}

func (df *DynamicForm) Render() string {
    return fmt.Sprintf(`
        <div class="dynamic-form-container">
            <h2>Dynamic Form</h2>
            <div data-html="form">%s</div>
        </div>
    `, df.formHTML.Get())
}

func (df *DynamicForm) Attach() {
    // Bind all field handlers
    for _, field := range df.config {
        switch field.Type {
        case FieldText, FieldEmail, FieldNumber, FieldTextarea:
            df.BindInput(field.Name, df.values[field.Name])
        case FieldSelect, FieldRadio:
            df.BindChange(field.Name, df.values[field.Name])
        case FieldCheckbox:
            df.BindChange(field.Name, func(checked bool) {
                df.values[field.Name].Set(fmt.Sprintf("%t", checked))
            })
        }
        
        // Bind blur handlers
        df.BindBlur(field.Name+"-blur", func() {
            df.touched[field.Name].Set(true)
        })
    }
    
    df.BindSubmit("handleSubmit", df.handleSubmit)
    df.BindHTML("form", df.formHTML)
}

func (df *DynamicForm) handleSubmit() {
    // Mark all fields as touched
    for _, field := range df.config {
        df.touched[field.Name].Set(true)
    }
    
    // Check for validation errors
    hasErrors := false
    for _, field := range df.config {
        if df.errors[field.Name].Get() != "" {
            hasErrors = true
            break
        }
    }
    
    if hasErrors {
        df.submitResult.Set("Please fix the errors above.")
        return
    }
    
    if df.isSubmitting.Get() {
        return
    }
    
    df.isSubmitting.Set(true)
    df.submitResult.Set("")
    
    // Collect form data
    formData := make(map[string]string)
    for _, field := range df.config {
        formData[field.Name] = df.values[field.Name].Get()
    }
    
    // Simulate API call
    go func() {
        time.Sleep(2 * time.Second)
        
        logutil.Logf("Form submitted with data: %+v", formData)
        df.submitResult.Set("Form submitted successfully!")
        
        // Clear form
        for _, field := range df.config {
            df.values[field.Name].Set("")
            df.touched[field.Name].Set(false)
        }
        
        df.isSubmitting.Set(false)
    }()
}

// Example usage
func CreateSampleForm() *DynamicForm {
    config := []FieldConfig{
        {
            Name:        "name",
            Label:       "Full Name",
            Type:        FieldText,
            Required:    true,
            Placeholder: "Enter your full name",
            Validation:  []ValidationRule{MinLength(2)},
        },
        {
            Name:        "email",
            Label:       "Email Address",
            Type:        FieldEmail,
            Required:    true,
            Placeholder: "Enter your email",
            Validation:  []ValidationRule{Email},
        },
        {
            Name:        "age",
            Label:       "Age",
            Type:        FieldNumber,
            Required:    true,
            Placeholder: "Enter your age",
        },
        {
            Name:    "country",
            Label:   "Country",
            Type:    FieldSelect,
            Required: true,
            Options: []string{"USA", "Canada", "UK", "Australia", "Other"},
        },
        {
            Name:    "gender",
            Label:   "Gender",
            Type:    FieldRadio,
            Options: []string{"Male", "Female", "Other", "Prefer not to say"},
        },
        {
            Name:        "bio",
            Label:       "Bio",
            Type:        FieldTextarea,
            Placeholder: "Tell us about yourself...",
        },
        {
            Name:  "newsletter",
            Label: "Subscribe to newsletter",
            Type:  FieldCheckbox,
        },
    }
    
    return NewDynamicForm(config)
}
```

## File Uploads

### File Upload Component

Handle file uploads with progress and validation:

```go
type FileUpload struct {
    selectedFiles *reactivity.Signal[[]File]
    uploadProgress *reactivity.Signal[map[string]int]
    uploadStatus  *reactivity.Signal[map[string]string]
    
    maxFileSize   int64
    allowedTypes  []string
    
    uploadHTML *reactivity.Memo[string]
}

type File struct {
    Name string
    Size int64
    Type string
}

func NewFileUpload(maxFileSize int64, allowedTypes []string) *FileUpload {
    fu := &FileUpload{
        selectedFiles:  reactivity.NewSignal([]File{}),
        uploadProgress: reactivity.NewSignal(make(map[string]int)),
        uploadStatus:   reactivity.NewSignal(make(map[string]string)),
        maxFileSize:    maxFileSize,
        allowedTypes:   allowedTypes,
    }
    
    fu.uploadHTML = reactivity.NewMemo(func() string {
        files := fu.selectedFiles.Get()
        progress := fu.uploadProgress.Get()
        status := fu.uploadStatus.Get()
        
        if len(files) == 0 {
            return `
                <div class="upload-area">
                    <p>Drag and drop files here or click to select</p>
                    <input type="file" data-change="fileSelect" multiple accept=".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx">
                </div>
            `
        }
        
        var fileList strings.Builder
        fileList.WriteString(`<div class="file-list">`)
        
        for _, file := range files {
            prog := progress[file.Name]
            stat := status[file.Name]
            
            if stat == "" {
                stat = "Ready"
            }
            
            fileList.WriteString(fmt.Sprintf(`
                <div class="file-item">
                    <div class="file-info">
                        <strong>%s</strong>
                        <span class="file-size">(%s)</span>
                    </div>
                    <div class="file-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: %d%%"></div>
                        </div>
                        <span class="status">%s (%d%%)</span>
                    </div>
                    <button data-click="remove-%s">Remove</button>
                </div>
            `, file.Name, formatFileSize(file.Size), prog, stat, prog, file.Name))
        }
        
        fileList.WriteString(`</div>`)
        
        return fmt.Sprintf(`
            %s
            <div class="upload-controls">
                <button data-click="uploadAll">Upload All</button>
                <button data-click="clearAll">Clear All</button>
                <input type="file" data-change="fileSelect" multiple accept=".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx">
            </div>
        `, fileList.String())
    })
    
    return fu
}

func formatFileSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (fu *FileUpload) Render() string {
    return fmt.Sprintf(`
        <div class="file-upload-container">
            <h2>File Upload</h2>
            <div class="upload-info">
                <p>Max file size: %s</p>
                <p>Allowed types: %s</p>
            </div>
            <div data-html="upload">%s</div>
        </div>
    `, formatFileSize(fu.maxFileSize), strings.Join(fu.allowedTypes, ", "), fu.uploadHTML.Get())
}

func (fu *FileUpload) Attach() {
    fu.BindHTML("upload", fu.uploadHTML)
    fu.BindChange("fileSelect", fu.handleFileSelect)
    fu.BindClick("uploadAll", fu.uploadAll)
    fu.BindClick("clearAll", fu.clearAll)
    
    // Bind remove handlers for each file
    reactivity.NewEffect(func() {
        files := fu.selectedFiles.Get()
        for _, file := range files {
            fu.bindRemoveHandler(file.Name)
        }
    })
}

func (fu *FileUpload) handleFileSelect(files []File) {
    validFiles := []File{}
    
    for _, file := range files {
        // Validate file size
        if file.Size > fu.maxFileSize {
            fu.setFileStatus(file.Name, fmt.Sprintf("Error: File too large (max %s)", formatFileSize(fu.maxFileSize)))
            continue
        }
        
        // Validate file type
        validType := false
        for _, allowedType := range fu.allowedTypes {
            if strings.HasSuffix(strings.ToLower(file.Name), allowedType) {
                validType = true
                break
            }
        }
        
        if !validType {
            fu.setFileStatus(file.Name, fmt.Sprintf("Error: Invalid file type (allowed: %s)", strings.Join(fu.allowedTypes, ", ")))
            continue
        }
        
        validFiles = append(validFiles, file)
        fu.setFileProgress(file.Name, 0)
        fu.setFileStatus(file.Name, "Ready")
    }
    
    fu.selectedFiles.Update(func(existing []File) []File {
        return append(existing, validFiles...)
    })
}

func (fu *FileUpload) bindRemoveHandler(fileName string) {
    selector := fmt.Sprintf("remove-%s", fileName)
    fu.BindClick(selector, func() {
        fu.removeFile(fileName)
    })
}

func (fu *FileUpload) removeFile(fileName string) {
    fu.selectedFiles.Update(func(files []File) []File {
        filtered := []File{}
        for _, file := range files {
            if file.Name != fileName {
                filtered = append(filtered, file)
            }
        }
        return filtered
    })
    
    // Clean up progress and status
    fu.uploadProgress.Update(func(progress map[string]int) map[string]int {
        delete(progress, fileName)
        return progress
    })
    
    fu.uploadStatus.Update(func(status map[string]string) map[string]string {
        delete(status, fileName)
        return status
    })
}

func (fu *FileUpload) uploadAll() {
    files := fu.selectedFiles.Get()
    
    for _, file := range files {
        if fu.getFileStatus(file.Name) == "Ready" {
            go fu.uploadFile(file)
        }
    }
}

func (fu *FileUpload) uploadFile(file File) {
    fu.setFileStatus(file.Name, "Uploading")
    
    // Simulate upload progress
    for i := 0; i <= 100; i += 10 {
        time.Sleep(200 * time.Millisecond)
        fu.setFileProgress(file.Name, i)
        
        if i == 100 {
            fu.setFileStatus(file.Name, "Completed")
        }
    }
}

func (fu *FileUpload) clearAll() {
    fu.selectedFiles.Set([]File{})
    fu.uploadProgress.Set(make(map[string]int))
    fu.uploadStatus.Set(make(map[string]string))
}

func (fu *FileUpload) setFileProgress(fileName string, progress int) {
    fu.uploadProgress.Update(func(progressMap map[string]int) map[string]int {
        progressMap[fileName] = progress
        return progressMap
    })
}

func (fu *FileUpload) setFileStatus(fileName, status string) {
    fu.uploadStatus.Update(func(statusMap map[string]string) map[string]string {
        statusMap[fileName] = status
        return statusMap
    })
}

func (fu *FileUpload) getFileStatus(fileName string) string {
    status := fu.uploadStatus.Get()
    return status[fileName]
}
```

## Best Practices

### 1. Form State Management

```go
// GOOD: Centralized form state
type FormState struct {
    values   map[string]*reactivity.Signal[string]
    errors   map[string]*reactivity.Memo[string]
    touched  map[string]*reactivity.Signal[bool]
    isValid  *reactivity.Memo[bool]
}

// BAD: Scattered state
type BadForm struct {
    name      *reactivity.Signal[string]
    nameError *reactivity.Signal[string]
    email     *reactivity.Signal[string]
    emailError *reactivity.Signal[string]
    // ... scattered everywhere
}
```

### 2. Event Handler Organization

```go
// GOOD: Organized event binding
func (c *Component) Attach() {
    // Group related bindings
    c.bindInputs()
    c.bindButtons()
    c.bindCustomEvents()
}

func (c *Component) bindInputs() {
    c.BindInput("name", c.name)
    c.BindInput("email", c.email)
}

func (c *Component) bindButtons() {
    c.BindClick("submit", c.handleSubmit)
    c.BindClick("cancel", c.handleCancel)
}

// BAD: All bindings mixed together
func (c *BadComponent) Attach() {
    c.BindInput("name", c.name)
    c.BindClick("submit", c.handleSubmit)
    c.BindInput("email", c.email)
    c.BindClick("cancel", c.handleCancel)
    // ... mixed and confusing
}
```

### 3. Validation Strategy

```go
// GOOD: Reusable validation rules
var CommonRules = struct {
    Required    ValidationRule
    Email       ValidationRule
    MinLength   func(int) ValidationRule
    MaxLength   func(int) ValidationRule
    Pattern     func(string, string) ValidationRule
}{
    Required: func(value string) string {
        if strings.TrimSpace(value) == "" {
            return "This field is required"
        }
        return ""
    },
    Email: func(value string) string {
        if value != "" && !isValidEmail(value) {
            return "Must be a valid email address"
        }
        return ""
    },
    // ... other rules
}

// Usage
field := NewValidatedField("", CommonRules.Required, CommonRules.Email)
```

### 4. Error Handling

```go
// GOOD: Comprehensive error handling
func (f *Form) handleSubmit() {
    // Validate before submission
    if !f.isValid() {
        f.showValidationErrors()
        return
    }
    
    // Prevent double submission
    if f.isSubmitting.Get() {
        return
    }
    
    f.isSubmitting.Set(true)
    
    go func() {
        defer f.isSubmitting.Set(false)
        
        if err := f.submitToAPI(); err != nil {
            f.handleSubmissionError(err)
            return
        }
        
        f.handleSubmissionSuccess()
    }()
}

// BAD: No error handling
func (f *BadForm) handleSubmit() {
    f.submitToAPI() // What if this fails?
}
```

### 5. Performance Optimization

```go
// GOOD: Debounced validation
func (f *Field) setupValidation() {
    var validationTimer *time.Timer
    
    reactivity.NewEffect(func() {
        value := f.value.Get()
        
        if validationTimer != nil {
            validationTimer.Stop()
        }
        
        validationTimer = time.AfterFunc(300*time.Millisecond, func() {
            f.validateValue(value)
        })
    })
}

// BAD: Immediate validation on every keystroke
func (f *BadField) setupValidation() {
    reactivity.NewEffect(func() {
        value := f.value.Get()
        f.validateValue(value) // Runs on every keystroke
    })
}
```

Next: Explore [API Reference](../api/core-apis.md) or check [Troubleshooting](../troubleshooting.md) for common issues.