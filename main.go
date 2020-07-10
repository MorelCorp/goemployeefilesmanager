package main

import (
	"fmt"
	"io/ioutil"

	"google.golang.org/api/drive/v3"
)

func listRecursiveFoldersInFolder(driveSrv *drive.Service, folderID string, compoundList *drive.FileList) (*drive.FileList, error) {

	r, err := driveSrv.Files.List().
		/////PageSize(100). ////TODO handle pagination
		Fields("files(id, name)").
		Q("'" + folderID + "' in parents and trashed=false and mimeType = 'application/vnd.google-apps.folder'").
		Do()

	check(err)

	fmt.Println("Searching in:" + folderID + "{")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			fmt.Printf("\t%s (%s)\n", i.Name, i.Id)
		}
	}
	fmt.Println("}")

	//we append results to master list
	if compoundList == nil {
		compoundList = r
	} else {
		compoundList.Files = append(compoundList.Files, r.Files...)
	}

	for _, curFile := range r.Files {
		r, err = listRecursiveFoldersInFolder(driveSrv, curFile.Id, compoundList)
		check(err)
	}

	return compoundList, err
}

func main() {

	srv, err := createDriveService()
	check(err)

	r, err := listRecursiveFoldersInFolder(srv, "1U36pQdHin4TFOmPikHBtk83rAf-Qgb6n", nil)
	check(err)

	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			fmt.Printf("%s (%s) [%s]\n", i.Name, i.Id, i.MimeType)
		}
	}

	jsonOutput, err := r.MarshalJSON()
	check(err)

	err = ioutil.WriteFile("jsonOutput.json", jsonOutput, 0644)
	check(err)

}
