package util

import (
	"fmt"
	"oneTiny/model"
	"strconv"
	"strings"
)

// GenerateHTML 生成首页HTML内容，包括 head、fileList、
func GenerateHTML(files []model.FileStruction, pathTitle string, IsAllowUpload bool) string {
	headHTML := head(pathTitle)
	if IsAllowUpload {
		headHTML += upload()
	}
	fileListHTML := fileList(files)
	tailHTML := `<br /><hr /></body></html>`
	return headHTML + fileListHTML + tailHTML
}

func head(pathTitle string) string {
	head := []string{
		`<!DOCTYPE html>`,
		`<html lang="zh-CN">`,
		`<head>`,
		`<meta charset="UTF-8">`,
		`<link rel="icon" href="data:image/ico;base64,aWNv">`,
		`<title>OneTiny Server</title>`,
		`</head>`,
		`<body>`,
		`<h1 style="display:inline;">Directory listing for `,
		pathTitle,
		`</h1><hr /><br />`,
	}
	return strings.Join(head, "")
}

func upload() string {
	form := []string{
		`<div style="position: absolute;right: 30px;top: 10px;font-size: 24px;width: 500px;border: 1px solid #000;padding: 7px;border-radius: 8px;">`,
		`<form action="/upload" method="post" enctype="multipart/form-data">`,
		`<input type="file" name="upload_file" style="width: 400px; float: left;">`,
		`<input type="submit" style="float: right;">`,
		`</form>`,
		`</div>`,
	}
	return strings.Join(form, "")
}

func fileList(files []model.FileStruction) string {
	tableHead := `<table style="width:100%">`
	tHead := `<thead><tr><td style='width:30%'>文件名</td><td style="width:100px;text-align:center;">文件大小</td><td></td></tr></thead>`
	tBody := "<tbody>"
	tBody += "<tr><td><a href='../'> &nbsp;. . /</a></td></tr>"
	for i, f := range files {
		link := ""
		trHead := "<tr>"
		fileLink := fmt.Sprintf("<td>%d. <a href='%s'> %s </a></td>", i, f.Rel, f.Name)
		fileSize := fmt.Sprintf("<td style='text-align:right'>%s</td>", sizeFmt(f.Size))
		trTail := "</tr>"
		link = trHead + fileLink + fileSize + trTail
		tBody += link
	}
	tBody += `</tbody>`
	tableTail := `</table>`
	return tableHead + tHead + tBody + tableTail
}

// sizeFmt 格式化给定的文件大小，如果文件类型为目录，则返回 "-"
//
// 参数：
//		bit int64: 文件大小的比特数
// 返回值：
// 		string: 格式化后的文件大小，单位 KByte、MByte、GByte、TByte、PByte
func sizeFmt(bit int64) string {
	const (
		_        = iota
		KB int64 = 1 << (10*iota + 3)
		MB
		GB
		TB
		PB
	)
	sizeFloat := float64(bit)
	size := "-"
	unit := "b"
	if bit == 0 {
		return size
	}
	switch {
	case bit < KB:
		return strconv.FormatInt(bit, 10) + "b"
	case bit >= KB && bit < MB:
		sizeFloat /= 1 << 10
		unit = "K"
	case bit >= MB && bit < GB:
		sizeFloat /= 1 << 20
		unit = "M"
	case bit >= GB && bit < TB:
		sizeFloat /= 1 << 30
		unit = "G"
	case bit >= TB && bit < PB:
		sizeFloat /= 1 << 40
		unit = "T"
	case bit >= PB:
		sizeFloat /= 1 << 50
		unit = "P"
	}
	return strconv.FormatFloat(sizeFloat, 'f', 2, 64) + unit
}
