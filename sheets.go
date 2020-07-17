package main

import (
	"google.golang.org/api/sheets/v4"
)

func employeeListToSheet(sheetTitle string, employeeList []Employee) (string, error) {

	valueRange := employeeListToValueRange(employeeList)

	srv, err := createSheetsService()
	check(err)

	newSheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: sheetTitle,
		},
	}

	newSheet, err = srv.Spreadsheets.Create(newSheet).Do()
	check(err)

	spreadsheetID := newSheet.SpreadsheetId
	_, err = srv.Spreadsheets.Values.Append(spreadsheetID, valueRange.Range, valueRange).ValueInputOption("USER_ENTERED").Do()
	check(err)

	return spreadsheetID, nil
}
