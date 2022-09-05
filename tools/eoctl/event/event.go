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

package event

import (
	"eventops/apistructs"
	"eventops/internal/core/token"
	"eventops/tools/eoctl/conf"
	"eventops/tools/eoctl/login"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var sendFilePath string

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Mock send event",
	Long:  `You can use this command to simulate sending events`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var eventSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Mock send event",
	Long:  `Example: eoctl event send -f event.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		applyUser := login.GetEditUserInfo()

		content, err := os.ReadFile(sendFilePath)
		if err != nil {
			fmt.Printf("read file %v content error: %v \n", sendFilePath, err)
			os.Exit(1)
		}

		err = sendEvent(applyUser, content)
		if err != nil {
			fmt.Printf("send event error: %v \n", err)
			os.Exit(1)
		}
	},
}

type Resp struct {
	Status int
	Msg    string
	Data   interface{}
}

func sendEvent(user *conf.UserInfo, content []byte) error {
	var eventInfo apistructs.Event
	err := yaml.Unmarshal(content, &eventInfo)
	if err != nil {
		return err
	}

	var resp Resp
	err = gout.
		POST(fmt.Sprintf("%s/%s", user.Server, "api/event/send")).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		SetJSON(eventInfo).
		BindJSON(&resp).
		Do()
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("send event error status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return nil
}

func BuildEventCmd() *cobra.Command {
	login.BindUserAndServerFlag(eventCmd)
	login.BindUserAndServerFlag(eventSendCmd)

	eventSendCmd.PersistentFlags().StringVarP(&sendFilePath, "f", "f", "", "event file location")

	eventCmd.AddCommand(eventSendCmd)
	return eventCmd
}
