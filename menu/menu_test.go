package menu

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

type id int

const (
	idA id = iota + 1
	idB
	idC
)

var testStyle = Style{
	Inactive: lipgloss.NewStyle(),
	Hover:    lipgloss.NewStyle().Bold(true),
	Disabled: lipgloss.NewStyle().Faint(true),
	Border:   lipgloss.NewStyle(),
}

// open builds a typical 3-item menu at (10, 5), Hover on the first
// enabled item.
func open() Group[id] {
	return Group[id]{
		Items: []Item[id]{
			{ID: idA, Label: "Alpha"},
			{ID: idB, Label: "Beta"},
			{ID: idC, Label: "Gamma"},
		},
		Open:    true,
		Hover:   0,
		AnchorX: 10,
		AnchorY: 5,
		Style:   testStyle,
	}
}

// TestRenderClosed verifies the closed-menu contract: empty output.
func TestRenderClosed(t *testing.T) {
	g := open()
	g.Open = false
	if got := g.Render(); got != "" {
		t.Errorf("closed menu must Render to \"\", got %q", got)
	}
}

// TestRenderDimensions verifies Width/Height match the rendered
// output and that all rows are the same width.
func TestRenderDimensions(t *testing.T) {
	g := open()
	out := g.Render()
	rows := strings.Split(out, "\n")
	if len(rows) != g.Height() {
		t.Errorf("got %d rows, want %d", len(rows), g.Height())
	}
	want := g.Width()
	for i, r := range rows {
		if w := lipgloss.Width(r); w != want {
			t.Errorf("row %d: width %d, want %d (%q)", i, w, want, r)
		}
	}
}

// TestInside tests the rect-detection helper at edges and corners.
func TestInside(t *testing.T) {
	g := open()
	w, h := g.Width(), g.Height()
	cases := []struct {
		x, y int
		want bool
	}{
		{10, 5, true},                 // top-left corner
		{10 + w - 1, 5, true},         // top-right
		{10, 5 + h - 1, true},         // bottom-left
		{10 + w - 1, 5 + h - 1, true}, // bottom-right
		{9, 5, false},                 // one column left
		{10 + w, 5, false},            // one column past right
		{10, 4, false},                // one row above
		{10, 5 + h, false},            // one row past bottom
		{50, 50, false},               // far away
	}
	for _, tc := range cases {
		got := g.Inside(tc.x, tc.y)
		if got != tc.want {
			t.Errorf("Inside(%d,%d) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

// TestHitTest covers item rows, border rows, and out-of-bounds.
func TestHitTest(t *testing.T) {
	g := open() // 3 items at AnchorX=10, AnchorY=5
	cases := []struct {
		name string
		x, y int
		want id
		ok   bool
	}{
		{"top border", 12, 5, 0, false},
		{"item row 0 (Alpha) content", 12, 6, idA, true},
		{"item row 1 (Beta) content", 12, 7, idB, true},
		{"item row 2 (Gamma) content", 12, 8, idC, true},
		{"bottom border", 12, 5 + g.Height() - 1, 0, false},
		{"left side border", 10, 6, 0, false},
		{"right side border", 10 + g.Width() - 1, 6, 0, false},
		{"way outside", 0, 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := g.HitTest(tc.x, tc.y)
			if got != tc.want || ok != tc.ok {
				t.Errorf("HitTest(%d,%d) = (%v, %v), want (%v, %v)",
					tc.x, tc.y, got, ok, tc.want, tc.ok)
			}
		})
	}
}

// TestApplyKeyConfirm verifies enter on a hovered enabled item
// returns Confirmed=true, Chosen=its ID, and closes the menu.
func TestApplyKeyConfirm(t *testing.T) {
	g := open()
	g.Hover = 1 // Beta
	out := g.ApplyKey("enter")
	if !out.Confirmed || out.Chosen != idB {
		t.Errorf("enter on Beta: Confirmed=%v Chosen=%v, want true %v", out.Confirmed, out.Chosen, idB)
	}
	if out.Group.Open {
		t.Errorf("Open should be false after confirm")
	}
}

// TestApplyKeyCancel verifies esc closes the menu and sets Cancelled.
func TestApplyKeyCancel(t *testing.T) {
	out := open().ApplyKey("esc")
	if !out.Cancelled || out.Confirmed {
		t.Errorf("esc: Cancelled=%v Confirmed=%v, want true false", out.Cancelled, out.Confirmed)
	}
	if out.Group.Open {
		t.Errorf("Open should be false after esc")
	}
}

// TestApplyKeyNavigation verifies up/down move Hover; wrap behavior
// is honored.
func TestApplyKeyNavigation(t *testing.T) {
	g := open()
	g.Wrap = true

	// down from 0 → 1
	out := g.ApplyKey("down")
	if out.Group.Hover != 1 {
		t.Errorf("down from 0: Hover=%d, want 1", out.Group.Hover)
	}

	// down from 2 with wrap → 0
	g.Hover = 2
	out = g.ApplyKey("down")
	if out.Group.Hover != 0 {
		t.Errorf("down from 2 wrap: Hover=%d, want 0", out.Group.Hover)
	}

	// up from 0 with wrap → 2
	g.Hover = 0
	out = g.ApplyKey("up")
	if out.Group.Hover != 2 {
		t.Errorf("up from 0 wrap: Hover=%d, want 2", out.Group.Hover)
	}
}

// TestApplyKeyClosed verifies key events on a closed menu are no-op.
func TestApplyKeyClosed(t *testing.T) {
	g := open()
	g.Open = false
	out := g.ApplyKey("enter")
	if out.Confirmed || out.Cancelled {
		t.Errorf("closed menu must ignore keys; got Confirmed=%v Cancelled=%v",
			out.Confirmed, out.Cancelled)
	}
}

// TestHoverAt verifies mouse-motion-driven hover updates: cursor on
// an item sets Hover; cursor on border / outside leaves Hover alone.
func TestHoverAt(t *testing.T) {
	g := open() // 3 items, Hover=0
	g.Hover = 0

	// Mouse on Beta row → Hover=1
	g = g.HoverAt(12, 7)
	if g.Hover != 1 {
		t.Errorf("HoverAt on Beta: Hover=%d, want 1", g.Hover)
	}

	// Mouse on border → Hover stays at 1
	g = g.HoverAt(10, 5)
	if g.Hover != 1 {
		t.Errorf("HoverAt on border: Hover=%d, want unchanged 1", g.Hover)
	}

	// Mouse outside menu → Hover stays at 1
	g = g.HoverAt(0, 0)
	if g.Hover != 1 {
		t.Errorf("HoverAt outside: Hover=%d, want unchanged 1", g.Hover)
	}

	// Closed menu → no-op
	g.Open = false
	g.Hover = 5
	g = g.HoverAt(12, 7)
	if g.Hover != 5 {
		t.Errorf("HoverAt on closed menu should be no-op; Hover=%d", g.Hover)
	}
}

// TestApplyClickOnItem verifies clicking an item confirms it.
func TestApplyClickOnItem(t *testing.T) {
	g := open()
	out := g.ApplyClick(12, 7) // Beta row
	if !out.Confirmed || out.Chosen != idB {
		t.Errorf("click on Beta: Confirmed=%v Chosen=%v, want true %v",
			out.Confirmed, out.Chosen, idB)
	}
	if out.Group.Open {
		t.Errorf("Open should be false after item click")
	}
}

// TestApplyClickOutside verifies clicking outside the menu cancels.
func TestApplyClickOutside(t *testing.T) {
	out := open().ApplyClick(0, 0)
	if !out.Cancelled || out.Confirmed {
		t.Errorf("click outside: Cancelled=%v Confirmed=%v, want true false",
			out.Cancelled, out.Confirmed)
	}
	if out.Group.Open {
		t.Errorf("Open should be false after outside click")
	}
}

// TestApplyClickOnBorder verifies clicks on border/padding are no-op
// (menu stays open).
func TestApplyClickOnBorder(t *testing.T) {
	g := open()
	out := g.ApplyClick(10, 5) // top-left corner (border)
	if out.Confirmed || out.Cancelled {
		t.Errorf("border click: Confirmed=%v Cancelled=%v, want both false",
			out.Confirmed, out.Cancelled)
	}
	if !out.Group.Open {
		t.Errorf("Open should remain true after border click")
	}
}

// TestDisabledItemCannotConfirm verifies disabled items can't be
// confirmed via enter or click.
func TestDisabledItemCannotConfirm(t *testing.T) {
	g := open()
	g.Items[1].Disabled = true // Beta disabled
	g.Hover = 1

	// Enter on disabled hovered item → no-op
	out := g.ApplyKey("enter")
	if out.Confirmed {
		t.Errorf("enter on disabled item should not confirm")
	}
	if !out.Group.Open {
		t.Errorf("menu should stay open when confirming disabled item")
	}

	// Click on disabled item → no-op
	out = g.ApplyClick(12, 7) // Beta row
	if out.Confirmed {
		t.Errorf("click on disabled item should not confirm")
	}
	if !out.Group.Open {
		t.Errorf("menu should stay open when clicking disabled item")
	}
}

// TestNavigationSkipsDisabled verifies up/down skip disabled items.
func TestNavigationSkipsDisabled(t *testing.T) {
	g := open()
	g.Items[1].Disabled = true // Beta disabled
	g.Hover = 0
	g.Wrap = true

	// down from Alpha should skip Beta and land on Gamma
	out := g.ApplyKey("down")
	if out.Group.Hover != 2 {
		t.Errorf("down skipping disabled: Hover=%d, want 2", out.Group.Hover)
	}
}
