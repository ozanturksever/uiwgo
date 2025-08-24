//go:build js && wasm

package dom

import (
	"testing"
	"time"

	reactivity "github.com/ozanturksever/uiwgo/reactivity"
)

func TestElementBuilderScopeCreation(t *testing.T) {
	eb := NewElement("div")
	if eb.scope == nil {
		t.Error("ElementBuilder should have a cleanup scope")
	}
	if eb.scope.GetParent() != nil {
		t.Error("Root element should have no parent scope")
	}
}

func TestElementBuilderScopeWithParent(t *testing.T) {
	parentScope := reactivity.NewCleanupScope(nil)
	reactivity.SetCurrentCleanupScope(parentScope)
	
	eb := NewElement("div")
	if eb.scope == nil {
		t.Error("ElementBuilder should have a cleanup scope")
	}
	if eb.scope.GetParent() != parentScope {
		t.Error("Element scope should have correct parent")
	}
	
	// Clean up
	reactivity.SetCurrentCleanupScope(nil)
	parentScope.Dispose()
}

func TestElementBuilderReactiveTextCleanup(t *testing.T) {
	signal := reactivity.CreateSignal("initial")
	eb := NewElement("div")
	
	eb.BindReactiveText(func() string {
		return signal.Get()
	})
	
	element := eb.Build()
	if element.TextContent() != "initial" {
		t.Errorf("Expected 'initial', got '%s'", element.TextContent())
	}
	
	// Update signal
	signal.Set("updated")
	if element.TextContent() != "updated" {
		t.Errorf("Expected 'updated', got '%s'", element.TextContent())
	}
	
	// Dispose scope should stop reactivity
	eb.scope.Dispose()
	signal.Set("should not update")
	
	// Give a moment for any potential updates
	time.Sleep(10 * time.Millisecond)
	
	if element.TextContent() != "updated" {
		t.Errorf("Text should not update after scope disposal, got '%s'", element.TextContent())
	}
}

func TestElementBuilderReactiveHTMLCleanup(t *testing.T) {
	signal := reactivity.CreateSignal("<span>initial</span>")
	eb := NewElement("div")
	
	eb.BindReactiveHTML(func() string {
		return signal.Get()
	})
	
	element := eb.Build()
	if element.InnerHTML() != "<span>initial</span>" {
		t.Errorf("Expected '<span>initial</span>', got '%s'", element.InnerHTML())
	}
	
	// Update signal
	signal.Set("<span>updated</span>")
	if element.InnerHTML() != "<span>updated</span>" {
		t.Errorf("Expected '<span>updated</span>', got '%s'", element.InnerHTML())
	}
	
	// Dispose scope should stop reactivity
	eb.scope.Dispose()
	signal.Set("<span>should not update</span>")
	
	// Give a moment for any potential updates
	time.Sleep(10 * time.Millisecond)
	
	if element.InnerHTML() != "<span>updated</span>" {
		t.Errorf("HTML should not update after scope disposal, got '%s'", element.InnerHTML())
	}
}

func TestElementBuilderReactiveAttributeCleanup(t *testing.T) {
	signal := reactivity.CreateSignal("initial-class")
	eb := NewElement("div")
	
	eb.BindReactiveAttribute("class", func() string {
		return signal.Get()
	})
	
	element := eb.Build()
	if element.GetAttribute("class") != "initial-class" {
		t.Errorf("Expected 'initial-class', got '%s'", element.GetAttribute("class"))
	}
	
	// Update signal
	signal.Set("updated-class")
	if element.GetAttribute("class") != "updated-class" {
		t.Errorf("Expected 'updated-class', got '%s'", element.GetAttribute("class"))
	}
	
	// Dispose scope should stop reactivity
	eb.scope.Dispose()
	signal.Set("should-not-update")
	
	// Give a moment for any potential updates
	time.Sleep(10 * time.Millisecond)
	
	if element.GetAttribute("class") != "updated-class" {
		t.Errorf("Attribute should not update after scope disposal, got '%s'", element.GetAttribute("class"))
	}
}

func TestElementBuilderBuildWithCleanup(t *testing.T) {
	signal := reactivity.CreateSignal("initial")
	eb := NewElement("div")
	
	eb.BindReactiveText(func() string {
		return signal.Get()
	})
	
	element, cleanup := eb.BuildWithCleanup()
	if element.TextContent() != "initial" {
		t.Errorf("Expected 'initial', got '%s'", element.TextContent())
	}
	
	// Update signal
	signal.Set("updated")
	if element.TextContent() != "updated" {
		t.Errorf("Expected 'updated', got '%s'", element.TextContent())
	}
	
	// Call cleanup function
	cleanup()
	signal.Set("should not update")
	
	// Give a moment for any potential updates
	time.Sleep(10 * time.Millisecond)
	
	if element.TextContent() != "updated" {
		t.Errorf("Text should not update after cleanup, got '%s'", element.TextContent())
	}
}

func TestElementBuilderNestedScopes(t *testing.T) {
	parentScope := reactivity.NewCleanupScope(nil)
	reactivity.SetCurrentCleanupScope(parentScope)
	
	parentSignal := reactivity.CreateSignal("parent")
	childSignal := reactivity.CreateSignal("child")
	
	parentEB := NewElement("div")
	parentEB.BindReactiveText(func() string {
		return parentSignal.Get()
	})
	
	// Create child element within parent's scope
	reactivity.SetCurrentCleanupScope(parentEB.scope)
	childEB := NewElement("span")
	childEB.BindReactiveText(func() string {
		return childSignal.Get()
	})
	
	parentElement := parentEB.Build()
	childElement := childEB.Build()
	
	if parentElement.TextContent() != "parent" {
		t.Errorf("Expected 'parent', got '%s'", parentElement.TextContent())
	}
	if childElement.TextContent() != "child" {
		t.Errorf("Expected 'child', got '%s'", childElement.TextContent())
	}
	
	// Update both signals
	parentSignal.Set("parent-updated")
	childSignal.Set("child-updated")
	
	if parentElement.TextContent() != "parent-updated" {
		t.Errorf("Expected 'parent-updated', got '%s'", parentElement.TextContent())
	}
	if childElement.TextContent() != "child-updated" {
		t.Errorf("Expected 'child-updated', got '%s'", childElement.TextContent())
	}
	
	// Dispose parent scope should dispose both parent and child
	parentScope.Dispose()
	parentSignal.Set("parent-should-not-update")
	childSignal.Set("child-should-not-update")
	
	// Give a moment for any potential updates
	time.Sleep(10 * time.Millisecond)
	
	if parentElement.TextContent() != "parent-updated" {
		t.Errorf("Parent text should not update after scope disposal, got '%s'", parentElement.TextContent())
	}
	if childElement.TextContent() != "child-updated" {
		t.Errorf("Child text should not update after scope disposal, got '%s'", childElement.TextContent())
	}
	
	// Clean up
	reactivity.SetCurrentCleanupScope(nil)
}