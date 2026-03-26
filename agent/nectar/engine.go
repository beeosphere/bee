package nectar

import (
	"errors"
	"time"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

type CollectorBase struct {
	Collected   chan *CollectorData
	ChannelId   string
	ChannelSpec *NectarSpec
}

func NewCollectorBase(channel *Channel) CollectorBase {
	return CollectorBase{
		ChannelId:   channel.Id,
		ChannelSpec: channel.Spec,
	}
}

func (c CollectorBase) Initialize() error {
	if c.Collected != nil {
		return errors.New("collector is already initialized")
	}
	c.Collected = make(chan *CollectorData)
	return nil
}

func (c CollectorBase) Finalize() error {
	if c.Collected != nil {
		close(c.Collected)
	}
	return nil
}

type DistributorBase struct {
	ChannelId   string
	ChannelSpec *NectarSpec
}

func NewDistributorBase(channel *Channel) DistributorBase {
	return DistributorBase{
		ChannelId:   channel.Id,
		ChannelSpec: channel.Spec,
	}
}

func (c DistributorBase) Initialize() error {
	return nil
}

func (c DistributorBase) Finalize() error {
	return nil
}

func DeserializeSettings[T any](settings any) (*T, error) {
	var model T
	if err := core.Deserialize(settings.([]byte), &model); err != nil {
		return nil, err
	}
	return &model, nil
}

type NectarOption func(*Engine) error

type Initializer interface {
	Initialize() error
}

type Finalizer interface {
	Finalize() error
}

type CollectorData struct {
	Channel string
	Message *NectarMessage
}

type Collector interface {
	Initializer
	Finalizer

	OnCollectSignals() <-chan *CollectorData
}

type Distributor interface {
	Initializer
	Finalizer

	DistributeSignals(msg *NectarMessage) error
}

type Driver interface {
	IsConnected() bool
	Connect() error
	Disconnect() error
}

type Channel struct {
	Id   string
	Spec *NectarSpec
}

type CollectorBuilder func(channel *Channel) (Collector, error)
type DistributorBuilder func(channel *Channel) (Distributor, error)
type DriverBuilder func(settings any) (Driver, error)

type EngineBuilders struct {
	DriverBuilder      DriverBuilder
	CollectorBuilder   CollectorBuilder
	DistributorBuilder DistributorBuilder
}

func WithCollector(builder CollectorBuilder) NectarOption {
	return func(e *Engine) error {
		e.builders.CollectorBuilder = builder
		return nil
	}
}

func WithDistributor(builder DistributorBuilder) NectarOption {
	return func(e *Engine) error {
		e.builders.DistributorBuilder = builder
		return nil
	}
}

func WithDriver(builder DriverBuilder) NectarOption {
	return func(e *Engine) error {
		e.builders.DriverBuilder = builder
		return nil
	}
}

type EngineConfiguration struct {
}

type Engine struct {
	context      *NectarContext
	bus          NectarBus
	builders     EngineBuilders
	driver       Driver
	collectors   []Collector
	distributors map[string]Distributor
}

func NewEngine[TContext NectarContext | models.ConnectorContext](c *TContext, options ...NectarOption) *Engine {
	if nc, ok := any(c).(*NectarContext); ok {
		return createEngine(nc, options...)
	} else {
		nc := ToNectarContext(any(c).(*models.ConnectorContext))
		return createEngine(nc, options...)
	}
}

func createEngine(nc *NectarContext, options ...NectarOption) *Engine {
	bus := nc.Messaging

	e := &Engine{
		context:      nc,
		bus:          bus,
		builders:     EngineBuilders{},
		collectors:   []Collector{},
		distributors: make(map[string]Distributor),
	}
	// Apply options to the engine configuration
	for _, op := range options {
		op(e)
	}
	// If a distributor provider is defined, subscribe to signals received on the bus and route them to the appropriate distributor
	if e.builders.DistributorBuilder != nil {
		bus.OnSignalsReceived(e.signalsReceived)
	}

	return e
}

func (e *Engine) Run(config *models.ConnectorConfiguration) error {
	// Ensure any existing resources are cleaned up before starting the engine
	err := e.Stop()
	if err != nil {
		return err
	}
	// Create a new driver instance with the provided settings and connect it
	driver, err := e.builders.DriverBuilder(config.Settings)
	if err != nil {
		return err
	}
	if err := driver.Connect(); err != nil {
		return err
	}

	e.driver = driver
	e.collectors = []Collector{}
	e.distributors = make(map[string]Distributor)

	// Create collectors and distributors based on the channel configurations
	for _, chn := range config.Channels {

		// deserialize the channel spec into a NectarSpec struct
		var nectarSpec NectarSpec
		if err := core.Deserialize(chn.Spec, &nectarSpec); err != nil {
			return err
		}
		channel := &Channel{
			Id:   chn.Id,
			Spec: &nectarSpec,
		}

		if core.IsEmitter(chn.Pattern) {
			// If it's an emitter channel, we create a collector if the CollectorProvider is defined
			if e.builders.CollectorBuilder == nil {
				continue
			}
			collector, err := e.builders.CollectorBuilder(channel)
			if err != nil {
				return err
			}
			e.collectors = append(e.collectors, collector)

		} else {
			// If it's a receiver channel, we create a distributor if the DistributorProvider is defined
			if e.builders.DistributorBuilder == nil {
				continue
			}
			distributor, err := e.builders.DistributorBuilder(channel)
			if err != nil {
				return err
			}
			e.distributors[chn.Id] = distributor
		}
	}
	return e.initializeCollectorsAndDistributors()
}

func (e *Engine) Stop() error {
	// Finalize all collectors and distributors
	if err := e.finalizeCollectorsAndDistributors(); err != nil {
		return err
	}
	// Disconnect the driver if it's connected
	if e.driver != nil && e.driver.IsConnected() {
		if err := e.driver.Disconnect(); err != nil {
			return err
		}
	}
	// Delay to ensure all resources are properly released before the engine is stopped
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (e *Engine) signalsReceived(channel string, msg *NectarMessage) error {
	// Check if there is a distributor for the received channel and distribute the signals
	if distributor, exists := e.distributors[channel]; exists {
		return distributor.DistributeSignals(msg)
	}
	// If no distributor is found for the channel, log a warning and ignore the message
	e.context.Log.Warnf("No distributor found for channel %s. Ignoring received signals.", channel)

	return nil
}

func (e *Engine) processCollectedSignals(c Collector) {
	// Continuously listen for collected data from the collector and emit signals to the bus until the collector chan is closed
	for data := range c.OnCollectSignals() {

		if err := e.bus.EmitSignals(data.Channel, data.Message); err != nil {
			// Log the error and continue processing the next data
			e.context.Log.Errorf("Failed to emit signals for channel %s: %v", data.Channel, err)
			continue
		}
	}
}

func (e *Engine) initializeCollectorsAndDistributors() error {
	// Initialize all distributors
	for _, d := range e.distributors {
		if err := d.Initialize(); err != nil {
			return err
		}
	}
	// Initialize all collectors
	for _, c := range e.collectors {
		if err := c.Initialize(); err != nil {
			return err
		}
		// Start a goroutine to process collected data from the collector
		go e.processCollectedSignals(c)
	}
	return nil
}

func (e *Engine) finalizeCollectorsAndDistributors() error {
	// Finalize all distributors
	for _, d := range e.distributors {
		if err := d.Finalize(); err != nil {
			return err
		}
	}
	// Finalize all collectors
	for _, c := range e.collectors {
		if err := c.Finalize(); err != nil {
			return err
		}
	}
	return nil
}
