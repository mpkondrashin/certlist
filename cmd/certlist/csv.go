package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"

	"github.com/mpkondrashin/certlist/pkg/smsbackup"
)

// getHeaders extracts struct field names as CSV headers
func getHeaders[T any]() []string {
	var t T
	typ := reflect.TypeOf(t)
	headers := make([]string, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		headers[i] = typ.Field(i).Name
	}

	return headers
}

// structToSlice converts a struct to a string slice for CSV
func structToSlice(v any) []string {
	val := reflect.ValueOf(v)
	fields := val.NumField()
	row := make([]string, fields)

	for i := 0; i < fields; i++ {
		row[i] = fmt.Sprintf("%v", val.Field(i).Interface()) // Convert to string
	}

	return row
}

func SaveCSV(filename string, data []smsbackup.ReportLine) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := getHeaders[smsbackup.ReportLine]()
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, line := range data {
		row := structToSlice(line)
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
