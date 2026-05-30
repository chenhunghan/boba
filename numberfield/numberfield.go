// Package numberfield renders a headless numeric field with steppers: a
// composed text input for typing the number plus up/down glyphs that step
// Value by Step, clamped to [Min, Max]. The package owns behavior — parsing
// typed text back to a number, stepping, clamping, and hit-testing the
// stepper zones — but every visual decision lives in caller-supplied styles
// and glyph strings.
//
// A Number is a value type: store it on your model, route key messages
// through Update, mouse through ClickAt/HoverAt, and render with View. Reach
// for ApplyKey / HitTest when you'd rather handle the result synchronously.
//
//	case tea.KeyMsg:
//	    m.qty, cmd = m.qty.Update(msg)
//	    return m, cmd
//	case numberfield.ChangedMsg:
//	    return onChange(m, msg.Value), nil
package numberfield

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/input"
)

// Style holds the stepper glyphs and their styles. The value text is styled
// through the embedded Input's own Style. Up / Down are the stepper glyph
// strings (empty falls back to "▲" / "▼"); UpStyle / DownStyle style them.
type Style struct {
	UpStyle   lipgloss.Style
	DownStyle lipgloss.Style
	Up        string
	Down      string
}

func (s Style) up() string {
	if s.Up != "" {
		return s.Up
	}
	return "▲"
}

func (s Style) down() string {
	if s.Down != "" {
		return s.Down
	}
	return "▼"
}

// Number is a numeric field with steppers. Value is the current number, kept
// within [Min, Max] by the stepping/parsing methods. Step is the increment
// for up/down (a non-positive Step falls back to 1). Focused reports whether
// the field owns keyboard input (the caller sets it); Update and the embedded
// cursor render only apply when Focused. Input is the composed text field
// that backs typing — its Value is the source of typed text, parsed back into
// Value on edit.
type Number struct {
	Value   float64
	Min     float64
	Max     float64
	Step    float64
	Focused bool
	Input   input.Model
	Style   Style

	// Increment / Decrement are the keys that step Value; nil falls back to
	// up/down. Set e.g. {"up", "+"} to also accept "+".
	Increment []string
	Decrement []string
}

// ChangedMsg is emitted, via the cmd from Update or ClickAt, when Value
// changes. Value is the new number.
type ChangedMsg struct{ Value float64 }

func (n Number) step() float64 {
	if n.Step <= 0 {
		return 1
	}
	return n.Step
}

// clamp keeps v within [Min, Max]; an inverted or zero range (Max <= Min)
// leaves v untouched so a caller that sets no bounds is unconstrained.
func (n Number) clamp(v float64) float64 {
	if n.Max > n.Min {
		if v < n.Min {
			return n.Min
		}
		if v > n.Max {
			return n.Max
		}
	}
	return v
}

// format renders Value as the canonical text shown in the input. -1 precision
// yields the shortest round-trippable form, so whole numbers read as "3" not
// "3.000000".
func (n Number) format() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

// SetValue clamps v into range, stores it, and syncs the input's text.
func (n Number) SetValue(v float64) Number {
	n.Value = n.clamp(v)
	n.Input = n.Input.SetValue(n.format())
	return n
}

// stepped returns n with Value moved by delta*Step (clamped) and the input
// text resynced, plus whether Value actually changed.
func (n Number) stepped(delta float64) (Number, bool) {
	next := n.clamp(n.Value + delta*n.step())
	if next == n.Value {
		// Keep the input text canonical even when clamped at a bound.
		n.Input = n.Input.SetValue(n.format())
		return n, false
	}
	return n.SetValue(next), true
}

// sync parses the input's current text into Value (clamped) and reports
// whether Value changed. Unparseable text (empty, "-", a lone ".") leaves
// Value untouched so the caller can keep typing.
func (n Number) sync() (Number, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(n.Input.Value), 64)
	if err != nil {
		return n, false
	}
	v = n.clamp(v)
	if v == n.Value {
		return n, false
	}
	n.Value = v
	return n, true
}

// ApplyKey applies a single key and returns the new Number plus whether Value
// changed. Increment/Decrement keys step Value; every other key is routed to
// the embedded input and its text reparsed into Value. This is the pure core
// that Update wraps.
func (n Number) ApplyKey(key tea.KeyMsg) (Number, bool) {
	k := key.String()
	if contains(n.keys(n.Increment, "up"), k) {
		return n.stepped(+1)
	}
	if contains(n.keys(n.Decrement, "down"), k) {
		return n.stepped(-1)
	}
	field, _ := n.Input.ApplyKey(key)
	n.Input = field
	return n.sync()
}

// Update routes a key message to the field and returns the new Number plus a
// cmd carrying a ChangedMsg when Value changed (nil otherwise). It is a no-op
// unless Focused, and ignores non-key messages.
func (n Number) Update(msg tea.Msg) (Number, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !n.Focused {
		return n, nil
	}
	next, changed := n.ApplyKey(key)
	if !changed {
		return next, nil
	}
	return next, fire(next.Value)
}

// Zone identifies a clickable region of the field.
type Zone int

const (
	// ZoneNone: outside the field, or on the value area.
	ZoneNone Zone = iota
	// ZoneUp: the up stepper glyph.
	ZoneUp
	// ZoneDown: the down stepper glyph.
	ZoneDown
)

// field returns the embedded input with Focused synced from the Number, so
// the value text and every width derived from it agree on the cursor block.
func (n Number) field() input.Model {
	f := n.Input
	f.Focused = n.Focused
	return f
}

// inputWidth is the width reserved for the value text. It honors the embedded
// input's fixed Width when set; otherwise it measures the rendered text
// (including the focused trailing cursor block, so steppers don't shift).
func (n Number) inputWidth() int {
	if n.Input.Width > 0 {
		return n.Input.Width
	}
	return lipgloss.Width(n.field().Render())
}

// Width is the rendered width in cells (value + both stepper glyphs) — the
// caller needs it to route clicks, which take panel-local coordinates.
func (n Number) Width() int {
	return n.inputWidth() + lipgloss.Width(n.Style.up()) + lipgloss.Width(n.Style.down())
}

// HitTest returns the stepper zone at panel-local (x, y). The steppers sit at
// the right of the field: the up glyph, then the down glyph. Returns ZoneNone
// for the value area or any miss.
func (n Number) HitTest(x, y int) Zone {
	if y != 0 || x < 0 {
		return ZoneNone
	}
	upW := lipgloss.Width(n.Style.up())
	downW := lipgloss.Width(n.Style.down())
	upStart := n.inputWidth()
	downStart := upStart + upW
	switch {
	case x >= upStart && x < upStart+upW:
		return ZoneUp
	case x >= downStart && x < downStart+downW:
		return ZoneDown
	}
	return ZoneNone
}

// ClickAt steps Value when (x, y) lands on a stepper zone, emitting a
// ChangedMsg via the returned cmd; a miss (or a step clamped at a bound) is a
// no-op. Coordinates are panel-local — typically the LocalX/LocalY from
// panel.HitTest.
func (n Number) ClickAt(x, y int) (Number, tea.Cmd) {
	var delta float64
	switch n.HitTest(x, y) {
	case ZoneUp:
		delta = +1
	case ZoneDown:
		delta = -1
	default:
		return n, nil
	}
	next, changed := n.stepped(delta)
	if !changed {
		return next, nil
	}
	return next, fire(next.Value)
}

// HoverAt takes panel-local coordinates for parity with other pointer-driven
// components. The field has no hover state, so it returns n unchanged; it
// exists so callers can wire HoverAt uniformly across components.
func (n Number) HoverAt(_, _ int) Number { return n }

// Render draws the field on a single row: the value text followed by the up
// and down stepper glyphs.
func (n Number) Render() string {
	return n.field().Render() +
		n.Style.UpStyle.Render(n.Style.up()) +
		n.Style.DownStyle.Render(n.Style.down())
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (n Number) View() string { return n.Render() }

func (n Number) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func fire(v float64) tea.Cmd {
	return func() tea.Msg { return ChangedMsg{Value: v} }
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
