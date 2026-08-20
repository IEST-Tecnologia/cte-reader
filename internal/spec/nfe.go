package spec

// NFe emits one row per product item: the identification, party, and total
// columns repeat on every row of the same invoice.
//
// Paths are relative to the det node and fall back to the document, which is
// why item groups (prod/..., imposto/...) and document groups (NFe/infNFe/...)
// sit in the same table. Delete Repeat and the item columns for one row per
// invoice instead.
var NFe = &Doc{
	Name:    "nfe",
	Sheet:   "NF-e",
	Root:    "nfeProc",
	Aliases: []string{"NFe"},
	Repeat:  "NFe/infNFe/det",
	Columns: []Column{
		// ---- Identificação da nota ----
		col("Número NF-e", KindNumber, "NFe/infNFe/ide/nNF"),
		col("Série", KindText, "NFe/infNFe/ide/serie"),
		// The protocol carries the key; unprotocolled files still have it in the
		// Id attribute, prefixed with "NFe".
		col("Chave de Acesso", KindText, "protNFe/infProt/chNFe", "NFe/infNFe/@Id").trimPrefix("NFe"),
		col("Natureza da Operação", KindText, "NFe/infNFe/ide/natOp"),
		col("Data de Emissão", KindDate, "NFe/infNFe/ide/dhEmi", "NFe/infNFe/ide/dEmi"),
		col("Data/Hora de Saída ou Entrada", KindDate, "NFe/infNFe/ide/dhSaiEnt", "NFe/infNFe/ide/dSaiEnt"),
		col("Tipo de Operação", KindText, "NFe/infNFe/ide/tpNF").
			decode(map[string]string{"0": "Entrada", "1": "Saída"}),
		// The protocol reports the authorisation; a cancellation lives in a
		// separate event file, matched back to this document by access key.
		col("Status", KindText, "protNFe/infProt/xMotivo").
			cancelledBy("protNFe/infProt/chNFe", "NFe/infNFe/@Id"),

		// ---- Emitente ----
		col("Razão Social Emitente", KindText, "NFe/infNFe/emit/xNome"),
		col("CNPJ Emitente", KindText, "NFe/infNFe/emit/CNPJ", "NFe/infNFe/emit/CPF"),
		col("IE Emitente", KindText, "NFe/infNFe/emit/IE"),
		col("Município Emitente", KindText, "NFe/infNFe/emit/enderEmit/xMun"),
		col("UF Emitente", KindText, "NFe/infNFe/emit/enderEmit/UF"),

		// ---- Destinatário ----
		col("Razão Social Destinatário", KindText, "NFe/infNFe/dest/xNome"),
		col("CNPJ/CPF Destinatário", KindText, "NFe/infNFe/dest/CNPJ", "NFe/infNFe/dest/CPF"),
		col("IE Destinatário", KindText, "NFe/infNFe/dest/IE"),
		col("Município Destinatário", KindText, "NFe/infNFe/dest/enderDest/xMun"),
		col("UF Destinatário", KindText, "NFe/infNFe/dest/enderDest/UF"),

		// ---- Item ----
		col("Item", KindNumber, "@nItem"),
		col("Código do Produto", KindText, "prod/cProd"),
		col("Descrição do Produto", KindText, "prod/xProd"),
		col("NCM", KindText, "prod/NCM"),
		col("CFOP", KindText, "prod/CFOP"),
		col("Unidade Comercial", KindText, "prod/uCom"),
		col("Quantidade", KindNumber, "prod/qCom"),
		col("Valor Unitário", KindNumber, "prod/vUnCom"),
		col("Valor Total do Produto", KindNumber, "prod/vProd"),

		// ---- ICMS do item ----
		// "*" is whichever group the document uses: ICMS00, ICMS10, ICMS20,
		// ICMS40, ICMS51, ICMS60, ICMS70, ICMS90 or the Simples Nacional ones.
		col("Origem da Mercadoria", KindText, "imposto/ICMS/*/orig"),
		col("CST do ICMS", KindText, "imposto/ICMS/*/CST", "imposto/ICMS/*/CSOSN"),
		col("Base de Cálculo do ICMS", KindNumber, "imposto/ICMS/*/vBC"),
		col("Alíquota do ICMS (%)", KindNumber, "imposto/ICMS/*/pICMS"),
		col("Valor do ICMS", KindNumber, "imposto/ICMS/*/vICMS"),

		// ---- ICMS-ST do item ----
		col("Base de Cálculo do ICMS-ST", KindNumber, "imposto/ICMS/*/vBCST"),
		col("Alíquota do ICMS-ST (%)", KindNumber, "imposto/ICMS/*/pICMSST"),
		col("Valor do ICMS-ST", KindNumber, "imposto/ICMS/*/vICMSST"),

		// ---- DIFAL do item ----
		col("Base de Cálculo do DIFAL", KindNumber, "imposto/ICMSUFDest/vBCUFDest"),
		col("Valor do ICMS UF Destino", KindNumber, "imposto/ICMSUFDest/vICMSUFDest"),

		// ---- IPI do item ----
		col("CST do IPI", KindText, "imposto/IPI/*/CST"),
		col("Enquadramento Legal do IPI", KindText, "imposto/IPI/cEnq"),
		col("Base de Cálculo do IPI", KindNumber, "imposto/IPI/IPITrib/vBC"),
		col("Alíquota do IPI (%)", KindNumber, "imposto/IPI/IPITrib/pIPI"),
		col("Valor do IPI", KindNumber, "imposto/IPI/IPITrib/vIPI"),

		// ---- PIS do item ----
		col("CST do PIS", KindText, "imposto/PIS/*/CST"),
		col("Base de Cálculo do PIS", KindNumber, "imposto/PIS/*/vBC"),
		col("Alíquota do PIS (%)", KindNumber, "imposto/PIS/*/pPIS"),
		col("Valor do PIS", KindNumber, "imposto/PIS/*/vPIS"),

		// ---- COFINS do item ----
		col("CST da COFINS", KindText, "imposto/COFINS/*/CST"),
		col("Base de Cálculo da COFINS", KindNumber, "imposto/COFINS/*/vBC"),
		col("Alíquota da COFINS (%)", KindNumber, "imposto/COFINS/*/pCOFINS"),
		col("Valor da COFINS", KindNumber, "imposto/COFINS/*/vCOFINS"),

		// ---- IBS e CBS do item ----
		// "**" walks whatever g* subgroups the layout version nests these in.
		col("CST do IBS/CBS", KindText, "imposto/IBSCBS/CST"),
		col("Classificação Tributária IBS/CBS", KindText, "imposto/IBSCBS/cClassTrib"),
		col("Base de Cálculo IBS/CBS", KindNumber, "imposto/IBSCBS/**/vBC"),
		col("Alíquota IBS Estadual (%)", KindNumber, "imposto/IBSCBS/**/pIBSUF"),
		col("Valor IBS Estadual", KindNumber, "imposto/IBSCBS/**/vIBSUF"),
		col("Alíquota IBS Municipal (%)", KindNumber, "imposto/IBSCBS/**/pIBSMun"),
		col("Valor IBS Municipal", KindNumber, "imposto/IBSCBS/**/vIBSMun"),
		col("Valor Total do IBS", KindNumber, "imposto/IBSCBS/**/vIBS"),
		col("Alíquota CBS (%)", KindNumber, "imposto/IBSCBS/**/pCBS"),
		col("Valor da CBS", KindNumber, "imposto/IBSCBS/**/vCBS"),

		// ---- Totais da nota ----
		col("Base de Cálculo do ICMS (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vBC"),
		col("Valor do ICMS (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vICMS"),
		col("Base de Cálculo do ICMS-ST (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vBCST"),
		col("Valor do ICMS-ST (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vST"),
		col("Valor do DIFAL (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vICMSUFDest"),
		col("Valor do IPI (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vIPI"),
		col("Valor do PIS (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vPIS"),
		col("Valor da COFINS (Total)", KindNumber, "NFe/infNFe/total/ICMSTot/vCOFINS"),
		col("Base de Cálculo IBS/CBS (Total)", KindNumber, "NFe/infNFe/total/IBSCBSTot/**/vBCIBSCBS"),
		col("Valor IBS Estadual (Total)", KindNumber, "NFe/infNFe/total/IBSCBSTot/**/vIBSUF"),
		col("Valor IBS Municipal (Total)", KindNumber, "NFe/infNFe/total/IBSCBSTot/**/vIBSMun"),
		col("Valor Total do IBS (Total)", KindNumber, "NFe/infNFe/total/IBSCBSTot/**/vIBS"),
		col("Valor da CBS (Total)", KindNumber, "NFe/infNFe/total/IBSCBSTot/**/vCBS"),
		col("Total dos Produtos", KindNumber, "NFe/infNFe/total/ICMSTot/vProd"),
		col("Total da NF-e", KindNumber, "NFe/infNFe/total/ICMSTot/vNF"),
	},
}
