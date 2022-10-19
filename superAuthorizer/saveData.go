package main

import (
	"encoding/json"
	"os"
)

var saveData map[string]interface{}

func loadSaveData() error {
	data, err := os.ReadFile("multiSav.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &saveData)
}

func saveSaveData() error {
	data, err := json.Marshal(saveData)
	if err != nil {
		return err
	}
	return os.WriteFile("multiSav.json", data, 0600)
}
