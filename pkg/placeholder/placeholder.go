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

package placeholder

import (
	"eventops/apistructs"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var PhRe = regexp.MustCompile(`\${{[ ]{1}([^{}\s]+)[ ]{1}}}`)

const Left = "${{ "
const Right = " }}"

type Type string

func (t Type) String() string {
	return t
}

const (
	ContextType Type = "contexts"
	InputType   Type = "inputs"
	OutputType  Type = "outputs"
	RandomType  Type = "randoms"
)

type Handler func(placeholder string, values ...string) error

func MatchHolderFromHandler(needMatchString string, handlers map[Type]Handler) error {
	matchStrings := PhRe.FindAllString(needMatchString, -1)
	for _, placeholder := range matchStrings {

		value := strings.TrimPrefix(placeholder, Left)
		value = strings.TrimSuffix(value, Right)

		ss := strings.SplitN(value, ".", 2)

		split := strings.Split(value, ".")

		switch ss[0] {
		case ContextType.String():
			if handlers[ContextType] == nil {
				continue
			}
			if len(split) != 2 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.xxx }}", ContextType, placeholder, ContextType)
			}
			err := handlers[ContextType](placeholder, split[0], split[1])
			if err != nil {
				return err
			}
		case InputType.String():
			if handlers[InputType] == nil {
				continue
			}
			if len(split) != 2 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.xxx }}", InputType, placeholder, InputType)
			}

			err := handlers[InputType](placeholder, split[0], split[1])
			if err != nil {
				return err
			}
		case OutputType.String():
			if handlers[OutputType] == nil {
				continue
			}
			if len(split) != 3 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.alias.xxx }}", OutputType, placeholder, OutputType)
			}
			err := handlers[OutputType](placeholder, split[0], split[1], split[2])
			if err != nil {
				return err
			}
		case RandomType.String():
			if handlers[RandomType] == nil {
				continue
			}
			if len(split) != 3 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.alias.xxx }}", RandomType, placeholder, RandomType)
			}
			err := handlers[RandomType](placeholder, split[0], split[1])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type ReplaceValue struct {
	Inputs   apistructs.Inputs
	Outputs  apistructs.Outputs
	Contexts apistructs.Contexts

	PipelineId uint64
	TaskId     uint64
}

func MakeOutputKey(taskAlias string, taskOutputName string) string {
	return fmt.Sprintf("%v.%v", taskAlias, taskOutputName)
}

func MakeRealFilePath(pipelineId, taskId uint64, name string, typeName string) string {
	return path.Join("/root", "pipelines", strconv.FormatUint(pipelineId, 10), "tasks", strconv.FormatUint(taskId, 10), typeName, name)
}

func ReplacePlaceholder(needMatchString string, replaceValue *ReplaceValue, makeFileTypeValueRealPath bool) string {
	_ = MatchHolderFromHandler(needMatchString, map[Type]Handler{
		ContextType: func(placeholder string, values ...string) error {
			// ${{ contexts.xxx }}
			if replaceValue.Contexts == nil {
				return nil
			}

			contextName := values[1]
			context := replaceValue.Contexts[contextName]
			if context.Type == apistructs.EnvType {
				needMatchString = strings.ReplaceAll(needMatchString, placeholder, context.Value)
			} else {
				if makeFileTypeValueRealPath {
					needMatchString = strings.ReplaceAll(needMatchString, placeholder, MakeRealFilePath(replaceValue.PipelineId, replaceValue.TaskId, context.Name, ContextType.String()))
				} else {
					needMatchString = strings.ReplaceAll(needMatchString, placeholder, context.Value)
				}
			}
			return nil
		},
		InputType: func(placeholder string, values ...string) error {
			// ${{ inputs.xxx }}
			if replaceValue.Inputs == nil {
				return nil
			}

			inputName := values[1]
			input := replaceValue.Inputs[inputName]

			if input.Type == apistructs.EnvType {
				needMatchString = strings.ReplaceAll(needMatchString, placeholder, input.Value)
			} else {
				if makeFileTypeValueRealPath {
					needMatchString = strings.ReplaceAll(needMatchString, placeholder, MakeRealFilePath(replaceValue.PipelineId, replaceValue.TaskId, input.Name, InputType.String()))
				} else {
					needMatchString = strings.ReplaceAll(needMatchString, placeholder, input.Value)
				}
			}
			return nil
		},
		OutputType: func(placeholder string, values ...string) error {
			// ${{ outputs.alias.xxx }}
			if replaceValue.Outputs == nil {
				return nil
			}

			taskName := values[1]
			taskOutput := values[2]
			output := replaceValue.Outputs[MakeOutputKey(taskName, taskOutput)]

			if output.Type == apistructs.EnvType {
				needMatchString = strings.ReplaceAll(needMatchString, placeholder, output.Value)
			} else {
				if makeFileTypeValueRealPath {
					needMatchString = strings.ReplaceAll(needMatchString, placeholder, MakeRealFilePath(replaceValue.PipelineId, replaceValue.TaskId, output.Name, OutputType.String()))
				} else {
					needMatchString = strings.ReplaceAll(needMatchString, placeholder, output.Value)
				}
			}
			return nil
		},
		RandomType: func(placeholder string, values ...string) error {
			// todo
			return nil
		},
	})
	return needMatchString
}
