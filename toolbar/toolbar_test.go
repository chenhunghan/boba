package toolbar_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/toolbar"
)

func TestRenderShape(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"New", "Open", "Save"}, Hover: -1, Gap: 1}
	out := bar.Render()
	if got := lipgloss.Width(out); got != 19 {
		t.Fatalf("width = %d, want 19", got)
	}
	if want := " New   Open   Save "; out != want {
		t.Fatalf("render = %q, want %q", out, want)
	}
}

func TestRenderEmpty(t *testing.T) {
	var bar toolbar.Toolbar
	if out := bar.Render(); out != "" {
		t.Fatalf("empty render = %q, want \"\"", out)
	}
	if out := bar.View(); out != "" {
		t.Fatalf("empty view = %q, want \"\"", out)
	}
}

func TestViewMatchesRender(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B"}, Hover: -1}
	if bar.View() != bar.Render() {
		t.Fatal("View must equal Render")
	}
}

func TestHitTest(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"New", "Open", "Save"}, Hover: -1, Gap: 1}
	cases := []struct {
		x, y int
		want int
	}{
		{-1, 0, -1}, // before first item
		{0, 0, 0},   // left edge of "New"
		{4, 0, 0},   // right edge of "New"
		{5, 0, -1},  // gap between item 0 and 1
		{6, 0, 1},   // left edge of "Open"
		{12, 0, -1}, // gap between item 1 and 2
		{13, 0, 2},  // left edge of "Save"
		{18, 0, 2},  // right edge of "Save"
		{19, 0, -1}, // past the last item
		{0, 1, -1},  // below the row
		{0, -1, -1}, // above the row
	}
	for _, c := range cases {
		if got := bar.HitTest(c.x, c.y); got != c.want {
			t.Errorf("HitTest(%d, %d) = %d, want %d", c.x, c.y, got, c.want)
		}
	}
}

func TestHitTestNoGap(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"New", "Open"}, Hover: -1}
	// "New" = 0..4, "Open" = 5..10; no dead zone between.
	if got := bar.HitTest(4, 0); got != 0 {
		t.Errorf("HitTest(4,0) = %d, want 0", got)
	}
	if got := bar.HitTest(5, 0); got != 1 {
		t.Errorf("HitTest(5,0) = %d, want 1", got)
	}
}

func TestApplyKeyMovesSelected(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B", "C"}, Selected: 0, Hover: -1}

	bar, _, _ = bar.ApplyKey("right")
	if bar.Selected != 1 {
		t.Fatalf("after right Selected = %d, want 1", bar.Selected)
	}
	bar, _, _ = bar.ApplyKey("l")
	if bar.Selected != 2 {
		t.Fatalf("after l Selected = %d, want 2", bar.Selected)
	}
	// Clamp at the right edge (Wrap is false).
	bar, _, _ = bar.ApplyKey("right")
	if bar.Selected != 2 {
		t.Fatalf("clamped Selected = %d, want 2", bar.Selected)
	}
	bar, _, _ = bar.ApplyKey("left")
	if bar.Selected != 1 {
		t.Fatalf("after left Selected = %d, want 1", bar.Selected)
	}
	bar, _, _ = bar.ApplyKey("h")
	bar, _, _ = bar.ApplyKey("h")
	if bar.Selected != 0 {
		t.Fatalf("clamped-left Selected = %d, want 0", bar.Selected)
	}
}

func TestApplyKeyWrap(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B", "C"}, Selected: 0, Hover: -1, Wrap: true}
	bar, _, _ = bar.ApplyKey("left")
	if bar.Selected != 2 {
		t.Fatalf("wrap left Selected = %d, want 2", bar.Selected)
	}
	bar, _, _ = bar.ApplyKey("right")
	if bar.Selected != 0 {
		t.Fatalf("wrap right Selected = %d, want 0", bar.Selected)
	}
}

func TestApplyKeyConfirm(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B", "C"}, Selected: 1, Hover: -1}
	_, idx, fired := bar.ApplyKey("enter")
	if !fired || idx != 1 {
		t.Fatalf("confirm: fired=%v idx=%d, want true 1", fired, idx)
	}
	if _, _, fired := bar.ApplyKey("x"); fired {
		t.Fatal("unbound key must not fire")
	}
}

func TestApplyKeyCustomBindings(t *testing.T) {
	bar := toolbar.Toolbar{
		Items:    []string{"A", "B"},
		Hover:    -1,
		Right:    []string{"tab"},
		Left:     []string{"shift+tab"},
		Confirm:  []string{" "},
		Selected: 0,
	}
	bar, _, _ = bar.ApplyKey("tab")
	if bar.Selected != 1 {
		t.Fatalf("custom right Selected = %d, want 1", bar.Selected)
	}
	// Default "right" is overridden, so it should no longer move.
	before := bar.Selected
	bar, _, _ = bar.ApplyKey("right")
	if bar.Selected != before {
		t.Fatal("overridden default key must be inert")
	}
	if _, _, fired := bar.ApplyKey(" "); !fired {
		t.Fatal("custom confirm must fire")
	}
}

func TestApplyKeyEmpty(t *testing.T) {
	var bar toolbar.Toolbar
	got, idx, fired := bar.ApplyKey("enter")
	if fired || idx != 0 || len(got.Items) != 0 {
		t.Fatal("ApplyKey on empty toolbar must be inert")
	}
}

func TestUpdateIgnoredWhenUnfocused(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B"}, Selected: 0, Hover: -1}
	got, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got.Selected != 0 || cmd != nil {
		t.Fatal("Update must be a no-op when unfocused")
	}
}

func TestUpdateIgnoresNonKey(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B"}, Focused: true, Hover: -1}
	got, cmd := bar.Update(tea.MouseMsg{})
	if cmd != nil || got.Selected != 0 {
		t.Fatal("Update must ignore non-key messages")
	}
}

func TestUpdateMovesWhenFocused(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B", "C"}, Focused: true, Hover: -1}
	got, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got.Selected != 1 {
		t.Fatalf("focused Update right Selected = %d, want 1", got.Selected)
	}
	if cmd != nil {
		t.Fatal("movement must not emit a cmd")
	}
}

func TestUpdateConfirmEmitsActivated(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"A", "B", "C"}, Selected: 2, Focused: true, Hover: -1}
	_, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("confirm must emit a cmd")
	}
	msg, ok := cmd().(toolbar.ActivatedMsg)
	if !ok || msg.Index != 2 {
		t.Fatalf("ActivatedMsg = %#v, want Index 2", msg)
	}
}

func TestClickAtSelectsAndActivates(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"New", "Open", "Save"}, Hover: -1, Gap: 1}
	got, cmd := bar.ClickAt(6, 0) // inside "Open"
	if got.Selected != 1 {
		t.Fatalf("ClickAt Selected = %d, want 1", got.Selected)
	}
	if cmd == nil {
		t.Fatal("click hit must emit a cmd")
	}
	msg, ok := cmd().(toolbar.ActivatedMsg)
	if !ok || msg.Index != 1 {
		t.Fatalf("ActivatedMsg = %#v, want Index 1", msg)
	}
}

func TestClickAtMissIsNoOp(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"New", "Open"}, Selected: 0, Hover: -1, Gap: 1}
	got, cmd := bar.ClickAt(5, 0) // gap
	if cmd != nil || got.Selected != 0 {
		t.Fatal("click miss must be a no-op")
	}
}

func TestHoverAt(t *testing.T) {
	bar := toolbar.Toolbar{Items: []string{"New", "Open", "Save"}, Hover: -1, Gap: 1}
	bar = bar.HoverAt(13, 0) // inside "Save"
	if bar.Hover != 2 {
		t.Fatalf("HoverAt Hover = %d, want 2", bar.Hover)
	}
	bar = bar.HoverAt(5, 0) // gap clears hover
	if bar.Hover != -1 {
		t.Fatalf("HoverAt miss Hover = %d, want -1", bar.Hover)
	}
}
