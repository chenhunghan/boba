package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/menu"
)

func menuStories() Component {
	demoStyle := func() menu.Style {
		surface := lipgloss.Color("236")
		return menu.Style{
			Inactive: muted.Background(surface),
			Hover:    chosen.Background(lipgloss.Color("238")),
			Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(surface),
			Border:   rule.Background(surface),
			Surface:  surface,
		}
	}

	items := func() []menu.Item[string] {
		return []menu.Item[string]{
			{ID: "copy", Label: "Copy"},
			{ID: "paste", Label: "Paste"},
			{ID: "rename", Label: "Rename"},
			{ID: "delete", Label: "Delete", Disabled: true},
		}
	}

	return Component{Name: "menu", Stories: []Story{
		{Name: "open", Render: func(w, h int) string {
			g := menu.Group[string]{
				Items:   items(),
				Open:    true,
				Hover:   1,
				AnchorX: 0,
				AnchorY: 0,
				Style:   demoStyle(),
			}
			return g.View()
		}},
		{Name: "no highlight", Render: func(w, h int) string {
			g := menu.Group[string]{
				Items:   items(),
				Open:    true,
				Hover:   -1,
				AnchorX: 0,
				AnchorY: 0,
				Style:   demoStyle(),
			}
			return g.View()
		}},
		{Name: "closed", Render: func(w, h int) string {
			g := menu.Group[string]{
				Items: items(),
				Open:  false,
				Style: demoStyle(),
			}
			return g.View()
		}},
	}}
}
