package main

import (
	"strings"

	"github.com/chenhunghan/boba/scrollarea"
)

func scrollareaStories() Component {
	demoStyle := scrollarea.Style{
		Bar:   muted,
		Thumb: rule,
	}

	lines := func(n int) string {
		rows := make([]string, n)
		for i := range rows {
			rows[i] = "line " + string(rune('0'+i%10))
		}
		return strings.Join(rows, "\n")
	}

	return Component{Name: "scrollarea", Stories: []Story{
		{Name: "top", Render: func(w, h int) string {
			a := scrollarea.ScrollArea{
				Content: lines(min(h, 8) + 6),
				Height:  min(h, 8),
				Width:   min(w, 16),
				Style:   demoStyle,
			}
			return a.Render()
		}},
		{Name: "scrolled", Render: func(w, h int) string {
			a := scrollarea.ScrollArea{
				Content: lines(min(h, 8) + 6),
				Height:  min(h, 8),
				Width:   min(w, 16),
				Focused: true,
				Style:   demoStyle,
			}
			a.Scroll.Offset = a.MaxOffset() / 2
			return a.Render()
		}},
		{Name: "custom glyphs", Render: func(w, h int) string {
			a := scrollarea.ScrollArea{
				Content: lines(min(h, 8) + 6),
				Height:  min(h, 8),
				Width:   min(w, 16),
				Style: scrollarea.Style{
					Bar:       muted,
					Thumb:     chosen,
					BarChar:   "┊",
					ThumbChar: "┃",
				},
			}
			a.Scroll.Offset = a.MaxOffset()
			return a.Render()
		}},
	}}
}
