package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/navcard"
	"github.com/chenhunghan/boba/statusbar"
)

type state int

const (
	running state = iota
	starting
	stopped
)

type service struct {
	id, name, image string
	st              state
	uptime          string
}

func seed() []service {
	return []service{
		{"web", "web", "nginx:1.25", running, "3d 4h"},
		{"api", "api", "go-api:2.1", running, "3d 4h"},
		{"db", "db", "postgres:16", running, "12d 2h"},
		{"cache", "cache", "redis:7", stopped, "—"},
		{"worker", "worker", "worker:1.0", starting, "—"},
		{"proxy", "proxy", "traefik:3", running, "8h"},
	}
}

func stateColor(s state) lipgloss.Color {
	switch s {
	case running:
		return "#73c990"
	case starting:
		return "#e5c07b"
	default:
		return "#888888"
	}
}

func stateLabel(s state) string {
	switch s {
	case running:
		return "running"
	case starting:
		return "starting…"
	default:
		return "stopped"
	}
}

// Outer panel borders use a single neutral scheme so the three columns read
// as one set; the per-region accents below are reserved for sub-component
// chrome (tab borders, the menu border).
var (
	borderIdle   = lipgloss.Color("#4b5563")
	borderActive = lipgloss.Color("#8c939e")

	navAccent  = lipgloss.Color("#bd93f9") // navigator region
	mainAccent = lipgloss.Color("#73c990") // main region

	dim = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	barStyle = statusbar.ItemStyle{
		Key:  lipgloss.NewStyle().Foreground(lipgloss.Color("#bbbbbb")).Bold(true),
		Text: dim,
	}

	startBtn = button.Style{
		Inactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#9ed99e")).Background(lipgloss.Color("#1e3a1e")),
		Hover:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#3d8b3d")).Bold(true),
		Active:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#4caf50")),
	}
	stopBtn = button.Style{
		Inactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#d9b58e")).Background(lipgloss.Color("#3a2e1e")),
		Hover:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#cc7a1f")).Bold(true),
		Active:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#e88e22")),
	}
	dotsBtn = button.Style{
		Inactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#dddddd")).Background(lipgloss.Color("#3a3a4a")),
		Hover:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#5a5a6a")).Bold(true),
		Active:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#7a6dab")),
	}
	hotBtn = button.Style{
		Inactive: lipgloss.NewStyle().Background(lipgloss.Color("237")),
		Hover:    lipgloss.NewStyle().Background(lipgloss.Color("#3a5a8a")),
		Active:   lipgloss.NewStyle().Background(lipgloss.Color("#5294e2")),
	}
)

func cardStyle() navcard.Style {
	mk := func(barColor, titleColor, subColor lipgloss.Color) navcard.StateStyle {
		return navcard.StateStyle{
			Bar:         lipgloss.NewStyle().Foreground(barColor),
			BarChar:     "▌",
			Fill:        lipgloss.NewStyle(),
			Title:       lipgloss.NewStyle().Foreground(titleColor).Bold(true),
			Subtitle:    lipgloss.NewStyle().Foreground(subColor),
			Description: lipgloss.NewStyle().Foreground(subColor),
		}
	}
	return navcard.Style{
		Inactive: mk("#4b5563", "#bbbbbb", "#888888"),
		Hover:    mk("#2a6da3", "#ffffff", "#aaaaaa"),
		Active:   mk("#3d90ce", "#ffffff", "#aaaaaa"),
	}
}

var navStyle = cardStyle()

func card(s service) navcard.Card {
	glyph := lipgloss.NewStyle().Foreground(stateColor(s.st)).Render("●")
	c := navcard.Card{
		Title:    glyph + " " + s.name,
		Subtitle: stateLabel(s.st),
		Style:    navStyle,
	}
	switch s.st {
	case running:
		c.Buttons = []button.Button{{Text: "stop", Style: stopBtn}}
	case stopped:
		c.Buttons = []button.Button{{Text: "start", Style: startBtn}}
	}
	c.RightButtons = []button.Button{{Text: "⋯", Style: dotsBtn}}
	return c
}

func properties(s service) string {
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render
	rows := [][2]string{
		{"Status", stateLabel(s.st)},
		{"Image", s.image},
		{"Uptime", s.uptime},
	}
	out := ""
	for _, r := range rows {
		out += fmt.Sprintf("%s  %s\n", dim.Render(r[0]+":"), val(r[1]))
	}
	return out
}

func logs(s service) string {
	return dim.Render(fmt.Sprintf("[%s] listening\n[%s] ready\n[%s] ok", s.name, s.name, s.name))
}
