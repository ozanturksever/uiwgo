package dom

// Good idea, not implemented
//import (
//	"github.com/ozanturksever/uiwgo/reactivity"
//	g "maragu.dev/gomponents"
//)

//// Show renders its children only when the provided signal is true.
//// It renders a placeholder comment node when the signal is false.
//func Show(when reactivity.Signal[bool], children ...g.Node) g.Node {
//	// Return a dynamic node that renders based on the signal's value.
//	return Dynamic(func() g.Node {
//		if when.Get() {
//			return g.Group(children)
//		}
//		// Render a placeholder so the DOM structure is maintained,
//		// which is important for frameworks that rely on node positions.
//		return Comment("hidden")
//	})
//}