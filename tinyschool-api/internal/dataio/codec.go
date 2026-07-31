package dataio

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/xuri/excelize/v2"

	"tinyschool-api/internal/model"
)

// MaxUploadBytes caps an imported file. A workspace of a few thousand students
// stays far below this even as an uncompressed workbook.
const MaxUploadBytes = 25 << 20

// Encode writes the dataset as an XLSX workbook or as a ZIP of CSV files.
func Encode(data model.Dataset, format Format) ([]byte, error) {
	if format == FormatCSV {
		return encodeCSV(tables(data))
	}
	return encodeXLSX(tables(data))
}

// Decode reads a dataset back from an uploaded file. The filename picks the
// container: .xlsx workbook, .zip of CSVs, or a single .csv holding one sheet.
func Decode(filename string, content []byte) (model.Dataset, error) {
	var (
		parsed []Table
		err    error
	)
	switch strings.ToLower(path.Ext(strings.TrimSpace(filename))) {
	case ".xlsx", ".xlsm":
		parsed, err = decodeXLSX(content)
	case ".zip":
		parsed, err = decodeZIP(content)
	case ".csv", ".txt", ".tsv":
		parsed, err = decodeSingleCSV(filename, content)
	default:
		return model.Dataset{}, fmt.Errorf("unsupported file type; upload the .xlsx workbook, the .zip of CSV files, or a single .csv sheet")
	}
	if err != nil {
		return model.Dataset{}, err
	}
	return dataset(parsed)
}

func encodeXLSX(sheets []Table) ([]byte, error) {
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	header, err := file.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, fmt.Errorf("create header style: %w", err)
	}
	for index, sheet := range sheets {
		if index == 0 {
			if err := file.SetSheetName(file.GetSheetName(0), sheet.Name); err != nil {
				return nil, fmt.Errorf("name sheet %s: %w", sheet.Name, err)
			}
		} else if _, err := file.NewSheet(sheet.Name); err != nil {
			return nil, fmt.Errorf("create sheet %s: %w", sheet.Name, err)
		}
		writer, err := file.NewStreamWriter(sheet.Name)
		if err != nil {
			return nil, fmt.Errorf("write sheet %s: %w", sheet.Name, err)
		}
		if err := writer.SetRow("A1", cells(sheet.Columns), excelize.RowOpts{StyleID: header}); err != nil {
			return nil, fmt.Errorf("write header of %s: %w", sheet.Name, err)
		}
		for offset, row := range sheet.Rows {
			axis, err := excelize.CoordinatesToCellName(1, offset+2)
			if err != nil {
				return nil, fmt.Errorf("write row of %s: %w", sheet.Name, err)
			}
			if err := writer.SetRow(axis, cells(row)); err != nil {
				return nil, fmt.Errorf("write row of %s: %w", sheet.Name, err)
			}
		}
		if err := writer.Flush(); err != nil {
			return nil, fmt.Errorf("flush sheet %s: %w", sheet.Name, err)
		}
	}
	file.SetActiveSheet(0)

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("encode workbook: %w", err)
	}
	return buffer.Bytes(), nil
}

// cells keeps every value a string so ids such as "0012" and dates such as
// "2026-01-05" survive the round trip without Excel reinterpreting them.
func cells(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func encodeCSV(sheets []Table) ([]byte, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for _, sheet := range sheets {
		entry, err := archive.Create(sheet.Name + ".csv")
		if err != nil {
			return nil, fmt.Errorf("create %s.csv: %w", sheet.Name, err)
		}
		writer := csv.NewWriter(entry)
		if err := writer.Write(sheet.Columns); err != nil {
			return nil, fmt.Errorf("write %s.csv: %w", sheet.Name, err)
		}
		if err := writer.WriteAll(sheet.Rows); err != nil {
			return nil, fmt.Errorf("write %s.csv: %w", sheet.Name, err)
		}
	}
	if err := archive.Close(); err != nil {
		return nil, fmt.Errorf("close archive: %w", err)
	}
	return buffer.Bytes(), nil
}

func decodeXLSX(content []byte) ([]Table, error) {
	file, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("the workbook could not be read: %w", err)
	}
	defer func() { _ = file.Close() }()

	var result []Table
	for _, name := range file.GetSheetList() {
		rows, err := file.GetRows(name)
		if err != nil {
			return nil, fmt.Errorf("read sheet %s: %w", name, err)
		}
		if len(rows) == 0 {
			continue
		}
		result = append(result, Table{Name: name, Columns: rows[0], Rows: rows[1:]})
	}
	return result, nil
}

func decodeZIP(content []byte) ([]Table, error) {
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("the archive could not be read: %w", err)
	}
	var result []Table
	for _, entry := range archive.File {
		if entry.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name), ".csv") {
			continue
		}
		file, err := entry.Open()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name, err)
		}
		table, err := readCSV(entry.Name, io.LimitReader(file, MaxUploadBytes))
		_ = file.Close()
		if err != nil {
			return nil, err
		}
		if table != nil {
			result = append(result, *table)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("the archive contains no CSV files")
	}
	return result, nil
}

func decodeSingleCSV(filename string, content []byte) ([]Table, error) {
	table, err := readCSV(filename, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, errors.New("the file is empty")
	}
	if Columns(normaliseSheet(table.Name)) == nil {
		return nil, fmt.Errorf("name the file after the sheet it holds, for example students.csv (known sheets: %s)", strings.Join(SheetNames, ", "))
	}
	return []Table{*table}, nil
}

func readCSV(name string, reader io.Reader) (*Table, error) {
	parser := csv.NewReader(reader)
	parser.FieldsPerRecord = -1
	parser.TrimLeadingSpace = true
	rows, err := parser.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s could not be read as CSV: %w", path.Base(name), err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	// Strip the byte order mark Excel writes in front of the first header.
	if len(rows[0]) > 0 {
		rows[0][0] = strings.TrimPrefix(rows[0][0], "\ufeff")
	}
	return &Table{Name: path.Base(name), Columns: rows[0], Rows: rows[1:]}, nil
}
