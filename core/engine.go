package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	log "github.com/sirupsen/logrus"
)

type BeeOptions struct {
	Id       string
	Key      string
	HiveUri  string
	LogLevel string
}

type BeeEngine struct {
	session     *session
	provisioner *provisioner
	controller  *controller
}

var busClient *bus

func NewBee(options *BeeOptions, provider ChannelProvider) *BeeEngine {

	if !validateOptions(options) {
		os.Exit(1)
	}

	log.SetLevel(convertLogLevel(options.LogLevel))

	log.Infof("BEE ID:      %s", options.Id)
	log.Infof("HIVE HOST:   %s", options.HiveUri)
	log.Infof("LOG LEVEL:   %s", options.LogLevel)
	log.Info("")
	log.Info("           oO              ")
	log.Info(" Zzzz..  -{||\")   Bee agent is flying!")
	log.Info("")

	session := newSession(options.Id, options.Key, formattedHiveUri(options.HiveUri))
	http := NewHttpClient(session)

	busClient = newBus(session)

	return &BeeEngine{
		session:     session,
		provisioner: newProvisioner(session, http),
		controller:  newController(session, http, provider),
	}
}

func (e *BeeEngine) Buzz() error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := e.connect(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	<-ctx.Done()

	fmt.Println("")

	err = e.disconnect()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (e *BeeEngine) connect(ctx context.Context) error {

	// Step 1. Connects to BeeOS Server and authenticates using NKeys
	if err := e.provisioner.openSession(ctx); err != nil {
		return err
	}

	// Step 2. Subscribes to connector system topics
	if err := e.controller.startup(); err != nil {
		return err
	}

	return nil
}

func (e *BeeEngine) disconnect() error {
	defer log.Info("Bee disconnected")

	return e.controller.shutdown()
}

func validateOptions(options *BeeOptions) bool {
	messages := []string{}
	// Validate required parameters
	if options.Id == "" {
		messages = append(messages, "Bee ID")
	}
	if options.Key == "" {
		messages = append(messages, "Private key")
	}
	if options.HiveUri == "" {
		messages = append(messages, "Hive address")
	}
	if len(messages) > 0 {
		log.Error(fmt.Sprintf("Missing required params: %s", strings.Join(messages, ", ")))
		return false
	}
	return true
}

func convertLogLevel(level string) log.Level {
	switch level {
	case "trace":
		return log.TraceLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
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
