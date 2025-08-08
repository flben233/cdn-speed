package speedtest

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Reset  = "\033[0m"
)

func SpeedTest(testServer TestServer, multiThread bool, banner string) {
	if multiThread {
		fmt.Println(string(banner) + `
大陆三网+教育网 IPv4 多线程测速，version 1
-----------------------------------------------------------------------------
Node            Download/Mbps      Upload/Mbps      Latency/ms      Jitter/ms`)
	} else {
		fmt.Println(string(banner) + `
大陆三网+教育网 IPv4 单线程测速，version 1
-----------------------------------------------------------------------------
Node            Download/Mbps      Upload/Mbps      Latency/ms      Jitter/ms`)
	}
	parts := strings.Split(testServer.URL, "://")
	protocol := parts[0]
	parts = strings.Split(parts[1], "/")
	var port int32 = 443
	if parts = strings.Split(parts[0], ":"); len(parts) > 1 {
		port1, _ := strconv.Atoi(parts[1])
		port = int32(port1)
	} else if protocol == "http" {
		port = 80
	}
	for _, server := range testServer.Servers {
		ctx, cancel := context.WithCancel(context.Background())
		testing(ctx)
		var down, up float32
		if multiThread {
			down = DownloadMultiThread(testServer.URL, server.IP)
			up = UploadMultiThread(server.IP, port)
		} else {
			down = DownloadSingleThread(testServer.URL, server.IP)
			up = UploadSingleThread(server.IP, port)
		}
		ping := Ping(server.IP)
		cancel()
		down, up = down*8/(1024*1024), up*8/(1024*1024)
		downStr, upStr := fmt.Sprintf("%.2f Mbps", down), fmt.Sprintf("%.2f Mbps", up)
		lat, jit := fmt.Sprintf("%.2f ms", ping.AvgRtt), fmt.Sprintf("%.2f ms", ping.Jitter)
		fmt.Printf("\r%s%-13s%s %s%-18s %-16s %-15s %-15s%s\n", Yellow, server.Name, Reset, Blue, downStr, upStr, lat, jit, Reset)
	}
	fmt.Println("-----------------------------------------------------------------------------")
	fmt.Println("系统时间：", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Println("北京时间：", time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"), "CST")
	fmt.Println("-----------------------------------------------------------------------------")
}
