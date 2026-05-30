// Package progress renders a headless progress bar: a fixed-width run of
// filled and empty cells sized to Value/Max. It is a static primitive — no
// state, no input — so the caller owns every glyph and style.
package progress

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's per-segment styles and glyphs. Filled / Empty
// style the two runs; FilledChar / EmptyChar are the per-cell glyphs, empty
// falling back to "█" / " ".
type Style struct {
	Filled     lipgloss.Style
	Empty      lipgloss.Style
	FilledChar string
	EmptyChar  string
}

func (s Style) filledChar() string {
	if s.FilledChar != "" {
		return s.FilledChar
	}
	return "█"
}

func (s Style) emptyChar() string {
	if s.EmptyChar != "" {
		return s.EmptyChar
	}
	return " "
}

// Progress is a single bar. Value is clamped to [0, Max] before rendering;
// Width is the total cell count. Max <= 0 renders an all-empty bar.
type Progress struct {
	Value float64
	Max   float64
	Width int
	Style Style
}

// filledCells returns how many of Width cells are filled for the clamped
// Value/Max ratio.
func (p Progress) filledCells() int {
	if p.Width <= 0 {
		return 0
	}
	if p.Max <= 0 {
		return 0
	}
	v := p.Value
	if v < 0 {
		v = 0
	}
	if v > p.Max {
		v = p.Max
	}
	n := int(v / p.Max * float64(p.Width))
	if n > p.Width {
		n = p.Width
	}
	return n
}

// Render returns the bar as a single row of Width cells, or "" when
// Width <= 0.
func (p Progress) Render() string {
	if p.Width <= 0 {
		return ""
	}
	filled := p.filledCells()
	var b strings.Builder
	if filled > 0 {
		b.WriteString(p.Style.Filled.Render(strings.Repeat(p.Style.filledChar(), filled)))
	}
	if empty := p.Width - filled; empty > 0 {
		b.WriteString(p.Style.Empty.Render(strings.Repeat(p.Style.emptyChar(), empty)))
	}
	return b.String()
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (p Progress) View() string { return p.Render() }
