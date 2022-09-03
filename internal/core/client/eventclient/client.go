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

package eventclient

import (
	"database/sql/driver"
	"encoding/json"
	"eventops/apistructs"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Event struct {
	Id            uint64                 `json:"id"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Content       string                 `json:"content"`
	Creater       string                 `json:"creater"`
	Status        apistructs.EventStatus `json:"status"`
	StatusMessage string                 `json:"status_Message"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (event *Event) ToApiStructs() apistructs.EventDetail {
	return apistructs.EventDetail{
		Id:            event.Id,
		Name:          event.Name,
		Version:       event.Version,
		Content:       event.Content,
		Creater:       event.Creater,
		Status:        event.Status,
		StatusMessage: event.StatusMessage,

		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}
}

func (args *Event) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &args)
}

func (args *Event) Value() (driver.Value, error) {
	if args == nil {
		return nil, nil
	}

	return json.Marshal(args)
}

type EventTrigger struct {
	Id      uint64                        `json:"id"`
	Status  apistructs.EventTriggerStatus `json:"status"`
	Message string                        `json:"message"`

	EventName    string    `json:"event_name"`
	EventVersion string    `json:"event_version"`
	EventCreater string    `json:"event_creater"`
	EventTime    time.Time `json:"event_time"`
	EventId      uint64    `json:"event_id"`

	TriggerName    string    `json:"trigger_name"`
	TriggerCreater string    `json:"trigger_creater"`
	TriggerTime    time.Time `json:"trigger_time"`

	PipelineImage string `json:"pipeline_image"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (event *EventTrigger) ToApiStructs() apistructs.EventTrigger {
	return apistructs.EventTrigger{
		Id:      event.Id,
		Status:  event.Status,
		Message: event.Message,

		EventName:    event.EventName,
		EventVersion: event.EventVersion,
		EventCreater: event.EventCreater,
		EventTime:    event.EventTime,
		EventId:      event.Id,

		TriggerName:    event.TriggerName,
		TriggerCreater: event.TriggerCreater,
		TriggerTime:    event.TriggerTime,

		PipelineImage: event.PipelineImage,

		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}
}

func (args *EventTrigger) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &args)
}

func (args *EventTrigger) Value() (driver.Value, error) {
	if args == nil {
		return nil, nil
	}

	return json.Marshal(args)
}

type Client struct {
	client *gorm.DB
}

func NewEventClient(client *gorm.DB) *Client {
	return &Client{client: client}
}

func (client *Client) BatchCreateEventTrigger(tx *gorm.DB, eventTriggers []EventTrigger) error {
	if tx == nil {
		tx = client.client
	}

	return tx.Create(eventTriggers).Error
}

func (client *Client) GetEventTrigger(tx *gorm.DB, id uint64) (*EventTrigger, bool, error) {
	if tx == nil {
		tx = client.client
	}

	var eventTrigger EventTrigger
	err := tx.Where("id = ?", id).First(&eventTrigger).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &eventTrigger, true, nil
}

func (client *Client) UpdateEventTriggerStatus(tx *gorm.DB, id uint64, preStatus apistructs.EventTriggerStatus, status apistructs.EventTriggerStatus, msg string) error {
	if tx == nil {
		tx = client.client
	}

	return tx.Model(&EventTrigger{}).Where("status = ? and id = ?", preStatus, id).Updates(map[string]interface{}{"status": status, "message": msg}).Error
}

func (client *Client) CreateEvent(tx *gorm.DB, t *Event) (*Event, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) GetEvent(tx *gorm.DB, name string, version string, creater string) (*Event, error) {
	if tx == nil {
		tx = client.client
	}

	var event Event
	err := tx.Where("name = ? and version = ? and creater = ?", name, version, creater).First(&event).Error
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (client *Client) GetEventById(tx *gorm.DB, id uint64) (*Event, error) {
	if tx == nil {
		tx = client.client
	}

	var event Event
	err := tx.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}

	return &event, nil
}

type EventQuery struct {
	Statues []apistructs.EventStatus
}

func (client *Client) ListEvent(tx *gorm.DB, query EventQuery) ([]Event, error) {
	if tx == nil {
		tx = client.client
	}

	if len(query.Statues) > 0 {
		tx = tx.Where("status in (?)", query.Statues)
	}

	var events []Event
	if err := tx.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

type EventTriggerQuery struct {
	Statues []apistructs.EventTriggerStatus
}

func (client *Client) ListEventTrigger(tx *gorm.DB, query EventTriggerQuery) ([]EventTrigger, error) {
	if tx == nil {
		tx = client.client
	}

	if len(query.Statues) > 0 {
		tx = tx.Where("status in (?)", query.Statues)
	}

	var eventTriggers []EventTrigger
	if err := tx.Find(&eventTriggers).Error; err != nil {
		return nil, err
	}
	return eventTriggers, nil
}

func (client *Client) UpdateEventStatus(tx *gorm.DB, id uint64, preStatus apistructs.EventStatus, status apistructs.EventStatus, msg string) error {
	if tx == nil {
		tx = client.client
	}

	return tx.Model(&Event{}).Where("status = ? and id = ?", preStatus, id).Updates(map[string]interface{}{"status": status, "status_message": msg}).Error
}
