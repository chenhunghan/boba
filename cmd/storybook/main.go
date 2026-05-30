package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/box"
	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/statusbar"
)

const sidebarW = 22

type model struct {
	catalog []Component
	sel     int
	story   int
	w, h    int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.sel > 0 {
				m.sel--
				m.story = 0
			}
		case "down", "j":
			if m.sel < len(m.catalog)-1 {
				m.sel++
				m.story = 0
			}
		case "left", "[":
			if m.story > 0 {
				m.story--
			}
		case "right", "]":
			if m.story < len(m.catalog[m.sel].Stories)-1 {
				m.story++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.w == 0 {
		return "loading…"
	}
	bodyH := max(m.h-1, 3)

	btns := make([]button.Button, len(m.catalog))
	for i, c := range m.catalog {
		btns[i] = button.Button{Text: c.Name, Style: sidebarBtn}
	}
	sidebar := box.Box{
		LeftNotches: []box.Notch{{Text: "boba"}},
		BorderColor: accent,
		Body: button.Stack{
			Buttons: btns, Width: sidebarW - 2, ItemHeight: 1,
			Selected: m.sel, Hover: -1, Active: true,
		}.Render(),
	}.Render(sidebarW, bodyH)

	comp := m.catalog[m.sel]
	st := comp.Stories[m.story]
	previewW := max(m.w-sidebarW, 12)
	preview := box.Box{
		LeftNotches:  []box.Notch{{Text: comp.Name}},
		RightNotches: []box.Notch{{Text: fmt.Sprintf("%d/%d %s", m.story+1, len(comp.Stories), st.Name)}},
		BorderColor:  accent,
		Body:         st.Render(previewW-2, bodyH-2),
	}.Render(previewW, bodyH)

	bar := statusbar.Bar{
		Left:  []statusbar.Item{{Key: "↑↓", Text: "component"}, {Key: "←→", Text: "variant"}},
		Right: []statusbar.Item{{Key: "q", Text: "quit"}},
	}.Render(m.w)

	return lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, sidebar, preview), bar)
}

func main() {
	if _, err := tea.NewProgram(model{catalog: catalog()}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
