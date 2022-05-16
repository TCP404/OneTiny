package common

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
)

func RandString(n int) string {
	var letters = []byte("!qaz@wsQWERTx#edYUIOPc$rfvtASDFGgb^yHJKLhnZXCVB&uNMjm*ik,(ol.)p;/_['+]")
	result := make([]byte, n)
	lettersLen := int64(len(letters))
	rand.Seed(time.Now().Unix())
	for i := range result {
		result[i] = letters[rand.Int63()%lettersLen]
	}
	return string(result)
}

func MD5(content string) string {
	h := md5.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
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

const (
	_        = iota
	KB int64 = 1 << (10*iota + 3)
	MB
	GB
	TB
	PB
)

// sizeFmt 格式化给定的文件大小，如果文件类型为目录，则返回 "-"
//
// 参数：
//		bit int64: 文件大小的比特数
// 返回值：
// 		string: 格式化后的文件大小，单位 KByte、MByte、GByte、TByte、PByte
func SizeFmt(bit int64) string {
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
		unit = "Kb"
	case bit >= MB && bit < GB:
		sizeFloat /= 1 << 23
		unit = "Mb"
	case bit >= GB && bit < TB:
		sizeFloat /= 1 << 33
		unit = "Gb"
	case bit >= TB && bit < PB:
		sizeFloat /= 1 << 43
		unit = "Tb"
	case bit >= PB:
		sizeFloat /= 1 << 53
		unit = "Pb"
	}
	return strconv.FormatFloat(sizeFloat, 'f', 2, 64) + unit
}

func DownloadBinary(url, path string) []error {
	_, body, errs := gorequest.New().Get(url).EndBytes()
	if len(errs) != 0 {
		return errs
	}
	err := ioutil.WriteFile(path, body, 0755)
	if err != nil {
		return []error{err}
	}
	return nil
}
