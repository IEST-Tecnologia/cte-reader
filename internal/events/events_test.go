package events_test

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"fiscal-reader/internal/events"
	"fiscal-reader/internal/fixtures"
)

const nfeKey = "35260311222333000181550010000012341000012348"

func scan(t *testing.T, files map[string]string) events.Scan {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
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

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return events.ScanZip(r, io.Discard)
}

func TestScanFindsCancellation(t *testing.T) {
	got := scan(t, map[string]string{
		"nfe.xml":   fixtures.NFe,
		"event.xml": fixtures.CancelEventNFe,
	})

	if !got.Cancelled[nfeKey] {
		t.Errorf("cancelled = %v, want the NF-e key", got.Cancelled)
	}
	if got.Total != 1 {
		t.Errorf("total = %d, want 1", got.Total)
	}
	// The document itself must not be claimed by the event pass.
	if got.Files["nfe.xml"] || !got.Files["event.xml"] {
		t.Errorf("files = %v", got.Files)
	}
}

// Only cancellations cancel: a letter of correction is an event like any other.
func TestScanIgnoresNonCancellationEvents(t *testing.T) {
	got := scan(t, map[string]string{"cc.xml": fixtures.CorrectionEventNFe})

	if len(got.Cancelled) != 0 {
		t.Errorf("cancelled = %v, want none", got.Cancelled)
	}
	if got.Total != 1 || !got.Files["cc.xml"] {
		t.Errorf("the correction should still count as an event: %+v", got)
	}
}

// A cancellation the authority rejected leaves the document standing.
func TestScanIgnoresRejectedCancellation(t *testing.T) {
	got := scan(t, map[string]string{"rejected.xml": fixtures.RejectedCancelEventNFe})

	if len(got.Cancelled) != 0 {
		t.Errorf("cancelled = %v, want none", got.Cancelled)
	}
}

// A rejected cancellation does not undo an accepted one for the same document.
func TestScanAcceptedWinsOverRejected(t *testing.T) {
	got := scan(t, map[string]string{
		"accepted.xml": fixtures.CancelEventNFe,
		"rejected.xml": fixtures.RejectedCancelEventNFe,
	})

	if !got.Cancelled[nfeKey] {
		t.Errorf("cancelled = %v, want the NF-e key", got.Cancelled)
	}
	if got.Total != 2 {
		t.Errorf("total = %d, want 2", got.Total)
	}
}

// An event distributed without the authority's reply has no cStat to check, so
// it is trusted.
func TestScanBareEventWithoutReply(t *testing.T) {
	got := scan(t, map[string]string{"bare.xml": `<evento versao="1.00"><infEvento Id="ID110111` + nfeKey + `01">
		<chNFe>` + nfeKey + `</chNFe><tpEvento>110111</tpEvento></infEvento></evento>`})

	if !got.Cancelled[nfeKey] {
		t.Errorf("cancelled = %v, want the NF-e key", got.Cancelled)
	}
}

// Without chNFe the key is recovered from the event's Id attribute.
func TestScanKeyFromIDAttribute(t *testing.T) {
	got := scan(t, map[string]string{"noch.xml": `<evento><infEvento Id="ID110111` + nfeKey + `01">
		<tpEvento>110111</tpEvento></infEvento></evento>`})

	if !got.Cancelled[nfeKey] {
		t.Errorf("cancelled = %v, want the key recovered from Id", got.Cancelled)
	}
}

// CT-e cancellations use the same event structure, keyed by chCTe.
func TestScanCteCancellation(t *testing.T) {
	const cteKey = "35260355442952000157570010152386931847651990"
	got := scan(t, map[string]string{"cte-event.xml": `<procEventoCTe><eventoCTe><infEvento>
		<chCTe>` + cteKey + `</chCTe><tpEvento>110111</tpEvento></infEvento></eventoCTe>
		<retEventoCTe><infEvento><cStat>135</cStat></infEvento></retEventoCTe></procEventoCTe>`})

	if !got.Cancelled[cteKey] {
		t.Errorf("cancelled = %v, want the CT-e key", got.Cancelled)
	}
}

// Ordinary documents are not events, however many times they say "evento".
func TestScanIgnoresDocuments(t *testing.T) {
	got := scan(t, map[string]string{"nfe.xml": fixtures.NFe, "cte.xml": fixtures.CTe})

	if got.Total != 0 || len(got.Files) != 0 || len(got.Cancelled) != 0 {
		t.Errorf("scan of plain documents = %+v, want nothing", got)
	}
}
