package util

import "fmt"

// 前景 背景 颜色
// ---------------------------------------
// 30  40  黑色
// 31  41  红色
// 32  42  绿色
// 33  43  黄色
// 34  44  蓝色
// 35  45  紫红色
// 36  46  青蓝色
// 37  47  白色

const (
	F_BLACK  uint8 = 30 + iota // 30  黑色
	F_RED                      // 31  红色
	F_GREEN                    // 32  绿色
	F_YELLOW                   // 33  黄色
	F_BLUE                     // 34  蓝色
	F_PURPLE                   // 35  紫红色
	F_CYAN                     // 36  青蓝色
	F_WHITE                    // 37  白色
)

var front = map[uint8]string{
	30: "30",
	31: "31",
	32: "32",
	33: "33",
	34: "34",
	35: "35",
	36: "36",
	37: "37",
}

const (
	B_BLACK  uint8 = 40 + iota // 40  黑色
	B_RED                      // 41  红色
	B_GREEN                    // 42  绿色
	B_YELLOW                   // 43  黄色
	B_BLUE                     // 44  蓝色
	B_PURPLE                   // 45  紫红色
	B_CYAN                     // 46  青蓝色
	B_WHITE                    // 47  白色
)

var back = map[uint8]string{
	40: "40",
	41: "41",
	42: "42",
	43: "43",
	44: "44",
	45: "45",
	46: "46",
	47: "47",
}

// 代码 意义
// -------------------------
//  0  终端默认设置
//  1  高亮显示
//  4  使用下划线
//  5  闪烁
//  7  反白显示
//  8  不可见
const (
	DEFAULT   uint8 = 0 //  0  终端默认设置
	HIGHLIGHT           //  1  高亮显示
	UNDERLINE uint8 = 4 //  4  使用下划线
	BLINK               //  5  闪烁
	REWHITE   uint8 = 7 //  7  反白显示
	HIDDEN              //  8  不可见
)

var style = map[uint8]string{
	0: "0",
	1: "1",
	4: "4",
	5: "5",
	7: "7",
	8: "8",
}

const (
	end string = "\033[0m"
)

func GetF(f uint8) string {
	begin := fmt.Sprintf("\033[%sm", front[f])
	return begin
}

func GetFB(f, b uint8) string {
	begin := fmt.Sprintf("\033[%s;%sm", front[f], back[b])
	return begin
}

func GetSFB(s, f, b uint8) (begin string) {
	begin = fmt.Sprintf("\033[%s;%s;%sm", style[s], front[f], back[b])
	return
}

func GetEnd() string {
	return end
}
