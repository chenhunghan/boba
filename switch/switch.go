// Package swtch renders a headless on/off switch: an indicator glyph plus an
// optional label, flipped by key or click. Styling and glyphs are the
// caller's; the package owns the toggle behavior.
//
// The directory is "switch" but the package clause is "swtch": "switch" is a
// Go keyword and cannot be used as a package identifier. Callers import
// "github.com/chenhunghan/boba/switch" and reference it as swtch.Switch.
package swtch

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's glyphs and per-state text styles. OnGlyph /
// OffGlyph are the indicator strings; empty falls back to "[on]" / "[off]".
type Style struct {
	OnGlyph  string
	OffGlyph string
	Normal   lipgloss.Style
	Focused  lipgloss.Style
}

// Switch is a single labeled boolean toggle. On is its state; Focused reports
// whether it owns keyboard input (the caller sets it), and Update only
// responds when Focused.
type Switch struct {
	On      bool
	Label   string
	Focused bool
	Style   Style

	// Toggle are the keys that flip On; nil falls back to enter/space.
	Toggle []string
}

// ToggledMsg is emitted, via the cmd from Update or ClickAt, when the switch
// flips. On is the new value.
type ToggledMsg struct{ On bool }

func (s Switch) glyph() string {
	if s.On {
		if s.Style.OnGlyph != "" {
			return s.Style.OnGlyph
		}
		return "[on]"
	}
	if s.Style.OffGlyph != "" {
		return s.Style.OffGlyph
	}
	return "[off]"
}

func (s Switch) text() string {
	if s.Label == "" {
		return s.glyph()
	}
	return s.glyph() + " " + s.Label
}

// Render draws the switch on a single row, styled by Focused.
func (s Switch) Render() string {
	st := s.Style.Normal
	if s.Focused {
		st = s.Style.Focused
	}
	return st.Render(s.text())
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (s Switch) View() string { return s.Render() }

// Width is the rendered width in cells — the caller needs it to route clicks
// (ClickAt takes panel-local coordinates).
func (s Switch) Width() int { return lipgloss.Width(s.text()) }

// Update flips On on a Toggle key and emits ToggledMsg. It is a no-op unless
// Focused, and ignores non-key messages.
func (s Switch) Update(msg tea.Msg) (Switch, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !s.Focused {
		return s, nil
	}
	toggle := s.Toggle
	if toggle == nil {
		toggle = []string{"enter", " "}
	}
	k := key.String()
	for _, t := range toggle {
		if k == t {
			return s.toggled()
		}
	}
	return s, nil
}

// ClickAt flips On when (x, y) lands on the switch's single row.
func (s Switch) ClickAt(x, y int) (Switch, tea.Cmd) {
	if y != 0 || x < 0 || x >= s.Width() {
		return s, nil
	}
	return s.toggled()
}

func (s Switch) toggled() (Switch, tea.Cmd) {
	s.On = !s.On
	on := s.On
	return s, func() tea.Msg { return ToggledMsg{On: on} }
}
