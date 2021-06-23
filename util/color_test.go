package util

import (
	"testing"
)

type test struct {
	want   string
	format string
	colors []uint8
	a      []interface{}
}

func TestRenderf(t *testing.T) {

	tests := map[string]test{
		"format_nil": {want: "\033[34mhttps://\033[0m", colors: []uint8{F_BLUE}, format: `https://`, a: nil},
		"getF":       {want: "\033[34mhttps://192.168.1.229:9090\033[0m", colors: []uint8{F_BLUE}, format: "https://%s:%s", a: []interface{}{"192.168.1.229", "9090"}},
		"getFB":      {want: "\033[34;46mhttps://192.168.1.229:9090\033[0m", colors: []uint8{F_BLUE, B_CYAN}, format: "https://%s:%s", a: []interface{}{"192.168.1.229", "9090"}},
		"getFB2":     {want: "\033[34;47mhttps://192.168.1.229:9090\033[0m", colors: []uint8{F_BLUE, B_WHITE}, format: "https://%s:%s", a: []interface{}{"192.168.1.229", "9090"}},
		"getSFB":     {want: "\033[1;34;47mhttps://192.168.1.229:9090\033[0m", colors: []uint8{HIGHLIGHT, F_BLUE, B_WHITE}, format: "https://%s:%s", a: []interface{}{"192.168.1.229", "9090"}},
		"errorIP": {
			want:   "\033[33m暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。\033[0m",
			colors: []uint8{F_YELLOW},
			format: "暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。",
			a:      nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := Renderf(tc.colors, tc.format, tc.a...)
			if got != tc.want {
				t.Errorf("want: %v  got: %v", tc.want, got)
			}
		})
	}
}

func BenchmarkRenderf(b *testing.B) {
	tc := test{want: "\033[1;34;47mhttps://192.168.1.229:9090\033[0m", colors: []uint8{HIGHLIGHT, F_BLUE, B_WHITE}, format: `https://%s:%s`, a: []interface{}{"192.168.1.229", "9090"}}
	for i := 0; i < b.N; i++ {
		Renderf(tc.colors, tc.format, tc.a)
	}
}
