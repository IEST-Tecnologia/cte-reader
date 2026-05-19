//go:build !prod

package main

import (
	"fmt"
	"strings"
)

func getFilename() string {
	return "data/test.zip"
}

func showResult(written, skipped int, outPaths []string) {
	fmt.Printf("Done. %d record(s) written to %s", written, strings.Join(outPaths, ", "))
	if skipped > 0 {
		fmt.Printf(", %d file(s) skipped", skipped)
	}
	fmt.Println()
}
