package tab_test

import (
	"fmt"

	"github.com/chenhunghan/boba/tab"
)

func ExampleGroup() {
	var g tab.Group[string]
	g, _ = g.AddTab(tab.Tab[string]{ID: "files", Label: "Files", Model: tab.Static("file list")})
	g, _ = g.AddTab(tab.Tab[string]{ID: "logs", Label: "Logs", Model: tab.Static("log output")})

	g.Selected = "files"
	g = g.ApplyKey("]")

	fmt.Println(g.Selected)
	// Output: logs
}
