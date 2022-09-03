package apistructs

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type EventStatus string

const EventCreatedStatus EventStatus = "created"
const EventProcessingStatus EventStatus = "processing"
const EventProcessFailedStatus EventStatus = "processFailed"
const EventProcessedStatus EventStatus = "processed"

type Value string

type FileValueType string

type FileValue struct {
	Value string        `json:"value"`
	Type  FileValueType `json:"type"`
}

type Event struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	Values map[string]Value     `json:"values"`
	Files  map[string]FileValue `json:"files"`
	Labels map[string]string    `json:"labels"`

	Timestamp    int64    `json:"timestamp"`
	SupportUsers []string `json:"users"`
}

type EventDetail struct {
	Id            uint64      `json:"id"`
	Name          string      `json:"name"`
	Version       string      `json:"version"`
	Content       string      `json:"content"`
	Creater       string      `json:"creater"`
	Status        EventStatus `json:"status"`
	StatusMessage string      `json:"statusMessage"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}

func (event Event) Check() error {
	if event.Name == "" {
		return fmt.Errorf("event name can not empty")
	}
	if event.Version == "" {
		return fmt.Errorf("event version can not empty")
	}
	if event.Timestamp <= 0 {
		return fmt.Errorf("event timestamp can not empty")
	}
	for key, value := range event.Values {
		if key == "" {
			return fmt.Errorf("event values key can not empty")
		}
		if value == "" {
			return fmt.Errorf("event values key %v value can not empty", key)
		}
	}

	for key, value := range event.Files {
		if key == "" {
			return fmt.Errorf("event files key can not empty")
		}
		if value.Type == "" {
			return fmt.Errorf("event files key %v type can not empty", key)
		}
		if value.Value == "" {
			return fmt.Errorf("event files key %v value can not empty", key)
		}
	}

	for key, value := range event.Labels {
		if key == "" {
			return fmt.Errorf("event labels key can not empty")
		}
		if value == "" {
			return fmt.Errorf("event labels key %v type can not empty", key)
		}
	}
	return nil
}
