package actuator

import (
	"encoding/json"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"tiggerops/apistructs"
	"tiggerops/pkg/schema/actuator"
	"tiggerops/pkg/token"
	"tiggerops/tools/eoctl/conf"
	"tiggerops/tools/eoctl/login"
	"time"
)

var applyFilePath string
var deleteFilePath string

var actuatorCmd = &cobra.Command{
	Use:   "actuator",
	Short: "Operate actuator",
	Long:  `You can perform a series of operations on the of the water line`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

var actuatorApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply actuator",
	Long:  `Example: eoctl actuator apply -f actuator.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		applyUser := login.GetEditUserInfo()

		content, err := os.ReadFile(applyFilePath)
		if err != nil {
			fmt.Printf("read file %v content error: %v \n", applyFilePath, err)
			os.Exit(1)
		}

		err = applyActuator(applyUser, string(content))
		if err != nil {
			fmt.Printf("apply actuator error: %v \n", err)
			os.Exit(1)
		}
	},
}

var actuatorDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete actuator",
	Long:  `Example: eoctl actuator delete -f actuator.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteUser := login.GetEditUserInfo()

		content, err := os.ReadFile(deleteFilePath)
		if err != nil {
			fmt.Printf("read file %v content error: %v \n", deleteFilePath, err)
			os.Exit(1)
		}

		var actuator actuator.Client
		err = yaml.Unmarshal(content, &actuator)
		if err != nil {
			fmt.Printf("unmarshal file %v content error: %v \n", deleteFilePath, err)
			os.Exit(1)
		}
		if actuator.Name == "" {
			fmt.Println("yaml content name can not empty")
			os.Exit(1)
		}
		err = deleteActuator(deleteUser, actuator.Name)
		if err != nil {
			fmt.Printf("delete actuator error: %v \n", err)
			os.Exit(1)
		}
	},
}

var actuatorListCmd = &cobra.Command{
	Use:   "list",
	Short: "list my actuator",
	Long:  `Example: eoctl actuator list`,
	Run: func(cmd *cobra.Command, args []string) {
		listUser := login.GetEditUserInfo()

		s, err := listMyActuator(listUser)
		if err != nil {
			fmt.Printf("list my actuator  error: %v \n", err)
			os.Exit(1)
		}
		jsonValue, err := json.Marshal(s)
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

func applyActuator(user *conf.UserInfo, content string) error {
	var resp Resp
	err := gout.
		POST(fmt.Sprintf("%s/%s", user.Server, "api/actuator/apply")).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		SetJSON(gout.H{"actuatorContent": content}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("apply actuator status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return nil
}

type ListMyActuatorResp struct {
	Status int
	Msg    string
	Data   []apistructs.Actuator
}

func listMyActuator(user *conf.UserInfo) ([]apistructs.Actuator, error) {
	var resp ListMyActuatorResp
	err := gout.
		GET(fmt.Sprintf("%s/api/actuator/", user.Server)).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("list my actuator status: %v, msg %s", resp.Status, resp.Msg)
	}

	return resp.Data, nil
}

func deleteActuator(user *conf.UserInfo, name string) error {
	var resp Resp
	err := gout.
		DELETE(fmt.Sprintf("%s/api/actuator/%s", user.Server, name)).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("delete actuator status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return nil
}

func BuildActuatorCmd() *cobra.Command {
	login.BindUserAndServerFlag(actuatorCmd)
	login.BindUserAndServerFlag(actuatorApplyCmd)
	login.BindUserAndServerFlag(actuatorDeleteCmd)
	login.BindUserAndServerFlag(actuatorListCmd)

	actuatorApplyCmd.PersistentFlags().StringVarP(&applyFilePath, "f", "f", "", "actuator defined file location")
	actuatorDeleteCmd.PersistentFlags().StringVarP(&deleteFilePath, "f", "f", "", "actuator defined file location")

	actuatorCmd.AddCommand(actuatorApplyCmd)
	actuatorCmd.AddCommand(actuatorDeleteCmd)
	actuatorCmd.AddCommand(actuatorListCmd)
	return actuatorCmd
}
