package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/scroll"
)

func scrollStories() Component {
	content := func(n int) string {
		ls := make([]string, n)
		for i := range ls {
			ls[i] = fmt.Sprintf("line %02d", i)
		}
		return strings.Join(ls, "\n")
	}
	const rows = 5

	return Component{Name: "scroll", Stories: []Story{
		{Name: "top", Render: func(w, h int) string {
			s := scroll.Scroll{Height: rows, Offset: 0, Style: muted}
			return s.Render(content(20))
		}},
		{Name: "scrolled", Render: func(w, h int) string {
			s := scroll.Scroll{Height: rows, Offset: 7, Style: bright}
			return s.Render(content(20))
		}},
		{Name: "focused", Render: func(w, h int) string {
			s := scroll.Scroll{
				Height:  rows,
				Offset:  3,
				Focused: true,
				Style:   rule.Border(lipgloss.NormalBorder(), false, false, false, true).PaddingLeft(1),
			}
			return s.Render(content(20))
		}},
	}}
}
