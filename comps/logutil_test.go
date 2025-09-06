//go:build js && wasm

package comps

import (
	"syscall/js"
	"testing"

	"github.com/ozanturksever/logutil"
)

type fakeWrapper struct{}

func (fakeWrapper) JSValue() js.Value { return js.Global() }

func TestLog_NoPanicWithMixedValues(t *testing.T) {
	// If any of these panic, the test will fail.
	logutil.Log("hello", 42, 3.14, true)
	logutil.Log(js.Undefined(), js.Null(), js.Global())
	logutil.Log(map[string]int{"a": 1}, []string{"x", "y"})
	logutil.Log(struct{ X int }{X: 2})
	logutil.Log(fakeWrapper{})
}

func TestLogf_NoPanicWithFormat(t *testing.T) {
	logutil.Logf("value=%d, pi=%.2f, s=%s", 7, 3.14159, "ok")
}
