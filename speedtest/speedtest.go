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

func SpeedTest(testServer TestServer, multiThread bool, banner string, ipv6 bool) {
	ipSymbol := "IPv4"
	if ipv6 {
		ipSymbol = "IPv6"
	}
	if multiThread {
		fmt.Printf(string(banner)+`
大陆三网+教育网 %s 多线程测速，version 1
-----------------------------------------------------------------------------
Node            Download/Mbps      Upload/Mbps      Latency/ms      Jitter/ms
`, ipSymbol)
	} else {
		fmt.Printf(string(banner)+`
大陆三网+教育网 %s 单线程测速，version 1
-----------------------------------------------------------------------------
Node            Download/Mbps      Upload/Mbps      Latency/ms      Jitter/ms
`, ipSymbol)
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
		ch := make(chan string, 1)
		Testing(ctx, ch)
		var down, up float32
		if multiThread {
			ch <- fmt.Sprintf("%s %s", server.Name, "Download")
			down = DownloadMultiThread(testServer.URL, server.IP)
			ch <- fmt.Sprintf("%s %s", server.Name, "Upload")
			up = UploadMultiThread(server.IP, port)
		} else {
			ch <- fmt.Sprintf("%s %s", server.Name, "Download")
			down = DownloadSingleThread(testServer.URL, server.IP)
			ch <- fmt.Sprintf("%s %s", server.Name, "Upload")
			up = UploadSingleThread(server.IP, port)
		}
		ch <- fmt.Sprintf("%s %s", server.Name, "Ping")
		ping := Ping(server.IP)
		cancel()
		down, up = down*8/(1024*1024), up*8/(1024*1024)
		downStr, upStr := fmt.Sprintf("%.2f Mbps", down), fmt.Sprintf("%.2f Mbps", up)
		lat, jit := fmt.Sprintf("%.2f ms", ping.AvgRtt), fmt.Sprintf("%.2f ms", ping.Jitter)
		fmt.Printf("\r%s%s%s %s%-18s %-16s %-15s %-9s%s\n", Yellow, AutoPad(server.Name, 15), Reset, Blue, downStr, upStr, lat, jit, Reset)
	}
	fmt.Println("-----------------------------------------------------------------------------")
	fmt.Println("系统时间：" + time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Println("北京时间: " + time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05") + " CST")
	fmt.Println("-----------------------------------------------------------------------------")
}
