package spec

// CTe reproduces the original CT-e column set. Edit this table to change
// which fields land in the spreadsheet; nothing else needs to change.
var CTe = &Doc{
	Name:    "cte",
	Sheet:   "CT-e",
	Root:    "cteProc",
	Aliases: []string{"CTe"},
	Columns: []Column{
		col("Número CT-e", KindNumber, "CTe/infCte/ide/nCT"),
		col("Data de Emissão", KindDate, "CTe/infCte/ide/dhEmi"),
		col("Razão Social Emitente", KindText, "CTe/infCte/emit/xNome"),
		col("CNPJ Emitente", KindText, "CTe/infCte/emit/CNPJ", "CTe/infCte/emit/CPF"),
		col("Razão Social Remetente", KindText, "CTe/infCte/rem/xNome"),
		col("CNPJ Remetente", KindText, "CTe/infCte/rem/CNPJ", "CTe/infCte/rem/CPF"),
		col("CFOP", KindText, "CTe/infCte/ide/CFOP"),
		joined("Início da Prestação", " - ", "CTe/infCte/ide/xMunIni", "CTe/infCte/ide/UFIni"),
		joined("Término da Prestação", " - ", "CTe/infCte/ide/xMunFim", "CTe/infCte/ide/UFFim"),
		col("Valor Total do Serviço", KindNumber, "CTe/infCte/vPrest/vTPrest"),
		col("Valor a Receber", KindNumber, "CTe/infCte/vPrest/vRec"),
		// "*" matches whichever ICMS variant the document uses (ICMS00, ICMS20,
		// ICMS40, ICMS45, ICMS60, ICMS90, ICMSOutraUF, ICMSSN).
		col("Situação Tributária (CST)", KindText, "CTe/infCte/imp/ICMS/*/CST"),
		col("Base de Cálculo ICMS", KindNumber, "CTe/infCte/imp/ICMS/*/vBC"),
		col("Alíquota ICMS (%)", KindNumber, "CTe/infCte/imp/ICMS/*/pICMS"),
		col("Valor ICMS", KindNumber, "CTe/infCte/imp/ICMS/*/vICMS"),
		col("Chave de Acesso", KindText, "protCTe/infProt/chCTe"),
	},
}
