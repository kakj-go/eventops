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

package flowmanager

import (
	"context"
	"encoding/json"
	"eventops/apistructs"
	"eventops/conf"
	"eventops/internal/core/actuator"
	"eventops/internal/core/client/pipelineclient"
	"eventops/internal/core/client/taskclient"
	"eventops/pkg/dag"
	"eventops/pkg/limit_sync_group"
	"eventops/pkg/placeholder"
	"eventops/pkg/schema/pipeline"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	flow        *Flow
	flowManager *FlowManager

	parentTaskId   uint64
	taskDefinition *pipeline.Task

	image string

	runner actuator.Actuator
	job    *actuator.Job
}

func NewNode(flow *Flow, parentTaskId uint64, taskDefinition *pipeline.Task, image string) *Node {
	node := &Node{
		flow:        flow,
		flowManager: flow.flowManager,

		taskDefinition: taskDefinition,
		parentTaskId:   parentTaskId,

		image: image,
	}

	dbTask := node.getTask()
	flow.addNode(dbTask.Id, node)
	return node
}

func (node *Node) Run() error {
	if node.flow.getPipe().Status.IsEnd() {
		return nil
	}

	if node.getTask().Status.IsFailedStatus() {
		node.flow.lazyStopPipeline(apistructs.PipelineFailedStatus, fmt.Sprintf("node alias: %v parent_task_id: %v pipeline image: %v exec error %v",
			node.getTask().Alias, node.getTask().ParentTaskId, node.image, node.getTask().Extra.Error))
		return fmt.Errorf(node.getTask().Extra.Error)
	}

	var err error
	if node.getTask().Type != apistructs.PipeType {
		err = node.exec()
	} else {
		err = node.execPipelineTypeTask()
	}
	if err == nil {
		err = node.setContext()
	}
	if err != nil {
		taskUpdateError := node.setDbTask(WithStatus(apistructs.ErrorTaskStatus), WithExtraError(err.Error()))
		if taskUpdateError != nil {
			logrus.Errorf("task %v extra error: %v update failed: %v", node.getTask().Id, err, taskUpdateError)
		}

		node.flow.lazyStopPipeline(apistructs.PipelineFailedStatus, fmt.Sprintf("node alias: %v parent_task_id: %v pipeline image: %v exec error: %v",
			node.getTask().Alias, node.getTask().ParentTaskId, node.image, err))
		return err
	}

	node.runNextNodes()
	return nil
}

func (node *Node) execPipelineTypeTask() error {
	if node.getTask().Status.IsDoneStatus() {
		return nil
	}

	err := node.setPipelineTypeTaskInputs()
	if err != nil {
		return err
	}

	if node == node.flow.rootNode {
		return nil
	}

	definition, err := node.flow.getAndSetPipelineVersionDefinition(node.taskDefinition.Image)
	if err != nil {
		return err
	}
	nodes := definition.Dag.GetNextNodes(dag.Root)

	if err := node.setDbTask(WithStatus(apistructs.RunningTaskStatus)); err != nil {
		return err
	}
	worker := limit_sync_group.NewWorker(len(nodes))
	for index := range nodes {
		worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			parentTaskId := node.getTask().Id
			taskDefinition, err := node.createNextRunDbTask(parentTaskId, nodes[i[0].(int)], node.taskDefinition.Image)
			if err != nil {
				node.flow.lazyStopPipeline(apistructs.PipelineFailedStatus, err.Error())
				return err
			}

			nextNode := NewNode(node.flow, parentTaskId, taskDefinition, node.taskDefinition.Image)
			return nextNode.Run()
		}, index)
	}
	err = worker.Do().Error()
	if err == nil {
		err = node.setPipelineTypeTaskOutputs()
	}
	if err != nil {
		if node.flow.getPipe().Status == apistructs.PipelineCancelStatus {
			return node.setDbTask(WithStatus(apistructs.CancelTaskStatus))
		}
		return node.setDbTask(WithStatus(apistructs.ErrorTaskStatus))
	}

	var allTaskIsSuccessStatus = true
	for _, task := range definition.Tasks {
		task := node.flow.getTask(node.getTask().Id, task.Alias)
		if task == nil {
			allTaskIsSuccessStatus = false
			break
		}
		if task.Status != apistructs.SuccessTaskStatus {
			allTaskIsSuccessStatus = false
			break
		}
	}

	if allTaskIsSuccessStatus {
		return node.setDbTask(WithStatus(apistructs.SuccessTaskStatus))
	} else {
		if node.flow.getPipe().Status == apistructs.PipelineCancelStatus {
			return node.setDbTask(WithStatus(apistructs.CancelTaskStatus))
		}
		return node.setDbTask(WithStatus(apistructs.FailedTaskStatus))
	}
}

func (node *Node) runNextNodes() {
	definition, err := node.flow.getAndSetPipelineVersionDefinition(node.image)
	if err != nil {
		node.flow.lazyStopPipeline(apistructs.PipelineFailedStatus, err.Error())
		return
	}

	nextNodes := definition.Dag.GetNextNodes(node.taskDefinition.Alias)
	var wait sync.WaitGroup
	for index, dagNode := range nextNodes {
		if !node.allNeedNodeIsDone(dagNode) {
			continue
		}
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			parentTaskId := node.getTask().ParentTaskId
			taskDefinition, err := node.createNextRunDbTask(parentTaskId, nextNodes[index], node.image)
			if err != nil {
				node.flow.lazyStopPipeline(apistructs.PipelineFailedStatus, err.Error())
				return
			}

			nextNode := NewNode(node.flow, parentTaskId, taskDefinition, node.image)
			_ = nextNode.Run()
		}(index)
	}
	wait.Wait()
}

func (node *Node) createNextRunDbTask(parentTaskId uint64, dagNode pipeline.Node, image string) (*pipeline.Task, error) {
	// 检测是否存在环
	if !node.flow.checkDag(fmt.Sprintf("%v_%v->%v_%v", node.getTask().ParentTaskId, node.getTask().Alias, parentTaskId, dagNode.Name)) {
		return nil, fmt.Errorf("[dag] node alias: %v, parent_task_id: %v -> node alias: %v, parent_task_id: %v has cycle",
			node.getTask().Alias, node.getTask().ParentTaskId, dagNode.Name, parentTaskId)
	}

	definition, err := node.flow.getAndSetPipelineVersionDefinition(image)
	if err != nil {
		return nil, err
	}

	taskDefinition := definition.GetTaskByAlias(dagNode.Name)
	if taskDefinition == nil {
		return nil, fmt.Errorf("not find task alias: %v definition in pipeline image: %v", dagNode.Name, image)
	}

	if node.flow.getTask(parentTaskId, dagNode.Name) == nil {
		var extra = taskclient.TaskExtra{
			Auth: getRandomString(32),
		}
		dbTask := &taskclient.Task{
			PipelineId:   node.flow.getPipe().Id,
			Alias:        dagNode.Name,
			Type:         taskDefinition.Type,
			Status:       apistructs.InitTaskStatus,
			Extra:        &extra,
			Outputs:      &taskclient.Outputs{},
			CostTimeSec:  0,
			Creater:      node.flow.getPipe().Creater,
			ParentTaskId: parentTaskId,
		}

		_, err := node.flowManager.clientManager.taskClient.CreateTask(nil, dbTask)
		if err != nil {
			return nil, err
		}

		node.flow.addTask(dbTask)
	}

	return taskDefinition, nil
}

func getRandomString(n int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	for i := 0; i < n; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

func (node *Node) buildReplaceValue() (*placeholder.ReplaceValue, error) {
	var replaceValue = placeholder.ReplaceValue{}

	inputs, err := node.getPlaceholderInputsValue()
	if err != nil {
		return nil, err
	}
	replaceValue.Inputs = inputs
	replaceValue.Contexts = node.getPlaceholderContextValue()

	outputs, err := node.getPlaceholderOutputValue()
	if err != nil {
		return nil, err
	}
	replaceValue.Outputs = outputs

	replaceValue.TaskId = node.getTask().Id
	replaceValue.PipelineId = node.flow.getPipe().Id

	return &replaceValue, nil
}

func (node *Node) getTask() *taskclient.Task {
	return node.flow.getTask(node.parentTaskId, node.taskDefinition.Alias)
}

func (node *Node) setTask(opts ...Opt) {
	node.flow.setTask(node.getTask().ParentTaskId, node.getTask().Alias, opts...)
}

func (node *Node) setDbTask(opts ...Opt) error {
	node.flow.setTask(node.getTask().ParentTaskId, node.getTask().Alias, opts...)

	_, taskUpdateError := node.flowManager.clientManager.taskClient.UpdateTask(nil, node.getTask())
	if taskUpdateError != nil {
		return taskUpdateError
	}
	return nil
}

func (node *Node) setPipelineTypeTaskInputs() error {
	if node.taskDefinition.Type != apistructs.PipeType {
		return nil
	}

	var taskInput = taskclient.Inputs{}
	if node == node.flow.rootNode {
		inputValues, err := node.getPlaceholderInputsValue()
		if err != nil {
			return err
		}
		definition, err := node.flow.getAndSetPipelineVersionDefinition(node.image)
		if err != nil {
			return err
		}
		for _, input := range definition.Inputs {
			taskInput[input.Name] = apistructs.Input{
				Name:  input.Name,
				Value: inputValues[input.Name].Value,
				Type:  input.Type,
			}
		}
	} else {
		replaceValue, err := node.buildReplaceValue()
		if err != nil {
			return err
		}

		definition, err := node.flow.getAndSetPipelineVersionDefinition(node.taskDefinition.Image)
		if err != nil {
			return err
		}
		var inputNameValue = map[string]pipeline.Input{}
		for _, definitionInput := range definition.Inputs {
			inputNameValue[definitionInput.Name] = definitionInput
		}

		for _, pipeInput := range node.taskDefinition.Inputs {
			taskInput[pipeInput.Name] = apistructs.Input{
				Name:  pipeInput.Name,
				Value: placeholder.ReplacePlaceholder(pipeInput.Value, replaceValue, false),
				Type:  inputNameValue[pipeInput.Name].Type,
			}
		}
	}

	node.setTask(WithExtraInputs(taskInput))
	return nil
}

func (node *Node) setPipelineTypeTaskOutputs() error {
	if node == node.flow.rootNode {
		return nil
	}

	if node.taskDefinition.Type != apistructs.PipeType {
		return nil
	}

	definition, err := node.flow.getAndSetPipelineVersionDefinition(node.taskDefinition.Image)
	if err != nil {
		return err
	}

	var definitionOutputMap = make(map[string]string, len(definition.Outputs))
	for _, output := range definition.Outputs {
		definitionOutputMap[output.Name] = output.Value
	}

	var outputs = make(taskclient.Outputs, len(node.taskDefinition.Outputs))
	for _, pipelineTaskOutput := range node.taskDefinition.Outputs {
		realValue := definitionOutputMap[pipelineTaskOutput.Value]
		_ = placeholder.MatchHolderFromHandler(realValue, map[placeholder.Type]placeholder.Handler{
			placeholder.OutputType: func(placeholder string, values ...string) error {
				taskName := values[1]
				taskOutputName := values[2]

				outputTask := node.flow.getTask(node.getTask().Id, taskName)

				if outputTask.Outputs != nil {
					for _, taskOutput := range *outputTask.Outputs {
						if taskOutput.Name != taskOutputName {
							continue
						}
						outputs[pipelineTaskOutput.Name] = apistructs.Output{
							Name:  pipelineTaskOutput.Name,
							Value: taskOutput.Value,
							Type:  taskOutput.Type,
						}
					}
				}
				return nil
			},
		})
	}
	node.setTask(WithExtraOutputs(outputs))
	return nil
}

func (node *Node) setContext() error {

	parentNode := node.flow.getNode(node.parentTaskId)

	taskOutputs := node.getTask().Outputs
	definitionOutputs := node.taskDefinition.Outputs

	var contexts apistructs.Contexts
	if parentNode == node.flow.rootNode {
		contexts = apistructs.Contexts(*node.flow.getPipeExtra().Contexts)
	} else {
		contexts = apistructs.Contexts(parentNode.getTask().Extra.Contexts)
	}
	if contexts == nil {
		contexts = map[string]apistructs.Context{}
	}

	for _, definitionOutput := range definitionOutputs {
		if definitionOutput.SetToContext == "" {
			continue
		}
		if taskOutputs != nil {
			value := *taskOutputs
			output := value[definitionOutput.Name]
			contexts[definitionOutput.SetToContext] = apistructs.Context{
				Name:  definitionOutput.SetToContext,
				Value: output.Value,
				Type:  output.Type,
			}
		}
	}

	if parentNode == node.flow.rootNode {
		extra := node.flow.getPipeExtra()
		pipelineContexts := pipelineclient.PipelineExtraContents(contexts)
		extra.Contexts = &pipelineContexts
		return node.flow.setDbPipeExtra(extra)
	} else {
		return parentNode.setDbTask(WithExtraContexts(taskclient.Contexts(contexts)))
	}
}

func (node *Node) allNeedNodeIsDone(dagNode pipeline.Node) bool {
	var allDone = true
	for _, need := range dagNode.Needs {
		if need == dag.Root {
			continue
		}

		task := node.flow.getTask(node.parentTaskId, need)
		if !task.Status.IsDoneStatus() {
			allDone = false
			break
		}
	}
	return allDone
}

func (node *Node) exec() error {
	if node.getTask().Status.IsDoneStatus() {
		return nil
	}

	if node.job == nil {
		job, err := node.buildActuatorJob()
		if err != nil {
			return err
		}
		node.job = job
	}

	if node.runner == nil {
		runner, chooseTag, err := node.actuator()
		if err != nil {
			return err
		}
		node.runner = runner
		node.setTask(WithExtraTag(chooseTag))
	}

	var waitTime = 1
	switch node.getTask().Status {
	case apistructs.InitTaskStatus:
		err := node.flowManager.clientManager.db.Transaction(func(tx *gorm.DB) error {
			createJob, err := node.runner.Create(node.flow.ctx, node.job)
			if err != nil {
				return err
			}

			return node.setDbTask(WithStatus(apistructs.CreatedTaskStatus), WithJobSign(createJob.JobSign))
		})
		if err != nil {
			return err
		}
		fallthrough
	case apistructs.CreatedTaskStatus:
		err := node.flowManager.clientManager.db.Transaction(func(tx *gorm.DB) error {
			err := node.runner.Start(node.flow.ctx, node.job)
			if err != nil {
				return err
			}

			return node.setDbTask(WithStatus(apistructs.RunningTaskStatus), WithJobSign(node.job.JobSign))
		})
		if err != nil {
			return err
		}
		fallthrough
	case apistructs.RunningTaskStatus:
		for {
			select {
			case <-node.flow.ctx.Done():
				return node.flowManager.clientManager.db.Transaction(func(tx *gorm.DB) error {
					err := node.setDbTask(WithStatus(apistructs.CancelTaskStatus))
					if err != nil {
						return err
					}

					return node.runner.Cancel(context.Background(), node.job)
				})
			case <-time.After(time.Duration(waitTime) * time.Second):
				if waitTime < 10 {
					waitTime = waitTime + 2
				}

				value := int64(time.Now().Sub(*node.getTask().TimeBegin) / time.Second)
				if value > node.taskDefinition.Timeout {
					return node.flowManager.clientManager.db.Transaction(func(tx *gorm.DB) error {
						err := node.setDbTask(WithStatus(apistructs.TimeoutTaskStatus))
						if err != nil {
							return err
						}

						return node.runner.Cancel(context.Background(), node.job)
					})
				}

				status, err := node.runner.Status(node.flow.ctx, node.job)
				if err != nil {
					return err
				}

				if node.getTask().Status.IsDoneStatus() {
					return nil
				}

				if status != node.getTask().Status {
					var opts = []Opt{WithStatus(status)}
					if node.job.Error != "" {
						opts = append(opts, WithExtraError(node.job.Error))
					}

					err := node.setDbTask(opts...)
					if err != nil {
						return err
					}
				}

				if status.IsDoneStatus() {
					return nil
				}
			}
		}
	}

	return nil
}

func buildMinioAlias(user string, pipelineId uint64, taskId uint64) string {
	return fmt.Sprintf("%v_%v_%v", user, pipelineId, taskId)
}

func buildMinioUploadPath(pipelineId, taskId uint64, name string) string {
	return fmt.Sprintf("pipeline-%v/task-%v/%v", pipelineId, taskId, name)
}

func buildMinioPath(alias string, name string) string {
	return fmt.Sprintf("%v/%v/%v", alias, conf.GetMinio().BasePath, name)
}

func (node *Node) buildActuatorJob() (*actuator.Job, error) {
	job := actuator.Job{
		PipelineId:     strconv.FormatUint(node.getTask().PipelineId, 10),
		TaskId:         strconv.FormatUint(node.getTask().Id, 10),
		DefinitionTask: node.taskDefinition,
		JobSign:        node.getTask().JobSign,
	}

	replaceValue, err := node.buildReplaceValue()
	if err != nil {
		return nil, err
	}

	var newCommands []string
	for _, command := range node.taskDefinition.Commands {
		newCommands = append(newCommands, placeholder.ReplacePlaceholder(command, replaceValue, true))
	}

	var minioAlias = buildMinioAlias(node.getTask().Creater, node.getTask().PipelineId, node.getTask().Id)

	// 入参出参和全局变量的值构建成 minio 的下载文件的命令，文件下载到对应的目录上
	var preCommands []string
	for _, input := range replaceValue.Inputs {
		if input.Type != apistructs.FileType {
			continue
		}
		minioPath := buildMinioPath(minioAlias, input.Value)
		localPath := placeholder.MakeRealFilePath(node.getTask().PipelineId, node.getTask().Id, input.Name, placeholder.InputType.String())
		preCommands = append(preCommands, fmt.Sprintf("mc cp %v %v", minioPath, localPath))
	}
	for _, output := range replaceValue.Outputs {
		if output.Type != apistructs.FileType {
			continue
		}

		minioPath := buildMinioPath(minioAlias, output.Value)
		localPath := placeholder.MakeRealFilePath(node.getTask().PipelineId, node.getTask().Id, output.Name, placeholder.OutputType.String())
		preCommands = append(preCommands, fmt.Sprintf("mc cp %v %v", minioPath, localPath))
	}
	for _, ctx := range replaceValue.Contexts {
		if ctx.Type != apistructs.FileType {
			continue
		}

		minioPath := buildMinioPath(minioAlias, ctx.Value)
		localPath := placeholder.MakeRealFilePath(node.getTask().PipelineId, node.getTask().Id, ctx.Name, placeholder.ContextType.String())
		preCommands = append(preCommands, fmt.Sprintf("mc cp %v %v", minioPath, localPath))
	}

	// 如果存在下载命令，则在最前面构建 mc alias 命令
	var minioAliasServerCommand = fmt.Sprintf("mc alias set %v %v %v %v", minioAlias, conf.GetMinio().Server, conf.GetMinio().AccessKeyId, conf.GetMinio().SecretAccessKey)
	if len(preCommands) > 0 {
		preCommands = append([]string{minioAliasServerCommand}, preCommands...)
	}

	// 如果出参是文件类型，则构建 mc 上传命令
	var nextCommands []string
	var curlOutputs = map[string]string{}
	for _, output := range node.taskDefinition.Outputs {
		if output.Type == apistructs.FileType {
			minioPath := buildMinioPath(minioAlias, buildMinioUploadPath(node.getTask().PipelineId, node.getTask().Id, output.Name))
			localPath := output.Value
			nextCommands = append(nextCommands, fmt.Sprintf("mc cp %v %v", localPath, minioPath))
			curlOutputs[output.Name] = buildMinioUploadPath(node.getTask().PipelineId, node.getTask().Id, output.Name)
		} else {
			curlOutputs[output.Name] = fmt.Sprintf("${%s}", output.Value)
		}
	}

	// 如果存在上传命令，则在最前面构建 mc alias 命令
	if len(nextCommands) > 0 {
		nextCommands = append([]string{minioAliasServerCommand}, nextCommands...)
	}

	// 只要存在 mc 命令就最后将别名去除
	if len(preCommands) > 0 || len(nextCommands) > 0 {
		nextCommands = append(nextCommands, fmt.Sprintf("mc alias remove %v", minioAlias))
	}

	var callback apistructs.CallbackBody
	callback.PipelineId = node.getTask().PipelineId
	callback.TaskId = node.getTask().Id
	callback.Auth = node.getTask().Extra.Auth
	callback.Outputs = curlOutputs
	callbackJson, err := json.Marshal(callback)
	if err != nil {
		return nil, err
	}
	callbackJson, err = json.Marshal(string(callbackJson))
	if err != nil {
		return nil, err
	}

	nextCommands = append(nextCommands, fmt.Sprintf("curl -o response.txt %v/api/pipeline/callback -X POST -d %v --header \"Content-Type: application/json\"", conf.GetCallbackAddress(), string(callbackJson)))
	nextCommands = append(nextCommands, fmt.Sprintf("if [ '\"success\"' != `cat response.txt` ]; then cat response.txt && exit 1; fi;"))

	job.PreCommands = preCommands
	job.DefinitionTask.Commands = newCommands
	job.NextCommands = nextCommands
	return &job, nil
}
