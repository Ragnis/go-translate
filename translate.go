package translate

import (
	"encoding/json"
	"errors"
	"os"
)

// Domain keeps track a set of loaded languages
type Domain struct {
	langs       map[string]*Language
	versionHash string
}

// DefaultDomain is the default domain
var DefaultDomain = NewDomain()

// NewDomain creates a new empty translation domain
func NewDomain() *Domain {
	return &Domain{
		langs: make(map[string]*Language),
	}
}

// SetVersionHash requires all future translation files to have the specified
// hash value set. Setting this to an empty string disables any checks.
func (d *Domain) SetVersionHash(h string) {
	d.versionHash = h
}

// Language returns a loaded language by it's name
func (d Domain) Language(name string) (*Language, error) {
	if l, ok := d.langs[name]; ok {
		return l, nil
	}
	return nil, errors.New("unknown language: " + name)
}

// LoadStrings loads strings from the provided path into this domain
func (d *Domain) LoadStrings(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("opening file: " + err.Error())
	}
	defer f.Close()
	sd := &PackedStringsData{}
	if err := json.NewDecoder(f).Decode(sd); err != nil {
		return errors.New("could not decode strings: " + err.Error())
	}
	if d.versionHash != "" && d.versionHash != sd.VersionHash {
		return errors.New("version hash mismatch")
	}
	lang, ok := d.langs[sd.Lang]
	if !ok {
		lang = &Language{name: sd.Lang}
		d.langs[sd.Lang] = lang
	}
	return lang.addStrings(sd)
}

// MustLoadStrings is the same as LoadStrings, except that it panics when an
// error occurs
func (d *Domain) MustLoadStrings(path string) {
	if err := d.LoadStrings(path); err != nil {
		panic(err)
	}
}
