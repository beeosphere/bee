<img src="assets/beeos_logo.png" alt="BeeOS" height="150" />

## Connect anything anywhere

BeeOS is a **scalable** and **flexible** IoT platform for **distributed connectivity** and **remote deployment** of workloads. It enables seamless integration of services using a diverse range of protocols and can be deployed in a variety of environments, from cloud to edge.

Designed as an end-to-end data pipeline, the platform **translates between protocols**, **collects data from local or remote** environments, **transforms** and **organizes** information, and delivers integrated **storage** and **visualization** capabilities. It brings **computation and intelligence to where the data is generated**, enabling code workloads and AI-driven processing to run even in remote or constrained environments for real-time decision-making.

Distributed **collectors and actuators can be deployed wherever they are needed** — at the edge or in the cloud — and **remotely provisioned and managed from a centralized control panel**, enabling scalable and unified operations across the entire infrastructure.

---

![BeeOS Diagram](assets/beeos_diagram.png)

📖 Full documentation at [docs.beeos.net](https://docs.beeos.net)

# Bee Agents

The bee module is the core of BeeOS agents. It connects bees to a hive and acts as the proxy between the hive and the concrete bee implementation.

Two execution models are supported:

- **Agent context** – a lightweight agent that receives lifecycle events and a model configuration. Suitable for simple integrations or services that manage their own internal logic.
- **Connector context** – a structured agent backed by the `nectar` engine, which composes a driver, collector, and distributor to handle signal acquisition and distribution.

## Installation

```bash
go get github.com/beeosphere/bee
```

## Configuration

Create a `config.toml` file to configure the agent identity and hive connection:

```toml
[agent]
id = "bee0"
key = "SUACEN57PJE3EQWYLQIY2SAYFU7O72VLFL3PQQIYJEZMW3VPT3FSB6GQLM"
hive = "localhost:8080"
log = "debug"
```

| Field  | Description                                              |
|--------|----------------------------------------------------------|
| `id`   | Unique identifier for this agent                         |
| `key`  | Authentication key used to connect to the hive           |
| `hive` | Address of the hive server                               |
| `log`  | Log level (`debug`, `info`, `warn`, `error`)             |

Then start your bee executable passing the config file:

```bash
yourbee -c config.toml
```

## Agents development

### Using agent context

Use `WithAgentProvider` when you need a simple agent that responds to lifecycle callbacks. The `AgentContext` is passed to `Started` and gives access to hive connectivity.

```go
package main

import (
	"github.com/beeosphere/bee/agent/bee"
	"github.com/beeosphere/bee/agent/models"
)

func main() {
	var Manifest = models.AgentManifest{
		Name:    "MyAgent",
		Type:    "my-agent",
		Version: "1.0.0",
	}

	bee.NewBee(
		bee.WithManifest(Manifest),
		bee.WithAgentProvider(func(logger models.Logger) models.Agent {
			return NewMyAgent(logger)
		}),
	).Buzz()
}

type MyModel struct {
	Data string `json:"data"`
}

type MyAgent struct {
	log   models.Logger
	model *MyModel
}

func NewMyAgent(logger models.Logger) models.Agent {
	return &MyAgent{log: logger}
}

func (ag *MyAgent) Started(ac *models.AgentContext) error {
	ag.log.Info("MyAgent started")
	return nil
}

func (ag *MyAgent) Configured(model *models.Model) error {
	ag.log.Infof("MyAgent configured with model ID: %s", model.Id)
	return nil
}

func (ag *MyAgent) Stopped() error {
	ag.log.Info("MyAgent stopped")
	return nil
}
```

### Using connector context

Use `WithConnectorProvider` when your agent needs to collect and distribute signals via the `nectar` engine. The engine is wired with a driver (connection management), a collector (signal acquisition), and a distributor (signal delivery).

```go
package main

import (
	"github.com/beeosphere/bee/agent/bee"
	"github.com/beeosphere/bee/agent/models"
	"github.com/beeosphere/bee/agent/nectar"
)

func main() {
	var Manifest = models.AgentManifest{
		Name:    "MyConnector",
		Type:    "my-connector",
		Version: "1.0.0",
	}

	bee.NewBee(
		bee.WithManifest(Manifest),
		bee.WithConnectorProvider(func(connectorType string, logger models.Logger) models.Connector {
			return NewMyConnector(logger)
		}),
	).Buzz()
}

type MySettings struct {
	Data string `json:"data"`
}

type MyConnector struct {
	log    models.Logger
	engine *nectar.Engine
}

func NewMyConnector(logger models.Logger) models.Connector {
	return &MyConnector{log: logger}
}

func (c *MyConnector) Started(cc *models.ConnectorContext) error {
	c.log.Info("MyConnector started")

	c.engine = nectar.NewEngine(cc,
		nectar.WithDriver(NewMyDriverBuilder()),
		nectar.WithCollector(NewMyCollectorBuilder()),
		nectar.WithDistributor(NewMyDistributorBuilder()))
	return nil
}

func (c *MyConnector) Configured(model *models.ConnectorSetup) error {
	if err := c.engine.Run(model.Config); err != nil {
		return err
	}
	c.log.Infof("MyConnector configured with model ID: %s", model.Config.Id)
	return nil
}

func (c *MyConnector) Stopped() error {
	if err := c.engine.Stop(); err != nil {
		return err
	}
	c.log.Info("MyConnector stopped")
	return nil
}
```

#### Driver

The driver manages the physical or logical connection to the data source.

```go
type MyDriver struct {
	connected bool
	settings  *MySettings
}

func (d *MyDriver) Connect() error    { d.connected = true; return nil }
func (d *MyDriver) Disconnect() error { d.connected = false; return nil }
func (d *MyDriver) IsConnected() bool { return d.connected }

func NewMyDriverBuilder() nectar.DriverBuilder {
	return func(settings any) (nectar.Driver, error) {
		s, err := nectar.DeserializeSettings[MySettings](settings)
		if err != nil {
			return nil, errors.New("invalid settings for MyDriver")
		}
		return &MyDriver{settings: s}, nil
	}
}
```

#### Collector

The collector samples signals at a configurable interval and forwards them to the nectar engine.

```go
type MyCollector struct {
	nectar.CollectorBase
}

func NewMyCollectorBuilder() nectar.CollectorBuilder {
	return func(channel *nectar.Channel) (nectar.Collector, error) {
		return &MyCollector{
			CollectorBase: nectar.NewCollectorBase(channel),
		}, nil
	}
}

func (c *MyCollector) OnCollectSignals() <-chan *nectar.CollectorData {
	go func() {
		for {
			msg := &nectar.NectarMessage{
				Labels:  nectar.Labels{"device": "d1", "sensor": "s1"},
				Signals: []nectar.Signal{ /* populate signals */ },
			}
			c.Collected <- &nectar.CollectorData{
				Channel: c.ChannelId,
				Message: msg,
			}
			time.Sleep(time.Second)
		}
	}()
	return c.Collected
}
```

#### Distributor

The distributor receives collected signals and delivers them to their destination.

```go
type MyDistributor struct {
	nectar.DistributorBase
}

func NewMyDistributorBuilder() nectar.DistributorBuilder {
	return func(channel *nectar.Channel) (nectar.Distributor, error) {
		return &MyDistributor{
			DistributorBase: nectar.NewDistributorBase(channel),
		}, nil
	}
}

func (d *MyDistributor) Initialize() error { return nil }

func (d *MyDistributor) DistributeSignals(msg *nectar.NectarMessage) error {
	// Publish to a broker, write to a database, forward to an external service, etc.
	return nil
}
```

## License

This project is licensed under the [MIT License](LICENSE).
