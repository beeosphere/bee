package core

import (
	"fmt"
)

type Controller interface {
}

type controller struct {
	session         *session
	errors          *errorTracker
	channelProvider ChannelProvider
	channels        map[string]Channel
	channelProxies  map[string]*channelProxy
	subscribers     map[string]*subscriber
	publisher       *publisher
	commands        map[string]func()
	synchronizer    *synchronizer
}

func newController(session *session, httpClient *HttpClient, channelProvider ChannelProvider) *controller {
	return &controller{
		session:         session,
		errors:          &errorTracker{},
		channels:        make(map[string]Channel),
		channelProvider: channelProvider,
		channelProxies:  make(map[string]*channelProxy),
		subscribers:     make(map[string]*subscriber),
		commands:        make(map[string]func()),
		synchronizer:    newSynchronizer(session, httpClient),
	}
}

type errorTracker struct {
}

func (en *errorTracker) enqueue(err error) {
}
func (en *errorTracker) flush() {
}

func (c *controller) startup() error {

	busClient = newBus(c.session)

	err := busClient.Connect()
	if err != nil {
		return err
	}

	c.synchronizer.Startup()

	config, err := c.synchronizer.Synchronize()
	if err != nil {
		return err
	}
	if err = c.manageChannels(config); err != nil {
		return err
	}

	// Instantiates publisher
	c.publisher = newCommandPublisher()

	// Subscribe to commands
	if err := c.subscribeToCommand(SyncTopic(c.session.bee)); err != nil {
		return err
	}

	return nil
}

func (c *controller) shutdown() error {

	for _, proxy := range c.channelProxies {
		// c.connector.RemoveChannel(proxy.Metadata().ChannelId)
		if err := c.stopAndRemoveChannel(proxy.Metadata().ChannelId); err != nil {
			c.errors.enqueue(fmt.Errorf("Shutdown: %w", err))
		}
	}
	return nil

	// for _, proxy := range c.channelProxies {

	// 	cmd := &CommandMessage{Cmd: Shutdown}
	// 	if err := proxy.executeCommand(cmd); err != nil {
	// 		return err
	// 	}
	// }
	// <-time.After(1 * time.Second)

	// return nil
}

func (c *controller) SyncNotification() error {
	// log.Info("Sync notification from hive")

	// TODO: Review sync...
	// TEST start
	config, err := c.synchronizer.Synchronize()
	if err != nil {
		return err
	}
	if err = c.manageChannels(config); err != nil {
		return err
	}
	// TEST end

	// for _, proxy := range c.channelProxies {
	// 	proxy.executeCommand(&CommandMessage{Cmd: Sync})
	// }
	return nil
}

// messageHandler implementation: Start

func (c *controller) processMessage(msg *DataMessage) error {
	return nil
}
func (c *controller) executeCommand(cmd *CommandMessage) error {
	switch cmd.Cmd {
	case Sync:
		return c.SyncNotification()
	}
	return nil
}

// messageHandler implementation: End

func (c *controller) manageChannels(config *BeeConfiguration) error {
	for _, channelConfig := range config.Channels {
		channelId := channelConfig.Metadata.ChannelId

		if proxy, ok := c.channelProxies[channelId]; ok {
			if channel, ok := c.channels[channelId]; ok {

				// Restart channel
				err := proxy.configureSubscriptions(channelConfig)
				if err == nil {
					err = channel.Restart()
				}
				if err != nil {
					c.errors.enqueue(fmt.Errorf("Sync: %w", err))
				}
			}
		} else {
			// Create, add and start new channel
			if err := c.createAndStartChannel(channelConfig); err != nil {
				c.errors.enqueue(fmt.Errorf("Sync: %w", err))
			}
		}
		// Stop and remove unused channels
		for channelId := range c.channelProxies {
			found := false
			for _, channelCfg := range config.Channels {
				if channelCfg.Metadata.ChannelId == channelId {
					found = true
					break
				}
			}
			if !found {
				if err := c.stopAndRemoveChannel(channelId); err != nil {
					c.errors.enqueue(fmt.Errorf("Sync: %w", err))
				}
			}
		}
	}
	return nil
}

// func (c *controller) organizeChannels(config *BeeConfiguration) error {
// 	for _, channelConfig := range config.Channels {
// 		var err error
// 		channelId := channelConfig.Metadata.ChannelId

// 		channelProxy := newChannelProxy(channelConfig, c)
// 		if _, ok := c.channelProxies[channelId]; ok {

// 			err = channelProxy.configureChannel(channelConfig)
// 		} else {
// 			err = c.connector.CreateChannel(channelProxy)
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		c.channelProxies[channelId] = channelProxy
// 	}
// 	return nil
// }

func (c *controller) subscribeToCommand(topic string) error {
	subscriber, err := newCommandSubscriber(topic, c)
	if err != nil {
		return err
	}
	c.subscribers[topic] = subscriber
	return nil
}

func (c *controller) createAndStartChannel(config *ChannelConfiguration) error {
	channelId := config.Metadata.ChannelId
	channelType := config.Metadata.ChannelType

	// Create and configure a new channel proxy
	channelProxy := newChannelProxy(c)
	if err := channelProxy.configureSubscriptions(config); err != nil {
		return err
	}
	channel := c.channelProvider(channelType)
	if channel == nil {
		return fmt.Errorf("Unknown channel type (%s) in channel ID '%s'", channelType, channelId)
	}
	if err := channel.Start(channelProxy); err != nil {
		// TODO: Dispose created channel proxy (channelProxy.Dispose() ?)
		return err
	}
	c.channels[channelId] = channel
	c.channelProxies[channelId] = channelProxy
	return nil
}

func (c *controller) stopAndRemoveChannel(channelId string) error {
	// channelProxy.executeCommand(&CommandMessage{Cmd: Shutdown})
	defer delete(c.channelProxies, channelId)
	defer delete(c.channels, channelId)

	return c.channels[channelId].Stop()
}
