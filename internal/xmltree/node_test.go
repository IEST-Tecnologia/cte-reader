package xmltree_test

import (
	"testing"

	"fiscal-reader/internal/fixtures"
	"fiscal-reader/internal/xmltree"
)

func mustParse(t *testing.T, s string) *xmltree.Node {
	t.Helper()
	n, err := xmltree.Parse([]byte(s))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return n
}

func TestParseIgnoresNamespaces(t *testing.T) {
	root := mustParse(t, fixtures.CTe)
	if root.Name != "cteProc" {
		t.Fatalf("root = %q, want cteProc", root.Name)
	}
	if got := root.Value("CTe/infCte/emit/xNome"); got != "Transportadora Exemplo LTDA" {
		t.Errorf("emitter = %q", got)
	}
}

func TestParseHandlesPrefixedTags(t *testing.T) {
	root := mustParse(t, `<ns:cteProc xmlns:ns="http://x"><ns:CTe><ns:infCte><ns:ide><ns:nCT>42</ns:nCT></ns:ide></ns:infCte></ns:CTe></ns:cteProc>`)
	if got := root.Value("CTe/infCte/ide/nCT"); got != "42" {
		t.Errorf("nCT = %q, want 42", got)
	}
}

func TestParseLatin1(t *testing.T) {
	// "Sã" encoded as ISO-8859-1: 0xE3 for "ã".
	raw := append([]byte(`<?xml version="1.0" encoding="ISO-8859-1"?><cteProc><CTe><infCte><ide><xMunIni>S`), 0xE3, 'o')
	raw = append(raw, []byte(` Paulo</xMunIni></ide></infCte></CTe></cteProc>`)...)

	root, err := xmltree.Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := root.Value("CTe/infCte/ide/xMunIni"); got != "São Paulo" {
		t.Errorf("xMunIni = %q, want São Paulo", got)
	}
}

func TestValueWildcardAndAttribute(t *testing.T) {
	root := mustParse(t, fixtures.CTe)

	if got := root.Value("CTe/infCte/imp/ICMS/*/pICMS"); got != "12.00" {
		t.Errorf("pICMS via wildcard = %q", got)
	}
	if got := root.Value("CTe/infCte/@Id"); got != "CTe35260355442952000157570010152386931847651990" {
		t.Errorf("Id attribute = %q", got)
	}
	if got := root.Value("CTe/infCte/ide/naoExiste"); got != "" {
		t.Errorf("missing path = %q, want empty", got)
	}
}

func TestFindAll(t *testing.T) {
	root := mustParse(t, `<nfeProc><NFe><infNFe><det nItem="1"/><det nItem="2"/><det nItem="3"/></infNFe></NFe></nfeProc>`)

	items := root.FindAll("NFe/infNFe/det")
	if len(items) != 3 {
		t.Fatalf("got %d det nodes, want 3", len(items))
	}
	if got := items[1].Value("@nItem"); got != "2" {
		t.Errorf("second item = %q, want 2", got)
	}
}

func TestValueDescendantWildcard(t *testing.T) {
	root := mustParse(t, `<nfeProc><NFe><infNFe><det nItem="1"><imposto><IBSCBS><CST>000</CST>
		<gIBSCBS><vBC>25.00</vBC>
			<gIBSUF><pIBSUF>0.1000</pIBSUF><vIBSUF>0.03</vIBSUF></gIBSUF>
			<gCBS><vCBS>0.23</vCBS></gCBS>
		</gIBSCBS></IBSCBS></imposto></det></infNFe></NFe></nfeProc>`)

	det := root.Find("NFe/infNFe/det")
	if det == nil {
		t.Fatal("det not found")
	}

	// "**" crosses however many group levels the layout nests.
	if got := det.Value("imposto/IBSCBS/**/vCBS"); got != "0.23" {
		t.Errorf("vCBS through ** = %q", got)
	}
	if got := det.Value("imposto/IBSCBS/**/vIBSUF"); got != "0.03" {
		t.Errorf("vIBSUF through ** = %q", got)
	}
	// It also matches zero levels, so a direct child still resolves.
	if got := det.Value("imposto/**/CST"); got != "000" {
		t.Errorf("CST through ** = %q", got)
	}
	if got := det.Value("imposto/IBSCBS/**/vNaoExiste"); got != "" {
		t.Errorf("missing name through ** = %q, want empty", got)
	}
}

// "**" returns the shallowest match first, so a nested name of the same
// element does not shadow the outer one.
func TestDescendantWildcardPrefersShallowest(t *testing.T) {
	root := mustParse(t, `<a><b><v>outer</v><c><v>inner</v></c></b></a>`)

	if got := root.Value("b/**/v"); got != "outer" {
		t.Errorf("got %q, want outer", got)
	}
	if all := root.FindAll("b/**/v"); len(all) != 2 {
		t.Errorf("FindAll returned %d nodes, want 2", len(all))
	}
}
