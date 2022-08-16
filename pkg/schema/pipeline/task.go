package pipeline

import (
	"fmt"
	"strings"
)

type TaskType string

const (
	K8sType    TaskType = "k8s"
	DockerType TaskType = "docker"
	OsType     TaskType = "os"
	PipeType   TaskType = "pipeline"
)

const ImageCreaterNameSplitWord = "/"
const ImageNameVersionSplitWord = ":"

var TaskTypeList = []TaskType{K8sType, DockerType, OsType, PipeType}

type Task struct {
	Image            string           `yaml:"image,omitempty"`
	Alias            string           `yaml:"alias,omitempty"`
	Type             TaskType         `yaml:"type,omitempty"`
	ActuatorSelector ActuatorSelector `yaml:"actuatorSelector,omitempty"`
	Inputs           []Input          `yaml:"inputs,omitempty"`
	Commands         []string         `yaml:"commands,omitempty"`
	Outputs          []Output         `yaml:"outputs,omitempty"`
	Timeout          int64            `yaml:"timeout,omitempty"`
	Resources        *Resources       `yaml:"resources,omitempty"`
}

func (t Task) GetPipelineTypeTaskVersion() string {
	if t.Type != PipeType {
		return ""
	}

	return GetImageVersion(t.Image)
}

func GetImageVersion(image string) string {
	split := strings.SplitN(image, ImageNameVersionSplitWord, 2)
	if len(split) == 1 {
		return ""
	}

	return split[1]
}

func (t Task) GetPipelineTypeTaskCreater() string {
	if t.Type != PipeType {
		return ""
	}

	return GetImageCreater(t.Image)
}

func GetImageCreater(image string) string {
	split := strings.SplitN(image, ImageCreaterNameSplitWord, 2)
	if len(split) == 1 {
		return ""
	}

	return split[0]
}

func (t Task) GetPipelineTypeTaskName() string {
	if t.Type != PipeType {
		return ""
	}

	var name = t.Image

	version := t.GetPipelineTypeTaskVersion()
	creater := t.GetPipelineTypeTaskCreater()

	if version != "" {
		name = strings.TrimRight(name, fmt.Sprintf(ImageNameVersionSplitWord+"%s", version))
	}

	if creater != "" {
		name = strings.TrimLeft(name, fmt.Sprintf("%s"+ImageCreaterNameSplitWord, creater))
	}
	return name
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

func (t Task) Check(pipelineContexts []Context) error {
	if t.Alias == "" {
		return fmt.Errorf("task alias can not empty")
	}

	if !t.Type.Check() {
		return fmt.Errorf("use [%s %s %s %s] these task type", K8sType, DockerType, OsType, PipeType)
	}

	if t.Type != OsType && len(t.Image) == 0 {
		return fmt.Errorf("[%s, %s, %s] type task image can not empty", K8sType, DockerType, PipeType)
	}

	if err := t.ActuatorSelector.check(); err != nil {
		return err
	}

	if err := t.inputCheck(); err != nil {
		return err
	}

	if t.Type != PipeType && len(t.Commands) == 0 {
		return fmt.Errorf("[%s, %s, %s] task type, commands can not empty", K8sType, DockerType, OsType)
	}

	if err := t.outputCheck(pipelineContexts); err != nil {
		return err
	}

	if err := t.resourceCheck(); err != nil {
		return err
	}

	return nil
}

func (t Task) outputCheck(pipelineContexts []Context) error {
	if t.Outputs == nil {
		return nil
	}

	for _, output := range t.Outputs {
		if output.Name == "" {
			return fmt.Errorf("output name can not empty")
		}

		if output.Value == "" {
			return fmt.Errorf("output name %v value can not empty", output.Name)
		}

		if t.Type != PipeType {
			err := output.Type.ValueTypeCheck()
			if err != nil {
				return fmt.Errorf("output name %v type check error %v", output.Name, err)
			}
		}

		if output.SetToContext != "" {
			var find = false
			for _, context := range pipelineContexts {
				if context.Name != output.SetToContext {
					continue
				}
				if context.Type != output.Type {
					return fmt.Errorf("output name %v setToContext %v type not match", output.Name, output.SetToContext)
				}
				find = true
				break
			}
			if !find {
				return fmt.Errorf("output name %v setToContext %v not find", output.Name, output.SetToContext)
			}
		}
	}
	return nil
}

func (t Task) inputCheck() error {
	if t.Inputs == nil {
		return nil
	}

	for _, input := range t.Inputs {
		if input.Name == "" {
			return fmt.Errorf("input name can not empty")
		}
		if input.Value == "" {
			return fmt.Errorf("input name %v value can not empty", input.Name)
		}
	}
	return nil
}

func (t Task) resourceCheck() error {
	if t.Resources == nil {
		return nil
	}

	if t.Resources.Limit != nil {
		if t.Resources.Limit.Cpu == "" {
			return fmt.Errorf("resources limit cpu can not empty")
		}
		if t.Resources.Limit.Mem == "" {
			return fmt.Errorf("resources limit cpu can not empty")
		}
	}

	if t.Resources.Request != nil {
		if t.Resources.Request.Cpu == "" {
			return fmt.Errorf("resources request cpu can not empty")
		}
		if t.Resources.Request.Mem == "" {
			return fmt.Errorf("resources request cpu can not empty")
		}
	}

	return nil
}
