// Package slider renders a headless horizontal value slider: a fixed-width
// track with a filled portion and a handle at the current value. Left/right
// keys (when Focused) and clicks adjust the value; the package owns that
// behavior while the caller owns styling and glyphs.
package slider

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's per-segment styles and glyph fallbacks. Track
// styles the unfilled run, Filled the run left of the handle, Handle the
// handle cell. TrackChar is the cell drawn for both track and filled runs
// (empty falls back to "─"); HandleChar is the handle cell (empty falls back
// to "●").
type Style struct {
	Track      lipgloss.Style
	Filled     lipgloss.Style
	Handle     lipgloss.Style
	TrackChar  string
	HandleChar string
}

// Slider is a horizontal value slider. Value is clamped to [Min, Max]; Step
// is the per-key increment. Width is the track length in cells. Focused
// reports whether the slider owns keyboard input (the caller sets it); Update
// only responds when Focused.
type Slider struct {
	Value   float64
	Min     float64
	Max     float64
	Step    float64
	Width   int
	Focused bool
	Style   Style

	// Left/Right are the keys that decrement/increment Value; nil falls
	// back to ["left", "h"] / ["right", "l"].
	Left  []string
	Right []string
}

// ChangedMsg is emitted, via the cmd from Update or ClickAt, when Value
// changes. Value is the new clamped value.
type ChangedMsg struct{ Value float64 }

func (s Slider) trackChar() string {
	if s.Style.TrackChar != "" {
		return s.Style.TrackChar
	}
	return "─"
}

func (s Slider) handleChar() string {
	if s.Style.HandleChar != "" {
		return s.Style.HandleChar
	}
	return "●"
}

// clamp constrains v to [Min, Max], guarding against an inverted range.
func (s Slider) clamp(v float64) float64 {
	lo, hi := s.Min, s.Max
	if hi < lo {
		lo, hi = hi, lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// fraction is Value's position in [0, 1] across the range; 0 when the range
// is empty.
func (s Slider) fraction() float64 {
	span := s.Max - s.Min
	if span == 0 {
		return 0
	}
	return (s.clamp(s.Value) - s.Min) / span
}

// handlePos is the handle's cell index in [0, Width-1].
func (s Slider) handlePos() int {
	if s.Width <= 1 {
		return 0
	}
	return int(math.Round(s.fraction() * float64(s.Width-1)))
}

// Render draws the track on a single row: the filled run, the handle, then
// the remaining track. Returns "" when Width <= 0.
func (s Slider) Render() string {
	if s.Width <= 0 {
		return ""
	}
	pos := s.handlePos()
	var b strings.Builder
	if pos > 0 {
		b.WriteString(s.Style.Filled.Render(strings.Repeat(s.trackChar(), pos)))
	}
	b.WriteString(s.Style.Handle.Render(s.handleChar()))
	if rest := s.Width - pos - 1; rest > 0 {
		b.WriteString(s.Style.Track.Render(strings.Repeat(s.trackChar(), rest)))
	}
	return b.String()
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (s Slider) View() string { return s.Render() }

// ValueAt maps a track-local x (0..Width-1) to a clamped, Step-snapped value.
func (s Slider) ValueAt(x int) float64 {
	if s.Width <= 1 {
		return s.clamp(s.Value)
	}
	if x < 0 {
		x = 0
	}
	if x > s.Width-1 {
		x = s.Width - 1
	}
	f := float64(x) / float64(s.Width-1)
	return s.snap(s.Min + f*(s.Max-s.Min))
}

// snap rounds v to the nearest Step (relative to Min) before clamping; a
// non-positive Step leaves v unrounded.
func (s Slider) snap(v float64) float64 {
	if s.Step > 0 {
		v = s.Min + math.Round((v-s.Min)/s.Step)*s.Step
	}
	return s.clamp(v)
}

// ApplyKey is the pure key handler Update wraps. It returns the updated
// slider and whether Value changed; it is a no-op for keys it does not bind.
func (s Slider) ApplyKey(key string) (Slider, bool) {
	switch {
	case contains(s.keys(s.Left, "left", "h"), key):
		return s.setValue(s.clamp(s.Value - s.Step))
	case contains(s.keys(s.Right, "right", "l"), key):
		return s.setValue(s.clamp(s.Value + s.Step))
	}
	return s, false
}

// Update adjusts Value on a Left/Right key and emits ChangedMsg. It is a
// no-op unless Focused, and ignores non-key messages.
func (s Slider) Update(msg tea.Msg) (Slider, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !s.Focused {
		return s, nil
	}
	s, changed := s.ApplyKey(key.String())
	if !changed {
		return s, nil
	}
	return s, fire(s.Value)
}

// ClickAt sets Value from a click at panel-local (x, y) on the track's single
// row, emitting ChangedMsg when the value changes. A click off the row is a
// no-op.
func (s Slider) ClickAt(x, y int) (Slider, tea.Cmd) {
	if y != 0 || x < 0 || x >= s.Width {
		return s, nil
	}
	s, changed := s.setValue(s.ValueAt(x))
	if !changed {
		return s, nil
	}
	return s, fire(s.Value)
}

// HoverAt sets Value while dragging over the track's single row, mirroring
// ClickAt; a hover off the row is a no-op.
func (s Slider) HoverAt(x, y int) (Slider, tea.Cmd) {
	return s.ClickAt(x, y)
}

func (s Slider) setValue(v float64) (Slider, bool) {
	if v == s.Value {
		return s, false
	}
	s.Value = v
	return s, true
}

func (s Slider) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func fire(v float64) tea.Cmd { return func() tea.Msg { return ChangedMsg{Value: v} } }

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
