package speedtest

import (
	"context"
	"fmt"
	"time"
)

func Testing(ctx context.Context, msg chan string) {
	s := ""
	go func() {
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case s = <-msg:
				s = fmt.Sprintf("\rTesting %-20s", s+"...")
			default:
				switch i % 4 {
				case 0:
					fmt.Printf("%s |", s)
				case 1:
					fmt.Printf("%s /", s)
				case 2:
					fmt.Printf("%s -", s)
				case 3:
					fmt.Printf("%s \\", s)
				}
				i++
			}
			time.Sleep(350 * time.Millisecond)
		}
	}()
}

func AutoPad(str string, length int) string {
	runeLen := 0
	for _, s := range str {
		// 全角ASCII、全角空格、汉字、平假名、片假名和CJK符号和标点
		if (s >= 0xFF01 && s <= 0xFF5E) || (s == 0x3000) || (s >= 0x4E00 && s <= 0x9FFF) || (s >= 0x3040 && s <= 0x309F) || (s >= 0x30A0 && s <= 0x30FF) || (s >= 0x3000 && s <= 0x303F) {
			runeLen += 2 // 全角字符占用两个位置
		} else {
			runeLen += 1 // 半角字符占用一个位置
		}
	}
	for i := runeLen; i < length; i++ {
		str += " "
	}
	return str
}
