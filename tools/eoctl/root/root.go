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

package root

import (
	"eventops/tools/eoctl/actuator"
	"eventops/tools/eoctl/conf"
	"eventops/tools/eoctl/event"
	"eventops/tools/eoctl/login"
	"eventops/tools/eoctl/pipeline"
	"eventops/tools/eoctl/register"
	"eventops/tools/eoctl/runtime"
	"eventops/tools/eoctl/trigger"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "eoctl",
	Short: "eoctl is the client of eventops API server",
	Long:  `Please read the eoctl documentation and log in before starting`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	rootCmd.AddCommand(login.BuildLoginCmd())
	rootCmd.AddCommand(register.BuildRegisterCmd())
	rootCmd.AddCommand(pipeline.BuildPipelineCmd())
	rootCmd.AddCommand(trigger.BuildTriggerCmd())
	rootCmd.AddCommand(actuator.BuildActuatorCmd())
	rootCmd.AddCommand(event.BuildEventCmd())
	rootCmd.AddCommand(runtime.BuildRuntimeCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&conf.CfgFile, "config", "", "config file (default is $HOME/.eoctl.yaml)")
	cobra.OnInitialize(conf.InitConfig)
}
