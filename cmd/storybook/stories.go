package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/checkbox"
)

// Story is one named variant of a component, rendered into a w×h area.
type Story struct {
	Name   string
	Render func(w, h int) string
}

// Component is a catalog entry: a package name and its stories.
type Component struct {
	Name    string
	Stories []Story
}

var (
	accent = lipgloss.Color("63")
	muted  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	bright = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	chosen = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true)
	rule   = lipgloss.NewStyle().Foreground(accent)

	sidebarBtn = button.Style{Inactive: muted, Hover: bright, Active: chosen}
	cbStyle    = checkbox.Style{Normal: muted, Focused: chosen}
	demoBtn    = button.Style{
		Inactive: muted.Background(lipgloss.Color("236")),
		Hover:    bright.Background(lipgloss.Color("238")),
		Active:   chosen.Background(lipgloss.Color("238")),
	}
)

func catalog() []Component {
	return []Component{
		separatorStories(), checkboxStories(), buttonStories(), boxStories(),
		statusbarStories(), glyphStories(),
		inputStories(), scrollStories(), switchStories(), toggleStories(),
		togglegroupStories(), radiogroupStories(), checkboxgroupStories(),
		sliderStories(), progressStories(), meterStories(), collapsibleStories(),
		accordionStories(), toolbarStories(), tooltipStories(), popoverStories(),
		dialogStories(), selectboxStories(), numberfieldStories(), scrollareaStories(),
		tabStories(), navcardStories(), menuStories(),
	}
}
