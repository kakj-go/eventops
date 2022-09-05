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

package register

import (
	"eventops/apistructs"
	"eventops/internal/core/client/triggerdefinitionclient"
	"eventops/internal/core/token"
	"eventops/pkg/responsehandler"
	"eventops/pkg/schema/event"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"net/http"
)

func (s *Service) ListMyTriggerDefinition(c *gin.Context) {
	dbTriggers, err := s.triggerDefinitionClient.ListEventTriggerDefinition(nil, triggerdefinitionclient.ListEventTriggerDefinitionQuery{
		Creater: token.GetUserName(c),
	})
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to list event trigger list error: %v", err), nil))
		return
	}

	var result []apistructs.EventTriggerDefinition
	for _, trigger := range dbTriggers {
		result = append(result, trigger.ToApiStructs())
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", result))
}

//type CancelPipelineQuery struct {
//	TriggerName string
//}
//
//func (s *Service) ListEventTrigger(c *gin.Context) {
//	dbTriggers, err := s.triggerDefinitionClient.ListEventTriggerDefinition(nil, triggerdefinitionclient.ListEventTriggerDefinitionQuery{
//		Creater: token.GetUserName(c),
//	})
//	if err != nil {
//		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to list trigger list error: %v", err), nil))
//		return
//	}
//
//	var result []apistructs.EventTriggerDefinition
//	for _, trigger := range dbTriggers {
//		result = append(result, trigger.ToApiStructs())
//	}
//
//	c.JSON(responsehandler.Build(http.StatusOK, "", result))
//}

type ApplyTriggerRequest struct {
	TriggerContent string `json:"triggerContent"`
}

func (s *Service) ApplyTriggerDefinition(c *gin.Context) {
	var applyInfo ApplyTriggerRequest
	if err := c.ShouldBind(&applyInfo); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	var trigger event.Trigger
	err := yaml.Unmarshal([]byte(applyInfo.TriggerContent), &trigger)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("yaml content unmarshal error: %v", err), nil))
		return
	}

	if err := trigger.Mutating(token.GetUserName(c)); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	if err := trigger.Check(token.GetUserName(c)); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	newContent, err := yaml.Marshal(trigger)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}
	applyInfo.TriggerContent = string(newContent)

	triggerDefinition, find, err := s.triggerDefinitionClient.GetEventTriggerDefinition(nil, trigger.Name, token.GetUserName(c))
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("get event trigger definition error: %v", err), nil))
		return
	}
	if !find {
		var createTrigger = triggerdefinitionclient.EventTriggerDefinition{
			Name:         trigger.Name,
			Creater:      token.GetUserName(c),
			Content:      applyInfo.TriggerContent,
			EventName:    trigger.EventName,
			EventCreater: trigger.EventCreater,
			EventVersion: trigger.EventVersion,
		}
		_, err := s.triggerDefinitionClient.CreateEventTriggerDefinition(nil, &createTrigger)
		if err != nil {
			c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("create event trigger definition error: %v", err), nil))
			return
		}
	} else {
		triggerDefinition.EventName = trigger.EventName
		triggerDefinition.EventCreater = trigger.EventCreater
		triggerDefinition.EventVersion = trigger.EventVersion
		triggerDefinition.Content = applyInfo.TriggerContent

		_, err := s.triggerDefinitionClient.UpdateEventTriggerDefinition(nil, triggerDefinition)
		if err != nil {
			c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("update event trigger definition error: %v", err), nil))
			return
		}
	}
	s.eventProcess.DeleteTriggerCache(s.eventProcess.MakeCacheKey(trigger.EventName, trigger.EventVersion, trigger.EventCreater))

	c.JSON(responsehandler.Build(http.StatusOK, "", nil))
}

type DeleteNameUrlQuery struct {
	Name string `uri:"name"`
}

func (s *Service) DeleteTriggerDefinition(c *gin.Context) {
	var deleteQuery = DeleteNameUrlQuery{}
	if err := c.ShouldBindUri(&deleteQuery); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get name from uri error: %v", err), nil))
		return
	}

	dbEventTriggerDefinition, find, err := s.triggerDefinitionClient.GetEventTriggerDefinition(nil, deleteQuery.Name, token.GetUserName(c))
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("Get event trigger definition error: %v", err), nil))
		return
	}
	if !find {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("not find event trigger definition"), nil))
		return
	}

	err = s.triggerDefinitionClient.DeleteEventTriggerDefinition(nil, dbEventTriggerDefinition.Name, token.GetUserName(c))
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("delete event trigger definition error: %v", err), nil))
		return
	}
	s.eventProcess.DeleteTriggerCache(s.eventProcess.MakeCacheKey(dbEventTriggerDefinition.EventName, dbEventTriggerDefinition.EventVersion, dbEventTriggerDefinition.EventCreater))

	c.JSON(responsehandler.Build(http.StatusOK, "", nil))
}
