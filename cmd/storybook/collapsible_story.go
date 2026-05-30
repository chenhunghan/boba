package main

import (
	"strings"

	"github.com/chenhunghan/boba/collapsible"
)

func collapsibleStories() Component {
	style := collapsible.Style{
		Header: chosen,
		Body:   muted,
	}
	return Component{Name: "collapsible", Stories: []Story{
		{Name: "states", Render: func(w, h int) string {
			body := "first line\nsecond line"
			return strings.Join([]string{
				collapsible.Collapsible{Title: "Collapsed", Body: body, Style: style}.Render(),
				collapsible.Collapsible{Title: "Expanded", Body: body, Expanded: true, Style: style}.Render(),
			}, "\n")
		}},
		{Name: "focused header", Render: func(w, h int) string {
			focusStyle := collapsible.Style{Header: bright, Body: muted}
			return collapsible.Collapsible{
				Title:    "Network",
				Body:     "eth0  up\nwlan0 down",
				Expanded: true,
				Focused:  true,
				Style:    focusStyle,
			}.Render()
		}},
		{Name: "custom glyphs", Render: func(w, h int) string {
			glyphStyle := collapsible.Style{
				Header:      chosen,
				Body:        muted,
				OpenGlyph:   "▾",
				ClosedGlyph: "▸",
			}
			return strings.Join([]string{
				collapsible.Collapsible{Title: "Closed", Body: "hidden", Style: glyphStyle}.Render(),
				collapsible.Collapsible{Title: "Open", Body: "shown", Expanded: true, Style: glyphStyle}.Render(),
			}, "\n")
		}},
	}}
}
