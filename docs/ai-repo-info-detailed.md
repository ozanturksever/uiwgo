# AI Repo Info (Detailed)

This document provides detailed information for AI code generation, including full API signatures for key libraries.

## File/Package Structure

*   `/action`: Core package for managing asynchronous operations, side effects, and state with lifecycle-aware components. Includes concepts like actions, queries, streams, and subscriptions.
*   `/appmanager`: Provides a high-level manager for bootstrapping and running the entire application.
*   `/compat`: Contains compatibility helpers, particularly for integrating with React components.
*   `/comps`: Home to the core component model, page management, and mounting logic.
*   `/debug`: Utilities and examples for debugging applications.
*   `/docs`: Project documentation.
*   `/dom`: Handles DOM manipulation, event handling, and reactive rendering logic.
*   `/examples`: A directory containing various standalone examples demonstrating different features of the framework. Each subdirectory is a runnable application.
*   `/form`: Utilities and components for handling form state and validation.
*   `/internal`: Internal packages used by the framework, not intended for public use. Includes test helpers.
*   `/js`: Low-level JavaScript interop helpers.
*   `/reactivity`: The heart of the reactive system, containing signals, effects, memos, and stores for state management.
*   `/router`: Provides client-side routing capabilities.
*   `/scripts`: Shell scripts for development, building, and testing.
*   `/spec`: Specification documents.
*   `/templates`: Go templates for code generation or other purposes.
*   `/viteassets`: Assets related to the Vite development server.
*   `/wasm`: WASM-related utilities or build artifacts.

## `maragu.dev/gomponents` API (Detailed)

`gomponents` is a library for writing declarative, type-safe HTML in pure Go.

### Index

- [Constants](https://pkg.go.dev/maragu.dev/gomponents#pkg-constants)
- [type Group](https://pkg.go.dev/maragu.dev/gomponents#Group)
  - [func Map[T any](ts []T, cb func(T) Node) Group](https://pkg.go.dev/maragu.dev/gomponents#Map)
  - [func (g Group) Render(w io.Writer) error](https://pkg.go.dev/maragu.dev/gomponents#Group.Render)
  - [func (g Group) String() string](https://pkg.go.dev/maragu.dev/gomponents#Group.String)
- [type Node](https://pkg.go.dev/maragu.dev/gomponents#Node)
  - [func Attr(name string, value ...string) Node](https://pkg.go.dev/maragu.dev/gomponents#Attr)
  - [func El(name string, children ...Node) Node](https://pkg.go.dev/maragu.dev/gomponents#El)
  - [func If(condition bool, n Node) Node](https://pkg.go.dev/maragu.dev/gomponents#If)
  - [func Iff(condition bool, f func() Node) Node](https://pkg.go.dev/maragu.dev/gomponents#Iff)
  - [func Raw(t string) Node](https://pkg.go.dev/maragu.dev/gomponents#Raw)
  - [func Rawf(format string, a ...interface{}) Node](https://pkg.go.dev/maragu.dev/gomponents#Rawf)
  - [func Text(t string) Node](https://pkg.go.dev/maragu.dev/gomponents#Text)
  - [func Textf(format string, a ...interface{}) Node](https://pkg.go.dev/maragu.dev/gomponents#Textf)
- [type NodeFunc](https://pkg.go.dev/maragu.dev/gomponents#NodeFunc)
  - [func (n NodeFunc) Render(w io.Writer) error](https://pkg.go.dev/maragu.dev/gomponents#NodeFunc.Render)
  - [func (n NodeFunc) String() string](https://pkg.go.dev/maragu.dev/gomponents#NodeFunc.String)
  - [func (n NodeFunc) Type() NodeType](https://pkg.go.dev/maragu.dev/gomponents#NodeFunc.Type)
- [type NodeType](https://pkg.go.dev/maragu.dev/gomponents#NodeType)

### Type Details

#### `type Group []Node`
Group is a slice of Nodes.

- `func Map[T any](ts []T, cb func(T) Node) Group`
- `func (g Group) Render(w io.Writer) error`
- `func (g Group) String() string`

#### `type Node interface`
Node is a DOM node that can Render itself to an io.Writer.
- `Render(w io.Writer) error`

##### Functions
- `func Attr(name string, value ...string) Node`
- `func El(name string, children ...Node) Node`
- `func If(condition bool, n Node) Node`
- `func Iff(condition bool, f func() Node) Node`
- `func Raw(t string) Node`
- `func Rawf(format string, a ...interface{}) Node`
- `func Text(t string) Node`
- `func Textf(format string, a ...interface{}) Node`

#### `type NodeFunc func(io.Writer) error`
NodeFunc is a render function that is also a Node.
- `func (n NodeFunc) Render(w io.Writer) error`
- `func (n NodeFunc) String() string`
- `func (n NodeFunc) Type() NodeType`

#### `type NodeType int`
Describes the type of Node (ElementType or AttributeType).

## `honnef.co/go/js/dom/v2` API (Detailed)

This package provides Go bindings for the JavaScript DOM APIs.

### Key Interfaces & Types

- **`Window`**: `func GetWindow() Window`
- **`Document`**: `doc := window.Document()`
- **`Element`**: `el := doc.QuerySelector(...)`
- **`Node`**: Base interface for all DOM nodes.
- **`Event`**: Base interface for all events.

A full list of types and their methods is below.

### Index (Partial)

- [type AnimationEvent](https://pkg.go.dev/honnef.co/go/js/dom/v2#AnimationEvent)
- [type BasicElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicElement)
  - [func (e *BasicElement) Attributes() map[string]string](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicElement.Attributes)
  - [func (e *BasicElement) Class() *TokenList](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicElement.Class)
  - [func (e *BasicElement) QuerySelector(s string) Element](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicElement.QuerySelector)
  - [func (e *BasicElement) QuerySelectorAll(s string) []Element](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicElement.QuerySelectorAll)
- [type BasicEvent](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicEvent)
  - [func (ev *BasicEvent) Target() Element](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicEvent.Target)
  - [func (ev *BasicEvent) PreventDefault()](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicEvent.PreventDefault)
- [type BasicHTMLElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicHTMLElement)
- [type BasicNode](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicNode)
  - [func (n *BasicNode) AddEventListener(typ string, useCapture bool, listener func(Event)) js.Func](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicNode.AddEventListener)
  - [func (n *BasicNode) AppendChild(newchild Node)](https://pkg.go.dev/honnef.co/go/js/dom/v2#BasicNode.AppendChild)
- [type Document](https://pkg.go.dev/honnef.co/go/js/dom/v2#Document)
  - [func WrapDocument(o js.Value) Document](https://pkg.go.dev/honnef.co/go/js/dom/v2#WrapDocument)
- [type Element](https://pkg.go.dev/honnef.co/go/js/dom/v2#Element)
  - [func WrapElement(o js.Value) Element](https://pkg.go.dev/honnef.co/go/js/dom/v2#WrapElement)
- [type Event](https://pkg.go.dev/honnef.co/go/js/dom/v2#Event)
  - [func WrapEvent(o js.Value) Event](https://pkg.go.dev/honnef.co/go/js/dom/v2#WrapEvent)
- [type HTMLButtonElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#HTMLButtonElement)
- [type HTMLCanvasElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#HTMLCanvasElement)
- [type HTMLElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#HTMLElement)
- [type HTMLInputElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#HTMLInputElement)
- [type HTMLSelectElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#HTMLSelectElement)
- [type HTMLTextAreaElement](https://pkg.go.dev/honnef.co/go/js/dom/v2#HTMLTextAreaElement)
- [type KeyboardEvent](https://pkg.go.dev/honnef.co/go/js/dom/v2#KeyboardEvent)
- [type MouseEvent](https://pkg.go.dev/honnef.co/go/js/dom/v2#MouseEvent)
- [type Node](https://pkg.go.dev/honnef.co/go/js/dom/v2#Node)
  - [func WrapNode(o js.Value) Node](https://pkg.go.dev/honnef.co/go/js/dom/v2#WrapNode)
- [type Window](https://pkg.go.dev/honnef.co/go/js/dom/v2#Window)
  - [func GetWindow() Window](https://pkg.go.dev/honnef.co/go/js/dom/v2#GetWindow)

... and many more. Refer to the full `firecrawl_search` output from previous steps for a comprehensive list.
