package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	//DomainName is the email domain name used access rights attribution
	DomainName = "coveo.com"

	//ArchiveFolderName is the string name of the folder where archived(deleted) files will be put
	ArchiveFolderName = "_archive"

	helpFile = "help.txt"
)

//TrialRunOnly is a global flag preventing any change to files in Drive if enabled. Only log output will be generated.
var TrialRunOnly bool = false

func usage() {
	fmt.Println("GoEmployeeFilesManager helper program")
	fmt.Println("Possible commands: (help for full details)")
	fmt.Println("- help:\t\t\tgo run goemployeefilesmanager help")
	fmt.Println("- authenticate:\t\tgo run goemployeefilesmanager authenticate")
	fmt.Println("- crawl:\t\tgo run goemployeefilesmanager crawl <ROOT_FOLDER_ID>")
	fmt.Println("- updateHierarchy:\tgo run goemployeefilesmanager updateHierarchy <ROOT_FOLDER_ID> <TARGET_EMPLOYEE_ROSTER_FILD_ID>")
	fmt.Println("- updateaccessrights:\tgo run goemployeefilesmanager updateHierarchy <ROOT_FOLDER_ID>")
	fmt.Println("- distribute:\t\tgo run goemployeefilesmanager updateHierarchy <ROOT_FOLDER_ID> <SOURCE_FILE_ID> <FILE_COPIES_PREFIX>")
}

// help displays the different options to the user
func help() {
	file, err := os.Open(helpFile)

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	for _, eachline := range txtlines {
		fmt.Println(eachline)
	}
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

	if sheetOutput && !TrialRunOnly {
		newSheetID, err := employeeListToSheet("Employee Roster", employeeList)
		debugLog("Sheet created: %s", spreadsheetLinkFormat(newSheetID))
		check(err)
	}

	if (!sheetOutput && !jsonOutput) || TrialRunOnly {
		fmt.Println(employeeList)
	}

	return nil
}

//updateHierarchy will use the spreadsheet id in param and parse the folder hierarchy to define and apply what needs to be updated
func updateHierarchy(rootFolderID string, employeeRosterSheetID string) {

	var found bool

	//read roster tree
	expectedHierarchy := importHierarchy(employeeRosterSheetID)

	//parse folder hierarchy and note problems
	curHierarchy, err := crawlHierarchy(rootFolderID)
	check(err)

	curHierarchyMap := employeeListToMap(curHierarchy)

	//let's make sure we have the archive folder (if needed)
	_, found = curHierarchyMap[ArchiveFolderName]
	if !found {

		newArchiveFolderID, err := createFolder(rootFolderID, ArchiveFolderName)
		check(err)

		fakeEmployee := Employee{
			Pseudo:   ArchiveFolderName,
			FolderID: newArchiveFolderID,
		}

		oneElementSlice := []Employee{fakeEmployee}
		curHierarchy = append(oneElementSlice, curHierarchy...)
	}

	//and in the same way, we'll have to pass along the root folder
	curHierarchy = append(curHierarchy, Employee{Pseudo: "", FolderID: rootFolderID})

	//let's list all the work to do
	var workToDo []Work

	for _, curExpectedEmployee := range expectedHierarchy {

		newWork := Work{
			target: curExpectedEmployee,
		}

		realEmployee, found := curHierarchyMap[curExpectedEmployee.Pseudo]

		if found != true {
			newWork.op = create
			workToDo = append(workToDo, newWork)
		} else if realEmployee.ManagerPseudo != curExpectedEmployee.ManagerPseudo {
			newWork.op = move
			newWork.cur = *realEmployee
			workToDo = append(workToDo, newWork)
		}
	}

	//check the ones to delete (move to archive folder)
	expectedHierarchyMap := employeeListToMap(expectedHierarchy)
	for _, curRealEmployee := range curHierarchy {

		//sad not very nice hack...
		if curRealEmployee.Pseudo == ArchiveFolderName || curRealEmployee.Pseudo == "" {
			//we never want to remove the Archive or root folder
			continue
		}

		_, found = expectedHierarchyMap[curRealEmployee.Pseudo]
		if !found {
			newWork := Work{
				cur: curRealEmployee,
				op:  archive,
			}
			workToDo = append(workToDo, newWork)
		}
	}

	doWork(workToDo, curHierarchy)

	//to improve: update the roster file (even though it is possible to crawl for it from scratch...)
}

// distribute will add one copy of the provided document in each folder of the hierarchy
func distribute(rootFolderID string, sourceDocumentID string, prefix string) error {
	return distributeDocument(rootFolderID, sourceDocumentID, prefix)
}

func validateParamsNumber(requiredParamsNumber int, params []string, silent bool) bool {
	if len(params) >= requiredParamsNumber {
		return true
	}

	if !silent {
		fmt.Println("Expected more parameters.")
	}
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
		if validateParamsNumber(1, params, false) {
			crawl(params[0], false, true)
		}

	case "updatehierarchy":
		if validateParamsNumber(2, params, false) {
			updateHierarchy(params[0], params[1])
		}

	case "updateaccessrights":
		if validateParamsNumber(1, params, false) {
			updateAccessRights(params[0], false)
		}

	case "distribute":
		if validateParamsNumber(3, params, true) {
			distribute(params[0], params[1], params[2])
		} else if validateParamsNumber(2, params, false) {
			distribute(params[0], params[1], "")
		}

	case "help":
		help()

	default:
		usage()
	}
}
