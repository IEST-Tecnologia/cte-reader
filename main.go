package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ---- XML structs ----

type CteProc struct {
	XMLName xml.Name `xml:"cteProc"`
	CTe     CTe      `xml:"CTe"`
	ProtCTe ProtCTe  `xml:"protCTe"`
}

type CTe struct {
	InfCte InfCte `xml:"infCte"`
}

type InfCte struct {
	Ide    Ide    `xml:"ide"`
	Emit   Emit   `xml:"emit"`
	Rem    Rem    `xml:"rem"`
	VPrest VPrest `xml:"vPrest"`
	Imp    Imp    `xml:"imp"`
}

type Ide struct {
	NCT     string `xml:"nCT"`
	DhEmi   string `xml:"dhEmi"`
	CFOP    string `xml:"CFOP"`
	XMunIni string `xml:"xMunIni"`
	UFIni   string `xml:"UFIni"`
	XMunFim string `xml:"xMunFim"`
	UFFim   string `xml:"UFFim"`
}

type Emit struct {
	CNPJ  string `xml:"CNPJ"`
	XNome string `xml:"xNome"`
}

type Rem struct {
	CNPJ  string `xml:"CNPJ"`
	XNome string `xml:"xNome"`
}

type VPrest struct {
	VTPrest string `xml:"vTPrest"`
	VRec    string `xml:"vRec"`
}

type Imp struct {
	ICMS ICMS `xml:"ICMS"`
}

type ICMS struct {
	ICMS00 *ICMSDetail `xml:"ICMS00"`
	ICMS20 *ICMSDetail `xml:"ICMS20"`
	ICMS40 *ICMSDetail `xml:"ICMS40"`
	ICMS45 *ICMSDetail `xml:"ICMS45"`
	ICMS60 *ICMSDetail `xml:"ICMS60"`
	ICMS90 *ICMSDetail `xml:"ICMS90"`
}

// Active returns whichever ICMS variant is present.
func (i ICMS) Active() *ICMSDetail {
	for _, d := range []*ICMSDetail{i.ICMS00, i.ICMS20, i.ICMS40, i.ICMS45, i.ICMS60, i.ICMS90} {
		if d != nil {
			return d
		}
	}
	return nil
}

type ICMSDetail struct {
	CST   string `xml:"CST"`
	VBC   string `xml:"vBC"`
	PICMS string `xml:"pICMS"`
	VICMS string `xml:"vICMS"`
}

type ProtCTe struct {
	InfProt InfProt `xml:"infProt"`
}

type InfProt struct {
	ChCTe string `xml:"chCTe"`
}

// ---- Helpers ----

var (
	reXmlns  = regexp.MustCompile(`\s+xmlns(?::\w+)?="[^"]*"`)
	rePrefix = regexp.MustCompile(`(</?)\w+:`)
)

// stripNamespaces removes xmlns declarations and tag prefixes so Go can match by local name.
func stripNamespaces(data []byte) []byte {
	data = reXmlns.ReplaceAll(data, nil)
	data = rePrefix.ReplaceAll(data, []byte("$1"))
	return data
}

func parseCte(data []byte) (*CteProc, error) {
	var cte CteProc
	if err := xml.Unmarshal(stripNamespaces(data), &cte); err != nil {
		return nil, err
	}
	return &cte, nil
}

// ---- Workbook helpers ----

const sheet = "CTe"

// chunkSizeStr is overridden at build time via -ldflags "-X main.chunkSizeStr=N".
var chunkSizeStr = "500000"

var chunkSize int

func init() {
	n, err := strconv.Atoi(chunkSizeStr)
	if err != nil || n <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid chunk size %q, using 500000\n", chunkSizeStr)
		n = 500_000
	}
	chunkSize = n
}

var headers = []string{
	"Número CT-e",
	"Data de Emissão",
	"Razão Social Emitente",
	"CNPJ Emitente",
	"Razão Social Remetente",
	"CNPJ Remetente",
	"CFOP",
	"Início da Prestação",
	"Término da Prestação",
	"Valor Total do Serviço",
	"Valor a Receber",
	"Situação Tributária (CST)",
	"Base de Cálculo ICMS",
	"Alíquota ICMS (%)",
	"Valor ICMS",
	"Chave de Acesso",
}

func newWorkbook() *excelize.File {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheet)
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle(sheet, 1, 1, style)
	return f
}

func saveWorkbook(f *excelize.File, path string) error {
	cols, _ := f.GetCols(sheet)
	for i, col := range cols {
		maxLen := 0
		for _, c := range col {
			if len(c) > maxLen {
				maxLen = len(c)
			}
		}
		name, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, name, name, float64(maxLen)+2)
	}
	return f.SaveAs(path)
}

// ---- Main ----

func main() {
	zipPath := getFilename()
	if zipPath == "" {
		os.Exit(1)
	}
	base := strings.TrimSuffix(zipPath, filepath.Ext(zipPath))

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening zip: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	f := newWorkbook()
	row := 2
	chunkNum := 1
	totalWritten := 0
	skipped := 0
	var outPaths []string

	for _, file := range r.File {
		if !strings.HasSuffix(strings.ToLower(file.Name), ".xml") {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", file.Name, err)
			skipped++
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", file.Name, err)
			skipped++
			continue
		}

		cte, err := parseCte(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", file.Name, err)
			skipped++
			continue
		}

		inf := cte.CTe.InfCte
		icms := inf.Imp.ICMS.Active()

		var cst, vbc, picms, vicms string
		if icms != nil {
			cst = icms.CST
			vbc = icms.VBC
			picms = icms.PICMS
			vicms = icms.VICMS
		}

		inicio := inf.Ide.XMunIni + " - " + inf.Ide.UFIni
		fim := inf.Ide.XMunFim + " - " + inf.Ide.UFFim

		values := []any{
			inf.Ide.NCT,
			inf.Ide.DhEmi,
			inf.Emit.XNome,
			inf.Emit.CNPJ,
			inf.Rem.XNome,
			inf.Rem.CNPJ,
			inf.Ide.CFOP,
			inicio,
			fim,
			inf.VPrest.VTPrest,
			inf.VPrest.VRec,
			cst,
			vbc,
			picms,
			vicms,
			cte.ProtCTe.InfProt.ChCTe,
		}

		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, v)
		}
		row++
		totalWritten++

		if totalWritten%chunkSize == 0 {
			path := fmt.Sprintf("%s_parte%d.xlsx", base, chunkNum)
			if err := saveWorkbook(f, path); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving %s: %v\n", path, err)
				os.Exit(1)
			}
			outPaths = append(outPaths, path)
			f = newWorkbook()
			row = 2
			chunkNum++
		}
	}

	// Save the last (or only) chunk if it has data rows.
	if row > 2 {
		path := fmt.Sprintf("%s_parte%d.xlsx", base, chunkNum)
		if err := saveWorkbook(f, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving %s: %v\n", path, err)
			os.Exit(1)
		}
		outPaths = append(outPaths, path)
	}

	// Single chunk: drop the _part1 suffix for a cleaner output name.
	if len(outPaths) == 1 {
		clean := base + ".xlsx"
		if err := os.Rename(outPaths[0], clean); err != nil {
			fmt.Fprintf(os.Stderr, "Error renaming output: %v\n", err)
			os.Exit(1)
		}
		outPaths[0] = clean
	}

	showResult(totalWritten, skipped, outPaths)
}
