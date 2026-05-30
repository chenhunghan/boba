package checkboxgroup_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/checkboxgroup"
)

func key(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestRenderShape(t *testing.T) {
	g := checkboxgroup.Group{
		Options: []string{"One", "Two", "Three"},
		Checked: []bool{false, true, false},
	}
	want := "[ ] One\n[x] Two\n[ ] Three"
	if got := g.Render(); got != want {
		t.Fatalf("Render()\n got %q\nwant %q", got, want)
	}
}

func TestRenderEmptyIsBlank(t *testing.T) {
	var g checkboxgroup.Group
	if got := g.Render(); got != "" {
		t.Fatalf("empty group should render %q, got %q", "", got)
	}
}

func TestRenderTreatsShortCheckedAsUnchecked(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"A", "B"}, Checked: []bool{true}}
	if got, lines := g.Render(), 2; strings.Count(got, "\n")+1 != lines {
		t.Fatalf("want %d lines, got %q", lines, got)
	}
	if !strings.HasPrefix(g.Render(), "[x] A") {
		t.Fatalf("row 0 should be checked: %q", g.Render())
	}
	if !strings.Contains(g.Render(), "[ ] B") {
		t.Fatalf("row 1 (past Checked) should read unchecked: %q", g.Render())
	}
}

func TestDownMovesCursorAndClamps(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"A", "B"}, Focused: true}
	g, _ = g.Update(key("j"))
	if g.Cursor != 1 {
		t.Fatalf("down should move cursor to 1, got %d", g.Cursor)
	}
	g, cmd := g.Update(tea.KeyMsg{Type: tea.KeyDown})
	if g.Cursor != 1 {
		t.Fatalf("down at the end should clamp at 1, got %d", g.Cursor)
	}
	if cmd != nil {
		t.Fatal("a navigation key should not emit a cmd")
	}
}

func TestUpClampsAtTop(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"A", "B"}, Cursor: 0, Focused: true}
	g, _ = g.Update(tea.KeyMsg{Type: tea.KeyUp})
	if g.Cursor != 0 {
		t.Fatalf("up at the top should clamp at 0, got %d", g.Cursor)
	}
}

func TestSpaceTogglesCursorRow(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"A", "B", "C"}, Cursor: 1, Focused: true}
	g, cmd := g.Update(key(" "))
	if !g.Checked[1] {
		t.Fatalf("space should check the cursor row, got %v", g.Checked)
	}
	if g.Checked[0] || g.Checked[2] {
		t.Fatalf("space should not touch other rows, got %v", g.Checked)
	}
	msg, ok := cmd().(checkboxgroup.ChangedMsg)
	if !ok || !msg.Checked[1] {
		t.Fatalf("got %#v, want ChangedMsg with row 1 checked", cmd())
	}
}

func TestChangedMsgIsACopy(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"A"}, Focused: true}
	g, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd().(checkboxgroup.ChangedMsg)
	msg.Checked[0] = false
	if !g.Checked[0] {
		t.Fatal("mutating the emitted slice must not affect the group state")
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"A"}}
	g2, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || g2.Cursor != 0 || len(g2.Checked) != 0 {
		t.Fatal("an unfocused group should ignore keys")
	}
}

func TestUpdateIgnoresNonKeyMsg(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"A"}, Focused: true}
	if _, cmd := g.Update(tea.MouseMsg{}); cmd != nil {
		t.Fatal("non-key messages should be ignored")
	}
}

func TestHitTest(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"One", "Two"}}
	if got := g.HitTest(0, 1); got != 1 {
		t.Fatalf("hit on row 1 should return 1, got %d", got)
	}
	if got := g.HitTest(0, 2); got != -1 {
		t.Fatalf("hit below the list should return -1, got %d", got)
	}
	if got := g.HitTest(99, 0); got != -1 {
		t.Fatalf("hit past the row width should return -1, got %d", got)
	}
}

func TestClickTogglesAndMovesCursor(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"One", "Two"}, Cursor: 0}
	g, cmd := g.ClickAt(1, 1)
	if g.Cursor != 1 {
		t.Fatalf("click should move cursor to row 1, got %d", g.Cursor)
	}
	if !g.Checked[1] {
		t.Fatalf("click should check row 1, got %v", g.Checked)
	}
	if msg, ok := cmd().(checkboxgroup.ChangedMsg); !ok || !msg.Checked[1] {
		t.Fatalf("got %#v, want ChangedMsg with row 1 checked", cmd())
	}
}

func TestClickMissIsNoop(t *testing.T) {
	g := checkboxgroup.Group{Options: []string{"One"}}
	if _, cmd := g.ClickAt(0, 5); cmd != nil {
		t.Fatal("a click off the list should be a no-op")
	}
}

func TestCustomGlyphsAndKeys(t *testing.T) {
	g := checkboxgroup.Group{
		Options: []string{"A"},
		Focused: true,
		Style:   checkboxgroup.Style{Checked: "(*)", Unchecked: "( )"},
		Toggle:  []string{"x"},
	}
	if !strings.HasPrefix(g.Render(), "( ) A") {
		t.Fatalf("custom unchecked glyph not used: %q", g.Render())
	}
	if _, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("enter should not toggle once Toggle is rebound to x")
	}
	g, cmd := g.Update(key("x"))
	if !g.Checked[0] || cmd == nil {
		t.Fatalf("custom toggle key x should check row 0, got %v", g.Checked)
	}
	if !strings.HasPrefix(g.Render(), "(*) A") {
		t.Fatalf("custom checked glyph not used: %q", g.Render())
	}
}
