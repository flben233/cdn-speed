package speedtest

import (
	"context"
	"fmt"
	"time"
)

func testing(ctx context.Context) {
	go func() {
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				switch i % 4 {
				case 0:
					fmt.Printf("\rTesting... |")
				case 1:
					fmt.Printf("\rTesting... /")
				case 2:
					fmt.Printf("\rTesting... -")
				case 3:
					fmt.Printf("\rTesting... \\")
				}
				i++
			}
			time.Sleep(350 * time.Millisecond)
		}
	}()
}
