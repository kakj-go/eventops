package pipeline

import (
	"eventops/apistructs"
	"eventops/pkg/dag"
	"eventops/pkg/placeholder"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type Pipeline struct {
	Version          string           `yaml:"version,omitempty"`
	Name             string           `yaml:"name,omitempty"`
	ActuatorSelector ActuatorSelector `yaml:"actuatorSelector,omitempty"`
	Inputs           []Input          `yaml:"inputs,omitempty"`
	Contexts         []Context        `yaml:"contexts,omitempty"`
	Dag              Dag              `yaml:"dag,omitempty"`
	Tasks            []Task           `yaml:"tasks,omitempty"`
	Outputs          []Output         `yaml:"outputs,omitempty"`
}

func BuildImage(name string, creater string, version string) string {
	return fmt.Sprintf("%v/%v:%v", creater, name, version)
}

func (p *Pipeline) GetPipelineTypeTask() (pipelineTask []Task) {
	for _, task := range p.Tasks {
		if task.Type != apistructs.PipeType {
			continue
		}
		pipelineTask = append(pipelineTask, task)
	}

	return pipelineTask
}

func (p *Pipeline) GetTaskByAlias(alias string) (pipelineTask *Task) {
	for _, task := range p.Tasks {
		if task.Alias == alias {
			return &task
		}
	}
	return nil
}

func (p *Pipeline) Mutating(pipelineTypeTaskDefinitionMap map[string]Pipeline) error {
	err := p.pipelineTypeTaskOutputTypeMutating(pipelineTypeTaskDefinitionMap)
	if err != nil {
		return err
	}

	err = p.pipelineOutputTypeMutating()
	if err != nil {
		return err
	}

	p.taskTimeoutMutating()
	return nil
}

func (p *Pipeline) PipelineTypeTaskImageMutating(creater string) {
	for index, task := range p.Tasks {
		if task.Type != apistructs.PipeType {
			continue
		}

		imageCreater := task.GetPipelineCreater()
		if strings.TrimSpace(imageCreater) == "" {
			p.Tasks[index].Image = fmt.Sprintf("%s%s%s", creater, ImageCreaterNameSplitWord, p.Tasks[index].Image)
		}

		imageVersion := task.GetPipelineVersion()
		if strings.TrimSpace(imageVersion) == "" {
			p.Tasks[index].Image = fmt.Sprintf("%s%s%s", p.Tasks[index].Image, ImageNameVersionSplitWord, apistructs.LatestVersion)
		}
	}
}

func (p *Pipeline) pipelineTypeTaskOutputTypeMutating(pipelineTypeTaskDefinitionMap map[string]Pipeline) error {
	var mapKeyBuild = func(image, outputName string) string {
		return fmt.Sprintf("%s-%s", image, outputName)
	}

	var pipelineOutputMap = map[string]Output{}
	for key, pipeline := range pipelineTypeTaskDefinitionMap {
		for _, output := range pipeline.Outputs {
			pipelineOutputMap[mapKeyBuild(key, output.Name)] = output
		}
	}

	for index, task := range p.Tasks {
		if task.Type != apistructs.PipeType {
			continue
		}
		for outputIndex, output := range task.Outputs {
			pipelineOutput, ok := pipelineOutputMap[mapKeyBuild(task.Image, output.Name)]
			if !ok {
				continue
			}
			p.Tasks[index].Outputs[outputIndex].Type = pipelineOutput.Type
		}
	}
	return nil
}

func (p *Pipeline) pipelineOutputTypeMutating() error {
	var mapKeyBuild = func(taskName, outputName string) string {
		return fmt.Sprintf("%s-%s", taskName, outputName)
	}

	var taskOutputMap = map[string]Output{}
	for _, task := range p.Tasks {
		for _, output := range task.Outputs {
			taskOutputMap[mapKeyBuild(task.Alias, output.Name)] = output
		}
	}

	for index, output := range p.Outputs {
		if output.Value == "" {
			continue
		}
		err := placeholder.MatchHolderFromHandler(output.Value, map[placeholder.Type]placeholder.Handler{
			placeholder.OutputType: func(placeholder string, values ...string) error {
				taskName := values[1]
				taskOutputName := values[2]

				output, ok := taskOutputMap[mapKeyBuild(taskName, taskOutputName)]
				if !ok {
					return nil
				}
				p.Outputs[index].Type = output.Type
				return nil
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pipeline) taskTimeoutMutating() {

	for index, task := range p.Tasks {
		if task.Type == apistructs.PipeType {
			continue
		}
		if task.Timeout <= 0 {
			p.Tasks[index].Timeout = 3600
		}
	}
}

func (p *Pipeline) Check(yamlContent string, pipelineTypeTaskDefinitionMap map[string]Pipeline) error {
	if err := p.checkFormat(); err != nil {
		return err
	}

	if err := p.checkPlaceholder(yamlContent, pipelineTypeTaskDefinitionMap); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) checkFormat() error {
	if err := p.checkVersion(); err != nil {
		return err
	}

	if len(p.Name) == 0 {
		return fmt.Errorf("pipeline name can not empty")
	}

	if err := p.ActuatorSelector.check(); err != nil {
		return err
	}

	if err := p.checkInput(); err != nil {
		return err
	}

	if err := p.checkContext(); err != nil {
		return err
	}

	if err := p.Dag.Check(); err != nil {
		return err
	}

	if err := p.checkTasks(); err != nil {
		return err
	}

	if err := p.checkDagTask(); err != nil {
		return err
	}

	if err := p.checkOutput(); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) checkInput() error {
	for _, input := range p.Inputs {
		if err := input.check(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) checkContext() error {
	if p.Contexts == nil {
		return nil
	}

	for _, context := range p.Contexts {
		if err := context.check(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) checkOutput() error {
	if p.Outputs == nil {
		return nil
	}

	for _, output := range p.Outputs {
		if output.Name == "" {
			return fmt.Errorf("output name can not empty")
		}
		if output.Value == "" {
			return fmt.Errorf("output name %v value can not empty", output.Name)
		}

		if output.Type != "" {
			err := output.Type.ValueTypeCheck()
			if err != nil {
				return fmt.Errorf("output name %v type check error %v", output.Name, err)
			}
		}
	}
	return nil
}

func (p *Pipeline) checkVersion() error {
	if len(strings.TrimSpace(p.Version)) == 0 {
		return fmt.Errorf("pipeline version can not empty")
	}

	if p.Version == apistructs.LatestVersion {
		return fmt.Errorf("version can not use latest")
	}
	return nil
}

func (p *Pipeline) checkTasks() error {
	if len(p.Tasks) == 0 {
		return fmt.Errorf("pipeline tasks can not empty")
	}

	var taskAliasOnly = map[string]bool{}
	for _, task := range p.Tasks {
		if err := task.Check(p.Contexts); err != nil {
			return err
		}

		find := taskAliasOnly[task.Alias]
		if find {
			return fmt.Errorf("task alias %s not only", task.Alias)
		} else {
			taskAliasOnly[task.Alias] = true
		}
	}
	return nil
}

func (p *Pipeline) checkDagTask() error {
	var findNeedRoot = false
	for _, node := range p.Dag {
		for _, need := range node.Needs {
			if need == dag.Root {
				findNeedRoot = true
			}
		}
	}
	if !findNeedRoot {
		return fmt.Errorf("dag needs should have root dependent")
	}

	var dagMap = map[string]interface{}{}
	for _, node := range p.Dag {
		dagMap[node.Name] = node.Name
	}

	var notInDagTask []string
	for _, task := range p.Tasks {
		if dagMap[task.Alias] == nil {
			notInDagTask = append(notInDagTask, task.Alias)
		}
	}

	if len(notInDagTask) > 0 {
		return fmt.Errorf("task %v not described depends in dag", notInDagTask)
	}
	return nil
}

func (p *Pipeline) checkPlaceholder(pipelineContent string, taskImagePipelineDefinitionMap map[string]Pipeline) error {
	dagInfo, err := p.dagCheck()
	if err != nil {
		return fmt.Errorf("dag cycle check error %v", err)
	}

	err = p.checkPlaceholderExist(pipelineContent)
	if err != nil {
		return err
	}

	err = p.checkOutputsPlaceholder(dagInfo)
	if err != nil {
		return fmt.Errorf("placeholder detection error %v", err)
	}

	// 校验流水线类型的任务是否在定义中有对应的 input
	err = p.checkPipelineTypeTaskInputDefinition(taskImagePipelineDefinitionMap)
	if err != nil {
		return err
	}

	// 校验流水线类型的任务是否在定义中有对应的 output
	err = p.checkPipelineTypeTaskOutputDefinition(taskImagePipelineDefinitionMap)
	if err != nil {
		return fmt.Errorf("placeholder detection error %v", err)
	}

	// 校验流水线类型的任务入参值的类型是否和对应的类型匹配
	err = p.checkPipelineTypeTaskInputValueType(taskImagePipelineDefinitionMap)
	if err != nil {
		return fmt.Errorf("placeholder detection error %v", err)
	}
	return nil
}

// 校验流水线类型的任务是否在定义中有对应的 input
func (p *Pipeline) checkPipelineTypeTaskInputDefinition(taskAliasPipelineInfoMap map[string]Pipeline) error {
	var taskMap = map[string]Task{}
	for _, task := range p.Tasks {
		taskMap[task.Alias] = task
	}

	var mapKeyBuild = func(image string, name string) string {
		return fmt.Sprintf("%v-%v", image, name)
	}

	var pipelineInputMap = map[string]Input{}
	for _, task := range p.GetPipelineTypeTask() {
		// 获取流水线类型的任务对应的流水线入参的 map
		for _, input := range taskAliasPipelineInfoMap[task.Image].Inputs {
			pipelineInputMap[mapKeyBuild(task.Image, input.Name)] = input
		}
	}

	for _, task := range p.GetPipelineTypeTask() {
		// 遍历任务的入参
		for _, taskInput := range task.Inputs {
			// 将流水线入参类型定义赋值给任务入参的定义
			_, ok := pipelineInputMap[mapKeyBuild(task.Image, taskInput.Name)]
			if !ok {
				return fmt.Errorf("task (alias %v) input (name %v) No definition in pipeline (image %v)", task.Alias, taskInput.Name, task.Image)
			}
		}
	}
	return nil
}

// 校验流水线类型的任务是否在定义中有对应的 output
func (p *Pipeline) checkPipelineTypeTaskOutputDefinition(taskAliasPipelineInfoMap map[string]Pipeline) error {
	var taskMap = map[string]Task{}
	for _, task := range p.Tasks {
		taskMap[task.Alias] = task
	}

	var mapKeyBuild = func(image string, name string) string {
		return fmt.Sprintf("%v-%v", image, name)
	}

	var pipelineOutputMap = map[string]Output{}
	for _, task := range p.GetPipelineTypeTask() {
		// 获取流水线类型的任务对应的流水线入参的 map
		for _, output := range taskAliasPipelineInfoMap[task.Image].Outputs {
			pipelineOutputMap[mapKeyBuild(task.Alias, output.Name)] = output
		}
	}

	for _, task := range p.GetPipelineTypeTask() {
		// 遍历任务的入参
		for _, taskOutput := range task.Outputs {
			pipelineOutput, ok := pipelineOutputMap[mapKeyBuild(task.Alias, taskOutput.Value)]
			if !ok {
				return fmt.Errorf("task (alias %v) output (name %v) No definition in pipeline (image %v)", task.Alias, taskOutput.Name, task.Image)
			}
			if pipelineOutput.Type != taskOutput.Type {
				return fmt.Errorf("task (alias %v) output (name %v) type not match", task.Alias, taskOutput.Name)
			}
		}
	}
	return nil
}

// 校验流水线类型的任务入参的值的类型是否对应上
func (p *Pipeline) checkPipelineTypeTaskInputValueType(taskImagePipelineInfoMap map[string]Pipeline) error {
	var taskMap = map[string]Task{}
	for _, task := range p.Tasks {
		taskMap[task.Alias] = task
	}

	var mapKeyBuild = func(image string, name string) string {
		return fmt.Sprintf("%v-%v", image, name)
	}

	var pipelineInputMap = map[string]Input{}
	for _, task := range p.GetPipelineTypeTask() {
		// 获取流水线类型的任务对应的流水线入参的 map
		for _, input := range taskImagePipelineInfoMap[task.Image].Inputs {
			pipelineInputMap[mapKeyBuild(task.Image, input.Name)] = input
		}
	}

	for _, task := range p.GetPipelineTypeTask() {
		// 遍历任务的入参
		for _, taskInput := range task.Inputs {
			// 判定任务的入参是否以 ${{ 开头 以 }} 结尾
			if !strings.HasPrefix(taskInput.Value, placeholder.Left) || !strings.HasSuffix(taskInput.Value, placeholder.Right) {
				return fmt.Errorf("task (alias %v) input (name %v) value should use ${{ xxx }} placeholder", task.Alias, taskInput.Name)
			}

			// 将流水线入参类型定义赋值给任务入参的定义
			taskInput.Type = pipelineInputMap[mapKeyBuild(task.Image, taskInput.Name)].Type

			err := placeholder.MatchHolderFromHandler(taskInput.Value, map[placeholder.Type]placeholder.Handler{
				placeholder.ContextType: func(placeholder string, values ...string) error {
					contextName := values[1]

					// 判定任务入参 value 是 contexts 占位符的时候, 其入参值定义是否和 context 的值定义相同
					for _, ctx := range p.Contexts {
						if ctx.Name == contextName && ctx.Type != taskInput.Type {
							return fmt.Errorf("task (alias %v) input (name %v) value type not match pipeline context (name %v) type", task.Alias, taskInput.Name, ctx.Name)
						}
					}
					return nil
				},
				placeholder.InputType: func(placeholder string, values ...string) error {
					inputName := values[1]

					// 判定任务入参 value 是 inputs 占位符的时候, 其入参值定义是否和 inputs 的值定义相同
					for _, input := range p.Inputs {
						if input.Name == inputName && input.Type != taskInput.Type {
							return fmt.Errorf("task (alias %v) input (name %v) value type not match pipeline input (name %v) type", task.Alias, taskInput.Name, input.Name)
						}
					}
					return nil
				},
				placeholder.OutputType: func(placeholder string, values ...string) error {
					taskAlias := values[1]
					taskOutputName := values[2]

					outputTask := taskMap[taskAlias]
					for _, output := range outputTask.Outputs {
						if taskOutputName == output.Name && taskInput.Type != output.Type {
							return fmt.Errorf("task (alias %v) input (name %v) value type not match task (alias %v) output (name %v) type", task.Alias, taskInput.Name, taskAlias, taskOutputName)
						}
					}
					return nil
				},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Pipeline) checkOutputsPlaceholder(dagInfo *dag.Dag) error {
	var taskMap = map[string]Task{}
	for _, task := range p.Tasks {
		taskMap[task.Alias] = task
	}

	for _, task := range p.Tasks {
		taskYaml, err := yaml.Marshal(task)
		if err != nil {
			return fmt.Errorf("task alias %v yaml marshal error %v", task.Alias, err)
		}

		err = placeholder.MatchHolderFromHandler(string(taskYaml), map[placeholder.Type]placeholder.Handler{
			placeholder.OutputType: func(placeholder string, values ...string) error {
				taskAlias := values[1]
				taskOutputName := values[2]

				// 校验出参占位符中的任务是否是 dag 中之前可以被执行到的任务
				var inNeedTask = false
				allNeedTasks := dagInfo.GetAllNeedsTask(task.Alias)
				for _, needTask := range allNeedTasks {
					if taskAlias == needTask {
						inNeedTask = true
						break
					}
				}
				if !inNeedTask {
					return fmt.Errorf("this task (alias %v) will not be scheduled before this placeholder %v task (alias %v)", task.Alias, placeholder, taskAlias)
				}

				// 校验出参是否在当前 pipeline 中定义
				var isTaskOutput = false
				outputTask := taskMap[taskAlias]
				for _, output := range outputTask.Outputs {
					if output.Name == taskOutputName {
						isTaskOutput = true
						break
					}
				}
				if !isTaskOutput {
					return fmt.Errorf("this placeholder %v task (alias %v) not have output %v", placeholder, outputTask.Alias, taskOutputName)
				}
				return nil
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// 校验入参和全局参数是否在当前 pipeline 中定义
func (p *Pipeline) checkPlaceholderExist(pipelineContent string) error {
	var mapKeyBuild = func(alias string, name string) string {
		return fmt.Sprintf("%v-%v", alias, name)
	}

	var taskAliasOutputMap = map[string]Output{}
	for _, task := range p.Tasks {
		for _, output := range task.Outputs {
			taskAliasOutputMap[mapKeyBuild(task.Alias, output.Name)] = output
		}
	}

	err := placeholder.MatchHolderFromHandler(pipelineContent, map[placeholder.Type]placeholder.Handler{
		placeholder.ContextType: func(placeholder string, values ...string) error {
			contextName := values[1]

			var findContextName = false
			for _, ctx := range p.Contexts {
				if ctx.Name == contextName {
					findContextName = true
					break
				}
			}

			if !findContextName {
				return fmt.Errorf("pipeline not has context %v", contextName)
			}
			return nil
		},
		placeholder.InputType: func(placeholder string, values ...string) error {
			inputName := values[1]

			var findInputName = false
			for _, input := range p.Inputs {
				if input.Name == inputName {
					findInputName = true
					break
				}
			}

			if !findInputName {
				return fmt.Errorf("pipeline not has input %v", inputName)
			}
			return nil
		},
		placeholder.OutputType: func(placeholder string, values ...string) error {
			taskAlias := values[1]
			outputName := values[2]

			_, ok := taskAliasOutputMap[mapKeyBuild(taskAlias, outputName)]
			if !ok {
				return fmt.Errorf("pipeline output (value %v) not find in task %v", placeholder, taskAlias)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) dagCheck() (*dag.Dag, error) {
	dagInfo := dag.NewDag()
	for _, task := range p.Tasks {
		dagInfo.AddNode(task.Alias)
	}

	for _, node := range p.Dag {
		err := dagInfo.AddEdge(node.Name, node.Needs)
		if err != nil {
			return nil, err
		}
	}
	err := dagInfo.Check()
	if err != nil {
		return nil, err
	}

	return dagInfo, nil
}
