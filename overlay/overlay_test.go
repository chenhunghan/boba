package overlay

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestOverlayEmptyFg verifies that an empty foreground returns the
// background unchanged — the no-op contract callers rely on when
// their popup is closed.
func TestOverlayEmptyFg(t *testing.T) {
	bg := "hello\nworld"
	if got := Overlay(bg, "", 1, 1); got != bg {
		t.Errorf("Overlay with empty fg should return bg unchanged; got %q want %q", got, bg)
	}
}

// TestOverlayBasicPlacement verifies a single-line fg lands at the
// requested column on the requested row, with bg's left and right
// portions preserved.
func TestOverlayBasicPlacement(t *testing.T) {
	bg := "abcdefghij\n0123456789"
	fg := "XYZ"
	got := Overlay(bg, fg, 3, 0)
	want := "abcXYZghij\n0123456789"
	if got != want {
		t.Errorf("Overlay = %q, want %q", got, want)
	}
}

// TestOverlayMultiLineFg verifies multi-line fg replaces the
// matching span on each row, leaving rows above and below untouched.
func TestOverlayMultiLineFg(t *testing.T) {
	bg := "row0row0row0\nrow1row1row1\nrow2row2row2"
	fg := "AA\nBB"
	got := Overlay(bg, fg, 4, 1)
	// At x=4 with 2-cell fg, prefix consumes "row1", drops 2 cells of
	// rest ("ro"), suffix is the remaining "w1row1".
	want := "row0row0row0\nrow1AAw1row1\nrow2BBw2row2"
	if got != want {
		t.Errorf("Overlay = %q, want %q", got, want)
	}
}

// TestOverlayPastBgWidth verifies that when x exceeds the bg row's
// visible width, the row is space-padded to reach x.
func TestOverlayPastBgWidth(t *testing.T) {
	bg := "abc"
	fg := "XYZ"
	got := Overlay(bg, fg, 5, 0)
	want := "abc  XYZ"
	if got != want {
		t.Errorf("Overlay = %q, want %q", got, want)
	}
}

// TestOverlayRowOutOfRange verifies fg rows that fall above or below
// bg's row range are dropped, not synthesized.
func TestOverlayRowOutOfRange(t *testing.T) {
	bg := "row0\nrow1"
	fg := "AA\nBB\nCC"
	got := Overlay(bg, fg, 0, 1)
	// fg row 0 lands on bg row 1; fg row 1 would land on bg row 2 (out of range, dropped);
	// fg row 2 would land on bg row 3 (out of range, dropped).
	want := "row0\nAA"
	// The row 1 result: x=0, fg=AA; bg "row1" -> prefix "" + fg "AA" + suffix "w1" → "AAw1"
	want = "row0\nAAw1"
	if got != want {
		t.Errorf("Overlay = %q, want %q", got, want)
	}
}

// TestOverlayNegativeY verifies that fg rows mapping to negative
// row indices (y < 0) are dropped silently.
func TestOverlayNegativeY(t *testing.T) {
	bg := "row0\nrow1"
	fg := "AAAA\nBBBB"
	// y=-1: fg row 0 → bg row -1 (drop); fg row 1 → bg row 0 (place).
	got := Overlay(bg, fg, 0, -1)
	want := "BBBB\nrow1"
	if got != want {
		t.Errorf("Overlay = %q, want %q", got, want)
	}
}

// TestOverlayPreservesWidth verifies that the resulting row's
// visible width matches the bg's pre-overlay width when x + fgW
// fits within bg's row.
func TestOverlayPreservesWidth(t *testing.T) {
	bg := "0123456789" // 10 cells
	fg := "ABC"        // 3 cells
	got := Overlay(bg, fg, 4, 0)
	if w := lipgloss.Width(got); w != 10 {
		t.Errorf("post-overlay width = %d, want 10 (got %q)", w, got)
	}
}

// TestSplitAtColPlain verifies the split helper on plain ASCII at
// various column boundaries.
func TestSplitAtColPlain(t *testing.T) {
	cases := []struct {
		s            string
		col          int
		wantL, wantR string
	}{
		{"abcdef", 0, "", "abcdef"},
		{"abcdef", 3, "abc", "def"},
		{"abcdef", 6, "abcdef", ""},
		{"abcdef", 99, "abcdef", ""},
	}
	for _, tc := range cases {
		gotL, gotR := splitAtCol(tc.s, tc.col)
		if gotL != tc.wantL || gotR != tc.wantR {
			t.Errorf("splitAtCol(%q, %d) = (%q, %q), want (%q, %q)",
				tc.s, tc.col, gotL, gotR, tc.wantL, tc.wantR)
		}
	}
}

// TestSplitAtColPassesThroughCSI verifies that CSI escape sequences
// are passed through to the left half (their pass-through behavior is
// what makes Overlay style-blind — see the package doc).
func TestSplitAtColPassesThroughCSI(t *testing.T) {
	// "AB" styled with red fg: \x1b[31mAB\x1b[0m
	// Cut at col=1 should put the open SGR + 'A' in left, 'B' + reset in right.
	s := "\x1b[31mAB\x1b[0m"
	left, right := splitAtCol(s, 1)
	if !strings.Contains(left, "\x1b[31m") || !strings.Contains(left, "A") {
		t.Errorf("left should contain open SGR and A: got %q", left)
	}
	if !strings.Contains(right, "B") {
		t.Errorf("right should contain B: got %q", right)
	}
}

// TestOverlayAnsiAroundFg verifies the styled portions of bg
// outside the overlay region survive the splice. Foreground here is
// plain (no ANSI); we only test that bg's color is preserved on
// either side of the cut.
func TestOverlayAnsiAroundFg(t *testing.T) {
	// Build a styled bg row: red "0123456789" (10 cells).
	bg := "\x1b[31m0123456789\x1b[0m"
	fg := "AB"
	// Place AB at col 4: prefix should retain the red SGR + "0123",
	// suffix should retain "6789" + reset (the reset from bg may have
	// been emitted into the dropped middle, but the visible cells
	// after the splice still appear).
	got := Overlay(bg, fg, 4, 0)
	if !strings.Contains(got, "\x1b[31m") {
		t.Errorf("Overlay should preserve bg's SGR: got %q", got)
	}
	if !strings.Contains(got, "0123") {
		t.Errorf("Overlay should preserve bg's prefix: got %q", got)
	}
	if !strings.Contains(got, "AB") {
		t.Errorf("Overlay should contain fg: got %q", got)
	}
	if !strings.Contains(got, "6789") {
		t.Errorf("Overlay should preserve bg's suffix: got %q", got)
	}
}
