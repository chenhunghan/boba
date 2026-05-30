package main

import (
	"github.com/chenhunghan/boba/accordion"
)

func accordionStories() Component {
	demoStyle := accordion.Style{
		Header:       muted,
		HeaderCursor: chosen,
		Body:         bright,
	}

	sections := []accordion.Section{
		{Title: "General", Body: "  display options"},
		{Title: "Network", Body: "  bandwidth limits"},
		{Title: "Storage", Body: "  disk thresholds"},
	}

	return Component{Name: "accordion", Stories: []Story{
		{Name: "mixed", Render: func(w, h int) string {
			a := accordion.Accordion{
				Sections: sections,
				Expanded: []bool{true, false, false},
				Cursor:   0,
				Style:    demoStyle,
			}
			return a.Render()
		}},
		{Name: "all collapsed", Render: func(w, h int) string {
			a := accordion.Accordion{
				Sections: sections,
				Cursor:   1,
				Style:    demoStyle,
			}
			return a.Render()
		}},
		{Name: "all expanded", Render: func(w, h int) string {
			a := accordion.Accordion{
				Sections: sections,
				Expanded: []bool{true, true, true},
				Cursor:   2,
				Style:    demoStyle,
			}
			return a.Render()
		}},
	}}
}
