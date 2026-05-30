package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/togglegroup"
)

func togglegroupStories() Component {
	style := togglegroup.Style{
		Inactive: muted.Padding(0, 1),
		Hover:    bright.Background(lipgloss.Color("238")).Padding(0, 1),
		Active:   chosen.Background(lipgloss.Color("238")).Padding(0, 1),
	}

	return Component{Name: "togglegroup", Stories: []Story{
		{Name: "selected", Render: func(w, h int) string {
			return togglegroup.Group{
				Items:    []string{"Day", "Week", "Month"},
				Selected: 1,
				Gap:      1,
				Hover:    -1,
				Style:    style,
			}.Render()
		}},
		{Name: "hover", Render: func(w, h int) string {
			return togglegroup.Group{
				Items:    []string{"Day", "Week", "Month"},
				Selected: 0,
				Gap:      1,
				Hover:    2,
				Style:    style,
			}.Render()
		}},
		{Name: "no gap", Render: func(w, h int) string {
			return togglegroup.Group{
				Items:    []string{"Off", "On", "Auto"},
				Selected: 2,
				Hover:    -1,
				Style:    style,
			}.Render()
		}},
	}}
}
