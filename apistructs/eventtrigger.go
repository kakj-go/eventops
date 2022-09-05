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

package apistructs

import (
	"gorm.io/gorm"
	"time"
)

type EventTriggerStatus string

const PassEventTriggerStatus = "pass"
const ProcessedEventTriggerStatus = "processed"
const ProcessFailedEventTriggerStatus = "processFailed"
const UnPassEventTriggerStatus = "unPass"

type EventTrigger struct {
	Id      uint64             `json:"id"`
	Status  EventTriggerStatus `json:"status"`
	Message string             `json:"message"`

	EventName    string    `json:"eventName"`
	EventVersion string    `json:"eventVersion"`
	EventCreater string    `json:"eventCreater"`
	EventTime    time.Time `json:"eventTime"`
	EventId      uint64    `json:"eventId"`

	TriggerName    string    `json:"triggerName"`
	TriggerCreater string    `json:"triggerCreater"`
	TriggerTime    time.Time `json:"triggerTime"`

	PipelineImage string `json:"pipelineImage"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}
