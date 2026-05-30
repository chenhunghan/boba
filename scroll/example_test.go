package scroll_test

import (
	"fmt"

	"github.com/chenhunghan/boba/scroll"
)

func ExampleScroll() {
	content := "line0\nline1\nline2\nline3\nline4"
	s := scroll.Scroll{Height: 2, Offset: 2}
	fmt.Println(s.Render(content))
	// Output:
	// line2
	// line3
}
