package pipeline

import (
	"fmt"
)

type Dag []Node

type Node struct {
	Name  string   `yaml:"name,omitempty"`
	Needs []string `yaml:"needs,omitempty"`
}

func (d Dag) GetNextNodes(alias string) []Node {
	var nextNodes []Node
	for _, node := range d {
		var isNext bool
		for _, need := range node.Needs {
			if need == alias {
				isNext = true
				break
			}
		}
		if isNext {
			nextNodes = append(nextNodes, node)
		}
	}
	return nextNodes
}

func (d Dag) Check() error {
	if len(d) == 0 {
		return fmt.Errorf("pipeline dag can not empty")
	}

	for _, node := range d {
		if node.Name == "" {
			return fmt.Errorf("dag node name can not empty")
		}
		if len(node.Needs) == 0 {
			return fmt.Errorf("dag node name %v needs can not empty", node.Name)
		}
	}
	return nil
}
