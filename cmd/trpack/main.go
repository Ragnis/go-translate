package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ragnis/go-translate"
)

func main() {
	ids := &ResourceIDSet{}
	if err := ids.LoadPackage("resid"); err != nil {
		fmt.Println("error reading resource IDs: " + err.Error())
		os.Exit(1)
	}

	for _, f := range os.Args[1:] {
		if err := packFile(f, ids); err != nil {
			fmt.Printf("error creating pack for '%s': %v\n", f, err)
		}
	}
}

// readJSON parses JSON from a file
func readJSON(path string, into interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("opening file: " + err.Error())
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(into)
}

// writeJSON writes JSON to a file
func writeJSON(path string, data interface{}) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.New("opening file: " + err.Error())
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(data)
}

// packFileName returns the file name for packed output
func packFileName(path string) string {
	if ext := filepath.Ext(path); ext == ".json" {
		path = path[:len(path)-len(ext)] + ".pak" + ext
	}
	return path
}

// packFile creates a .pak.json file for the input json file.
func packFile(path string, ids *ResourceIDSet) error {
	lang := &translate.StringsData{}
	if err := readJSON(path, lang); err != nil {
		return errors.New("could not read language file: " + err.Error())
	}
	pak := &translate.PackedStringsData{
		Lang:        lang.Lang,
		VersionHash: ids.VersionHash,
		Strings:     make([]string, ids.MaxID()+1),
	}
	for name, value := range lang.Strings {
		id, ok := ids.Names[name]
		if !ok {
			return fmt.Errorf("no resource ID for name '%s'", name)
		}
		pak.Strings[id] = value
	}
	return writeJSON(packFileName(path), pak)
}
