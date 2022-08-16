package pipeline

type Resources struct {
	Limit   *Limit   `yaml:"limit,omitempty"`
	Request *Request `yaml:"request,omitempty"`
}

type Limit struct {
	Cpu string `yaml:"cpu,omitempty"`
	Mem string `yaml:"mem,omitempty"`
}

type Request struct {
	Cpu string `yaml:"cpu,omitempty"`
	Mem string `yaml:"mem,omitempty"`
}
