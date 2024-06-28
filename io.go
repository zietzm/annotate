package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func readCsv(inputFile string, selectedColumn string, annotationColumn string) ([]item, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	headers := records[0]
	textColumnIndex := -1
	annotationColumnIndex := -1
	for i, header := range headers {
		if header == selectedColumn {
			textColumnIndex = i
		}
		if header == annotationColumn {
			annotationColumnIndex = i
		}
	}
	if textColumnIndex == -1 {
		return nil, fmt.Errorf("no 'text' column found in CSV")
	}

	items := make([]item, len(records)-1)
	for i, record := range records[1:] {
		title := record[0]
		description := record[textColumnIndex]
		var annotation = ""
		if annotationColumnIndex != -1 {
			annotation = record[annotationColumnIndex]
		}
		items[i] = item{title: title, text: description, annotation: annotation}
	}

	return items, nil
}

func writeCsv(outputFile string, records []item) error {
	outputF, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outputF.Close()

	writer := csv.NewWriter(outputF)
	defer writer.Flush()

	headers := []string{"title", "text", "annotation"}
	err = writer.Write(headers)
	if err != nil {
		return fmt.Errorf("error writing CSV: %w", err)
	}

	for _, record := range records {
		err = writer.Write([]string{record.title, record.text, record.annotation})
		if err != nil {
			return fmt.Errorf("error writing CSV: %w", err)
		}
	}

	return nil
}
