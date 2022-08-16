package root

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"tiggerops/tools/eoctl/conf"
	"tiggerops/tools/eoctl/login"
	"tiggerops/tools/eoctl/pipeline"
	"tiggerops/tools/eoctl/trigger"
)

var rootCmd = &cobra.Command{
	Use:   "eoctl",
	Short: "eoctl is the client of eventops API server",
	Long:  `Please read the eoctl documentation and log in before starting`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	rootCmd.AddCommand(login.BuildLoginCmd())
	rootCmd.AddCommand(pipeline.BuildPipelineCmd())
	rootCmd.AddCommand(trigger.BuildTriggerCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&conf.CfgFile, "config", "", "config file (default is $HOME/.eoctl.yaml)")
	cobra.OnInitialize(conf.InitConfig)
}
