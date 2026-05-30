package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/box"
	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/statusbar"
)

const sidebarW = 22

//go:embed *_story.go
var storyFS embed.FS

func storySource(name string) string {
	b, err := storyFS.ReadFile(name + "_story.go")
	if err != nil {
		return "(source unavailable)"
	}
	s := strings.ReplaceAll(string(b), "\t", "  ")
	if i := strings.Index(s, "func "); i >= 0 {
		return s[i:]
	}
	return s
}

type model struct {
	catalog  []Component
	sel      int
	story    int
	hover    int
	showCode bool
	w, h     int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) sidebarStack() button.Stack {
	btns := make([]button.Button, len(m.catalog))
	for i, c := range m.catalog {
		btns[i] = button.Button{Text: c.Name, Style: sidebarBtn}
	}
	return button.Stack{
		Buttons: btns, Width: sidebarW - 2, ItemHeight: 1,
		Selected: m.sel, Hover: m.hover, Active: true,
	}
}

func dec(x int) int {
	if x > 0 {
		return x - 1
	}
	return 0
}

func inc(x, n int) int {
	if x < n-1 {
		return x + 1
	}
	return x
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			m.showCode = !m.showCode
		case "up", "k":
			m.sel, m.story = dec(m.sel), 0
		case "down", "j":
			m.sel, m.story = inc(m.sel, len(m.catalog)), 0
		case "left", "[":
			m.story = dec(m.story)
		case "right", "]":
			m.story = inc(m.story, len(m.catalog[m.sel].Stories))
		}
	case tea.MouseMsg:
		m = m.mouse(msg)
	}
	return m, nil
}

func (m model) mouse(msg tea.MouseMsg) model {
	overSidebar := msg.X < sidebarW
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if overSidebar {
			m.sel, m.story = dec(m.sel), 0
		} else {
			m.story = dec(m.story)
		}
		return m
	case tea.MouseButtonWheelDown:
		if overSidebar {
			m.sel, m.story = inc(m.sel, len(m.catalog)), 0
		} else {
			m.story = inc(m.story, len(m.catalog[m.sel].Stories))
		}
		return m
	}

	if msg.X >= 1 && msg.X < sidebarW-1 && msg.Y >= 1 {
		if idx, area := m.sidebarStack().HitTest(msg.X-1, msg.Y-1); area == button.HitBody {
			m.hover = idx
			if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
				m.sel, m.story = idx, 0
			}
		}
		return m
	}

	m.hover = -1
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft && msg.X >= sidebarW {
		if n := len(m.catalog[m.sel].Stories); n > 0 {
			m.story = (m.story + 1) % n
		}
	}
	return m
}

func (m model) View() string {
	if m.w == 0 {
		return "loading…"
	}
	bodyH := max(m.h-1, 3)

	sidebar := box.Box{
		LeftNotches: []box.Notch{{Text: "boba"}},
		BorderColor: accent,
		Body:        m.sidebarStack().Render(),
	}.Render(sidebarW, bodyH)

	comp := m.catalog[m.sel]
	st := comp.Stories[m.story]
	previewW := max(m.w-sidebarW, 12)

	var right string
	if m.showCode {
		ph := max(bodyH/2, 3)
		ch := max(bodyH-ph, 3)
		preview := box.Box{
			LeftNotches:  []box.Notch{{Text: comp.Name}},
			RightNotches: []box.Notch{{Text: fmt.Sprintf("%d/%d %s", m.story+1, len(comp.Stories), st.Name)}},
			BorderColor:  accent,
			Body:         st.Render(previewW-2, ph-2),
		}.Render(previewW, ph)
		code := box.Box{
			LeftNotches: []box.Notch{{Text: "code"}},
			BorderColor: accent,
			Body:        storySource(comp.Name),
		}.Render(previewW, ch)
		right = lipgloss.JoinVertical(lipgloss.Left, preview, code)
	} else {
		right = box.Box{
			LeftNotches:  []box.Notch{{Text: comp.Name}},
			RightNotches: []box.Notch{{Text: fmt.Sprintf("%d/%d %s", m.story+1, len(comp.Stories), st.Name)}},
			BorderColor:  accent,
			Body:         st.Render(previewW-2, bodyH-2),
		}.Render(previewW, bodyH)
	}

	bar := statusbar.Bar{
		Left:  []statusbar.Item{{Key: "↑↓/click", Text: "component"}, {Key: "←→/scroll", Text: "variant"}, {Key: "c", Text: "code"}},
		Right: []statusbar.Item{{Key: "q", Text: "quit"}},
	}.Render(m.w)

	return lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, sidebar, right), bar)
}

func main() {
	p := tea.NewProgram(model{catalog: catalog(), hover: -1}, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
