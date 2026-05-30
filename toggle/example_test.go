package toggle_test

import (
	"fmt"

	"github.com/chenhunghan/boba/toggle"
)

func ExampleToggle() {
	tg := toggle.Toggle{Label: "Auto-scroll", Pressed: true}
	fmt.Println(tg.Render())
	fmt.Println(tg.Width())
	// Output:
	// Auto-scroll
	// 11
}
