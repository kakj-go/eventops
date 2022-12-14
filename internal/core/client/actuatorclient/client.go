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

package actuatorclient

import (
	"eventops/apistructs"
	"eventops/pkg/schema/actuator"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"time"
)

type Client struct {
	client *gorm.DB
}

func NewActuatorsClient(client *gorm.DB) *Client {
	return &Client{client: client}
}

type Actuator struct {
	Id          uint64              `json:"id"`
	Name        string              `json:"name"`
	Creater     string              `json:"creater"`
	Type        apistructs.TaskType `json:"type"`
	Status      string              `json:"status"`
	Content     string              `json:"content"`
	ClientId    string              `json:"client_id"`
	ClientToken string              `json:"client_token"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (a Actuator) ToApiStructs() (apistructs.Actuator, error) {
	var actuatorInfo actuator.Client
	err := yaml.Unmarshal([]byte(a.Content), &actuatorInfo)
	if err != nil {
		return apistructs.Actuator{}, err
	}

	return apistructs.Actuator{
		Name:      a.Name,
		Creater:   a.Creater,
		Type:      a.Type.String(),
		Status:    a.Status,
		Tags:      actuatorInfo.Tags,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}, nil
}

type ActuatorTag struct {
	Id              uint64              `json:"id"`
	ActuatorName    string              `json:"actuator_name"`
	ActuatorCreater string              `json:"actuator_creater"`
	ActuatorType    apistructs.TaskType `json:"actuator_type"`
	Tag             string              `json:"tag"`
	ActuatorId      uint64              `json:"actuator_id"`
}

func (client *Client) GetActuator(tx *gorm.DB, name string, creater string) (*Actuator, bool, error) {
	if tx == nil {
		tx = client.client
	}

	var result Actuator
	err := tx.Where("name = ? and creater = ?", name, creater).First(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &result, true, nil
}

func (client *Client) UpdateActuator(tx *gorm.DB, t *Actuator) (*Actuator, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Model(&Actuator{}).Select("*").Where("id = ?", t.Id).Updates(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) CreateActuator(tx *gorm.DB, t *Actuator) (*Actuator, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) DeleteActuator(tx *gorm.DB, name, creater string) error {
	if tx == nil {
		tx = client.client
	}
	return tx.Model(&Actuator{}).Where("name = ? and creater = ?", name, creater).Delete(&Actuator{}).Error
}

type ListActuatorQuery struct {
	Creater string
	IdList  []uint64

	ClientId    string
	ClientToken string
}

func (client *Client) ListActuator(tx *gorm.DB, query ListActuatorQuery) ([]Actuator, error) {
	if tx == nil {
		tx = client.client
	}
	if query.Creater != "" {
		tx = tx.Where("creater = ?", query.Creater)
	}

	if len(query.IdList) > 0 {
		tx = tx.Where("id in (?)", query.IdList)
	}
	if query.ClientId != "" {
		tx = tx.Where("client_id = ?", query.ClientId)
	}
	if query.ClientToken != "" {
		tx = tx.Where("client_token = ?", query.ClientToken)
	}

	var list []Actuator
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (client *Client) DeleteActuatorTags(tx *gorm.DB, actuatorName, actuatorCreater string) error {
	if tx == nil {
		tx = client.client
	}
	return tx.Model(&ActuatorTag{}).Where("actuator_name = ? and actuator_creater = ?", actuatorName, actuatorCreater).Delete(&ActuatorTag{}).Error
}

type ListActuatorTagQuery struct {
	Tags            []string
	ActuatorName    string
	ActuatorCreater string
	ActuatorType    string
}

func (client *Client) ListActuatorTags(tx *gorm.DB, query ListActuatorTagQuery) ([]ActuatorTag, error) {
	if tx == nil {
		tx = client.client
	}
	if len(query.Tags) > 0 {
		tx = tx.Where("tag in (?)", query.Tags)
	}

	if query.ActuatorName != "" {
		tx = tx.Where("actuator_name = ?", query.ActuatorName)
	}
	if query.ActuatorCreater != "" {
		tx = tx.Where("actuator_creater = ?", query.ActuatorCreater)
	}
	if query.ActuatorType != "" {
		tx = tx.Where("actuator_type = ?", query.ActuatorType)
	}
	var list []ActuatorTag
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (client *Client) BatchCreateActuatorTags(tx *gorm.DB, batches []ActuatorTag) error {
	if tx == nil {
		tx = client.client
	}

	return tx.Create(batches).Error
}
