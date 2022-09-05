/*
 * Copyright 2022 The kakj-go Authors.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
