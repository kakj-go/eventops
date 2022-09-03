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
	"context"
	"eventops/apistructs"
	"eventops/internal/core/client/eventclient"
	"eventops/internal/core/client/pipelineclient"
	"eventops/internal/core/client/pipelinedefinitionclient"
	"eventops/internal/core/client/taskclient"
	"eventops/internal/core/client/triggerdefinitionclient"
	"eventops/internal/core/dialer"
	"eventops/pkg/limit_sync_group"
	"eventops/pkg/schema/pipeline"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"sync"
)

type FlowManager struct {
	ctx context.Context

	flows map[uint64]*Flow
	lock  sync.Mutex

	clientManager *clientManager
	dialerServer  *dialer.Server
}

func NewFlowManager(parentCtx context.Context, client *gorm.DB, dialerServer *dialer.Server) *FlowManager {
	clientManager := newClientManager(client)

	return &FlowManager{
		ctx:   parentCtx,
		flows: map[uint64]*Flow{},

		clientManager: clientManager,
		dialerServer:  dialerServer,
	}
}

func (m *FlowManager) Run() error {
	runningPipelines, err := m.clientManager.pipelineClient.ListPipeline(nil, pipelineclient.ListPipelineQuery{
		Statuses: []apistructs.PipelineStatus{
			apistructs.PipelineRunningStatus,
		},
	})
	if err != nil {
		return err
	}

	go func() {
		worker := limit_sync_group.NewWorker(10)
		for index := range runningPipelines {
			worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
				dbPipeline := runningPipelines[i[0].(int)]

				extra, find, err := m.clientManager.pipelineClient.GetPipelineExtra(nil, dbPipeline.Id)
				if err != nil {
					logrus.Errorf("get pipelineId: %v pipelineExtra error: %v", dbPipeline.Id, err)
					return nil
				}
				if !find {
					logrus.Errorf("not find pipelineId: %v pipelineExtra", dbPipeline.Id)
					return nil
				}

				flow, err := newFlow(m, &dbPipeline, extra)
				if err != nil {
					logrus.Errorf("build pipelineId: %v flow error %v", dbPipeline.Id, err)
					return nil
				}
				go m.runFlow(flow)
				return nil
			}, index)
		}
		worker.Do()
	}()

	return nil
}

func (m *FlowManager) runFlow(flow *Flow) {
	m.lock.Lock()
	m.flows[flow.dbPipe.Id] = flow
	m.lock.Unlock()

	flow.run()

	m.lock.Lock()
	delete(m.flows, flow.dbPipe.Id)
	m.lock.Unlock()
}

func (m *FlowManager) GetFlow(id uint64) *Flow {
	m.lock.Lock()
	flow := m.flows[id]
	m.lock.Unlock()

	return flow
}

func (m *FlowManager) CancelFlow(id uint64, user string, callback func()) {
	m.lock.Lock()
	flow := m.flows[id]
	m.lock.Unlock()

	if flow == nil {
		return
	}

	flow.lazyStopPipelineWithCallback(apistructs.PipelineCancelStatus, fmt.Sprintf("user: %v stop", user), callback)
}

func (m *FlowManager) Callback(body apistructs.CallbackBody) error {
	flow := m.GetFlow(body.PipelineId)
	if flow == nil {
		return nil
	}

	node := flow.getNode(body.TaskId)
	if node == nil {
		return nil
	}

	if len(body.Outputs) == 0 {
		return nil
	}

	if node.getTask().Extra.Auth != body.Auth {
		return fmt.Errorf("callback auth failed")
	}

	var setOutputs = apistructs.Outputs{}
	for _, output := range node.taskDefinition.Outputs {
		value := body.Outputs[output.Name]
		setOutputs[output.Name] = apistructs.Output{
			Name:  output.Name,
			Value: value,
			Type:  output.Type,
		}
	}

	return node.setDbTask(WithExtraOutputs(taskclient.Outputs(setOutputs)))
}

func (m *FlowManager) RunFlowByEventTrigger(id uint64) error {
	eventTrigger, find, err := m.clientManager.eventTriggerClient.GetEventTrigger(nil, id)
	if err != nil {
		return err
	}
	if !find {
		return fmt.Errorf("not find id: %v event trigger", id)
	}

	var dbPipeline *pipelineclient.Pipeline
	var dbPipelineExtra *pipelineclient.PipelineExtra
	var flow *Flow

	dbPipeline, _, err = m.clientManager.pipelineClient.GetPipelineByEventTriggerId(nil, id)
	if err != nil {
		return err
	}

	if dbPipeline != nil {
		dbPipelineExtra, _, err = m.clientManager.pipelineClient.GetPipelineExtra(nil, dbPipeline.Id)
		if err != nil {
			return err
		}

		if dbPipelineExtra == nil {
			logrus.Errorf("pipelineid: %v pipelineExtra was empry", id)
			return nil
		}

		err = m.clientManager.db.Transaction(func(tx *gorm.DB) error {
			err = m.clientManager.eventTriggerClient.UpdateEventTriggerStatus(nil, eventTrigger.Id, eventTrigger.Status, apistructs.ProcessedEventTriggerStatus, "")
			if err != nil {
				return err
			}

			flow, err = newFlow(m, dbPipeline, dbPipelineExtra)
			if err != nil {
				return err
			}
			go m.runFlow(flow)
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		associatedData, err := m.getAssociatedDataFromDb(eventTrigger)
		if err != nil {
			return err
		}

		dbPipeline = &pipelineclient.Pipeline{
			EventTriggerId:      eventTrigger.Id,
			EventId:             eventTrigger.EventId,
			TriggerDefinitionId: associatedData.triggerDefinition.Id,
			DefinitionName:      associatedData.pipelineVersionDefinition.Name,
			DefinitionCreater:   associatedData.pipelineVersionDefinition.Creater,
			DefinitionVersion:   associatedData.pipelineVersionDefinition.Version,
			Creater:             eventTrigger.TriggerCreater,
			Status:              apistructs.PipelineRunningStatus,
			CostTimeSec:         0,
		}
		dbPipelineExtra = buildPipelineExtra(&associatedData)

		err = m.clientManager.db.Transaction(func(tx *gorm.DB) error {
			dbPipeline, err := m.clientManager.pipelineClient.CreatePipeline(tx, dbPipeline)
			if err != nil {
				return err
			}

			dbPipelineExtra.PipelineId = dbPipeline.Id
			_, err = m.clientManager.pipelineClient.CreatePipelineExtra(tx, dbPipelineExtra)
			if err != nil {
				return err
			}

			err = m.clientManager.eventTriggerClient.UpdateEventTriggerStatus(tx, eventTrigger.Id, eventTrigger.Status, apistructs.ProcessedEventTriggerStatus, "")
			if err != nil {
				return err
			}

			flow, err = newFlow(m, dbPipeline, dbPipelineExtra)
			if err != nil {
				return err
			}

			go m.runFlow(flow)
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type AssociatedData struct {
	eventTrigger              *eventclient.EventTrigger
	event                     *eventclient.Event
	triggerDefinition         *triggerdefinitionclient.EventTriggerDefinition
	pipelineVersionDefinition *pipelinedefinitionclient.PipelineVersionDefinition
}

func (m *FlowManager) getAssociatedDataFromDb(eventTrigger *eventclient.EventTrigger) (data AssociatedData, err error) {
	data.eventTrigger = eventTrigger
	worker := limit_sync_group.NewWorker(3)
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		data.event, err = m.clientManager.eventTriggerClient.GetEventById(nil, eventTrigger.EventId)
		if err != nil {
			return err
		}
		return nil
	})
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		var find = false
		data.triggerDefinition, find, err = m.clientManager.triggerDefinitionClient.GetEventTriggerDefinition(nil, eventTrigger.TriggerName, eventTrigger.TriggerCreater)
		if err != nil {
			return err
		}
		if !find {
			return fmt.Errorf("not find trigger definition name: %v craeter: %v", eventTrigger.TriggerName, eventTrigger.TriggerCreater)
		}
		return nil
	})

	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		var find = false
		pipelineName := pipeline.GetImageName(eventTrigger.PipelineImage)
		pipelineVersion := pipeline.GetImageVersion(eventTrigger.PipelineImage)
		pipelineCreater := pipeline.GetImageCreater(eventTrigger.PipelineImage)
		data.pipelineVersionDefinition, find, err = m.clientManager.pipelineDefinitionClient.GetPipelineVersionDefinition(nil, pipelineName, pipelineVersion, pipelineCreater)
		if err != nil {
			return err
		}
		if !find {
			return fmt.Errorf("not find pipeline definition name: %v version: %v creater: %v", pipelineName, pipelineVersion, pipelineCreater)
		}
		return nil
	})
	err = worker.Do().Error()
	if err != nil {
		return
	}
	return
}

func buildPipelineExtra(data *AssociatedData) *pipelineclient.PipelineExtra {
	pipelineExtra := pipelineclient.PipelineExtra{
		DefinitionContent:        data.pipelineVersionDefinition,
		EventContent:             data.event,
		EventTriggerContent:      data.eventTrigger,
		TriggerDefinitionContent: data.triggerDefinition,
		Extra:                    &pipelineclient.PipelineExtraInfo{},
		Contexts:                 &pipelineclient.PipelineExtraContents{},
	}
	return &pipelineExtra
}
