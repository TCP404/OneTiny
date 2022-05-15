package model

import "io/fs"

// FileStruction 记录文件的各种相关信息
//
// 例:
//		DeviceAbsPath: /home/boii/---File---/---CODE/Go/a.go
//		URLRelPath:    /---CODE/Go/a.go
//		Name:          a.go
//		Size:          64
//		Mode:          110100100
type FileStruction struct {
	Size          string
	Mode          fs.FileMode
	DeviceAbsPath string
	URLRelPath    string
	Name          string
}
