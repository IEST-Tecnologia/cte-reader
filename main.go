// Command fiscal-reader converts a ZIP of fiscal document XMLs into spreadsheets.
package main

import (
	"fmt"
	"os"
	"strconv"

	"fiscal-reader/internal/convert"
)

// chunkSizeStr is overridden at build time via -ldflags "-X main.chunkSizeStr=N".
var chunkSizeStr = "500000"

func main() {
	zipPath := getFilename()
	if zipPath == "" {
		os.Exit(1)
	}

	res, err := convert.Run(zipPath, chunkSize(), os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	showResult(res)
}

func chunkSize() int {
	n, err := strconv.Atoi(chunkSizeStr)
	if err != nil || n <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid chunk size %q, using 500000\n", chunkSizeStr)
		return 500_000
	}
	return n
}
