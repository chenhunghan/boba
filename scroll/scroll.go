// Package scroll is a headless vertical viewport over multi-line content.
// It owns the scroll behavior — clamping the offset and clipping content to
// a fixed height, plus the keys that move the offset — while the caller owns
// styling via a per-instance Style. The offset is plain state the caller
// reads off the struct; nothing is emitted.
//
//	s := scroll.Scroll{Height: 10, Focused: true}
//	s, _ = s.Update(msg)        // up/down/pgup/pgdn/home/end move s.Offset
//	out := s.Render(longText)   // 10 lines starting at the clamped offset
package scroll

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Scroll is a vertical viewport. Offset is the index of the first visible
// line; Height is the number of visible rows. Focused reports whether the
// viewport owns keyboard input (the caller sets it); Update only responds
// when Focused. Style is applied to the clipped block.
type Scroll struct {
	Offset  int
	Height  int
	Focused bool
	Style   lipgloss.Style

	// Key bindings (defaults applied internally when nil): Up/Down move by
	// one line, PageUp/PageDown by Height, Home/End jump to the extremes.
	Up, Down, PageUp, PageDown, Home, End []string
}

func lines(content string) []string {
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}

// MaxOffset is the largest in-range Offset for content: the line count minus
// Height, floored at 0. Offsets above it would show blank rows past the end.
func (s Scroll) MaxOffset(content string) int {
	max := len(lines(content)) - s.Height
	if max < 0 {
		return 0
	}
	return max
}

func clamp(v, max int) int {
	if v < 0 {
		return 0
	}
	if v > max {
		return max
	}
	return v
}

// Render returns Height lines of content beginning at the clamped Offset,
// styled by Style. Offset is clamped to [0, MaxOffset] for this call without
// mutating the struct, so out-of-range stored offsets still render sanely.
func (s Scroll) Render(content string) string {
	if s.Height <= 0 {
		return ""
	}
	ls := lines(content)
	start := clamp(s.Offset, s.MaxOffset(content))
	end := start + s.Height
	if end > len(ls) {
		end = len(ls)
	}
	return s.Style.Render(strings.Join(ls[start:end], "\n"))
}

// ApplyKey moves Offset for a navigation key, clamped to [0, MaxOffset]. It
// is the pure handler Update wraps; keys it doesn't recognize return Scroll
// unchanged. content is needed to resolve the lower-bound jumps (end) and the
// clamp ceiling.
func (s Scroll) ApplyKey(key, content string) Scroll {
	max := s.MaxOffset(content)
	switch {
	case contains(orDefault(s.Up, "up", "k"), key):
		s.Offset = clamp(s.Offset-1, max)
	case contains(orDefault(s.Down, "down", "j"), key):
		s.Offset = clamp(s.Offset+1, max)
	case contains(orDefault(s.PageUp, "pgup"), key):
		s.Offset = clamp(s.Offset-s.Height, max)
	case contains(orDefault(s.PageDown, "pgdown"), key):
		s.Offset = clamp(s.Offset+s.Height, max)
	case contains(orDefault(s.Home, "home", "g"), key):
		s.Offset = 0
	case contains(orDefault(s.End, "end", "G"), key):
		s.Offset = max
	}
	return s
}

// Update routes a key message through ApplyKey and returns the new Scroll.
// It is a no-op unless Focused and ignores non-key messages. The returned
// cmd is always nil — scroll emits no events; the caller reads Offset
// directly. content is the same string the caller passes to Render, needed
// to clamp the offset.
//
//	case tea.KeyMsg:
//	    m.scroll, _ = m.scroll.Update(msg, m.content)
//	    return m, nil
func (s Scroll) Update(msg tea.Msg, content string) (Scroll, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !s.Focused {
		return s, nil
	}
	return s.ApplyKey(key.String(), content), nil
}

func orDefault(binding []string, def ...string) []string {
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
