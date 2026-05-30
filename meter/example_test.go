package meter_test

import (
	"fmt"

	"github.com/chenhunghan/boba/meter"
)

func ExampleMeter() {
	m := meter.Meter{
		Value: 6,
		Max:   10,
		Width: 10,
		Style: meter.Style{FillChar: "#", EmptyChar: "-"},
	}
	fmt.Println(m.Render())
	// Output: ######----
}
