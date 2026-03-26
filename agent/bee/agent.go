package bee

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/beeosphere/bee/agent/internal/adapters/busnats"
	"github.com/beeosphere/bee/agent/internal/adapters/storefs"
	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/internal/runtime"
	"github.com/beeosphere/bee/agent/models"
	"github.com/sirupsen/logrus"
)

type Agent interface {
	Run() error
	Stop() error
}

type agentEngine struct {
	environment       *models.AgentOptions
	session           *core.Session
	log               models.Logger
	hosted            bool
	manifest          models.AgentManifest
	bus               core.Bus
	store             models.Store
	authenticator     runtime.Authenticator
	controller        *runtime.Controller
	agentProvider     models.AgentProvider
	connectorProvider models.ConnectorProvider
	// controller  Controller
}

func NewAgent(ops ...Opt) Agent {

	agent := &agentEngine{}
	for _, op := range ops {
		op(agent)
	}

	// ENVIRONMENT
	if agent.environment == nil {
		// Get from config system (koanf)
		config := getOptions()
		if config == nil || config.Agent == nil {
			agent.log.Errorf("No configuration found")
			os.Exit(1)
		}
		agent.environment = config.Agent
	}

	if !agent.validateEnvironment() {
		os.Exit(1)
	}

	// LOGGER
	if agent.log == nil {
		logger := core.NewLogrusLogger()
		logger.SetLevel(convertLogLevel(agent.environment.LogLevel))
		logger.SetPrefix("core", "")
		agent.log = logger
	}

	// SESSION
	agent.session = core.NewSession(agent.environment, &agent.manifest, agent.log)

	// HTTP CLIENT
	httpClient := core.NewHttpClient(agent.session)

	// BUS
	if agent.bus == nil {
		if agent.session.IsEmbedded() {
			// engine.bus = buslocal.NewBus(engine.session, engine.log)
			agent.log.Error("embedded bus not implemented")
			os.Exit(1)
		} else {
			agent.bus = busnats.NewBus(agent.session)
		}
	}

	// STORE
	if agent.store == nil {
		basePath := agent.environment.Path
		if basePath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				agent.log.Errorf("Failed to get current working directory: %v", err)
				os.Exit(1)
			}
			basePath = filepath.Join(cwd, "data")
		}
		agent.store = storefs.NewFileSystemStore(agent.session, basePath)
	}

	// AUTHENTICATOR
	if agent.authenticator == nil {
		if agent.session.IsEmbedded() {
			agent.authenticator = runtime.NewLocalAuthenticator(agent.session)
		} else {
			agent.authenticator = runtime.NewTwoStepAuthenticator(agent.session, httpClient)
		}
	}

	// COMMANDER
	commander := runtime.NewCommander(agent.session, agent.bus)

	// SYNCHRONIZER
	synchronizer := runtime.NewSynchronizer(agent.session, httpClient, commander, agent.store)

	// PROCESSOR
	var processor runtime.Processor
	if agent.agentProvider != nil {
		processor = runtime.NewAgentProcessor(agent.session, agent.agentProvider, commander, agent.bus)
	} else if agent.connectorProvider != nil {
		processor = runtime.NewConnectorProcessor(agent.session, agent.connectorProvider, agent.bus)
	} else {
		agent.log.Error("Missing provider for agent or connector")
		os.Exit(1)
	}

	// CONTROLLER
	agent.controller = runtime.NewController(agent.session, agent.bus, synchronizer, commander, processor)

	agent.logConfig(agent.environment)
	agent.environment = nil // IMPORTANT: Clear sensitive info. no longer needed. It is in session

	return agent
}

func (a *agentEngine) Run() error {

	ctx := context.Background()

	// Step 1. Connects to BeeOS Hive and authenticates using NKeys
	if err := a.authenticator.OpenSession(ctx); err != nil {
		a.log.Error(err)
		return err
	}

	// Step 2. Subscribes to connector system topics
	if err := a.controller.Startup(); err != nil {
		a.log.Error(err)
		return err
	}

	return nil
}

func (a *agentEngine) Stop() error {
	err := a.controller.Shutdown()
	if err != nil {
		a.log.Error(err)
		return err
	}
	return nil
}

func (a *agentEngine) logConfig(environment *models.AgentOptions) {
	log := a.log

	log.Infof("BEE ID:           %s", a.session.Bee)
	log.Infof("BEE TYPE:         %s (version: %s)", a.manifest.Name, a.manifest.Version)
	// log.Infof("BEE VERSION:      %s", )
	if a.session.IsEmbedded() {
		// log.Infof("MODE:             Standalone")
		// log.Infof("DATA SOURCE:      %s", session.ConfigPath)
	} else {
		log.Infof("MODE:             Hive bus")
		log.Infof("HIVE HOST:        %s", a.session.HiveAddress)
	}
	log.Infof("LOG LEVEL:        %s", environment.LogLevel)
	log.Infof("DATA PATH:        %s", a.store.BasePath())

	log.Info("")
	log.Info("           oO              ")
	log.Info(" Zzzz..  -{||\")   Bee agent is flying!")
	log.Info("")
	log.Info("---------------------------------------------------")
}

func (a *agentEngine) validateEnvironment() bool {
	e := a.environment
	messages := []string{}
	// Validate required parameters
	if e.Id == "" {
		messages = append(messages, "Bee ID")
	}
	// if o.Key == "" && o.Source == nil {
	// 	messages = append(messages, "Private key")
	// }
	// if o.HiveUri == "" && o.Source == nil {
	// 	messages = append(messages, "Hive address")
	// }
	// if o.Source != nil && o.Source.Path == "" && o.HiveUri == "" {
	// 	messages = append(messages, "Data source path")
	// }
	if len(messages) > 0 {
		a.log.Errorf("Missing required params: %s", strings.Join(messages, ", "))
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
