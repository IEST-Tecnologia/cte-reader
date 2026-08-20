// Package fixtures holds sample documents shared by the tests of several
// packages. It is test-only: nothing in the command imports it.
package fixtures

// CTe is a processed CT-e with an ICMS20 group, covering every column of the
// CT-e spec.
const CTe = `<?xml version="1.0" encoding="utf-8"?>
<cteProc versao="4.00" xmlns="http://www.portalfiscal.inf.br/cte">
  <CTe xmlns="http://www.portalfiscal.inf.br/cte">
    <infCte Id="CTe35260355442952000157570010152386931847651990" versao="4.00">
      <ide>
        <CFOP>6352</CFOP>
        <nCT>1523869</nCT>
        <dhEmi>2026-03-11T14:32:05-03:00</dhEmi>
        <xMunIni>São Paulo</xMunIni>
        <UFIni>SP</UFIni>
        <xMunFim>Curitiba</xMunFim>
        <UFFim>PR</UFFim>
      </ide>
      <emit><CNPJ>55442952000157</CNPJ><xNome>Transportadora Exemplo LTDA</xNome></emit>
      <rem><CNPJ>11222333000181</CNPJ><xNome>Remetente Exemplo SA</xNome></rem>
      <vPrest><vTPrest>1234.56</vTPrest><vRec>1200.00</vRec></vPrest>
      <imp><ICMS><ICMS20><CST>20</CST><vBC>1000.00</vBC><pICMS>12.00</pICMS><vICMS>120.00</vICMS></ICMS20></ICMS></imp>
    </infCte>
  </CTe>
  <protCTe versao="4.00">
    <infProt><chCTe>35260355442952000157570010152386931847651990</chCTe></infProt>
  </protCTe>
</cteProc>`

// NFe is a processed NF-e with two items: the first taxed under ICMS00 with
// IPI, PIS, COFINS and IBS/CBS groups, the second under ICMS10 with ICMS-ST and
// DIFAL but no IPI or IBS/CBS, so tests cover both present and absent groups.
const NFe = `<?xml version="1.0" encoding="utf-8"?>
<nfeProc versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">
  <NFe>
    <infNFe Id="NFe35260311222333000181550010000012341000012348" versao="4.00">
      <ide>
        <nNF>1234</nNF><serie>1</serie><natOp>Venda de mercadoria</natOp>
        <dhEmi>2026-03-11T09:05:00-03:00</dhEmi>
        <dhSaiEnt>2026-03-11T18:30:00-03:00</dhSaiEnt>
        <tpNF>1</tpNF>
      </ide>
      <emit>
        <CNPJ>11222333000181</CNPJ><xNome>Indústria Exemplo SA</xNome><IE>123456789012</IE>
        <enderEmit><xMun>São Paulo</xMun><UF>SP</UF></enderEmit>
      </emit>
      <dest>
        <CPF>12345678909</CPF><xNome>Cliente Pessoa Física</xNome><IE>ISENTO</IE>
        <enderDest><xMun>Nova Iguaçu</xMun><UF>RJ</UF></enderDest>
      </dest>
      <det nItem="1">
        <prod><cProd>A-1</cProd><xProd>Parafuso</xProd><NCM>73181500</NCM><CFOP>5102</CFOP>
          <uCom>UN</uCom><qCom>10.0000</qCom><vUnCom>2.5000</vUnCom><vProd>25.00</vProd></prod>
        <imposto>
          <ICMS><ICMS00><orig>0</orig><CST>00</CST><vBC>25.00</vBC><pICMS>18.00</pICMS><vICMS>4.50</vICMS></ICMS00></ICMS>
          <IPI><cEnq>999</cEnq>
            <IPITrib><CST>50</CST><vBC>25.00</vBC><pIPI>5.00</pIPI><vIPI>1.25</vIPI></IPITrib></IPI>
          <PIS><PISAliq><CST>01</CST><vBC>25.00</vBC><pPIS>1.65</pPIS><vPIS>0.41</vPIS></PISAliq></PIS>
          <COFINS><COFINSAliq><CST>01</CST><vBC>25.00</vBC><pCOFINS>7.60</pCOFINS><vCOFINS>1.90</vCOFINS></COFINSAliq></COFINS>
          <IBSCBS><CST>000</CST><cClassTrib>000001</cClassTrib>
            <gIBSCBS><vBC>25.00</vBC>
              <gIBSUF><pIBSUF>0.1000</pIBSUF><vIBSUF>0.03</vIBSUF></gIBSUF>
              <gIBSMun><pIBSMun>0.0500</pIBSMun><vIBSMun>0.01</vIBSMun></gIBSMun>
              <vIBS>0.04</vIBS>
              <gCBS><pCBS>0.9000</pCBS><vCBS>0.23</vCBS></gCBS>
            </gIBSCBS>
          </IBSCBS>
        </imposto>
      </det>
      <det nItem="2">
        <prod><cProd>B-2</cProd><xProd>Porca</xProd><NCM>73181600</NCM><CFOP>6102</CFOP>
          <uCom>CX</uCom><qCom>4.0000</qCom><vUnCom>1.0000</vUnCom><vProd>4.00</vProd></prod>
        <imposto>
          <ICMS><ICMS10><orig>0</orig><CST>10</CST><vBC>4.00</vBC><pICMS>12.00</pICMS><vICMS>0.48</vICMS>
            <vBCST>6.00</vBCST><pICMSST>18.00</pICMSST><vICMSST>0.60</vICMSST></ICMS10></ICMS>
          <ICMSUFDest><vBCUFDest>4.00</vBCUFDest><vICMSUFDest>0.24</vICMSUFDest></ICMSUFDest>
          <PIS><PISNT><CST>07</CST></PISNT></PIS>
          <COFINS><COFINSNT><CST>07</CST></COFINSNT></COFINS>
        </imposto>
      </det>
      <total>
        <ICMSTot><vBC>29.00</vBC><vICMS>4.98</vICMS><vBCST>6.00</vBCST><vST>0.60</vST>
          <vICMSUFDest>0.24</vICMSUFDest><vProd>29.00</vProd><vIPI>1.25</vIPI>
          <vPIS>0.41</vPIS><vCOFINS>1.90</vCOFINS><vNF>29.00</vNF></ICMSTot>
        <IBSCBSTot><vBCIBSCBS>29.00</vBCIBSCBS>
          <gIBS>
            <gIBSUF><vIBSUF>0.03</vIBSUF></gIBSUF>
            <gIBSMun><vIBSMun>0.01</vIBSMun></gIBSMun>
            <vIBS>0.04</vIBS>
          </gIBS>
          <gCBS><vCBS>0.23</vCBS></gCBS>
        </IBSCBSTot>
      </total>
    </infNFe>
  </NFe>
  <protNFe>
    <infProt><chNFe>35260311222333000181550010000012341000012348</chNFe>
      <cStat>100</cStat><xMotivo>Autorizado o uso da NF-e</xMotivo></infProt>
  </protNFe>
</nfeProc>`

// CancelEventNFe cancels the NF-e above: tpEvento 110111, accepted by the
// authority with cStat 135, matched back to the document by chNFe.
const CancelEventNFe = `<?xml version="1.0" encoding="UTF-8"?>
<procEventoNFe versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">
  <evento versao="1.00">
    <infEvento Id="ID11011135260311222333000181550010000012341000012348 01">
      <cOrgao>35</cOrgao><tpAmb>1</tpAmb><CNPJ>11222333000181</CNPJ>
      <chNFe>35260311222333000181550010000012341000012348</chNFe>
      <dhEvento>2026-03-12T10:00:00-03:00</dhEvento>
      <tpEvento>110111</tpEvento><nSeqEvento>1</nSeqEvento>
      <detEvento versao="1.00">
        <descEvento>Cancelamento</descEvento><nProt>135260000000001</nProt>
        <xJust>Erro de digitação nos valores</xJust>
      </detEvento>
    </infEvento>
  </evento>
  <retEvento versao="1.00">
    <infEvento>
      <tpAmb>1</tpAmb><cOrgao>35</cOrgao><cStat>135</cStat>
      <xMotivo>Evento registrado e vinculado a NF-e</xMotivo>
      <chNFe>35260311222333000181550010000012341000012348</chNFe>
      <tpEvento>110111</tpEvento><xEvento>Cancelamento registrado</xEvento>
      <nSeqEvento>1</nSeqEvento><dhRegEvento>2026-03-12T10:00:01-03:00</dhRegEvento>
      <nProt>135260000000002</nProt>
    </infEvento>
  </retEvento>
</procEventoNFe>`

// CorrectionEventNFe is a letter of correction (tpEvento 110110) for the same
// NF-e: an event, but not a cancellation.
const CorrectionEventNFe = `<?xml version="1.0" encoding="UTF-8"?>
<procEventoNFe versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">
  <evento versao="1.00">
    <infEvento Id="ID11011035260311222333000181550010000012341000012348 01">
      <chNFe>35260311222333000181550010000012341000012348</chNFe>
      <tpEvento>110110</tpEvento><nSeqEvento>1</nSeqEvento>
      <detEvento versao="1.00"><descEvento>Carta de Correcao</descEvento></detEvento>
    </infEvento>
  </evento>
  <retEvento versao="1.00">
    <infEvento><cStat>135</cStat><xMotivo>Evento registrado e vinculado a NF-e</xMotivo>
      <chNFe>35260311222333000181550010000012341000012348</chNFe>
      <tpEvento>110110</tpEvento></infEvento>
  </retEvento>
</procEventoNFe>`

// RejectedCancelEventNFe is a cancellation the authority refused (cStat 573),
// so the document it names is still valid.
const RejectedCancelEventNFe = `<?xml version="1.0" encoding="UTF-8"?>
<procEventoNFe versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">
  <evento versao="1.00">
    <infEvento Id="ID11011135260311222333000181550010000012341000012348 02">
      <chNFe>35260311222333000181550010000012341000012348</chNFe>
      <tpEvento>110111</tpEvento><nSeqEvento>2</nSeqEvento>
      <detEvento versao="1.00"><descEvento>Cancelamento</descEvento></detEvento>
    </infEvento>
  </evento>
  <retEvento versao="1.00">
    <infEvento><cStat>573</cStat><xMotivo>Rejeicao: Duplicidade de evento</xMotivo>
      <chNFe>35260311222333000181550010000012341000012348</chNFe>
      <tpEvento>110111</tpEvento></infEvento>
  </retEvento>
</procEventoNFe>`
