package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"google.golang.org/api/sheets/v4"
)

//An Employee represents an employee within the folder hiearchy
type Employee struct {
	Pseudo        string
	ManagerPseudo string
	FolderID      string
}

// ByEmployee implements sort.Interface based on the Pseudo(Employee) field.
type ByEmployee []Employee

func (a ByEmployee) Len() int           { return len(a) }
func (a ByEmployee) Less(i, j int) bool { return a[i].Pseudo < a[j].Pseudo }
func (a ByEmployee) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func employeeListToMap(employeeList []Employee) map[string]*Employee {

	employeeMap := make(map[string]*Employee)

	for i, curEmployee := range employeeList {
		employeeMap[curEmployee.Pseudo] = &employeeList[i]
	}

	return employeeMap
}

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

//extractFileID removes everything before the last /
func extractFileID(fileLink string) string {

	subStrings := strings.Split(fileLink, "/")
	fileID := subStrings[len(subStrings)-1]
	return fileID
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

func valueRangeToEmployeeList(valueRange *sheets.ValueRange) []Employee {

	var employeeList []Employee

	for _, curLine := range valueRange.Values {

		//skipping empty lines
		if len(curLine) > 0 {

			var newEmployee Employee
			newEmployee.Pseudo = curLine[0].(string)

			if len(curLine) > 1 {
				newEmployee.ManagerPseudo = curLine[1].(string)
			}
			if len(curLine) > 2 {
				newEmployee.FolderID = extractFileID(curLine[2].(string))
			}

			employeeList = append(employeeList, newEmployee)
		}
	}

	return employeeList
}
