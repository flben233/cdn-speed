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

func downloadWorker(done context.Context, quit context.CancelFunc, url string, serverIp string, size int32, ch chan SpeedResult) {
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
	go func() {
		start := time.Now()
		n, err := io.Copy(io.Discard, io.LimitReader(resp.Body, int64(size)))
		if err != nil {
			ch <- SpeedResult{float32(n) / float32(time.Since(start).Seconds()), n, err}
			quit()
		}
		elapse := float32(time.Since(start).Seconds())
		ch <- SpeedResult{float32(min(int64(size), n)) / elapse, n, err}
		// 本线程完成，通知其他线程停下
		quit()
	}()
	// 一旦有线程完成，立刻停下
	<-done.Done()
	resp.Body.Close()
}

func DownloadMultiThread(url string, serverIp string) float32 {
	jobs := make([]chan SpeedResult, 8)
	sharedCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < 8; i++ {
		jobs[i] = make(chan SpeedResult, 1)
		go downloadWorker(sharedCtx, cancel, url, serverIp, 32*1024*1024, jobs[i])
	}
	start := time.Now()
	var size int64 = 0
	for i, job := range jobs {
		spd := <-job
		size += spd.Size
		if spd.Err != nil {
			log.Printf("Error occured while mt downloading from %s #%d: %v\n", serverIp, i, spd.Err)
		}
	}
	lapse := float32(time.Since(start).Seconds())
	return float32(size) / lapse
}

func DownloadSingleThread(url string, serverIp string) float32 {
	job := make(chan SpeedResult, 1)
	sharedCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go downloadWorker(sharedCtx, cancel, url, serverIp, 64*1024*1024, job)
	spd := <-job
	if spd.Err != nil {
		log.Printf("Error occured while downloading from %s: %v\n", serverIp, spd.Err)
	}
	return spd.Result
}
