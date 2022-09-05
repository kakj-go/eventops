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

package triggerdefinitionclient

import (
	"database/sql/driver"
	"encoding/json"
	"eventops/apistructs"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type EventTriggerDefinition struct {
	Id           uint64 `json:"id"`
	Name         string `json:"name"`
	Content      string `json:"content"`
	Creater      string `json:"creater"`
	EventName    string `json:"event_name"`
	EventVersion string `json:"event_version"`
	EventCreater string `json:"event_creater"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (args *EventTriggerDefinition) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &args)
}

func (args *EventTriggerDefinition) Value() (driver.Value, error) {
	if args == nil {
		return nil, nil
	}

	return json.Marshal(args)
}

func (d EventTriggerDefinition) ToApiStructs() apistructs.EventTriggerDefinition {
	return apistructs.EventTriggerDefinition{
		Name:         d.Name,
		Content:      d.Content,
		Creater:      d.Creater,
		EventName:    d.EventName,
		EventVersion: d.EventVersion,
		EventCreater: d.EventCreater,
		CreatedAt:    d.CreatedAt,
	}
}

type Client struct {
	client *gorm.DB
}

func NewTriggerDefinitionClient(client *gorm.DB) *Client {
	return &Client{client: client}
}

func (client *Client) GetEventTriggerDefinition(tx *gorm.DB, name string, creater string) (*EventTriggerDefinition, bool, error) {
	if tx == nil {
		tx = client.client
	}

	var pipelineDefinition EventTriggerDefinition
	err := tx.Where("name = ? and creater = ?", name, creater).First(&pipelineDefinition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &pipelineDefinition, true, nil
}

func (client *Client) UpdateEventTriggerDefinition(tx *gorm.DB, t *EventTriggerDefinition) (*EventTriggerDefinition, error) {
	if tx == nil {
		tx = client.client
	}
	err := tx.Model(&EventTriggerDefinition{}).Select("*").Where("id = ?", t.Id).Updates(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) CreateEventTriggerDefinition(tx *gorm.DB, t *EventTriggerDefinition) (*EventTriggerDefinition, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) DeleteEventTriggerDefinition(tx *gorm.DB, name, creater string) error {
	if tx == nil {
		tx = client.client
	}
	return tx.Model(&EventTriggerDefinition{}).Where("name = ? and creater = ?", name, creater).Delete(&EventTriggerDefinition{}).Error
}

type ListEventTriggerDefinitionQuery struct {
	Creater      string
	EventName    string
	EventVersion string
	EventCreater string
	TriggerName  string
}

func (client *Client) ListEventTriggerDefinition(tx *gorm.DB, query ListEventTriggerDefinitionQuery) ([]EventTriggerDefinition, error) {
	if tx == nil {
		tx = client.client
	}
	if query.Creater != "" {
		tx = tx.Where("trigger_creater = ?", query.Creater)
	}
	if query.TriggerName != "" {
		tx = tx.Where("trigger_name = ?", query.TriggerName)
	}

	if query.EventName != "" {
		tx = tx.Where("event_name = ?", query.EventName)
	}
	if query.EventVersion != "" {
		tx = tx.Where("event_version = ?", query.EventVersion)
	}
	if query.EventCreater != "" {
		tx = tx.Where("event_creater = ?", query.EventCreater)
	}

	var list []EventTriggerDefinition
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}
