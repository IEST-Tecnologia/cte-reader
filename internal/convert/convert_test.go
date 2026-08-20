package convert_test

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"

	"fiscal-reader/internal/convert"
	"fiscal-reader/internal/fixtures"
	"fiscal-reader/internal/spec"
)

// testChunkSize is large enough that only the chunking test splits files.
const testChunkSize = 1000

func writeZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "docs.zip")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func readSheet(t *testing.T, path, sheet string) [][]string {
	t.Helper()
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatalf("reading sheet %s: %v", sheet, err)
	}
	return rows
}

func run(t *testing.T, zipPath string, chunkSize int) convert.Result {
	t.Helper()
	res, err := convert.Run(zipPath, chunkSize, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestConvertSingleType(t *testing.T) {
	zipPath := writeZip(t, map[string]string{
		"a.xml":     fixtures.CTe,
		"notes.txt": "ignored, not xml",
	})

	res := run(t, zipPath, testChunkSize)
	if len(res.Outputs) != 1 {
		t.Fatalf("outputs = %v, want one file", res.Outputs)
	}
	// One type, one chunk: named after the ZIP, no suffixes.
	if want := filepath.Join(filepath.Dir(zipPath), "docs.xlsx"); res.Outputs[0] != want {
		t.Errorf("output = %s, want %s", res.Outputs[0], want)
	}
	if len(res.Counts) != 1 || res.Counts[0].Rows != 1 {
		t.Errorf("counts = %+v, want 1 CT-e row", res.Counts)
	}

	rows := readSheet(t, res.Outputs[0], spec.CTe.Sheet)
	if len(rows) != 2 {
		t.Fatalf("got %d rows including header, want 2", len(rows))
	}
	if rows[0][0] != "Número CT-e" {
		t.Errorf("header = %q", rows[0][0])
	}
	if rows[1][2] != "Transportadora Exemplo LTDA" {
		t.Errorf("emitter cell = %q", rows[1][2])
	}
}

func TestConvertAppliesCancellation(t *testing.T) {
	zipPath := writeZip(t, map[string]string{
		// The event sorts before the document, and must apply either way.
		"a-event.xml": fixtures.CancelEventNFe,
		"b-nfe.xml":   fixtures.NFe,
	})

	res := run(t, zipPath, testChunkSize)
	if res.Events != 1 || res.Cancelled != 1 {
		t.Errorf("events = %d, cancelled = %d, want 1 and 1", res.Events, res.Cancelled)
	}
	// The event file is consumed, not reported as an unsupported document.
	if res.Ignored != 0 || res.Skipped != 0 {
		t.Errorf("ignored = %d, skipped = %d, want 0 and 0", res.Ignored, res.Skipped)
	}
	if len(res.Outputs) != 1 {
		t.Fatalf("outputs = %v, want one workbook", res.Outputs)
	}

	rows := readSheet(t, res.Outputs[0], spec.NFe.Sheet)
	statusCol := -1
	for i, h := range rows[0] {
		if h == "Status" {
			statusCol = i
		}
	}
	if statusCol < 0 {
		t.Fatal("no Status column")
	}
	for _, row := range rows[1:] {
		if row[statusCol] != "Cancelado" {
			t.Errorf("status = %q, want Cancelado on every item row", row[statusCol])
		}
	}
}

func TestConvertMixedTypes(t *testing.T) {
	zipPath := writeZip(t, map[string]string{
		"cte.xml":    fixtures.CTe,
		"nfe.xml":    fixtures.NFe,
		"event.xml":  fixtures.CorrectionEventNFe,
		"outro.xml":  `<naoSuportado><x/></naoSuportado>`,
		"broken.xml": `<cteProc><CTe>`,
	})

	res := run(t, zipPath, testChunkSize)
	if len(res.Outputs) != 2 {
		t.Fatalf("outputs = %v, want one per document type", res.Outputs)
	}
	if res.Ignored != 1 {
		t.Errorf("ignored = %d, want 1 (the unsupported file)", res.Ignored)
	}
	// The event is read by the event pass, and cancels nothing.
	if res.Events != 1 || res.Cancelled != 0 {
		t.Errorf("events = %d, cancelled = %d, want 1 and 0", res.Events, res.Cancelled)
	}
	if res.Skipped != 1 {
		t.Errorf("skipped = %d, want 1 (the truncated file)", res.Skipped)
	}

	dir := filepath.Dir(zipPath)
	for _, want := range []string{"docs_cte.xlsx", "docs_nfe.xlsx"} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("missing expected output %s", want)
		}
	}

	// The NF-e workbook holds one row per item.
	nfeRows := readSheet(t, filepath.Join(dir, "docs_nfe.xlsx"), spec.NFe.Sheet)
	if len(nfeRows) != 3 {
		t.Errorf("NF-e sheet has %d rows including header, want 3", len(nfeRows))
	}
}

func TestConvertChunking(t *testing.T) {
	files := map[string]string{}
	for i := 0; i < 5; i++ {
		files[string(rune('a'+i))+".xml"] = fixtures.CTe
	}
	zipPath := writeZip(t, files)

	res := run(t, zipPath, 2)
	// 5 rows at 2 per chunk: three files, the last holding one row.
	if len(res.Outputs) != 3 {
		t.Fatalf("outputs = %v, want 3 chunks", res.Outputs)
	}
	if want := filepath.Join(filepath.Dir(zipPath), "docs_parte1.xlsx"); res.Outputs[0] != want {
		t.Errorf("first chunk = %s, want %s", res.Outputs[0], want)
	}

	total := 0
	for _, path := range res.Outputs {
		rows := readSheet(t, path, spec.CTe.Sheet)
		if rows[0][0] != "Número CT-e" {
			t.Errorf("%s is missing its header row", path)
		}
		total += len(rows) - 1
	}
	if total != 5 {
		t.Errorf("rows across chunks = %d, want 5", total)
	}
	if res.Counts[0].Rows != 5 {
		t.Errorf("reported count = %d, want 5", res.Counts[0].Rows)
	}
}

// A ZIP with nothing to extract must not leave an empty workbook behind.
func TestConvertNoDocuments(t *testing.T) {
	zipPath := writeZip(t, map[string]string{"outro.xml": `<naoSuportado><x/></naoSuportado>`})

	res := run(t, zipPath, testChunkSize)
	if len(res.Outputs) != 0 || len(res.Counts) != 0 {
		t.Errorf("outputs = %v, counts = %+v, want none", res.Outputs, res.Counts)
	}
	if res.Ignored != 1 {
		t.Errorf("ignored = %d, want 1", res.Ignored)
	}
}

func TestConvertMissingZip(t *testing.T) {
	if _, err := convert.Run(filepath.Join(t.TempDir(), "nope.zip"), testChunkSize, io.Discard); err == nil {
		t.Error("expected an error for a missing ZIP")
	}
}
