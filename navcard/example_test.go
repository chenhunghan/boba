package navcard_test

import (
	"fmt"

	"github.com/chenhunghan/boba/navcard"
)

func ExampleStack() {
	s := navcard.Stack{
		Cards: []navcard.Card{
			{Title: "First"},
			{Title: "Second"},
		},
		Width:       20,
		Hover:       -1,
		HoverButton: -1,
	}

	hit := s.HitTest(2, 0)
	fmt.Println(hit.Card)
	// Output: 0
}
