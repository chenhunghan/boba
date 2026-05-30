package main

import (
	"github.com/chenhunghan/boba/radiogroup"
)

func radiogroupStories() Component {
	rgStyle := radiogroup.Style{
		Selected: chosen,
		Normal:   muted,
		Focused:  bright,
	}
	return Component{Name: "radiogroup", Stories: []Story{
		{Name: "default", Render: func(w, h int) string {
			return radiogroup.RadioGroup{
				Options:  []string{"Low", "Medium", "High"},
				Selected: 1,
				Style:    rgStyle,
			}.Render()
		}},
		{Name: "focused", Render: func(w, h int) string {
			return radiogroup.RadioGroup{
				Options:  []string{"CPU", "Memory", "Network", "Disk"},
				Selected: 0,
				Focused:  true,
				Style:    rgStyle,
			}.Render()
		}},
		{Name: "custom glyphs", Render: func(w, h int) string {
			st := rgStyle
			st.SelectedGlyph = "◉"
			st.UnselectedGlyph = "○"
			return radiogroup.RadioGroup{
				Options:  []string{"Auto", "Manual"},
				Selected: 1,
				Style:    st,
			}.Render()
		}},
	}}
}
