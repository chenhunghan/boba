// Package overlay places one rendered string on top of another at a
// fixed (x, y) screen position — the foundational primitive for
// popups, tooltips, modals, and any other "draw this rectangle on
// top of that one" composition in a TUI.
//
// Overlay is a positioning function only. It splices the foreground
// string into the background at cell-level granularity (x measured in
// visible cells, y measured in lines), preserving the background's
// ANSI styling outside the overlaid rectangle and inserting the
// foreground's own ANSI styling inside it.
//
// # ANSI isolation contract
//
// Overlay does NOT track ANSI state across the splice. CSI escape
// sequences in the background are passed through to whichever side
// of the cut they happen to fall on, but no codes are synthesized
// or rebalanced. As a consequence:
//
//   - If the background row has an open SGR (e.g., a colored cell)
//     whose closing reset lives inside the dropped middle slice, the
//     foreground's first cells will inherit that open style.
//   - If the foreground does not end with an SGR reset, its trailing
//     state can leak into the background's right portion (the suffix
//     half of the cut).
//
// The fix for both is the SAME: the FOREGROUND must own its ANSI
// state. Practically: every row of the foreground should begin with
// "\x1b[0m" (or set explicit fg+bg+attrs on every cell), so the
// foreground reads as a self-contained rectangle no matter what's
// underneath. Popup-style components (menu, dropdown, modal) should
// emit a leading SGR reset per row and rely on lipgloss's trailing
// reset between styled pieces to keep state contained.
//
// This is a deliberate design choice: keeping Overlay style-blind
// makes it small, predictable, and free of an ANSI parser. The
// trade-off is that consumers — not the overlay primitive — own
// ANSI isolation.
//
// # Limitations
//
// Only ANSI SGR (CSI ... m) is considered when measuring visible
// width. Non-SGR CSI codes are passed through but their content
// width is treated as 0. OSC sequences (e.g., OSC8 hyperlinks),
// DCS, and APC are NOT recognized — if a foreground or background
// uses them, splicing may misbehave. For the typical lipgloss-only
// content this codebase produces, that's not a concern.
package overlay

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Overlay places fg on top of bg starting at screen (x, y). Each
// fg row replaces a horizontal span of the corresponding bg row
// from cell x to cell x + lipgloss.Width(fgRow). Rows of fg that
// fall outside bg's row range are dropped. Returns bg unchanged
// when fg is empty.
//
// Coordinate semantics:
//   - x is in visible cells (lipgloss.Width units), not bytes
//   - y is in lines (0-indexed)
//   - When x exceeds bg's row width, the row is padded with spaces
//     to reach x before fg is appended
//
// See the package doc for the ANSI isolation contract — Overlay
// does not synthesize state at the splice seams; the foreground
// must own its ANSI envelope.
func Overlay(bg, fg string, x, y int) string {
	if fg == "" {
		return bg
	}
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")
	for i, fgLine := range fgLines {
		row := y + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		bgLine := bgLines[row]
		bgWidth := lipgloss.Width(bgLine)
		fgWidth := lipgloss.Width(fgLine)

		// Pad bg with spaces if x lies past its visible width, so the
		// fg lands at the requested column rather than butting up
		// against the last visible character.
		if x > bgWidth {
			bgLine += strings.Repeat(" ", x-bgWidth)
		}
		prefix, rest := splitAtCol(bgLine, x)
		_, suffix := splitAtCol(rest, fgWidth)
		bgLines[row] = prefix + fgLine + suffix
	}
	return strings.Join(bgLines, "\n")
}

// splitAtCol splits s at the given visible column boundary, where
// "visible" means cells that count toward lipgloss.Width — CSI
// escape codes (the most common form for color/style) are passed
// through and stay attached to the half they appear in. Wide runes
// (CJK, emoji) are kept whole; the split may overshoot by one cell
// in that case.
func splitAtCol(s string, col int) (left, right string) {
	runes := []rune(s)
	var b strings.Builder
	visible := 0
	var i int
	for i = 0; i < len(runes); {
		r := runes[i]
		// Pass through CSI escape sequences: ESC [ params terminator.
		if r == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) {
				rj := runes[j]
				j++
				if rj >= 0x40 && rj <= 0x7e {
					break // CSI terminator (any uppercase or lowercase letter)
				}
			}
			b.WriteString(string(runes[i:j]))
			i = j
			continue
		}
		if visible >= col {
			break
		}
		b.WriteRune(r)
		visible += lipgloss.Width(string(r))
		i++
	}
	return b.String(), string(runes[i:])
}
