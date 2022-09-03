package k8s

import (
	"bytes"
	"context"
	"eventops/apistructs"
	"eventops/internal/core/actuator"
	client "eventops/pkg/schema/actuator"
	"fmt"
	"github.com/rancher/remotedialer"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"net"
)

type Actuator struct {
	client *kubernetes.Clientset
	config *restclient.Config
}

func isPodNotFindError(err error, podName string) bool {
	if err.Error() == fmt.Sprintf("pods \"%v\" not found", podName) {
		return true
	}
	return false
}

func isNsNotFindError(err error, ns string) bool {
	if err.Error() == fmt.Sprintf("namespaces \"%v\" not found", ns) {
		return true
	}
	return false
}

func makeNamespace(namespaceId string) string {
	return fmt.Sprintf("pipelines-%v", namespaceId)
}

func (a Actuator) Create(ctx context.Context, task *actuator.Job) (*actuator.Job, error) {
	var ns = corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: makeNamespace(task.PipelineId),
		},
	}

	exist, err := a.client.CoreV1().Namespaces().Get(ctx, makeNamespace(task.PipelineId), metav1.GetOptions{})
	if err != nil {
		if !isNsNotFindError(err, makeNamespace(task.PipelineId)) {
			return nil, err
		}
		exist = nil
	}
	if exist == nil {
		_, err := a.client.CoreV1().Namespaces().Create(ctx, &ns, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	task.JobSign = fmt.Sprintf("%v-%v", task.DefinitionTask.Alias, task.TaskId)
	return task, nil
}

func (a Actuator) Start(ctx context.Context, task *actuator.Job) error {
	if exist, err := a.Exist(ctx, task); err != nil {
		return err
	} else {
		if exist {
			return nil
		}
	}

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

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: task.JobSign,
		},
		Spec: corev1.PodSpec{
			RestartPolicy:         corev1.RestartPolicyNever,
			ShareProcessNamespace: &[]bool{true}[0],
			Containers: []corev1.Container{
				{
					Name:    task.DefinitionTask.Alias,
					Image:   task.DefinitionTask.Image,
					Command: []string{"sh"},
					Args:    []string{"-c", command},
					Stdin:   true,
					TTY:     true,
				},
			},
		},
	}

	_, err := a.client.CoreV1().Pods(makeNamespace(task.PipelineId)).Create(ctx, &pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (a Actuator) Remove(ctx context.Context, task *actuator.Job) error {
	return a.client.CoreV1().Pods(makeNamespace(task.PipelineId)).Delete(ctx, task.JobSign, metav1.DeleteOptions{})
}

func (a Actuator) Cancel(ctx context.Context, task *actuator.Job) error {
	status, err := a.Status(ctx, task)
	if err != nil {
		if err == actuator.JobNotFindError {
			return nil
		}
		return err
	}
	if status.IsDoneStatus() {
		return nil
	}
	// 构造执行命令请求
	req := a.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(task.JobSign).
		Namespace(makeNamespace(task.PipelineId)).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"sh", "-c", "kill -15 `ps | awk '{print $1}' | awk 'NR == 3'`"},
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	exec, err := remotecommand.NewSPDYExecutor(a.config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return err
	}
	if stdout.String() != "" {
		return fmt.Errorf(stdout.String())
	}
	if stderr.String() != "" {
		return fmt.Errorf(stderr.String())
	}

	return nil
}

func (a Actuator) Exist(ctx context.Context, task *actuator.Job) (bool, error) {
	pod, err := a.client.CoreV1().Pods(makeNamespace(task.PipelineId)).Get(ctx, task.JobSign, metav1.GetOptions{})
	if err != nil {
		if !isPodNotFindError(err, task.JobSign) {
			return false, err
		}
		return false, nil
	}
	if pod != nil {
		return true, nil
	}
	return false, nil
}

func (a Actuator) Type() apistructs.TaskType {
	return apistructs.K8sType
}

func (a Actuator) Status(ctx context.Context, task *actuator.Job) (apistructs.TaskStatus, error) {
	pod, err := a.client.CoreV1().Pods(makeNamespace(task.PipelineId)).Get(ctx, task.JobSign, metav1.GetOptions{})
	if err != nil {
		if !isPodNotFindError(err, task.JobSign) {
			return "", err
		}
		return "", actuator.JobNotFindError
	}
	if pod == nil {
		return "", actuator.JobNotFindError
	}

	var result apistructs.TaskStatus

	switch pod.Status.Phase {
	case corev1.PodPending:
		result = apistructs.RunningTaskStatus
	case corev1.PodRunning:
		result = apistructs.RunningTaskStatus
	case corev1.PodSucceeded:
		result = apistructs.SuccessTaskStatus
	case corev1.PodFailed:
		result = apistructs.FailedTaskStatus
		task.Error = getContainerExitCode(pod)
	default:
		result = apistructs.UnKnowTaskStatus
		task.Error = getContainerExitCode(pod)
	}
	return result, nil
}

func getContainerExitCode(pod *corev1.Pod) string {
	if pod.Status.Reason != "" {
		return pod.Status.Reason
	}

	if pod.Status.Message != "" {
		return pod.Status.Message
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Sprintf("Exited (1)")
	}

	if pod.Status.ContainerStatuses[0].State.Terminated == nil {
		return fmt.Sprintf("Exited (1)")
	}

	return fmt.Sprintf("Exited (%v)", pod.Status.ContainerStatuses[0].State.Terminated.ExitCode)
}

func NewKubernetesClient(k8sConfig *client.Kubernetes, dialer remotedialer.Dialer) (*Actuator, error) {
	f, err := ioutil.TempFile("", "*.config")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.WriteString(k8sConfig.Config)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", f.Name())
	if err != nil {
		return nil, err
	}
	if dialer != nil {
		config.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer(network, address)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Actuator{
		client: clientset,
		config: config,
	}, nil
}
