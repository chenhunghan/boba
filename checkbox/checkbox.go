// Package checkbox renders a headless checkbox: an indicator glyph plus an
// optional label, toggled by key or click. Styling and glyphs are the
// caller's; the package owns the toggle behavior.
package checkbox

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's glyphs and per-state text styles. Checked /
// Unchecked are the indicator strings; empty falls back to "[x]" / "[ ]".
type Style struct {
	Checked   string
	Unchecked string
	Normal    lipgloss.Style
	Focused   lipgloss.Style
}

// Checkbox is a single labeled checkbox. Focused reports whether it owns
// keyboard input (the caller sets it); Update only responds when Focused.
type Checkbox struct {
	Label   string
	Checked bool
	Focused bool
	Style   Style

	// Toggle are the keys that flip Checked; nil falls back to enter/space.
	Toggle []string
}

// ToggledMsg is emitted, via the cmd from Update or ClickAt, when the
// checkbox flips. Checked is the new value.
type ToggledMsg struct{ Checked bool }

func (c Checkbox) glyph() string {
	if c.Checked {
		if c.Style.Checked != "" {
			return c.Style.Checked
		}
		return "[x]"
	}
	if c.Style.Unchecked != "" {
		return c.Style.Unchecked
	}
	return "[ ]"
}

func (c Checkbox) text() string {
	if c.Label == "" {
		return c.glyph()
	}
	return c.glyph() + " " + c.Label
}

// Render draws the checkbox on a single row.
func (c Checkbox) Render() string {
	st := c.Style.Normal
	if c.Focused {
		st = c.Style.Focused
	}
	return st.Render(c.text())
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (c Checkbox) View() string { return c.Render() }

// Width is the rendered width in cells — the caller needs it to route
// clicks (ClickAt takes panel-local coordinates).
func (c Checkbox) Width() int { return lipgloss.Width(c.text()) }

// Update flips Checked on a Toggle key and emits ToggledMsg. It is a no-op
// unless Focused, and ignores non-key messages.
func (c Checkbox) Update(msg tea.Msg) (Checkbox, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !c.Focused {
		return c, nil
	}
	toggle := c.Toggle
	if toggle == nil {
		toggle = []string{"enter", " "}
	}
	k := key.String()
	for _, t := range toggle {
		if k == t {
			return c.toggled()
		}
	}
	return c, nil
}

// ClickAt flips Checked when (x, y) lands on the checkbox's single row.
func (c Checkbox) ClickAt(x, y int) (Checkbox, tea.Cmd) {
	if y != 0 || x < 0 || x >= c.Width() {
		return c, nil
	}
	return c.toggled()
}

func (c Checkbox) toggled() (Checkbox, tea.Cmd) {
	c.Checked = !c.Checked
	checked := c.Checked
	return c, func() tea.Msg { return ToggledMsg{Checked: checked} }
}
