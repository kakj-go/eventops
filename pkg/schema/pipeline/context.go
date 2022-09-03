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
	"eventops/apistructs"
	"fmt"
)

type Context struct {
	Name string               `yaml:"name,omitempty"`
	Type apistructs.ValueType `yaml:"type,omitempty"`
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
