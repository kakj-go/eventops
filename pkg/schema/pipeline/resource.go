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

package pipeline

type Resources struct {
	Limit   *Limit   `yaml:"limit,omitempty"`
	Request *Request `yaml:"request,omitempty"`
}

type Limit struct {
	Cpu string `yaml:"cpu,omitempty"`
	Mem string `yaml:"mem,omitempty"`
}

type Request struct {
	Cpu string `yaml:"cpu,omitempty"`
	Mem string `yaml:"mem,omitempty"`
}
