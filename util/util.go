package util

import (
	"errors"
	"fmt"
	"net"
)

// 生成首页HTML内容
func GenerateHTML(relPath []string, pathTitle string) string {
	headHTML := `<!DOCTYPE html><html lang="zh-CN"><head><meta charset="UTF-8"><link rel="icon" href="data:image/ico;base64,aWNv"><title>Tiny Server</title></head><body><h1>Directory listing for ` + pathTitle + `</h1><hr /><br />`
	tailHTML := `<br /><hr /></body></html>`

	html := headHTML
	for i, f := range relPath {
		link := fmt.Sprintf("<a href='%s'> %d - %s </a><br />", f, i, f) // <a href='example.go'> 1 - example.go </a><br />
		html += link
	}
	html += tailHTML
	return html
}

func GetIP() (string, error) {
	ip, err := externalIP()
	return ip.String(), err
}

func externalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("你好像没有接入局域网？")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}
