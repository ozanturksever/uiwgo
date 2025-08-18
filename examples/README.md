# Golid Examples

This directory contains simple examples demonstrating various features of the Golid framework. Each example is self-contained and can be run independently using the provided Makefile commands.

## Available Examples

### 1. Counter Example (`counter/`)
**Demonstrates:** Basic signals and reactive updates

**Features:**
- Reactive signals with `golid.NewSignal()`
- Button click handlers with `golid.OnClick()`
- Reactive text binding with `golid.BindText()`
- State-based conditional rendering

**Run with:**
```bash
make example-counter
```

### 2. Todo List Example (`todo/`)
**Demonstrates:** List rendering and form handling

**Features:**
- Dynamic list rendering with `golid.ForEach()`
- Form input handling with `golid.OnInput()`
- Adding, removing, and toggling todo items
- Reactive statistics display
- Complex state management with arrays

**Run with:**
```bash
make example-todo
```

### 3. Router Example (`router/`)
**Demonstrates:** Client-side routing

**Features:**
- Multi-page application with `golid.NewRouter()`
- Route parameter extraction (`:id`, `:category/:id`)
- Navigation with `golid.RouterLink()`
- Route outlets with `golid.RouterOutlet()`
- Form handling within routed pages
- Static and dynamic routes

**Run with:**
```bash
make example-router
```

### 4. Conditional Rendering Example (`conditional/`)
**Demonstrates:** Dynamic UI updates based on state

**Features:**
- Show/hide functionality with `golid.Bind()`
- Multi-state conditional rendering (switch statements)
- Dynamic styling based on state
- User role-based UI changes
- State-dependent visual feedback

**Run with:**
```bash
make example-conditional
```

### 5. Lifecycle Example (`lifecycle/`)
**Demonstrates:** Component lifecycle hooks and cleanup patterns

**Features:**
- Component lifecycle hooks (`OnInit`, `OnMount`, `OnDismount`)
- Dynamic component creation and destruction
- Resource cleanup patterns (timers, connections)
- Real-time lifecycle event logging
- Timer management and goroutine cleanup
- State reset on component unmount

**Run with:**
```bash
make example-lifecycle
```

## How to Run Examples

### Prerequisites
- Go 1.23.0+ installed
- Run `make setup` from the project root to build the development server

### Running Individual Examples

1. **List all examples:**
   ```bash
   make examples-list
   ```

2. **Run a specific example:**
   ```bash
   make example-counter      # Counter example
   make example-todo         # Todo list example
   make example-router       # Router example
   make example-conditional  # Conditional rendering example
   ```

3. **Access the example:**
   - Each command will start a development server on http://localhost:8090
   - Open your browser and navigate to the URL
   - Press Ctrl+C to stop the server

### Manual Compilation (Optional)

If you want to manually compile an example without running the server:

```bash
# Compile to WebAssembly
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm ./examples/counter/main.go

# Or use TinyGo for smaller bundle size
tinygo build -o main.wasm -target wasm ./examples/counter/main.go
```

## Example Structure

Each example follows the same structure:
```
examples/[example-name]/
├── main.go          # Main application entry point
└── README.md        # Individual example documentation (optional)
```

### Code Patterns

All examples demonstrate these common Golid patterns:

1. **Signal Creation:**
   ```go
   count := golid.NewSignal(0)
   items := golid.NewSignal([]Item{})
   ```

2. **Reactive Binding:**
   ```go
   golid.BindText(func() string {
       return fmt.Sprintf("Count: %d", count.Get())
   })
   ```

3. **Event Handling:**
   ```go
   golid.OnClick(func() {
       count.Set(count.Get() + 1)
   })
   ```

4. **Conditional Rendering:**
   ```go
   golid.Bind(func() Node {
       if showDetails.Get() {
           return Div(Text("Details"))
       }
       return Text("")
   })
   ```

## Development Tips

1. **Hot Reload:** The development server automatically rebuilds WASM when you modify Go files
2. **Browser DevTools:** Use the browser console to see Go runtime messages and errors
3. **Debugging:** Add `golid.Log()` or `golid.Logf()` statements for debugging
4. **Styling:** Examples use inline styles for simplicity, but you can use CSS classes too

## Learn More

- Check the main project documentation for advanced features
- Explore the `golid/golid.go` source code for all available functions
- Look at the main application (`main.go`) for more complex examples

## Contributing

Feel free to add more examples! Each example should:
- Be self-contained in its own directory
- Demonstrate a specific Golid feature or pattern
- Include clear comments explaining the concepts
- Be added to the main Makefile with a corresponding `example-[name]` target