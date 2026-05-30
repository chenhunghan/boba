package selectbox_test

import (
	"fmt"

	"github.com/chenhunghan/boba/selectbox"
)

// ExampleSelectBox shows the closed control: the selected value with the
// chevron pinned to the right within W. Styles are left zero (no color) so the
// output is plain text.
func ExampleSelectBox() {
	s := selectbox.SelectBox{
		Options:  []string{"Low", "Medium", "High"},
		Selected: 1,
		W:        10,
	}
	fmt.Printf("%q\n", s.View())
	// Output: "Medium   ▾"
}

// ExampleSelectBox_applyKey drives the pure core: open the list, move the
// highlight down, and confirm — selecting the option under the cursor.
func ExampleSelectBox_applyKey() {
	s := selectbox.SelectBox{Options: []string{"Low", "Medium", "High"}, W: 10}

	s, _ = s.ApplyKey("enter") // open, highlight current Selected (0)
	s, _ = s.ApplyKey("down")  // highlight 1
	s, confirmed := s.ApplyKey("enter")

	fmt.Println(confirmed, s.Selected, s.Options[s.Selected])
	// Output: true 1 Medium
}
