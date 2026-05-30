// Package button renders headless icon buttons: solid-fill rectangles
// with optional badge accents. The package owns *behavior* and
// *layout* — it produces correctly-sized strings for a button at a
// given state — but it does not own *styling*. Each Button carries
// its own Style; the package applies the right per-state variant
// based on the State the parent passes in. This is the pattern
// base-ui uses for React: behavior is part of the component, but
// appearance is composed in per-instance, so the same component can
// be reused with completely different looks.
//
// Single button:
//
//	b := button.Button{Text: "Edit", Style: mystyle}
//	out := b.Render(button.StateInactive, 0, 1)
//
// Vertical list of buttons (e.g., a sidebar or shortcut bar) — also supports
// hit-testing for click and hover handling:
//
//	stack := button.Stack{Buttons: buttons, Width: 5, ItemHeight: 1, Selected: 0, Hover: -1, Active: true}
//	out := stack.Render()
//	clickedIdx, area := stack.HitTest(panelLocalX, panelLocalY)
//
// Horizontal row of buttons (e.g., inline action buttons in a card):
//
//	out := button.HorizontalRow(buttons, hoverIdx, height, gap)
package button

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// State is the visual state a button is rendered in. Inactive is the
// default; Hover indicates the cursor is over the button; Active means
// the button is the chosen item in its parent (selected + parent has
// focus). Stacks compute the right State from their own fields and
// pass it to each Button.Render.
type State int

const (
	StateInactive State = iota
	StateHover
	StateActive
)

// Style holds one lipgloss.Style per State plus an optional Badge
// override. ForState returns the style for the given state.
//
// The styles are expected to fully define both foreground and
// background — the package applies them as-is to the button content,
// so they should not require any ambient styling to look right.
//
// Badge, if non-empty, overrides the foreground color of the badge
// portion only; the background comes from the state's style. Useful
// for keyboard-shortcut hints (a red ¹ on a colored button bg).
type Style struct {
	Inactive lipgloss.Style
	Hover    lipgloss.Style
	Active   lipgloss.Style
	Badge    lipgloss.Style
}

// ForState returns the lipgloss.Style for the given state.
func (s Style) ForState(state State) lipgloss.Style {
	switch state {
	case StateActive:
		return s.Active
	case StateHover:
		return s.Hover
	default:
		return s.Inactive
	}
}

// Button is a single icon/label button. Text is the visible label
// (typically 1 character for icon buttons, several for text buttons).
// Name is a human-readable label that callers can show elsewhere when
// the button is selected — the renderer never displays it. Badge is
// an optional accent shown before Text in the badge style. Trailing
// is an optional glyph anchored at the rightmost cell — used for
// inline affordances like a close-x; Stack.HitTest returns
// HitTrailing when a click lands on it. Style is owned by the button
// itself: each instance can have its own look.
type Button struct {
	Text     string
	Name     string
	Badge    string
	Trailing string
	Style    Style
}

// Render produces a string of exactly width × height visible cells
// representing the button in the given state, styled per the button's
// own Style.
//
// width <= 0 auto-sizes to fit the content (badge + text + 1 cell of
// padding on each side). height <= 0 defaults to 1.
//
// The implementation renders each segment of the line (left padding,
// badge, text, right padding) with explicit foreground+background
// from the state style, so the bg stays uniform across the whole
// row — there are no nested ANSI resets to break it.
func (b Button) Render(state State, width, height int) string {
	base := b.Style.ForState(state)

	trailingW := lipgloss.Width(b.Trailing)
	contentWidth := lipgloss.Width(b.Badge) + lipgloss.Width(b.Text) + trailingW
	if width <= 0 {
		width = contentWidth + 2
	}
	if height <= 0 {
		height = 1
	}

	padTotal := width - contentWidth
	if padTotal < 0 {
		padTotal = 0
	}
	leftPad := padTotal / 2
	rightPad := padTotal - leftPad

	var parts []string
	if leftPad > 0 {
		parts = append(parts, base.Render(strings.Repeat(" ", leftPad)))
	}
	if b.Badge != "" {
		badgeStyle := base
		if fg := b.Style.Badge.GetForeground(); fg != nil {
			badgeStyle = base.Foreground(fg)
		}
		parts = append(parts, badgeStyle.Render(b.Badge))
	}
	if b.Text != "" {
		parts = append(parts, base.Render(b.Text))
	}
	if rightPad > 0 {
		parts = append(parts, base.Render(strings.Repeat(" ", rightPad)))
	}
	if b.Trailing != "" {
		parts = append(parts, base.Render(b.Trailing))
	}
	line := strings.Join(parts, "")

	if height == 1 {
		return line
	}

	emptyRow := base.Render(strings.Repeat(" ", width))
	rows := make([]string, 0, height)
	topPad := (height - 1) / 2
	for i := 0; i < topPad; i++ {
		rows = append(rows, emptyRow)
	}
	rows = append(rows, line)
	for len(rows) < height {
		rows = append(rows, emptyRow)
	}
	return strings.Join(rows, "\n")
}

// Stack is a vertical list of buttons. The same struct backs both
// rendering and hit-testing — Render and HitTest read from the same
// fields, so they cannot drift out of sync as the layout evolves.
//
// Width must be set: Render needs it for per-button width, and
// HitTest needs it to know where the trailing glyph (if any) sits.
// Storing it on the Stack keeps the HitTest signature uniform with
// the rest of the design system: HitTest(x, y int) — same shape as
// navcard.Stack and tab.Group.
type Stack struct {
	Buttons    []Button
	Width      int  // outer width per button; required for rendering and hit-testing
	ItemHeight int  // rows per button (must be > 0 for Render to produce output)
	Selected   int  // index of the chosen button (visible only when Active)
	Hover      int  // index under the cursor; -1 = none
	Active     bool // whether the parent panel currently has focus

	// Keyboard bindings for Update (nil → sensible defaults): Up/Down
	// move Selected, Confirm activates it; Wrap cycles at the ends.
	Up, Down, Confirm []string
	Wrap              bool
}

// Render produces the stack as a multi-line string of Width cells per
// row. Each button is rendered with the State derived from Active,
// Selected, and Hover (priority: Active > Hover > Inactive). The
// Stack does not own button styling — each Button carries its own.
func (s Stack) Render() string {
	parts := make([]string, 0, len(s.Buttons))
	for i, btn := range s.Buttons {
		state := StateInactive
		switch {
		case s.Active && i == s.Selected:
			state = StateActive
		case i == s.Hover:
			state = StateHover
		}
		parts = append(parts, btn.Render(state, s.Width, s.ItemHeight))
	}
	return strings.Join(parts, "\n")
}

// HitArea identifies which region of a button was hit. HitNone is
// returned alongside Index == -1 when the click misses the stack
// entirely; HitBody for clicks on the button body; HitTrailing for
// clicks on the optional trailing glyph (only buttons with a
// non-empty Trailing have a HitTrailing region).
type HitArea int

const (
	HitNone HitArea = iota
	HitBody
	HitTrailing
)

// HitTest returns the index and area of the button at panel-local
// (x, y). Returns (-1, HitNone) if y is outside the stack. When the
// hit button has a non-empty Trailing and x falls within the
// rightmost trailingW cells of the Stack's Width, the area is
// HitTrailing; otherwise HitBody. Coordinates are stack-local —
// caller is responsible for translating from app-global coordinates.
func (s Stack) HitTest(x, y int) (int, HitArea) {
	if y < 0 || s.ItemHeight <= 0 {
		return -1, HitNone
	}
	idx := y / s.ItemHeight
	if idx >= len(s.Buttons) {
		return -1, HitNone
	}
	if t := s.Buttons[idx].Trailing; t != "" {
		trailingW := lipgloss.Width(t)
		if s.Width > 0 && trailingW > 0 && x >= s.Width-trailingW {
			return idx, HitTrailing
		}
	}
	return idx, HitBody
}

// ActivatedMsg is emitted — via the cmd from Update or ClickAt — when a
// button is activated (the Confirm key, or a click on its body). Index
// is the activated button.
type ActivatedMsg struct{ Index int }

// TrailingMsg is emitted — via the cmd from ClickAt — when a button's
// trailing glyph is clicked (e.g. a close/remove affordance). Index is
// the button whose trailing glyph was hit.
type TrailingMsg struct{ Index int }

// Update routes a key message to the stack: Up/Down move Selected and
// Confirm activates it (emitting ActivatedMsg). It is a no-op unless
// Active, and ignores non-key messages — mouse goes through ClickAt /
// HoverAt, which take panel-local coordinates.
//
//	case tea.KeyMsg:
//	    m.actions, cmd = m.actions.Update(msg)
//	    return m, cmd
//	case button.ActivatedMsg:
//	    return run(m, msg.Index), nil
func (s Stack) Update(msg tea.Msg) (Stack, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !s.Active || len(s.Buttons) == 0 {
		return s, nil
	}
	switch k := key.String(); {
	case contains(s.keys(s.Up, "up", "k"), k):
		s.Selected = step(s.Selected, -1, len(s.Buttons), s.Wrap)
	case contains(s.keys(s.Down, "down", "j"), k):
		s.Selected = step(s.Selected, +1, len(s.Buttons), s.Wrap)
	case contains(s.keys(s.Confirm, "enter"), k):
		return s, fire(ActivatedMsg{Index: s.Selected})
	}
	return s, nil
}

// ClickAt applies a click at panel-local (x, y) — typically the LocalX /
// LocalY from panel.HitTest. A hit on a button body selects it and emits
// ActivatedMsg; a hit on a trailing glyph emits TrailingMsg; a miss is a
// no-op.
func (s Stack) ClickAt(x, y int) (Stack, tea.Cmd) {
	switch idx, area := s.HitTest(x, y); area {
	case HitBody:
		s.Selected = idx
		return s, fire(ActivatedMsg{Index: idx})
	case HitTrailing:
		return s, fire(TrailingMsg{Index: idx})
	}
	return s, nil
}

// HoverAt sets Hover from panel-local (x, y), clearing it to -1 when the
// coordinate misses the stack.
func (s Stack) HoverAt(x, y int) Stack {
	if idx, area := s.HitTest(x, y); area != HitNone {
		s.Hover = idx
	} else {
		s.Hover = -1
	}
	return s
}

// View renders the stack — an alias for Render that matches the View()
// convention of Bubble Tea models and charmbracelet/bubbles components.
func (s Stack) View() string { return s.Render() }

func (s Stack) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func fire(msg tea.Msg) tea.Cmd { return func() tea.Msg { return msg } }

func step(cur, delta, n int, wrap bool) int {
	if n == 0 {
		return cur
	}
	next := cur + delta
	if wrap {
		return ((next % n) + n) % n
	}
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

// HorizontalRow renders buttons left-to-right with `gap` blank cells
// between them. Each button auto-sizes to its content (width=0) and
// uses its own Style. hover is the index under the cursor (-1 for
// none); there's no "selected" concept for an inline button row —
// those are typically momentary actions, not list items.
func HorizontalRow(buttons []Button, hover, height, gap int) string {
	if len(buttons) == 0 {
		return ""
	}
	if height <= 0 {
		height = 1
	}
	rendered := make([]string, 0, len(buttons)*2)
	for i, btn := range buttons {
		if i > 0 && gap > 0 {
			rendered = append(rendered, spacer(gap, height))
		}
		state := StateInactive
		if i == hover {
			state = StateHover
		}
		rendered = append(rendered, btn.Render(state, 0, height))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// spacer is a width × height rectangle of plain spaces. Used between
// buttons in HorizontalRow.
func spacer(width, height int) string {
	line := strings.Repeat(" ", width)
	if height == 1 {
		return line
	}
	rows := make([]string, height)
	for i := range rows {
		rows[i] = line
	}
	return strings.Join(rows, "\n")
}
