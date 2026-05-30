// Package radiogroup renders a headless single-select list: one option per
// row, each prefixed by a selected/unselected glyph. The package owns the
// selection behavior (key navigation, hit-testing); styling and glyphs are
// the caller's, supplied per-instance via Style.
package radiogroup

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's glyphs and per-state row styles. SelectedGlyph /
// UnselectedGlyph are the indicator strings; empty falls back to "(*)" /
// "( )". Selected styles the chosen row, Focused the rows while the group
// owns keyboard input, Normal everything else.
type Style struct {
	SelectedGlyph   string
	UnselectedGlyph string
	Selected        lipgloss.Style
	Normal          lipgloss.Style
	Focused         lipgloss.Style
}

func (s Style) selectedGlyph() string {
	if s.SelectedGlyph != "" {
		return s.SelectedGlyph
	}
	return "(*)"
}

func (s Style) unselectedGlyph() string {
	if s.UnselectedGlyph != "" {
		return s.UnselectedGlyph
	}
	return "( )"
}

// RadioGroup is a vertical single-select list. Selected is the index of the
// chosen option. Focused reports whether the group owns keyboard input (the
// caller sets it); Update only responds when Focused.
type RadioGroup struct {
	Options  []string
	Selected int
	Focused  bool
	Style    Style

	// Up/Down are the keys that move Selected; nil falls back to
	// ["up", "k"] / ["down", "j"].
	Up, Down []string
}

// ChangedMsg is emitted, via the cmd from Update or ClickAt, when Selected
// changes. Selected is the new index.
type ChangedMsg struct{ Selected int }

func (r RadioGroup) row(i string, selected bool) string {
	if selected {
		return r.Style.selectedGlyph() + " " + i
	}
	return r.Style.unselectedGlyph() + " " + i
}

// Render draws one option per line, the selected row styled with
// Style.Selected and the rest with Focused (when Focused) or Normal.
func (r RadioGroup) Render() string {
	rows := make([]string, len(r.Options))
	for i, opt := range r.Options {
		st := r.Style.Normal
		if r.Focused {
			st = r.Style.Focused
		}
		if i == r.Selected {
			st = r.Style.Selected
		}
		rows[i] = st.Render(r.row(opt, i == r.Selected))
	}
	return strings.Join(rows, "\n")
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (r RadioGroup) View() string { return r.Render() }

// ApplyKey moves Selected on an Up/Down key and reports whether it changed.
// It is the pure key handler Update wraps; it does not check Focused.
func (r RadioGroup) ApplyKey(key string) (RadioGroup, bool) {
	if len(r.Options) == 0 {
		return r, false
	}
	prev := r.Selected
	switch {
	case contains(r.keys(r.Up, "up", "k"), key):
		if r.Selected > 0 {
			r.Selected--
		}
	case contains(r.keys(r.Down, "down", "j"), key):
		if r.Selected < len(r.Options)-1 {
			r.Selected++
		}
	}
	return r, r.Selected != prev
}

// Update moves Selected on an Up/Down key and emits ChangedMsg when it
// changes. It is a no-op unless Focused, and ignores non-key messages.
func (r RadioGroup) Update(msg tea.Msg) (RadioGroup, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !r.Focused {
		return r, nil
	}
	r, changed := r.ApplyKey(key.String())
	if !changed {
		return r, nil
	}
	sel := r.Selected
	return r, func() tea.Msg { return ChangedMsg{Selected: sel} }
}

// HitTest returns the option index at panel-local (x, y) and whether the
// coordinate lands on a row. Rows are one line tall; x is ignored beyond
// requiring a non-negative value, since a row spans the full group width.
func (r RadioGroup) HitTest(x, y int) (int, bool) {
	if x < 0 || y < 0 || y >= len(r.Options) {
		return 0, false
	}
	return y, true
}

// ClickAt selects the option at panel-local (x, y) and emits ChangedMsg when
// the selection changes; a miss, or a click on the already-selected row, is a
// no-op.
func (r RadioGroup) ClickAt(x, y int) (RadioGroup, tea.Cmd) {
	idx, ok := r.HitTest(x, y)
	if !ok || idx == r.Selected {
		return r, nil
	}
	r.Selected = idx
	return r, func() tea.Msg { return ChangedMsg{Selected: idx} }
}

func (r RadioGroup) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
