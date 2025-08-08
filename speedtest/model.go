package speedtest

type SpeedResult struct {
	Result float32
	Err    error
}

type PingResult struct {
	AvgRtt float32
	Jitter float32
	Err    error
}

type TestServer struct {
	URL     string `json:"url"`
	Servers []struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
	} `json:"servers"`
}
