package statusbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestRenderWidth verifies the contract: Render produces exactly
// `width` visible cells, regardless of how much content fits.
func TestRenderWidth(t *testing.T) {
	cases := []struct {
		name  string
		bar   Bar
		width int
	}{
		{"empty", Bar{}, 40},
		{"left only", Bar{Left: []Item{{Key: "esc", Text: "back"}}}, 40},
		{"right only", Bar{Right: []Item{{Key: "q", Text: "quit"}}}, 40},
		{"both",
			Bar{
				Left:  []Item{{Key: "1/h", Text: "hot"}, {Key: "↑↓", Text: "select"}},
				Right: []Item{{Key: "esc", Text: "back"}, {Key: "q", Text: "quit"}},
			},
			80,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.bar.Render(tc.width)
			if got := lipgloss.Width(out); got != tc.width {
				t.Errorf("width = %d, want %d (output %q)", got, tc.width, out)
			}
			if strings.Contains(out, "\n") {
				t.Errorf("status bar must be one row, got newline: %q", out)
			}
		})
	}
}

// TestRenderOverflow verifies the documented overflow behavior: when
// the left group alone exceeds width, the bar emits the full left
// content and may end up wider than requested.
func TestRenderOverflow(t *testing.T) {
	bar := Bar{
		Left: []Item{
			{Key: "long", Text: "label"}, {Key: "long", Text: "label"},
		},
		Right: []Item{{Key: "x", Text: "x"}},
	}
	out := bar.Render(10)
	if lipgloss.Width(out) < 10 {
		t.Errorf("overflow output should not be shorter than width")
	}
	if strings.Contains(out, "\n") {
		t.Errorf("status bar must be one row, got newline")
	}
}

// TestItemRenderShape verifies the visible structure of a single Item.
// Letter highlighting must not change the visible width (it only
// re-styles characters that are already in Text).
func TestItemRenderShape(t *testing.T) {
	cases := []struct {
		name string
		item Item
		want string // visible text (after stripping ANSI)
	}{
		{"key+text", Item{Key: "esc", Text: "back"}, "esc back"},
		{"key only", Item{Key: "esc"}, "esc"},
		{"text only", Item{Text: "ready"}, "ready"},
		{"with letter", Item{Key: "1", Text: "hot", Letter: "h"}, "1 hot"},
		{"letter not in text", Item{Key: "1", Text: "hot", Letter: "z"}, "1 hot"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.item.render()
			width := lipgloss.Width(out)
			if width != lipgloss.Width(tc.want) {
				t.Errorf("visible width = %d, want %d", width, lipgloss.Width(tc.want))
			}
		})
	}
}
