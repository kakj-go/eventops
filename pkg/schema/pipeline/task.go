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
	"strings"
)

const ImageCreaterNameSplitWord = "/"
const ImageNameVersionSplitWord = ":"

type Task struct {
	Image            string              `yaml:"image,omitempty"`
	Alias            string              `yaml:"alias,omitempty"`
	Type             apistructs.TaskType `yaml:"type,omitempty"`
	ActuatorSelector ActuatorSelector    `yaml:"actuatorSelector,omitempty"`
	Inputs           []Input             `yaml:"inputs,omitempty"`
	Commands         []string            `yaml:"commands,omitempty"`
	Outputs          []Output            `yaml:"outputs,omitempty"`
	Timeout          int64               `yaml:"timeout,omitempty"`
	Resources        *Resources          `yaml:"resources,omitempty"`
}

func (t Task) GetPipelineVersion() string {
	if t.Type != apistructs.PipeType {
		return ""
	}

	return GetImageVersion(t.Image)
}

func GetImageVersion(image string) string {
	split := strings.SplitN(image, ImageNameVersionSplitWord, 2)
	if len(split) == 1 {
		return ""
	}

	return split[1]
}

func (t Task) GetPipelineCreater() string {
	if t.Type != apistructs.PipeType {
		return ""
	}

	return GetImageCreater(t.Image)
}

func GetImageCreater(image string) string {
	split := strings.SplitN(image, ImageCreaterNameSplitWord, 2)
	if len(split) == 1 {
		return ""
	}

	return split[0]
}

func GetImageName(image string) string {
	var name = image

	version := GetImageVersion(image)
	creater := GetImageCreater(image)

	if version != "" {
		name = strings.TrimSuffix(name, ImageNameVersionSplitWord+version)
	}

	if creater != "" {
		name = strings.TrimPrefix(name, creater+ImageCreaterNameSplitWord)
	}
	return name
}

func (t Task) GetPipelineName() string {
	if t.Type != apistructs.PipeType {
		return ""
	}

	return GetImageName(t.Image)
}

func (t Task) Check(pipelineContexts []Context) error {
	if t.Alias == "" {
		return fmt.Errorf("task alias can not empty")
	}

	if !t.Type.Check() {
		return fmt.Errorf("use [%s %s %s %s] these task type", apistructs.K8sType, apistructs.DockerType, apistructs.OsType, apistructs.PipeType)
	}

	if t.Type != apistructs.OsType && len(t.Image) == 0 {
		return fmt.Errorf("[%s, %s, %s] type task image can not empty", apistructs.K8sType, apistructs.DockerType, apistructs.PipeType)
	}

	if err := t.ActuatorSelector.check(); err != nil {
		return err
	}

	if err := t.inputCheck(); err != nil {
		return err
	}

	if t.Type != apistructs.PipeType && len(t.Commands) == 0 {
		return fmt.Errorf("[%s, %s, %s] task type, commands can not empty", apistructs.K8sType, apistructs.DockerType, apistructs.OsType)
	}

	if err := t.outputCheck(pipelineContexts); err != nil {
		return err
	}

	if err := t.resourceCheck(); err != nil {
		return err
	}

	return nil
}

func (t Task) outputCheck(pipelineContexts []Context) error {
	if t.Outputs == nil {
		return nil
	}

	for _, output := range t.Outputs {
		if output.Name == "" {
			return fmt.Errorf("output name can not empty")
		}

		if output.Value == "" {
			return fmt.Errorf("output name %v value can not empty", output.Name)
		}

		if t.Type != apistructs.PipeType {
			err := output.Type.ValueTypeCheck()
			if err != nil {
				return fmt.Errorf("output name %v type check error %v", output.Name, err)
			}
		}

		if output.SetToContext != "" {
			var find = false
			for _, context := range pipelineContexts {
				if context.Name != output.SetToContext {
					continue
				}
				if context.Type != output.Type {
					return fmt.Errorf("output name %v setToContext %v type not match", output.Name, output.SetToContext)
				}
				find = true
				break
			}
			if !find {
				return fmt.Errorf("output name %v setToContext %v not find in pipeline content", output.Name, output.SetToContext)
			}
		}
	}
	return nil
}

func (t Task) inputCheck() error {
	if t.Inputs == nil {
		return nil
	}

	for _, input := range t.Inputs {
		if input.Name == "" {
			return fmt.Errorf("input name can not empty")
		}
		if input.Value == "" {
			return fmt.Errorf("input name %v value can not empty", input.Name)
		}
	}
	return nil
}

func (t Task) resourceCheck() error {
	if t.Resources == nil {
		return nil
	}

	if t.Resources.Limit != nil {
		if t.Resources.Limit.Cpu == "" {
			return fmt.Errorf("resources limit cpu can not empty")
		}
		if t.Resources.Limit.Mem == "" {
			return fmt.Errorf("resources limit cpu can not empty")
		}
	}

	if t.Resources.Request != nil {
		if t.Resources.Request.Cpu == "" {
			return fmt.Errorf("resources request cpu can not empty")
		}
		if t.Resources.Request.Mem == "" {
			return fmt.Errorf("resources request cpu can not empty")
		}
	}

	return nil
}
