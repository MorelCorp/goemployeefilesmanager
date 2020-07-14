package main

import "google.golang.org/api/sheets/v4"

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

	/////TODO #1 c'est là qu'on est rendu!
	////srv.Spreadsheets.BatchUpdate

	newSheet.

	spreadsheetID := newSheet.SpreadsheetId
	debugLog("New spreadsheet:%s", spreadsheetID)

	return "", nil
}
