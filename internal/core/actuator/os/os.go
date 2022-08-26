package os

import (
	"context"
	"eventops/apistructs"
	"eventops/internal/core/actuator"
	client "eventops/pkg/schema/actuator"
	"eventops/pkg/schema/pipeline"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/rancher/remotedialer"
	"golang.org/x/crypto/ssh"
	"html/template"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

type Actuator struct {
	client goph.Client
}

const bashShellName = "bash.sh"
const runShellName = "run.sh"

var runShellTemplate = `
{{.Bash}}

{{range $index, $value := .Command}}
{{$value}}
{{end}}
`

var bashShell = `
#!/bin/bash
chmod +x ./run.sh
./run.sh
code=$?
echo $code > exit.code
`

type ShellTemplateObject struct {
	Bash    string
	Command []string
}

func workDir(pipelineId string, sign string) string {
	return fmt.Sprintf("%v/%v", pipelineId, sign)
}

func (a Actuator) Create(ctx context.Context, task *actuator.Job) (*actuator.Job, error) {
	session, err := a.client.NewSession()
	if err != nil {
		return nil, err
	}
	err = session.Run(fmt.Sprintf("rm -rf %v", workDir(task.PipelineId, task.TaskId)))
	if err != nil {
		return nil, err
	}
	err = session.Run(fmt.Sprintf("mkdir -p %v", workDir(task.PipelineId, task.TaskId)))
	if err != nil {
		return nil, err
	}
	pathOutput, err := session.Output(fmt.Sprintf("cd %v && pwd", workDir(task.PipelineId, task.TaskId)))
	if err != nil {
		return nil, err
	}
	if string(pathOutput) == "" {
		return nil, fmt.Errorf("work path name can not empty")
	}

	err = a.createRunShell(task, string(pathOutput))
	if err != nil {
		return nil, err
	}

	err = a.createBashShell(task, string(pathOutput))
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (a Actuator) createRunShell(task *actuator.Job, workdir string) error {
	tempFile, err := ioutil.TempFile(fmt.Sprintf("%v/%v", task.PipelineId, task.TaskId), runShellName)
	if err != nil {
		return err
	}
	defer tempFile.Close()
	shellTpl, err := template.New("shell").Parse(runShellTemplate)
	if err != nil {
		return err
	}
	err = shellTpl.Execute(tempFile, ShellTemplateObject{
		Bash:    "#!/bin/bash",
		Command: task.DefinitionTask.Commands,
	})
	if err != nil {
		return err
	}
	return a.client.Upload(tempFile.Name(), workdir)
}

func (a Actuator) createBashShell(task *actuator.Job, workdir string) error {
	bashFile, err := ioutil.TempFile(fmt.Sprintf("%v/%v", task.PipelineId, task.TaskId), bashShellName)
	if err != nil {
		return err
	}
	defer bashFile.Close()
	_, err = bashFile.WriteString(bashShell)
	if err != nil {
		return err
	}

	return a.client.Upload(bashFile.Name(), workdir)
}

func (a Actuator) Start(ctx context.Context, task *actuator.Job) error {
	session, err := a.client.NewSession()
	if err != nil {
		return err
	}
	err = session.Run(fmt.Sprintf("cd %v", workDir(task.PipelineId, task.TaskId)))
	if err != nil {
		return err
	}
	output, err := session.Output(fmt.Sprintf("chmod +x ./%v && echo '' > nohup.out  && nohup ./%v > nohup.out 2>&1 &", bashShellName, bashShellName))
	if err != nil {
		return err
	}

	task.JobSign = strings.TrimLeft(string(output), "[1]")
	task.JobSign = strings.TrimSpace(task.JobSign)
	return nil
}

func (a Actuator) Remove(ctx context.Context, task *actuator.Job) error {
	session, err := a.client.NewSession()
	if err != nil {
		return err
	}
	return session.Run(fmt.Sprintf("rm -rf %v", workDir(task.PipelineId, task.TaskId)))
}

func (a Actuator) Cancel(ctx context.Context, task *actuator.Job) error {
	session, err := a.client.NewSession()
	if err != nil {
		return err
	}
	return session.Run(fmt.Sprintf("kill -15 %v", task.JobSign))
}

func (a Actuator) Exist(ctx context.Context, task *actuator.Job) (bool, error) {
	var shell = `
if [ ! -d "` + workDir(task.PipelineId, task.TaskId) + `" ]; then
    echo "1"
else
    echo "0"
fi
`

	output, err := a.client.Run(shell)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(output)) == "1" {
		return true, nil
	}
	return false, nil
}

func (a Actuator) Status(ctx context.Context, task *actuator.Job) (apistructs.TaskStatus, error) {
	session, err := a.client.NewSession()
	if err != nil {
		return "", err
	}
	outputs, err := session.Output(fmt.Sprintf("cd %v", workDir(task.PipelineId, task.TaskId)))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(outputs)) != "" {
		return "", actuator.JobNotFindError
	}

	var shell = `
if [ ! -f "exit.code" ]; then
    cat exit.code
else
    echo ""
fi
`
	output, err := session.Output(shell)
	if err != nil {
		return "", err
	}
	switch string(output) {
	case "0":
		return apistructs.SuccessTaskStatus, nil
	case "1", "2", "126", "127", "128", "255", "130":
		return apistructs.FailedTaskStatus, nil
	case "15":
		return apistructs.CancelTaskStatus, nil
	case "":
		return apistructs.RunningTaskStatus, nil
	default:
		return apistructs.UnKnowTaskStatus, nil
	}
}

func (a Actuator) Type() pipeline.TaskType {
	return pipeline.OsType
}

func NewOsClient(osConfig *client.Os, dialer remotedialer.Dialer) (*Actuator, error) {
	var conn net.Conn
	var err error

	if dialer != nil {
		conn, err = dialer("tcp", fmt.Sprintf("%s:%s", osConfig.Ip, osConfig.Port))
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", osConfig.Ip, osConfig.Port), 5*time.Second)
		if err != nil {
			return nil, err
		}
	}

	sshConn, a, b, err := ssh.NewClientConn(conn, fmt.Sprintf("%s:%s", osConfig.Ip, osConfig.Port), &ssh.ClientConfig{
		User:            osConfig.User,
		Auth:            []ssh.AuthMethod{ssh.Password(osConfig.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}

	return &Actuator{
		client: goph.Client{
			Client: ssh.NewClient(sshConn, a, b),
		},
	}, nil
}
