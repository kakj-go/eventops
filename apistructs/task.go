package apistructs

import (
	"time"
)

type TaskType string

const (
	K8sType    TaskType = "k8s"
	DockerType TaskType = "docker"
	OsType     TaskType = "os"
	PipeType   TaskType = "pipeline"
)

var TaskTypeList = []TaskType{K8sType, DockerType, OsType, PipeType}

func (t TaskType) String() string {
	return string(t)
}

func (t TaskType) Check() bool {
	var find = false
	for _, taskType := range TaskTypeList {
		if t == taskType {
			find = true
			break
		}
	}
	return find
}

type TaskStatus string

const InitTaskStatus TaskStatus = "initializing"
const CreatedTaskStatus TaskStatus = "created"
const RunningTaskStatus TaskStatus = "running"

const CancelTaskStatus TaskStatus = "cancel"
const FailedTaskStatus TaskStatus = "failed"
const SuccessTaskStatus TaskStatus = "success"
const UnKnowTaskStatus TaskStatus = "unknow"
const ErrorTaskStatus TaskStatus = "error"
const TimeoutTaskStatus TaskStatus = "timeout"

var DoneTaskStatuses = []TaskStatus{SuccessTaskStatus, FailedTaskStatus, CancelTaskStatus, UnKnowTaskStatus, ErrorTaskStatus, TimeoutTaskStatus}
var FailedTaskStatuses = []TaskStatus{FailedTaskStatus, CancelTaskStatus, UnKnowTaskStatus, ErrorTaskStatus, TimeoutTaskStatus}

func (taskStatus TaskStatus) IsDoneStatus() bool {
	for _, status := range DoneTaskStatuses {
		if status == taskStatus {
			return true
		}
	}
	return false
}

func (taskStatus TaskStatus) IsFailedStatus() bool {
	for _, status := range FailedTaskStatuses {
		if status == taskStatus {
			return true
		}
	}
	return false
}

type Task struct {
	Id           uint64     `json:"id"`
	PipelineId   uint64     `json:"pipelineId"`
	ParentTaskId uint64     `json:"parentTaskId"`
	JobSign      string     `json:"jobSign"`
	Alias        string     `json:"alias"`
	Type         TaskType   `json:"type"`
	Status       TaskStatus `json:"status"`
	Extra        *TaskExtra `json:"extra"`
	Outputs      Outputs    `json:"outputs"`
	CostTimeSec  uint64     `json:"costTimeSec"`
	TimeBegin    *time.Time `json:"timeBegin"`
	TimeEnd      *time.Time `json:"timeEnd"`
	Creater      string     `json:"creater"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskExtra struct {
	Error     string   `json:"error"`
	ChooseTag string   `json:"chooseTag"`
	Inputs    Inputs   `json:"inputs"`
	Contexts  Contexts `json:"contexts"`
}

type Inputs map[string]Input

type Input struct {
	Name  string    `json:"name"`
	Value string    `json:"value"`
	Type  ValueType `json:"type"`
}

type Contexts map[string]Context

type Context struct {
	Name  string    `json:"name"`
	Value string    `json:"value"`
	Type  ValueType `json:"type"`
}

type Outputs map[string]Output

type Output struct {
	Name  string    `json:"name"`
	Value string    `json:"value"`
	Type  ValueType `json:"type"`
}
