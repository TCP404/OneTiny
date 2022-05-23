package util

import (
	"fmt"
	"strings"

	"github.com/TCP404/OneTiny-cli/internal/model"
)

type generator struct {
	strings.Builder
}

// GenerateIndexHTML 生成首页HTML内容，包括 head、fileList、
func GenerateIndexHTML(files []model.FileStructure, pathTitle string, IsAllowUpload bool) string {
	gen := &generator{}
	gen.indexHead(pathTitle)
	if IsAllowUpload {
		gen.upload()
	}
	gen.fileList(files)
	gen.WriteString(`<br /><hr /></body></html>`)
	return gen.String()
}

func (g *generator) indexHead(pathTitle string) {
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
	g.WriteString(strings.Join(head, ""))
}

func (g *generator) upload() {
	form := []string{
		`<div style="position: absolute;right: 30px;top: 10px;font-size: 24px;width: 500px;border: 1px solid #000;padding: 7px;border-radius: 8px;">`,
		`<form action="/file/upload" method="post" enctype="multipart/form-data">`,
		`<input type="file" name="upload_file" style="width: 400px; float: left;">`,
		`<input type="submit" value="上传" style="float: right;">`,
		`</form>`,
		`</div>`,
	}
	g.WriteString(strings.Join(form, ""))
}

func (g *generator) fileList(files []model.FileStructure) {
	g.WriteString(`<table style="width:100%">`)
	g.WriteString(`<thead><tr><td style='width:30%'>文件名</td><td style="width:100px;text-align:right;">文件大小</td><td style="width:100px;text-align:right;">操作</td><td></td></tr></thead>`)
	g.WriteString(`<tbody>`)
	g.WriteString(`<tr><td><a href='../?action=view'> &nbsp;. . /</a></td></tr>`)
	for i, f := range files {
		g.WriteString("<tr>")
		fileLink := fmt.Sprintf("<td>%d. <a href='%s?action=view'> %s </a></td>", i, f.URLRelPath, f.Name)
		fileSize := fmt.Sprintf("<td style='text-align:right'>%s</td>", f.Size)
		g.WriteString(fileLink)
		g.WriteString(fileSize)

		if !f.Mode.IsDir() {
			g.WriteString(fmt.Sprintf("<td style='text-align:right'><a href='%s?action=dl' download='%s'>下载</a> <a href='%s?action=view' target='_blank'>查看</a></td>", f.URLRelPath, f.Name, f.URLRelPath))
		} else {
			g.WriteString(fmt.Sprintf("<td style='text-align:right'><a href='%s?action=dl' download='%s'>下载</a> <a href='%s?action=view'>&nbsp;&nbsp;-></a></td>", f.URLRelPath, f.Name, f.URLRelPath))
		}
		g.WriteString("</tr>")
	}
	g.WriteString(`</tbody>`)
	g.WriteString(`</table>`)
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
