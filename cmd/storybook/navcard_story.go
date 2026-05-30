package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/navcard"
)

func navcardStories() Component {
	cardStyle := func() navcard.Style {
		stateStyle := func(barFg, bg lipgloss.Color) navcard.StateStyle {
			fill := lipgloss.NewStyle().Background(bg)
			return navcard.StateStyle{
				Bar:         lipgloss.NewStyle().Foreground(barFg).Background(bg),
				BarChar:     "▌",
				Fill:        fill,
				Title:       bright.Background(bg).Bold(true),
				Subtitle:    muted.Background(bg),
				Description: muted.Background(bg),
			}
		}
		return navcard.Style{
			Inactive: stateStyle(lipgloss.Color("240"), lipgloss.Color("235")),
			Hover:    stateStyle(lipgloss.Color("245"), lipgloss.Color("237")),
			Active:   stateStyle(accent, lipgloss.Color("237")),
		}
	}()

	btnStyle := button.Style{
		Inactive: muted.Background(lipgloss.Color("236")),
		Hover:    bright.Background(lipgloss.Color("238")),
		Active:   chosen.Background(lipgloss.Color("238")),
	}

	cards := func() []navcard.Card {
		return []navcard.Card{
			{
				Title:       "nginx",
				Subtitle:    "running · 12 MB",
				Description: "reverse proxy",
				Buttons: []button.Button{
					{Text: "Edit", Style: btnStyle},
					{Text: "Stop", Style: btnStyle},
				},
				Style: cardStyle,
			},
			{
				Title:    "redis",
				Subtitle: "stopped",
				Style:    cardStyle,
			},
			{
				Title:    "postgres",
				Subtitle: "running · 88 MB",
				Style:    cardStyle,
			},
		}
	}

	return Component{Name: "navcard", Stories: []Story{
		{Name: "focused selection", Render: func(w, h int) string {
			return navcard.Stack{
				Cards:       cards(),
				Width:       min(w, 30),
				Gap:         1,
				Selected:    0,
				Hover:       -1,
				HoverButton: -1,
				Active:      true,
			}.View()
		}},
		{Name: "hover with buttons", Render: func(w, h int) string {
			return navcard.Stack{
				Cards:       cards(),
				Width:       min(w, 30),
				Gap:         1,
				Selected:    0,
				Hover:       0,
				HoverButton: 1,
				Active:      false,
			}.View()
		}},
		{Name: "blurred", Render: func(w, h int) string {
			return navcard.Stack{
				Cards:       cards(),
				Width:       min(w, 30),
				Gap:         1,
				Selected:    0,
				Hover:       -1,
				HoverButton: -1,
				Active:      false,
			}.View()
		}},
	}}
}
