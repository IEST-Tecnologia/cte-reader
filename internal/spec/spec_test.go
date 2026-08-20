package spec_test

import (
	"testing"
	"time"

	"fiscal-reader/internal/fixtures"
	"fiscal-reader/internal/spec"
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

func rowMap(t *testing.T, doc *spec.Doc, row []any) map[string]any {
	t.Helper()
	if len(row) != len(doc.Columns) {
		t.Fatalf("row has %d values, spec has %d columns", len(row), len(doc.Columns))
	}
	m := make(map[string]any, len(row))
	for i, c := range doc.Columns {
		m[c.Header] = row[i]
	}
	return m
}

func TestSpecForRoot(t *testing.T) {
	tests := []struct {
		xml      string
		wantSpec string
	}{
		{fixtures.CTe, "cte"},
		{fixtures.NFe, "nfe"},
		{`<CTe><infCte><ide><nCT>7</nCT></ide></infCte></CTe>`, "cte"},
		{`<NFe><infNFe><det nItem="1"><prod><cProd>X</cProd></prod></det></infNFe></NFe>`, "nfe"},
	}
	for _, tc := range tests {
		doc, _ := spec.ForRoot(mustParse(t, tc.xml))
		if doc == nil || doc.Name != tc.wantSpec {
			t.Errorf("ForRoot = %v, want %s", doc, tc.wantSpec)
		}
	}

	if doc, _ := spec.ForRoot(mustParse(t, `<procEventoCTe><evento/></procEventoCTe>`)); doc != nil {
		t.Errorf("event document matched spec %s, want no match", doc.Name)
	}
}

// An unenveloped <CTe> is wrapped so the spec's paths still resolve.
func TestSpecForRootWrapsUnenveloped(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, `<CTe><infCte><ide><nCT>7</nCT></ide></infCte></CTe>`))
	if doc == nil {
		t.Fatal("no spec matched")
	}
	if got := node.Value("CTe/infCte/ide/nCT"); got != "7" {
		t.Errorf("nCT after wrapping = %q, want 7", got)
	}
}

func TestCteRows(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, fixtures.CTe))
	rows := doc.Rows(node, spec.Env{})
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	got := rowMap(t, doc, rows[0])

	want := map[string]any{
		"Número CT-e":               float64(1523869),
		"Razão Social Emitente":     "Transportadora Exemplo LTDA",
		"CNPJ Emitente":             "55442952000157",
		"CFOP":                      "6352",
		"Início da Prestação":       "São Paulo - SP",
		"Término da Prestação":      "Curitiba - PR",
		"Valor Total do Serviço":    1234.56,
		"Situação Tributária (CST)": "20",
		"Alíquota ICMS (%)":         12.0,
		"Chave de Acesso":           "35260355442952000157570010152386931847651990",
	}
	for header, wantVal := range want {
		if got[header] != wantVal {
			t.Errorf("%s = %#v, want %#v", header, got[header], wantVal)
		}
	}

	emitted, ok := got["Data de Emissão"].(time.Time)
	if !ok {
		t.Fatalf("Data de Emissão = %#v, want time.Time", got["Data de Emissão"])
	}
	if want := time.Date(2026, 3, 11, 14, 32, 5, 0, time.UTC); !emitted.Equal(want) {
		t.Errorf("Data de Emissão = %v, want %v", emitted, want)
	}
}

// A missing ICMS group leaves its columns empty rather than failing the row.
func TestCteRowsMissingFields(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, `<cteProc><CTe><infCte><ide><nCT>9</nCT></ide></infCte></CTe></cteProc>`))
	got := rowMap(t, doc, doc.Rows(node, spec.Env{})[0])

	if got["Número CT-e"] != float64(9) {
		t.Errorf("nCT = %#v", got["Número CT-e"])
	}
	for _, header := range []string{"Situação Tributária (CST)", "Valor ICMS", "Chave de Acesso", "Início da Prestação"} {
		if got[header] != "" {
			t.Errorf("%s = %#v, want empty string", header, got[header])
		}
	}
}

// Repeat fans one document out into one row per item, and document-level
// columns are repeated on each row.
func TestNfeRowsPerItem(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, fixtures.NFe))
	rows := doc.Rows(node, spec.Env{})
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}

	first := rowMap(t, doc, rows[0])
	second := rowMap(t, doc, rows[1])

	if first["Código do Produto"] != "A-1" || second["Código do Produto"] != "B-2" {
		t.Errorf("item codes = %v / %v", first["Código do Produto"], second["Código do Produto"])
	}
	if first["Item"] != float64(1) || second["Item"] != float64(2) {
		t.Errorf("item numbers = %v / %v", first["Item"], second["Item"])
	}
	if first["Quantidade"] != 10.0 || second["Quantidade"] != 4.0 {
		t.Errorf("quantities = %v / %v", first["Quantidade"], second["Quantidade"])
	}
	// Document-level values resolve through the fallback on every row.
	for i, row := range []map[string]any{first, second} {
		if row["Total da NF-e"] != 29.0 {
			t.Errorf("row %d total = %#v, want 29", i, row["Total da NF-e"])
		}
		if row["Razão Social Emitente"] != "Indústria Exemplo SA" {
			t.Errorf("row %d emitter = %#v", i, row["Razão Social Emitente"])
		}
	}
}

func TestNfeIdentification(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, fixtures.NFe))
	got := rowMap(t, doc, doc.Rows(node, spec.Env{})[0])

	want := map[string]any{
		"Número NF-e":          float64(1234),
		"Série":                "1",
		"Chave de Acesso":      "35260311222333000181550010000012341000012348",
		"Natureza da Operação": "Venda de mercadoria",
		"Tipo de Operação":     "Saída",
		"Status":               "Autorizado o uso da NF-e",
	}
	for header, wantVal := range want {
		if got[header] != wantVal {
			t.Errorf("%s = %#v, want %#v", header, got[header], wantVal)
		}
	}

	if emitted, ok := got["Data de Emissão"].(time.Time); !ok || !emitted.Equal(time.Date(2026, 3, 11, 9, 5, 0, 0, time.UTC)) {
		t.Errorf("Data de Emissão = %#v", got["Data de Emissão"])
	}
	if left, ok := got["Data/Hora de Saída ou Entrada"].(time.Time); !ok || !left.Equal(time.Date(2026, 3, 11, 18, 30, 0, 0, time.UTC)) {
		t.Errorf("Data/Hora de Saída ou Entrada = %#v", got["Data/Hora de Saída ou Entrada"])
	}
}

func TestNfeParties(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, fixtures.NFe))
	got := rowMap(t, doc, doc.Rows(node, spec.Env{})[0])

	want := map[string]any{
		"Razão Social Emitente":     "Indústria Exemplo SA",
		"CNPJ Emitente":             "11222333000181",
		"IE Emitente":               "123456789012",
		"Município Emitente":        "São Paulo",
		"UF Emitente":               "SP",
		"Razão Social Destinatário": "Cliente Pessoa Física",
		// CNPJ else CPF: the first non-empty path wins.
		"CNPJ/CPF Destinatário":  "12345678909",
		"IE Destinatário":        "ISENTO",
		"Município Destinatário": "Nova Iguaçu",
		"UF Destinatário":        "RJ",
	}
	for header, wantVal := range want {
		if got[header] != wantVal {
			t.Errorf("%s = %#v, want %#v", header, got[header], wantVal)
		}
	}
}

// Item taxes come from whichever group the item uses: ICMS00 with IPI and
// IBS/CBS on the first item, ICMS10 with ST and DIFAL on the second.
func TestNfeItemTaxes(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, fixtures.NFe))
	rows := doc.Rows(node, spec.Env{})
	first := rowMap(t, doc, rows[0])
	second := rowMap(t, doc, rows[1])

	wantFirst := map[string]any{
		"Origem da Mercadoria":             "0",
		"CST do ICMS":                      "00",
		"Base de Cálculo do ICMS":          25.0,
		"Alíquota do ICMS (%)":             18.0,
		"Valor do ICMS":                    4.5,
		"CST do IPI":                       "50",
		"Enquadramento Legal do IPI":       "999",
		"Base de Cálculo do IPI":           25.0,
		"Alíquota do IPI (%)":              5.0,
		"Valor do IPI":                     1.25,
		"CST do PIS":                       "01",
		"Alíquota do PIS (%)":              1.65,
		"Valor do PIS":                     0.41,
		"CST da COFINS":                    "01",
		"Alíquota da COFINS (%)":           7.6,
		"Valor da COFINS":                  1.9,
		"CST do IBS/CBS":                   "000",
		"Classificação Tributária IBS/CBS": "000001",
		"Base de Cálculo IBS/CBS":          25.0,
		"Alíquota IBS Estadual (%)":        0.1,
		"Valor IBS Estadual":               0.03,
		"Alíquota IBS Municipal (%)":       0.05,
		"Valor IBS Municipal":              0.01,
		"Valor Total do IBS":               0.04,
		"Alíquota CBS (%)":                 0.9,
		"Valor da CBS":                     0.23,
	}
	for header, wantVal := range wantFirst {
		if first[header] != wantVal {
			t.Errorf("item 1 %s = %#v, want %#v", header, first[header], wantVal)
		}
	}

	wantSecond := map[string]any{
		"CST do ICMS":                "10",
		"Base de Cálculo do ICMS-ST": 6.0,
		"Alíquota do ICMS-ST (%)":    18.0,
		"Valor do ICMS-ST":           0.6,
		"Base de Cálculo do DIFAL":   4.0,
		"Valor do ICMS UF Destino":   0.24,
		"CST do PIS":                 "07",
	}
	for header, wantVal := range wantSecond {
		if second[header] != wantVal {
			t.Errorf("item 2 %s = %#v, want %#v", header, second[header], wantVal)
		}
	}

	// Groups the item does not carry leave their columns empty.
	for _, header := range []string{"Valor do IPI", "Valor da CBS", "Base de Cálculo IBS/CBS", "Valor do PIS"} {
		if second[header] != "" {
			t.Errorf("item 2 %s = %#v, want empty", header, second[header])
		}
	}
	// ...and the first item carries no ST or DIFAL.
	for _, header := range []string{"Valor do ICMS-ST", "Valor do ICMS UF Destino"} {
		if first[header] != "" {
			t.Errorf("item 1 %s = %#v, want empty", header, first[header])
		}
	}
}

func TestNfeTotals(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, fixtures.NFe))
	got := rowMap(t, doc, doc.Rows(node, spec.Env{})[0])

	want := map[string]any{
		"Base de Cálculo do ICMS (Total)":    29.0,
		"Valor do ICMS (Total)":              4.98,
		"Base de Cálculo do ICMS-ST (Total)": 6.0,
		"Valor do ICMS-ST (Total)":           0.6,
		"Valor do DIFAL (Total)":             0.24,
		"Valor do IPI (Total)":               1.25,
		"Valor do PIS (Total)":               0.41,
		"Valor da COFINS (Total)":            1.9,
		"Base de Cálculo IBS/CBS (Total)":    29.0,
		"Valor IBS Estadual (Total)":         0.03,
		"Valor IBS Municipal (Total)":        0.01,
		"Valor Total do IBS (Total)":         0.04,
		"Valor da CBS (Total)":               0.23,
		"Total dos Produtos":                 29.0,
		"Total da NF-e":                      29.0,
	}
	for header, wantVal := range want {
		if got[header] != wantVal {
			t.Errorf("%s = %#v, want %#v", header, got[header], wantVal)
		}
	}
}

// Simples Nacional items carry CSOSN where others carry CST.
func TestNfeSimplesNacional(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, `<nfeProc><NFe><infNFe><det nItem="1">
		<prod><cProd>X</cProd></prod>
		<imposto><ICMS><ICMSSN102><orig>0</orig><CSOSN>102</CSOSN></ICMSSN102></ICMS></imposto>
	</det></infNFe></NFe></nfeProc>`))
	got := rowMap(t, doc, doc.Rows(node, spec.Env{})[0])

	if got["CST do ICMS"] != "102" {
		t.Errorf("CSOSN fallback = %#v, want 102", got["CST do ICMS"])
	}
}

// Without a protocol the access key comes from infNFe/@Id, minus its "NFe"
// prefix.
func TestNfeKeyFromIdAttribute(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, `<NFe><infNFe Id="NFe35260311222333000181550010000012341000012348">
		<det nItem="1"><prod><cProd>X</cProd></prod></det></infNFe></NFe>`))
	got := rowMap(t, doc, doc.Rows(node, spec.Env{})[0])

	if got["Chave de Acesso"] != "35260311222333000181550010000012341000012348" {
		t.Errorf("key = %#v, want the Id without its NFe prefix", got["Chave de Acesso"])
	}
}

// A document with no matching Repeat node contributes no rows.
func TestRepeatWithNoMatches(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, `<nfeProc><NFe><infNFe><ide><nNF>1</nNF></ide></infNFe></NFe></nfeProc>`))
	if rows := doc.Rows(node, spec.Env{}); len(rows) != 0 {
		t.Errorf("got %d rows, want 0", len(rows))
	}
}

// Every registered spec must have unique, non-empty identifiers.
func TestSpecRegistryIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range spec.All {
		if s.Name == "" || s.Sheet == "" || s.Root == "" || len(s.Columns) == 0 {
			t.Errorf("spec %q is incomplete", s.Name)
		}
		for _, root := range append([]string{s.Root}, s.Aliases...) {
			if seen[root] {
				t.Errorf("root %q claimed by more than one spec", root)
			}
			seen[root] = true
		}
		for _, c := range s.Columns {
			if c.Header == "" || len(c.Paths) == 0 {
				t.Errorf("spec %q has a column without header or paths", s.Name)
			}
		}
	}
}

// The archive-wide environment overrides the status the document reports about
// itself.
func TestNfeCancelledByEvent(t *testing.T) {
	doc, node := spec.ForRoot(mustParse(t, fixtures.NFe))
	const key = "35260311222333000181550010000012341000012348"

	got := rowMap(t, doc, doc.Rows(node, spec.Env{})[0])
	if got["Status"] != "Autorizado o uso da NF-e" {
		t.Errorf("status without events = %#v", got["Status"])
	}

	env := spec.Env{Cancelled: map[string]bool{key: true}}
	got = rowMap(t, doc, doc.Rows(node, env)[0])
	if got["Status"] != "Cancelado" {
		t.Errorf("status with a cancellation = %#v, want Cancelado", got["Status"])
	}

	// A cancellation for some other document leaves this one alone.
	env = spec.Env{Cancelled: map[string]bool{"99999999999999999999999999999999999999999999": true}}
	got = rowMap(t, doc, doc.Rows(node, env)[0])
	if got["Status"] != "Autorizado o uso da NF-e" {
		t.Errorf("status = %#v, want the document's own protocol", got["Status"])
	}
}

// Unprotocolled documents are matched on the key held in infNFe/@Id.
func TestNfeCancelledByEventWithoutProtocol(t *testing.T) {
	const key = "35260311222333000181550010000012341000012348"
	doc, node := spec.ForRoot(mustParse(t, `<NFe><infNFe Id="NFe`+key+`">
		<det nItem="1"><prod><cProd>X</cProd></prod></det></infNFe></NFe>`))

	env := spec.Env{Cancelled: map[string]bool{key: true}}
	got := rowMap(t, doc, doc.Rows(node, env)[0])
	if got["Status"] != "Cancelado" {
		t.Errorf("status = %#v, want Cancelado", got["Status"])
	}
}
