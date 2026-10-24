package cryptodatadownload

// ReturnFormat selects the wire format on the 43 legacy operations that
// declare a `return` query parameter in the OpenAPI schema. JSON is the API
// default and is what every typed method decodes; CSV and XLSX are only
// reachable through the explicit *Raw variants, which return the raw body.
//
// No operation accepts a `format` parameter — that name appears only in
// marketing material. Do not confuse the two.
type ReturnFormat string

const (
	// FormatJSON returns the standard JSON envelope.
	FormatJSON ReturnFormat = "JSON"
	// FormatCSV returns a CSV document.
	FormatCSV ReturnFormat = "CSV"
	// FormatXLSX returns an XLSX workbook.
	FormatXLSX ReturnFormat = "XLSX"
)
