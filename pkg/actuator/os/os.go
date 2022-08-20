package os

import (
	"context"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/rancher/remotedialer"
	"golang.org/x/crypto/ssh"
	"html/template"
	"io/ioutil"
	"net"
	"strings"
	"tiggerops/pkg/actuator"
	client "tiggerops/pkg/schema/actuator"
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

func workDir(pipelineId uint64, sign string) string {
	return fmt.Sprintf("%v/%v", pipelineId, sign)
}

func (a Actuator) Create(ctx context.Context, task *actuator.Task) (*actuator.Task, error) {
	session, err := a.client.NewSession()
	if err != nil {
		return nil, err
	}
	err = session.Run(fmt.Sprintf("rm -rf %v", workDir(task.PipelineID, task.Sign)))
	if err != nil {
		return nil, err
	}
	err = session.Run(fmt.Sprintf("mkdir -p %v", workDir(task.PipelineID, task.Sign)))
	if err != nil {
		return nil, err
	}
	pathOutput, err := session.Output(fmt.Sprintf("cd %v && pwd", workDir(task.PipelineID, task.Sign)))
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

func (a Actuator) createRunShell(task *actuator.Task, workdir string) error {
	tempFile, err := ioutil.TempFile(fmt.Sprintf("%v/%v", task.PipelineID, task.Sign), runShellName)
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

func (a Actuator) createBashShell(task *actuator.Task, workdir string) error {
	bashFile, err := ioutil.TempFile(fmt.Sprintf("%v/%v", task.PipelineID, task.Sign), bashShellName)
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

func (a Actuator) Start(ctx context.Context, task *actuator.Task) error {
	session, err := a.client.NewSession()
	if err != nil {
		return err
	}
	err = session.Run(fmt.Sprintf("cd %v", workDir(task.PipelineID, task.Sign)))
	if err != nil {
		return err
	}
	output, err := session.Output(fmt.Sprintf("chmod +x ./%v && nohup ./%v > nohup.out 2>&1 &", bashShellName, bashShellName))
	if err != nil {
		return err
	}

	task.InstanceSign = strings.TrimLeft(string(output), "[1]")
	task.InstanceSign = strings.TrimSpace(task.InstanceSign)
	return nil
}

func (a Actuator) Remove(ctx context.Context, task *actuator.Task) error {
	session, err := a.client.NewSession()
	if err != nil {
		return err
	}
	return session.Run(fmt.Sprintf("rm -rf %v", task.PipelineID))
}

func (a Actuator) Cancel(ctx context.Context, task *actuator.Task) error {
	session, err := a.client.NewSession()
	if err != nil {
		return err
	}
	return session.Run(fmt.Sprintf("kill -15 %v", task.InstanceSign))
}

func (a Actuator) Exist(ctx context.Context, task *actuator.Task) (bool, error) {
	output, err := a.client.Run(fmt.Sprintf("ps -q %v | awk '{print $1}' | awk 'NR == 2'", task.InstanceSign))
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(output)) == task.InstanceSign {
		return true, nil
	}
	return false, nil
}

func (a Actuator) Status(ctx context.Context, task *actuator.Task) (actuator.TaskStatus, error) {
	session, err := a.client.NewSession()
	if err != nil {
		return "", err
	}
	err = session.Run(fmt.Sprintf("cd %v", workDir(task.PipelineID, task.Sign)))
	if err != nil {
		return "", err
	}

	output, err := session.Output("cat exit.code")
	if err != nil {
		return "", err
	}
	switch string(output) {
	case "0":
		return actuator.SuccessTaskStatus, nil
	case "1", "2", "126", "127", "128", "255", "130":
		return actuator.FailedTaskStatus, nil
	case "15":
		return actuator.CancelTaskStatus, nil
	case "":
		return actuator.RunningTaskStatus, nil
	default:
		return actuator.UnKnowTaskStatus, nil
	}
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
