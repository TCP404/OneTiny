package model

import "io/fs"

// FileStruction 记录文件的各种相关信息
//
// 例:
//		Abs:  	/home/boii/---File---/---CODE/Go/a.go
//		RootPath: /home/boii/---File---
//		Rel:      /---CODE/Go/a.go
//		Name:   a.go
//		Size:   64
//		Mode:   110100100
type FileStruction struct {
	Size int64
	Mode fs.FileMode
	Abs  string
	Rel  string
	Name string
}
