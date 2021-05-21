package util

import (
	"fmt"
	"net"
	"os"
)

// 生成首页HTML内容
func GenerateHTML(relPath []string, pathTitle string) string {
	head := `<!DOCTYPE html><html lang="zh-CN"><head><meta charset="UTF-8"><link rel="icon" href="data:image/ico;base64,aWNv"><title>Tiny Server</title></head><body><h1>Directory listing for ` + pathTitle + `</h1>`
	tail := `</body></html>`

	html := head
	for i, f := range relPath {
		link := fmt.Sprintf("<a href='%s'> %d - %s </a><br />", f, i, f) // <a href='example.go'> 1 - example.go </a><br />
		html += link
	}
	html += tail
	return html
}

func GetIP() (string, bool) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), true
			}
		}
	}
	return "", false
}
