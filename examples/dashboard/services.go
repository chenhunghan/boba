package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/navcard"
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

var (
	accent   = lipgloss.Color("63")
	dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
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
		Inactive: dim,
		Hover:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
		Active:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
	}
	hotBtn = button.Style{
		Inactive: lipgloss.NewStyle().Background(lipgloss.Color("237")),
		Hover:    lipgloss.NewStyle().Background(lipgloss.Color("#3a5a8a")),
		Active:   lipgloss.NewStyle().Background(lipgloss.Color("#5294e2")),
	}
)

func cardStyle() navcard.Style {
	mk := func(barColor lipgloss.Color, titleColor lipgloss.Color) navcard.StateStyle {
		return navcard.StateStyle{
			Bar:         lipgloss.NewStyle().Foreground(barColor),
			BarChar:     "▌",
			Fill:        lipgloss.NewStyle(),
			Title:       lipgloss.NewStyle().Foreground(titleColor).Bold(true),
			Subtitle:    dim,
			Description: dim,
		}
	}
	return navcard.Style{
		Inactive: mk("#4b5563", "#bbbbbb"),
		Hover:    mk("#2a6da3", "#ffffff"),
		Active:   mk(accent, "#ffffff"),
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
	key := dim.Render
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render
	rows := [][2]string{
		{"Status", stateLabel(s.st)},
		{"Image", s.image},
		{"Uptime", s.uptime},
	}
	out := ""
	for _, r := range rows {
		out += fmt.Sprintf("%s  %s\n", key(r[0]+":"), val(r[1]))
	}
	return out
}

func logs(s service) string {
	return dim.Render(fmt.Sprintf("[%s] listening\n[%s] ready\n[%s] ok", s.name, s.name, s.name))
}
