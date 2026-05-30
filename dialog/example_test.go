package dialog_test

import (
	"fmt"

	"github.com/chenhunghan/boba/dialog"
)

func ExampleDialog() {
	d := dialog.Dialog{
		Title:   "Quit?",
		Body:    "Discard changes?",
		Buttons: []string{"OK", "Cancel"},
		Open:    true,
	}

	d, _, _ = d.ApplyKey("right")
	_, chosen, _ := d.ApplyKey("enter")

	fmt.Println(d.Selected, chosen)
	// Output: 1 true
}
