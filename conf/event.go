package conf

type Event struct {
	Process Process `json:"process"`
}

type Process struct {
	BufferSize            int64 `yaml:"bufferSize"`
	WorkNum               int64 `yaml:"workNum"`
	CacheSize             int   `yaml:"cacheSize"`
	LoopLoadEventInterval int   `yaml:"loopLoadEventInterval"`
}
