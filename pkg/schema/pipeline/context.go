package pipeline

import "fmt"

type Context struct {
	Name string    `yaml:"name,omitempty"`
	Type ValueType `yaml:"type,omitempty"`
}

func (c Context) check() error {
	if c.Name == "" {
		return fmt.Errorf("context name cannot empty")
	}

	if err := c.Type.ValueTypeCheck(); err != nil {
		return fmt.Errorf("context name %v type check error %v", c.Name, err)
	}
	return nil
}
