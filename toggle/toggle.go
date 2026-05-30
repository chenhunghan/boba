// Package toggle renders a headless toggle button: a pressable label that
// stays pressed until toggled again. Styling is the caller's via a per-state
// Style; the package owns the press/toggle behavior.
package toggle

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's per-state text styles. Released is the resting
// look, Pressed the toggled-on look, and Focused overrides either when the
// toggle owns keyboard input.
type Style struct {
	Released lipgloss.Style
	Pressed  lipgloss.Style
	Focused  lipgloss.Style
}

// Toggle is a single labeled toggle button. Pressed is its sticky on/off
// state; Focused reports whether it owns keyboard input (the caller sets it),
// and Update only responds when Focused.
type Toggle struct {
	Label   string
	Pressed bool
	Focused bool
	Style   Style

	// Toggle are the keys that flip Pressed; nil falls back to enter/space.
	Toggle []string
}

// ToggledMsg is emitted, via the cmd from Update or ClickAt, when the toggle
// flips. Pressed is the new value.
type ToggledMsg struct{ Pressed bool }

// Render draws the toggle's label on a single row, styled by Pressed/Focused.
// Focused takes priority over the Pressed/Released look.
func (t Toggle) Render() string {
	st := t.Style.Released
	if t.Pressed {
		st = t.Style.Pressed
	}
	if t.Focused {
		st = t.Style.Focused
	}
	return st.Render(t.Label)
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (t Toggle) View() string { return t.Render() }

// Width is the rendered width in cells — the caller needs it to route clicks
// (ClickAt takes panel-local coordinates).
func (t Toggle) Width() int { return lipgloss.Width(t.Label) }

// Update flips Pressed on a Toggle key and emits ToggledMsg. It is a no-op
// unless Focused, and ignores non-key messages.
func (t Toggle) Update(msg tea.Msg) (Toggle, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !t.Focused {
		return t, nil
	}
	toggle := t.Toggle
	if toggle == nil {
		toggle = []string{"enter", " "}
	}
	k := key.String()
	for _, kb := range toggle {
		if k == kb {
			return t.toggled()
		}
	}
	return t, nil
}

// ClickAt flips Pressed when (x, y) lands on the toggle's single row.
func (t Toggle) ClickAt(x, y int) (Toggle, tea.Cmd) {
	if y != 0 || x < 0 || x >= t.Width() {
		return t, nil
	}
	return t.toggled()
}

func (t Toggle) toggled() (Toggle, tea.Cmd) {
	t.Pressed = !t.Pressed
	pressed := t.Pressed
	return t, func() tea.Msg { return ToggledMsg{Pressed: pressed} }
}
