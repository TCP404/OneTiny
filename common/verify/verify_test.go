package verify

import (
	"strings"
	"testing"
)

func TestVerifyPort(t *testing.T) {
	type test struct {
		iPort   int
		wantStr string
	}
	tests := map[string]test{
		"normal":           {9532, ""},
		"low bound":        {1024, ""},
		"high bound":       {65535, ""},
		"low bound sub 1":  {1023, "不可以设置系统预留端口 1023"},
		"high bound sub 1": {65536, "不可以设置系统预留端口 65536"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := verifyPort(tc.iPort)
			if !ErrorCheck(got, tc.wantStr) {
				t.Errorf("want %#v, got: %#v", tc.wantStr, got)
			}
		})
	}
}

func TestVerifyPath(t *testing.T) {
	tests := map[string]struct {
		path    string
		wantStr string
	}{
		"normal": {"/home/boii", ""},
		"not":    {"/home/abc", "无法设置您指定的共享路径，"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := verifyPath(tc.path)
			if !ErrorCheck(got, tc.wantStr) {
				t.Errorf("want: %#v, got: %#v", tc.wantStr, got)
			}
		})
	}
}

func ErrorCheck(got error, wantStr string) bool {
	if got == nil { // err 为空时，说明程序没有发生错误，errStr 应该为空，检查测试用例中的wantStr 是否为空
		return wantStr == ""
	}
	// err 不为空时，检查测试用例中的 wantStr 是否为空
	if wantStr == "" {
		return false
	}
	// err 不为空，测试用例中的 wantStr 也不为空，则检查 err 中的报错信息是否包含测试用例指定的报错信息
	return strings.Contains(got.Error(), wantStr)
}
