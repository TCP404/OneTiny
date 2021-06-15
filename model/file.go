package model

import "io/fs"

// For example
// RootPath: /home/boii/---File---
// Abs:  /home/boii/---File---/---CODE/Go/a.go
// Rel:  /---CODE/Go/a.go
// Name: a.go
// Size: 64
// Mode: 110100100
type FileStruction struct {
	Size int64
	Mode fs.FileMode
	Abs  string
	Rel  string
	Name string
}
