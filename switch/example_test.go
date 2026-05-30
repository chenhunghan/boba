package swtch_test

import (
	"fmt"

	swtch "github.com/chenhunghan/boba/switch"
)

func ExampleSwitch() {
	s := swtch.Switch{On: true, Label: "Wi-Fi"}
	fmt.Println(s.Render())
	// Output: [on] Wi-Fi
}
