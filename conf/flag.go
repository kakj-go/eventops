/*
 * Copyright 2022 The kakj-go Authors.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package conf

import (
	"flag"
)

var configPath *string

func parseFlag() {
	configPath = flag.String("configFile", "/etc/eventops/config.yaml", "Enter -configFile path to set the configuration file address, the default configuration file address is /etc/eventops/config.yaml")
	flag.Parse()
}
