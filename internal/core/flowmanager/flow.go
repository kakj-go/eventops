package flowmanager

import (
	"context"
	"eventops/apistructs"
	"eventops/internal/core/client/pipelineclient"
	"eventops/internal/core/client/taskclient"
	"eventops/pkg/dag"
	"eventops/pkg/schema/pipeline"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"sync"
	"time"
)

type Flow struct {
	ctx         context.Context
	cancel      func()
	lock        sync.Mutex
	flowManager *FlowManager

	stopOnce     sync.Once
	lazyStopFunc func()

	rootNode *Node
	nodes    map[uint64]*Node

	dagCheckMap map[string]struct{}

	dbTasks     map[string]*taskclient.Task
	dbPipe      *pipelineclient.Pipeline
	dbPipeExtra *pipelineclient.PipelineExtra

	pipelineDefinitions map[string]*pipeline.Pipeline
}

func newFlow(flowManager *FlowManager, dbPipe *pipelineclient.Pipeline, dbPipeExtra *pipelineclient.PipelineExtra) (*Flow, error) {
	var version = dbPipeExtra.DefinitionContent

	var pipelineDefinition pipeline.Pipeline
	err := yaml.Unmarshal([]byte(version.Content), &pipelineDefinition)
	if err != nil {
		return nil, err
	}

	var image = pipeline.BuildImage(version.Name, version.Creater, version.Version)
	newCtx, canFunc := context.WithCancel(context.Background())

	runPipeline := &Flow{
		ctx:         newCtx,
		cancel:      canFunc,
		lock:        sync.Mutex{},
		dagCheckMap: map[string]struct{}{},
		dbTasks:     map[string]*taskclient.Task{},
		nodes:       map[uint64]*Node{},
		flowManager: flowManager,
		pipelineDefinitions: map[string]*pipeline.Pipeline{
			image: &pipelineDefinition,
		},
		dbPipeExtra: dbPipeExtra,
		dbPipe:      dbPipe,
	}

	rootTask := taskclient.Task{
		Alias:        dag.Root,
		ParentTaskId: 0,
		Extra:        &taskclient.TaskExtra{},
		Outputs:      &taskclient.Outputs{},
		Type:         apistructs.PipeType,
		Id:           0,
		Status:       apistructs.SuccessTaskStatus,
	}
	runPipeline.addTask(&rootTask)

	runPipeline.rootNode = NewNode(runPipeline,
		0,
		&pipeline.Task{
			Alias: dag.Root,
			Image: image,
			Type:  apistructs.PipeType,
		}, image)

	return runPipeline, nil
}

func (p *Flow) checkDag(vectorKey string) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, find := p.dagCheckMap[vectorKey]
	if find {
		return false
	}
	p.dagCheckMap[vectorKey] = struct{}{}
	return true
}

func buildTaskKey(parentTaskId uint64, taskAlias string) string {
	return fmt.Sprintf("%v-%v", parentTaskId, taskAlias)
}

type Opt func(task *taskclient.Task)

func WithStatus(status apistructs.TaskStatus) Opt {
	return func(task *taskclient.Task) {
		task.Status = status

		if task.TimeBegin == nil {
			task.TimeBegin = &[]time.Time{time.Now()}[0]
		}

		if status.IsDoneStatus() {
			task.TimeEnd = &[]time.Time{time.Now()}[0]
			task.CostTimeSec = uint64(task.TimeEnd.Sub(*task.TimeBegin) / time.Second)
		} else {
			task.CostTimeSec = uint64(time.Now().Sub(*task.TimeBegin) / time.Second)
		}
	}
}

func WithExtraError(err string) Opt {
	return func(task *taskclient.Task) {
		task.Extra.Error = err
	}
}

func WithExtraTag(tag string) Opt {
	return func(task *taskclient.Task) {
		task.Extra.ChooseTag = tag
	}
}

func WithExtraInputs(inputs taskclient.Inputs) Opt {
	return func(task *taskclient.Task) {
		if task.Extra.Inputs == nil {
			task.Extra.Inputs = inputs
			return
		}

		for key, input := range inputs {
			task.Extra.Inputs[key] = input
		}
	}
}

func WithExtraOutputs(outputs taskclient.Outputs) Opt {
	return func(task *taskclient.Task) {
		if task.Outputs == nil {
			task.Outputs = &outputs
			return
		}

		taskOutputs := *task.Outputs
		for key, output := range outputs {
			taskOutputs[key] = output
		}
	}
}

func WithExtraContexts(contexts taskclient.Contexts) Opt {
	return func(task *taskclient.Task) {
		if task.Extra.Contexts == nil {
			task.Extra.Contexts = contexts
			return
		}

		for key, ctxValue := range contexts {
			task.Extra.Contexts[key] = ctxValue
		}
	}
}

func WithJobSign(jobSign string) Opt {
	return func(task *taskclient.Task) {
		task.JobSign = jobSign
	}
}

func (p *Flow) setTask(parentTaskId uint64, taskAlias string, opts ...Opt) {
	p.lock.Lock()
	defer p.lock.Unlock()

	task := p.dbTasks[buildTaskKey(parentTaskId, taskAlias)]
	for _, opt := range opts {
		opt(task)
	}

	p.dbTasks[buildTaskKey(parentTaskId, taskAlias)] = task
}

func (p *Flow) getTask(parentTaskId uint64, taskAlias string) *taskclient.Task {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.dbTasks[buildTaskKey(parentTaskId, taskAlias)]
}

func (p *Flow) addTask(task *taskclient.Task) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.dbTasks[buildTaskKey(task.ParentTaskId, task.Alias)] = task
}

func (p *Flow) getNode(id uint64) *Node {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.nodes[id]
}

func (p *Flow) addNode(id uint64, node *Node) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.nodes[id] = node
}

func (p *Flow) getPipe() *pipelineclient.Pipeline {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.dbPipe
}

func (p *Flow) getPipeExtra() *pipelineclient.PipelineExtra {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.dbPipeExtra
}

func (p *Flow) setDbPipeExtra(extra *pipelineclient.PipelineExtra) error {
	_, err := p.flowManager.clientManager.pipelineClient.UpdatePipelineExtra(nil, extra)
	if err != nil {
		return err
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	p.dbPipeExtra = extra
	return nil
}

func (p *Flow) setDbPipe(pipe *pipelineclient.Pipeline) error {
	_, err := p.flowManager.clientManager.pipelineClient.UpdatePipeline(nil, pipe)
	if err != nil {
		return err
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	p.dbPipe = pipe
	return nil
}

// 当 task 是 pipeline 类型的时候，动态查询定义可能会出现报错。
// 如果全部定义都存储将会浪费存储存储。
func (p *Flow) getAndSetPipelineVersionDefinition(image string) (*pipeline.Pipeline, error) {
	p.lock.Lock()
	definition, find := p.pipelineDefinitions[image]
	p.lock.Unlock()
	if find {
		return definition, nil
	}

	dbDefinition, find, err := p.flowManager.clientManager.pipelineDefinitionClient.GetPipelineVersionDefinition(nil,
		pipeline.GetImageName(image), pipeline.GetImageVersion(image), pipeline.GetImageCreater(image))
	if err != nil {
		return nil, err
	}
	if !find {
		return nil, fmt.Errorf("not find definition image: %v", image)
	}

	var pipelineDefinition pipeline.Pipeline
	err = yaml.Unmarshal([]byte(dbDefinition.Content), &pipelineDefinition)
	if err != nil {
		return nil, err
	}

	p.lock.Lock()
	p.pipelineDefinitions[image] = &pipelineDefinition
	p.lock.Unlock()
	return &pipelineDefinition, nil
}

func (p *Flow) runLazyStopFunc() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.lazyStopFunc != nil {
		p.lazyStopFunc()
		p.lazyStopFunc = nil
	}
}

func (p *Flow) run() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			p.lazyStopPipeline(apistructs.PipelineFailedStatus, "panic error")
		}
		p.lazyStopPipeline(apistructs.PipelineSuccessStatus, "")
		p.runLazyStopFunc()
		logrus.Debugf("pipeline %v stop", p.getPipe().Id)
	}()

	go func() {
		select {
		case <-p.ctx.Done():
			return
		case <-time.After(24 * time.Hour):
			p.lazyStopPipeline(apistructs.PipelineFailedStatus, "timeout")
		}
	}()

	dbTasks, err := p.flowManager.clientManager.taskClient.ListTasks(nil, p.getPipe().Id, p.getPipe().Creater)
	if err != nil {
		p.lazyStopPipeline(apistructs.PipelineFailedStatus, err.Error())
		return
	}
	for _, task := range dbTasks {
		p.addTask(task)
	}

	if p.getPipe().TimeBegin == nil {
		pipe := p.getPipe()
		pipe.TimeBegin = &[]time.Time{time.Now()}[0]
		err := p.setDbPipe(pipe)
		if err != nil {
			p.lazyStopPipeline(apistructs.PipelineFailedStatus, err.Error())
			return
		}
	}
	err = p.rootNode.Run()
	if err != nil {
		logrus.Debugf("pipeline: %v run error: %v", p.getPipe().Id, err)
	}
}

func (p *Flow) lazyStopPipeline(status apistructs.PipelineStatus, stopReason string) {
	p.lazyStopPipelineWithCallback(status, stopReason, nil)
}

func (p *Flow) lazyStopPipelineWithCallback(status apistructs.PipelineStatus, stopReason string, callback func()) {
	if !status.IsEnd() {
		return
	}

	p.stopOnce.Do(func() {
		p.lock.Lock()
		p.lazyStopFunc = func() {
			p.dbPipe.Status = status
			p.dbPipeExtra.Extra.StopReason = stopReason
			if p.dbPipe.Status.IsEnd() {
				if p.dbPipe.TimeEnd == nil {
					p.dbPipe.TimeEnd = &[]time.Time{time.Now()}[0]
				}
				p.dbPipe.CostTimeSec = uint64(p.dbPipe.TimeEnd.Sub(*p.dbPipe.TimeBegin) / time.Second)
			}
			err := p.flowManager.clientManager.db.Transaction(func(tx *gorm.DB) error {
				_, err := p.flowManager.clientManager.pipelineClient.UpdatePipelineExtra(tx, p.dbPipeExtra)
				if err != nil {
					return err
				}

				_, err = p.flowManager.clientManager.pipelineClient.UpdatePipeline(tx, p.dbPipe)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				logrus.Errorf("failed to update pipeline status: %v error: %v", apistructs.PipelineFailedStatus, err.Error())
			}

			if callback != nil {
				callback()
			}
		}
		p.lock.Unlock()
		p.cancel()
	})
}

func (p *Flow) clear() {

}
