package conf

import (
	"flag"
)

var configPath *string

func parseFlag() {
	configPath = flag.String("configFile", "/etc/eventops/config.yaml", "Enter -configFile path to set the configuration file address, the default configuration file address is /etc/eventops/config.yaml")
	flag.Parse()
}
