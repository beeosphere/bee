package host

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beeosphere/bee/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var beeId string = "bee"
var beeKey string
var hiveUri string
var configPath string

var initialized bool = false

func startup() {
	// ENVIRONMENT VARIABLES
	viper.SetEnvPrefix("bee")
	viper.BindEnv("id")
	viper.BindEnv("key")
	viper.BindEnv("hive")
	viper.BindEnv("config")

	// COMMAND FLAGS
	pflag.StringVarP(&beeId, "id", "i", "", "Bee identifier")
	pflag.StringVarP(&beeKey, "key", "k", "", "Bee private seed key")
	pflag.StringVarP(&hiveUri, "hive", "h", "localhost:8080", "Hive URI address")
	pflag.StringVarP(&configPath, "config", "c", "", "Config file path")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// CONFIG FILE
	cfgPath := viper.GetString("config")
	cfgFilename := "bee"
	cfgDir := "."
	if cfgPath != "" {
		cfgDir = filepath.Dir(cfgPath)
		cfgFilename = filepath.Base(cfgPath)
		cfgFilename = strings.TrimSuffix(cfgFilename, filepath.Ext(cfgFilename))
	}

	if cfgPath != "" {

		viper.SetConfigName(cfgFilename)
		viper.SetConfigType("yaml")
		// TODO: Add config path from flats/envvars
		viper.AddConfigPath(cfgDir)
		// if err := viper.ReadInConfig(); err != nil {
		// 	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// 		// Config file not found; ignore error if desired
		// 		if cfgPath != "" {
		// 			log.Errorf("Config file: %s/%s.yaml not found\n", cfgDir, cfgFilename)
		// 			os.Exit(0)
		// 		}
		// 	} else {
		// 		// Config file was found but another error was produced
		// 		log.Error(err)
		// 		os.Exit(0)
		// 	}
		// }
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				// Config file was found but another error was produced
				log.Error(err)
				os.Exit(0)
			}
		}
	}

	if cfgPath != "" {
		log.Infof("Using config file: %s/%s.yaml\n", cfgDir, cfgFilename)
	} else {
		log.Infof("A config file is not being used\n")
	}

	initialized = true
}

func GetOptions() *core.BeeOptions {
	if !initialized {
		startup()
	}
	options := &core.BeeOptions{
		Id:      viper.GetString("id"),
		Key:     viper.GetString("key"),
		HiveUri: fmt.Sprintf("http://%s", viper.GetString("hive")), // TODO: http or https...
	}
	return options
}
