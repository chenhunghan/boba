package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/box"
	"github.com/chenhunghan/boba/button"
)

var btnStyle = button.Style{
	Inactive: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
	Hover:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	Active:   lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true),
}

type model struct {
	menu   button.Stack
	chosen string
}

func initialModel() model {
	return model{
		menu: button.Stack{
			Buttons: []button.Button{
				{Text: "Overview", Style: btnStyle},
				{Text: "Metrics", Style: btnStyle},
				{Text: "Logs", Style: btnStyle},
			},
			Width:      14,
			ItemHeight: 1,
			Hover:      -1,
			Active:     true,
		},
		chosen: "Overview",
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.menu, cmd = m.menu.Update(msg)
		return m, cmd
	case button.ActivatedMsg:
		m.chosen = m.menu.Buttons[msg.Index].Text
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	sidebar := box.Box{
		LeftNotches: []box.Notch{{Text: "menu"}},
		Body:        m.menu.View(),
	}.Render(18, 8)
	main := box.Box{
		LeftNotches: []box.Notch{{Text: m.chosen}},
		Body:        fmt.Sprintf("Selected: %s\n\n↑/↓ move · enter select · q quit", m.chosen),
	}.Render(42, 8)
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
