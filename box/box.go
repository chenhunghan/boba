// Package box renders a bordered TUI region with optional labeled
// notches in its top edge.
//
// Notches are small inline labels in the top border — they look like
// cutouts in the edge, drawn with the downward-corner glyphs ┐ and ┌:
//
//	┌──┐cpu┌─┐menu┌┐preset┌──────────┐2100ms┌┐
//	│                                        │
//	│  body content goes here                │
//	└────────────────────────────────────────┘
//
// Each Notch carries its own Gap (number of ─ dashes drawn before its
// opening ┐), so callers control spacing per-label rather than via a
// single global setting.
//
// Render is a pure function from (Box, width, height) to a string of
// exactly width × height cells — making it safe to compose into larger
// layouts via lipgloss.JoinHorizontal / JoinVertical.
package box

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Box is a bordered TUI region with bracketed labels in its top edge.
// The zero value is a box with no labels, no body, and no border color
// (the terminal default) — useful as a placeholder.
type Box struct {
	// LeftNotches are labeled notches on the left side of the top border,
	// rendered after the box's top-left corner. By convention the first
	// left notch is the box's title.
	LeftNotches []Notch

	// RightNotches are labeled notches on the right side of the top
	// border, rendered before the box's top-right corner.
	RightNotches []Notch

	// Body is the box's interior content. Newline-separated lines are
	// placed in the box's interior (width-2 × height-2 with borders,
	// or the full width × height when Borderless); short content is
	// padded with empty rows, long content is truncated.
	Body string

	// BorderColor accents the box's box-drawing characters. Empty
	// applies no color (the terminal default). Ignored when Borderless.
	BorderColor lipgloss.Color

	// DimBorder renders the border with the terminal's "faint"
	// attribute. Ignored when Borderless.
	DimBorder bool

	// Borderless skips the box-drawing characters (no top/bottom
	// edges, no side bars) and gives the body the full width × height
	// area. LeftNotches and RightNotches are not rendered in this mode.
	// Useful for solid-fill rectangles, e.g. icon buttons.
	Borderless bool

	// HideTopBorder, HideRightBorder, HideBottomBorder, HideLeftBorder
	// turn off individual edges. Notches are only rendered when the top
	// edge is shown. This is finer-grained than Borderless (which
	// hides everything) — useful for "card" style cells with only a
	// thick left bar, etc.
	HideTopBorder    bool
	HideRightBorder  bool
	HideBottomBorder bool
	HideLeftBorder   bool

	// LeftBorderChar overrides the glyph used for the left side of
	// the box. Empty falls back to "│". Useful for thicker bars like
	// "▌" or "┃".
	LeftBorderChar string

	// FillColor sets the background color of the box's interior cells
	// (and the entire area in Borderless mode). Empty means terminal
	// default (transparent).
	FillColor lipgloss.Color
}

// borderStyle returns a lipgloss style for this box's box-drawing runes.
func (b Box) borderStyle() lipgloss.Style {
	s := lipgloss.NewStyle()
	if b.BorderColor != "" {
		s = s.Foreground(b.BorderColor)
	}
	if b.DimBorder {
		s = s.Faint(true)
	}
	return s
}

// Render assembles the box as a string of exactly width × height
// cells. Returns "" if width or height is too small for the requested
// borders to fit (each enabled side consumes one row/column).
func (b Box) Render(width, height int) string {
	// Borderless is shorthand for "hide every side and skip notches".
	showTop := !b.Borderless && !b.HideTopBorder
	showRight := !b.Borderless && !b.HideRightBorder
	showBottom := !b.Borderless && !b.HideBottomBorder
	showLeft := !b.Borderless && !b.HideLeftBorder

	horizBorders := 0
	if showLeft {
		horizBorders++
	}
	if showRight {
		horizBorders++
	}
	vertBorders := 0
	if showTop {
		vertBorders++
	}
	if showBottom {
		vertBorders++
	}
	minWidth := horizBorders + 1
	minHeight := vertBorders + 1
	// The fully-bordered case historically required at least 4×3 so
	// that notches and corner glyphs have somewhere to go. Preserve that.
	if showTop && showRight && showBottom && showLeft {
		if minWidth < 4 {
			minWidth = 4
		}
		if minHeight < 3 {
			minHeight = 3
		}
	}
	if width < minWidth || height < minHeight {
		return ""
	}

	innerWidth := width - horizBorders
	innerHeight := height - vertBorders

	bodyLines := strings.Split(b.Body, "\n")
	if len(bodyLines) > innerHeight {
		bodyLines = bodyLines[:innerHeight]
	}
	for len(bodyLines) < innerHeight {
		bodyLines = append(bodyLines, "")
	}

	leftChar := "│"
	if b.LeftBorderChar != "" {
		leftChar = b.LeftBorderChar
	}
	border := b.borderStyle()
	fill := b.fillStyle()

	for i, line := range bodyLines {
		pad := innerWidth - lipgloss.Width(line)
		if pad < 0 {
			line = line[:innerWidth]
			pad = 0
		}
		inner := line + strings.Repeat(" ", pad)
		if b.FillColor != "" {
			inner = fill.Render(inner)
		}
		var prefix, suffix string
		if showLeft {
			prefix = border.Render(leftChar)
		}
		if showRight {
			suffix = border.Render("│")
		}
		bodyLines[i] = prefix + inner + suffix
	}

	var rows []string
	if showTop {
		rows = append(rows, b.renderTop(width))
	}
	rows = append(rows, bodyLines...)
	if showBottom {
		rows = append(rows, b.renderBottom(width))
	}
	return strings.Join(rows, "\n")
}

// fillStyle returns a lipgloss style applying the box's FillColor as
// a background. Returns a no-op style when FillColor is empty.
func (b Box) fillStyle() lipgloss.Style {
	if b.FillColor == "" {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Background(b.FillColor)
}

// renderTop builds the top edge of the box.
func (b Box) renderTop(width int) string {
	border := b.borderStyle()

	// Render a single notch as ┐<badge><text>┌. Each Notch carries its
	// own NotchStyle; the package just stitches the segments together
	// between the border corner glyphs.
	renderNotch := func(t Notch) (string, int) {
		var content string
		if t.Badge != "" {
			content = t.Style.Badge.Render(t.Badge)
		}
		content += t.Style.Text.Render(t.Text)
		seg := border.Render("┐") + content + border.Render("┌")
		return seg, lipgloss.Width(seg)
	}

	// Left side: ┌ + (gap + notch) for each left notch.
	var leftSb strings.Builder
	leftSb.WriteString(border.Render("┌"))
	leftWidth := 1 // box TL ┌
	for _, t := range b.LeftNotches {
		if t.Gap > 0 {
			leftSb.WriteString(border.Render(strings.Repeat("─", t.Gap)))
			leftWidth += t.Gap
		}
		seg, w := renderNotch(t)
		leftSb.WriteString(seg)
		leftWidth += w
	}

	// Right side (excluding box TR ┐).
	var rightSb strings.Builder
	rightWidth := 0
	for _, t := range b.RightNotches {
		if t.Gap > 0 {
			rightSb.WriteString(border.Render(strings.Repeat("─", t.Gap)))
			rightWidth += t.Gap
		}
		seg, w := renderNotch(t)
		rightSb.WriteString(seg)
		rightWidth += w
	}

	// Trailing fill between the two sides.
	trailing := width - leftWidth - rightWidth - 1
	if trailing < 0 {
		trailing = 0
	}

	return leftSb.String() +
		border.Render(strings.Repeat("─", trailing)) +
		rightSb.String() +
		border.Render("┐")
}

// renderBottom builds a plain bottom edge.
func (b Box) renderBottom(width int) string {
	if width < 2 {
		return ""
	}
	return b.borderStyle().Render("└" + strings.Repeat("─", width-2) + "┘")
}
