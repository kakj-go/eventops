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

package flowmanager

import (
	"eventops/internal/core/actuator"
	"eventops/internal/core/actuator/docker"
	"eventops/internal/core/actuator/k8s"
	"eventops/internal/core/actuator/os"
	"eventops/internal/core/client/actuatorclient"
	"eventops/pkg/retry"
	actuatordefinition "eventops/pkg/schema/actuator"
	"fmt"
	"github.com/rancher/remotedialer"
	"gopkg.in/yaml.v3"
	"time"
)

func (node *Node) actuator() (actuator.Actuator, string, error) {
	definition, err := node.flow.getAndSetPipelineVersionDefinition(node.flow.rootNode.image)
	if err != nil {
		return nil, "", err
	}

	taskTags := node.taskDefinition.ActuatorSelector.Tags
	pipelineTags := definition.ActuatorSelector.Tags

	var allTags []string
	allTags = append(allTags, taskTags...)
	allTags = append(allTags, pipelineTags...)
	actuatorDefinitionMap, err := node.getTagsActuatorDefinitionMap(allTags)
	if err != nil {
		return nil, "", err
	}

	var chooseActuatorDefinition *actuatordefinition.Client
	chooseTag := node.getTask().Extra.ChooseTag
	if chooseTag != "" {
		list := actuatorDefinitionMap[chooseTag]
		if len(list) > 0 {
			firstDefinition := list[0]
			chooseActuatorDefinition = &firstDefinition
		}
	}

	if chooseActuatorDefinition == nil && len(taskTags) > 0 {
		for _, taskTag := range taskTags {
			list := actuatorDefinitionMap[taskTag]
			if len(list) > 0 {
				firstDefinition := list[0]
				chooseActuatorDefinition = &firstDefinition
				chooseTag = taskTag
				break
			}
		}
	}

	if chooseActuatorDefinition == nil && len(pipelineTags) > 0 {
		for _, pipelineTag := range pipelineTags {
			list := actuatorDefinitionMap[pipelineTag]
			if len(list) > 0 {
				firstDefinition := list[0]
				chooseActuatorDefinition = &firstDefinition
				chooseTag = pipelineTag
				break
			}
		}
	}

	if chooseActuatorDefinition == nil {
		return nil, "", fmt.Errorf("task alias: %v type: %v No suitable actuatordefinition found", node.getTask().Alias, node.getTask().Type)
	}

	var dialer remotedialer.Dialer
	if chooseActuatorDefinition.Tunnel != nil {
		err := retry.DoWithInterval(func() error {
			dialer = node.flowManager.dialerServer.GetClient(node.getTask().Creater, chooseActuatorDefinition.Tunnel.ClientId)
			if dialer == nil {
				return fmt.Errorf("not find dialer client")
			}
			return nil
		}, 10, 5*time.Second)
		if err != nil {
			return nil, "", fmt.Errorf("not find tag: %v actuator name: %v tunnel client", chooseTag, chooseActuatorDefinition.Name)
		}
	}

	newActuator, err := NewActuator(*chooseActuatorDefinition, dialer)
	if err != nil {
		return nil, "", err
	}
	return newActuator, chooseTag, nil
}

func (node *Node) getTagsActuatorDefinitionMap(tags []string) (map[string][]actuatordefinition.Client, error) {
	actuatorTag, err := node.flowManager.clientManager.actuatorClient.ListActuatorTags(nil, actuatorclient.ListActuatorTagQuery{
		Tags:            tags,
		ActuatorCreater: node.getTask().Creater,
	})
	if err != nil {
		return nil, err
	}
	if len(actuatorTag) == 0 {
		return nil, nil
	}

	var actuatorIdList []uint64
	for _, tag := range actuatorTag {
		actuatorIdList = append(actuatorIdList, tag.ActuatorId)
	}

	dbActuators, err := node.flowManager.clientManager.actuatorClient.ListActuator(nil, actuatorclient.ListActuatorQuery{
		IdList: actuatorIdList,
	})
	if err != nil {
		return nil, err
	}

	var result = map[string][]actuatordefinition.Client{}
	for _, dbActuator := range dbActuators {
		var actuatorInfo actuatordefinition.Client
		err := yaml.Unmarshal([]byte(dbActuator.Content), &actuatorInfo)
		if err != nil {
			return nil, err
		}
		if actuatorInfo.GetType() != node.getTask().Type {
			continue
		}

		for _, tag := range actuatorInfo.Tags {
			result[tag] = append(result[tag], actuatorInfo)
		}
	}

	return result, nil
}

func NewActuator(client actuatordefinition.Client, dialer remotedialer.Dialer) (actuator.Actuator, error) {
	if client.Os != nil {
		return os.NewOsClient(client.Os, dialer)
	}
	if client.Kubernetes != nil {
		return k8s.NewKubernetesClient(client.Kubernetes, dialer)
	}
	if client.Docker != nil {
		return docker.NewDockerClient(client.Docker, dialer)
	}
	return nil, fmt.Errorf("not support actuator type")
}
