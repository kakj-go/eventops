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

package docker

import (
	"context"
	"eventops/apistructs"
	"eventops/internal/core/actuator"
	actuatorclient "eventops/pkg/schema/actuator"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/rancher/remotedialer"
	"golang.org/x/crypto/ssh"
	"net"
	"strings"
	"time"
)

type Actuator struct {
	client *client.Client
}

func NewDockerClient(dockerConfig *actuatorclient.Docker, dialer remotedialer.Dialer) (*Actuator, error) {
	var ops []client.Opt
	ops = append(ops, client.WithAPIVersionNegotiation())
	if dialer != nil {
		ops = append(ops, client.WithHost(fmt.Sprintf("tcp://%v:%v", dockerConfig.Ip, dockerConfig.Port)))
		ops = append(ops, client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer(network, addr)
		}))
		goto BUILD
	}

	if dockerConfig.Ssh != nil {
		sshConfig := &ssh.ClientConfig{
			User:            dockerConfig.Ssh.User,
			Timeout:         time.Second * 5,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth:            []ssh.AuthMethod{ssh.Password(dockerConfig.Ssh.Password)},
		}
		conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", dockerConfig.Ssh.Ip, dockerConfig.Ssh.Port), sshConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "dail tcp %s err: %v", dockerConfig.Ssh.Ip, err)
		}
		ops = append(ops, client.WithHost(fmt.Sprintf("tcp://%v:%v", dockerConfig.Ip, dockerConfig.Port)))
		ops = append(ops, client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
			return conn.Dial(network, addr)
		}))
		goto BUILD
	}

	ops = append(ops, client.WithHost(fmt.Sprintf("tcp://%v:%v", dockerConfig.Ip, dockerConfig.Port)))
BUILD:
	if len(ops) == 0 {
		return nil, fmt.Errorf("not find docker config opt")
	}
	cl, err := client.NewClientWithOpts(
		ops...,
	)
	if err != nil {
		return nil, err
	}

	return &Actuator{
		client: cl,
	}, nil
}

func (a *Actuator) Create(ctx context.Context, task *actuator.Job) (*actuator.Job, error) {
	out, err := a.client.ImagePull(ctx, task.DefinitionTask.Image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	defer out.Close()

	command := fmt.Sprintf("echo 'task %v start'", task.DefinitionTask.Alias)
	for _, cmd := range task.PreCommands {
		command = fmt.Sprintf("%s && %s", command, cmd)
	}
	for _, cmd := range task.DefinitionTask.Commands {
		command = fmt.Sprintf("%s && %s", command, cmd)
	}
	for _, cmd := range task.NextCommands {
		command = fmt.Sprintf("%s && %s", command, cmd)
	}

	resp, err := a.client.ContainerCreate(ctx, &container.Config{
		Image: task.DefinitionTask.Image,
		Cmd:   []string{"sh", "-c", command},
	}, nil, nil, nil, task.TaskId)
	if err != nil {
		return nil, err
	}

	task.JobSign = resp.ID
	return task, nil
}

func (a *Actuator) Start(ctx context.Context, task *actuator.Job) error {
	return a.client.ContainerStart(ctx, task.JobSign, types.ContainerStartOptions{})
}

func (a *Actuator) Exist(ctx context.Context, task *actuator.Job) (bool, error) {
	containers, err := a.getContainers(ctx, task.JobSign)
	if err != nil {
		return false, err
	}
	if len(containers) > 0 {
		return true, nil
	}
	return false, nil
}

func (a *Actuator) Remove(ctx context.Context, task *actuator.Job) error {
	return a.client.ContainerRemove(ctx, task.JobSign, types.ContainerRemoveOptions{})
}

func (a *Actuator) Cancel(ctx context.Context, task *actuator.Job) error {
	status, err := a.Status(ctx, task)
	if err != nil {
		if err == actuator.JobNotFindError {
			return nil
		}
		return err
	}
	if status != apistructs.RunningTaskStatus {
		return nil
	}

	return a.client.ContainerStop(ctx, task.JobSign, nil)
}

func (a Actuator) Type() apistructs.TaskType {
	return apistructs.DockerType
}

func (a *Actuator) Status(ctx context.Context, task *actuator.Job) (apistructs.TaskStatus, error) {
	containers, err := a.getContainers(ctx, task.JobSign)
	if err != nil {
		return "", err
	}
	if len(containers) == 0 {
		return "", actuator.JobNotFindError
	}

	var dockerContainer = containers[0]
	var resultStatus apistructs.TaskStatus
	switch dockerContainer.State {
	case "created":
		resultStatus = apistructs.CreatedTaskStatus
	case "restarting", "running":
		resultStatus = apistructs.RunningTaskStatus
	case "paused":
		resultStatus = apistructs.RunningTaskStatus
	case "dead":
		task.Error = dockerContainer.Status
		resultStatus = apistructs.FailedTaskStatus
	case "exited":
		// todo dockerContainer.Status(example: Exited (127) Less than a second ago)后面会带无用字段, 获取错误原因的时候会有误判
		if strings.HasPrefix(dockerContainer.Status, "Exited (0)") {
			resultStatus = apistructs.SuccessTaskStatus
		} else {
			task.Error = dockerContainer.Status
			resultStatus = apistructs.FailedTaskStatus
		}
	default:
		task.Error = dockerContainer.Status
		resultStatus = apistructs.UnKnowTaskStatus
	}
	return resultStatus, nil
}

func (a *Actuator) getContainers(ctx context.Context, id string) ([]types.Container, error) {
	return a.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "id",
			Value: id,
		}),
	})
}
