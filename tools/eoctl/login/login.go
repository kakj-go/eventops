package login

import (
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"os"
	"tiggerops/apistructs"
	"tiggerops/tools/eoctl/conf"
	"time"
)

var Server string
var Username string
var Password string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login eventops api-server",
	Long:  `Example: eoctl login --server=value --username=userUsername --password=userPassword`,
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
		return apistructs.User{}, fmt.Errorf("login request error %v", err)
	}
	if resp.Status != 200 {
		return apistructs.User{}, fmt.Errorf("login failed status %v, msg %v", err, resp.Msg)
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
	loginCmd.PersistentFlags().StringVarP(&Server, "server", "s", "", "The eventops api-server address you want to sign in to. Example: --server http://api-server:port")
	loginCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "The eventops api-server username. Example: --username username")
	loginCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "The eventops api-server password. Example: --password password")
	return loginCmd
}
