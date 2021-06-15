package util

import (
	"fmt"
	"strconv"
	"strings"
	"oneTiny/model"
)

// 生成首页HTML内容
func GenerateHTML(files []model.FileStruction, pathTitle string) string {
	headHTML := head(pathTitle)
	formHTML := upload()
	fileListHTML := fileList(files)
	tailHTML := `<br /><hr /></body></html>`
	return headHTML + formHTML + fileListHTML + tailHTML
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
		fileLink := fmt.Sprintf("<td><a href='%s'> %d. %s </a></td>", f.Rel, i, f.Name)
		fileSize := fmt.Sprintf("<td style='text-align:right'>%s</td>", sizeFmt(f.Size))
		trTail := "</tr>"
		link = trHead + fileLink + fileSize + trTail
		tBody += link
	}
	tBody += `</tbody>`
	tableTail := `</table>`
	return tableHead + tHead + tBody + tableTail
}

func sizeFmt(bit int64) string {
	const (
		_ = iota
		K = 1 << (10 * iota)
		M
		G
	)
	size := "-"
	sizeFloat := float64(bit)
	unit := "b"
	if bit == 0 {
		return size
	}
	switch {
	case bit < K:
		return strconv.FormatInt(bit, 10) + "b"
	case bit >= K && bit < M:
		sizeFloat /= 1 << 10
		unit = "k"
	case bit >= M && bit < G:
		sizeFloat /= 1 << 20
		unit = "M"
	case bit >= G:
		sizeFloat /= 1 << 30
		unit = "G"
	}
	return strconv.FormatFloat(sizeFloat, 'f', 2, 64) + unit
}
