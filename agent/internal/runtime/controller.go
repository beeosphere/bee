package runtime

import (
	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

type Controller struct {
	session      *core.Session
	log          models.Logger
	bus          core.Bus
	commander    *commander
	synchronizer Synchronizer
	processor    Processor
}

func NewController(session *core.Session, bus core.Bus, synchronizer Synchronizer, commander *commander, processor Processor) *Controller {
	return &Controller{
		session:      session,
		log:          session.Log,
		bus:          bus,
		commander:    commander,
		synchronizer: synchronizer,
		processor:    processor,
	}
}

func (c *Controller) Startup() error {
	// BUS CONNECT
	if err := c.bus.Connect(); err != nil {
		return err
	}
	// COMMAND SUBSCRIPTIONS
	if err := c.commander.SubscribeCommands(); err != nil {
		return err
	}
	// SYNCHRONIZER STARTUP
	c.synchronizer.OnSynchronized(c.onModelSynchronized)
	if err := c.synchronizer.Startup(); err != nil {
		return err
	}
	// START AN INITIAL SYNCHRONIZATION
	if !c.synchronizer.SynchronizeFromStore() {
		c.synchronizer.SynchronizeFromHive()
	}
	return nil
}

func (c *Controller) Shutdown() error {
	// COMMAND UNSUBSCRIPTIONS
	if err := c.commander.UnsubscribeCommands(); err != nil {
		return err
	}
	if err := c.processor.Dispose(); err != nil {
		return err
	}
	// BUS DISCONNECT
	if err := c.bus.Disconnect(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) onModelSynchronized(model *SyncedModel) {
	c.processor.Process(&models.Model{
		Id:        model.ModelId,
		Hash:      model.ModelHash,
		Data:      model.Model,
		Resources: model.Resources,
	})
}
