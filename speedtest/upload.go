package speedtest

import (
	"fmt"
	"io"
	"log"
	"net"
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

func uploadWorker(serverIp string, serverPort int32, size int32, ch chan SpeedResult) {
	timeout := 30 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort), timeout)
	defer conn.Close()
	if err != nil {
		ch <- SpeedResult{0, err}
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
	start := time.Now()
	if nn, err := io.Copy(conn, &nullReader{total: int64(size), head: []byte(req)}); err != nil {
		ch <- SpeedResult{
			float32(nn) / float32(time.Since(start).Seconds()),
			err}
		return
	}
	elapse := float32(time.Since(start).Seconds())
	speed := float32(size) / elapse
	ch <- SpeedResult{speed, nil}
}

func UploadMultiThread(serverIp string, serverPort int32) float32 {
	jobs := make([]chan SpeedResult, 8)
	for i := 0; i < 8; i++ {
		jobs[i] = make(chan SpeedResult, 1)
		go uploadWorker(serverIp, serverPort, 16*1024*1024, jobs[i])
	}
	var result float32
	for i, job := range jobs {
		select {
		case spd := <-job:
			result += spd.Result
			if spd.Err != nil {
				log.Printf("Error occured while mt uploading to %s:%d #%d: %v\n", serverIp, serverPort, i, spd.Err)
			}
		case <-time.After(60 * time.Second):
			result += 0
		}
	}
	return result
}

func UploadSingleThread(serverIp string, serverPort int32) float32 {
	job := make(chan SpeedResult, 1)
	go uploadWorker(serverIp, serverPort, 64*1024*1024, job)
	select {
	case spd := <-job:
		if spd.Err != nil {
			log.Printf("Error occured while uploading to %s:%d: %v\n", serverIp, serverPort, spd.Err)
		}
		return spd.Result
	case <-time.After(60 * time.Second):
		return 0.0
	}
}
