package pipeline

import (
	"eventops/apistructs"
	"fmt"
)

type Input struct {
	Name    string               `yaml:"name,omitempty"`
	Value   string               `yaml:"value,omitempty"`
	Type    apistructs.ValueType `yaml:"type,omitempty"`
	Default string               `yaml:"default,omitempty"`
}

func (i Input) check() error {
	if i.Name == "" {
		return fmt.Errorf("input name can not empty")
	}
	if err := i.Type.ValueTypeCheck(); err != nil {
		return err
	}

	return nil
}
