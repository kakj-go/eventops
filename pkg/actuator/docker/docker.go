package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/rancher/remotedialer"
	"golang.org/x/crypto/ssh"
	"net"
	"net/http"
	"tiggerops/pkg/actuator"
	actuatorclient "tiggerops/pkg/schema/actuator"
	"time"
)

type Actuator struct {
	client *client.Client
}

func NewDockerClient(dockerConfig *actuatorclient.Docker, dialer remotedialer.Dialer) (*Actuator, error) {
	var ops []client.Opt

	if dialer != nil {
		ops = append(ops, client.WithHost(fmt.Sprintf("tcp://%v:%v", dockerConfig.Ip, &dockerConfig.Port)))
		ops = append(ops, client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) { return dialer(network, addr) }))
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
		sockConn, err := conn.Dial("unix", "/var/run/docker.sock")
		if err != nil {
			return nil, err
		}

		httpClient := http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return sockConn, nil
				},
				TLSHandshakeTimeout:   time.Second * 10,
				IdleConnTimeout:       time.Second * 10,
				ResponseHeaderTimeout: time.Second * 10,
			},
		}

		ops = append(ops, client.WithHTTPClient(&httpClient))
		ops = append(ops, client.WithAPIVersionNegotiation())
		goto BUILD
	}

	ops = append(ops, client.WithHost(fmt.Sprintf("tcp://%v:%v", dockerConfig.Ip, &dockerConfig.Port)))
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

func (a *Actuator) Create(ctx context.Context, task *actuator.Task) (*actuator.Task, error) {
	out, err := a.client.ImagePull(ctx, task.DefinitionTask.Image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	defer out.Close()

	command := fmt.Sprintf("echo 'task %v start'", task.DefinitionTask.Alias)
	for _, cmd := range task.DefinitionTask.Commands {
		command = fmt.Sprintf("%s && %s", command, cmd)
	}

	resp, err := a.client.ContainerCreate(ctx, &container.Config{
		Image: task.DefinitionTask.Image,
		Cmd:   []string{command},
	}, nil, nil, nil, task.Sign)
	if err != nil {
		return nil, err
	}

	task.InstanceSign = resp.ID
	return task, nil
}

func (a *Actuator) Start(ctx context.Context, task *actuator.Task) error {
	return a.client.ContainerStart(ctx, task.InstanceSign, types.ContainerStartOptions{})
}

func (a *Actuator) Exist(ctx context.Context, task *actuator.Task) (bool, error) {
	containers, err := a.getContainers(ctx, task.InstanceSign)
	if err != nil {
		return false, err
	}
	if len(containers) > 0 {
		return true, nil
	}
	return false, nil
}

func (a *Actuator) Remove(ctx context.Context, task *actuator.Task) error {
	return a.client.ContainerRemove(ctx, task.InstanceSign, types.ContainerRemoveOptions{})
}

func (a *Actuator) Cancel(ctx context.Context, task *actuator.Task) error {
	return a.client.ContainerStop(ctx, task.InstanceSign, nil)
}

func (a *Actuator) Status(ctx context.Context, task *actuator.Task) (actuator.TaskStatus, error) {
	containers, err := a.getContainers(ctx, task.InstanceSign)
	if err != nil {
		return "", err
	}
	if len(containers) == 0 {
		return "", actuator.TaskNotFindError
	}
	var dockerContainer = containers[0]
	var resultStatus actuator.TaskStatus
	switch dockerContainer.State {
	case "created":
		resultStatus = actuator.CreatedTaskStatus
	case "restarting", "running":
		resultStatus = actuator.RunningTaskStatus
	case "paused":
		resultStatus = actuator.PausedTaskStatus
	case "exited", "dead", "removing":
		if dockerContainer.Status == "Exit 0" {
			resultStatus = actuator.SuccessTaskStatus
		} else {
			resultStatus = actuator.FailedTaskStatus
		}
	default:
		resultStatus = actuator.UnKnowTaskStatus
	}
	return resultStatus, nil
}

func (a *Actuator) getContainers(ctx context.Context, id string) ([]types.Container, error) {
	return a.client.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "id",
			Value: id,
		}),
	})
}
