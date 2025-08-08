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
	fmt.Print(banner + `
1.  中国大陆三网 IPv4 多线程测速
2.  中国大陆三网 IPv4 单线程测速
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
	}
}
