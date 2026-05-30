package main

import "github.com/chenhunghan/boba/box"

func boxStories() Component {
	return Component{Name: "box", Stories: []Story{
		{Name: "notched", Render: func(w, h int) string {
			return box.Box{
				LeftNotches:  []box.Notch{{Text: "title"}},
				RightNotches: []box.Notch{{Text: "1", Badge: "q"}},
				BorderColor:  accent,
				Body:         "headless bordered region",
			}.Render(min(w, 32), min(h, 5))
		}},
	}}
}
