package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/box"
	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/checkbox"
	"github.com/chenhunghan/boba/glyph"
	"github.com/chenhunghan/boba/separator"
	"github.com/chenhunghan/boba/statusbar"
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

// The storybook is a consumer, so it owns all styling here (the library
// packages ship none).
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
		{Name: "separator", Stories: []Story{
			{Name: "horizontal", Render: func(w, h int) string {
				return separator.Separator{Length: w, Style: rule}.Render()
			}},
			{Name: "vertical", Render: func(w, h int) string {
				return separator.Separator{Orientation: separator.Vertical, Length: h, Style: rule}.Render()
			}},
			{Name: "custom char", Render: func(w, h int) string {
				return separator.Separator{Length: w, Char: "·", Style: rule}.Render()
			}},
		}},
		{Name: "checkbox", Stories: []Story{
			{Name: "states", Render: func(w, h int) string {
				return strings.Join([]string{
					checkbox.Checkbox{Label: "unchecked", Style: cbStyle}.Render(),
					checkbox.Checkbox{Label: "checked", Checked: true, Style: cbStyle}.Render(),
					checkbox.Checkbox{Label: "focused", Focused: true, Style: cbStyle}.Render(),
				}, "\n")
			}},
		}},
		{Name: "button", Stories: []Story{
			{Name: "stack", Render: func(w, h int) string {
				s := button.Stack{
					Buttons:    []button.Button{{Text: "Start"}, {Text: "Stop"}, {Text: "Restart"}},
					Width:      14,
					ItemHeight: 1,
					Selected:   1,
					Hover:      -1,
					Active:     true,
				}
				for i := range s.Buttons {
					s.Buttons[i].Style = demoBtn
				}
				return s.Render()
			}},
		}},
		{Name: "box", Stories: []Story{
			{Name: "notched", Render: func(w, h int) string {
				return box.Box{
					LeftNotches:  []box.Notch{{Text: "title"}},
					RightNotches: []box.Notch{{Text: "1", Badge: "q"}},
					BorderColor:  accent,
					Body:         "headless bordered region",
				}.Render(min(w, 32), min(h, 5))
			}},
		}},
		{Name: "statusbar", Stories: []Story{
			{Name: "row", Render: func(w, h int) string {
				return statusbar.Bar{
					Left:  []statusbar.Item{{Key: "↑↓", Text: "move"}, {Key: "1", Letter: "h", Text: "hot"}},
					Right: []statusbar.Item{{Key: "esc", Text: "back"}},
				}.Render(min(w, 40))
			}},
		}},
		{Name: "glyph", Stories: []Story{
			{Name: "scripts", Render: func(w, h int) string {
				return "x" + glyph.Superscript(2) + "   H" + glyph.Subscript(2) + "O"
			}},
		}},
	}
}
