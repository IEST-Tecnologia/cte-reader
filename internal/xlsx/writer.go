// Package xlsx writes extracted rows to spreadsheets, spilling to a new file
// every so many rows so no single workbook grows unbounded.
package xlsx

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"

	"fiscal-reader/internal/spec"
)

const (
	maxColWidth = 60
	dateFormat  = "dd/mm/yyyy hh:mm"
)

// Writer accumulates rows for one document type.
type Writer struct {
	// Doc is the spec whose rows this writer holds.
	Doc *spec.Doc

	base      string // output path stem, without extension
	chunkSize int

	file      *excelize.File
	dateStyle int
	widths    []int
	row       int // next row to write
	inChunk   int
	chunk     int

	Total int
	Paths []string
}

// NewWriter starts a writer for one document type. Output files are named from
// base (a path stem without extension); a new one is started every chunkSize
// rows.
func NewWriter(doc *spec.Doc, base string, chunkSize int) *Writer {
	w := &Writer{Doc: doc, base: base, chunkSize: chunkSize}
	w.reset()
	return w
}

func (w *Writer) reset() {
	f := excelize.NewFile()
	f.SetSheetName(f.GetSheetName(0), w.Doc.Sheet)

	headers := w.Doc.Headers()
	w.widths = make([]int, len(headers))
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(w.Doc.Sheet, cell, h)
		w.widths[i] = utf8.RuneCountInString(h)
	}

	if style, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}}); err == nil {
		f.SetRowStyle(w.Doc.Sheet, 1, 1, style)
	}
	w.dateStyle, _ = f.NewStyle(&excelize.Style{CustomNumFmt: strPtr(dateFormat)})

	w.file = f
	w.row = 2
	w.inChunk = 0
}

func strPtr(s string) *string { return &s }

// Add writes every row of one document, then flushes if the chunk is full.
// Chunks break between documents, so an NF-e's items stay in one file.
func (w *Writer) Add(rows [][]any) error {
	for _, row := range rows {
		for i, v := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, w.row)
			if err := w.file.SetCellValue(w.Doc.Sheet, cell, v); err != nil {
				return fmt.Errorf("writing %s: %w", cell, err)
			}
			if _, ok := v.(time.Time); ok {
				w.file.SetCellStyle(w.Doc.Sheet, cell, cell, w.dateStyle)
			}
			if n := displayWidth(v); n > w.widths[i] {
				w.widths[i] = n
			}
		}
		w.row++
		w.inChunk++
		w.Total++
	}

	if w.inChunk >= w.chunkSize {
		return w.flush()
	}
	return nil
}

// Close writes the trailing chunk, if any rows are pending.
func (w *Writer) Close() error {
	if w.inChunk == 0 {
		return nil
	}
	return w.flush()
}

func (w *Writer) flush() error {
	for i, width := range w.widths {
		if width > maxColWidth {
			width = maxColWidth
		}
		name, _ := excelize.ColumnNumberToName(i + 1)
		w.file.SetColWidth(w.Doc.Sheet, name, name, float64(width)+2)
	}

	w.chunk++
	path := fmt.Sprintf("%s_%s_parte%d.xlsx", w.base, w.Doc.Name, w.chunk)
	if err := w.file.SaveAs(path); err != nil {
		return fmt.Errorf("saving %s: %w", path, err)
	}
	w.file.Close()
	w.Paths = append(w.Paths, path)

	w.reset()
	return nil
}

func displayWidth(v any) int {
	if t, ok := v.(time.Time); ok {
		return len(t.Format("02/01/2006 15:04"))
	}
	return utf8.RuneCountInString(fmt.Sprintf("%v", v))
}
