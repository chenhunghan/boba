package main

import "github.com/chenhunghan/boba/separator"

func separatorStories() Component {
	return Component{Name: "separator", Stories: []Story{
		{Name: "horizontal", Render: func(w, h int) string {
			return separator.Separator{Length: w, Style: rule}.Render()
		}},
		{Name: "vertical", Render: func(w, h int) string {
			return separator.Separator{Orientation: separator.Vertical, Length: h, Style: rule}.Render()
		}},
		{Name: "custom char", Render: func(w, h int) string {
			return separator.Separator{Length: w, Char: "·", Style: rule}.Render()
		}},
	}}
}
