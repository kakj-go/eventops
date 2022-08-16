package pipeline

type Output struct {
	Name         string    `yaml:"name,omitempty"`
	Value        string    `yaml:"value,omitempty"`
	Type         ValueType `yaml:"type,omitempty"`
	SetToContext string    `yaml:"setToContext,omitempty"`
}
