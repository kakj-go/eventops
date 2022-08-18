package conf

type Event struct {
	Process Process `json:"process"`
}

type Process struct {
	BufferSize         int64 `yaml:"bufferSize"`
	WorkNum            int64 `yaml:"workNum"`
	ProcessingOverTime int64 `yaml:"processingOverTime"`

	TriggerCacheSize      int `yaml:"triggerCacheSize"`
	LoopLoadEventInterval int `yaml:"loopLoadEventInterval"`
}
