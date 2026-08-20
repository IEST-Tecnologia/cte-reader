//go:build !prod

package main

import (
	"fmt"
	"strings"

	"fiscal-reader/internal/convert"
)

func getFilename() string {
	return "data/test.zip"
}

func showResult(res convert.Result) {
	var parts []string
	for _, c := range res.Counts {
		parts = append(parts, fmt.Sprintf("%d %s row(s)", c.Rows, c.Sheet))
	}
	if len(parts) == 0 {
		parts = append(parts, "no records")
	}

	fmt.Printf("Done. %s written to %s", strings.Join(parts, ", "), strings.Join(res.Outputs, ", "))
	if res.Cancelled > 0 {
		fmt.Printf(", %d cancellation(s) applied", res.Cancelled)
	}
	if res.Skipped > 0 {
		fmt.Printf(", %d file(s) skipped", res.Skipped)
	}
	if res.Ignored > 0 {
		fmt.Printf(", %d file(s) of unsupported type", res.Ignored)
	}
	fmt.Println()
}
