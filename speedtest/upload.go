package speedtest

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type nullReader struct {
	total int64
	count int64
	head  []byte
}

func (r *nullReader) Read(p []byte) (n int, err error) {
	if r.count >= r.total {
		return 0, io.EOF
	}
	n = len(p)
	if int64(n) > r.total-r.count {
		n = int(r.total - r.count)
	}
	for i := 0; i < n; i++ {
		if r.count+int64(i) < int64(len(r.head)) {
			p[i] = r.head[r.count+int64(i)]
		} else {
			p[i] = 'A'
		}
	}
	r.count += int64(n)
	return n, nil
}

func uploadWorker(done context.Context, quit context.CancelFunc, serverIp string, serverPort int32, size int32, ch chan SpeedResult) {
	timeout := 30 * time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(serverIp, strconv.Itoa(int(serverPort))), timeout)
	if err != nil {
		ch <- SpeedResult{0, 0, err}
		return
	}
	_ = conn.SetDeadline(time.Now().Add(timeout))
	req := fmt.Sprintf(
		"GET %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Content-Type: application/json\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n"+
			"\r\n",
		"/", serverIp, size)
	go func() {
		start := time.Now()
		nn, err := io.Copy(conn, &nullReader{total: int64(size), head: []byte(req)})
		if err != nil {
			ch <- SpeedResult{float32(nn) / float32(time.Since(start).Seconds()), nn, err}
			return
		}
		elapse := float32(time.Since(start).Seconds())
		speed := float32(size) / elapse
		quit()
		ch <- SpeedResult{speed, nn, nil}
	}()
	// 一旦有线程完成，立刻停下
	<-done.Done()
	conn.Close()
}

func UploadMultiThread(serverIp string, serverPort int32) float32 {
	jobs := make([]chan SpeedResult, 8)
	var size int64
	sharedCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < 8; i++ {
		jobs[i] = make(chan SpeedResult, 1)
		go uploadWorker(sharedCtx, cancel, serverIp, serverPort, 32*1024*1024, jobs[i])
	}
	timer := time.After(60 * time.Second)
	start := time.Now()
	for i, job := range jobs {
		// 优先执行
		select {
		case spd := <-job:
			size += spd.Size
			continue
		default:
		}

		select {
		case spd := <-job:
			size += spd.Size
		case <-timer:
			log.Printf("Timeout while mt uploading to %s:%d #%d\n", serverIp, serverPort, i)
		}
	}
	elapse := float32(time.Since(start).Seconds())
	return float32(size) / elapse
}

func UploadSingleThread(serverIp string, serverPort int32) float32 {
	job := make(chan SpeedResult, 1)
	sharedCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go uploadWorker(sharedCtx, cancel, serverIp, serverPort, 64*1024*1024, job)
	select {
	case spd := <-job:
		if spd.Err != nil {
			log.Printf("Error occured while uploading to %s:%d: %v\n", serverIp, serverPort, spd.Err)
		}
		return spd.Result
	case <-time.After(60 * time.Second):
		log.Printf("Timeout while uploading to %s:%d\n", serverIp, serverPort)
		return 0.0
	}
}
