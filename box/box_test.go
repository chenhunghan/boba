package box

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/glyph"
)

// TestRenderDimensions verifies that Box.Render always emits exactly
// height lines, each width visible cells wide. This is the contract
// callers rely on when composing boxes with lipgloss.JoinVertical /
// JoinHorizontal — a single off-by-one breaks the whole layout.
//
// NOTE: width must be at least the sum of (notch widths + gaps) + 2
// for the box's own corners. If width is smaller than that, Render
// currently overflows. TODO: drop or truncate notches to honor the
// width contract gracefully.
func TestRenderDimensions(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
	}{
		{"small", 40, 5},
		{"square", 30, 30},
		{"wide", 200, 10},
		{"tall", 30, 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := Box{
				LeftNotches: []Notch{
					{Badge: glyph.Superscript(1), Text: "cpu", Gap: 2},
					{Text: "menu", Gap: 1},
				},
				RightNotches: []Notch{{Text: "now", Gap: 1}},
				Body:         "hello",
			}
			out := b.Render(tc.width, tc.height)

			lines := strings.Split(out, "\n")
			if len(lines) != tc.height {
				t.Fatalf("got %d lines, want %d", len(lines), tc.height)
			}
			for i, line := range lines {
				if w := lipgloss.Width(line); w != tc.width {
					t.Errorf("line %d: visible width %d, want %d (%q)", i, w, tc.width, line)
				}
			}
		})
	}
}

// TestRenderEmptyNotches verifies that a box with no notches still meets
// the dimension contract for the smallest reasonable size.
func TestRenderEmptyNotches(t *testing.T) {
	b := Box{Body: ""}
	out := b.Render(4, 3)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 4 {
			t.Errorf("line %d: visible width %d, want 4", i, w)
		}
	}
}

// TestRenderTooSmall verifies the empty-string contract for sub-minimum
// dimensions — callers can detect "box won't fit" via len(out) == 0.
func TestRenderTooSmall(t *testing.T) {
	b := Box{LeftNotches: []Notch{{Text: "x"}}}
	if got := b.Render(3, 5); got != "" {
		t.Errorf("width=3 should return empty, got %q", got)
	}
	if got := b.Render(20, 2); got != "" {
		t.Errorf("height=2 should return empty, got %q", got)
	}
}

// TestRenderEmptyBody verifies the body-padding fix: short or empty
// content must still produce innerHeight (= height-2) interior rows.
func TestRenderEmptyBody(t *testing.T) {
	b := Box{LeftNotches: []Notch{{Text: "x", Gap: 1}}}
	out := b.Render(20, 10)
	lines := strings.Split(out, "\n")
	if len(lines) != 10 {
		t.Fatalf("empty body: got %d lines, want 10", len(lines))
	}
}

// TestRenderBorderless verifies that Borderless mode produces the
// expected dimensions with no box-drawing characters in the output.
func TestRenderBorderless(t *testing.T) {
	b := Box{Body: "X", Borderless: true}
	out := b.Render(5, 3)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 5 {
			t.Errorf("line %d: visible width %d, want 5", i, w)
		}
	}
	// No box-drawing runes should appear in borderless output.
	for _, ch := range []string{"┌", "┐", "└", "┘", "│", "─"} {
		if strings.Contains(out, ch) {
			t.Errorf("borderless output contains %q", ch)
		}
	}
}

// TestRenderFillColor verifies that FillColor doesn't break the
// dimension contract — content is preserved at the expected size.
// (Whether ANSI escape codes appear depends on lipgloss's color
// detection, which strips colors in non-TTY test environments.)
func TestRenderFillColor(t *testing.T) {
	b := Box{
		Body:       "X",
		Borderless: true,
		FillColor:  lipgloss.Color("63"),
	}
	out := b.Render(5, 3)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 5 {
			t.Errorf("line %d: visible width %d, want 5", i, w)
		}
	}
	if !strings.Contains(out, "X") {
		t.Error("body content X not found in output")
	}
}
