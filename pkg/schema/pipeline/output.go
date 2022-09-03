package pipeline

import "eventops/apistructs"

type Output struct {
	Name         string               `yaml:"name,omitempty"`
	Value        string               `yaml:"value,omitempty"`
	Type         apistructs.ValueType `yaml:"type,omitempty"`
	SetToContext string               `yaml:"setToContext,omitempty"`
}
