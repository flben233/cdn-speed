package speedtest

import (
	"context"
	"fmt"
	"net"
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
			default:
				switch i % 4 {
				case 0:
					fmt.Printf("\rTesting %s... |", s)
				case 1:
					fmt.Printf("\rTesting %s... /", s)
				case 2:
					fmt.Printf("\rTesting %s... -", s)
				case 3:
					fmt.Printf("\rTesting %s... \\", s)
				}
				i++
			}
			time.Sleep(350 * time.Millisecond)
		}
	}()
}

func ForceExitIfTimeout(conn net.Conn, timeout time.Duration) {
	go func() {
		time.Sleep(timeout)
		_ = conn.Close()
	}()
}
