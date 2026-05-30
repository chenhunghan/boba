package scrollarea_test

import (
	"fmt"

	"github.com/chenhunghan/boba/scrollarea"
)

func ExampleScrollArea() {
	content := "line0\nline1\nline2\nline3\nline4\nline5"
	a := scrollarea.ScrollArea{Content: content, Height: 3, Width: 7}
	a.Scroll.Offset = 1
	fmt.Println(a.Render())
	// Output:
	// line1 █
	// line2 █
	// line3 │
}
