package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/popover"
	"github.com/chenhunghan/boba/popup"
)

func popoverStories() Component {
	demoStyle := popover.Style{
		Surface: bright.Background(lipgloss.Color("236")),
		Border:  rule,
	}

	blank := func(w, h int) string {
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		row := strings.Repeat(" ", w)
		rows := make([]string, h)
		for i := range rows {
			rows[i] = row
		}
		return strings.Join(rows, "\n")
	}

	return Component{Name: "popover", Stories: []Story{
		{Name: "below anchor", Render: func(w, h int) string {
			p := popover.Popover{
				Content:   "save changes?\n[y] yes  [n] no",
				Open:      true,
				X:         2,
				Y:         1,
				W:         6,
				H:         1,
				Placement: popup.Below,
				Style:     demoStyle,
			}
			return p.Over(blank(w, h), w, h)
		}},
		{Name: "above anchor", Render: func(w, h int) string {
			p := popover.Popover{
				Content:   "tooltip",
				Open:      true,
				X:         3,
				Y:         h - 1,
				W:         4,
				H:         1,
				Placement: popup.Above,
				Style:     demoStyle,
			}
			return p.Over(blank(w, h), w, h)
		}},
		{Name: "closed", Render: func(w, h int) string {
			p := popover.Popover{Content: "hidden", Open: false, Style: demoStyle}
			return p.Over(blank(w, h), w, h)
		}},
	}}
}
