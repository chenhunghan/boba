package tab_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/tab"
)

// recorder is a sub-model used to assert that Update forwards messages
// to the active tab.
type recorder struct{ last string }

func (r recorder) Init() tea.Cmd { return nil }
func (r recorder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		r.last = k.String()
	}
	return r, nil
}
func (r recorder) View() string { return r.last }

func tabGroup() tab.Group[string] {
	g := tab.Group[string]{}
	g, _ = g.AddTab(tab.Tab[string]{ID: "a", Label: "A", Model: recorder{}})
	g, _ = g.AddTab(tab.Tab[string]{ID: "b", Label: "B", Closable: true, Model: recorder{}})
	return g
}

func key(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestUpdateBoundKeyCycles(t *testing.T) {
	g := tabGroup()
	g.Selected = "a"
	g, _ = g.Update(key("]"))
	if g.Selected != "b" {
		t.Fatalf("Selected=%q, want \"b\"", g.Selected)
	}
}

func TestUpdateForwardsToActiveTab(t *testing.T) {
	g := tabGroup()
	g, _ = g.Update(key("x"))
	if got := g.RenderContent(); got != "x" {
		t.Fatalf("active tab content=%q, want \"x\" (key should forward)", got)
	}
}

func TestClickSelectsAndCloses(t *testing.T) {
	g := tabGroup()

	g, _ = g.ClickAt(2, 0)
	if g.Selected != "a" {
		t.Fatalf("after click: Selected=%q, want \"a\"", g.Selected)
	}

	g, _ = g.ClickAt(9, 0)
	if _, ok := g.Find("b"); ok {
		t.Fatal("tab \"b\" should be closed after clicking its close glyph")
	}
}

func TestHoverState(t *testing.T) {
	g := tabGroup()
	if rs := g.HoverState(2, 0); !rs.HasHover || rs.Hover != "a" {
		t.Fatalf("HoverState=%+v, want {Hover:a HasHover:true}", rs)
	}
	if rs := g.HoverState(0, 99); rs.HasHover {
		t.Fatalf("HoverState=%+v, want no hover", rs)
	}
}
