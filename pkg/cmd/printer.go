package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const maxCellWidth = 50

// renderTable safely formats slices of maps or structs into a psql-style terminal table.
func renderTable(data any) {
	val, elemType, n, valid := normalizeData(data)
	if !valid {
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Emulate psql tabular styling
	t.SetStyle(table.StyleLight)
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateHeader = true
	t.Style().Format.Header = text.FormatUpper

	headers := extractHeaders(val, elemType, n)
	if len(headers) == 0 {
		fmt.Println("(0 columns)")
		return
	}

	t.AppendHeader(convertToTableRow(headers))

	for i := 0; i < n; i++ {
		rowVal := val.Index(i)
		if rowVal.Kind() == reflect.Ptr {
			rowVal = rowVal.Elem()
		}
		var rowData table.Row
		for _, h := range headers {
			fieldVal := getFieldValue(rowVal, elemType, h)
			rowData = append(rowData, formatCell(fieldVal, true))
		}
		t.AppendRow(rowData)
	}

	t.Render()
	fmt.Printf("(%d rows)\n\n", n)
}

// renderExpanded formats data vertically (Key | Value) per record, emulating psql's \x mode.
func renderExpanded(data any) {
	val, elemType, n, valid := normalizeData(data)
	if !valid {
		return
	}

	headers := extractHeaders(val, elemType, n)
	if len(headers) == 0 {
		fmt.Println("(0 columns)")
		return
	}

	for i := 0; i < n; i++ {
		fmt.Printf("-[ RECORD %d ]-------------------------\n", i+1)

		rowVal := val.Index(i)
		if rowVal.Kind() == reflect.Ptr {
			rowVal = rowVal.Elem()
		}

		for _, h := range headers {
			fieldVal := getFieldValue(rowVal, elemType, h)
			// Print without truncating, formatting JSON beautifully
			rawStr := formatCell(fieldVal, false)
			fmt.Printf("%-20s | %s\n", h, rawStr)
		}
	}
	fmt.Printf("\n(%d rows)\n\n", n)
}

// normalizeData ensures the input is always a slice and extracts its underlying reflection properties.
func normalizeData(data any) (reflect.Value, reflect.Type, int, bool) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if !val.IsValid() {
		fmt.Println("(null)")
		return reflect.Value{}, nil, 0, false
	}

	// Normalize single items into a slice for uniform processing
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		slice := reflect.MakeSlice(reflect.SliceOf(val.Type()), 1, 1)
		slice.Index(0).Set(val)
		val = slice
	}

	n := val.Len()
	if n == 0 {
		fmt.Println("(0 rows)")
		return reflect.Value{}, nil, 0, false
	}

	elemType := val.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	return val, elemType, n, true
}

// extractHeaders dynamically aggregates column names depending on if the data is a Map or a Struct.
func extractHeaders(val reflect.Value, elemType reflect.Type, n int) []string {
	var headers []string

	if elemType.Kind() == reflect.Map {
		headerSet := make(map[string]bool)
		// Aggregate all unique keys across all documents to handle schemaless data
		for i := 0; i < n; i++ {
			rowVal := val.Index(i)
			if rowVal.Kind() == reflect.Ptr {
				rowVal = rowVal.Elem()
			}
			for _, key := range rowVal.MapKeys() {
				headerSet[key.String()] = true
			}
		}
		for k := range headerSet {
			headers = append(headers, k)
		}
		sort.Strings(headers) // Sort headers for deterministic column ordering
	} else if elemType.Kind() == reflect.Struct {
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			if !field.IsExported() {
				continue
			}
			name := field.Tag.Get("json")
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			} else {
				name = strings.Split(name, ",")[0]
			}
			headers = append(headers, name)
		}
	} else {
		fmt.Printf("Unsupported table type: %s\n", elemType.Kind())
	}

	return headers
}

// getFieldValue abstracts retrieving a value by name for both Maps and Structs.
func getFieldValue(rowVal reflect.Value, elemType reflect.Type, header string) reflect.Value {
	if elemType.Kind() == reflect.Map {
		return rowVal.MapIndex(reflect.ValueOf(header))
	}

	// For structs, we need to match the header back to the struct field (either by JSON tag or Field Name)
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		name := field.Tag.Get("json")
		if name != "" {
			name = strings.Split(name, ",")[0]
		} else {
			name = field.Name
		}
		if name == header {
			return rowVal.Field(i)
		}
	}
	return reflect.Value{}
}

// formatCell handles nested JSON, arrays, and long strings gracefully.
func formatCell(v reflect.Value, truncate bool) string {
	if !v.IsValid() || v.IsZero() {
		return ""
	}

	// Unwrap interfaces
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	var strVal string

	// Minify nested maps and slices into compact JSON strings
	if v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Struct {
		// Use pretty JSON if we are not truncating (expanded mode)
		var b []byte
		var err error
		if !truncate {
			b, err = json.MarshalIndent(v.Interface(), "", "  ")
		} else {
			b, err = json.Marshal(v.Interface())
		}

		if err == nil {
			strVal = string(b)
		} else {
			strVal = fmt.Sprintf("%v", v.Interface())
		}
	} else {
		strVal = fmt.Sprintf("%v", v.Interface())
	}

	if truncate {
		// Clean up formatting for the standard table
		strVal = strings.ReplaceAll(strVal, "\n", " ")
		strVal = strings.ReplaceAll(strVal, "\r", "")
		strVal = strings.TrimSpace(strVal)

		// Truncate long text to prevent terminal stretching
		if len(strVal) > maxCellWidth {
			return strVal[:maxCellWidth-3] + "..."
		}
	}

	return strVal
}

func convertToTableRow(headers []string) table.Row {
	row := make(table.Row, len(headers))
	for i, h := range headers {
		row[i] = h
	}
	return row
}
