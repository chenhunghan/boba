package button

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// testStyle is a minimal Style used across tests. Concrete colors are
// irrelevant — tests check dimensions and structure, not appearance.
var testStyle = Style{
	Inactive: lipgloss.NewStyle().Background(lipgloss.Color("237")),
	Hover:    lipgloss.NewStyle().Background(lipgloss.Color("63")),
	Active:   lipgloss.NewStyle().Background(lipgloss.Color("33")),
}

// withStyle is a tiny helper for table-driven tests: assigns the
// shared testStyle to a Button so each test row reads cleanly.
func withStyle(b Button) Button {
	b.Style = testStyle
	return b
}

// TestButtonRenderDimensions verifies the visible width × height
// contract for various input combinations.
func TestButtonRenderDimensions(t *testing.T) {
	cases := []struct {
		name          string
		btn           Button
		state         State
		width, height int
		wantWidth     int
		wantHeight    int
	}{
		{"auto width 1-row", Button{Text: "X"}, StateInactive, 0, 0, 3, 1}, // " X " = 3 cells
		{"explicit 5×1", Button{Text: "X"}, StateHover, 5, 1, 5, 1},
		{"explicit 5×3", Button{Text: "X"}, StateActive, 5, 3, 5, 3},
		{"badge + text auto", Button{Text: "go", Badge: "1"}, StateInactive, 0, 0, 5, 1}, // " 1go "
		{"empty label auto", Button{}, StateInactive, 0, 0, 2, 1},                        // 2 padding only
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := withStyle(tc.btn).Render(tc.state, tc.width, tc.height)
			lines := strings.Split(out, "\n")
			if len(lines) != tc.wantHeight {
				t.Fatalf("got %d rows, want %d", len(lines), tc.wantHeight)
			}
			for i, line := range lines {
				if w := lipgloss.Width(line); w != tc.wantWidth {
					t.Errorf("row %d: visible width %d, want %d (%q)", i, w, tc.wantWidth, line)
				}
			}
		})
	}
}

// TestStackRender checks the rendered Stack is exactly Width × (n *
// ItemHeight) cells.
func TestStackRender(t *testing.T) {
	stack := Stack{
		Buttons: []Button{
			withStyle(Button{Text: "A"}),
			withStyle(Button{Text: "B"}),
			withStyle(Button{Text: "C"}),
		},
		Width:      5,
		ItemHeight: 2,
		Selected:   0,
		Hover:      -1,
		Active:     true,
	}
	out := stack.Render()
	lines := strings.Split(out, "\n")
	if len(lines) != 6 {
		t.Fatalf("got %d rows, want 6 (3 buttons × 2 rows)", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 5 {
			t.Errorf("row %d: visible width %d, want 5", i, w)
		}
	}
}

// TestStackHitTest checks the y → index mapping with various boundary
// conditions: negative y, y inside item, y past the last item.
func TestStackHitTest(t *testing.T) {
	stack := Stack{
		Buttons:    []Button{{Text: "A"}, {Text: "B"}, {Text: "C"}},
		Width:      5,
		ItemHeight: 2,
	}
	cases := []struct {
		y    int
		want int
	}{
		{-1, -1}, // before stack
		{0, 0},   // first cell of first button
		{1, 0},   // second cell of first button
		{2, 1},   // first cell of second button
		{3, 1},   // second cell of second button
		{4, 2},   // first cell of third button
		{5, 2},   // last cell of third button
		{6, -1},  // past the last button
		{99, -1},
	}
	for _, tc := range cases {
		// x is immaterial when no button has a Trailing.
		gotIdx, gotArea := stack.HitTest(0, tc.y)
		if gotIdx != tc.want {
			t.Errorf("HitTest(0, %d) idx = %d, want %d", tc.y, gotIdx, tc.want)
		}
		wantArea := HitBody
		if tc.want == -1 {
			wantArea = HitNone
		}
		if gotArea != wantArea {
			t.Errorf("HitTest(0, %d) area = %d, want %d", tc.y, gotArea, wantArea)
		}
	}
}

// TestStackHitTestTrailing verifies that clicks on the rightmost
// cell of a button with a non-empty Trailing return HitTrailing,
// while clicks elsewhere on the same button return HitBody. Buttons
// without a Trailing always return HitBody regardless of x.
func TestStackHitTestTrailing(t *testing.T) {
	stack := Stack{
		Buttons: []Button{
			{Text: "A", Trailing: "×"}, // has trailing
			{Text: "B"},                // no trailing
		},
		Width:      3,
		ItemHeight: 1,
	}
	cases := []struct {
		x, y     int
		wantIdx  int
		wantArea HitArea
	}{
		{0, 0, 0, HitBody},     // leftmost cell of button 0 → body
		{1, 0, 0, HitBody},     // middle cell of button 0 → body
		{2, 0, 0, HitTrailing}, // rightmost cell of button 0 → trailing (× lives here)
		{0, 1, 1, HitBody},     // leftmost cell of button 1 → body (no trailing)
		{2, 1, 1, HitBody},     // rightmost cell of button 1 → body (no trailing)
		{0, 2, -1, HitNone},    // past the last button
	}
	for _, tc := range cases {
		gotIdx, gotArea := stack.HitTest(tc.x, tc.y)
		if gotIdx != tc.wantIdx || gotArea != tc.wantArea {
			t.Errorf("HitTest(%d, %d) = (%d, %d), want (%d, %d)",
				tc.x, tc.y, gotIdx, gotArea, tc.wantIdx, tc.wantArea)
		}
	}
}

// TestHorizontalRowGap checks that gap cells appear between buttons.
func TestHorizontalRowGap(t *testing.T) {
	buttons := []Button{
		withStyle(Button{Text: "A"}),
		withStyle(Button{Text: "B"}),
	}
	out := HorizontalRow(buttons, -1, 1, 2)
	// Two buttons of auto-width 3 each, plus 2 gap cells = 8 total.
	if got := lipgloss.Width(out); got != 8 {
		t.Errorf("visible width %d, want 8 (two 3-wide buttons + 2 gap)", got)
	}
}
