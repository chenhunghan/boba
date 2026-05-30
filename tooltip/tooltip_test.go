package tooltip_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/popup"
	"github.com/chenhunghan/boba/tooltip"
)

func shown() tooltip.Tooltip {
	return tooltip.Tooltip{
		Content:   "rename",
		Visible:   true,
		X:         5,
		Y:         5,
		W:         4,
		H:         1,
		Placement: popup.Below,
		Style:     tooltip.Style{Surface: lipgloss.NewStyle()},
	}
}

func TestHiddenRendersEmpty(t *testing.T) {
	tip := shown()
	tip.Visible = false
	if got := tip.Render(); got != "" {
		t.Fatalf("hidden Render = %q, want empty", got)
	}
	if w, h := tip.Width(), tip.Height(); w != 0 || h != 0 {
		t.Fatalf("hidden size = (%d,%d), want (0,0)", w, h)
	}
}

func TestRenderDimensions(t *testing.T) {
	tip := shown()
	out := tip.Render()
	rows := strings.Split(out, "\n")
	if len(rows) != tip.Height() {
		t.Fatalf("got %d rows, want Height %d", len(rows), tip.Height())
	}
	// "rename" is 6 cells + 2 borders = 8 wide, 1 content row + 2 borders = 3 tall.
	if tip.Width() != 8 || tip.Height() != 3 {
		t.Fatalf("size = (%d,%d), want (8,3)", tip.Width(), tip.Height())
	}
	for i, r := range rows {
		if w := lipgloss.Width(r); w != tip.Width() {
			t.Fatalf("row %d width %d, want %d (%q)", i, w, tip.Width(), r)
		}
	}
}

func TestMultiLineHeightAndWidth(t *testing.T) {
	tip := shown()
	tip.Content = "ab\nlonger" // widest line is 6 cells
	if tip.Height() != 4 {
		t.Fatalf("two-line Height = %d, want 4", tip.Height())
	}
	if tip.Width() != 8 {
		t.Fatalf("Width = %d, want 8 (6 + 2 borders)", tip.Width())
	}
	rows := strings.Split(tip.Render(), "\n")
	for i, r := range rows {
		if w := lipgloss.Width(r); w != tip.Width() {
			t.Fatalf("row %d width %d, want %d", i, w, tip.Width())
		}
	}
}

func TestCustomGlyphsAppear(t *testing.T) {
	tip := shown()
	tip.Content = "x"
	tip.Style.TopLeft = "+"
	tip.Style.Horizontal = "="
	out := tip.Render()
	if !strings.Contains(out, "+") || !strings.Contains(out, "=") {
		t.Fatalf("custom glyphs missing from %q", out)
	}
}

func TestRowsBeginWithReset(t *testing.T) {
	const reset = "\x1b[0m"
	for _, r := range strings.Split(shown().Render(), "\n") {
		if !strings.HasPrefix(r, reset) {
			t.Fatalf("row not isolated by leading SGR reset: %q", r)
		}
	}
}

func TestOverHiddenReturnsBackgroundUnchanged(t *testing.T) {
	bg := "....\n....\n...."
	tip := shown()
	tip.Visible = false
	if got := tip.Over(bg, 20, 10); got != bg {
		t.Fatalf("hidden Over changed bg: %q", got)
	}
}

func TestOverPreservesBackgroundLineCount(t *testing.T) {
	bg := strings.Repeat(".", 20)
	bg = strings.Join([]string{bg, bg, bg, bg, bg, bg, bg, bg, bg, bg}, "\n")
	got := shown().Over(bg, 20, 10)
	if a, b := strings.Count(got, "\n"), strings.Count(bg, "\n"); a != b {
		t.Fatalf("Over changed line count: got %d, want %d", a, b)
	}
}

func TestViewMatchesRender(t *testing.T) {
	tip := shown()
	if tip.View() != tip.Render() {
		t.Fatal("View must alias Render")
	}
}
