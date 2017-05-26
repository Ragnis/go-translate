package translate

import (
	"encoding/json"
	"errors"
	"os"
)

// Language represents a group of resource values
type Language struct {
	name        string
	versionHash string
	strings     []string
}

// String returns the string value of a resource, or an empty string if no such
// resource exists
func (l Language) String(id uint) string {
	if id < uint(len(l.strings)) {
		return l.strings[id]
	}
	return ""
}

// LoadStrings loads strings from a file
func (l *Language) LoadStrings(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sd := &PackedStringsData{}
	if err := json.NewDecoder(f).Decode(sd); err != nil {
		return errors.New("could not decode strings: " + err.Error())
	}
	return l.addStrings(sd)
}

// addStrings adds strings from a PackedStringsData structure
func (l *Language) addStrings(sd *PackedStringsData) error {
	if sd.Lang != l.name {
		return errors.New("language names mismatch")
	}
	if l.versionHash != "" && l.versionHash != sd.VersionHash {
		return errors.New("version hash mismatch")
	}
	l.name = sd.Lang
	l.versionHash = sd.VersionHash
	l.ensureStringsLen(len(sd.Strings) + 1)
	for id, s := range sd.Strings {
		if s != "" {
			l.strings[id] = s
		}
	}
	return nil
}

// ensureStringsLen extends the string array to the required length. If the
// array is already at the required length, or larger, nothing will be done.
func (l *Language) ensureStringsLen(req int) {
	if req <= len(l.strings) {
		return
	}
	new := make([]string, req)
	copy(l.strings[:len(l.strings)], new[:len(l.strings)])
	l.strings = new
}
