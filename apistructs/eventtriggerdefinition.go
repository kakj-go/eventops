package apistructs

import (
	"time"
)

type EventTriggerDefinition struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Creater string `json:"creater"`

	EventName    string `json:"event_name"`
	EventVersion string `json:"event_version"`
	EventCreater string `json:"event_creater"`
	CreatedAt    time.Time
}
