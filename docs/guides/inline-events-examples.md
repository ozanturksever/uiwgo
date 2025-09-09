# Inline Events: Examples and Best Practices

This guide provides practical examples and best practices for using inline event binding in uiwgo applications.

## Table of Contents

- [Basic Examples](#basic-examples)
- [Advanced Patterns](#advanced-patterns)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Integration with Signals](#integration-with-signals)
- [Form Handling](#form-handling)
- [Keyboard Navigation](#keyboard-navigation)

## Basic Examples

### Simple Counter

```go
func Counter() g.Node {
    count := reactivity.NewSignal(0)

    return Div(
        H2(Text("Counter Example")),
        P(comps.BindText(func() string {
            return fmt.Sprintf("Count: %d", count.Get())
        })),
        Button(
            Text("Increment"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(count.Get() + 1)
            }),
        ),
        Button(
            Text("Decrement"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(count.Get() - 1)
            }),
        ),
        Button(
            Text("Reset"),
            dom.OnClickInline(func(el dom.Element) {
                count.Set(0)
            }),
        ),
    )
}
```

### Input Handling

```go
func InputExample() g.Node {
    text := reactivity.NewSignal("")
    wordCount := reactivity.NewComputed(func() int {
        words := strings.Fields(text.Get())
        return len(words)
    })

    return Div(
        H2(Text("Input Example")),
        Input(
            Type("text"),
            Placeholder("Type something..."),
            dom.OnInputInline(func(el dom.Element) {
                text.Set(el.Get("value").String())
            }),
        ),
        P(comps.BindText(func() string {
            return fmt.Sprintf("Text: %s", text.Get())
        })),
        P(comps.BindText(func() string {
            return fmt.Sprintf("Word count: %d", wordCount.Get())
        })),
    )
}
```

## Advanced Patterns

### Todo List with Multiple Event Types

```go
type Todo struct {
    ID   int
    Text string
    Done bool
}

func TodoApp() g.Node {
    todos := reactivity.NewSignal([]Todo{})
    newTodoText := reactivity.NewSignal("")
    nextID := reactivity.NewSignal(1)

    addTodo := func() {
        text := strings.TrimSpace(newTodoText.Get())
        if text == "" {
            return
        }
        
        currentTodos := todos.Get()
        newTodo := Todo{
            ID:   nextID.Get(),
            Text: text,
            Done: false,
        }
        
        todos.Set(append(currentTodos, newTodo))
        newTodoText.Set("")
        nextID.Set(nextID.Get() + 1)
    }

    toggleTodo := func(id int) {
        currentTodos := todos.Get()
        for i, todo := range currentTodos {
            if todo.ID == id {
                currentTodos[i].Done = !currentTodos[i].Done
                break
            }
        }
        todos.Set(currentTodos)
    }

    deleteTodo := func(id int) {
        currentTodos := todos.Get()
        filtered := make([]Todo, 0, len(currentTodos))
        for _, todo := range currentTodos {
            if todo.ID != id {
                filtered = append(filtered, todo)
            }
        }
        todos.Set(filtered)
    }

    return Div(
        H1(Text("Todo App")),
        
        // Add todo form
        Div(
            Input(
                Type("text"),
                Placeholder("Add a new todo..."),
                comps.BindValue(newTodoText),
                dom.OnInputInline(func(el dom.Element) {
                    newTodoText.Set(el.Get("value").String())
                }),
                dom.OnEnterInline(func(el dom.Element) {
                    addTodo()
                }),
                dom.OnEscapeInline(func(el dom.Element) {
                    newTodoText.Set("")
                }),
            ),
            Button(
                Text("Add"),
                dom.OnClickInline(func(el dom.Element) {
                    addTodo()
                }),
            ),
        ),
        
        // Todo list
        comps.BindList(todos, func(todo Todo) g.Node {
            return Div(
                Class("todo-item"),
                Input(
                    Type("checkbox"),
                    If(todo.Done, Checked()),
                    dom.OnChangeInline(func(el dom.Element) {
                        toggleTodo(todo.ID)
                    }),
                ),
                Span(
                    Text(todo.Text),
                    If(todo.Done, Class("completed")),
                ),
                Button(
                    Text("Delete"),
                    dom.OnClickInline(func(el dom.Element) {
                        deleteTodo(todo.ID)
                    }),
                ),
            )
        }),
    )
}
```

### Dynamic Form with Validation

```go
func ContactForm() g.Node {
    name := reactivity.NewSignal("")
    email := reactivity.NewSignal("")
    message := reactivity.NewSignal("")
    
    nameError := reactivity.NewComputed(func() string {
        if len(strings.TrimSpace(name.Get())) < 2 {
            return "Name must be at least 2 characters"
        }
        return ""
    })
    
    emailError := reactivity.NewComputed(func() string {
        email := strings.TrimSpace(email.Get())
        if email == "" {
            return "Email is required"
        }
        if !strings.Contains(email, "@") {
            return "Invalid email format"
        }
        return ""
    })
    
    isValid := reactivity.NewComputed(func() bool {
        return nameError.Get() == "" && emailError.Get() == "" && 
               strings.TrimSpace(message.Get()) != ""
    })

    submitForm := func() {
        if !isValid.Get() {
            return
        }
        
        logutil.Logf("Submitting form: %s, %s, %s", 
            name.Get(), email.Get(), message.Get())
        
        // Reset form
        name.Set("")
        email.Set("")
        message.Set("")
    }

    return Form(
        H2(Text("Contact Form")),
        
        Div(
            Label(Text("Name:")),
            Input(
                Type("text"),
                comps.BindValue(name),
                dom.OnInputInline(func(el dom.Element) {
                    name.Set(el.Get("value").String())
                }),
            ),
            comps.BindIf(func() bool { return nameError.Get() != "" }, func() g.Node {
                return Div(
                    Class("error"),
                    Text(nameError.Get()),
                )
            }),
        ),
        
        Div(
            Label(Text("Email:")),
            Input(
                Type("email"),
                comps.BindValue(email),
                dom.OnInputInline(func(el dom.Element) {
                    email.Set(el.Get("value").String())
                }),
            ),
            comps.BindIf(func() bool { return emailError.Get() != "" }, func() g.Node {
                return Div(
                    Class("error"),
                    Text(emailError.Get()),
                )
            }),
        ),
        
        Div(
            Label(Text("Message:")),
            Textarea(
                Placeholder("Your message..."),
                comps.BindValue(message),
                dom.OnInputInline(func(el dom.Element) {
                    message.Set(el.Get("value").String())
                }),
                dom.OnKeyDownInline(func(el dom.Element) {
                    // Ctrl+Enter to submit
                    event := el.Get("event")
                    if event.Get("ctrlKey").Bool() {
                        submitForm()
                    }
                }, "Enter"),
            ),
        ),
        
        Button(
            Type("button"),
            Text("Submit"),
            comps.BindDisabled(func() bool { return !isValid.Get() }),
            dom.OnClickInline(func(el dom.Element) {
                submitForm()
            }),
        ),
    )
}
```

## Best Practices

### 1. Keep Handlers Simple

```go
// ✅ Good: Simple, focused handler
dom.OnClickInline(func(el dom.Element) {
    count.Set(count.Get() + 1)
})

// ❌ Avoid: Complex logic in handlers
dom.OnClickInline(func(el dom.Element) {
    // 50 lines of complex business logic...
})

// ✅ Better: Extract to functions
handleIncrement := func() {
    // Complex logic here
}

dom.OnClickInline(func(el dom.Element) {
    handleIncrement()
})
```

### 2. Use Closures for Context

```go
// ✅ Good: Capture necessary context
func TodoItem(todo Todo, onToggle func(int), onDelete func(int)) g.Node {
    return Div(
        Input(
            Type("checkbox"),
            If(todo.Done, Checked()),
            dom.OnChangeInline(func(el dom.Element) {
                onToggle(todo.ID) // Captured from closure
            }),
        ),
        Button(
            Text("Delete"),
            dom.OnClickInline(func(el dom.Element) {
                onDelete(todo.ID) // Captured from closure
            }),
        ),
    )
}
```

### 3. Combine with Signals Effectively

```go
// ✅ Good: Direct signal updates
dom.OnInputInline(func(el dom.Element) {
    searchTerm.Set(el.Get("value").String())
})

// ✅ Good: Computed values for derived state
filteredItems := reactivity.NewComputed(func() []Item {
    term := strings.ToLower(searchTerm.Get())
    var filtered []Item
    for _, item := range allItems.Get() {
        if strings.Contains(strings.ToLower(item.Name), term) {
            filtered = append(filtered, item)
        }
    }
    return filtered
})
```

### 4. Handle Edge Cases

```go
// ✅ Good: Validate input
dom.OnInputInline(func(el dom.Element) {
    value := el.Get("value").String()
    if value == "" {
        return // Handle empty input
    }
    
    if len(value) > 100 {
        value = value[:100] // Limit length
    }
    
    textSignal.Set(value)
})

// ✅ Good: Prevent double-clicks
var submitting bool
dom.OnClickInline(func(el dom.Element) {
    if submitting {
        return
    }
    submitting = true
    defer func() { submitting = false }()
    
    submitForm()
})
```

## Common Patterns

### Toggle Pattern

```go
func ToggleButton(label string, state reactivity.Signal[bool]) g.Node {
    return Button(
        comps.BindText(func() string {
            if state.Get() {
                return "Hide " + label
            }
            return "Show " + label
        }),
        dom.OnClickInline(func(el dom.Element) {
            state.Set(!state.Get())
        }),
    )
}
```

### Increment/Decrement Pattern

```go
func NumberInput(value reactivity.Signal[int], min, max int) g.Node {
    return Div(
        Button(
            Text("-"),
            comps.BindDisabled(func() bool { return value.Get() <= min }),
            dom.OnClickInline(func(el dom.Element) {
                if value.Get() > min {
                    value.Set(value.Get() - 1)
                }
            }),
        ),
        Span(
            comps.BindText(func() string {
                return fmt.Sprintf("%d", value.Get())
            }),
        ),
        Button(
            Text("+"),
            comps.BindDisabled(func() bool { return value.Get() >= max }),
            dom.OnClickInline(func(el dom.Element) {
                if value.Get() < max {
                    value.Set(value.Get() + 1)
                }
            }),
        ),
    )
}
```

### Selection Pattern

```go
func RadioGroup(options []string, selected reactivity.Signal[string]) g.Node {
    return Div(
        comps.BindList(reactivity.NewSignal(options), func(option string) g.Node {
            return Label(
                Input(
                    Type("radio"),
                    Name("radio-group"),
                    Value(option),
                    comps.BindChecked(func() bool {
                        return selected.Get() == option
                    }),
                    dom.OnChangeInline(func(el dom.Element) {
                        if el.Get("checked").Bool() {
                            selected.Set(option)
                        }
                    }),
                ),
                Text(option),
            )
        }),
    )
}
```

## Integration with Signals

### Reactive Forms

```go
func ReactiveForm() g.Node {
    formData := reactivity.NewSignal(map[string]string{
        "name":  "",
        "email": "",
    })
    
    updateField := func(field, value string) {
        data := formData.Get()
        newData := make(map[string]string)
        for k, v := range data {
            newData[k] = v
        }
        newData[field] = value
        formData.Set(newData)
    }
    
    return Div(
        Input(
            Type("text"),
            Placeholder("Name"),
            dom.OnInputInline(func(el dom.Element) {
                updateField("name", el.Get("value").String())
            }),
        ),
        Input(
            Type("email"),
            Placeholder("Email"),
            dom.OnInputInline(func(el dom.Element) {
                updateField("email", el.Get("value").String())
            }),
        ),
        Pre(
            comps.BindText(func() string {
                data := formData.Get()
                return fmt.Sprintf("Form Data: %+v", data)
            }),
        ),
    )
}
```

### Debounced Input

```go
func DebouncedSearch() g.Node {
    searchTerm := reactivity.NewSignal("")
    debouncedTerm := reactivity.NewSignal("")
    
    var debounceTimer *time.Timer
    
    return Div(
        Input(
            Type("text"),
            Placeholder("Search..."),
            dom.OnInputInline(func(el dom.Element) {
                value := el.Get("value").String()
                searchTerm.Set(value)
                
                // Debounce the search
                if debounceTimer != nil {
                    debounceTimer.Stop()
                }
                
                debounceTimer = time.AfterFunc(300*time.Millisecond, func() {
                    debouncedTerm.Set(value)
                })
            }),
        ),
        P(
            comps.BindText(func() string {
                return fmt.Sprintf("Searching for: %s", debouncedTerm.Get())
            }),
        ),
    )
}
```

## Form Handling

### Multi-Step Form

```go
func MultiStepForm() g.Node {
    currentStep := reactivity.NewSignal(1)
    formData := reactivity.NewSignal(map[string]interface{}{})
    
    nextStep := func() {
        if currentStep.Get() < 3 {
            currentStep.Set(currentStep.Get() + 1)
        }
    }
    
    prevStep := func() {
        if currentStep.Get() > 1 {
            currentStep.Set(currentStep.Get() - 1)
        }
    }
    
    return Div(
        H2(comps.BindText(func() string {
            return fmt.Sprintf("Step %d of 3", currentStep.Get())
        })),
        
        comps.BindSwitch(currentStep, map[int]func() g.Node{
            1: func() g.Node {
                return Div(
                    H3(Text("Personal Information")),
                    Input(
                        Type("text"),
                        Placeholder("First Name"),
                        dom.OnInputInline(func(el dom.Element) {
                            data := formData.Get()
                            data["firstName"] = el.Get("value").String()
                            formData.Set(data)
                        }),
                    ),
                )
            },
            2: func() g.Node {
                return Div(
                    H3(Text("Contact Information")),
                    Input(
                        Type("email"),
                        Placeholder("Email"),
                        dom.OnInputInline(func(el dom.Element) {
                            data := formData.Get()
                            data["email"] = el.Get("value").String()
                            formData.Set(data)
                        }),
                    ),
                )
            },
            3: func() g.Node {
                return Div(
                    H3(Text("Review")),
                    Pre(comps.BindText(func() string {
                        return fmt.Sprintf("%+v", formData.Get())
                    })),
                )
            },
        }),
        
        Div(
            comps.BindIf(func() bool { return currentStep.Get() > 1 }, func() g.Node {
                return Button(
                    Text("Previous"),
                    dom.OnClickInline(func(el dom.Element) {
                        prevStep()
                    }),
                )
            }),
            comps.BindIf(func() bool { return currentStep.Get() < 3 }, func() g.Node {
                return Button(
                    Text("Next"),
                    dom.OnClickInline(func(el dom.Element) {
                        nextStep()
                    }),
                )
            }),
            comps.BindIf(func() bool { return currentStep.Get() == 3 }, func() g.Node {
                return Button(
                    Text("Submit"),
                    dom.OnClickInline(func(el dom.Element) {
                        logutil.Log("Form submitted:", formData.Get())
                    }),
                )
            }),
        ),
    )
}
```

## Keyboard Navigation

### Arrow Key Navigation

```go
func NavigableList(items []string) g.Node {
    selectedIndex := reactivity.NewSignal(0)
    
    return Div(
        Tabindex("0"), // Make div focusable
        dom.OnKeyDownInline(func(el dom.Element) {
            current := selectedIndex.Get()
            switch {
            case current > 0:
                selectedIndex.Set(current - 1)
            }
        }, "ArrowUp"),
        dom.OnKeyDownInline(func(el dom.Element) {
            current := selectedIndex.Get()
            if current < len(items)-1 {
                selectedIndex.Set(current + 1)
            }
        }, "ArrowDown"),
        dom.OnEnterInline(func(el dom.Element) {
            selected := items[selectedIndex.Get()]
            logutil.Log("Selected:", selected)
        }),
        
        comps.BindList(reactivity.NewSignal(items), func(item string) g.Node {
            index := 0
            for i, it := range items {
                if it == item {
                    index = i
                    break
                }
            }
            
            return Div(
                comps.BindClass(func() string {
                    if selectedIndex.Get() == index {
                        return "selected"
                    }
                    return ""
                }),
                Text(item),
                dom.OnClickInline(func(el dom.Element) {
                    selectedIndex.Set(index)
                }),
            )
        }),
    )
}
```

This guide provides comprehensive examples and patterns for effectively using inline events in your uiwgo applications. Remember to keep handlers simple, leverage closures for context, and integrate seamlessly with the reactivity system.