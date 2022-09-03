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

package login

import (
	"eventops/apistructs"
	"eventops/tools/eoctl/conf"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var Server string
var Username string
var Password string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login eventops",
	Long:  `Example: eoctl login --server=xxx --username=xxx --password=xxx`,
	Run: func(cmd *cobra.Command, args []string) {
		if Server == "" {
			fmt.Println("The --server declaration was not found")
			os.Exit(1)
		}
		if Username == "" {
			fmt.Println("The --username declaration was not found")
			os.Exit(1)
		}
		if Password == "" {
			fmt.Println("The --password declaration was not found")
			os.Exit(1)
		}

		user, err := login(Server, Username, Password)
		if err != nil {
			fmt.Println("login failed error: ", err)
			os.Exit(1)
		}

		cfg, err := conf.NewConfig()
		if err != nil {
			fmt.Println("failed to get config error: ", err)
			os.Exit(1)
		}

		err = cfg.SetUser(user, Server)
		if err != nil {
			fmt.Println("login failed error: ", err)
			os.Exit(1)
		}

		fmt.Println("login success")
	},
}

type Resp struct {
	Status int
	Msg    string
	Data   apistructs.User
}

func login(server, username, password string) (apistructs.User, error) {
	var resp Resp
	err := gout.
		POST(fmt.Sprintf("%s/%s", server, "api/user/login")).
		SetHeader(gout.H{"X-IP": "127.0.0.1", "sid": fmt.Sprintf("%x", time.Now().UnixNano())}).
		SetJSON(gout.H{"username": username, "password": password}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return apistructs.User{}, fmt.Errorf("login request error: %v", err)
	}
	if resp.Status != 200 {
		return apistructs.User{}, fmt.Errorf("login failed status %v, msg: %v", err, resp.Msg)
	}
	return resp.Data, nil
}

func GetEditUserInfo() *conf.UserInfo {
	cfg, err := conf.NewConfig()
	if err != nil {
		fmt.Println("failed to get config error: ", err)
		os.Exit(1)
	}

	defaultUser := cfg.GetDefaultUser()
	var editUsername = Username
	var editServer = Server
	if editUsername == "" && defaultUser != nil {
		editUsername = defaultUser.Username
	}
	if editServer == "" && defaultUser != nil {
		editServer = defaultUser.Server
	}

	return cfg.GetUser(editUsername, editServer)
}

func BindUserAndServerFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "Declare which user is used to save the pipeline definition")
	cmd.PersistentFlags().StringVarP(&Server, "server", "s", "", "Declare to save a pipeline definition on that server")
}

func BuildLoginCmd() *cobra.Command {
	loginCmd.PersistentFlags().StringVarP(&Server, "server", "s", "", "The eventops address you want to sign in to. Example: --server http://ip:port")
	loginCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "The eventops username. Example: --username username")
	loginCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "The eventops password. Example: --password password")
	return loginCmd
}
