// Package events reads the event documents that accompany fiscal documents in
// an archive — cancellations, corrections, acknowledgements — and reports which
// access keys they cancel.
package events

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"fiscal-reader/internal/xmltree"
)

// Cancellation event codes: 110111 is the ordinary cancellation, 110112 the
// cancellation by substitution.
var cancelCodes = map[string]bool{"110111": true, "110112": true}

// Codes the authority returns for an accepted event. Anything else — a
// rejection, a duplicate — leaves the document standing.
var acceptedCodes = map[string]bool{
	"135": true, // evento registrado e vinculado
	"136": true, // evento registrado, não vinculado
	"155": true, // cancelamento homologado fora de prazo
}

// eventRoots are the roots this package claims. Bare evento or retEvento files
// occur when an issuer distributes the halves separately.
var eventRoots = map[string]bool{
	"procEventoNFe": true,
	"procEventoCTe": true,
	"evento":        true,
	"retEvento":     true,
}

// Entries are checked for this marker before being parsed, so a large archive of
// ordinary documents is not parsed twice.
var eventMarker = []byte("vento")

// Scan is what a pass over the archive's event documents found.
type Scan struct {
	// Cancelled holds the access keys an accepted cancellation event applies to.
	Cancelled map[string]bool
	// Files names the ZIP entries that were event documents, so the main pass
	// can skip them instead of reporting them as unsupported.
	Files map[string]bool
	// Total counts the event documents seen, cancellations or not.
	Total int
}

// ScanZip reads every event document in r. Problems with individual files are
// reported to logw and otherwise ignored: without a readable event, a document
// simply keeps the status its own protocol reports.
func ScanZip(r *zip.Reader, logw io.Writer) Scan {
	scan := Scan{Cancelled: map[string]bool{}, Files: map[string]bool{}}

	for _, file := range r.File {
		if !strings.HasSuffix(strings.ToLower(file.Name), ".xml") {
			continue
		}

		data, err := readEntry(file)
		if err != nil {
			// The main pass reports this file as skipped; nothing to add here.
			continue
		}
		if !bytes.Contains(data, eventMarker) {
			continue
		}

		root, err := xmltree.Parse(data)
		if err != nil || !eventRoots[root.Name] {
			continue
		}

		scan.Files[file.Name] = true
		scan.Total++

		key, ok := cancels(root)
		if !ok {
			continue
		}
		if key == "" {
			fmt.Fprintf(logw, "EVENT %s: cancellation without an access key\n", file.Name)
			continue
		}
		scan.Cancelled[key] = true
	}
	return scan
}

// cancels reports whether the event cancels its document, and for which access
// key. The event half carries tpEvento and the key; the authority's reply half
// carries cStat. A file holding only the event half is trusted, since there is
// no reply to check it against.
func cancels(root *xmltree.Node) (string, bool) {
	if !cancelCodes[root.Value("**/infEvento/tpEvento")] {
		return "", false
	}
	if cStat := root.Value("**/infEvento/cStat"); cStat != "" && !acceptedCodes[cStat] {
		return "", false
	}

	for _, path := range []string{"**/infEvento/chNFe", "**/infEvento/chCTe"} {
		if key := root.Value(path); key != "" {
			return key, true
		}
	}
	return keyFromID(root.Value("**/infEvento/@Id")), true
}

// keyFromID recovers the access key from an event's Id attribute, which is
// "ID" + the six-digit event code + the 44-digit key + the sequence number.
func keyFromID(id string) string {
	const prefix, code, key = 2, 6, 44
	if len(id) < prefix+code+key || !strings.HasPrefix(id, "ID") {
		return ""
	}
	return id[prefix+code : prefix+code+key]
}

func readEntry(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
