package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/beeosphere/bee/core/ebus"
	"github.com/sirupsen/logrus"
)

// type BeeOptions struct {
// 	Id       string
// 	Key      string
// 	HiveUri  string
// 	LogLevel string
// }

var log logrusLogger

type ConfigOptions struct {
	Agent  *AgentOptions            `koanf:"agent"`
	Agents map[string]*AgentOptions `koanf:"agents"`
}

type AgentOptions struct {
	Id       string         `koanf:"id"`
	Key      string         `koanf:"key"`
	HiveUri  string         `koanf:"hive"`
	LogLevel string         `koanf:"log"`
	Source   *SourceOptions `koanf:"config"`
}

type SourceOptions struct {
	Path  string `koanf:"path"`
	Watch bool   `koanf:"watch"`
}

type AgentManifest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type BeeEngine struct {
	session     *session
	provisioner Provisioner
	controller  Controller
	manifest    AgentManifest
}

var busClient *bus

var eBus ebus.Bus

// var eBus EmbedBus

// type PrefixedFormatter struct {
// 	Prefix string
// 	log.TextFormatter
// }

// func (f *PrefixedFormatter) Format(entry *log.Entry) ([]byte, error) {
// 	bytes, _ := f.TextFormatter.Format(entry)
// 	return append(bytes, []byte(f.Prefix+" > ")...), nil
// }

// Opt is a configuration option to initialize a client
type Opt func(*BeeEngine) error

func WithManifest(manifest AgentManifest) Opt {
	return func(c *BeeEngine) error {
		c.manifest = manifest
		return nil
	}
}

func NewBee(options *AgentOptions, provider ChannelProvider, ops ...Opt) *BeeEngine {

	if !validateOptions(options) {
		os.Exit(1)
	}

	// log.SetFormatter(&PrefixedFormatter{Prefix: options.Id})
	// log.SetLevel(convertLogLevel(options.LogLevel))
	log = *newLogrusLogger()
	log.SetLevel(convertLogLevel(options.LogLevel))
	log.SetPrefix("core", "")

	session := newSession(options)
	http := NewHttpClient(session)

	busClient = newBus(session)

	var provisioner Provisioner
	if session.IsEmbedded() {
		provisioner = newLocalProvisioner(session)
	} else {
		provisioner = newRemoteProvisioner(session, http)
	}

	engine := &BeeEngine{
		session:     session,
		provisioner: provisioner,
		controller:  newController(session, http, provider),
	}

	for _, op := range ops {
		op(engine)
	}

	log.Infof("BEE ID:           %s", session.bee)
	log.Infof("BEE TYPE:         %s (version: %s)", engine.manifest.Name, engine.manifest.Version)
	// log.Infof("BEE VERSION:      %s", )
	if session.IsEmbedded() {
		log.Infof("MODE:             Standalone")
		log.Infof("DATA SOURCE:      %s", session.configPath)
	} else {
		log.Infof("MODE:             Hive bus")
		log.Infof("HIVE HOST:        %s", session.hiveAddress)
	}
	log.Infof("LOG LEVEL:        %s", options.LogLevel)

	log.Info("")
	log.Info("           oO              ")
	log.Info(" Zzzz..  -{||\")   Bee agent is flying!")
	log.Info("")
	log.Info("---------------------------------------------------")

	return engine
}

func (e *BeeEngine) Buzz() error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	bus := ebus.New()
	// bus := NewEmbedBus()

	return e.EmbededBuzz(ctx, bus)

	// err := e.connect(ctx)
	// if err != nil {
	// 	log.Error(err)
	// 	return err
	// }
	// <-ctx.Done()

	// fmt.Println("")

	// err = e.disconnect()
	// if err != nil {
	// 	log.Error(err)
	// 	return err
	// }
	// return nil
}

func (e *BeeEngine) EmbededBuzz(ctx context.Context, embeddedBus ebus.Bus) error {
	if eBus == nil {
		eBus = embeddedBus
	}

	err := e.connect(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	<-ctx.Done()

	// fmt.Println("")

	err = e.disconnect()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// func (e *BeeEngine) StartBuzzing(ctx context.Context, embeddedBus ebus.Bus) error {
// 	if eBus == nil {
// 		eBus = embeddedBus
// 	}

// 	fmt.Printf("Bee %s is starting to buzz...\n", e.session.bee)
// 	err := e.connect(ctx)
// 	if err != nil {
// 		log.Error(err)
// 		return err
// 	}
// 	fmt.Printf("Bee %s is buzzing...\n", e.session.bee)
// 	return nil
// }

// func (e *BeeEngine) StopBuzzing() error {
// 	fmt.Println("")

// 	err := e.disconnect()
// 	if err != nil {
// 		log.Error(err)
// 		return err
// 	}
// 	return nil
// }

func (e *BeeEngine) connect(ctx context.Context) error {

	// Step 1. Connects to BeeOS Hive and authenticates using NKeys
	if err := e.provisioner.OpenSession(ctx); err != nil {
		return err
	}

	// Step 2. Subscribes to connector system topics
	if err := e.controller.Startup(); err != nil {
		return err
	}

	// <-e.controller.Started()

	return nil
}

func (e *BeeEngine) disconnect() error {
	defer log.Info("Bee disconnected")

	return e.controller.Shutdown()
}

func validateOptions(options *AgentOptions) bool {
	o := options
	messages := []string{}
	// Validate required parameters
	if o.Id == "" {
		messages = append(messages, "Bee ID")
	}
	if o.Key == "" && o.Source == nil {
		messages = append(messages, "Private key")
	}
	if o.HiveUri == "" && o.Source == nil {
		messages = append(messages, "Hive address")
	}
	if o.Source != nil && o.Source.Path == "" && o.HiveUri == "" {
		messages = append(messages, "Data source path")
	}
	if len(messages) > 0 {
		log.Error(fmt.Sprintf("Missing required params: %s", strings.Join(messages, ", ")))
		return false
	}
	return true
}

func convertLogLevel(level string) logrus.Level {
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

func formattedHiveUri(hive string) string {
	if strings.HasPrefix(hive, "http") {
		return hive
	}
	return fmt.Sprintf("http://%s", hive) // TODO: http or https...
}

// log.Info("                           ")
// log.Info("      ████  ████           ")
// log.Info("    ██    ██    ██         ")
// log.Info("      ██    ██  ██         ")
// log.Info("        ██████████         ")
// log.Info("      ████░░██░░░░██       ")
// log.Info("    ██░░██░░██░░░░░░▓▓     ")
// log.Info("▓▓▓▓██░░██░░██░░▓▓░░██     ")
// log.Info("    ██░░██░░██░░░░░░██     ")
// log.Info("      ████░░██░░░░██       ")
// log.Info("        ██████████         ")
// log.Info("                           ")

// log.Info("                      oO //      ")
// log.Info(" bzzz''-.._.-''-.._ -(||)(')     ")
// log.Info("                     '''         ")

// log.Info("   oO //     ")
// log.Info("-(||)(')     ")
// log.Info(" '''         ")

// log.Info("")
// log.Info(fmt.Sprintf("  oO     Bee: %s", options.Id))
// log.Info(fmt.Sprintf("-{||')   Hive: %s", options.HiveUri))
// log.Info("")
