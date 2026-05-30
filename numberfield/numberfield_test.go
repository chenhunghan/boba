package numberfield_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/input"
	"github.com/chenhunghan/boba/numberfield"
)

func runes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestStepUpClampsToMax(t *testing.T) {
	n := numberfield.Number{Value: 9, Min: 0, Max: 10, Step: 2, Focused: true}
	n, changed := n.ApplyKey(tea.KeyMsg{Type: tea.KeyUp})
	if !changed || n.Value != 10 {
		t.Fatalf("got %v changed=%v, want 10 true", n.Value, changed)
	}
	// Already at Max: stepping up is a no-op.
	n, changed = n.ApplyKey(tea.KeyMsg{Type: tea.KeyUp})
	if changed || n.Value != 10 {
		t.Fatalf("got %v changed=%v, want 10 false", n.Value, changed)
	}
}

func TestStepDownClampsToMin(t *testing.T) {
	n := numberfield.Number{Value: 1, Min: 0, Max: 10, Step: 5, Focused: true}
	n, changed := n.ApplyKey(tea.KeyMsg{Type: tea.KeyDown})
	if !changed || n.Value != 0 {
		t.Fatalf("got %v changed=%v, want 0 true", n.Value, changed)
	}
}

func TestStepDefaultsToOne(t *testing.T) {
	n := numberfield.Number{Value: 5, Focused: true}
	n, _ = n.ApplyKey(tea.KeyMsg{Type: tea.KeyUp})
	if n.Value != 6 {
		t.Fatalf("got %v, want 6 (Step defaults to 1)", n.Value)
	}
}

func TestStepSyncsInputText(t *testing.T) {
	n := numberfield.Number{Value: 2, Step: 1, Focused: true}
	n, _ = n.ApplyKey(tea.KeyMsg{Type: tea.KeyUp})
	if n.Input.Value != "3" {
		t.Fatalf("input text = %q, want %q", n.Input.Value, "3")
	}
}

func TestNoBoundsLeavesValueUnconstrained(t *testing.T) {
	// Max <= Min means no clamping.
	n := numberfield.Number{Value: 100, Step: 50, Focused: true}
	n, _ = n.ApplyKey(tea.KeyMsg{Type: tea.KeyUp})
	if n.Value != 150 {
		t.Fatalf("got %v, want 150", n.Value)
	}
}

func TestTypingUpdatesValue(t *testing.T) {
	n := numberfield.Number{Focused: true}
	for _, r := range []string{"4", "2"} {
		var changed bool
		n, changed = n.ApplyKey(runes(r))
		if !changed {
			t.Fatalf("typing %q should change Value", r)
		}
	}
	if n.Value != 42 {
		t.Fatalf("got %v, want 42", n.Value)
	}
}

func TestTypingClampsParsedValue(t *testing.T) {
	n := numberfield.Number{Min: 0, Max: 50, Focused: true}
	for _, r := range []string{"9", "9"} {
		n, _ = n.ApplyKey(runes(r))
	}
	if n.Value != 50 {
		t.Fatalf("got %v, want 50 (clamped)", n.Value)
	}
}

func TestPartialNumberIsNoChange(t *testing.T) {
	// A lone "-" doesn't parse: Value stays put, typing continues.
	n := numberfield.Number{Value: 7, Focused: true}
	n, changed := n.ApplyKey(runes("-"))
	if changed || n.Value != 7 {
		t.Fatalf("got %v changed=%v, want 7 false", n.Value, changed)
	}
	if n.Input.Value != "-" {
		t.Fatalf("input text = %q, want %q (kept for further typing)", n.Input.Value, "-")
	}
}

func TestCustomIncrementKey(t *testing.T) {
	n := numberfield.Number{Value: 0, Step: 1, Focused: true, Increment: []string{"up", "+"}}
	n, changed := n.ApplyKey(runes("+"))
	if !changed || n.Value != 1 {
		t.Fatalf("got %v changed=%v, want 1 true", n.Value, changed)
	}
}

func TestUpdateNoopWhenUnfocused(t *testing.T) {
	n := numberfield.Number{Value: 1, Step: 1}
	got, cmd := n.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got.Value != 1 || cmd != nil {
		t.Fatal("unfocused field must ignore keys")
	}
}

func TestUpdateIgnoresNonKeyMsg(t *testing.T) {
	n := numberfield.Number{Value: 1, Step: 1, Focused: true}
	if _, cmd := n.Update(struct{}{}); cmd != nil {
		t.Fatal("non-key message must be a no-op")
	}
}

func TestUpdateEmitsChangedMsg(t *testing.T) {
	n := numberfield.Number{Value: 1, Step: 1, Focused: true}
	_, cmd := n.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cmd == nil {
		t.Fatal("expected a cmd carrying ChangedMsg")
	}
	if msg, ok := cmd().(numberfield.ChangedMsg); !ok || msg.Value != 2 {
		t.Fatalf("got %#v, want numberfield.ChangedMsg{Value: 2}", cmd())
	}
}

func TestUpdateNoCmdWhenUnchanged(t *testing.T) {
	n := numberfield.Number{Value: 10, Min: 0, Max: 10, Step: 1, Focused: true}
	if _, cmd := n.Update(tea.KeyMsg{Type: tea.KeyUp}); cmd != nil {
		t.Fatal("a clamped step must not emit ChangedMsg")
	}
}

func TestHitTestZones(t *testing.T) {
	// Value "5" is 1 cell wide; up glyph at x=1, down glyph at x=2.
	n := numberfield.Number{Value: 5}.SetValue(5)
	for _, tc := range []struct {
		x    int
		want numberfield.Zone
	}{
		{0, numberfield.ZoneNone},
		{1, numberfield.ZoneUp},
		{2, numberfield.ZoneDown},
		{3, numberfield.ZoneNone},
	} {
		if got := n.HitTest(tc.x, 0); got != tc.want {
			t.Fatalf("HitTest(%d,0) = %v, want %v", tc.x, got, tc.want)
		}
	}
	if got := n.HitTest(1, 1); got != numberfield.ZoneNone {
		t.Fatalf("HitTest off-row = %v, want ZoneNone", got)
	}
}

func TestClickUpSteps(t *testing.T) {
	n := numberfield.Number{Step: 1}.SetValue(3)
	n, cmd := n.ClickAt(1, 0) // up glyph
	if n.Value != 4 {
		t.Fatalf("got %v, want 4", n.Value)
	}
	if msg, ok := cmd().(numberfield.ChangedMsg); !ok || msg.Value != 4 {
		t.Fatalf("got %#v, want numberfield.ChangedMsg{Value: 4}", cmd())
	}
}

func TestClickDownSteps(t *testing.T) {
	n := numberfield.Number{Step: 1}.SetValue(3)
	n, cmd := n.ClickAt(2, 0) // down glyph
	if n.Value != 2 || cmd == nil {
		t.Fatalf("got %v cmd=%v, want 2 + cmd", n.Value, cmd)
	}
}

func TestClickOnValueIsNoop(t *testing.T) {
	n := numberfield.Number{Step: 1}.SetValue(3)
	if _, cmd := n.ClickAt(0, 0); cmd != nil {
		t.Fatal("click on the value area should be a no-op")
	}
}

func TestClickClampedIsNoop(t *testing.T) {
	n := numberfield.Number{Min: 0, Max: 5, Step: 1}.SetValue(5)
	if _, cmd := n.ClickAt(1, 0); cmd != nil {
		t.Fatal("clicking up at Max should be a no-op")
	}
}

func TestSetValueClampsAndFormats(t *testing.T) {
	n := numberfield.Number{Min: 0, Max: 10}.SetValue(99)
	if n.Value != 10 {
		t.Fatalf("value = %v, want 10", n.Value)
	}
	if n.Input.Value != "10" {
		t.Fatalf("input text = %q, want %q", n.Input.Value, "10")
	}
}

func TestWidthCountsValueAndSteppers(t *testing.T) {
	// "12" (2) + up (1) + down (1) = 4.
	n := numberfield.Number{}.SetValue(12)
	if got := n.Width(); got != 4 {
		t.Fatalf("Width = %d, want 4", got)
	}
}

func TestRenderPlainValuePlusGlyphs(t *testing.T) {
	n := numberfield.Number{}.SetValue(7)
	if got := n.Render(); got != "7▲▼" {
		t.Fatalf("render = %q, want %q", got, "7▲▼")
	}
}

func TestCustomGlyphs(t *testing.T) {
	n := numberfield.Number{Style: numberfield.Style{Up: "+", Down: "-"}}.SetValue(7)
	if got := n.Render(); got != "7+-" {
		t.Fatalf("render = %q, want %q", got, "7+-")
	}
}

func TestViewMatchesRender(t *testing.T) {
	n := numberfield.Number{}.SetValue(3)
	if n.View() != n.Render() {
		t.Fatal("View must alias Render")
	}
}

func TestHoverAtIsIdentity(t *testing.T) {
	n := numberfield.Number{}.SetValue(3)
	got := n.HoverAt(1, 0)
	if got.Value != n.Value || got.Input.Value != n.Input.Value {
		t.Fatal("HoverAt must return the field unchanged")
	}
}

func TestFixedInputWidthGovernsZones(t *testing.T) {
	// A fixed input width reserves that many cells before the steppers,
	// regardless of the value text length.
	n := numberfield.Number{Input: input.Model{Width: 5}}.SetValue(2)
	if got := n.HitTest(5, 0); got != numberfield.ZoneUp {
		t.Fatalf("up zone at x=5 = %v, want ZoneUp", got)
	}
	if got := n.Width(); got != 7 {
		t.Fatalf("Width = %d, want 7 (5 + up + down)", got)
	}
}
