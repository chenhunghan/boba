package main

import "github.com/chenhunghan/boba/statusbar"

func statusbarStories() Component {
	return Component{Name: "statusbar", Stories: []Story{
		{Name: "row", Render: func(w, h int) string {
			return statusbar.Bar{
				Left:  []statusbar.Item{{Key: "↑↓", Text: "move"}, {Key: "1", Letter: "h", Text: "hot"}},
				Right: []statusbar.Item{{Key: "esc", Text: "back"}},
			}.Render(min(w, 40))
		}},
	}}
}
