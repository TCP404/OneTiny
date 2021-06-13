package model

import "io/fs"

// RootPath: /home/boii/---File---
type FileStruction struct {
	Abs  string      // /home/boii/---File---/---CODE/Go/a.go
	Rel  string      // /---CODE/Go/a.go
	Name string      // a.go
	Size int64       // 64
	Mode fs.FileMode //
}
