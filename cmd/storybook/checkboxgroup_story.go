package main

import (
	"github.com/chenhunghan/boba/checkboxgroup"
)

func checkboxgroupStories() Component {
	style := checkboxgroup.Style{
		Checked:   "[x]",
		Unchecked: "[ ]",
		Normal:    muted,
		Cursor:    chosen,
	}

	return Component{Name: "checkboxgroup", Stories: []Story{
		{Name: "mixed", Render: func(w, h int) string {
			return checkboxgroup.Group{
				Options: []string{"Apples", "Bananas", "Cherries"},
				Checked: []bool{true, false, true},
				Style:   style,
			}.Render()
		}},
		{Name: "focused cursor", Render: func(w, h int) string {
			return checkboxgroup.Group{
				Options: []string{"CPU", "Memory", "Network", "Disk"},
				Checked: []bool{true, true, false, false},
				Cursor:  2,
				Focused: true,
				Style:   style,
			}.Render()
		}},
		{Name: "custom glyphs", Render: func(w, h int) string {
			alt := checkboxgroup.Style{
				Checked:   "(*)",
				Unchecked: "( )",
				Normal:    muted,
				Cursor:    chosen,
			}
			return checkboxgroup.Group{
				Options: []string{"Low", "Medium", "High"},
				Checked: []bool{false, true, false},
				Cursor:  1,
				Focused: true,
				Style:   alt,
			}.Render()
		}},
	}}
}
