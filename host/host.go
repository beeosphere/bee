package host

import (
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
var logLevel string

var initialized bool = false

func startup() {
	// ENVIRONMENT VARIABLES
	viper.SetEnvPrefix("bee")
	viper.BindEnv("id")
	viper.BindEnv("key")
	viper.BindEnv("hive")
	viper.BindEnv("config")
	viper.BindEnv("log")

	// COMMAND FLAGS
	pflag.StringVarP(&beeId, "id", "i", "", "Bee identifier")
	pflag.StringVarP(&beeKey, "key", "k", "", "Bee private seed key")
	pflag.StringVarP(&hiveUri, "hive", "h", "localhost:8080", "Hive URI address")
	pflag.StringVarP(&configPath, "config", "c", "", "Config file path")
	pflag.StringVarP(&logLevel, "log", "l", "info", "Log level (debug, info, warn, error)")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// CONFIG FILE
	cfgPath := viper.GetString("config")
	cfgFilename := "bee"
	cfgDir := "."

	if cfgPath != "" {
		// Prepare config file path
		cfgDir = filepath.Dir(cfgPath)
		cfgFilename = filepath.Base(cfgPath)
		cfgFilename = strings.TrimSuffix(cfgFilename, filepath.Ext(cfgFilename))

		// Set config file name and type in viper
		viper.SetConfigName(cfgFilename)
		viper.SetConfigType("yaml")
		viper.AddConfigPath(cfgDir)

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Errorf("Config file not found (%s)\n", cfgPath)
			} else {
				// Config file was found but another error was produced
				log.Error(err)
			}
			os.Exit(1)
		}

		log.Infof("CONFIG FILE: %s/%s.yaml\n", cfgDir, cfgFilename)
	}
	initialized = true
}

func GetOptions() *core.BeeOptions {
	if !initialized {
		startup()
	}
	options := &core.BeeOptions{
		LogLevel: viper.GetString("log"),
		Id:       viper.GetString("id"),
		Key:      viper.GetString("key"),
		HiveUri:  viper.GetString("hive"),
	}
	return options
}
