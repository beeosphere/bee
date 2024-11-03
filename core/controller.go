package core

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// type ProtoTest struct {
// 	Data string `json:"data"`
// }

type Controller interface {
}

type controller struct {
	logger          Logger
	session         *session
	errors          *errorTracker
	channelProvider ChannelProvider
	channels        map[string]Channel
	channelProxies  map[string]*channelProxy
	subscribers     map[string]*subscriber
	publisher       *publisher
	commands        map[string]func()
	deployer        *deployer
	cancel          context.CancelFunc
	resources       *resourceManager
	// synchronizer    *synchronizer
}

func newController(session *session, httpClient *HttpClient, channelProvider ChannelProvider) *controller {
	return &controller{
		logger:          newLogrusLogger(),
		session:         session,
		errors:          &errorTracker{},
		channels:        make(map[string]Channel),
		channelProvider: channelProvider,
		channelProxies:  make(map[string]*channelProxy),
		subscribers:     make(map[string]*subscriber),
		commands:        make(map[string]func()),
		deployer:        newDeployer(session, httpClient),
		resources:       newResourceManager(session.bee, httpClient),
		// synchronizer:    newSynchronizer(session, httpClient),
	}
}

type errorTracker struct {
}

func (en *errorTracker) enqueue(err error) {
}
func (en *errorTracker) flush() {
}

func (c *controller) startup() error {

	// busClient = newBus(c.session)

	err := busClient.Connect()
	if err != nil {
		return err
	}

	c.resources.Run()

	// c.synchronizer.Startup()
	c.deployer.Startup()

	// TODO: Activate this code!!

	// config, err := c.synchronizer.Synchronize()
	// if err != nil {
	// 	return err
	// }
	// if err = c.manageChannels(config); err != nil {
	// 	return err
	// }

	// Instantiates publisher
	c.publisher = newCommandPublisher()

	// // Subscribe to commands
	// if err := c.subscribeToCommand(SyncTopic(c.session.bee)); err != nil {
	// 	return err
	// }

	// cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	go c.deployTracker(ctx)

	c.deployer.Deploy(ctx)

	// TEST protocol
	// proto := newProtocol(c.session.bee)
	// proto.On("test", func(p *protocol, data interface{}) {
	// 	fmt.Println("Received test signal: ", data)
	// })

	// proto.Emit("$HIVE.TEST", &ProtoTest{Data: "Hello!"}, nil)

	// go func() {
	// 	<-time.After(18 * time.Second)
	// 	proto.AbortEmit("$HIVE.TEST")

	// 	proto.Emit("$HIVE.TEST", &ProtoTest{Data: "Hello!"}, nil)
	// }()

	return nil
}

func (c *controller) shutdown() error {
	if c.cancel != nil {
		c.cancel()
	}

	c.resources.Stop()

	for _, proxy := range c.channelProxies {
		// c.connector.RemoveChannel(proxy.Metadata().ChannelId)
		if err := c.stopAndRemoveChannel(proxy.Metadata().ChannelId, true); err != nil {
			c.errors.enqueue(fmt.Errorf("shutdown: %w", err))
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

func (c *controller) deployTracker(ctx context.Context) {
	for {
		select {
		case data := <-c.deployer.ConfigDeployed:

			if !data.HasConfig() {
				data.Config = &BeeConfiguration{Channels: []*ChannelConfiguration{}}
			}
			if err := c.manageChannels(data.ConfigId, data.ConfigHash, data.Config, data.Resources); err != nil {
				// TODO: log error...
				log.Warnf("failed to process channels: %s", err)
			}
			// fmt.Println("-- DEPLOY TRACKER DATA: ", data.Config)

		case <-ctx.Done():
			return
		}
	}
}

// func (c *controller) SyncNotification() error {
// 	// log.Info("Sync notification from hive")

// 	// TODO: Review sync...
// 	// TEST start
// 	config, err := c.synchronizer.Synchronize()
// 	if err != nil {
// 		return err
// 	}
// 	if err = c.manageChannels(config); err != nil {
// 		return err
// 	}
// 	// TEST end

// 	// for _, proxy := range c.channelProxies {
// 	// 	proxy.executeCommand(&CommandMessage{Cmd: Sync})
// 	// }
// 	return nil
// }

// // messageHandler implementation: Start

// func (c *controller) processMessage(msg *DataMessage) error {
// 	return nil
// }
// func (c *controller) executeCommand(cmd *CommandMessage) error {
// 	switch cmd.Cmd {
// 	case Sync:
// 		return c.SyncNotification()
// 	}
// 	return nil
// }

// // messageHandler implementation: End

func (c *controller) manageChannels(configId string, configHash string, config *BeeConfiguration, resources map[string][]byte) error {
	for _, channelConfig := range config.Channels {
		channelId := channelConfig.Metadata.ChannelId

		if proxy, ok := c.channelProxies[channelId]; ok {
			if channel, ok := c.channels[channelId]; ok {

				// Restart channel
				proxy.storeResources(resources)
				err := proxy.configureSubscriptions(channelConfig)
				if err == nil {
					err = channel.Configure(proxy.Config())
				}
				if err != nil {
					c.errors.enqueue(fmt.Errorf("sync: %w", err))
				} else {
					log.Infof("Reconfigured (channel: %s, config: %s, hash: %s)", channelId, configId, shortValue(configHash))
				}
			}
		} else {
			// Create, add and start new channel
			if err := c.createAndStartChannel(configId, configHash, channelConfig, resources); err != nil {
				c.errors.enqueue(fmt.Errorf("sync: %w", err))
			}
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
			if err := c.stopAndRemoveChannel(channelId, false); err != nil {
				c.errors.enqueue(fmt.Errorf("sync: %w", err))
			}
		}
	}
	return nil
}

func (c *controller) createAndStartChannel(configId string, configHash string, config *ChannelConfiguration, resources map[string][]byte) error {
	channelId := config.Metadata.ChannelId
	channelType := config.Metadata.ChannelType

	// Create and configure a new channel proxy
	channelProxy := newChannelProxy(c)
	channelProxy.storeResources(resources)
	if err := channelProxy.configureSubscriptions(config); err != nil {
		return err
	}

	channel := c.channelProvider(channelType, c.logger)
	if channel == nil {
		return fmt.Errorf("unknown channel type (%s) in channel ID '%s'", channelType, channelId)
	}

	if err := channel.Start(channelProxy); err != nil {
		// TODO: Dispose created channel proxy (channelProxy.Dispose() ?)
		fmt.Println("ERROR STARTING CHANNEL")
		return err
	}
	log.Infof("Started (channel: %s)", channelId)
	if err := channel.Configure(channelProxy.Config()); err != nil {
		return err
	}
	log.Infof("Configured (channel: %s, config: %s, hash: %s)", channelId, configId, shortValue(configHash))

	c.channels[channelId] = channel
	c.channelProxies[channelId] = channelProxy
	return nil
}

func (c *controller) stopAndRemoveChannel(channelId string, destroy bool) error {
	// channelProxy.executeCommand(&CommandMessage{Cmd: Shutdown})
	defer delete(c.channelProxies, channelId)
	defer delete(c.channels, channelId)

	err := c.channels[channelId].Stop(destroy)
	if destroy {
		log.Infof("Stopped and disposed (channel: %s)", channelId)
	} else {
		log.Infof("Stopped (channel: %s)", channelId)
	}
	return err
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

// func (c *controller) subscribeToCommand(topic string) error {
// 	subscriber, err := newCommandSubscriber(topic, c)
// 	if err != nil {
// 		return err
// 	}
// 	c.subscribers[topic] = subscriber
// 	return nil
// }
