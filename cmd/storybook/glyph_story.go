package main

import "github.com/chenhunghan/boba/glyph"

func glyphStories() Component {
	return Component{Name: "glyph", Stories: []Story{
		{Name: "scripts", Render: func(w, h int) string {
			return "x" + glyph.Superscript(2) + "   H" + glyph.Subscript(2) + "O"
		}},
	}}
}
