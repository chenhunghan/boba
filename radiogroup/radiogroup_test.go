package radiogroup_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/radiogroup"
)

func TestRenderShape(t *testing.T) {
	r := radiogroup.RadioGroup{Options: []string{"One", "Two", "Three"}, Selected: 1}
	out := r.Render()
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if !strings.HasPrefix(lines[1], "(*)") {
		t.Fatalf("selected row should use the selected glyph, got %q", lines[1])
	}
	if !strings.HasPrefix(lines[0], "( )") || !strings.HasPrefix(lines[2], "( )") {
		t.Fatalf("unselected rows should use the unselected glyph, got %q / %q", lines[0], lines[2])
	}
}

func TestUpdateMovesWhenFocused(t *testing.T) {
	r := radiogroup.RadioGroup{Options: []string{"a", "b", "c"}, Selected: 0, Focused: true}
	r, cmd := r.Update(tea.KeyMsg{Type: tea.KeyDown})
	if r.Selected != 1 {
		t.Fatalf("down should move Selected to 1, got %d", r.Selected)
	}
	if msg, ok := cmd().(radiogroup.ChangedMsg); !ok || msg.Selected != 1 {
		t.Fatalf("got %#v, want radiogroup.ChangedMsg{Selected: 1}", cmd())
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	r := radiogroup.RadioGroup{Options: []string{"a", "b"}, Selected: 0}
	r2, cmd := r.Update(tea.KeyMsg{Type: tea.KeyDown})
	if r2.Selected != 0 || cmd != nil {
		t.Fatal("an unfocused group should ignore keys")
	}
}

func TestUpdateClampsAtEdges(t *testing.T) {
	r := radiogroup.RadioGroup{Options: []string{"a", "b"}, Selected: 0, Focused: true}
	r, cmd := r.Update(tea.KeyMsg{Type: tea.KeyUp})
	if r.Selected != 0 {
		t.Fatalf("up at the top should stay at 0, got %d", r.Selected)
	}
	if cmd != nil {
		t.Fatal("an edge move that changes nothing should emit no cmd")
	}
}

func TestHitTest(t *testing.T) {
	r := radiogroup.RadioGroup{Options: []string{"a", "b", "c"}}
	if idx, ok := r.HitTest(0, 2); !ok || idx != 2 {
		t.Fatalf("HitTest(0,2) = (%d,%v), want (2,true)", idx, ok)
	}
	if _, ok := r.HitTest(0, 3); ok {
		t.Fatal("a y past the last row should miss")
	}
	if _, ok := r.HitTest(-1, 0); ok {
		t.Fatal("a negative x should miss")
	}
}

func TestClickSelects(t *testing.T) {
	r := radiogroup.RadioGroup{Options: []string{"a", "b", "c"}, Selected: 0}
	r, cmd := r.ClickAt(2, 2)
	if r.Selected != 2 {
		t.Fatalf("click on row 2 should select it, got %d", r.Selected)
	}
	if msg, ok := cmd().(radiogroup.ChangedMsg); !ok || msg.Selected != 2 {
		t.Fatalf("got %#v, want radiogroup.ChangedMsg{Selected: 2}", cmd())
	}
}

func TestClickSelectedRowIsNoop(t *testing.T) {
	r := radiogroup.RadioGroup{Options: []string{"a", "b"}, Selected: 1}
	if _, cmd := r.ClickAt(0, 1); cmd != nil {
		t.Fatal("clicking the already-selected row should be a no-op")
	}
}

func TestCustomKeysAndGlyphs(t *testing.T) {
	r := radiogroup.RadioGroup{
		Options:  []string{"a", "b"},
		Selected: 0,
		Focused:  true,
		Down:     []string{"tab"},
		Style:    radiogroup.Style{SelectedGlyph: "[x]", UnselectedGlyph: "[ ]"},
	}
	// "down" is no longer bound, so it must not move.
	if r2, cmd := r.Update(tea.KeyMsg{Type: tea.KeyDown}); r2.Selected != 0 || cmd != nil {
		t.Fatal("a key not in the custom binding should be ignored")
	}
	r, _ = r.Update(tea.KeyMsg{Type: tea.KeyTab})
	if r.Selected != 1 {
		t.Fatalf("custom Down key should move Selected, got %d", r.Selected)
	}
	if !strings.HasPrefix(strings.Split(r.Render(), "\n")[1], "[x]") {
		t.Fatal("custom selected glyph should render")
	}
}
