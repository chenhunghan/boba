package main

import "github.com/chenhunghan/boba/button"

func buttonStories() Component {
	return Component{Name: "button", Stories: []Story{
		{Name: "stack", Render: func(w, h int) string {
			s := button.Stack{
				Buttons:    []button.Button{{Text: "Start"}, {Text: "Stop"}, {Text: "Restart"}},
				Width:      14,
				ItemHeight: 1,
				Selected:   1,
				Hover:      -1,
				Active:     true,
			}
			for i := range s.Buttons {
				s.Buttons[i].Style = demoBtn
			}
			return s.Render()
		}},
	}}
}
