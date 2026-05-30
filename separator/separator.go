// Package separator renders a headless divider: a run of a single rune,
// horizontal or vertical, styled by the caller.
package separator

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Orientation selects a horizontal or vertical divider.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// Separator is a divider line. Length is the number of cells (Horizontal)
// or rows (Vertical). Char is the repeated rune; empty falls back to "─"
// for Horizontal and "│" for Vertical.
type Separator struct {
	Orientation Orientation
	Length      int
	Char        string
	Style       lipgloss.Style
}

func (s Separator) char() string {
	if s.Char != "" {
		return s.Char
	}
	if s.Orientation == Vertical {
		return "│"
	}
	return "─"
}

// Render returns the divider, or "" when Length <= 0.
func (s Separator) Render() string {
	if s.Length <= 0 {
		return ""
	}
	ch := s.char()
	if s.Orientation == Vertical {
		rows := make([]string, s.Length)
		for i := range rows {
			rows[i] = s.Style.Render(ch)
		}
		return strings.Join(rows, "\n")
	}
	return s.Style.Render(strings.Repeat(ch, s.Length))
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (s Separator) View() string { return s.Render() }
