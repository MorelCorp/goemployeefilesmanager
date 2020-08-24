package main

import (
	"google.golang.org/api/drive/v3"
)

// insertFolder will create a new folder under the manager folder name and return the new folder ID as well
// as the updated full array of employees
func insertFolder(rootFolderID string, managerFolderName string, newFolderName string) (string, []Employee) {

	employees, err := crawlHierarchy(rootFolderID)
	check(err)

	var manager *Employee = nil

	for _, employee := range employees {

		if employee.Pseudo == managerFolderName {
			manager = &employee
			break
		}
	}

	if manager == nil {
		return "", nil
	}

	newFolderID, err := createFolder(manager.FolderID, newFolderName)
	check(err)

	newEmployee := Employee{Pseudo: newFolderName, ManagerPseudo: manager.Pseudo, FolderID: newFolderID}
	employees = append(employees, newEmployee)

	return newFolderID, employees
}

func crawlHierarchy(folderID string) ([]Employee, error) {
	srv, err := createDriveService()
	check(err)

	r, err := crawlHierarchyRecursive(srv, "", folderID, nil)

	return r, err
}

func crawlHierarchyRecursive(driveSrv *drive.Service, folderName string, folderID string, compoundList []Employee) ([]Employee, error) {

	query := "'" + folderID + "' in parents and trashed=false and mimeType='application/vnd.google-apps.folder'"
	r, err := driveSrv.Files.List().
		Fields("files(id, name)").
		Q(query).
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

	if TrialRunOnly {
		debugLog("Trial Run: Giving %s access to %s.", emailAdress, itemID)
		return nil
	}

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

		//we never want to update access rights to the archive folder
		if curFile.Name == ArchiveFolderName {
			continue
		}

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

	if TrialRunOnly {
		debugLog("Trial Run: Copying %s (Potential New title:%s) in %s.", sourceDocumentID, newTitle, targetFolderID)
		return nil
	}

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

	if TrialRunOnly {
		debugLog("Trial Run: Move required for %s from %s to %s.", documentID, curParentID, targetFolderID)
		return nil
	}

	srv, err := createDriveService()
	check(err)

	_, err = srv.Files.Update(documentID, nil).AddParents(targetFolderID).RemoveParents(curParentID).Do()
	check(err)

	return nil
}

func createFolder(parentFolderID string, folderName string) (string, error) {

	if TrialRunOnly {
		debugLog("Trial Run: Create folder \"%s\" required in folder %s.", folderName, parentFolderID)
		return "", nil
	}

	srv, err := createDriveService()
	check(err)

	f := &drive.File{
		Name:     folderName,
		Parents:  []string{parentFolderID},
		MimeType: "application/vnd.google-apps.folder",
	}

	createdFileInfo, err := srv.Files.Create(f).Do()
	check(err)

	return createdFileInfo.Id, nil
}
