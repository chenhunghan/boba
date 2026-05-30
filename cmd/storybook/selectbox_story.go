package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/selectbox"
)

func selectboxStories() Component {
	demoStyle := func() selectbox.Style {
		return selectbox.Style{
			Chevron:     "▾",
			Closed:      muted.Background(lipgloss.Color("236")),
			Option:      muted.Background(lipgloss.Color("236")),
			Highlighted: chosen.Background(lipgloss.Color("238")),
		}
	}

	background := func(top string, w, h int) string {
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		blank := strings.Repeat(" ", w)
		rows := make([]string, h)
		rows[0] = top
		for i := 1; i < h; i++ {
			rows[i] = blank
		}
		return strings.Join(rows, "\n")
	}

	return Component{Name: "selectbox", Stories: []Story{
		{Name: "closed", Render: func(w, h int) string {
			s := selectbox.SelectBox{
				Options:  []string{"Low", "Medium", "High"},
				Selected: 1,
				W:        12,
				Style:    demoStyle(),
			}
			return s.View()
		}},
		{Name: "open", Render: func(w, h int) string {
			s := selectbox.SelectBox{
				Options:   []string{"Low", "Medium", "High"},
				Selected:  1,
				Open:      true,
				Highlight: 2,
				W:         12,
				H:         1,
				Style:     demoStyle(),
			}
			bg := background(s.View(), w, h)
			return s.OpenView(bg, w, h)
		}},
	}}
}
