package main

import (
	"encoding/json"
	"io/ioutil"
)

func main() {
	r, err := crawlHierarchy("1U36pQdHin4TFOmPikHBtk83rAf-Qgb6n")
	check(err)

	jsonOutput, err := json.Marshal(r)
	check(err)

	err = ioutil.WriteFile("jsonOutput.json", jsonOutput, 0644)
	check(err)

}
