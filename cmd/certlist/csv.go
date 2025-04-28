package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"

	"github.com/mpkondrashin/certlist/pkg/smsbackup"
)

// getHeaders extracts struct field names as CSV headers
func getHeaders[T any](useTags bool) []string {
	var t T
	typ := reflect.TypeOf(t)
	headers := make([]string, typ.NumField())

	for i := range typ.NumField() {
		if !useTags {
			headers[i] = typ.Field(i).Name
			continue
		}
		header := typ.Field(i).Tag.Get("csv")
		if header == "" {
			continue
		}
		headers[i] = header
	}

	return headers
}

// structToSlice converts a struct to a string slice for CSV
func structToSlice(v any, useTags bool) []string {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	//fields := val.NumField()
	var row []string

	for i := range typ.NumField() {
		if useTags && typ.Field(i).Tag.Get("csv") == "" {
			continue
		}
		row = append(row, fmt.Sprintf("%v", val.Field(i).Interface()))
	}
	return row
}

func SaveCSV(filename string, data []smsbackup.ReportLine, useTags bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := getHeaders[smsbackup.ReportLine](useTags)
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, line := range data {
		row := structToSlice(line, useTags)
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
