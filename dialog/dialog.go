// Package dialog is a headless modal dialog: a centered box with a title, a
// body, and a row of buttons, drawn on top of arbitrary background content.
// It owns behavior — button navigation, key handling, centering and ANSI
// isolation over the background — while every visual decision lives in
// caller-supplied styles.
//
// A Dialog is a value type: store it on your model, route Bubble Tea
// messages through Update, render the box with View, and place it over your
// screen with Over.
//
//	case tea.KeyMsg:
//	    m.dialog, cmd = m.dialog.Update(msg)
//	    return m, cmd
//	case dialog.ChosenMsg:
//	    return choose(m, msg.Index), nil
//	case dialog.CancelledMsg:
//	    m.dialog.Open = false
//	    return m, nil
//
//	// in View, after rendering the rest of the screen:
//	return m.dialog.Over(screen, w, h)
package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/overlay"
	"github.com/chenhunghan/boba/popup"
)

// Style holds the per-instance styling for a dialog. Surface paints every
// owned cell (border, title, body, button row, padding) with one background;
// Title and Body style their respective text. Button and ButtonSelected
// style the unselected and selected buttons. Border styles the box-drawing
// runes. Glyph fields fall back to the box-drawing defaults when empty.
type Style struct {
	Surface lipgloss.Color

	Title  lipgloss.Style
	Body   lipgloss.Style
	Border lipgloss.Style

	Button         lipgloss.Style
	ButtonSelected lipgloss.Style

	// Glyphs. Empty fields fall back to the box-drawing defaults.
	TopLeft, TopRight, BottomLeft, BottomRight string
	Horizontal, Vertical                       string
}

func (s Style) withSurface(style lipgloss.Style) lipgloss.Style {
	if s.Surface == "" {
		return style
	}
	return style.Background(s.Surface)
}

func orDefault(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

func (s Style) topLeft() string     { return orDefault(s.TopLeft, "┌") }
func (s Style) topRight() string    { return orDefault(s.TopRight, "┐") }
func (s Style) bottomLeft() string  { return orDefault(s.BottomLeft, "└") }
func (s Style) bottomRight() string { return orDefault(s.BottomRight, "┘") }
func (s Style) horizontal() string  { return orDefault(s.Horizontal, "─") }
func (s Style) vertical() string    { return orDefault(s.Vertical, "│") }

// Dialog is the modal widget. Title and Body are the displayed text; Buttons
// is the action row; Selected is the index of the highlighted button; Open
// gates Update and Over. Style is owned per-instance.
type Dialog struct {
	Title    string
	Body     string
	Buttons  []string
	Selected int
	Open     bool
	Style    Style

	// Key bindings (defaults applied internally when nil): Prev/Next move
	// Selected across Buttons, Confirm chooses it, Cancel dismisses.
	Prev    []string // default ["left", "h"]
	Next    []string // default ["right", "l"]
	Confirm []string // default ["enter"]
	Cancel  []string // default ["esc"]
}

// ChosenMsg is emitted, via the cmd from Update, when a button is confirmed.
// Index is the chosen button.
type ChosenMsg struct{ Index int }

// CancelledMsg is emitted, via the cmd from Update, when the dialog is
// dismissed with the Cancel key.
type CancelledMsg struct{}

// Width returns the dialog's outer width in cells (border + 1-cell side
// padding + the widest of title, body, and the button row).
func (d Dialog) Width() int {
	inner := lipgloss.Width(d.Title)
	if w := lipgloss.Width(d.Body); w > inner {
		inner = w
	}
	if w := lipgloss.Width(d.buttonRow()); w > inner {
		inner = w
	}
	// 2 borders + 2 padding + content.
	return inner + 4
}

// Height returns the dialog's outer height in cells: top border, title row,
// body row, button row, bottom border.
func (d Dialog) Height() int { return 5 }

func (d Dialog) buttonRow() string {
	if len(d.Buttons) == 0 {
		return ""
	}
	parts := make([]string, len(d.Buttons))
	for i, label := range d.Buttons {
		st := d.Style.Button
		if i == d.Selected {
			st = d.Style.ButtonSelected
		}
		parts[i] = d.Style.withSurface(st).Render(label)
	}
	return strings.Join(parts, " ")
}

// Render returns the modal box sized Width x Height. Each row begins with an
// SGR reset so the box owns its ANSI envelope when overlaid; the output
// carries no positioning of its own — use Over to place it, or compose it
// yourself.
func (d Dialog) Render() string {
	w := d.Width()
	innerW := w - 2

	const sgrReset = "\x1b[0m"
	border := d.Style.withSurface(d.Style.Border)

	edge := func(left, right string) string {
		return sgrReset + border.Render(left+strings.Repeat(d.Style.horizontal(), innerW)+right)
	}
	row := func(content string) string {
		return sgrReset + border.Render(d.Style.vertical()) + content + border.Render(d.Style.vertical())
	}

	rows := []string{
		edge(d.Style.topLeft(), d.Style.topRight()),
		row(d.pad(d.Style.withSurface(d.Style.Title).Render(d.Title), innerW)),
		row(d.pad(d.Style.withSurface(d.Style.Body).Render(d.Body), innerW)),
		row(d.pad(d.buttonRow(), innerW)),
		edge(d.Style.bottomLeft(), d.Style.bottomRight()),
	}
	return strings.Join(rows, "\n")
}

// pad surrounds already-styled content with 1 cell of side padding and fills
// the remaining inner width, all on the dialog's surface, so the row spans
// exactly innerW cells with a uniform background.
func (d Dialog) pad(content string, innerW int) string {
	fill := innerW - 2 - lipgloss.Width(content)
	if fill < 0 {
		fill = 0
	}
	spaces := d.Style.withSurface(lipgloss.NewStyle())
	return spaces.Render(" ") + content + spaces.Render(strings.Repeat(" ", fill)+" ")
}

// Over centers the dialog on a screenW x screenH screen and overlays it on
// bg, returning the composited screen. When closed, bg is returned unchanged.
func (d Dialog) Over(bg string, screenW, screenH int) string {
	if !d.Open {
		return bg
	}
	x, y := popup.Center(d.Width(), d.Height(), screenW, screenH)
	return overlay.Overlay(bg, popup.Isolate(d.Render()), x, y)
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (d Dialog) View() string { return d.Render() }

// ApplyKey processes a key press, returning the updated Dialog plus optional
// chosen / cancelled flags. No-op when closed. Prev/Next move Selected
// (clamped at the ends), Confirm flags the choice, Cancel flags dismissal.
func (d Dialog) ApplyKey(key string) (Dialog, bool, bool) {
	if !d.Open {
		return d, false, false
	}
	switch {
	case contains(keys(d.Cancel, "esc"), key):
		return d, false, true
	case contains(keys(d.Confirm, "enter"), key):
		return d, len(d.Buttons) > 0, false
	case contains(keys(d.Prev, "left", "h"), key):
		d.Selected = clampStep(d.Selected, -1, len(d.Buttons))
	case contains(keys(d.Next, "right", "l"), key):
		d.Selected = clampStep(d.Selected, +1, len(d.Buttons))
	}
	return d, false, false
}

// Update routes a key message and returns the new Dialog plus a cmd carrying
// any event (ChosenMsg / CancelledMsg); the cmd is nil when nothing notable
// happened. It is a no-op while closed and ignores non-key messages.
func (d Dialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !d.Open {
		return d, nil
	}
	next, chosen, cancelled := d.ApplyKey(key.String())
	switch {
	case chosen:
		idx := next.Selected
		return next, func() tea.Msg { return ChosenMsg{Index: idx} }
	case cancelled:
		return next, func() tea.Msg { return CancelledMsg{} }
	}
	return next, nil
}

func keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func clampStep(cur, delta, n int) int {
	if n == 0 {
		return cur
	}
	next := cur + delta
	if next < 0 {
		return 0
	}
	if next >= n {
		return n - 1
	}
	return next
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
