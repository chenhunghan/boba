package main

import (
	"strings"

	swtch "github.com/chenhunghan/boba/switch"
)

func switchStories() Component {
	style := swtch.Style{
		OnGlyph:  "[x]",
		OffGlyph: "[ ]",
		Normal:   muted,
		Focused:  chosen,
	}
	return Component{Name: "switch", Stories: []Story{
		{Name: "states", Render: func(w, h int) string {
			return strings.Join([]string{
				swtch.Switch{Label: "off", Style: style}.Render(),
				swtch.Switch{Label: "on", On: true, Style: style}.Render(),
				swtch.Switch{Label: "focused", On: true, Focused: true, Style: style}.Render(),
			}, "\n")
		}},
		{Name: "glyph fallback", Render: func(w, h int) string {
			bare := swtch.Style{Normal: muted, Focused: chosen}
			return strings.Join([]string{
				swtch.Switch{Label: "Wi-Fi", Style: bare}.Render(),
				swtch.Switch{Label: "Wi-Fi", On: true, Style: bare}.Render(),
			}, "\n")
		}},
	}}
}
