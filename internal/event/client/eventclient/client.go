package eventclient

import (
	"gorm.io/gorm"
	"tiggerops/apistructs"
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

	var event []Event
	if err := tx.Find(&event).Error; err != nil {
		return nil, err
	}
	return event, nil
}

func (client *Client) UpdateEventStatus(tx *gorm.DB, id uint64, preStatus apistructs.EventStatus, status apistructs.EventStatus, msg string) error {
	if tx == nil {
		tx = client.client
	}

	return tx.Model(&Event{}).Where("status = ? and id = ?", preStatus, id).Updates(map[string]interface{}{"status": status, "status_message": msg}).Error
}
