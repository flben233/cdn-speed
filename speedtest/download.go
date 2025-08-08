package speedtest

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func downloadWorker(url string, serverIp string, size int32, ch chan SpeedResult) {
	timeout := 30 * time.Second
	// 自定义 DialContext
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 120 * time.Second,
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			port := strings.Split(addr, ":")[1]
			addr = net.JoinHostPort(serverIp, port)
			return dialer.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	handleErr := func(err error) {
		log.Printf("Request failed: %v\n", err)
		ch <- SpeedResult{Err: err}
	}

	// 提前建立连接
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		handleErr(err)
		return
	}
	req.Header.Set("Connection", "keep-alive")
	if _, err = client.Do(req); err != nil {
		handleErr(err)
		return
	}

	// 发起请求
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		handleErr(err)
		return
	}
	start := time.Now()
	n, err := io.Copy(io.Discard, io.LimitReader(resp.Body, int64(size)))
	if err != nil {
		if err.(net.Error).Timeout() {
			ch <- SpeedResult{
				float32(n) / float32(time.Since(start).Seconds()),
				nil}
		} else {
			ch <- SpeedResult{
				float32(n) / float32(time.Since(start).Seconds()),
				err}
		}
		return
	}
	lapse := float32(time.Since(start).Seconds())
	ch <- SpeedResult{
		float32(min(int64(size), n)) / lapse,
		err}
	defer resp.Body.Close()
}

func DownloadMultiThread(url string, serverIp string) float32 {
	jobs := make([]chan SpeedResult, 8)
	for i := 0; i < 8; i++ {
		jobs[i] = make(chan SpeedResult, 1)
		go downloadWorker(url, serverIp, 16*1024*1024, jobs[i])
	}
	var result float32
	for i, job := range jobs {
		spd := <-job
		result += spd.Result
		if spd.Err != nil {
			log.Printf("Error occured while mt downloading from %s #%d: %v\n", serverIp, i, spd.Err)
		}
	}
	return result
}

func DownloadSingleThread(url string, serverIp string) float32 {
	job := make(chan SpeedResult, 1)
	go downloadWorker(url, serverIp, 64*1024*1024, job)
	spd := <-job
	if spd.Err != nil {
		log.Printf("Error occured while downloading from %s: %v\n", serverIp, spd.Err)
	}
	return spd.Result
}
