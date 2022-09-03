/*
 * Copyright (c) 2021 Terminus, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
