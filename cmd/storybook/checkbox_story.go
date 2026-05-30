package main

import (
	"strings"

	"github.com/chenhunghan/boba/checkbox"
)

func checkboxStories() Component {
	return Component{Name: "checkbox", Stories: []Story{
		{Name: "states", Render: func(w, h int) string {
			return strings.Join([]string{
				checkbox.Checkbox{Label: "unchecked", Style: cbStyle}.Render(),
				checkbox.Checkbox{Label: "checked", Checked: true, Style: cbStyle}.Render(),
				checkbox.Checkbox{Label: "focused", Focused: true, Style: cbStyle}.Render(),
			}, "\n")
		}},
	}}
}
