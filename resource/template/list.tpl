<!DOCTYPE html>
<html lang="zh-CN">

<head>
    <meta charset="UTF-8">
    <link rel="icon" href="/favicon.ico">
    <title>OneTiny Server</title>
</head>

<body>
    <h1 style="display:inline;">Directory listing for {{- .pathTitle -}}</h1>
    <hr /><br />
    {{if .upload}}
    <div
        style="position: absolute;right: 30px;top: 10px;font-size: 24px;width: 500px;border: 1px solid #000;padding: 7px;border-radius: 8px;">
        <form action="/file/upload" method="post" enctype="multipart/form-data">
            <input type="file" name="upload_file" style="width: 400px; float: left;">
            <input type="hidden" name="path" value="{{- .pathTitle -}}">
            <input type="submit" value="上传" style="float: right;">
        </form>
    </div>
    {{end}}
    <table style="width:100%">
        <thead>
            <tr>
                <td style='width:30%'>文件名</td>
                <td style="width:100px;text-align:right;">文件大小</td>
                <td style="width:100px;text-align:right;">操作</td>
                <td></td>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td><a href='../'> &nbsp;. . /</a></td>
            </tr>
            {{range $i, $f := .files}}
            <tr>
                <td>{{$i}}.&nbsp;<a href='{{$f.URLRelPath}}?action=view'>{{$f.Name}}</a></td>
                <td style='text-align:right'>{{$f.Size}}</td>
                {{if $f.IsDir}}
                <td style='text-align:right'>
                    <a href='{{$f.URLRelPath}}?action=dl' download='{{$f.Name}}'>下载</a>
                    <a href='{{$f.URLRelPath}}?action=view'>查看</a>
                </td>
                {{else}}
                <td style='text-align:right'>
                    <a href='{{$f.URLRelPath}}?action=dl' download='{{$f.Name}}'>下载</a>
                    <a href='{{$f.URLRelPath}}?action=view' target='_blank'>&nbsp;->&nbsp;</a>
                </td>
                {{end}}
            </tr>
            {{end}}
        </tbody>
    </table>
</body>

</html>