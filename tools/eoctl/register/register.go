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

package register

import (
	"eventops/apistructs"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register eventops user",
	Long:  `Example: eoctl register --server=value --username=xxx --password=xxx --email=xxx`,
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
		if Email == "" {
			fmt.Println("The --email declaration was not found")
			os.Exit(1)
		}
		err := register()
		if err != nil {
			fmt.Printf("register error: %v \n", err)
			os.Exit(1)
		}

		fmt.Println("register success")
	},
}

type Resp struct {
	Status int
	Msg    string
	Data   interface{}
}

func register() error {
	var resp Resp

	var userRegister apistructs.UserRegisterInfo
	userRegister.Name = Username
	userRegister.Email = Email
	userRegister.Password = Password

	err := gout.
		POST(fmt.Sprintf("%s/%s", Server, "api/user/register")).
		SetHeader(gout.H{"X-IP": "127.0.0.1", "sid": fmt.Sprintf("%x", time.Now().UnixNano())}).
		SetJSON(userRegister).
		BindJSON(&resp).
		Do()
	if err != nil {
		return fmt.Errorf("register request error: %v", err)
	}
	if resp.Status != 200 {
		return fmt.Errorf("register failed status: %v, msg: %v", err, resp.Msg)
	}
	return nil
}

var Server string
var Username string
var Password string
var Email string

func BuildRegisterCmd() *cobra.Command {
	registerCmd.PersistentFlags().StringVarP(&Server, "server", "s", "", "The eventops address you want to register in to. Example: --server http://ip:port")
	registerCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "The eventops user username. Example: --username username")
	registerCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "The eventops user password. Example: --password password")
	registerCmd.PersistentFlags().StringVarP(&Email, "email", "e", "", "The eventops user email. Example: --email email")
	return registerCmd
}
