package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/slider"
)

func sliderStories() Component {
	demoStyle := slider.Style{
		Track:  muted,
		Filled: lipgloss.NewStyle().Foreground(accent),
		Handle: chosen,
	}

	build := func(value float64, focused bool) slider.Slider {
		return slider.Slider{
			Value:   value,
			Min:     0,
			Max:     10,
			Step:    1,
			Width:   21,
			Focused: focused,
			Style:   demoStyle,
		}
	}

	return Component{Name: "slider", Stories: []Story{
		{Name: "values", Render: func(w, h int) string {
			return strings.Join([]string{
				build(0, false).Render(),
				build(5, false).Render(),
				build(10, false).Render(),
			}, "\n")
		}},
		{Name: "focused", Render: func(w, h int) string {
			return build(3, true).Render()
		}},
		{Name: "custom glyphs", Render: func(w, h int) string {
			s := build(6, false)
			s.Style.TrackChar = "·"
			s.Style.HandleChar = "◆"
			return s.Render()
		}},
	}}
}
