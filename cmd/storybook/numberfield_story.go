package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/numberfield"
)

func numberfieldStories() Component {
	upStyle := lipgloss.NewStyle().Foreground(accent)
	downStyle := muted

	build := func(v float64, up, down string) string {
		return numberfield.Number{
			Min: 0, Max: 10, Step: 1,
			Style: numberfield.Style{
				Up: up, Down: down,
				UpStyle: upStyle, DownStyle: downStyle,
			},
		}.SetValue(v).View()
	}

	return Component{Name: "numberfield", Stories: []Story{
		{Name: "arrows", Render: func(w, h int) string {
			return build(3, "", "")
		}},
		{Name: "plus/minus", Render: func(w, h int) string {
			return build(5, "+", "-")
		}},
		{Name: "clamped at max", Render: func(w, h int) string {
			return strings.Join([]string{
				bright.Render("value clamped to Max"),
				build(99, "+", "-"),
			}, "\n")
		}},
	}}
}
