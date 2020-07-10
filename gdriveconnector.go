package main

import (
	"fmt"

	"google.golang.org/api/drive/v3"
)

//func crawlForMistakes

//func

func crawlHierarchy(folderID string) ([]Employee, error) {
	srv, err := createDriveService()
	check(err)

	r, err := crawlHierarchyRecursive(srv, "_hierarchy", folderID, nil)

	return r, err
}

func crawlHierarchyRecursive(driveSrv *drive.Service, folderName string, folderID string, compoundList []Employee) ([]Employee, error) {

	r, err := driveSrv.Files.List().
		Fields("files(id, name)").
		Q("'" + folderID + "' in parents and trashed=false and mimeType = 'application/vnd.google-apps.folder'").
		Do()

	check(err)

	//we append results to master list
	if compoundList == nil {
		compoundList = []Employee{}
	}

	for _, curFile := range r.Files {

		fmt.Printf("FOUND: %s (%s)\n", curFile.Name, curFile.Id)
		compoundList = append(compoundList, Employee{Pseudo: curFile.Name, ManagerPseudo: folderName, FolderID: curFile.Id})

		compoundList, err = crawlHierarchyRecursive(driveSrv, curFile.Name, curFile.Id, compoundList)
		check(err)
	}

	return compoundList, err
}
