# Fiscal Reader

Reads fiscal document XMLs from a ZIP archive and exports the relevant fields to
Excel spreadsheets (`.xlsx`). CT-e and NF-e are supported out of the box, and
both the document types and the exported columns are declarative tables.

## Usage

On Windows, double-click the `.exe`. A file picker will open — select the ZIP
containing the XMLs. The output is saved in the same folder, named after the ZIP.

Each document type gets its own workbook, since the columns differ:

| Contents of the ZIP        | Output                                     |
| -------------------------- | ------------------------------------------ |
| One type, fits in one file | `docs.xlsx`                                |
| One type, several chunks   | `docs_parte1.xlsx`, `docs_parte2.xlsx`, …  |
| Several types              | `docs_cte.xlsx`, `docs_nfe.xlsx`           |

Files that cannot be read or parsed are skipped, and files whose root element no
spec claims are ignored; both are counted in the completion dialog, along with
the event documents that were read.

## Cancellations

Event files (`procEventoNFe`, `procEventoCTe`) are read in a first pass over the
archive, before any document is converted, since a cancellation may sit anywhere
in the ZIP — including after the document it cancels. A document is marked
`Cancelado` when the archive holds an event that

- carries `tpEvento` 110111 or 110112 (cancellation, or cancellation by
  substitution), **and**
- was accepted by the authority — `cStat` 135, 136 or 155. A rejected event, a
  letter of correction, or an acknowledgement never cancels anything.

The event is matched to its document by access key (`chNFe` or `chCTe`, falling
back to the key embedded in the event's `Id` attribute). An event whose reply
half is missing has no `cStat` to check and is trusted.

Only NF-e has a `Status` column today. Giving CT-e one is a single line in its
table, since the event pass already collects `chCTe` keys:

```go
col("Status", KindText, "protCTe/infProt/xMotivo").
	cancelledBy("protCTe/infProt/chCTe"),
```

## Supported documents

| Type   | Root elements                     | Rows                 | Spec                                             |
| ------ | --------------------------------- | -------------------- | ------------------------------------------------ |
| CT-e   | `cteProc` or `CTe`                | one per document     | [internal/spec/cte.go](internal/spec/cte.go)     |
| NF-e   | `nfeProc` or `NFe`                | one per product item | [internal/spec/nfe.go](internal/spec/nfe.go)     |
| Events | `procEventoNFe`, `procEventoCTe`  | none, they set status | [internal/events](internal/events)              |

The NF-e sheet is item-level: identification, parties and totals repeat on every
row of the same invoice, and each row adds that item's product data and its
ICMS, ICMS-ST, DIFAL, IPI, PIS, COFINS and IBS/CBS figures. Groups an item does
not carry leave their columns empty.

Documents that arrive without their processing envelope (a bare `<CTe>` or
`<NFe>` root) are read too; an NF-e then takes its access key from `infNFe/@Id`,
while a CT-e has none.

`Status` is the authorisation message from the document's own protocol
(`xMotivo`), replaced by `Cancelado` when the archive also contains a
cancellation event for that access key.

## Layout

| Package                                    | Responsibility                                             |
| ------------------------------------------ | ---------------------------------------------------------- |
| `main`                                     | entry point, file picker, result dialog, build-time options |
| [internal/convert](internal/convert)       | walks the ZIP and routes each document to its writer        |
| [internal/spec](internal/spec)             | what to extract: document types and their column tables     |
| [internal/events](internal/events)         | cancellation events, matched to documents by access key     |
| [internal/xmltree](internal/xmltree)       | namespace-agnostic XML tree addressed by path               |
| [internal/xlsx](internal/xlsx)             | spreadsheet output and chunking                             |
| [internal/fixtures](internal/fixtures)     | sample documents shared by tests                            |

The file picker and the completion dialog live in `dialog_prod.go` (behind the
`prod` build tag) and `dialog_dev.go`, so a development build runs headless
against a fixed path.

## Changing the exported columns

Every column is one entry in a spec's `Columns` table. Paths address elements by
local name, so namespaces and tag prefixes are irrelevant.

```go
col("Valor Total do Serviço", KindNumber, "CTe/infCte/vPrest/vTPrest")
```

- **`col(header, kind, paths...)`** — the first non-empty path wins, which is how
  `CNPJ` falls back to `CPF`.
- **`joined(header, sep, paths...)`** — concatenates every non-empty path, e.g.
  `"São Paulo" + " - " + "SP"`.
- **`Kind`** — `KindText` keeps the raw string (right for codes: CNPJ, CFOP, CST,
  access keys, anything with leading zeros), `KindNumber` writes a real number so
  the column can be summed, `KindDate` writes a real date so it can be sorted and
  filtered. A value that does not parse falls back to its raw text.
- **`*`** matches any single element: `imposto/ICMS/*/vICMS` reads whichever ICMS
  group the document uses without naming ICMS00, ICMS10, ICMS20 and friends.
- **`**`** matches any number of levels: `imposto/IBSCBS/**/vCBS` finds the value
  through whatever `g*` subgroups a layout version nests it in. The shallowest
  match wins.
- **`@name`** as the last segment reads an attribute instead of element text.
- **`.trimPrefix("NFe")`** and **`.decode(map[string]string{...})`** post-process
  the raw value — stripping the prefix from an access key read out of
  `infNFe/@Id`, or turning `tpNF` 0/1 into Entrada/Saída.
- **`.cancelledBy(keyPaths...)`** overrides the value with `Cancelado` when the
  archive holds a cancellation event for the key at one of those paths. It is the
  one column source that looks beyond the document being converted.

Column order in the table is column order in the spreadsheet. A path that
matches nothing yields an empty cell rather than an error.

## Adding a document type

1. Copy [internal/spec/nfe.go](internal/spec/nfe.go) and edit `Name`, `Sheet`, `Root`, `Aliases`
   and `Columns`.
2. Register it in the `All` list in [internal/spec/spec.go](internal/spec/spec.go).

Set `Repeat` to a repeating node's path (NF-e uses `NFe/infNFe/det`) to emit one
row per match; item paths are then relative to that node, while document-level
paths keep resolving from the root. Leave `Repeat` empty for one row per file.

To discover paths, dump the element tree of a sample XML — `data/layout.txt` is
such a dump for CT-e.

## Building

Requires Go 1.21+.

```bash
# Windows release (no terminal window)
make build

# Rows per output file (default 500000)
make build CHUNK_SIZE=100000

# Development (hardcoded test path, no dialog, prints to stdout)
go build .

make test
```

## Releasing

Push a version tag to trigger the GitHub Actions workflow, which builds and
attaches `fiscal-reader.exe` to the release automatically.

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Commits

This repository follows the [conventional commits specification](https://www.conventionalcommits.org/en/v1.0.0/)
