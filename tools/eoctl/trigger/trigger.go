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

package trigger

import (
	"encoding/json"
	"eventops/apistructs"
	"eventops/internal/core/token"
	"eventops/pkg/schema/event"
	"eventops/tools/eoctl/conf"
	"eventops/tools/eoctl/login"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var applyFilePath string
var deleteFilePath string

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Operate trigger definition",
	Long:  `You can perform a series of operations on the definition of the water line`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

var triggerApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply trigger definition",
	Long:  `Example: eoctl trigger apply -f triggerDefinition.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		applyUser := login.GetEditUserInfo()

		content, err := os.ReadFile(applyFilePath)
		if err != nil {
			fmt.Printf("read file %v content error: %v \n", applyFilePath, err)
			os.Exit(1)
		}

		err = applyTriggerDefinition(applyUser, string(content))
		if err != nil {
			fmt.Printf("apply trigger definition error: %v \n", err)
			os.Exit(1)
		}
	},
}

var triggerDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete trigger definition",
	Long:  `Example: eoctl trigger delete -f triggerDefinition.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteUser := login.GetEditUserInfo()

		content, err := os.ReadFile(deleteFilePath)
		if err != nil {
			fmt.Printf("read file %v content error: %v \n", deleteFilePath, err)
			os.Exit(1)
		}

		var trigger event.Trigger
		err = yaml.Unmarshal(content, &trigger)
		if err != nil {
			fmt.Printf("unmarshal file %v content error: %v \n", deleteFilePath, err)
			os.Exit(1)
		}
		if trigger.Name == "" {
			fmt.Println("yaml content name can not empty")
			os.Exit(1)
		}
		err = deleteTriggerDefinition(deleteUser, trigger.Name)
		if err != nil {
			fmt.Printf("delete trigger definition error: %v \n", err)
			os.Exit(1)
		}
	},
}

var triggerListCmd = &cobra.Command{
	Use:   "list",
	Short: "list my trigger definition",
	Long:  `Example: eoctl trigger list`,
	Run: func(cmd *cobra.Command, args []string) {
		listUser := login.GetEditUserInfo()

		definitions, err := listMyTriggerDefinition(listUser)
		if err != nil {
			fmt.Printf("list my trigger definition error: %v \n", err)
			os.Exit(1)
		}
		jsonValue, err := json.Marshal(definitions)
		if err != nil {
			fmt.Printf("json marshal result error: %v \n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonValue))
	},
}

type Resp struct {
	Status int
	Msg    string
	Data   interface{}
}

func applyTriggerDefinition(user *conf.UserInfo, content string) error {
	var resp Resp
	err := gout.
		POST(fmt.Sprintf("%s/%s", user.Server, "api/trigger-definition/apply")).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		SetJSON(gout.H{"triggerContent": content}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("apply trigger definition status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return nil
}

type ListMyTriggerDefinitionResp struct {
	Status int
	Msg    string
	Data   []apistructs.EventTriggerDefinition
}

func listMyTriggerDefinition(user *conf.UserInfo) ([]apistructs.EventTriggerDefinition, error) {
	var resp ListMyTriggerDefinitionResp
	err := gout.
		GET(fmt.Sprintf("%s/api/trigger-definition/", user.Server)).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("list my trigger definition status: %v, msg: %s", resp.Status, resp.Msg)
	}

	return resp.Data, nil
}

func deleteTriggerDefinition(user *conf.UserInfo, name string) error {
	var resp Resp
	err := gout.
		DELETE(fmt.Sprintf("%s/api/trigger-definition/%s", user.Server, name)).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("delete triggers definition status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return nil
}

func BuildTriggerCmd() *cobra.Command {
	login.BindUserAndServerFlag(triggerCmd)
	login.BindUserAndServerFlag(triggerApplyCmd)
	login.BindUserAndServerFlag(triggerDeleteCmd)
	login.BindUserAndServerFlag(triggerListCmd)

	triggerApplyCmd.PersistentFlags().StringVarP(&applyFilePath, "f", "f", "", "trigger defined file location")
	triggerDeleteCmd.PersistentFlags().StringVarP(&deleteFilePath, "f", "f", "", "trigger defined file location")

	triggerCmd.AddCommand(triggerApplyCmd)
	triggerCmd.AddCommand(triggerDeleteCmd)
	triggerCmd.AddCommand(triggerListCmd)
	return triggerCmd
}
