// Package scrollarea renders a headless scrollable viewport: multi-line
// content clipped to a fixed Height with a one-cell scrollbar on the right
// edge whose thumb size and position reflect the scroll offset. It composes
// scroll.Scroll for the offset behavior and the navigation keys, and owns the
// scrollbar layout; the caller owns every style and glyph.
//
//	a := scrollarea.ScrollArea{Content: longText, Height: 10, Width: 40, Focused: true}
//	a, _ = a.Update(msg) // up/down/pgup/pgdn/home/end move the offset
//	out := a.Render()     // 10 rows of content + a right-edge scrollbar
package scrollarea

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/scroll"
)

// Style holds the caller's scrollbar styles and glyph fallbacks. Bar styles
// the track (the unfilled run of the scrollbar); Thumb styles the draggable
// indicator. BarChar is the track cell (empty falls back to "│"); ThumbChar
// is the thumb cell (empty falls back to "█").
type Style struct {
	Bar       lipgloss.Style
	Thumb     lipgloss.Style
	BarChar   string
	ThumbChar string
}

func (s Style) barChar() string {
	if s.BarChar != "" {
		return s.BarChar
	}
	return "│"
}

func (s Style) thumbChar() string {
	if s.ThumbChar != "" {
		return s.ThumbChar
	}
	return "█"
}

// ScrollArea is a scrollable viewport with a scrollbar. Content is the
// multi-line text; Height is the visible row count and Width the total cell
// width (content plus the one-cell scrollbar). Focused reports whether the
// area owns keyboard input (the caller sets it); Update only responds when
// Focused. Style covers the scrollbar; the embedded Scroll carries the offset,
// the content clip style, and the navigation key bindings.
type ScrollArea struct {
	Content string
	Height  int
	Width   int
	Focused bool
	Style   Style

	// Scroll holds the offset state and navigation keys. Height and Focused
	// are mirrored onto it by Render/Update, so callers set them on the
	// ScrollArea, not here.
	Scroll scroll.Scroll
}

// Offset is the index of the first visible content line.
func (a ScrollArea) Offset() int { return a.Scroll.Offset }

// MaxOffset is the largest in-range offset for the current Content and Height.
func (a ScrollArea) MaxOffset() int {
	return a.scroll().MaxOffset(a.Content)
}

// scroll returns the embedded Scroll with Height and Focused synced from the
// ScrollArea, so the two cannot drift apart.
func (a ScrollArea) scroll() scroll.Scroll {
	s := a.Scroll
	s.Height = a.Height
	s.Focused = a.Focused
	return s
}

// thumb returns the scrollbar thumb's size and top position in rows. The size
// is proportional to the visible fraction of the content (at least one row);
// the position spreads the offset across the free travel. When all content
// fits, the thumb fills the whole bar.
func (a ScrollArea) thumb() (size, pos int) {
	total := len(strings.Split(a.Content, "\n"))
	if a.Content == "" {
		total = 0
	}
	if total <= a.Height || a.Height <= 0 {
		return a.Height, 0
	}
	size = int(math.Round(float64(a.Height) * float64(a.Height) / float64(total)))
	if size < 1 {
		size = 1
	}
	if size > a.Height {
		size = a.Height
	}
	max := a.MaxOffset()
	travel := a.Height - size
	if max <= 0 || travel <= 0 {
		return size, 0
	}
	off := a.Offset()
	if off < 0 {
		off = 0
	}
	if off > max {
		off = max
	}
	pos = int(math.Round(float64(travel) * float64(off) / float64(max)))
	if pos > travel {
		pos = travel
	}
	return size, pos
}

// bar renders the one-cell-wide scrollbar as Height rows: thumb cells over the
// thumb's span, track cells elsewhere.
func (a ScrollArea) bar() string {
	size, pos := a.thumb()
	rows := make([]string, a.Height)
	for i := range rows {
		if i >= pos && i < pos+size {
			rows[i] = a.Style.Thumb.Render(a.Style.thumbChar())
		} else {
			rows[i] = a.Style.Bar.Render(a.Style.barChar())
		}
	}
	return strings.Join(rows, "\n")
}

// Render returns Height rows of Width cells: Content clipped to the viewport on
// the left and the scrollbar on the rightmost column. Returns "" when Height or
// Width is non-positive. With Width == 1 only the scrollbar is drawn.
func (a ScrollArea) Render() string {
	if a.Height <= 0 || a.Width <= 0 {
		return ""
	}
	bar := a.bar()
	if a.Width == 1 {
		return bar
	}
	contentW := a.Width - 1
	body := strings.Split(a.scroll().Render(a.Content), "\n")
	barRows := strings.Split(bar, "\n")
	rows := make([]string, a.Height)
	for i := range rows {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		rows[i] = fit(line, contentW) + barRows[i]
	}
	return strings.Join(rows, "\n")
}

// fit truncates line to width cells and right-pads it with spaces to exactly
// width, preserving any styling already on line.
func fit(line string, width int) string {
	line = lipgloss.NewStyle().MaxWidth(width).Render(line)
	if pad := width - lipgloss.Width(line); pad > 0 {
		line += strings.Repeat(" ", pad)
	}
	return line
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (a ScrollArea) View() string { return a.Render() }

// Update forwards a navigation key to the embedded Scroll, moving the offset.
// It is a no-op unless Focused and ignores non-key messages. The returned cmd
// is always nil — scroll emits no events; the caller reads Offset directly.
//
//	case tea.KeyMsg:
//	    m.area, _ = m.area.Update(msg)
//	    return m, nil
func (a ScrollArea) Update(msg tea.Msg) (ScrollArea, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); !ok || !a.Focused {
		return a, nil
	}
	s, _ := a.scroll().Update(msg, a.Content)
	a.Scroll.Offset = s.Offset
	return a, nil
}
