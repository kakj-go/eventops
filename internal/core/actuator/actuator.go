package actuator

import (
	"context"
	"eventops/apistructs"
	"eventops/pkg/schema/pipeline"
	"fmt"
)

type Actuator interface {
	Type() apistructs.TaskType

	Create(context.Context, *Job) (*Job, error)
	Start(context.Context, *Job) error
	Cancel(context.Context, *Job) error
	Status(context.Context, *Job) (apistructs.TaskStatus, error)

	Remove(context.Context, *Job) error
	Exist(context.Context, *Job) (bool, error)
}

var JobNotFindError = fmt.Errorf("task not find")

type Job struct {
	PipelineId string
	TaskId     string

	PreCommands    []string
	DefinitionTask *pipeline.Task
	NextCommands   []string

	JobSign string
	Error   string
}
