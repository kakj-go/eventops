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
