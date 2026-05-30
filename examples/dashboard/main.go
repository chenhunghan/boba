package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/box"
	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/focus"
	"github.com/chenhunghan/boba/menu"
	"github.com/chenhunghan/boba/navcard"
	"github.com/chenhunghan/boba/overlay"
	"github.com/chenhunghan/boba/panel"
	"github.com/chenhunghan/boba/pins"
	"github.com/chenhunghan/boba/statusbar"
	"github.com/chenhunghan/boba/tab"
)

type focusArea int

const (
	focusNone focusArea = iota
	focusHotbar
	focusNav
	focusMain
)

const (
	hotbarW = 5
	navW    = 30
)

var menuStyle = menu.Style{
	Surface:  lipgloss.Color("236"),
	Inactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#cccccc")).Background(lipgloss.Color("236")),
	Hover:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#3a5a8a")).Bold(true),
	Disabled: dim.Background(lipgloss.Color("236")),
	Border:   lipgloss.NewStyle().Foreground(navAccent).Background(lipgloss.Color("236")),
}

var tabStyle = tab.Style{
	Inactive:    dim,
	Hover:       lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
	Active:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
	Border:      lipgloss.NewStyle().Foreground(borderIdle),
	HoverBar:    lipgloss.NewStyle().Foreground(lipgloss.Color("#717780")),
	SelectedBar: lipgloss.NewStyle().Foreground(mainAccent),
}

type tabKey struct{ action, svc string }

type model struct {
	services []service
	focus    focusArea
	w, h     int

	navSel, navHover, navHoverBtn int
	hotHover                      int

	pins         pins.List
	tabs         tab.Group[tabKey]
	bottomTabs   tab.Group[tabKey]
	topRS        tab.RenderState[tabKey]
	botRS        tab.RenderState[tabKey]
	activeBottom bool

	menu    menu.Group[string]
	menuSvc string
}

func focusCfg() focus.Config[focusArea] {
	return focus.Config[focusArea]{
		KeyJumps:   map[string]focusArea{"1": focusHotbar, "2": focusNav, "3": focusMain},
		CycleOrder: []focusArea{focusHotbar, focusNav, focusMain},
		CycleNext:  []string{"tab"},
		CyclePrev:  []string{"shift+tab"},
		Clear:      []string{"esc"},
		Zero:       focusNone,
		NeedsMouse: func(focusArea) bool { return true },
	}
}

func (m model) bodyH() int  { return max(m.h-1, 3) }
func (m model) split() bool { return len(m.bottomTabs.Tabs) > 0 }

func (m model) layout() panel.Split[focusArea] {
	b := panel.Borders{Top: 1, Right: 1, Bottom: 1, Left: 1}
	return panel.Split[focusArea]{Axis: panel.Horizontal, Children: []panel.Node[focusArea]{
		panel.Panel[focusArea]{ID: focusHotbar, Size: hotbarW, Borders: b},
		panel.Panel[focusArea]{ID: focusNav, Size: navW, Borders: b},
		panel.Panel[focusArea]{ID: focusMain, Size: 0, Borders: b},
	}}
}

func (m model) navStack() navcard.Stack {
	cards := make([]navcard.Card, len(m.services))
	for i, s := range m.services {
		cards[i] = card(s)
	}
	return navcard.Stack{
		Cards: cards, Width: navW - 2, Gap: 1,
		Selected: m.navSel, Hover: m.navHover, HoverButton: m.navHoverBtn,
		Active: m.focus == focusNav,
	}
}

func abbrev(s string) string {
	if len(s) > 3 {
		return s[:3]
	}
	return s
}

func (m model) hotStack() button.Stack {
	ids := m.pins.IDs()
	btns := make([]button.Button, len(ids))
	for i, id := range ids {
		btns[i] = button.Button{Text: abbrev(id), Style: hotBtn}
	}
	return button.Stack{
		Buttons: btns, Width: hotbarW - 2, ItemHeight: 1,
		Selected: m.pins.Selected(), Hover: m.hotHover, Active: m.focus == focusHotbar,
	}
}

func (m model) svc(id string) service {
	for _, s := range m.services {
		if s.id == id {
			return s
		}
	}
	return service{}
}

func (m *model) openTab(action, id string) {
	key := tabKey{action, id}
	t := tab.Tab[tabKey]{ID: key, Label: action + ": " + id, Closable: true, Style: tabStyle}
	if action == "logs" {
		t.Model = tab.Static(logs(m.svc(id)))
		if _, ok := m.bottomTabs.Find(key); ok {
			m.bottomTabs.Selected = key
		} else {
			m.bottomTabs, _ = m.bottomTabs.AddTab(t)
		}
		m.activeBottom = true
		return
	}
	t.Model = tab.Static(properties(m.svc(id)))
	if _, ok := m.tabs.Find(key); ok {
		m.tabs.Selected = key
	} else {
		m.tabs, _ = m.tabs.AddTab(t)
	}
	m.activeBottom = false
}

func (m model) contextMenu(s service) menu.Group[string] {
	pinLabel := "Pin"
	if m.pins.Contains(s.id) {
		pinLabel = "Unpin"
	}
	return menu.Group[string]{
		Items: []menu.Item[string]{
			{ID: "props", Label: "Properties"},
			{ID: "logs", Label: "Logs"},
			{ID: "pin", Label: pinLabel},
		},
		Open: true, Hover: 0, Style: menuStyle,
	}
}

func toggleState(s state) state {
	if s == running {
		return stopped
	}
	return running
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil

	case menu.ChosenMsg[string]:
		m.menu.Open = false
		switch msg.ID {
		case "props":
			m.openTab("props", m.menuSvc)
		case "logs":
			m.openTab("logs", m.menuSvc)
		case "pin":
			if m.pins.Contains(m.menuSvc) {
				m.pins.Unpin(m.menuSvc)
			} else {
				m.pins.Pin(m.menuSvc)
			}
		}
		return m, nil
	case menu.CancelledMsg:
		m.menu.Open = false
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.menu.Open {
			var cmd tea.Cmd
			m.menu, cmd = m.menu.Update(msg)
			return m, cmd
		}
		r := focus.ApplyKey(m.focus, msg.String(), focusCfg())
		if r.Focus != m.focus {
			m.focus = r.Focus
			return m, r.Cmd
		}
		return m.key(msg.String()), nil

	case tea.MouseMsg:
		return m.mouse(msg)
	}
	return m, nil
}

func (m model) key(k string) model {
	switch m.focus {
	case focusNav:
		switch k {
		case "up", "k":
			if m.navSel > 0 {
				m.navSel--
			}
		case "down", "j":
			if m.navSel < len(m.services)-1 {
				m.navSel++
			}
		case "enter":
			m.openTab("props", m.services[m.navSel].id)
		}
	case focusHotbar:
		switch k {
		case "up", "k":
			m.pins.Move(-1)
		case "down", "j":
			m.pins.Move(1)
		}
	case focusMain:
		switch k {
		case "up", "k":
			m.activeBottom = false
		case "down", "j":
			m.activeBottom = true
		default:
			if m.activeBottom {
				m.bottomTabs, _ = m.bottomTabs.Update(keyMsg(k))
			} else {
				m.tabs, _ = m.tabs.Update(keyMsg(k))
			}
		}
	}
	return m
}

func keyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func (m model) mouse(msg tea.MouseMsg) (model, tea.Cmd) {
	if m.menu.Open {
		var cmd tea.Cmd
		m.menu, cmd = m.menu.Update(msg)
		return m, cmd
	}
	hit := panel.HitTest[focusArea](m.layout(), m.w, m.bodyH(), msg.X, msg.Y)
	if !hit.Found {
		return m, nil
	}
	press := msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft
	if press {
		m.focus = focus.ApplyClick(m.focus, hit.Panel, focusCfg()).Focus
	}

	m.navHover, m.navHoverBtn, m.hotHover = -1, -1, -1
	m.topRS, m.botRS = tab.RenderState[tabKey]{}, tab.RenderState[tabKey]{}
	switch hit.Panel {
	case focusNav:
		nh := m.navStack().HitTest(hit.LocalX, hit.LocalY)
		m.navHover, m.navHoverBtn = nh.Card, nh.Button
		if press && nh.Card >= 0 {
			m = m.navClick(nh, msg.X, msg.Y)
		}
	case focusHotbar:
		if idx, area := m.hotStack().HitTest(hit.LocalX, hit.LocalY); area == button.HitBody {
			m.hotHover = idx
			if press {
				if ids := m.pins.IDs(); idx < len(ids) {
					m.pins.SetSelected(idx)
					m.selectByID(ids[idx])
				}
			}
		}
	case focusMain:
		m = m.mainMouse(hit.LocalX, hit.LocalY, press)
	}
	return m, nil
}

// The body stacks the two groups as equal halves with no divider, so the bottom
// group's local y is the hit minus the top height.
func (m model) mainMouse(x, y int, press bool) model {
	topH := (m.bodyH() - 2) / 2
	if m.split() && y >= topH {
		ly := y - topH
		m.botRS = m.bottomTabs.HoverState(x, ly)
		if press {
			m.bottomTabs, _ = m.bottomTabs.ClickAt(x, ly)
			m.activeBottom = true
		}
		return m
	}
	m.topRS = m.tabs.HoverState(x, y)
	if press {
		m.tabs, _ = m.tabs.ClickAt(x, y)
		m.activeBottom = false
	}
	return m
}

func (m *model) selectByID(id string) {
	for i, s := range m.services {
		if s.id == id {
			m.navSel = i
			return
		}
	}
}

func (m model) navClick(nh navcard.Hit, x, y int) model {
	m.navSel = nh.Card
	s := m.services[nh.Card]
	if nh.Button < 0 {
		m.openTab("props", s.id)
		return m
	}
	leftCount := 0
	if s.st == running || s.st == stopped {
		leftCount = 1
	}
	if nh.Button < leftCount {
		m.services[nh.Card].st = toggleState(s.st)
		return m
	}
	m.menu = m.contextMenu(s)
	m.menu.AnchorX, m.menu.AnchorY = x, y
	m.menuSvc = s.id
	return m
}

func borderFor(focused bool) lipgloss.Color {
	if focused {
		return borderActive
	}
	return borderIdle
}

func fitLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines[:n], "\n")
}

func (m model) mainBody(w, innerH int) string {
	if !m.split() {
		return fitLines(m.tabs.Render(w, m.topRS), innerH)
	}
	topH := innerH / 2
	botH := innerH - topH
	top := fitLines(m.tabs.Render(w, m.topRS), topH)
	bot := fitLines(m.bottomTabs.Render(w, m.botRS), botH)
	return top + "\n" + bot
}

func (m model) View() string {
	if m.w == 0 {
		return "loading…"
	}
	bodyH := m.bodyH()

	hot := box.Box{
		LeftNotches: []box.Notch{{Text: "p"}},
		BorderColor: borderFor(m.focus == focusHotbar),
		Body:        m.hotStack().Render(),
	}.Render(hotbarW, bodyH)

	nav := box.Box{
		LeftNotches: []box.Notch{{Text: "services"}},
		BorderColor: borderFor(m.focus == focusNav),
		Body:        m.navStack().Render(),
	}.Render(navW, bodyH)

	mainW := max(m.w-hotbarW-navW, 12)
	main := box.Box{
		LeftNotches: []box.Notch{{Text: "main"}},
		BorderColor: borderFor(m.focus == focusMain),
		Body:        m.mainBody(mainW-2, bodyH-2),
	}.Render(mainW, bodyH)

	row := lipgloss.JoinHorizontal(lipgloss.Top, hot, nav, main)
	if m.menu.Open {
		row = overlay.Overlay(row, m.menu.Render(), m.menu.AnchorX, m.menu.AnchorY)
	}

	bar := statusbar.Bar{
		Left: []statusbar.Item{
			{Key: "1-3", Text: "focus", Style: barStyle},
			{Key: "↑↓", Text: "select", Style: barStyle},
			{Key: "[ ]", Text: "tabs", Style: barStyle},
			{Key: "⋯", Text: "menu", Style: barStyle},
		},
		Right: []statusbar.Item{{Key: "q", Text: "quit", Style: barStyle}},
	}.Render(m.w)

	return lipgloss.JoinVertical(lipgloss.Left, row, bar)
}

func main() {
	m := model{services: seed(), focus: focusNav, navHover: -1, navHoverBtn: -1, hotHover: -1}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
