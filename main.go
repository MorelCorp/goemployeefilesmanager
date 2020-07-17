package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	//DomainName is the email domain name used access rights attribution
	DomainName = "coveo.com"
)

// help displays the different options to the user
func help() {

}

// authenticate handles initial auth saving credentials in token.json
// requires credentials.json in the launch folder
func authenticate() error {

	//the first call to create drive service will actually perform the credentials required handshake
	_, err := createDriveService()
	check(err)

	return err
}

// crawl will go through the whole hierarchy and provide folders and links either as a json outputfile, a sheet or if nothing specified as consoole output
func crawl(rootFolderID string, jsonOutput bool, sheetOutput bool) error {

	employeeList, err := crawlHierarchy(rootFolderID)
	check(err)

	if jsonOutput {
		err = writeEmployeeListToFile(employeeList, "jsonOutput.json")
		check(err)
	}

	if sheetOutput {
		newSheetID, err := employeeListToSheet("Employee Roster", employeeList)
		debugLog("Sheet created: %s", spreadsheetLinkFormat(newSheetID))
		check(err)
	}

	if !sheetOutput && !jsonOutput {
		fmt.Println(employeeList)
	}

	return nil
}

// populateHierarchy uses the spreadsheet id in params to create folder hierarchy from specified rootFolderID
func populateHierarchy(rootFolderID string, employeeRosterSheetID string) error {

	return nil
}

//updateHierarchy will use the spreadsheet id in param and parse the folder hierarchy to define and apply what needs to be updated
func updateHierarchy(rootFolderID string, employeeRosterSheetID string) error {

	return nil
}

// distribute will add one copy of the provided document in each folder of the hierarchy
func distribute(rootFolderID string, sourceDocumentID string) error {

	return nil
}

func validateParamsNumber(requiredParamsNumber int, params []string) bool {
	if len(params) >= requiredParamsNumber {
		return true
	}

	fmt.Println("Expected more parameters.")
	return false
}

func main() {

	var functionCall = ""
	var params = []string{}

	if len(os.Args) > 1 {
		functionCall = strings.ToLower(os.Args[1])
		params = os.Args[2:]
	}

	switch functionCall {

	case "authenticate":
		authenticate()

	case "crawl":
		if validateParamsNumber(1, params) {
			crawl(params[0], false, true)
		}

	case "populate":

	case "update":

	case "updateaccessrights":
		if validateParamsNumber(1, params) {
			updateAccessRights(params[0], false)
		}

	case "distribute":
		if validateParamsNumber(2, params) {
			distribute(params[0], params[1])
		}

	case "help":
		fallthrough
	default:
		help()
	}
}
