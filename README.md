# TinyServer

TinyServer 是一个用于局域网内共享文件的微型程序，它能将当前工作目录临时共享目录，对局域网内其他主机共享，通过浏览器访问 `http://局域网IP:9090` 来访问和下载共享目录中的文件。

简而言之与命令 `python -m http.server 9090` 做的是同样的事情。

## 需求

我有两台设备，一台装着 Linux 系统，一台装着 Windows 系统，偶尔需要互相传输文件。

在 Linux 上我可以在任意一个目录下使用命令 `python -m http.server 9090`，从而在 Windows 上或局域网内其他主机上通过浏览器访问 `http://局域网IP:9090` 查看所有文件，也可以下载;

但是这条命令在 Windows 上不可行，所以需要编写一个程序可以运行在 Windows 上实现同样的功能。

## 开发技术
Python 太久没碰了也懒得复习，最近又在学 [Golang](https://golang.org) 和 [gin](https://gin-gonic.com/zh-cn/) 框架，且 Golang 可以编译出不依赖于虚拟机的独立的可执行文件，也可以交叉编译，所以采用了 gin 做一个简单的局域网微型服务器。

## 使用说明
可从本仓库的 [Release]() 中下载对应版本。已提供 [Linux 版]() 和 [Windows 版]()，Mac 的同学请下载后自行编译。

### 安装
```bash
$ git clone https://github.com/TCP404/TinyServer.git
$ go mod tidy
$ go build
```

### 运行
**Windows**: 
下载后双击 TinyServer.exe 即可运行（需管理员权限）。
可以在CMD中切换到 TinyServer.exe 锁在目录，执行一下命令：
```cmd
> TinyServer                              # 将运行在 http://本机局域网IP:9090，共享目录为当前工作目录
> TinyServer.exe                          # 将运行在 http://本机局域网IP:9090，共享目录为当前工作目录
> TinyServer -p {指定端口}                 # 将运行在 http://本机局域网IP:指定端口，共享目录为当前工作目录
$ TinyServer -r {指定目录}                 # 将运行在 http://本机局域网IP:9090，共享目录为指定目录
$ TinyServer -r {指定目录} -p {指定端口}    # 将运行在 http://本机局域网IP:指定端口，共享目录为指定目录
```

**Linux**: 
```bash
$ cd {TinyServer所在目录}
$ ./TinyServer                            # 将运行在 http://本机局域网IP:9090，共享目录为当前工作目录
$ ./TinyServer -p {指定端口}               # 将运行在 http://本机局域网IP:指定端口，共享目录为当前工作目录
$ ./TinyServer -r {指定目录}               # 将运行在 http://本机局域网IP:9090，共享目录为指定目录
$ ./TinyServer -r {指定目录} -p {指定端口}  # 将运行在 http://本机局域网IP:指定端口，共享目录为指定目录
```


## 程序说明
程序非常简单。

首先获得一个 gin 实例，注册一个404路由处理 [404 错误](https://en.wikipedia.org/wiki/HTTP_404)

然后再注册一个普通路由，使用通配符以便读取所有链接。

因为是局域网内，且主要是我自己两台设备之间互传文件，所以就使用简单的 GET 请求。（此处可改进）。

浏览器访问 `http://本机局域网IP:指定端口号` 时，将当前工作目录视为根目录，以链接形式展示根目录下所有文件及子目录;

点击链接将重新发起新的请求，解析请求中的路径，判断是否为目录;

是 -> 读取目录下所有文件及子目录，生成链接，返回给客户端
否 -> 调用下载函数，将文件内容返回给客户端完成下载。

```go
func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.NoRoute(controller.NotFound)
	r.GET("/*filename", controller.Handler)

	r.Run(":" + controller.Port)
}
```
![程序流程图](README/Flowchart.png)

## TODO
- [ ] 上传功能
- [ ] 密码验证功能（防止局域网内监听）
- [ ] 增加图形界面（使用 [fyne](https://fyne.io/))
