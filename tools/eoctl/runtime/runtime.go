package runtime

import (
	"encoding/json"
	"eventops/apistructs"
	"eventops/internal/core/token"
	"eventops/tools/eoctl/conf"
	"eventops/tools/eoctl/login"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var runtimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Get pipeline runtime",
	Long:  `You can use this command to get pipeline runtime`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var runtimeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pipeline runtime",
	Long:  `You can use this command to list pipeline runtime`,
	Run: func(cmd *cobra.Command, args []string) {
		if EventName != "" || EventVersion != "" || EventCreater != "" {
			if EventName == "" || EventVersion == "" || EventCreater == "" {
				fmt.Println("If you want to use event query, eventName, eventVersion and eventCreater cannot be empty")
				os.Exit(1)
			}
		}

		if PipelineDefinitionName != "" || PipelineDefinitionVersion != "" || PipelineDefinitionCreater != "" {
			if PipelineDefinitionName == "" || PipelineDefinitionVersion == "" || PipelineDefinitionCreater == "" {
				fmt.Println("If you want to use pipelineDefinition query, pipelineDefinitionName, pipelineDefinitionVersion and pipelineDefinitionCreater cannot be empty")
				os.Exit(1)
			}
		}

		listUser := login.GetEditUserInfo()
		result, err := ListPipelineRuntimes(listUser)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		resultJson, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println(string(resultJson))
	},
}

var runtimeGetDetailCmd = &cobra.Command{
	Use:   "get",
	Short: "Get pipeline runtime detail",
	Long:  `You can use this command to get pipeline runtime detail`,
	Run: func(cmd *cobra.Command, args []string) {
		if pipelineRuntimeId == "" {
			fmt.Println("id cannot be empty")
			os.Exit(1)
		}
		getUser := login.GetEditUserInfo()
		result, err := GetPipelineRuntimeDetail(getUser)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		resultJson, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println(string(resultJson))
	},
}

var runtimeCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel pipeline runtime detail",
	Long:  `You can use this command to cancel pipeline runtime`,
	Run: func(cmd *cobra.Command, args []string) {
		if pipelineRuntimeId == "" {
			fmt.Println("runtimeId cannot be empty")
			os.Exit(1)
		}

		cancelUser := login.GetEditUserInfo()
		result, err := CancelPipelineRuntimeDetail(cancelUser)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		resultJson, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println(string(resultJson))
	},
}

type ListResp struct {
	Status int
	Msg    string
	Data   []apistructs.Pipeline
}

var EventName string
var EventVersion string
var EventCreater string

var TriggerDefinitionName string

var PipelineDefinitionName string
var PipelineDefinitionVersion string
var PipelineDefinitionCreater string
var Top string

func ListPipelineRuntimes(user *conf.UserInfo) ([]apistructs.Pipeline, error) {
	var resp ListResp
	err := gout.
		GET(fmt.Sprintf("%s/%s", user.Server, fmt.Sprintf("api/pipeline/"))).
		SetQuery(gout.H{
			"eventName":                 EventName,
			"eventVersion":              EventVersion,
			"eventCreater":              EventCreater,
			"triggerDefinitionName":     TriggerDefinitionName,
			"pipelineDefinitionName":    PipelineDefinitionName,
			"pipelineDefinitionVersion": PipelineDefinitionVersion,
			"pipelineDefinitionCreater": PipelineDefinitionCreater,
			"top":                       Top,
		}).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("failed list pipeline runtime status: %v, msg: %s", resp.Status, resp.Msg)
	}

	return resp.Data, nil
}

var pipelineRuntimeId string

type GetResp struct {
	Status int
	Msg    string
	Data   apistructs.PipelineDetail
}

func GetPipelineRuntimeDetail(user *conf.UserInfo) (*apistructs.PipelineDetail, error) {
	var resp GetResp
	err := gout.
		GET(fmt.Sprintf("%s/%s", user.Server, fmt.Sprintf("api/pipeline/%v", pipelineRuntimeId))).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("failed get pipeline runtime detail status: %v, msg: %s", resp.Status, resp.Msg)
	}

	return &resp.Data, nil
}

type CancelResp struct {
	Status int
	Msg    string
	Data   string
}

func CancelPipelineRuntimeDetail(user *conf.UserInfo) (string, error) {
	var resp CancelResp
	err := gout.
		POST(fmt.Sprintf("%s/%s", user.Server, fmt.Sprintf("api/pipeline/%v/cancel", pipelineRuntimeId))).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return "", err
	}
	if resp.Status != 200 {
		return "", fmt.Errorf("failed cancel pipeline runtime status: %v, msg: %s", resp.Status, resp.Msg)
	}

	return resp.Data, nil
}

func BuildRuntimeCmd() *cobra.Command {
	login.BindUserAndServerFlag(runtimeCmd)
	login.BindUserAndServerFlag(runtimeListCmd)
	login.BindUserAndServerFlag(runtimeGetDetailCmd)
	login.BindUserAndServerFlag(runtimeCancelCmd)

	runtimeListCmd.PersistentFlags().StringVarP(&EventName, "en", "", "", "list pipeline runtime by eventName")
	runtimeListCmd.PersistentFlags().StringVarP(&EventVersion, "ev", "", "", "list pipeline runtime by eventVersion")
	runtimeListCmd.PersistentFlags().StringVarP(&EventCreater, "ec", "", "", "list pipeline runtime by eventCreater")

	runtimeListCmd.PersistentFlags().StringVarP(&TriggerDefinitionName, "tdn", "", "", "list pipeline runtime by triggerDefinitionName")

	runtimeListCmd.PersistentFlags().StringVarP(&PipelineDefinitionName, "pdn", "", "", "list pipeline runtime by pipelineDefinitionName")
	runtimeListCmd.PersistentFlags().StringVarP(&PipelineDefinitionVersion, "pdv", "", "", "list pipeline runtime by pipelineDefinitionVersion")
	runtimeListCmd.PersistentFlags().StringVarP(&PipelineDefinitionCreater, "pdc", "", "", "list pipeline runtime by pipelineDefinitionCreater")
	runtimeListCmd.PersistentFlags().StringVarP(&Top, "top", "", "20", "limit pipeline runtime result num. top max 100. default = 20")

	runtimeGetDetailCmd.PersistentFlags().StringVarP(&pipelineRuntimeId, "id", "", "", "get pipeline runtime detail by id")

	runtimeCancelCmd.PersistentFlags().StringVarP(&pipelineRuntimeId, "id", "", "", "cancel pipeline runtime by id")

	runtimeCmd.AddCommand(runtimeListCmd)
	runtimeCmd.AddCommand(runtimeGetDetailCmd)
	runtimeCmd.AddCommand(runtimeCancelCmd)
	return runtimeCmd
}
