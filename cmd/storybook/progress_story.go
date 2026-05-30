package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/progress"
)

func progressStories() Component {
	demoStyle := progress.Style{
		Filled: lipgloss.NewStyle().Foreground(accent),
		Empty:  muted,
	}
	bar := func(value, max float64, width int, st progress.Style) string {
		return progress.Progress{Value: value, Max: max, Width: width, Style: st}.Render()
	}

	return Component{Name: "progress", Stories: []Story{
		{Name: "levels", Render: func(w, h int) string {
			width := min(w, 24)
			return strings.Join([]string{
				bar(0, 10, width, demoStyle),
				bar(3, 10, width, demoStyle),
				bar(7, 10, width, demoStyle),
				bar(10, 10, width, demoStyle),
			}, "\n")
		}},
		{Name: "custom glyphs", Render: func(w, h int) string {
			st := progress.Style{
				Filled:     rule,
				Empty:      muted,
				FilledChar: "#",
				EmptyChar:  ".",
			}
			return bar(4, 10, min(w, 24), st)
		}},
	}}
}
