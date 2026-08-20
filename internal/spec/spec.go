// Package spec declares what to extract from each document type. A Doc names
// the XML roots it claims and lists the columns to produce; adding a document
// type or changing a report means editing a table, not writing code.
package spec

import (
	"strconv"
	"strings"
	"time"

	"fiscal-reader/internal/xmltree"
)

// Kind controls how a value is written to the spreadsheet. Text keeps the raw
// string (right for codes: CNPJ, CFOP, CST, access keys), Number writes a real
// number so the column can be summed, Date writes a real date so it can be
// sorted and filtered.
type Kind int

const (
	KindText Kind = iota
	KindNumber
	KindDate
)

// Column is one spreadsheet column. Paths are evaluated against the row node
// first and then against the whole document, so a spec with Repeat set can mix
// item-level and document-level fields freely.
//
// With several Paths and no Join, the first non-empty one wins (CNPJ else CPF).
// With Join set, every non-empty value is joined with it (city + " - " + state).
type Column struct {
	Header string
	Kind   Kind
	Paths  []string
	Join   string

	// Transform rewrites the raw string before it is typed; nil leaves it alone.
	Transform func(string) string

	// Resolve gets the last word on the value, given what the paths extracted
	// and what the rest of the archive says. Nil keeps the extracted value.
	Resolve func(row, doc *xmltree.Node, extracted string, env Env) string
}

// Env carries facts gathered from the rest of the archive, which a document
// cannot know about on its own.
type Env struct {
	// Cancelled holds the access keys cancelled by an event document found
	// elsewhere in the same archive.
	Cancelled map[string]bool
}

// trimPrefix strips a fixed prefix, e.g. the "NFe" that precedes the access key
// inside the infNFe/@Id attribute.
func (c Column) trimPrefix(prefix string) Column {
	c.Transform = func(v string) string { return strings.TrimPrefix(v, prefix) }
	return c
}

// cancelledBy overwrites the column with "Cancelado" when the access key at any
// of keyPaths was cancelled by an event elsewhere in the archive. The document
// itself cannot report this: its own protocol still says authorised.
func (c Column) cancelledBy(keyPaths ...string) Column {
	c.Resolve = func(row, doc *xmltree.Node, extracted string, env Env) string {
		for _, p := range keyPaths {
			if key := accessKey(lookup(row, doc, p)); key != "" && env.Cancelled[key] {
				return "Cancelado"
			}
		}
		return extracted
	}
	return c
}

// accessKey strips the document-type prefix that the Id attributes carry, so a
// key read from infNFe/@Id compares equal to one read from chNFe.
func accessKey(v string) string {
	for _, prefix := range []string{"NFe", "CTe"} {
		v = strings.TrimPrefix(v, prefix)
	}
	return v
}

// decode replaces coded values with their meaning, leaving unlisted codes as
// they are.
func (c Column) decode(codes map[string]string) Column {
	c.Transform = func(v string) string {
		if s, ok := codes[v]; ok {
			return s
		}
		return v
	}
	return c
}

// Doc describes one document type: which XML roots it claims, and which
// columns to extract.
type Doc struct {
	Name    string   // filename-safe identifier, e.g. "cte"
	Sheet   string   // worksheet name, e.g. "CT-e"
	Root    string   // canonical root element, e.g. "cteProc"
	Aliases []string // roots wrapped into Root, e.g. "CTe" for unprocessed files
	Repeat  string   // optional path to a repeating node: one output row per match
	Columns []Column
}

// col declares a column whose value is the first non-empty path.
func col(header string, kind Kind, paths ...string) Column {
	return Column{Header: header, Kind: kind, Paths: paths}
}

// joined declares a column that concatenates every non-empty path.
func joined(header, sep string, paths ...string) Column {
	return Column{Header: header, Kind: KindText, Paths: paths, Join: sep}
}

// All is the registry. Add a document type by appending its Doc here.
var All = []*Doc{CTe, NFe}

// ForRoot matches a parsed document to its spec, normalising unenveloped
// documents so all paths in a spec are written against the canonical root.
// It returns nil when no spec claims the document's root element.
func ForRoot(doc *xmltree.Node) (*Doc, *xmltree.Node) {
	for _, s := range All {
		if doc.Name == s.Root {
			return s, doc
		}
		for _, alias := range s.Aliases {
			if doc.Name == alias {
				return s, doc.Wrap(s.Root)
			}
		}
	}
	return nil, nil
}

// Headers returns the header row.
func (s *Doc) Headers() []string {
	out := make([]string, len(s.Columns))
	for i, c := range s.Columns {
		out[i] = c.Header
	}
	return out
}

// Rows extracts every output row for one document. Without Repeat that is
// exactly one row; with Repeat it is one row per matching node (NF-e items, for
// example), and a document with no matches yields no rows.
func (s *Doc) Rows(doc *xmltree.Node, env Env) [][]any {
	bases := []*xmltree.Node{doc}
	if s.Repeat != "" {
		bases = doc.FindAll(s.Repeat)
	}

	rows := make([][]any, 0, len(bases))
	for _, base := range bases {
		row := make([]any, len(s.Columns))
		for i, c := range s.Columns {
			row[i] = c.value(base, doc, env)
		}
		rows = append(rows, row)
	}
	return rows
}

func (c Column) value(row, doc *xmltree.Node, env Env) any {
	var raw string
	if c.Join != "" {
		var parts []string
		for _, p := range c.Paths {
			if v := c.apply(lookup(row, doc, p)); v != "" {
				parts = append(parts, v)
			}
		}
		raw = strings.Join(parts, c.Join)
	} else {
		for _, p := range c.Paths {
			if v := c.apply(lookup(row, doc, p)); v != "" {
				raw = v
				break
			}
		}
	}

	if c.Resolve != nil {
		raw = c.Resolve(row, doc, raw, env)
	}
	if raw == "" {
		return ""
	}
	return castValue(raw, c.Kind)
}

func (c Column) apply(v string) string {
	if c.Transform == nil || v == "" {
		return v
	}
	return c.Transform(v)
}

// lookup resolves a path against the row node, falling back to the document.
func lookup(row, doc *xmltree.Node, path string) string {
	if v := row.Value(path); v != "" {
		return v
	}
	if doc != row {
		return doc.Value(path)
	}
	return ""
}

var dateLayouts = []string{
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// castValue turns a raw XML string into the cell value for its Kind, falling back
// to the raw string whenever it does not parse.
func castValue(v string, kind Kind) any {
	switch kind {
	case KindNumber:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	case KindDate:
		for _, layout := range dateLayouts {
			t, err := time.Parse(layout, v)
			if err != nil {
				continue
			}
			// Keep the wall clock the issuer wrote; the offset would otherwise
			// shift the displayed time when Excel renders the serial value.
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
		}
	}
	return v
}
