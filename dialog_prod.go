//go:build prod

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sqweek/dialog"
)

func getFilename() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	startDir := filepath.Dir(exePath)

	filename, err := dialog.File().
		Title("Selecionar arquivo ZIP com CT-es").
		Filter("Arquivo ZIP", "zip").
		SetStartDir(startDir).
		Load()
	if err != nil {
		// user cancelled or error — exit silently
		os.Exit(0)
	}
	return filename
}

func showResult(written, skipped int, outPaths []string) {
	msg := fmt.Sprintf("%d CT-e(s) exportados para:\n%s", written, strings.Join(outPaths, "\n"))
	if skipped > 0 {
		msg += fmt.Sprintf("\n\n%d arquivo(s) ignorados por erro.", skipped)
	}
	dialog.Message("%s", msg).Title("CTE Reader — Concluído").Info()
}
