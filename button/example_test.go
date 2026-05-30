package button_test

import (
	"fmt"

	"github.com/chenhunghan/boba/button"
)

// A Stack is a vertical list of buttons. The same struct backs both
// Render and HitTest, so the two can't drift out of sync. Width and
// ItemHeight must be set for the stack to render and hit-test.
func ExampleStack() {
	s := button.Stack{
		Buttons:    []button.Button{{Text: "Start"}, {Text: "Stop"}},
		Width:      10,
		ItemHeight: 1,
		Hover:      -1, // -1 = nothing under the cursor
	}

	// HitTest maps panel-local (x, y) back to a button index; the second
	// row belongs to the second button.
	idx, _ := s.HitTest(2, 1)
	fmt.Println(idx)
	// Output: 1
}
