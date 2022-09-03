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

package pipelineclient

import (
	"database/sql/driver"
	"encoding/json"
	"eventops/apistructs"
	"eventops/internal/core/client/eventclient"
	"eventops/internal/core/client/pipelinedefinitionclient"
	"eventops/internal/core/client/triggerdefinitionclient"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Pipeline struct {
	Id                  uint64                    `json:"id"`
	EventTriggerId      uint64                    `json:"event_trigger_id"`
	EventId             uint64                    `json:"event_id"`
	TriggerDefinitionId uint64                    `json:"trigger_definition_id"`
	DefinitionName      string                    `json:"definition_name"`
	DefinitionVersion   string                    `json:"definition_version"`
	DefinitionCreater   string                    `json:"definition_creater"`
	Creater             string                    `json:"creater"`
	Status              apistructs.PipelineStatus `json:"status"`
	CostTimeSec         uint64                    `json:"cost_time_sec"`
	TimeBegin           *time.Time                `json:"time_begin"`
	TimeEnd             *time.Time                `json:"time_end"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p Pipeline) ToApiStruct() apistructs.Pipeline {
	return apistructs.Pipeline{
		Id:                  p.Id,
		EventTriggerId:      p.EventTriggerId,
		EventId:             p.EventId,
		TriggerDefinitionId: p.TriggerDefinitionId,
		DefinitionName:      p.DefinitionName,
		DefinitionVersion:   p.DefinitionVersion,
		DefinitionCreater:   p.DefinitionCreater,
		Creater:             p.Creater,
		Status:              p.Status,
		CostTimeSec:         p.CostTimeSec,
		TimeBegin:           p.TimeBegin,
		TimeEnd:             p.TimeEnd,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

type PipelineExtra struct {
	Id                       uint64                                              `json:"id"`
	PipelineId               uint64                                              `json:"pipeline_id"`
	DefinitionContent        *pipelinedefinitionclient.PipelineVersionDefinition `json:"definition_content" gorm:"type:text"`
	EventContent             *eventclient.Event                                  `json:"event_content" gorm:"type:text"`
	EventTriggerContent      *eventclient.EventTrigger                           `json:"event_trigger_content" gorm:"type:text"`
	TriggerDefinitionContent *triggerdefinitionclient.EventTriggerDefinition     `json:"trigger_definition_content" gorm:"type:text"`
	Extra                    *PipelineExtraInfo                                  `json:"extra" gorm:"type:text"`
	Contexts                 *PipelineExtraContents                              `json:"contexts" gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p PipelineExtra) ToApiStruct() apistructs.PipelineExtra {
	versionDefinition := p.DefinitionContent.ToApiStructs()
	event := p.EventContent.ToApiStructs()
	eventTrigger := p.EventTriggerContent.ToApiStructs()
	triggerDefinition := p.TriggerDefinitionContent.ToApiStructs()
	pipelineExtraInfo := p.Extra.ToApiStruct()
	contexts := p.Contexts.ToApiStruct()

	extra := apistructs.PipelineExtra{
		Id:                       p.Id,
		PipelineId:               p.PipelineId,
		DefinitionContent:        &versionDefinition,
		EventContent:             &event,
		EventTriggerContent:      &eventTrigger,
		TriggerDefinitionContent: &triggerDefinition,
		Extra:                    &pipelineExtraInfo,
		Contexts:                 &contexts,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
	}
	return extra
}

type PipelineExtraInfo struct {
	StopReason string `json:"stop_reason,omitempty"`
}

func (p PipelineExtraInfo) ToApiStruct() apistructs.PipelineExtraInfo {
	extra := apistructs.PipelineExtraInfo{
		StopReason: p.StopReason,
	}
	return extra
}

func (args *PipelineExtraInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &args)
}

func (args *PipelineExtraInfo) Value() (driver.Value, error) {
	if args == nil {
		return nil, nil
	}

	return json.Marshal(args)
}

type PipelineExtraContents apistructs.Contexts

func (p PipelineExtraContents) ToApiStruct() apistructs.PipelineExtraContents {
	return apistructs.PipelineExtraContents{}
}

func (args *PipelineExtraContents) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &args)
}

func (args *PipelineExtraContents) Value() (driver.Value, error) {
	if args == nil {
		return nil, nil
	}

	return json.Marshal(args)
}

type Client struct {
	client *gorm.DB
}

func NewPipelineClient(client *gorm.DB) *Client {
	return &Client{client: client}
}

func (client *Client) GetPipelineByEventTriggerId(tx *gorm.DB, id uint64) (*Pipeline, bool, error) {
	if tx == nil {
		tx = client.client
	}

	tx = tx.Where("event_trigger_id = ?", id)

	var pipeline Pipeline
	err := tx.First(&pipeline).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &pipeline, true, nil
}

func (client *Client) GetPipeline(tx *gorm.DB, id uint64, user string) (*Pipeline, bool, error) {
	if tx == nil {
		tx = client.client
	}

	tx = tx.Where("id = ? and creater = ?", id, user)

	var pipeline Pipeline
	err := tx.First(&pipeline).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &pipeline, true, nil
}

func (client *Client) GetPipelineExtra(tx *gorm.DB, pipelineId uint64) (*PipelineExtra, bool, error) {
	if tx == nil {
		tx = client.client
	}

	tx = tx.Where("pipeline_id = ?", pipelineId)

	var pipelineExtra PipelineExtra
	err := tx.First(&pipelineExtra).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &pipelineExtra, true, nil
}

type ListPipelineQuery struct {
	TriggerDefinitionId uint64

	EventTriggerId uint64
	EventId        uint64

	PipelineDefinitionName    string
	PipelineDefinitionVersion string
	PipelineDefinitionCreater string

	Creater string

	Top uint64

	Statuses []apistructs.PipelineStatus
}

func (client *Client) ListPipeline(tx *gorm.DB, query ListPipelineQuery) ([]Pipeline, error) {
	if tx == nil {
		tx = client.client
	}

	if query.Creater != "" {
		tx = tx.Where("creater = ?", query.Creater)
	}

	if len(query.Statuses) > 0 {
		tx = tx.Where("status in (?)", query.Statuses)
	}

	if query.EventTriggerId > 0 {
		tx = tx.Where("event_trigger_id = ?", query.EventTriggerId)
	}

	if query.TriggerDefinitionId > 0 {
		tx = tx.Where("trigger_definition_id = ?", query.TriggerDefinitionId)
	}

	if query.EventId > 0 {
		tx = tx.Where("event_id = ?", query.EventId)
	}

	if query.PipelineDefinitionName != "" && query.PipelineDefinitionVersion != "" && query.PipelineDefinitionCreater != "" {
		tx = tx.Where("definition_name = ? && definition_version = ? && definition_creater = ?",
			query.PipelineDefinitionName, query.PipelineDefinitionVersion, query.PipelineDefinitionCreater)
	}
	if query.Top > 0 {
		tx = tx.Limit(int(query.Top))
	}
	tx = tx.Order("id desc")

	var list []Pipeline
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (client *Client) ListPipelineExtra(tx *gorm.DB, pipelineIdList []uint64) ([]PipelineExtra, error) {
	if tx == nil {
		tx = client.client
	}

	tx = tx.Where("pipeline_id in (?)", pipelineIdList)

	var list []PipelineExtra
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (client *Client) CreatePipeline(tx *gorm.DB, t *Pipeline) (*Pipeline, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) CreatePipelineExtra(tx *gorm.DB, t *PipelineExtra) (*PipelineExtra, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) UpdatePipeline(tx *gorm.DB, t *Pipeline) (*Pipeline, error) {
	if tx == nil {
		tx = client.client
	}
	err := tx.Model(&Pipeline{}).Select("*").Where("id = ?", t.Id).Updates(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) UpdatePipelineExtra(tx *gorm.DB, t *PipelineExtra) (*PipelineExtra, error) {
	if tx == nil {
		tx = client.client
	}
	err := tx.Model(&PipelineExtra{}).Select("*").Where("id = ?", t.Id).Updates(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}
