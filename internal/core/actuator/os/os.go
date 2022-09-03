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

package os

import (
	"context"
	"eventops/apistructs"
	"eventops/internal/core/actuator"
	client "eventops/pkg/schema/actuator"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/rancher/remotedialer"
	"golang.org/x/crypto/ssh"
	"net"
	"strings"
	"time"
)

type Actuator struct {
	client goph.Client
}

const nohupShellName = "nohup.sh"
const runShellName = "run.sh"

var nohupShell = `
#!/bin/bash
( 
nohup ./run.sh > nohup.log 2>&1
echo "$?" > exit.code
) &
`

type ShellTemplateObject struct {
	Bash    string
	Command []string
}

func workDir(pipelineId string, sign string) string {
	return fmt.Sprintf("pipelines/%v/tasks/%v", pipelineId, sign)
}

func (a Actuator) Create(ctx context.Context, task *actuator.Job) (*actuator.Job, error) {
	exist, err := a.Exist(ctx, task)
	if err != nil {
		return nil, err
	}

	if !exist {
		command := fmt.Sprintf("mkdir -p %v", workDir(task.PipelineId, task.TaskId))
		_, err := a.client.RunContext(ctx, command)
		if err != nil {
			return nil, err
		}
	}

	err = a.createRunShell(ctx, task)
	if err != nil {
		return nil, err
	}

	err = a.createNohupShell(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (a Actuator) createRunShell(ctx context.Context, task *actuator.Job) error {

	var command = "#!/bin/bash\n"
	command += fmt.Sprintf("echo $$ > nohup.pid \n")

	for _, cmd := range task.PreCommands {
		command += fmt.Sprintf("%v\n", cmd)
	}
	for _, cmd := range task.DefinitionTask.Commands {
		command += fmt.Sprintf("%v\n", cmd)
	}
	for _, cmd := range task.NextCommands {
		command += fmt.Sprintf("%v\n", cmd)
	}

	outputs, err := a.client.RunContext(ctx, fmt.Sprintf("cd %v && touch %v && chmod +x %v && cat > %v <<'EOF'\n%v", workDir(task.PipelineId, task.TaskId), runShellName, runShellName, runShellName, command))
	if err != nil {
		return fmt.Errorf("error %v, output: %v", err, string(outputs))
	}
	return nil
}

func (a Actuator) createNohupShell(ctx context.Context, task *actuator.Job) error {
	outputs, err := a.client.RunContext(ctx, fmt.Sprintf("cd %v && echo '%v' > %v && chmod +x %v", workDir(task.PipelineId, task.TaskId), nohupShell, nohupShellName, nohupShellName))
	if err != nil {
		return fmt.Errorf("error %v, output: %v", err, string(outputs))
	}

	return nil
}
func (a Actuator) Start(ctx context.Context, task *actuator.Job) error {
	output, err := a.client.RunContext(ctx, fmt.Sprintf("cd %v && nohup ./%v > /dev/null 2>&1 &", workDir(task.PipelineId, task.TaskId), nohupShellName))
	if err != nil {
		return err
	}

	output, err = a.client.Run(fmt.Sprintf("cat %v/%v", workDir(task.PipelineId, task.TaskId), "nohup.pid"))
	if err != nil {
		return err
	}

	task.JobSign = strings.TrimSpace(string(output))
	return nil
}

func (a Actuator) Remove(ctx context.Context, task *actuator.Job) error {
	_, err := a.client.RunContext(ctx, fmt.Sprintf("rm -rf %v", workDir(task.PipelineId, task.TaskId)))
	return err
}

func (a Actuator) Cancel(ctx context.Context, task *actuator.Job) error {
	status, err := a.Status(ctx, task)
	if err != nil {
		return err
	}

	if !status.IsDoneStatus() {
		// todo run.sh 中正在执行的这条命令是不会被中止的，这里中止的只是 run.sh 这个脚本
		// 是否需要一个脚本递归的将 run.sh 的 pid 下的所有子进程都中止掉
		_, err := a.client.RunContext(ctx, fmt.Sprintf("kill -15 %v", task.JobSign))
		if err != nil {
			return err
		}
	}
	return nil
}

func (a Actuator) Exist(ctx context.Context, task *actuator.Job) (bool, error) {
	var shell = `
if [ ! -d "` + workDir(task.PipelineId, task.TaskId) + `" ]; then
    echo "1"
else
    echo "0"
fi
`

	output, err := a.client.RunContext(ctx, shell)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(output)) == "1" {
		return false, nil
	}
	return true, nil
}

func (a Actuator) Status(ctx context.Context, task *actuator.Job) (apistructs.TaskStatus, error) {
	exist, err := a.Exist(ctx, task)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", actuator.JobNotFindError
	}

	var shell = `
if [ ! -f "exit.code" ]; then
	echo ""
else
    cat exit.code
fi;
`
	command := fmt.Sprintf("cd %v", workDir(task.PipelineId, task.TaskId))
	command = fmt.Sprintf("%s && %s", command, shell)

	output, err := a.client.RunContext(ctx, command)
	if err != nil {
		return "", err
	}

	code := strings.TrimSpace(string(output))
	switch code {
	case "0":
		return apistructs.SuccessTaskStatus, nil
	case "1", "2", "126", "127", "128", "255", "130":
		task.Error = fmt.Sprintf("Exited (%v)", code)
		return apistructs.FailedTaskStatus, nil
	case "15":
		task.Error = fmt.Sprintf("Exited (%v)", code)
		return apistructs.CancelTaskStatus, nil
	case "":
		return apistructs.RunningTaskStatus, nil
	default:
		task.Error = fmt.Sprintf("Exited (%v)", code)
		return apistructs.UnKnowTaskStatus, nil
	}
}

func (a Actuator) Type() apistructs.TaskType {
	return apistructs.OsType
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
