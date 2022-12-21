package invoices

type InvoiceVersion uint64

const (
	INVOICE_VERSION_0 InvoiceVersion = iota
)

type InvoiceItemVersion uint64

const (
	INVOICE_ITEM_NEW InvoiceItemVersion = iota
	INVOICE_ITEM_ID
)
