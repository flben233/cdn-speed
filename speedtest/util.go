package speedtest

import (
	"context"
	"fmt"
	"time"
)

func testing(ctx context.Context, msg chan string) {
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
