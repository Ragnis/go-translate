package translate

// StringsData is the strings JSON structure
type StringsData struct {
	// Lang is the language identifier
	Lang string `json:"lang"`
	// Strings is a map of string and their translations
	Strings map[string]string `json:"strings"`
}

// PackedStringsData is the packed strings JSON structure. This structure
// shouldn't be used externally and may change at any time.
type PackedStringsData struct {
	// Lang is the language identifier
	Lang string
	// VersionHash is the resource IDs version for which this data was
	// generated
	VersionHash string
	// Strings is a list of translations
	Strings []string
}
