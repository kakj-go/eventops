package k8s

import (
	"bytes"
	"context"
	"eventops/apistructs"
	"eventops/internal/core/actuator"
	client "eventops/pkg/schema/actuator"
	"eventops/pkg/schema/pipeline"
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
	"strings"
)

type Actuator struct {
	client *kubernetes.Clientset
	config *restclient.Config
}

func makeNamespace(namespaceId string) string {
	return fmt.Sprintf("pipelines-%v", namespaceId)
}

func (a Actuator) Create(ctx context.Context, task *actuator.Job) (*actuator.Job, error) {
	task.JobSign = fmt.Sprintf("%v_%v", task.DefinitionTask.Alias, task.TaskId)
	return task, nil
}

func (a Actuator) Start(ctx context.Context, task *actuator.Job) error {
	command := fmt.Sprintf("echo 'task %v start'", task.DefinitionTask.Alias)
	for _, cmd := range task.DefinitionTask.Commands {
		command = fmt.Sprintf("%s && %s", command, cmd)
	}

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "v1",
			APIVersion: "pod",
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

	executor, err := remotecommand.NewSPDYExecutor(a.config, "POST", req.URL())
	if err != nil {
		return err
	}

	var stdout, stderr bytes.Buffer
	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return err
	}

	stdError := string(stderr.Bytes())
	if stdError != "" {
		return fmt.Errorf(stdError)
	}
	return nil
}

func (a Actuator) Exist(ctx context.Context, task *actuator.Job) (bool, error) {
	pod, err := a.client.CoreV1().Pods(makeNamespace(task.PipelineId)).Get(ctx, task.JobSign, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if pod != nil {
		return true, nil
	}
	return false, nil
}

func (a Actuator) Type() pipeline.TaskType {
	return pipeline.K8sType
}

func (a Actuator) Status(ctx context.Context, task *actuator.Job) (apistructs.TaskStatus, error) {
	pod, err := a.client.CoreV1().Pods(makeNamespace(task.PipelineId)).Get(ctx, task.JobSign, metav1.GetOptions{})
	if err != nil {
		return "", err
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
	default:
		result = apistructs.UnKnowTaskStatus
	}
	return result, nil
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
