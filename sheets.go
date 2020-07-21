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

func importHierarchy(hierarchyFileID string) []Employee {
	srv, err := createSheetsService()
	check(err)

	// The ranges to retrieve from the spreadsheet.
	// ranges := []string{} // TODO: Update placeholder value.

	// True if grid data should be returned.
	// This parameter is ignored if a field mask was set in the request.
	//includeGridData := false // TODO: Update placeholder value.

	readRange := "Sheet1!A2:C"

	rValue, err := srv.Spreadsheets.Values.Get(hierarchyFileID, readRange).Do()
	check(err)

	return valueRangeToEmployeeList(rValue)
}
