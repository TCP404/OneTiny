package util

import (
	"fmt"
	"oneTiny/core/model"
	"strconv"
	"strings"
)

// TODO 对字符串拼接进行优化，推荐采用 strings.Builder 并提前设置好 Grow()
// https://juejin.cn/post/6844903713241301006

// GenerateIndexHTML 生成首页HTML内容，包括 head、fileList、
func GenerateIndexHTML(files []model.FileStruction, pathTitle string, IsAllowUpload bool) string {
	headHTML := indexHead(pathTitle)
	if IsAllowUpload {
		headHTML += upload()
	}
	fileListHTML := fileList(files)
	tailHTML := `<br /><hr /></body></html>`
	return headHTML + fileListHTML + tailHTML
}

func indexHead(pathTitle string) string {
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
		`<form action="/file/upload" method="post" enctype="multipart/form-data">`,
		`<input type="file" name="upload_file" style="width: 400px; float: left;">`,
		`<input type="submit" value="上传" style="float: right;">`,
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
		return strconv.FormatInt(bit, 10) + unit
	case bit >= KB && bit < MB:
		sizeFloat /= 1 << 13
		unit = "K"
	case bit >= MB && bit < GB:
		sizeFloat /= 1 << 23
		unit = "M"
	case bit >= GB && bit < TB:
		sizeFloat /= 1 << 33
		unit = "G"
	case bit >= TB && bit < PB:
		sizeFloat /= 1 << 43
		unit = "T"
	case bit >= PB:
		sizeFloat /= 1 << 53
		unit = "P"
	}
	return strconv.FormatFloat(sizeFloat, 'f', 2, 64) + unit
}

func GenerateLoginHTML() string {
	return `<html lang="zh-CN">
	<head>
		<meta charset="utf-8">
		<title>OneTiny</title>
	<style>
	/* reset */
	body, h1, h2, h3, h4, h5, h6, p, ul, ol, li, form, input { margin: 0; padding: 0; }
	body { font-size: 14px; -webkit-font-smoothing: antialiased; }
	a { text-decoration: none; }
	
	/* ui */
	.ui-input { position: relative; padding: 15px 0; border-bottom: 1px solid #dfe6e5; font-size: 0; }
	.ui-input input { width: 100%; height: 30px; border: 0; font-size: 14px; outline: none; }
	
	.ui-button { height: 40px; border: 0; font-size: 14px; outline: none; cursor: pointer; }
	.ui-button--primary { color: #fff; background-color: #a6aaad; }
	.ui-button--success { color: #fff; background-color: #22d18e; }
	
	/* page */
	.form { width: 460px; margin: 0 auto; padding-top: 70px; }
	.form .captcha { height: 30px; vertical-align: top; cursor: pointer; }
	.form a { color: #7b7f81; }
	.form a:hover { color: #666; }
	
	.form-head { padding: 20px 0; text-align: center; }
	.form-head h2 { font-size: 24px; font-weight: 400; }
	.form-head p { margin-top: 12px; color: #7b7f81; }
	.form-head p a { text-decoration: underline; }
	
	.form-body { padding: 20px 40px; color: #222; }
	.form-body .err-msg { text-align: center; color: #fc5c5c; }
	
	.forget-password { margin-top: 10px; text-align: right; }
	.form .narrow-input input { width: 290px; margin-right: 10px; }
	.form .warn-msg { margin-bottom: 20px; }
	.form .err-msg + .warn-msg { margin-top: 12px; }
	.form .sms-button { display: inline-block; width: 80px; font-size: 14px; text-align: right; color: #22d18e; }
	.form .sms-button:hover { color: #56e9b2; }
	.form .form-notice { color: #22d18e; }
	.form .ui-input.focus { border-bottom-color: #22d18e; }
	.form .ui-button { width: 100%; margin: 40px 0; }
	</style>
	
	</head>
	<body>
	<div id="main">
		<div class="form">
			<div class="form-head">
				<h2>登录</h2>
			</div>
			<div class="form-body">
				<!-- <p class="err-msg">账号不存在</p> -->
				<div class="ui-input">
					<input name="username" id="username"　type="text" placeholder="帐号" autocomplete="off">
				</div>
				<div class="ui-input">
					<input name="password" id="password" type="password" placeholder="密码" autocomplete="off">
				</div>
				<button class="ui-button ui-button--primary" onclick="submit()">登录</button>
			</div>
		</div>
	</div>
	
	
	<script>
		var input = document.querySelectorAll('.ui-input input');
		input.forEach(function(val, i) { 
			val.onfocus = function() {
				this.parentNode.className += ' focus';
			}
			val.onblur = function() {
				this.parentNode.className = this.parentNode.className.replace(' focus', '');
			}
		});
		document.onkeydown=function(event) {
			if (event.keyCode == 13) submit();
		}
		function submit() {
			var formData = new FormData();
			formData.append("username", document.getElementById("username").value);
			formData.append("password", document.getElementById("password").value);
			var xhr = new XMLHttpRequest();
			xhr.timeout = 3000;
			xhr.onreadystatechange = callback;
			xhr.open('POST', '/login', true);
			xhr.send(formData);

			function callback() {
				if (xhr.readyState === 4 && xhr.status === 200) {
					var result = JSON.parse(xhr.responseText);	
					if(result.code === 0){alert(result.message);}
					else{window.location.href = "/file/";}
				} else {
					console.log(xhr.responseText);
				}
			}
		}
	</script>
	</body>
	</html>`
}
