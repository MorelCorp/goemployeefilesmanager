package main

import (
	"google.golang.org/api/drive/v3"
)

func crawlHierarchy(folderID string) ([]Employee, error) {
	srv, err := createDriveService()
	check(err)

	r, err := crawlHierarchyRecursive(srv, "", folderID, nil)

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

		debugLog("FOUND: %s (%s)\n", curFile.Name, curFile.Id)
		compoundList = append(compoundList, Employee{Pseudo: curFile.Name, ManagerPseudo: folderName, FolderID: curFile.Id})

		compoundList, err = crawlHierarchyRecursive(driveSrv, curFile.Name, curFile.Id, compoundList)
		check(err)
	}

	return compoundList, err
}

func allowAccess(driveSrv *drive.Service, itemID string, emailAdress string, notifyUser bool) error {

	newPermission := &drive.Permission{
		EmailAddress: emailAdress,
		Type:         "user",
		Role:         "writer",
	}

	_, err := driveSrv.Permissions.Create(itemID, newPermission).SendNotificationEmail(notifyUser).Do()
	return err
}

func updateAccessRightsRecursive(driveSrv *drive.Service, folderID string, notifyUsers bool) error {

	//get all sub folders
	r, err := driveSrv.Files.List().
		Fields("files(id, name)").
		Q("'" + folderID + "' in parents and trashed=false and mimeType = 'application/vnd.google-apps.folder'").
		Do()
	check(err)

	for _, curFile := range r.Files {

		debugLog("FOUND: %s (%s)\n", curFile.Name, curFile.Id)

		err = allowAccess(driveSrv, curFile.Id, curFile.Name+"@"+DomainName, notifyUsers)
		check(err)

		err = updateAccessRightsRecursive(driveSrv, curFile.Id, notifyUsers)
		check(err)
	}

	return err
}

func updateAccessRights(rootFolderID string, notifyUsers bool) error {

	srv, err := createDriveService()
	check(err)

	return updateAccessRightsRecursive(srv, rootFolderID, notifyUsers)
}

func copyDocument(driveSrv *drive.Service, targetFolderID string, sourceDocumentID string, newTitle string) error {

	// if new title = "" we just skip new title
	f := &drive.File{}
	if newTitle != "" {
		f.Name = newTitle
	}
	f.Parents = []string{targetFolderID}

	_, err := driveSrv.Files.Copy(sourceDocumentID, f).Do()
	check(err)

	return err
}

func distributeDocument(rootFolderID string, sourceDocumentID string, titlePrefix string) error {
	srv, err := createDriveService()
	check(err)

	err = distributeDocumentRecursive(srv, rootFolderID, sourceDocumentID, titlePrefix)

	return err
}

func distributeDocumentRecursive(driveSrv *drive.Service, folderID string, sourceDocumentID string, titlePrefix string) error {

	r, err := driveSrv.Files.List().
		Fields("files(id, name)").
		Q("'" + folderID + "' in parents and trashed=false and mimeType = 'application/vnd.google-apps.folder'").
		Do()

	check(err)

	for _, curFile := range r.Files {

		var newTitle = titlePrefix
		if newTitle != "" {
			newTitle = newTitle + curFile.Name
		}

		debugLog("FOUND: %s (%s)\n", curFile.Name, curFile.Id)
		err = copyDocument(driveSrv, curFile.Id, sourceDocumentID, newTitle)
		check(err)

		err = distributeDocumentRecursive(driveSrv, curFile.Id, sourceDocumentID, titlePrefix)
		check(err)
	}

	return err
}

func moveFile(documentID string, curParentID string, targetFolderID string) error {

	srv, err := createDriveService()
	check(err)

	_, err = srv.Files.Update(documentID, nil).AddParents(targetFolderID).RemoveParents(curParentID).Do()
	check(err)

	return nil
}
