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
	"eventops/internal/core/client/pipelineclient"
	"eventops/internal/core/token"
	"eventops/pkg/limit_sync_group"
	"eventops/pkg/responsehandler"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type CancelPipelineQuery struct {
	Id uint64 `uri:"id"`
}

func (s *Service) Cancel(c *gin.Context) {
	var cancel CancelPipelineQuery
	if err := c.ShouldBindUri(&cancel); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to cancel pipeline runtime: %v error: %v", cancel.Id, err), nil))
		return
	}

	flow := s.manager.GetFlow(cancel.Id)
	if flow == nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("not find this runtime: %v", cancel.Id), nil))
		return
	}

	var stopChan = make(chan struct{})
	s.manager.CancelFlow(cancel.Id, token.GetUserName(c), func() {
		go func() {
			select {
			case <-time.After(10 * time.Second):
				<-stopChan
			}
		}()
		stopChan <- struct{}{}
	})

	var status = ""
	select {
	case <-stopChan:
		status = "success"
	case <-time.After(5 * time.Second):
		status = "stopping"
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", fmt.Sprintf("pipeline runtime cancel %v", status)))
}

type GetPipelineQuery struct {
	Id uint64 `uri:"id"`
}

func (s *Service) Get(c *gin.Context) {
	var get GetPipelineQuery
	if err := c.ShouldBindUri(&get); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get pipeline runtime: %v error: %v", get.Id, err), nil))
		return
	}

	var pipelineDetail = apistructs.PipelineDetail{}
	worker := limit_sync_group.NewWorker(3)
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		dbPipeline, _, err := s.pipelineDbClient.GetPipeline(nil, get.Id, token.GetUserName(c))
		if err != nil {
			return err
		}
		if dbPipeline != nil {
			pipelineDetail.Pipeline = dbPipeline.ToApiStruct()
		}
		return nil
	})
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		dbPipelineExtra, _, err := s.pipelineDbClient.GetPipelineExtra(nil, get.Id)
		if err != nil {
			return err
		}
		if dbPipelineExtra != nil {
			pipelineDetail.PipelineExtra = dbPipelineExtra.ToApiStruct()
		}
		return nil
	})
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		dbTasks, err := s.taskDbClient.ListTasks(nil, get.Id, token.GetUserName(c))
		if err != nil {
			return err
		}
		var tasks []apistructs.Task
		for _, dbTask := range dbTasks {
			tasks = append(tasks, dbTask.ToApiStruct())
		}
		pipelineDetail.Tasks = tasks
		return nil
	})
	err := worker.Do().Error()
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get pipeline runtime detail error: %v", err), nil))
		return
	}

	if pipelineDetail.Pipeline.Id == 0 {
		c.JSON(responsehandler.Build(http.StatusOK, "", nil))
		return
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", pipelineDetail))
}

type pipelineListQuery struct {
	EventName    string
	EventVersion string
	EventCreater string

	TriggerDefinitionName string

	PipelineDefinitionName    string
	PipelineDefinitionVersion string
	PipelineDefinitionCreater string

	Top uint64
}

func (s *Service) List(c *gin.Context) {
	var body = pipelineListQuery{}
	body.EventName = c.Query("eventName")
	body.EventVersion = c.Query("eventVersion")
	body.EventCreater = c.Query("eventCreater")
	body.TriggerDefinitionName = c.Query("triggerDefinitionName")
	body.PipelineDefinitionName = c.Query("pipelineDefinitionName")
	body.PipelineDefinitionVersion = c.Query("pipelineDefinitionVersion")
	body.PipelineDefinitionCreater = c.Query("pipelineDefinitionCreater")
	top := c.Query("top")
	if len(top) > 0 {
		topInt, err := strconv.ParseUint(c.Query("top"), 10, 64)
		if err != nil {
			c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("parse top error: %v", err), nil))
			return
		}
		body.Top = topInt
	}

	if body.Top > 100 {
		body.Top = 100
	}
	if body.Top == 0 {
		body.Top = 20
	}

	var query pipelineclient.ListPipelineQuery
	query.Creater = token.GetUserName(c)
	query.Top = body.Top

	worker := limit_sync_group.NewWorker(3)
	if body.EventName != "" && body.EventVersion != "" && body.EventCreater != "" {
		worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			event, err := s.eventClient.GetEvent(nil, body.EventName, body.EventVersion, body.EventCreater)
			if err != nil {
				return err
			}
			if event == nil {
				return fmt.Errorf("not find name: %v version: %v creater: %v event", body.EventName, body.EventVersion, body.EventCreater)
			}

			query.EventId = event.Id
			return nil
		})
	}

	if body.TriggerDefinitionName != "" {
		worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			definition, find, err := s.triggerDefinitionClient.GetEventTriggerDefinition(nil, body.TriggerDefinitionName, token.GetUserName(c))
			if err != nil {
				return err
			}
			if !find {
				return fmt.Errorf("not find triggerDefinitionName: %v creater: %v triggerDefinition", body.TriggerDefinitionName, token.GetUserName(c))
			}
			query.TriggerDefinitionId = definition.Id
			return nil
		})
	}
	err := worker.Do().Error()
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	if body.PipelineDefinitionName != "" && body.PipelineDefinitionCreater != "" && body.PipelineDefinitionVersion != "" {
		query.PipelineDefinitionName = body.PipelineDefinitionName
		query.PipelineDefinitionCreater = body.PipelineDefinitionCreater
		query.PipelineDefinitionVersion = body.PipelineDefinitionVersion
	}

	result, err := s.pipelineDbClient.ListPipeline(nil, query)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to list pipeline runtime error: %v", err), nil))
		return
	}

	var pipelines []apistructs.Pipeline
	for _, dbPipeline := range result {
		pipelines = append(pipelines, dbPipeline.ToApiStruct())
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", pipelines))
}

func (s *Service) Callback(c *gin.Context) {
	var callback apistructs.CallbackBody
	if err := c.ShouldBind(&callback); err != nil {
		c.JSON(http.StatusServiceUnavailable, fmt.Sprintf("pipeline runtime callback error: %v", err))
		return
	}

	err := s.manager.Callback(callback)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, fmt.Sprintf("pipeline runtime callback error: %v", err))
		return
	}
	c.JSON(http.StatusOK, "success")
}
