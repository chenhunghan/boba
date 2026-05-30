package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/toggle"
)

func toggleStories() Component {
	style := toggle.Style{
		Released: muted,
		Pressed:  chosen,
		Focused:  bright.Foreground(accent).Bold(true),
	}
	return Component{Name: "toggle", Stories: []Story{
		{Name: "states", Render: func(w, h int) string {
			return strings.Join([]string{
				toggle.Toggle{Label: "Auto-scroll", Style: style}.Render(),
				toggle.Toggle{Label: "Auto-scroll", Pressed: true, Style: style}.Render(),
				toggle.Toggle{Label: "Auto-scroll", Focused: true, Style: style}.Render(),
			}, "\n")
		}},
		{Name: "framed", Render: func(w, h int) string {
			frame := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accent).
				Padding(0, 1)
			return frame.Render(toggle.Toggle{Label: "Show graph", Pressed: true, Style: style}.Render())
		}},
	}}
}
