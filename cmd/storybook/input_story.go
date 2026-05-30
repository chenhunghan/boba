package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/input"
)

func inputStories() Component {
	style := input.Style{
		Text:        bright,
		Placeholder: muted.Italic(true),
		Cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Background(accent),
	}

	return Component{Name: "input", Stories: []Story{
		{Name: "placeholder", Render: func(w, h int) string {
			m := input.Model{Placeholder: "search…", Style: style}
			return m.Render()
		}},
		{Name: "focused", Render: func(w, h int) string {
			m := input.Model{Value: "Ada", Cursor: 3, Focused: true, Style: style}
			return m.Render()
		}},
		{Name: "windowed", Render: func(w, h int) string {
			m := input.Model{
				Value:   "github.com/chenhunghan/boba",
				Cursor:  27,
				Focused: true,
				Width:   min(w, 12),
				Style:   style,
			}
			return m.Render()
		}},
	}}
}
