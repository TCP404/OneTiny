package util

import (
	"fmt"
	"strings"
)

// 生成首页HTML内容
func GenerateHTML(relPath []string, pathTitle string) string {
	headHTML := head(pathTitle)
	formHTML := upload()
	tailHTML := `<br /><hr /></body></html>`

	html := headHTML + formHTML
	for i, f := range relPath {
		link := fmt.Sprintf("<a href='%s'> %d. %s </a><br />", f, i, f) // <a href='example.go'> 1 - example.go </a><br />
		html += link
	}
	html += tailHTML
	return html
}

func head(pathTitle string) string {
	head := []string{
		`<!DOCTYPE html>`,
		`<html lang="zh-CN">`,
		`<head>`,
		`<meta charset="UTF-8">`,
		`<link rel="icon" href="data:image/ico;base64,aWNv">`,
		`<title>Tiny Server</title>`,
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
