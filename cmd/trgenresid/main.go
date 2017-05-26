package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Ragnis/go-translate"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("trgenresid: ")

	names := &NameSet{nil, make(map[string]bool), false}

	for _, path := range os.Args[1:] {
		if err := names.Load(path); err != nil {
			panic(err)
		}
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "// Code generated by \"trgenresid %s\"; DO NOT EDIT.\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintf(&buf, "\n")
	fmt.Fprintf(&buf, "package %s", os.Getenv("GOPACKAGE"))
	fmt.Fprintf(&buf, "\n")

	fmt.Fprintf(&buf, "// VersionHash is a string uniquely identifying this set of resource IDs\n")
	fmt.Fprintf(&buf, "const VersionHash = \"%s\"\n", names.Hash())
	fmt.Fprintf(&buf, "\n")

	fmt.Fprintf(&buf, "// Resource identifiers\n")
	fmt.Fprintf(&buf, "const (\n")

	for i, name := range names.Names() {
		fmt.Fprintf(&buf, "\t%s", name)
		if i == 0 {
			fmt.Fprintf(&buf, " uint = iota")
		}
		fmt.Fprintf(&buf, "\n")
	}

	fmt.Fprintf(&buf, ")\n")

	// Format the generated code

	src, err := format.Source(buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
	}

	if err := ioutil.WriteFile(outFileName(), src, 0644); err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// NameSet represents a set of translation names
type NameSet struct {
	names    []string
	set      map[string]bool
	needSort bool
}

// Load looks up translation files using the specified path and loads the names
// from them
func (ns *NameSet) Load(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return ns.loadDir(path)
	}
	return ns.loadFile(path)
}

// Add adds a name to this set
func (ns *NameSet) Add(name string) {
	if _, ok := ns.set[name]; !ok {
		ns.set[name] = true
		ns.names = append(ns.names, name)
		ns.needSort = true
	}
}

// Names returns the names in this set
func (ns *NameSet) Names() []string {
	if ns.needSort {
		sort.Strings(ns.names)
	}
	return ns.names
}

// Hash returns a unique hash returning this set of names
func (ns *NameSet) Hash() string {
	s := strings.Join(ns.Names(), ",")
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func (ns *NameSet) loadDir(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.New("reading directory: " + err.Error())
	}
	for _, file := range files {
		if file.Name()[0] == '.' {
			continue
		}
		fp := filepath.Join(path, file.Name())
		if file.IsDir() {
			err = ns.loadDir(fp)
		} else {
			err = ns.loadFile(fp)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (ns *NameSet) loadFile(path string) error {
	if filepath.Ext(path) != ".json" {
		return errors.New("not a JSON file")
	}
	strs := &translate.StringsData{}
	if err := readJSON(path, strs); err != nil {
		return errors.New("reading strings file: " + err.Error())
	}
	for k := range strs.Strings {
		ns.Add(k)
	}
	return nil
}

// outFileName returns a string to use as the output file name
func outFileName() string {
	file := os.Getenv("GOFILE")
	if file == "" {
		return "resid.go"
	}
	ext := filepath.Ext(file)
	if ext == ".go" {
		file = file[:len(file)-len(ext)] + "_resid" + ext
	}
	return file
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