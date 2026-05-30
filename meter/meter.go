// Package meter renders a headless gauge bar: a fixed-width strip whose
// filled portion reflects a value within [Min, Max], with optional Low/High
// thresholds that switch the fill style (e.g. green/yellow/red). It owns the
// fill math and threshold selection; the caller owns every style and glyph.
//
// A Meter is static — it has no state to mutate, so there is no Update; View
// is an alias for Render.
package meter

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's per-zone fill styles and the fill/empty glyphs.
// Normal styles a value below the High threshold (and at or above Low);
// LowFill styles a value below the Low threshold; HighFill styles a value at
// or above the High threshold. Empty styles the unfilled remainder.
//
// FillChar / EmptyChar are the single-cell glyphs; empty falls back to a solid
// block and a space.
type Style struct {
	Normal    lipgloss.Style
	LowFill   lipgloss.Style
	HighFill  lipgloss.Style
	Empty     lipgloss.Style
	FillChar  string
	EmptyChar string
}

func (s Style) fillChar() string {
	if s.FillChar != "" {
		return s.FillChar
	}
	return "█"
}

func (s Style) emptyChar() string {
	if s.EmptyChar != "" {
		return s.EmptyChar
	}
	return " "
}

// Meter is a gauge bar. Value is clamped to [Min, Max] for the fill fraction.
// Width is the bar's cell count. Low and High are optional thresholds in the
// value's own units that select the fill style: a zero threshold is treated as
// unset, so a bare Meter uses Style.Normal for the whole fill.
type Meter struct {
	Value float64
	Min   float64
	Max   float64
	Width int

	Low  float64
	High float64

	Style Style
}

// fraction returns Value's position in [Min, Max], clamped to [0, 1]. A
// non-positive span yields 0 (a zero-width or inverted range can't fill).
func (m Meter) fraction() float64 {
	span := m.Max - m.Min
	if span <= 0 {
		return 0
	}
	f := (m.Value - m.Min) / span
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// fillStyle picks the fill style from Value against the thresholds. High wins
// over Low when both apply. A zero threshold is unset and never matches.
func (m Meter) fillStyle() lipgloss.Style {
	if m.High != 0 && m.Value >= m.High {
		return m.Style.HighFill
	}
	if m.Low != 0 && m.Value < m.Low {
		return m.Style.LowFill
	}
	return m.Style.Normal
}

// Render returns the gauge as a single Width-cell row, or "" when Width <= 0.
// The first round(fraction*Width) cells use the threshold-selected fill style
// and glyph; the rest use the empty style and glyph.
func (m Meter) Render() string {
	if m.Width <= 0 {
		return ""
	}
	filled := int(m.fraction()*float64(m.Width) + 0.5)
	if filled > m.Width {
		filled = m.Width
	}
	fill := m.fillStyle().Render(strings.Repeat(m.Style.fillChar(), filled))
	empty := m.Style.Empty.Render(strings.Repeat(m.Style.emptyChar(), m.Width-filled))
	return fill + empty
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (m Meter) View() string { return m.Render() }
