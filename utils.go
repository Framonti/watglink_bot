package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	//"github.com/naoina/toml"
)

func readUnstructuredJson(filePath string, stringName string) string {
	jsonStringFile, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
	}
	defer jsonStringFile.Close()

	jsonByte, _ := ioutil.ReadAll(jsonStringFile)

	var readContent map[string]string
	json.Unmarshal(jsonByte, &readContent)

	return readContent[stringName]
}

func readStringJSON(language string, stringName string) string {

	var loadPath = "data/strings/" + language + "Strings.json" // todo remove hardcoded string and load from config.toml

	return readUnstructuredJson(loadPath, stringName)
}
