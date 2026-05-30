package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/popup"
	"github.com/chenhunghan/boba/tooltip"
)

func tooltipStories() Component {
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

	demoStyle := tooltip.Style{
		Surface: lipgloss.NewStyle().
			Foreground(bright.GetForeground()).
			Background(lipgloss.Color("238")),
	}

	return Component{Name: "tooltip", Stories: []Story{
		{Name: "below anchor", Render: func(w, h int) string {
			t := tooltip.Tooltip{
				Content: "rename",
				Visible: true,
				X:       2, Y: 1, W: 6, H: 1,
				Placement: popup.Below,
				Style:     demoStyle,
			}
			return t.Over(blank(w, h), w, h)
		}},
		{Name: "above anchor", Render: func(w, h int) string {
			t := tooltip.Tooltip{
				Content: "delete",
				Visible: true,
				X:       3, Y: h - 2, W: 6, H: 1,
				Placement: popup.Above,
				Style:     demoStyle,
			}
			return t.Over(blank(w, h), w, h)
		}},
		{Name: "multiline", Render: func(w, h int) string {
			t := tooltip.Tooltip{
				Content: "save changes\nesc to cancel",
				Visible: true,
				X:       2, Y: 1, W: 4, H: 1,
				Placement: popup.Below,
				Style:     demoStyle,
			}
			return t.Over(blank(w, h), w, h)
		}},
	}}
}
