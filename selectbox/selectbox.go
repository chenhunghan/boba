// Package selectbox renders a headless select / dropdown: a closed control
// showing the chosen value and a chevron, which opens into a popup list of
// options. It owns behavior — open/closed state, keyboard navigation,
// hit-testing, and the popup composition (placement, ANSI isolation, overlay)
// — while every visual decision lives in a caller-supplied Style.
//
// The closed control is positioned by the caller via its anchor rect
// (X, Y, W, H, in screen coordinates); the open list is placed against that
// rect with the popup and overlay primitives, so the same composition rules
// apply as the rest of boba.
//
// Store a SelectBox on your model, route messages through Update, render the
// closed control with View, and composite the open list onto your view with
// OpenView:
//
//	m.sel, cmd = m.sel.Update(msg)
//	...
//	base = overlayClosed(base, m.sel.View(), m.sel.X, m.sel.Y)
//	return m.sel.OpenView(base, screenW, screenH)
//
// The name is selectbox because select is a reserved keyword in Go.
package selectbox

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/overlay"
	"github.com/chenhunghan/boba/popup"
)

// Style holds the caller's chevron glyph and per-state styles. Closed styles
// the closed control, Option each unselected option row in the open list, and
// Highlighted the row under the cursor. Chevron is the trailing indicator on
// the closed control; empty falls back to "▾".
type Style struct {
	Chevron     string
	Closed      lipgloss.Style
	Option      lipgloss.Style
	Highlighted lipgloss.Style
}

func (s Style) chevron() string {
	if s.Chevron != "" {
		return s.Chevron
	}
	return "▾"
}

// SelectBox is a select / dropdown control. Options are the choices; Selected
// is the index of the current value. Open gates the popup list and its input.
// Focused reports whether the control owns keyboard input (the caller sets
// it); Update only responds when Focused. Highlight is the index of the row
// under the cursor while open. X, Y, W, H are the anchor rectangle of the
// closed control in screen coordinates, against which the open list is placed.
type SelectBox struct {
	Options   []string
	Selected  int
	Open      bool
	Focused   bool
	Highlight int

	X, Y, W, H int

	Style Style

	// Key bindings; nil falls back to the defaults noted on each field.
	OpenKeys []string // open the list while closed; default ["enter", " ", "down"]
	Up       []string // move Highlight up while open; default ["up", "k"]
	Down     []string // move Highlight down while open; default ["down", "j"]
	Confirm  []string // confirm Highlight while open; default ["enter", " "]
	Cancel   []string // close without confirming; default ["esc"]
}

// ChangedMsg is emitted, via the cmd from Update or ClickAt, when the user
// confirms an option. Selected is the new index.
type ChangedMsg struct{ Selected int }

func (s SelectBox) value() string {
	if s.Selected < 0 || s.Selected >= len(s.Options) {
		return ""
	}
	return s.Options[s.Selected]
}

// listWidth is the open list's width in cells: the widest option, the closed
// control's content width, and W are all candidates, so the list never renders
// narrower than the control it drops from.
func (s SelectBox) listWidth() int {
	w := s.W
	if cw := lipgloss.Width(s.closedContent()); cw > w {
		w = cw
	}
	for _, opt := range s.Options {
		if ow := lipgloss.Width(opt); ow > w {
			w = ow
		}
	}
	return w
}

// closedContent is the unstyled closed-control text: value left-aligned with
// the chevron pinned to the right within W (when W leaves room), else value
// and chevron separated by a single space.
func (s SelectBox) closedContent() string {
	val := s.value()
	chev := s.Style.chevron()
	gap := s.W - lipgloss.Width(val) - lipgloss.Width(chev)
	if gap < 1 {
		gap = 1
	}
	return val + repeat(" ", gap) + chev
}

// Render draws the closed control on a single row, styled with Style.Closed.
func (s SelectBox) Render() string {
	return s.Style.Closed.Render(s.closedContent())
}

// View is an alias for Render, matching the Bubble Tea View() convention. It
// is the closed control only; composite the open list separately via OpenView.
func (s SelectBox) View() string { return s.Render() }

// renderList returns the open list as a multi-line string sized listWidth wide,
// one row per option. Each row begins with an SGR reset and is padded to the
// full width with its row style so the popup owns its ANSI envelope and reads
// as a solid surface over whatever OpenView composites it onto.
func (s SelectBox) renderList() string {
	const sgrReset = "\x1b[0m"
	w := s.listWidth()
	rows := make([]string, len(s.Options))
	for i, opt := range s.Options {
		st := s.Style.Option
		if i == s.Highlight {
			st = s.Style.Highlighted
		}
		pad := w - lipgloss.Width(opt)
		if pad < 0 {
			pad = 0
		}
		rows[i] = sgrReset + st.Render(opt+repeat(" ", pad))
	}
	return join(rows)
}

// OpenView composites the open list onto bg and returns the result, or bg
// unchanged when closed. The list is placed directly below the anchor rect via
// popup.Place (flipping above when it would overflow), isolated so it cannot
// inherit bg's ANSI, and spliced in with overlay.Overlay.
func (s SelectBox) OpenView(bg string, screenW, screenH int) string {
	if !s.Open || len(s.Options) == 0 {
		return bg
	}
	list := s.renderList()
	x, y := popup.Place(s.X, s.Y, s.W, s.H, s.listWidth(), len(s.Options), screenW, screenH, popup.Below)
	return overlay.Overlay(bg, popup.Isolate(list), x, y)
}

// ApplyKey is the pure key handler Update wraps. It does not check Focused.
// While closed, an OpenKeys key opens the list (highlighting the current
// Selected). While open: Up/Down move Highlight, Confirm selects the
// highlighted option and closes (returning confirmed=true), Cancel closes
// without selecting. The returned confirmed is true only when an option was
// just chosen.
func (s SelectBox) ApplyKey(key string) (SelectBox, bool) {
	if len(s.Options) == 0 {
		return s, false
	}
	if !s.Open {
		if contains(s.keys(s.OpenKeys, "enter", " ", "down"), key) {
			s.Open = true
			s.Highlight = clampIndex(s.Selected, len(s.Options))
		}
		return s, false
	}
	switch {
	case contains(s.keys(s.Cancel, "esc"), key):
		s.Open = false
	case contains(s.keys(s.Confirm, "enter", " "), key):
		s.Open = false
		if s.Highlight >= 0 && s.Highlight < len(s.Options) {
			s.Selected = s.Highlight
			return s, true
		}
	case contains(s.keys(s.Up, "up", "k"), key):
		if s.Highlight > 0 {
			s.Highlight--
		}
	case contains(s.keys(s.Down, "down", "j"), key):
		if s.Highlight < len(s.Options)-1 {
			s.Highlight++
		}
	}
	return s, false
}

// Update routes a key message through ApplyKey and emits ChangedMsg when an
// option is confirmed. It is a no-op unless Focused, and ignores non-key
// messages — clicks go through ClickAt, which takes panel-local coordinates.
func (s SelectBox) Update(msg tea.Msg) (SelectBox, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !s.Focused {
		return s, nil
	}
	s, confirmed := s.ApplyKey(key.String())
	if !confirmed {
		return s, nil
	}
	sel := s.Selected
	return s, func() tea.Msg { return ChangedMsg{Selected: sel} }
}

// HitTest maps panel-local (x, y) to a target. y == 0 is the closed control.
// When Open, the list rows sit directly below it, so y in [1, 1+len(Options))
// addresses option y-1. row is the option index for a list hit, or -1 for a
// control hit; ok is false when the coordinate lands on neither.
func (s SelectBox) HitTest(x, y int) (row int, ok bool) {
	if x < 0 {
		return -1, false
	}
	if y == 0 {
		if x >= s.listWidth() {
			return -1, false
		}
		return -1, true
	}
	if s.Open {
		idx := y - 1
		if idx >= 0 && idx < len(s.Options) && x < s.listWidth() {
			return idx, true
		}
	}
	return -1, false
}

// ClickAt applies a click at panel-local (x, y) — typically the LocalX / LocalY
// from panel.HitTest. A click on the closed control toggles Open (highlighting
// the current Selected when opening). When open, a click on an option selects
// it, closes the list, and emits ChangedMsg. A miss is a no-op.
func (s SelectBox) ClickAt(x, y int) (SelectBox, tea.Cmd) {
	row, ok := s.HitTest(x, y)
	if !ok {
		return s, nil
	}
	if row < 0 {
		s.Open = !s.Open
		if s.Open {
			s.Highlight = clampIndex(s.Selected, len(s.Options))
		}
		return s, nil
	}
	s.Selected = row
	s.Highlight = row
	s.Open = false
	return s, func() tea.Msg { return ChangedMsg{Selected: row} }
}

func (s SelectBox) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func clampIndex(i, n int) int {
	if i < 0 || i >= n {
		return 0
	}
	return i
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}

func join(rows []string) string {
	out := ""
	for i, r := range rows {
		if i > 0 {
			out += "\n"
		}
		out += r
	}
	return out
}
