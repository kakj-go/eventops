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

package event

import (
	"eventops/apistructs"
	"eventops/pkg/schema/pipeline"
	"fmt"
)

type TriggerPipeline struct {
	Image   string        `yaml:"image,omitempty"`
	Inputs  []InputsValue `yaml:"inputs,omitempty"`
	Filters []Filter      `yaml:"filters,omitempty"`
}

type Trigger struct {
	Name         string            `yaml:"name,omitempty"`
	EventCreater string            `yaml:"eventCreater,omitempty"`
	EventName    string            `yaml:"eventName,omitempty"`
	EventVersion string            `yaml:"eventVersion,omitempty"`
	Pipelines    []TriggerPipeline `yaml:"pipelines,omitempty"`
	Filters      []Filter          `yaml:"filters,omitempty"`
}

type InputsValue struct {
	Name  string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (input InputsValue) check() error {
	if input.Name == "" {
		return fmt.Errorf("trigger definition inputs field: name can not empty")
	}
	if input.Value == "" {
		return fmt.Errorf("trigger definition inputs field: value can not empty")
	}
	return nil
}

type Filter struct {
	Expr    string   `yaml:"expr,omitempty"`
	Matches []string `yaml:"matches,omitempty"`
}

func (filter Filter) check() error {
	if filter.Expr == "" {
		return fmt.Errorf("trigger definition filters field: expr can not empty")
	}
	if len(filter.Matches) == 0 {
		return fmt.Errorf("trigger definition filters field: matches can not empty")
	}
	return nil
}

func (t *Trigger) Mutating(creater string) error {
	for index, pipe := range t.Pipelines {

		imageCreater := pipeline.GetImageCreater(pipe.Image)
		if imageCreater == "" {
			t.Pipelines[index].Image = fmt.Sprintf("%s%s%s", creater, pipeline.ImageCreaterNameSplitWord, t.Pipelines[index].Image)
		}

		imageVersion := pipeline.GetImageVersion(pipe.Image)
		if imageVersion == "" {
			t.Pipelines[index].Image = fmt.Sprintf("%s%s%s", t.Pipelines[index].Image, pipeline.ImageNameVersionSplitWord, apistructs.LatestVersion)
		}
	}

	return nil
}

func (t *Trigger) Check(creater string) error {
	if t.Name == "" {
		return fmt.Errorf("trigger definition field: name can not empty")
	}
	if t.EventCreater == "" {
		return fmt.Errorf("trigger definition field: eventCreater can not empty")
	}
	if t.EventName == "" {
		return fmt.Errorf("trigger definition field: eventName can not empty")
	}
	if t.EventVersion == "" {
		return fmt.Errorf("trigger definition field: eventVersion can not empty")
	}
	if len(t.Pipelines) == 0 {
		return fmt.Errorf("trigger definition field: pipelines can not empty")
	}
	for _, pipe := range t.Pipelines {
		if pipe.Image == "" {
			return fmt.Errorf("trigger definition pipeline field: image can not empty")
		}
		if pipeline.GetImageCreater(pipe.Image) != creater {
			return fmt.Errorf("trigger definition pipeline field: image user should use youself")
		}

		for _, input := range pipe.Inputs {
			if err := input.check(); err != nil {
				return err
			}
		}

		for _, filter := range pipe.Filters {
			if err := filter.check(); err != nil {
				return err
			}
		}
	}

	for _, filter := range t.Filters {
		if err := filter.check(); err != nil {
			return err
		}
	}

	return nil
}
