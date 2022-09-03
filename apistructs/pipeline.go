package apistructs

import (
	"fmt"
	"time"
)

type ValueType string

const (
	FileType ValueType = "file"
	EnvType  ValueType = "env"
)

var ValueTypeList = []ValueType{FileType, EnvType}

func (v ValueType) ValueTypeCheck() error {
	var find bool
	for _, value := range ValueTypeList {
		if value == v {
			find = true
			break
		}
	}
	if !find {
		return fmt.Errorf("value type not support, use [%v, %v]", FileType, EnvType)
	}

	return nil
}

type PipelineStatus string

const PipelineRunningStatus PipelineStatus = "running"

const PipelineSuccessStatus PipelineStatus = "success"
const PipelineFailedStatus PipelineStatus = "failed"
const PipelineCancelStatus PipelineStatus = "cancel"

func (status PipelineStatus) IsEnd() bool {
	if status == PipelineSuccessStatus || status == PipelineFailedStatus || status == PipelineCancelStatus {
		return true
	}
	return false
}

func (status PipelineStatus) IsCancel() bool {
	if status == PipelineCancelStatus {
		return true
	}
	return false
}

func (status PipelineStatus) IsRunning() bool {
	if status == PipelineRunningStatus {
		return true
	}
	return false
}

type PipelineDetail struct {
	Pipeline      Pipeline      `json:"pipeline"`
	PipelineExtra PipelineExtra `json:"pipelineExtra"`
	Tasks         []Task        `json:"tasks"`
}

type Pipeline struct {
	Id                  uint64         `json:"id"`
	EventTriggerId      uint64         `json:"eventTriggerId"`
	EventId             uint64         `json:"eventId"`
	TriggerDefinitionId uint64         `json:"triggerDefinitionId"`
	DefinitionName      string         `json:"definitionName"`
	DefinitionVersion   string         `json:"definitionVersion"`
	DefinitionCreater   string         `json:"definitionCreater"`
	Creater             string         `json:"creater"`
	Status              PipelineStatus `json:"status"`
	CostTimeSec         uint64         `json:"costTimeSec"`
	TimeBegin           *time.Time     `json:"timeBegin"`
	TimeEnd             *time.Time     `json:"timeEnd"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PipelineExtra struct {
	Id                       uint64                     `json:"id"`
	PipelineId               uint64                     `json:"pipelineId"`
	DefinitionContent        *PipelineVersionDefinition `json:"definitionContent"`
	EventContent             *EventDetail               `json:"eventContent"`
	EventTriggerContent      *EventTrigger              `json:"eventTriggerContent"`
	TriggerDefinitionContent *EventTriggerDefinition    `json:"triggerDefinitionContent"`
	Extra                    *PipelineExtraInfo         `json:"extra"`
	Contexts                 *PipelineExtraContents     `json:"contexts"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PipelineExtraInfo struct {
	StopReason string `json:"stopReason"`
}

type PipelineExtraContents struct {
}

type CallbackBody struct {
	TaskId     uint64            `json:"taskId"`
	PipelineId uint64            `json:"pipelineId"`
	Auth       string            `json:"auth"`
	Outputs    map[string]string `json:"outputs"`
}
