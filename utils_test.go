package main

import "testing"

func TestEmployeeListToArray(t *testing.T) {

	testList := []Employee{
		{
			Pseudo:        "pseudo",
			ManagerPseudo: "managerpseudo",
			FolderID:      "folderid",
		},
	}

	testArray := employeeListToArray(testList)

	if len(testArray) != 2 {
		t.Errorf("Expected 1 row, got %d", len(testArray))
	}

	if len(testArray[0]) != 3 {
		t.Errorf("Expected 3 colums, got %d", len(testArray))
	}

	if testArray[1][0] != "pseudo" {
		t.Errorf("First column is for employee pseudo, got %s instead.", testArray[0][0])
	}
	if testArray[1][1] != "managerpseudo" {
		t.Errorf("Second column is for manager pseudo, got %s instead.", testArray[0][1])
	}
	if testArray[1][2] != "folderid" {
		t.Errorf("Third column is for folder link pseudo, got %s instead.", testArray[0][2])
	}
}
