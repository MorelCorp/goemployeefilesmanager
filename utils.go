package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"google.golang.org/api/sheets/v4"
)

func check(err error) {
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func debugLog(message string, a ...interface{}) {
	log.Printf(message, a...)
}

func folderLinkFormat(folderID string) string {
	return "https://drive.google.com/drive/folders/" + folderID
}

func spreadsheetLinkFormat(spreadsheetID string) string {
	return "https://docs.google.com/spreadsheets/d/" + spreadsheetID
}

func writeEmployeeListToFile(employeeList []Employee, filename string) error {
	jsonOutput, err := json.Marshal(employeeList)
	check(err)

	if err == nil {
		err = ioutil.WriteFile(filename, jsonOutput, 0644)
	}

	return (err)
}

func employeeListToArray(employeeList []Employee) [][]interface{} {

	rArray := [][]interface{}{}
	rArray = append(rArray, []interface{}{"Employee", "Manager", "Folder Link"})

	for _, curEmployee := range employeeList {
		rArray = append(rArray, []interface{}{curEmployee.Pseudo, curEmployee.ManagerPseudo, folderLinkFormat(curEmployee.FolderID)})
	}

	return rArray
}

func employeeListToValueRange(employeeList []Employee) *sheets.ValueRange {

	rValue := new(sheets.ValueRange)

	dataArray := employeeListToArray(employeeList)
	rValue.Values = dataArray

	rValue.Range = fmt.Sprintf("A1:%s%d", string('A'-1+len(dataArray[0])), len(dataArray))

	return rValue
}
