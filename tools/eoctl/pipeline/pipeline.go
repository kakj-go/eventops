package pipeline

import (
	"encoding/json"
	"eventops/apistructs"
	"eventops/internal/core/token"
	"eventops/pkg/schema/pipeline"
	"eventops/tools/eoctl/conf"
	"eventops/tools/eoctl/login"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Operate pipeline definition",
	Long:  `You can perform a series of operations on the definition of the water line`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

var pipelineApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply pipeline definition",
	Long:  `Example: eoctl pipeline apply -f pipelineDefinition.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		applyUser := login.GetEditUserInfo()

		content, err := os.ReadFile(applyFilePath)
		if err != nil {
			fmt.Printf("read file %v content error: %v \n", applyFilePath, err)
			os.Exit(1)
		}

		err = applyPipelineDefinition(applyUser, string(content))
		if err != nil {
			fmt.Printf("apply pipeline definition error: %v \n", err)
			os.Exit(1)
		}
	},
}

var pipelineDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete pipeline definition",
	Long:  `Example: eoctl pipeline delete -f pipelineDefinition.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteUser := login.GetEditUserInfo()

		content, err := os.ReadFile(deleteFilePath)
		if err != nil {
			fmt.Printf("read file %v content error: %v \n", deleteFilePath, err)
			os.Exit(1)
		}

		var pipeInfo pipeline.Pipeline
		err = yaml.Unmarshal(content, &pipeInfo)
		if err != nil {
			fmt.Printf("unmarshal file %v content error: %v \n", deleteFilePath, err)
			os.Exit(1)
		}
		if pipeInfo.Name == "" {
			fmt.Println("yaml content name can not empty")
			os.Exit(1)
		}
		if pipeInfo.Version == "" {
			fmt.Println("yaml content version field can not empty")
			os.Exit(1)
		}

		err = deletePipelineDefinition(deleteUser, pipeInfo.Name, pipeInfo.Version)
		if err != nil {
			fmt.Printf("delete pipeline definition error: %v \n", err)
			os.Exit(1)
		}
	},
}

var pipelineListCmd = &cobra.Command{
	Use:   "list",
	Short: "list my pipeline definition",
	Long:  `Example: eoctl pipeline list`,
	Run: func(cmd *cobra.Command, args []string) {
		listUser := login.GetEditUserInfo()
		definitions, err := listMyPipelineDefinition(listUser)
		if err != nil {
			fmt.Printf("list my pipeline definition error: %v \n", err)
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

var pipelineDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "describe pipeline definition",
	Long:  `Example: eoctl pipeline describe --name logo --version 1.2`,
	Run: func(cmd *cobra.Command, args []string) {
		describeUser := login.GetEditUserInfo()
		if describePipelineName == "" {
			fmt.Println("describe name can not empty")
			os.Exit(1)
		}

		var result interface{}
		var err error
		if describePipelineVersion != "" {
			result, err = describePipelineVersionInfo(describeUser, describePipelineName, describePipelineVersion)
		} else {
			result, err = describePipelineInfo(describeUser, describePipelineName)
		}

		if err != nil {
			fmt.Printf("describe pipeline definition error: %v \n", err)
			os.Exit(1)
		}
		jsonValue, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("json marshal result error: %v \n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonValue))
		return

	},
}

type ApplyPipelineVersionResp struct {
	Status int
	Msg    string
	Data   interface{}
}

func applyPipelineDefinition(user *conf.UserInfo, content string) error {
	var resp ApplyPipelineVersionResp
	err := gout.
		POST(fmt.Sprintf("%s/%s", user.Server, "api/pipeline-definition/apply")).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		SetJSON(gout.H{"pipelineContent": content}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("apply pipeline definition status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return nil
}

type DeletePipelineVersionResp struct {
	Status int
	Msg    string
	Data   interface{}
}

func deletePipelineDefinition(user *conf.UserInfo, name, version string) error {
	var resp DeletePipelineVersionResp
	err := gout.
		DELETE(fmt.Sprintf("%s/%s", user.Server, fmt.Sprintf("api/pipeline-definition/%s/%s", name, version))).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("delete pipeline definition status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return nil
}

type ListMyPipelineVersionResp struct {
	Status int
	Msg    string
	Data   []apistructs.PipelineVersionDefinition
}

func listMyPipelineDefinition(user *conf.UserInfo) ([]apistructs.PipelineVersionDefinition, error) {
	var resp ListMyPipelineVersionResp
	err := gout.
		GET(fmt.Sprintf("%s/%s", user.Server, fmt.Sprintf("api/pipeline-definition/"))).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("list my pipeline definition status: %v, msg: %s", resp.Status, resp.Msg)
	}

	return resp.Data, nil
}

type describePipelineVersionInfoResp struct {
	Status int
	Msg    string
	Data   apistructs.PipelineVersionDefinition
}

func describePipelineVersionInfo(user *conf.UserInfo, name string, version string) (*apistructs.PipelineVersionDefinition, error) {
	var resp describePipelineVersionInfoResp
	err := gout.
		GET(fmt.Sprintf("%s/%s", user.Server, fmt.Sprintf("api/pipeline-definition/%s/%s", name, version))).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("get pipeline definition status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return &resp.Data, nil
}

type describePipelineInfoResp struct {
	Status int
	Msg    string
	Data   apistructs.PipelineDefinition
}

func describePipelineInfo(user *conf.UserInfo, name string) (*apistructs.PipelineDefinition, error) {
	var resp describePipelineInfoResp
	err := gout.
		GET(fmt.Sprintf("%s/%s", user.Server, fmt.Sprintf("api/pipeline-definition/%s", name))).
		SetHeader(gout.H{"sid": fmt.Sprintf("%x", time.Now().UnixNano()), token.AuthHeader: token.BuildTokenHeaderValue(user.Token)}).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("describe pipeline definition status: %v, msg: %s", resp.Status, resp.Msg)
	}
	return &resp.Data, nil
}

var applyFilePath string
var deleteFilePath string

var describePipelineName string
var describePipelineVersion string

func BuildPipelineCmd() *cobra.Command {
	login.BindUserAndServerFlag(pipelineCmd)
	login.BindUserAndServerFlag(pipelineApplyCmd)
	login.BindUserAndServerFlag(pipelineDeleteCmd)
	login.BindUserAndServerFlag(pipelineListCmd)
	login.BindUserAndServerFlag(pipelineDescribeCmd)

	pipelineApplyCmd.PersistentFlags().StringVarP(&applyFilePath, "f", "f", "", "pipeline defined file location")
	pipelineDeleteCmd.PersistentFlags().StringVarP(&deleteFilePath, "f", "f", "", "pipeline defined file location")

	pipelineDescribeCmd.PersistentFlags().StringVarP(&describePipelineName, "name", "n", "", "pipeline definition name")
	pipelineDescribeCmd.PersistentFlags().StringVarP(&describePipelineVersion, "version", "v", "", "pipeline definition version")

	pipelineCmd.AddCommand(pipelineApplyCmd)
	pipelineCmd.AddCommand(pipelineDeleteCmd)
	pipelineCmd.AddCommand(pipelineListCmd)
	pipelineCmd.AddCommand(pipelineDescribeCmd)
	return pipelineCmd
}
