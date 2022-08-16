package triggerdefinitionclient

import (
	"gorm.io/gorm"
	"tiggerops/apistructs"
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

	err := tx.Updates(t).Error
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
}

func (client *Client) ListEventTriggerDefinition(tx *gorm.DB, query ListEventTriggerDefinitionQuery) ([]EventTriggerDefinition, error) {
	if tx == nil {
		tx = client.client
	}
	if query.Creater != "" {
		tx = tx.Where("creater = ?", query.Creater)
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
