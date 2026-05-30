// Package input renders a headless single-line text field: a run of
// styled runes with an optional cursor and placeholder. The package owns
// behavior — rune insertion/deletion, cursor movement, and windowing the
// visible text to a fixed width — but every visual decision lives in
// caller-supplied styles. It works in runes, not bytes, so multi-byte
// input edits and windows correctly.
//
// A Model is a value type: store it on your model, route key messages
// through Update, and render with View. Reach for ApplyKey when you'd
// rather handle the result synchronously.
//
//	case tea.KeyMsg:
//	    m.field, cmd = m.field.Update(msg)
//	    return m, cmd
//	case input.ChangedMsg:
//	    return onChange(m, msg.Value), nil
package input

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the per-element styles for a text field. Text styles the
// content runes, Placeholder the placeholder shown when empty and
// unfocused, and Cursor the single cell under the cursor (or the trailing
// block past the end of text) while focused.
type Style struct {
	Text        lipgloss.Style
	Placeholder lipgloss.Style
	Cursor      lipgloss.Style
}

// Model is a single-line text field. Cursor is a rune index in
// [0, len(runes)]; Value, Cursor, and the limits are kept consistent by
// the editing methods. Focused reports whether the field owns keyboard
// input (the caller sets it); Update and the cursor render only apply
// when Focused. Width is the visible width in cells (<=0 = full); the
// visible runes are windowed to it with the cursor kept in view.
// CharLimit caps Value's rune length (<=0 = unlimited).
type Model struct {
	Value       string
	Placeholder string
	Cursor      int
	Focused     bool
	Width       int
	CharLimit   int
	Style       Style
}

// ChangedMsg is emitted, via the cmd from Update, when Value changes.
// Value is the new content.
type ChangedMsg struct{ Value string }

func (m Model) runes() []rune { return []rune(m.Value) }

// clamp keeps Cursor inside [0, len(runes)] — callers may set Value or
// Cursor directly, so render and editing both normalize before use.
func (m Model) clampCursor(n int) int {
	if m.Cursor < 0 {
		return 0
	}
	if m.Cursor > n {
		return n
	}
	return m.Cursor
}

// SetValue replaces the content, truncating to CharLimit and clamping the
// cursor into range.
func (m Model) SetValue(v string) Model {
	r := []rune(v)
	if m.CharLimit > 0 && len(r) > m.CharLimit {
		r = r[:m.CharLimit]
	}
	m.Value = string(r)
	m.Cursor = m.clampCursor(len(r))
	return m
}

// ApplyKey applies a single key to the field and returns the new Model
// plus whether Value changed. It edits only the cursor for movement keys
// and only Value+cursor for text keys; unhandled keys are a no-op. This
// is the pure core that Update wraps.
func (m Model) ApplyKey(key tea.KeyMsg) (Model, bool) {
	r := m.runes()
	m.Cursor = m.clampCursor(len(r))
	switch key.Type {
	case tea.KeyLeft:
		if m.Cursor > 0 {
			m.Cursor--
		}
	case tea.KeyRight:
		if m.Cursor < len(r) {
			m.Cursor++
		}
	case tea.KeyHome:
		m.Cursor = 0
	case tea.KeyEnd:
		m.Cursor = len(r)
	case tea.KeyBackspace:
		if m.Cursor > 0 {
			r = append(r[:m.Cursor-1], r[m.Cursor:]...)
			m.Cursor--
			return m.commit(r)
		}
	case tea.KeyDelete:
		if m.Cursor < len(r) {
			r = append(r[:m.Cursor], r[m.Cursor+1:]...)
			return m.commit(r)
		}
	case tea.KeyRunes, tea.KeySpace:
		return m.insert(r, key.Runes)
	}
	return m, false
}

// insert places ins at the cursor, honoring CharLimit, and advances the
// cursor past what was actually inserted.
func (m Model) insert(r, ins []rune) (Model, bool) {
	if len(ins) == 0 {
		return m, false
	}
	if m.CharLimit > 0 {
		room := m.CharLimit - len(r)
		if room <= 0 {
			return m, false
		}
		if len(ins) > room {
			ins = ins[:room]
		}
	}
	next := make([]rune, 0, len(r)+len(ins))
	next = append(next, r[:m.Cursor]...)
	next = append(next, ins...)
	next = append(next, r[m.Cursor:]...)
	m.Cursor += len(ins)
	return m.commit(next)
}

func (m Model) commit(r []rune) (Model, bool) {
	v := string(r)
	if v == m.Value {
		return m, false
	}
	m.Value = v
	return m, true
}

// Update routes a key message to the field and returns the new Model plus
// a cmd carrying a ChangedMsg when Value changed (nil otherwise). It is a
// no-op unless Focused, and ignores non-key messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !m.Focused {
		return m, nil
	}
	next, changed := m.ApplyKey(key)
	if !changed {
		return next, nil
	}
	v := next.Value
	return next, func() tea.Msg { return ChangedMsg{Value: v} }
}

// Render draws the field on a single row. When the field is empty and not
// focused, it shows the placeholder. The visible runes are windowed to
// Width (when > 0) with the cursor kept in view; when focused, the cell
// under the cursor (or a trailing block past the end) is drawn with
// Style.Cursor.
func (m Model) Render() string {
	r := m.runes()
	if len(r) == 0 && !m.Focused {
		return m.Style.Placeholder.Render(m.Placeholder)
	}

	cursor := m.clampCursor(len(r))
	// Slots are the display cells: each rune, plus a trailing block slot at
	// index len(r) when focused. Windowing over slots keeps the visible
	// width to exactly min(Width, slotCount) — block included.
	slots := len(r)
	if m.Focused {
		slots++
	}
	start, end := m.window(slots, cursor)

	out := ""
	for i := start; i < end; i++ {
		switch {
		case i == len(r): // trailing block (only reachable when focused)
			out += m.Style.Cursor.Render(" ")
		case m.Focused && i == cursor:
			out += m.Style.Cursor.Render(string(r[i]))
		default:
			out += m.Style.Text.Render(string(r[i]))
		}
	}
	return out
}

// window returns the [start, end) slot range to display so that cursor is
// visible. With Width <= 0 every slot is shown; otherwise the window is at
// most Width slots wide, scrolled to keep cursor in view (anchored to the
// right edge once the cursor passes Width).
func (m Model) window(slots, cursor int) (int, int) {
	if m.Width <= 0 || m.Width >= slots {
		return 0, slots
	}
	w := m.Width
	if cursor < w {
		return 0, w
	}
	end := cursor + 1
	if end > slots {
		end = slots
	}
	return end - w, end
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (m Model) View() string { return m.Render() }
