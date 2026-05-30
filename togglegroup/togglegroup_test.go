package togglegroup_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/togglegroup"
)

func group() togglegroup.Group {
	return togglegroup.Group{
		Items:    []string{"Day", "Week", "Month"},
		Selected: 0,
		Focused:  true,
		Hover:    -1,
	}
}

func TestRenderPlainShape(t *testing.T) {
	g := group()
	if got, want := g.Render(), "DayWeekMonth"; got != want {
		t.Fatalf("Render()=%q, want %q", got, want)
	}
	g.Gap = 1
	if got, want := g.Render(), "Day Week Month"; got != want {
		t.Fatalf("Render() with gap=%q, want %q", got, want)
	}
}

func TestRenderEmpty(t *testing.T) {
	var g togglegroup.Group
	if g.Render() != "" {
		t.Fatal("an empty group should render nothing")
	}
}

func TestUpdateMovesAndEmits(t *testing.T) {
	g := group()
	g, cmd := g.Update(tea.KeyMsg{Type: tea.KeyRight})
	if g.Selected != 1 {
		t.Fatalf("right: Selected=%d, want 1", g.Selected)
	}
	if msg, ok := cmd().(togglegroup.ChangedMsg); !ok || msg.Selected != 1 {
		t.Fatalf("got %#v, want togglegroup.ChangedMsg{Selected: 1}", cmd())
	}

	g, cmd = g.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if g.Selected != 0 {
		t.Fatalf("left: Selected=%d, want 0", g.Selected)
	}
	if msg, ok := cmd().(togglegroup.ChangedMsg); !ok || msg.Selected != 0 {
		t.Fatalf("got %#v, want togglegroup.ChangedMsg{Selected: 0}", cmd())
	}
}

func TestUpdateClampsAtEnds(t *testing.T) {
	g := group()
	g2, cmd := g.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if g2.Selected != 0 || cmd != nil {
		t.Fatalf("left at index 0 should clamp and emit nothing: Selected=%d", g2.Selected)
	}

	g.Selected = 2
	g3, cmd := g.Update(tea.KeyMsg{Type: tea.KeyRight})
	if g3.Selected != 2 || cmd != nil {
		t.Fatalf("right at last index should clamp and emit nothing: Selected=%d", g3.Selected)
	}
}

func TestUpdateWraps(t *testing.T) {
	g := group()
	g.Wrap = true
	g2, cmd := g.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if g2.Selected != 2 {
		t.Fatalf("left at 0 with wrap: Selected=%d, want 2", g2.Selected)
	}
	if msg, ok := cmd().(togglegroup.ChangedMsg); !ok || msg.Selected != 2 {
		t.Fatalf("got %#v, want togglegroup.ChangedMsg{Selected: 2}", cmd())
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	g := group()
	g.Focused = false
	g2, cmd := g.Update(tea.KeyMsg{Type: tea.KeyRight})
	if g2.Selected != g.Selected || cmd != nil {
		t.Fatal("an unfocused group should ignore keys")
	}
}

func TestApplyKeyCustomBinding(t *testing.T) {
	g := group()
	g.Right = []string{"tab"}
	if g2, _ := g.ApplyKey("l"); g2.Selected != 0 {
		t.Fatal("custom binding should replace the default, not extend it")
	}
	if g2, _ := g.ApplyKey("tab"); g2.Selected != 1 {
		t.Fatal("custom binding should move Selected")
	}
}

func TestHitTestWithGap(t *testing.T) {
	g := group()
	g.Gap = 1 // "Day Week Month" → Day=0..2, gap=3, Week=4..7, gap=8, Month=9..13
	cases := []struct {
		x, want int
	}{
		{0, 0}, {2, 0}, {3, -1}, {4, 1}, {7, 1}, {8, -1}, {9, 2}, {13, 2}, {14, -1},
	}
	for _, c := range cases {
		if got := g.HitTest(c.x, 0); got != c.want {
			t.Errorf("HitTest(%d,0)=%d, want %d", c.x, got, c.want)
		}
	}
	if g.HitTest(0, 1) != -1 {
		t.Error("a hit off the row should be -1")
	}
}

func TestClickSelectsAndEmits(t *testing.T) {
	g := group()
	g2, cmd := g.ClickAt(5, 0) // inside "Week"
	if g2.Selected != 1 {
		t.Fatalf("click: Selected=%d, want 1", g2.Selected)
	}
	if msg, ok := cmd().(togglegroup.ChangedMsg); !ok || msg.Selected != 1 {
		t.Fatalf("got %#v, want togglegroup.ChangedMsg{Selected: 1}", cmd())
	}
}

func TestClickSameSelectionIsNoop(t *testing.T) {
	g := group()
	if _, cmd := g.ClickAt(0, 0); cmd != nil {
		t.Fatal("clicking the already-selected item should emit nothing")
	}
	if _, cmd := g.ClickAt(99, 0); cmd != nil {
		t.Fatal("a click off the row should be a no-op")
	}
}

func TestHoverAtSetsAndClears(t *testing.T) {
	g := group().HoverAt(5, 0)
	if g.Hover != 1 {
		t.Fatalf("Hover=%d, want 1", g.Hover)
	}
	if g = g.HoverAt(99, 0); g.Hover != -1 {
		t.Fatalf("Hover=%d, want -1", g.Hover)
	}
}
