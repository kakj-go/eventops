/*
 * Copyright (c) 2021 Terminus, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
