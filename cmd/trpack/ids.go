package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
)

// ResourceIDSet represents a set of resource identifiers
type ResourceIDSet struct {
	// VersionHash is the VersionHash value from the resource identifiers package
	VersionHash string
	// Names maps string identifiers to their numeric counterparts
	Names map[string]uint64
}

// LoadPackage loads the identifiers from the named package
func (r *ResourceIDSet) LoadPackage(pkg string) error {
	r.VersionHash = ""
	r.Names = make(map[string]uint64)

	fset := token.NewFileSet()
	files, err := parsePackageFiles(fset, pkg)
	if err != nil {
		return err
	}
	// check types and get the objects
	defs := make(map[*ast.Ident]types.Object)
	if err := check(fset, files, defs); err != nil {
		return err
	}
	// find VersionHash ident
	ident, err := findVersionHashIdent(files)
	if err != nil {
		return err
	}
	// read VersionHash value
	if ident != nil {
		obj, ok := defs[ident]
		if !ok {
			return errors.New("no value for constant 'VersionHash'")
		}
		// the ident is already guaranteed to be a constant
		value := obj.(*types.Const).Val()
		r.VersionHash = constant.StringVal(value)
	}
	// find resource ID idents
	idents, err := findResourceIDIdents(files)
	if err != nil {
		return err
	}
	// read the ident values
	for _, ident := range idents {
		obj, ok := defs[ident]
		if !ok {
			return fmt.Errorf("no value for constant '%s'", ident.Name)
		}
		value := obj.(*types.Const).Val()
		u64, ok := constant.Uint64Val(value)
		if !ok {
			return fmt.Errorf("value of constant '%s' is not an integer", ident.Name)
		}
		r.Names[ident.Name] = u64
	}
	return nil
}

// MaxID returns the maximum identifier value that this set contains
func (r ResourceIDSet) MaxID() (max uint64) {
	for _, v := range r.Names {
		if v > max {
			max = v
		}
	}
	return
}

// parsePackage parses all go files from the specified path. If the path is a
// directory it will be treated as a package.
func parsePackageFiles(fset *token.FileSet, path string) ([]*ast.File, error) {
	var files []*ast.File
	if isDirectory(path) {
		pkgs, err := parser.ParseDir(fset, path, nil, 0)
		if err != nil {
			return nil, err
		}
		if len(pkgs) > 1 {
			return nil, errors.New("path contains files of multiple packages")
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				files = append(files, file)
			}
		}
	} else {
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

// check runs type information check and exposes the const values
func check(fset *token.FileSet, files []*ast.File, defs map[*ast.Ident]types.Object) error {
	conf := &types.Config{
		IgnoreFuncBodies: true,
		FakeImportC:      true,
	}
	info := &types.Info{
		Defs: defs,
	}
	if _, err := conf.Check("", fset, files, info); err != nil {
		return err
	}
	return nil
}

// findVersionHashIdent finds the go AST ident for the VersionHash constant
func findVersionHashIdent(files []*ast.File) (*ast.Ident, error) {
	for _, file := range files {
		for _, node := range file.Decls {
			decl, ok := node.(*ast.GenDecl)
			if !ok {
				continue
			}
			if decl.Tok != token.CONST {
				continue
			}
			for _, spec := range decl.Specs {
				var typ string // Type of the current value
				// Will succeed as we already checked decl.Tok
				vspec := spec.(*ast.ValueSpec)
				if vspec.Type == nil && len(vspec.Values) > 0 {
					typ = ""
				}
				if vspec.Type != nil {
					ident, ok := vspec.Type.(*ast.Ident)
					if !ok {
						continue
					}
					typ = ident.Name
				}
				if typ != "" && typ != "string" {
					continue
				}
				// All unwanted types have been filtered out. Find the ident we're
				// looking for
				for _, name := range vspec.Names {
					if name.Name == "VersionHash" {
						return name, nil
					}
				}
			}
		}
	}
	// No match was found
	return nil, nil
}

// findResourceIDIdents finds all go AST idents that could represent a resource
// ID
func findResourceIDIdents(files []*ast.File) ([]*ast.Ident, error) {
	var idents []*ast.Ident
	for _, file := range files {
		for _, node := range file.Decls {
			decl, ok := node.(*ast.GenDecl)
			if !ok {
				continue
			}
			if decl.Tok != token.CONST {
				continue
			}
			// The name of the type of the constants we are declaring.
			typ := ""
			for _, spec := range decl.Specs {
				// Will succeed as we already checked decl.Tok
				vspec := spec.(*ast.ValueSpec)
				if vspec.Type == nil && len(vspec.Values) > 0 {
					typ = ""
					continue
				}
				if vspec.Type != nil {
					ident, ok := vspec.Type.(*ast.Ident)
					if !ok {
						continue
					}
					typ = ident.Name
				}
				if typ != "uint" {
					continue
				}
				// All unwanted types have been filtered out. Find matching idents.
				for _, name := range vspec.Names {
					if name.Name == "_" {
						continue
					}
					idents = append(idents, name)
				}
			}
		}
	}
	return idents, nil
}
