//go:build prod

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sqweek/dialog"

	"fiscal-reader/internal/convert"
)

func getFilename() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	startDir := filepath.Dir(exePath)

	filename, err := dialog.File().
		Title("Selecionar arquivo ZIP com documentos fiscais").
		Filter("Arquivo ZIP", "zip").
		SetStartDir(startDir).
		Load()
	if err != nil {
		// user cancelled or error — exit silently
		os.Exit(0)
	}
	return filename
}

func showResult(res convert.Result) {
	var parts []string
	for _, c := range res.Counts {
		parts = append(parts, fmt.Sprintf("%d linha(s) de %s", c.Rows, c.Sheet))
	}
	if len(parts) == 0 {
		dialog.Message("Nenhum documento fiscal encontrado no arquivo ZIP.").
			Title("Fiscal Reader — Concluído").Info()
		return
	}

	msg := fmt.Sprintf("%s exportadas para:\n%s", strings.Join(parts, "\n"), strings.Join(res.Outputs, "\n"))
	if res.Events > 0 {
		msg += fmt.Sprintf("\n\n%d evento(s) lidos, %d cancelamento(s) aplicados.", res.Events, res.Cancelled)
	}
	if res.Skipped > 0 {
		msg += fmt.Sprintf("\n\n%d arquivo(s) ignorados por erro.", res.Skipped)
	}
	if res.Ignored > 0 {
		msg += fmt.Sprintf("\n%d arquivo(s) de tipo não suportado.", res.Ignored)
	}
	dialog.Message("%s", msg).Title("Fiscal Reader — Concluído").Info()
}
