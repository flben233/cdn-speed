package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/shirakawatyu/cdn-speed/speedtest"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

//go:embed conf/nodes.json
var nodesJson []byte

//go:embed conf/nodes_v6.json
var nodesV6Json []byte

//go:embed conf/banner.txt
var banner string

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	file, err := os.OpenFile("cdn-speed.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0755)
	// 从头写入
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.SetOutput(file)

	nodes := speedtest.TestServer{}
	err = json.Unmarshal(nodesJson, &nodes)
	if err != nil {
		panic(err)
	}
	nodesV6 := speedtest.TestServer{}
	err = json.Unmarshal(nodesV6Json, &nodesV6)
	if err != nil {
		panic(err)
	}
	fmt.Print(banner + `
1.  中国大陆三网 IPv4 多线程测速
2.  中国大陆三网 IPv4 单线程测速
3.  中国大陆三网 IPv6 多线程测速
4.  中国大陆三网 IPv6 单线程测速
5.  退出
请输入您想选择的节点序号: `)
	var choice int
	_, err = fmt.Scanf("%d", &choice)
	if err != nil {
		panic(err)
	}
	fmt.Println()
	switch choice {
	case 1:
		speedtest.SpeedTest(nodes, true, banner)
	case 2:
		speedtest.SpeedTest(nodes, false, banner)
	case 3:
		speedtest.SpeedTest(nodesV6, true, banner)
	case 4:
		speedtest.SpeedTest(nodesV6, false, banner)
	case 5:
		fmt.Println("退出程序")
		os.Exit(0)
	default:
		fmt.Println("无效输入，退出程序")
		os.Exit(1)
	}
}
