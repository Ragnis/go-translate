package main

import (
	"os"
)

func isDirectory(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir()
	}
	return false
}
