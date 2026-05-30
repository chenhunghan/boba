package collapsible_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/collapsible"
)

func TestCollapsedRendersHeaderOnly(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs", Body: "line one\nline two"}
	got := c.Render()
	if got != "> Logs" {
		t.Fatalf("Render = %q, want \"> Logs\"", got)
	}
	if n := strings.Count(got, "\n") + 1; n != 1 {
		t.Fatalf("collapsed rows = %d, want 1", n)
	}
}

func TestExpandedRendersBody(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs", Body: "line one\nline two", Expanded: true}
	got := c.Render()
	want := "v Logs\nline one\nline two"
	if got != want {
		t.Fatalf("Render = %q, want %q", got, want)
	}
	if c.Lines() != 3 {
		t.Fatalf("Lines = %d, want 3", c.Lines())
	}
}

func TestExpandedWithEmptyBodyIsHeaderOnly(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs", Expanded: true}
	if got := c.Render(); got != "v Logs" {
		t.Fatalf("Render = %q, want \"v Logs\"", got)
	}
	if c.Lines() != 1 {
		t.Fatalf("Lines = %d, want 1", c.Lines())
	}
}

func TestCustomGlyphs(t *testing.T) {
	c := collapsible.Collapsible{
		Title: "X",
		Style: collapsible.Style{OpenGlyph: "▼", ClosedGlyph: "▶"},
	}
	if got := c.Render(); got != "▶ X" {
		t.Fatalf("collapsed Render = %q, want \"▶ X\"", got)
	}
	c.Expanded = true
	if got := c.Render(); got != "▼ X" {
		t.Fatalf("expanded Render = %q, want \"▼ X\"", got)
	}
}

func TestUpdateTogglesWhenFocused(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs", Focused: true}
	c, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !c.Expanded {
		t.Fatal("enter should expand a focused section")
	}
	if msg, ok := cmd().(collapsible.ToggledMsg); !ok || !msg.Expanded {
		t.Fatalf("got %#v, want collapsible.ToggledMsg{Expanded: true}", cmd())
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs"}
	c2, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if c2.Expanded || cmd != nil {
		t.Fatal("an unfocused section should ignore keys")
	}
}

func TestUpdateIgnoresUnboundKey(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs", Focused: true}
	c2, cmd := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if c2.Expanded || cmd != nil {
		t.Fatal("a key not in Toggle should be a no-op")
	}
}

func TestClickOnHeaderToggles(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs", Expanded: true}
	c, cmd := c.ClickAt(1, 0)
	if c.Expanded {
		t.Fatal("click on header should collapse")
	}
	if msg, ok := cmd().(collapsible.ToggledMsg); !ok || msg.Expanded {
		t.Fatalf("got %#v, want collapsible.ToggledMsg{Expanded: false}", cmd())
	}
}

func TestClickOffHeaderIsNoop(t *testing.T) {
	c := collapsible.Collapsible{Title: "Logs", Body: "body", Expanded: true}
	if _, cmd := c.ClickAt(0, 1); cmd != nil {
		t.Fatal("a click on a body row should be a no-op")
	}
	if _, cmd := c.ClickAt(c.HeaderWidth(), 0); cmd != nil {
		t.Fatal("a click past the header width should be a no-op")
	}
}
