package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/toolbar"
)

func toolbarStories() Component {
	demo := toolbar.Style{
		Inactive: muted.Background(lipgloss.Color("236")),
		Hover:    bright.Background(lipgloss.Color("238")),
		Active:   chosen.Background(lipgloss.Color("238")),
	}

	return Component{Name: "toolbar", Stories: []Story{
		{Name: "focused selection", Render: func(w, h int) string {
			return toolbar.Toolbar{
				Items:    []string{"New", "Open", "Save"},
				Selected: 1,
				Hover:    -1,
				Focused:  true,
				Gap:      1,
				Style:    demo,
			}.View()
		}},
		{Name: "hover", Render: func(w, h int) string {
			return toolbar.Toolbar{
				Items:    []string{"New", "Open", "Save"},
				Selected: 0,
				Hover:    2,
				Focused:  false,
				Gap:      1,
				Style:    demo,
			}.View()
		}},
		{Name: "blurred", Render: func(w, h int) string {
			return toolbar.Toolbar{
				Items:    []string{"New", "Open", "Save"},
				Selected: 1,
				Hover:    -1,
				Focused:  false,
				Gap:      1,
				Style:    demo,
			}.View()
		}},
	}}
}
