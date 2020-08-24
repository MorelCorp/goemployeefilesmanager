package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type appConfiguration struct {
	Globals struct {
		RootFolderID      string `json:"root_folder_id"`
		ArchiveFolderName string `json:"archive_folder_name"`
		DomainName        string `json:"domain_name"`
		HelpFile          string `json:"help_file"`
	} `json:"globals"`
	Operations struct {
		SourceDocumentID string `json:"source_document_id"`
		TitlePrefix      string `json:"title_prefix"`
	} `json:"operations"`
}

//AppConfigs contains all global configurations imported from Config.JSON
var AppConfigs appConfiguration

func loadConfigs() {
	f, err := os.Open("./config.json")
	check(err)

	defer f.Close()

	err = json.NewDecoder(f).Decode(&AppConfigs)
	check(err)

}

func usage() {
	fmt.Println("GoEmployeeFilesManager helper program")
	fmt.Println("Possible commands: (help for full details)")
	fmt.Println("- help:\t\t\tgo run goemployeefilesmanager help")
	fmt.Println("- authenticate:\t\tgo run goemployeefilesmanager authenticate")
	fmt.Println("- crawl:\t\tgo run goemployeefilesmanager crawl")
	fmt.Println("- updatehierarchy:\tgo run goemployeefilesmanager updatehierarchy <TARGET_EMPLOYEE_ROSTER_FILD_ID>")
	fmt.Println("- updateaccessrights:\tgo run goemployeefilesmanager updateaccessrights")
	fmt.Println("- distribute:\t\tgo run goemployeefilesmanager distribute")
	fmt.Println("- insert:\t\tgo run goemployeefilesmanager insert <MANAGER_FOLDER_NAME> <EMPLOYEE_FOLDER_NAME>>")
}

// help displays the different options to the user
func help() {
	file, err := os.Open(AppConfigs.Globals.HelpFile)

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
func crawl(jsonOutput bool, sheetOutput bool) error {

	employeeList, err := crawlHierarchy(AppConfigs.Globals.RootFolderID)
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

// distribute will add one copy of the provided document in each folder of the hierarchy
func distribute() error {
	return distributeDocument(AppConfigs.Globals.RootFolderID, AppConfigs.Operations.SourceDocumentID, AppConfigs.Operations.TitlePrefix)
}

//updateHierarchy will use the spreadsheet id in param and parse the folder hierarchy to define and apply what needs to be updated
func updateHierarchy(employeeRosterSheetID string) {

	var found bool

	//read roster tree
	expectedHierarchy := importHierarchy(employeeRosterSheetID)

	//parse folder hierarchy and note problems
	curHierarchy, err := crawlHierarchy(AppConfigs.Globals.RootFolderID)
	check(err)

	curHierarchyMap := employeeListToMap(curHierarchy)

	//let's make sure we have the archive folder (if needed)
	_, found = curHierarchyMap[AppConfigs.Globals.ArchiveFolderName]
	if !found {

		newArchiveFolderID, err := createFolder(AppConfigs.Globals.RootFolderID, AppConfigs.Globals.ArchiveFolderName)
		check(err)

		fakeEmployee := Employee{
			Pseudo:   AppConfigs.Globals.ArchiveFolderName,
			FolderID: newArchiveFolderID,
		}

		oneElementSlice := []Employee{fakeEmployee}
		curHierarchy = append(oneElementSlice, curHierarchy...)
	}

	//and in the same way, we'll have to pass along the root folder
	curHierarchy = append(curHierarchy, Employee{Pseudo: "", FolderID: AppConfigs.Globals.RootFolderID})

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
		if curRealEmployee.Pseudo == AppConfigs.Globals.ArchiveFolderName || curRealEmployee.Pseudo == "" {
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

// insert will add one copy of the provided document in each folder of the hierarchy
func insert(managerFolderName string, newFolderName string) {

	newFolderID, employees := insertFolder(AppConfigs.Globals.RootFolderID, managerFolderName, newFolderName)

	srv, err := createDriveService()
	check(err)
	var newTitle = AppConfigs.Operations.TitlePrefix
	if newTitle != "" {
		newTitle = newTitle + newFolderName
	}
	err = copyDocument(srv, newFolderID, AppConfigs.Operations.SourceDocumentID, newTitle)
	check(err)

	newSheetID, err := employeeListToSheet("New Employees Roster", employees)
	debugLog("Sheet created: %s", spreadsheetLinkFormat(newSheetID))
	check(err)
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

	loadConfigs()

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
		crawl(false, true)

	case "updatehierarchy":
		if validateParamsNumber(1, params, false) {
			updateHierarchy(params[0])
		}

	case "updateaccessrights":
		updateAccessRights(AppConfigs.Globals.RootFolderID, false)

	case "distribute":
		distribute()

	case "insert":
		if validateParamsNumber(2, params, false) {
			insert(params[0], params[1])
		}

	case "help":
		help()

	default:
		usage()
	}
}
