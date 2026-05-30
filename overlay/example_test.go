package overlay_test

import (
	"fmt"
	"strings"

	"github.com/chenhunghan/boba/overlay"
)

func ExampleOverlay() {
	bg := "........\n........\n........"
	fg := "##\n##"

	out := overlay.Overlay(bg, fg, 3, 1)

	fmt.Println(strings.Count(out, "\n") + 1) // lines preserved
	// Output: 3
}
