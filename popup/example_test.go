package popup_test

import (
	"fmt"
	"strings"

	"github.com/chenhunghan/boba/popup"
)

func ExamplePlace() {
	// A 4x1 anchor at (2,2) on a 40x20 screen; a 4x3 box placed below it
	// sits just under the anchor's left edge.
	x, y := popup.Place(2, 2, 4, 1, 4, 3, 40, 20, popup.Below)
	fmt.Printf("%d,%d\n", x, y)
	// Output: 2,3
}

func ExampleCenter() {
	x, y := popup.Center(10, 4, 40, 20)
	fmt.Printf("%d,%d\n", x, y)
	// Output: 15,8
}

func ExampleIsolate() {
	// Each line gains a leading SGR reset; line count is unchanged.
	out := popup.Isolate("red\ngreen")
	fmt.Println(strings.Count(out, "\n") + 1)
	fmt.Println(strings.HasPrefix(out, "\x1b[0m"))
	// Output:
	// 2
	// true
}
