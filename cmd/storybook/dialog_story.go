package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/dialog"
)

func dialogStories() Component {
	demoStyle := func() dialog.Style {
		surface := lipgloss.Color("236")
		return dialog.Style{
			Surface:        surface,
			Title:          bright.Bold(true),
			Body:           muted,
			Border:         rule,
			Button:         muted.Background(surface),
			ButtonSelected: chosen.Background(lipgloss.Color("238")),
		}
	}

	blank := func(w, h int) string {
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		return strings.TrimRight(strings.Repeat(strings.Repeat(" ", w)+"\n", h), "\n")
	}

	return Component{Name: "dialog", Stories: []Story{
		{Name: "confirm", Render: func(w, h int) string {
			d := dialog.Dialog{
				Title:    "Quit?",
				Body:     "Discard changes?",
				Buttons:  []string{"OK", "Cancel"},
				Selected: 0,
				Open:     true,
				Style:    demoStyle(),
			}
			return d.Over(blank(w, h), w, h)
		}},
		{Name: "selected button", Render: func(w, h int) string {
			d := dialog.Dialog{
				Title:    "Delete file",
				Body:     "This cannot be undone.",
				Buttons:  []string{"Cancel", "Delete"},
				Selected: 1,
				Open:     true,
				Style:    demoStyle(),
			}
			return d.Over(blank(w, h), w, h)
		}},
		{Name: "no buttons", Render: func(w, h int) string {
			d := dialog.Dialog{
				Title: "Loading",
				Body:  "Please wait…",
				Open:  true,
				Style: demoStyle(),
			}
			return d.Over(blank(w, h), w, h)
		}},
	}}
}
