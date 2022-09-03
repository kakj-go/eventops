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

package flowmanager

import (
	"eventops/apistructs"
	"eventops/pkg/placeholder"
	"eventops/pkg/schema/event"
	"fmt"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

func (node *Node) getPlaceholderInputsValue() (apistructs.Inputs, error) {
	parentNode := node.flow.getNode(node.parentTaskId)
	if parentNode == nil {
		return nil, fmt.Errorf("task alias: %v, not find parentNode by taskId: %v", node.getTask().Alias, node.parentTaskId)
	}

	var inputs = apistructs.Inputs{}
	if parentNode == node.flow.rootNode {
		extra := node.flow.getPipeExtra()
		dbTriggerDefinition := extra.TriggerDefinitionContent
		dbEvent := extra.EventContent

		var triggerDefinition event.Trigger
		err := yaml.Unmarshal([]byte(dbTriggerDefinition.Content), &triggerDefinition)
		if err != nil {
			return nil, fmt.Errorf("task alias: %v parent_task_id: %v yaml unmarshal triggerDefinition error: %v", node.getTask().Alias, node.parentTaskId, err.Error())
		}

		var triggerDefinitionPipe event.TriggerPipeline
		for _, pipe := range triggerDefinition.Pipelines {
			if pipe.Image == node.flow.rootNode.image {
				triggerDefinitionPipe = pipe
			}
		}

		var triggerDefinitionPipeInputMap = make(map[string]string, len(triggerDefinitionPipe.Inputs))
		for _, input := range triggerDefinitionPipe.Inputs {
			inputValue := gjson.Get(dbEvent.Content, input.Value).String()
			triggerDefinitionPipeInputMap[input.Name] = inputValue
		}

		definition, err := node.flow.getAndSetPipelineVersionDefinition(node.flow.rootNode.image)
		if err != nil {
			return nil, fmt.Errorf("task alias: %v parent_task_id: %v getAndSetPipelineVersionDefinition image: %v error: %v", node.getTask().Alias, node.parentTaskId, node.flow.rootNode.image, err.Error())
		}
		for _, input := range definition.Inputs {
			inputs[input.Name] = apistructs.Input{
				Name:  input.Name,
				Value: triggerDefinitionPipeInputMap[input.Name],
				Type:  input.Type,
			}
		}
	} else {
		for _, input := range parentNode.getTask().Extra.Inputs {
			inputs[input.Name] = apistructs.Input{
				Name:  input.Name,
				Value: input.Value,
				Type:  input.Type,
			}
		}
	}

	return inputs, nil
}

func (node *Node) getPlaceholderContextValue() apistructs.Contexts {
	parentNode := node.flow.getNode(node.parentTaskId)

	var contexts = apistructs.Contexts{}
	if node.flow.dbPipeExtra.Contexts != nil {
		for _, rootContext := range *node.flow.dbPipeExtra.Contexts {
			contexts[rootContext.Name] = apistructs.Context{
				Name:  rootContext.Name,
				Value: rootContext.Value,
				Type:  rootContext.Type,
			}
		}
	}

	for _, nodeContext := range parentNode.getTask().Extra.Contexts {
		contexts[nodeContext.Name] = apistructs.Context{
			Name:  nodeContext.Name,
			Value: nodeContext.Value,
			Type:  nodeContext.Type,
		}
	}

	return contexts
}

func (node *Node) getPlaceholderOutputValue() (apistructs.Outputs, error) {
	if node == node.flow.rootNode {
		return nil, nil
	}

	var matchString string
	if node.taskDefinition.Type == apistructs.PipeType {
		inputsYaml, err := yaml.Marshal(node.taskDefinition.Inputs)
		if err != nil {
			return nil, err
		}
		matchString = string(inputsYaml)
	} else {
		commandsYaml, err := yaml.Marshal(node.taskDefinition.Commands)
		if err != nil {
			return nil, err
		}
		matchString = string(commandsYaml)
	}

	var outputTaskNames []string
	_ = placeholder.MatchHolderFromHandler(matchString, map[placeholder.Type]placeholder.Handler{
		placeholder.OutputType: func(holder string, values ...string) error {
			taskName := values[1]
			outputTaskNames = append(outputTaskNames, taskName)
			return nil
		},
	})

	var outputs = apistructs.Outputs{}
	for _, taskName := range outputTaskNames {
		task := node.flow.getTask(node.parentTaskId, taskName)
		if task == nil {
			continue
		}

		if task.Outputs != nil {
			for name, value := range *task.Outputs {
				outputs[placeholder.MakeOutputKey(taskName, name)] = apistructs.Output{
					Name:  value.Name,
					Value: value.Value,
					Type:  value.Type,
				}
			}
		}
	}

	return outputs, nil
}
