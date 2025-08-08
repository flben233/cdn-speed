package speedtest

import (
	probing "github.com/prometheus-community/pro-bing"
	"log"
	"math"
	"time"
)

func Ping(ip string) PingResult {
	pinger, err := probing.NewPinger(ip)
	if err != nil {
		log.Printf("Error creating pinger for %s: %v\n", ip, err)
		return PingResult{Err: err}
	}
	pinger.SetPrivileged(true)
	pinger.Count = 3
	pinger.Timeout = 5 * time.Second
	err = pinger.Run()
	if err != nil {
		log.Printf("Error running pinger for %s: %v\n", ip, err)
		return PingResult{Err: err}
	}
	stats := pinger.Statistics()
	jitter := 0.0
	for i := 0; i < len(stats.Rtts)-1; i++ {
		jitter += math.Abs(float64(stats.Rtts[i+1].Microseconds()-stats.Rtts[i].Microseconds())/1000.0) / float64(pinger.Count)
	}
	return PingResult{
		AvgRtt: float32(float64(stats.AvgRtt.Microseconds()) / 1000.0),
		Jitter: float32(jitter)}
}
