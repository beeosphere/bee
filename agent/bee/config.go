package bee

import (
	"fmt"
	"os"
	"strings"

	"github.com/beeosphere/bee/agent/models"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var k = koanf.New(".")

var options *models.ConfigOptions = nil

func initialize() {

	f := flag.NewFlagSet("cfg", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.StringP("config", "c", "", "path to bee config file")
	f.Lookup("config").NoOptDefVal = "config.toml"
	f.StringP("agent.id", "i", "", "bee identifier")
	f.StringP("agent.key", "k", "", "bee private seed key")
	f.StringP("agent.hive", "h", "", "hive URI address")
	f.StringP("agent.log", "l", "", "log level (debug, info, warn, error)")

	f.Parse(os.Args[1:])

	configPath, _ := f.GetString("config")

	// Load TOML config if flag "config" is set
	if configPath != "" {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			log.Fatalf("[core] error loading config: %v", err)
		}
	}

	// Load command line flags
	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		log.Fatalf("[core] error loading config: %v", err)
	}

	// Load env vars
	k.Load(env.Provider("BEE_", ".", func(s string) string {
		s = strings.Replace(s, "BEE_", "AGENT_", 1)
		s = strings.Replace(strings.ToLower(s), "_", ".", -1)
		return s
	}), nil)

	options = &models.ConfigOptions{}
	k.UnmarshalWithConf("", options, koanf.UnmarshalConf{Tag: "koanf"})

	if configPath != "" {
		log.Info("[core] ---------------------------------------------------")
		log.Infof("[core] CONFIG FILE:      %s\n", configPath)
		log.Info("[core] ---------------------------------------------------")
	}

	// fmt.Printf("config: %+v\n\n", k.All())

	// fmt.Printf("keys: %+v\n\n", k.MapKeys("agents"))

	// fmt.Printf("options: %+v\n", options)

}

func getOptions() *models.ConfigOptions {
	if options == nil {
		initialize()
	}
	return options
}
