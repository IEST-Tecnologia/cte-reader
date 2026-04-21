# CTE Reader

Reads CT-e XML files from a ZIP archive and exports the relevant fields to an Excel spreadsheet (`.xlsx`).

## Usage

On Windows, double-click the `.exe`. A file picker will open — select the ZIP containing the CT-e XMLs. The output `.xlsx` is saved in the same folder with the same name as the ZIP.

## Output columns

| Column                    | Description                |
| ------------------------- | -------------------------- |
| Razão Social Emitente     | Emitter company name       |
| CNPJ Emitente             | Emitter tax ID             |
| Razão Social Remetente    | Sender company name        |
| CNPJ Remetente            | Sender tax ID              |
| CFOP                      | Fiscal operation code      |
| Início da Prestação       | Origin city and state      |
| Término da Prestação      | Destination city and state |
| Valor Total do Serviço    | Total service value        |
| Valor a Receber           | Amount receivable          |
| Situação Tributária (CST) | ICMS tax situation code    |
| Base de Cálculo ICMS      | ICMS calculation base      |
| Alíquota ICMS (%)         | ICMS rate                  |
| Valor ICMS                | ICMS value                 |
| Chave de Acesso           | CT-e access key            |

## Building

Requires Go 1.21+.

```bash
# Windows release (no terminal window)
make build

# Development (hardcoded test path, no dialog)
go build .
```

## Releasing

Push a version tag to trigger the GitHub Actions workflow, which builds and attaches `cte-reader.exe` to the release automatically.

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Commits

This repository follows the [conventional commits specification](https://www.conventionalcommits.org/en/v1.0.0/)
