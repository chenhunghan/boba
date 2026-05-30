package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/tab"
)

func tabStories() Component {
	demoStyle := tab.Style{
		Inactive:    muted,
		Hover:       bright,
		Active:      chosen,
		Border:      muted,
		HoverBar:    bright,
		SelectedBar: rule,
	}

	newGroup := func(selected string) tab.Group[string] {
		return tab.Group[string]{
			Tabs: []tab.Tab[string]{
				{ID: "cpu", Label: "CPU", Icon: "▣", Style: demoStyle, Model: tab.Static("cpu graph")},
				{ID: "mem", Label: "Mem", Style: demoStyle, Model: tab.Static("memory graph")},
				{ID: "net", Label: "Net", Style: demoStyle, Closable: true, Model: tab.Static("network graph")},
			},
			Gap:      1,
			Selected: selected,
		}
	}

	return Component{Name: "tab", Stories: []Story{
		{Name: "selected", Render: func(w, h int) string {
			g := newGroup("cpu")
			return g.Render(min(w, 40), tab.RenderState[string]{})
		}},
		{Name: "hover", Render: func(w, h int) string {
			g := newGroup("cpu")
			return g.Render(min(w, 40), tab.RenderState[string]{Hover: "mem", HasHover: true})
		}},
		{Name: "header only", Render: func(w, h int) string {
			g := tab.Group[string]{
				Tabs: []tab.Tab[string]{
					{ID: "a", Label: "Overview", Style: demoStyle},
					{ID: "b", Label: "Details", Style: tab.Style{
						Inactive: lipgloss.NewStyle().Faint(true),
						Active:   chosen,
						Border:   muted,
					}},
				},
				Selected: "a",
			}
			return g.RenderHeader(min(w, 40), tab.RenderState[string]{})
		}},
	}}
}
