// Package xmltree parses XML into a namespace-agnostic element tree that can be
// addressed by path, so callers need no generated structs per document type.
package xmltree

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// Node is a namespace-agnostic XML element tree. Elements are keyed by their
// local name only, so the SEFAZ namespace declarations and any tag prefixes are
// irrelevant to lookups.
type Node struct {
	Name     string
	Attrs    map[string]string
	Text     string
	Children []*Node
}

// Parse builds a Node tree from raw XML bytes.
func Parse(data []byte) (*Node, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.CharsetReader = charsetReader

	var root *Node
	var stack []*Node

	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			n := &Node{Name: t.Name.Local}
			for _, a := range t.Attr {
				if n.Attrs == nil {
					n.Attrs = make(map[string]string, len(t.Attr))
				}
				n.Attrs[a.Name.Local] = a.Value
			}
			if len(stack) == 0 {
				root = n
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, n)
			}
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(t)
			}
		}
	}

	if root == nil {
		return nil, errors.New("no XML root element")
	}
	return root, nil
}

// charsetReader handles the ISO-8859-1 files some issuers still produce; any
// other declared encoding is passed through untouched.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "iso-8859-1", "latin1", "iso8859-1", "windows-1252":
		raw, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		buf.Grow(len(raw))
		for _, b := range raw {
			buf.WriteRune(rune(b))
		}
		return &buf, nil
	default:
		return input, nil
	}
}

// Wrap returns a synthetic parent named name holding n, so that documents
// distributed without their processing envelope can be read with the same paths
// as the enveloped form.
func (n *Node) Wrap(name string) *Node {
	return &Node{Name: name, Children: []*Node{n}}
}

// Find returns the first node reachable by path ("a/b/c"), or nil. A "*"
// segment matches any single element, which is how the ICMS variant groups
// (ICMS00, ICMS20, ...) are addressed without naming each one; "**" matches any
// number of levels, for groups whose nesting varies between layout versions.
func (n *Node) Find(path string) *Node {
	all := n.findAll(splitPath(path), true)
	if len(all) == 0 {
		return nil
	}
	return all[0]
}

// FindAll returns every node reachable by path, in document order.
func (n *Node) FindAll(path string) []*Node {
	return n.findAll(splitPath(path), false)
}

func (n *Node) findAll(segments []string, first bool) []*Node {
	if n == nil {
		return nil
	}
	if len(segments) == 0 {
		return []*Node{n}
	}

	var out []*Node
	seg := segments[0]
	rest := segments[1:]

	// "**" matches zero levels here, then any number of levels below, so the
	// shallowest match comes first.
	if seg == "**" {
		out = append(out, n.findAll(rest, first)...)
		if first && len(out) > 0 {
			return out[:1]
		}
		for _, c := range n.Children {
			found := c.findAll(segments, first)
			if len(found) == 0 {
				continue
			}
			if first {
				return found[:1]
			}
			out = append(out, found...)
		}
		return out
	}

	for _, c := range n.Children {
		if seg != "*" && c.Name != seg {
			continue
		}
		found := c.findAll(rest, first)
		if len(found) == 0 {
			continue
		}
		if first {
			return found[:1]
		}
		out = append(out, found...)
	}
	return out
}

// Value returns the trimmed text at path, or "" when the path is absent. A
// final "@name" segment reads that attribute instead of the element text.
func (n *Node) Value(path string) string {
	if n == nil {
		return ""
	}
	segments := splitPath(path)

	attr := ""
	if len(segments) > 0 {
		if last := segments[len(segments)-1]; strings.HasPrefix(last, "@") {
			attr = strings.TrimPrefix(last, "@")
			segments = segments[:len(segments)-1]
		}
	}

	target := n
	if len(segments) > 0 {
		matches := n.findAll(segments, true)
		if len(matches) == 0 {
			return ""
		}
		target = matches[0]
	}

	if attr != "" {
		return strings.TrimSpace(target.Attrs[attr])
	}
	return strings.TrimSpace(target.Text)
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
