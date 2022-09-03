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

package actuator

import (
	"eventops/apistructs"
	"fmt"
)

type Client struct {
	Name string `json:"name"`

	Os         *Os         `yaml:"os,omitempty"`
	Kubernetes *Kubernetes `yaml:"kubernetes,omitempty"`
	Docker     *Docker     `yaml:"docker,omitempty"`

	Tunnel *Tunnel `yaml:"tunnel,omitempty"`

	Tags []string `yaml:"tags"`
}

func (a Client) GetTunnelClientID() string {
	if a.Tunnel == nil {
		return ""
	}
	return a.Tunnel.ClientId
}

func (a Client) GetTunnelClientToken() string {
	if a.Tunnel == nil {
		return ""
	}
	return a.Tunnel.ClientToken
}

func (a Client) Check() error {
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

	if err := a.Tunnel.check(); err != nil {
		return err
	}

	if len(a.Tags) == 0 {
		return fmt.Errorf("tags can not empty")
	}
	return nil
}

func (a Client) Mutating() error {
	if a.Os != nil {
		if a.Os.Port == "" {
			a.Os.Port = "22"
		}
	}

	if a.Docker != nil {
		if a.Docker.Ssh != nil {
			if a.Docker.Ssh.Port == "" {
				a.Docker.Ssh.Port = "22"
			}
		}
	}

	return nil
}

func (a Client) GetType() apistructs.TaskType {
	if a.Os != nil {
		return apistructs.OsType
	}
	if a.Kubernetes != nil {
		return apistructs.K8sType
	}
	if a.Docker != nil {
		return apistructs.DockerType
	}
	return ""
}

func (k *Kubernetes) Check() error {
	if k == nil {
		return nil
	}
	if k.Config == "" {
		return fmt.Errorf("kubernetes type actuator config can not empty")
	}
	return nil
}

func (o *Os) Check() error {
	if o == nil {
		return nil
	}

	if o.User == "" {
		return fmt.Errorf("os type actuator user can not empty")
	}
	if o.Password == "" {
		return fmt.Errorf("os type actuator password can not empty")
	}
	if o.Ip == "" {
		return fmt.Errorf("os type actuator ip can not empty")
	}
	return nil
}

func (d *Docker) Check() error {
	if d == nil {
		return nil
	}
	if d.Ip == "" {
		return fmt.Errorf("docker type actuator ip can not empty")
	}
	if d.Port == "" {
		return fmt.Errorf("docker type actuator port can not empty")
	}

	if err := d.Ssh.Check(); err != nil {
		return err
	}

	return nil
}

type Kubernetes struct {
	Config string `yaml:"config"`
}

type Os struct {
	User     string `json:"user"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Password string `yaml:"password,omitempty"`
}

type Docker struct {
	Ip   string `json:"ip"`
	Port string `yaml:"port"`
	Ssh  *Os    `yaml:"ssh,omitempty"`
}

type Tunnel struct {
	ClientId    string `yaml:"clientId"`
	ClientToken string `yaml:"clientToken"`
}

func (t *Tunnel) check() error {
	if t == nil {
		return nil
	}
	if t.ClientId == "" {
		return fmt.Errorf("tunnel clientId can not empty")
	}
	if t.ClientToken == "" {
		return fmt.Errorf("tunnel clientToken can not empty")
	}
	return nil
}
