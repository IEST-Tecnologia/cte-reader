//go:build !prod

package main

import "fmt"

func getFilename() string {
	return "data/test.zip"
}

func showResult(written, skipped int, outPath string) {
	fmt.Printf("Done. %d record(s) written to %s", written, outPath)
	if skipped > 0 {
		fmt.Printf(", %d file(s) skipped", skipped)
	}
	fmt.Println()
}
