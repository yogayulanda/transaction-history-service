package domain

const (
	CategoryTransfer = "transfer"
	CategoryPPOB     = "ppob"
	CategoryCICO     = "cico"

	ProductTransferBiFast   = "transfer_bifast"
	ProductTransferRTOL     = "transfer_rtol"
	ProductTransferInternal = "transfer_internal"
	ProductPulsa            = "pulsa"
	ProductDataPackage      = "data_package"
	ProductPLNPrepaid       = "pln_prepaid"
	ProductPLNPostpaid      = "pln_postpaid"
	ProductBPJSKesehatan    = "bpjs_kesehatan"
	ProductCashIn           = "cash_in"
	ProductCashOut          = "cash_out"
)

type Catalog interface {
	ProductCategory(productType string) (string, bool)
	ValidProvider(productType, provider string) bool
	ProviderName(provider string) string
	ValidPartner(partner string) bool
	AllowedSource(source string) bool
	ValidApp(app string) bool
}
