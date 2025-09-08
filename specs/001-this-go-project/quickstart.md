# Quickstart: Using React Components

This guide demonstrates how to use a third-party React date picker component in your Go application.

## 1. JavaScript Setup
Create a JavaScript file to register your React components.

```javascript
// src/js/components.js
import React from 'react';
import ReactDOM from 'react-dom';
import DatePicker from 'react-datepicker';

window.renderComponent = (id, name, props) => {
  const container = document.getElementById(id);
  if (name === 'DatePicker') {
    ReactDOM.render(<DatePicker {...props} />, container);
  }
};

window.updateComponent = (id, props) => {
  // This is a simplified example. A real implementation would
  // need to manage component instances.
  window.renderComponent(id, 'DatePicker', props);
};

window.unmountComponent = (id) => {
  const container = document.getElementById(id);
  ReactDOM.unmountComponentAtNode(container);
};
```

## 2. Go Application
In your Go code, you can now render and interact with the `DatePicker` component.

```go
package main

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/ozanturksever/uiwgo/reactivity"
)

func main() {
	// Create a signal to hold the selected date.
	selectedDate := reactivity.CreateSignal(time.Now())

	// Create a container element for the date picker.
	js.Global().Get("document").Get("body").Set("innerHTML", `<div id="date-picker"></div>`)

	// Render the DatePicker component.
	props := map[string]interface{}{
		"selected": selectedDate.Get(),
		"onChange": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// When the date changes, update the Go signal.
			date, _ := time.Parse(time.RFC3339, args[0].String())
			selectedDate.Set(date)
			return nil
		}),
	}
	js.Global().Call("renderComponent", "date-picker", "DatePicker", props)

	// Create an effect to log the selected date when it changes.
	reactivity.CreateEffect(func() {
		fmt.Println("Selected date:", selectedDate.Get())
	})

	// Keep the application running.
	select {}
}

```
