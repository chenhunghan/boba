package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/meter"
)

func meterStories() Component {
	demo := meter.Style{
		Normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		LowFill:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		HighFill: lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		Empty:    muted,
	}

	return Component{Name: "meter", Stories: []Story{
		{Name: "fill levels", Render: func(w, h int) string {
			width := min(w, 24)
			rows := make([]string, 0, 3)
			for _, v := range []float64{20, 50, 90} {
				rows = append(rows, meter.Meter{
					Value: v, Max: 100, Width: width, Style: demo,
				}.Render())
			}
			return strings.Join(rows, "\n")
		}},
		{Name: "thresholds", Render: func(w, h int) string {
			width := min(w, 24)
			rows := make([]string, 0, 3)
			for _, v := range []float64{15, 50, 85} {
				rows = append(rows, meter.Meter{
					Value: v, Max: 100, Width: width,
					Low: 30, High: 70, Style: demo,
				}.Render())
			}
			return strings.Join(rows, "\n")
		}},
		{Name: "custom glyphs", Render: func(w, h int) string {
			return meter.Meter{
				Value: 6, Max: 10, Width: min(w, 24),
				Style: meter.Style{
					Normal:    bright,
					Empty:     muted,
					FillChar:  "#",
					EmptyChar: "-",
				},
			}.Render()
		}},
	}}
}
