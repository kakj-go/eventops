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

package taskclient

import (
	"database/sql/driver"
	"encoding/json"
	"eventops/apistructs"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Task struct {
	Id           uint64                `json:"id"`
	PipelineId   uint64                `json:"pipeline_id"`
	ParentTaskId uint64                `json:"parent_task_id"`
	JobSign      string                `json:"job_sign"`
	Alias        string                `json:"alias"`
	Type         apistructs.TaskType   `json:"type"`
	Status       apistructs.TaskStatus `json:"status"`
	Extra        *TaskExtra            `json:"extra" gorm:"type:text"`
	Outputs      *Outputs              `json:"outputs" gorm:"type:text"`
	CostTimeSec  uint64                `json:"cost_time_sec"`
	TimeBegin    *time.Time            `json:"time_begin"`
	TimeEnd      *time.Time            `json:"time_end"`
	Creater      string                `json:"creater"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t Task) ToApiStruct() apistructs.Task {
	return apistructs.Task{
		Id:           t.Id,
		PipelineId:   t.PipelineId,
		ParentTaskId: t.ParentTaskId,
		JobSign:      t.JobSign,
		Alias:        t.Alias,
		Type:         t.Type,
		Status:       t.Status,
		Extra: &apistructs.TaskExtra{
			Error:     t.Extra.Error,
			ChooseTag: t.Extra.ChooseTag,
			Inputs: func() apistructs.Inputs {
				if t.Extra.Inputs == nil {
					return nil
				}
				inputs := apistructs.Inputs(t.Extra.Inputs)
				return inputs
			}(),
			Contexts: func() apistructs.Contexts {
				if t.Extra.Contexts == nil {
					return nil
				}
				contexts := apistructs.Contexts(t.Extra.Contexts)
				return contexts
			}(),
		},
		Outputs: func() apistructs.Outputs {
			if t.Outputs == nil {
				return nil
			}
			outputs := apistructs.Outputs(*t.Outputs)
			return outputs
		}(),
		CostTimeSec: t.CostTimeSec,
		TimeBegin:   t.TimeBegin,
		TimeEnd:     t.TimeEnd,
		Creater:     t.Creater,

		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (t Task) TableName() string {
	return "pipeline_tasks"
}

type TaskExtra struct {
	Error     string   `json:"error,omitempty"`
	ChooseTag string   `json:"chooseTag,omitempty"`
	Inputs    Inputs   `json:"inputs,omitempty"`
	Contexts  Contexts `json:"contexts,omitempty"`
	Auth      string   `json:"auth,omitempty"`
}

type Inputs apistructs.Inputs

type Contexts apistructs.Contexts

func (args *TaskExtra) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &args)
}

func (args *TaskExtra) Value() (driver.Value, error) {
	if args == nil {
		return nil, nil
	}

	return json.Marshal(args)
}

type Outputs apistructs.Outputs

func (args *Outputs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &args)
}

func (args *Outputs) Value() (driver.Value, error) {
	if args == nil {
		return nil, nil
	}

	return json.Marshal(args)
}

type Client struct {
	client *gorm.DB
}

func NewTaskClient(client *gorm.DB) *Client {
	return &Client{client: client}
}

func (client *Client) CreateTask(tx *gorm.DB, t *Task) (*Task, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) UpdateTask(tx *gorm.DB, t *Task) (*Task, error) {
	if tx == nil {
		tx = client.client
	}
	err := tx.Model(&Task{}).Select("*").Where("id = ?", t.Id).Updates(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (client *Client) GetTask(tx *gorm.DB, pipelineId uint64, alias string, taskId uint64) (*Task, error) {
	if tx == nil {
		tx = client.client
	}
	tx = tx.Where("pipeline_id = ? and alias = ? and parent_task_id = ?", pipelineId, alias, taskId)

	var task Task
	err := tx.First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (client *Client) ListTasks(tx *gorm.DB, pipelineId uint64, user string) ([]*Task, error) {
	if tx == nil {
		tx = client.client
	}
	tx = tx.Where("pipeline_id = ? and creater = ?", pipelineId, user)

	var task []*Task
	err := tx.Find(&task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}
