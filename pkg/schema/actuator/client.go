package actuator

import (
	"fmt"
	"tiggerops/pkg/schema/pipeline"
)

type Actuator struct {
	Name string `json:"name"`

	Os         *Os         `yaml:"os,omitempty"`
	Kubernetes *Kubernetes `yaml:"kubernetes,omitempty"`
	Docker     *Docker     `yaml:"docker,omitempty"`

	Tags []string `yaml:"tags"`
}

func (a Actuator) Check() error {
	if a.Name == "" {
		return fmt.Errorf("name can not empty")
	}

	var configNum = 0
	if a.Os != nil {
		configNum++
	}
	if a.Kubernetes != nil {
		configNum++
	}
	if a.Docker != nil {
		configNum++
	}
	if configNum != 1 {
		return fmt.Errorf("[os, kubernetes, docker] only one of them can be configured")
	}

	if err := a.Os.Check(); err != nil {
		return err
	}
	if err := a.Kubernetes.Check(); err != nil {
		return err
	}
	if err := a.Docker.Check(); err != nil {
		return err
	}

	if len(a.Tags) == 0 {
		return fmt.Errorf("tags can not empty")
	}
	return nil
}

func (a Actuator) GetType() pipeline.TaskType {
	if a.Os != nil {
		return pipeline.OsType
	}
	if a.Kubernetes != nil {
		return pipeline.K8sType
	}
	if a.Docker != nil {
		return pipeline.DockerType
	}
	return ""
}

func (k *Kubernetes) Check() error {
	if k == nil {
		return nil
	}
	return nil
}

func (o *Os) Check() error {
	if o == nil {
		return nil
	}
	return nil
}

func (d *Docker) Check() error {
	if d == nil {
		return nil
	}
	return nil
}

type Kubernetes struct {
	Config string `yaml:"config"`

	Tunnel *Tunnel `yaml:"tunnel,omitempty"`
	Tls    *Tls    `yaml:"tls,omitempty"`
}

type Os struct {
	User     string `json:"user"`
	Ip       string `json:"ip"`
	Password string `yaml:"password,omitempty"`
	Rsa      string `yaml:"rsa,omitempty"`

	Tunnel Tunnel `yaml:"tunnel,omitempty"`
	Tls    *Tls   `yaml:"tls,omitempty"`
}

type Docker struct {
	Ip   string `json:"ip"`
	Port string `yaml:"port"`
	Ssh  *Os    `yaml:"ssh,omitempty"`

	Tunnel *Tunnel `yaml:"tunnel,omitempty"`
	Tls    *Tls    `yaml:"tls,omitempty"`
}

type Tunnel struct {
	ClientId    string `yaml:"clientId"`
	ClientToken string `yaml:"clientToken"`
}

type Tls struct {
	ServerPem string `yaml:"serverPem"`
	ServerKey string `yaml:"serverKey"`
	ClientPem string `yaml:"clientPem"`
}
