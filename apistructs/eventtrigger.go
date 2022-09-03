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
