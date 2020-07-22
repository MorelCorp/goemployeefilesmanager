package main

import (
	"fmt"
)

//Operation refers to the type of work needed to be performed
type Operation int

const (
	archive Operation = 0 //move the file to archive folder (create archive folder if none present)
	create  Operation = 1 //create the file and then list
	move    Operation = 2 //move the file to new parent folder
)

//Work is to define one change that needs to be performed in the hierarchy
type Work struct {
	op     Operation
	cur    Employee
	target Employee
}

//ToString produces nice output for the Work struct
func (w Work) ToString() string {

	var rString string

	switch w.op {
	case create:
		rString = fmt.Sprintf("{Operation: CREATE for employee %s UNDER %s}", w.target.Pseudo, w.target.ManagerPseudo)
	case move:
		rString = fmt.Sprintf("{Operation: MOVE for employee %s FROM %s TO %s}", w.cur.Pseudo, w.cur.ManagerPseudo, w.target.ManagerPseudo)
	case archive:
		rString = fmt.Sprintf("{Operation: ARCHIVE for employee %s}", w.cur.Pseudo)
	}
	return rString
}

func sortWorkOrders(workOrders *[]Work, curHierarchy []Employee) error {

	//if there is nothing to sort, let's not try to
	if workOrders == nil || len(*workOrders) < 2 {
		return nil
	}

	//init our work datastruct
	allPotentialEmployees := make(map[string]bool)
	employeesAlreadyCreated := make(map[string]bool)
	var createWorkOrders []Work
	var otherWorkOrders []Work

	for _, curEmployee := range curHierarchy {
		allPotentialEmployees[curEmployee.Pseudo] = true
		employeesAlreadyCreated[curEmployee.Pseudo] = true
	}

	for _, curWorkOrder := range *workOrders {
		if curWorkOrder.op == create {
			createWorkOrders = append(createWorkOrders, curWorkOrder)
			allPotentialEmployees[curWorkOrder.target.Pseudo] = true
		} else {
			otherWorkOrders = append(otherWorkOrders, curWorkOrder)
		}
	}

	var returnSlice []Work
	var popped Work
	var found bool

	for notFinished := true; notFinished; notFinished = len(createWorkOrders) > 0 {
		popped, createWorkOrders = createWorkOrders[0], createWorkOrders[1:len(createWorkOrders)]

		_, found = employeesAlreadyCreated[popped.target.ManagerPseudo]
		//if we can already create it without issue, add it to the results
		if found || popped.target.ManagerPseudo == "" {
			returnSlice = append(returnSlice, popped)
			employeesAlreadyCreated[popped.target.Pseudo] = true
		} else {

			//if there's no hope of finding a working solution better to know it earlier than later
			_, found = allPotentialEmployees[popped.target.ManagerPseudo]
			if !found {
				return fmt.Errorf("Missing manager declaration to complete work order. Missing manager:%s (for employee %s)", popped.target.ManagerPseudo, popped.target.Pseudo)
			}

			//otherwise we just add back to the end of the queue and we loop
			createWorkOrders = append(createWorkOrders, popped)
		}
	}

	returnSlice = append(returnSlice, otherWorkOrders...)
	*workOrders = returnSlice

	return nil
}

func doWork(workOrders []Work, curHierarchy []Employee) error {

	err := sortWorkOrders(&workOrders, curHierarchy)
	check(err)

	fileIDMap := make(map[string]string)
	for _, curEmployee := range curHierarchy {
		fileIDMap[curEmployee.Pseudo] = curEmployee.FolderID
	}

	for _, curWorkOrder := range workOrders {

		debugLog(curWorkOrder.ToString())

		switch curWorkOrder.op {
		case archive:
			//archiving is basically moving the folder to the archive folder.
			moveFile(curWorkOrder.cur.FolderID, fileIDMap[curWorkOrder.cur.ManagerPseudo], fileIDMap[ArchiveFolderName])
		case create:
			newFolderID, err := createFolder(fileIDMap[curWorkOrder.target.ManagerPseudo], curWorkOrder.target.Pseudo)
			if err == nil {
				fileIDMap[curWorkOrder.target.Pseudo] = newFolderID
				curWorkOrder.target.FolderID = newFolderID
			}
		case move:
			moveFile(curWorkOrder.cur.FolderID, fileIDMap[curWorkOrder.cur.ManagerPseudo], fileIDMap[curWorkOrder.target.ManagerPseudo])
		}

	}

	return nil
}
