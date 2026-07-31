// Package dataio converts a workspace dataset to and from the spreadsheet
// formats a user can open: one XLSX workbook with a sheet per table, or a ZIP
// archive holding the same tables as CSV files.
//
// Both formats share the intermediate Table representation below, so the sheet
// layout is defined exactly once (in tables.go) and each codec only deals with
// its own container format.
package dataio

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Table is one sheet: a name, a header row, and the cells below it.
type Table struct {
	Name    string
	Columns []string
	Rows    [][]string
}

// Format identifies a container the dataset can be written to.
type Format string

const (
	FormatXLSX Format = "xlsx"
	FormatCSV  Format = "csv"
)

// ParseFormat validates a caller-supplied format name.
func ParseFormat(value string) (Format, error) {
	switch Format(strings.ToLower(strings.TrimSpace(value))) {
	case FormatXLSX, "":
		return FormatXLSX, nil
	case FormatCSV:
		return FormatCSV, nil
	default:
		return "", fmt.Errorf("format must be xlsx or csv")
	}
}

// Extension is the file extension an export of this format is served with.
func (f Format) Extension() string {
	if f == FormatCSV {
		return "zip"
	}
	return "xlsx"
}

// ContentType is the MIME type an export of this format is served with.
func (f Format) ContentType() string {
	if f == FormatCSV {
		return "application/zip"
	}
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

// multiValueSeparator joins list-valued cells such as a school's classrooms so
// that one row stays one row in the spreadsheet.
const multiValueSeparator = " | "

func joinValues(values []string) string {
	return strings.Join(values, multiValueSeparator)
}

func splitValues(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, "|")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func boolCell(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func numberCell(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func floatCell(value *float64) string {
	if value == nil {
		return ""
	}
	return numberCell(*value)
}

func timeCell(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func timePointerCell(value *time.Time) string {
	if value == nil {
		return ""
	}
	return timeCell(*value)
}

// rowReader looks cells up by column name so that a hand-edited file may
// reorder columns, drop optional ones, or change their letter case.
type rowReader struct {
	table  string
	line   int
	index  map[string]int
	values []string
	errs   *[]string
}

func newRowReader(table string, line int, index map[string]int, values []string, errs *[]string) rowReader {
	return rowReader{table: table, line: line, index: index, values: values, errs: errs}
}

func (r rowReader) fail(format string, args ...any) {
	*r.errs = append(*r.errs, fmt.Sprintf("%s row %d: %s", r.table, r.line, fmt.Sprintf(format, args...)))
}

func (r rowReader) text(column string) string {
	position, ok := r.index[strings.ToLower(column)]
	if !ok || position >= len(r.values) {
		return ""
	}
	return strings.TrimSpace(r.values[position])
}

func (r rowReader) list(column string) []string {
	return splitValues(r.text(column))
}

func (r rowReader) boolean(column string) bool {
	switch strings.ToLower(r.text(column)) {
	case "true", "yes", "y", "1":
		return true
	default:
		return false
	}
}

func (r rowReader) number(column string) float64 {
	value := r.text(column)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		r.fail("%s must be a number", column)
		return 0
	}
	return parsed
}

func (r rowReader) integer(column string) int {
	return int(r.number(column))
}

func (r rowReader) optionalNumber(column string) *float64 {
	value := r.text(column)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		r.fail("%s must be a number", column)
		return nil
	}
	return &parsed
}

// timestamp accepts the RFC 3339 stamps written by the exporter as well as the
// plainer forms Excel produces when a user retypes a cell.
var timestampLayouts = []string{
	time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04", time.DateOnly,
}

func (r rowReader) timestamp(column string) *time.Time {
	value := r.text(column)
	if value == "" {
		return nil
	}
	for _, layout := range timestampLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			parsed = parsed.UTC()
			return &parsed
		}
	}
	r.fail("%s must be a date or timestamp", column)
	return nil
}

// forEachRow walks a table's data rows, skipping ones that are entirely blank.
func forEachRow(table Table, errs *[]string, visit func(rowReader)) {
	index := make(map[string]int, len(table.Columns))
	for position, column := range table.Columns {
		index[strings.ToLower(strings.TrimSpace(column))] = position
	}
	for offset, values := range table.Rows {
		if blankRow(values) {
			continue
		}
		// +2 so the reported line matches the spreadsheet, header included.
		visit(newRowReader(table.Name, offset+2, index, values, errs))
	}
}

func blankRow(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
