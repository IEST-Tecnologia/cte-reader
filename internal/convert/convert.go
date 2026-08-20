// Package convert walks a ZIP of fiscal documents and writes one spreadsheet
// per document type found inside it.
package convert

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fiscal-reader/internal/events"
	"fiscal-reader/internal/spec"
	"fiscal-reader/internal/xlsx"
	"fiscal-reader/internal/xmltree"
)

// Result is what a run produced, for reporting back to the user.
type Result struct {
	Counts    []TypeCount
	Events    int // event documents read, cancellations or not
	Cancelled int // access keys an accepted cancellation event applies to
	Skipped   int // unreadable or unparseable files
	Ignored   int // parsed, but no spec claims the root element
	Outputs   []string
}

// TypeCount is the number of rows written for one document type.
type TypeCount struct {
	Sheet string
	Rows  int
}

// Run converts every document in zipPath, writing spreadsheets alongside it.
// Output files hold at most chunkSize rows each. Diagnostics for individual
// files go to logw; only a failure of the whole run returns an error.
func Run(zipPath string, chunkSize int, logw io.Writer) (Result, error) {
	base := strings.TrimSuffix(zipPath, filepath.Ext(zipPath))

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return Result{}, fmt.Errorf("error opening zip: %w", err)
	}
	defer r.Close()

	// Events are read first: a document's status depends on a cancellation
	// that may sit anywhere in the archive, including after the document.
	scan := events.ScanZip(&r.Reader, logw)
	env := spec.Env{Cancelled: scan.Cancelled}

	res := Result{Events: scan.Total, Cancelled: len(scan.Cancelled)}
	// One writer per document type, created on first sighting so a ZIP holding
	// only CT-e never produces an empty NF-e workbook.
	writers := make(map[string]*xlsx.Writer)
	var order []*xlsx.Writer

	for _, file := range r.File {
		if !strings.HasSuffix(strings.ToLower(file.Name), ".xml") {
			continue
		}
		if scan.Files[file.Name] {
			continue // already accounted for by the event pass
		}

		data, err := readZipEntry(file)
		if err != nil {
			fmt.Fprintf(logw, "SKIP %s: %v\n", file.Name, err)
			res.Skipped++
			continue
		}

		root, err := xmltree.Parse(data)
		if err != nil {
			fmt.Fprintf(logw, "SKIP %s: %v\n", file.Name, err)
			res.Skipped++
			continue
		}

		doc, node := spec.ForRoot(root)
		if doc == nil {
			fmt.Fprintf(logw, "IGNORE %s: unsupported root element <%s>\n", file.Name, root.Name)
			res.Ignored++
			continue
		}

		w := writers[doc.Name]
		if w == nil {
			w = xlsx.NewWriter(doc, base, chunkSize)
			writers[doc.Name] = w
			order = append(order, w)
		}
		if err := w.Add(doc.Rows(node, env)); err != nil {
			return res, err
		}
	}

	for _, w := range order {
		if err := w.Close(); err != nil {
			return res, err
		}
		res.Counts = append(res.Counts, TypeCount{Sheet: w.Doc.Sheet, Rows: w.Total})
	}

	outputs, err := finalizeNames(order, base)
	if err != nil {
		return res, err
	}
	res.Outputs = outputs
	return res, nil
}

func readZipEntry(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// finalizeNames drops the type and chunk suffixes that a given run does not
// need: a ZIP of one document type producing one file is named after the ZIP.
func finalizeNames(writers []*xlsx.Writer, base string) ([]string, error) {
	multiType := len(writers) > 1

	var outputs []string
	for _, w := range writers {
		multiChunk := len(w.Paths) > 1
		for i, path := range w.Paths {
			name := base
			if multiType {
				name += "_" + w.Doc.Name
			}
			if multiChunk {
				name += fmt.Sprintf("_parte%d", i+1)
			}
			name += ".xlsx"

			if name != path {
				if err := os.Rename(path, name); err != nil {
					return nil, fmt.Errorf("error renaming output: %w", err)
				}
			}
			outputs = append(outputs, name)
		}
	}
	return outputs, nil
}
