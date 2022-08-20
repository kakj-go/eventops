package actuator

import (
	"context"
	"fmt"
	"tiggerops/pkg/schema/pipeline"
	"time"
)

type Actuator interface {
	Create(context.Context, *Task) (*Task, error)
	Start(context.Context, *Task) error
	Remove(context.Context, *Task) error
	Cancel(context.Context, *Task) error
	Exist(context.Context, *Task) (bool, error)
	Status(context.Context, *Task) (TaskStatus, error)
}

var TaskNotFindError error = fmt.Errorf("task not find")

type TaskStatus string

const CreatedTaskStatus TaskStatus = "created"

const RunningTaskStatus TaskStatus = "running"
const SuccessTaskStatus TaskStatus = "success"
const FailedTaskStatus TaskStatus = "failed"
const CancelTaskStatus TaskStatus = "cancel"

const UnKnowTaskStatus TaskStatus = "unknow"
const ErrorTaskStatus TaskStatus = "error"
const PausedTaskStatus TaskStatus = "paused"

type Task struct {
	Id           uint64
	PipelineID   uint64
	Sign         string
	InstanceSign string

	DefinitionTask     pipeline.Task
	DefinitionPipeline pipeline.Pipeline

	Status  TaskStatus
	RunUser string
	Outputs []pipeline.Output
	Errors  []string
	Waring  []string

	TimeBegin   time.Time `json:"time_begin"`
	TimeEnd     time.Time `json:"time_end"`
	TimeCreated time.Time `json:"time_Created"`
	TimeUpdated time.Time `json:"timeUpdated"`
}
