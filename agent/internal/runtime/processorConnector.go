package runtime

import (
	"fmt"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

// CONNECTOR PROCESSOR STRATEGY

type ConnectorProcessor struct {
	session           *core.Session
	log               models.Logger
	busClient         models.BusClient
	connectorProvider models.ConnectorProvider
	connectors        map[string]models.Connector
	channelBuses      map[string]*channelBus
}

func NewConnectorProcessor(session *core.Session, provider models.ConnectorProvider, busClient models.BusClient) *ConnectorProcessor {
	return &ConnectorProcessor{
		session:           session,
		log:               session.Log,
		busClient:         busClient,
		connectorProvider: provider,
		connectors:        make(map[string]models.Connector),
		channelBuses:      make(map[string]*channelBus),
	}
}

func (p *ConnectorProcessor) Process(model *models.Model) error {
	var config models.AgentConfiguration
	err := core.Deserialize(model.Data, &config)
	if err != nil {
		return nil
	}
	return p.manageConnectors(model.Id, model.Hash, &config)
}

func (p *ConnectorProcessor) Dispose() error {
	// Iterate and delete connectors
	for _, connectorId := range p.connectorIds() {
		if err := p.stopAndRemoveConnector(connectorId); err != nil {
			p.log.Errorf("Error stopping connector %s: %v", connectorId, err)
		}
	}
	return nil
}

func (p *ConnectorProcessor) connectorIds() []string {
	ids := make([]string, 0, len(p.connectors))
	for id := range p.connectors {
		ids = append(ids, id)
	}
	return ids
}

func (p *ConnectorProcessor) manageConnectors(configId string, configHash string, config *models.AgentConfiguration) error {
	for _, connectorConfig := range config.Connectors {
		connectorId := connectorConfig.Id

		if connector, ok := p.connectors[connectorId]; ok {
			if channelBus, ok := p.channelBuses[connectorId]; ok {

				// Reconfigure connector
				err := channelBus.reconcileSubscriptions(connectorConfig.Channels)
				if err != nil {
					// p.errors.enqueue(fmt.Errorf("sync: %w", err))
				}
				setup := &models.ConnectorSetup{
					Config: connectorConfig,
				}
				err = connector.Configured(setup)
				if err != nil {
					// p.errors.enqueue(fmt.Errorf("sync: %w", err))
				}
				p.log.Infof("Reconfigured (bee: %s, connector: %s, model: %s, hash: %s)", p.session.Bee, connectorId, configId, core.ShortValue(configHash))
			}
		} else {
			// Create, add and start new connector
			if err := p.createAndStartConnector(configId, configHash, connectorConfig); err != nil {
				// p.errors.enqueue(fmt.Errorf("sync: %w", err))
			}
		}
	}
	// Stop and remove unused connectors
	for _, connectorId := range p.connectorIds() {
		found := false
		for _, connectorConfig := range config.Connectors {
			if connectorConfig.Id == connectorId {
				found = true
				break
			}
		}
		if !found {
			if err := p.stopAndRemoveConnector(connectorId); err != nil {
				// p.errors.enqueue(fmt.Errorf("sync: %w", err))
			}
		}
	}
	return nil
}

func (p *ConnectorProcessor) createAndStartConnector(configId string, configHash string, config *models.ConnectorConfiguration) error {
	log := p.log

	connectorId := config.Id
	connectorType := config.ConnectorType
	chBus := NewChannelsBus(p.busClient, log)

	// Create and configure a new channel proxy
	cctx := &models.ConnectorContext{
		AgentId:    p.session.Bee,
		InstanceId: p.session.PublicKey,
		Manifest:   nil, // TODO: Provide manifest if needed
		Log:        p.log,
		Channels:   chBus,
	}

	if err := chBus.reconcileSubscriptions(config.Channels); err != nil {
		return err
	}

	connector := p.connectorProvider(connectorType, p.log)
	if connector == nil {
		return fmt.Errorf("unknown connector type (%s) in connector ID '%s' (bee: %s)", connectorType, connectorId, p.session.Bee)
	}

	if err := connector.Started(cctx); err != nil {
		log.Errorf("Start error (bee: %s, connector: %s): %v", p.session.Bee, connectorId, err)
		chBus.unsubscribeAll()
		return err
	}
	log.Infof("Started (bee: %s, connector: %s)", p.session.Bee, connectorId)

	setup := &models.ConnectorSetup{
		Config: config,
	}
	if err := connector.Configured(setup); err != nil {
		return err
	}
	log.Infof("Configured (bee: %s, connector: %s, model: %s, hash: %s)", p.session.Bee, connectorId, configId, core.ShortValue(configHash))

	p.connectors[connectorId] = connector
	p.channelBuses[connectorId] = chBus

	// TODO: We need to send a bool to a global channel when all the channels are started ???
	// c.started <- true

	return nil
}

func (p *ConnectorProcessor) stopAndRemoveConnector(connectorId string) error {
	log := p.log
	defer delete(p.connectors, connectorId)
	defer delete(p.channelBuses, connectorId)

	err := p.connectors[connectorId].Stopped()
	p.channelBuses[connectorId].unsubscribeAll()
	// TODO: Errors handling
	log.Infof("Stopped (bee: %s, connector: %s)", p.session.Bee, connectorId)
	return err
}

// // CHANNEL BUS IMPLEMENTATION

// const (
// 	PatternPublisher  = "pub"
// 	PatternSubscriber = "sub"
// 	PatternService    = "srv"
// 	PatternClient     = "cli"
// 	PatternProducer   = "prod"
// 	PatternConsumer   = "cons"
// )

// type channelBus struct {
// 	bus           models.BusClient
// 	Log           models.Logger
// 	subscriptions map[string]models.BusSubscription
// 	handler       models.ChannelHandler
// 	topicsMap     map[string]string
// }

// func NewChannelsBus(bus models.BusClient, log models.Logger) *channelBus {
// 	return &channelBus{
// 		bus:           bus,
// 		Log:           log,
// 		subscriptions: make(map[string]models.BusSubscription),
// 		topicsMap:     make(map[string]string),
// 	}
// }

// func (b *channelBus) Emit(channel string, data []byte) error {
// 	topic, ok := b.topicsMap[channel]
// 	if !ok {
// 		return fmt.Errorf("unknown topic for channel: %s", channel)
// 	}
// 	headers := core.NewHeaders()
// 	headers.Set("beeos-channel", channel)
// 	// headers.Set("beeos-agent-id"), "")

// 	return b.bus.Publish(topic, data, headers)
// }

// func (b *channelBus) OnReceived(handler models.ChannelHandler) {
// 	b.handler = handler
// }

// func (b *channelBus) hasTopicChanged(channelId string, newTopic string) bool {
// 	sub, exists := b.subscriptions[channelId]
// 	return exists && sub.Topic() != newTopic
// }

// func (b *channelBus) configureTopicsMap(channels []*models.ChannelConfiguration) {
// 	b.topicsMap = make(map[string]string)
// 	for _, ch := range channels {
// 		b.topicsMap[ch.Id] = ch.Topic
// 	}
// }

// type channelMeta struct {
// 	Channel string
// 	Topic   string
// 	Pattern string
// }

// func (b *channelBus) processMessage(channelId string, msg *models.BusMessage) {
// 	// Process message
// 	if b.handler != nil {
// 		channelMsg := &models.ChannelMessage{
// 			Channel: channelId,
// 			// Topic:   msg.Topic.String(),
// 			Data: msg.Data.([]byte),
// 		}
// 		if err := b.handler(channelMsg); err != nil {
// 			b.Log.Errorf("channel bus handler error (topic: %s): %v", msg.Topic, err)
// 		}
// 	}
// }

// func (b *channelBus) reconcileSubscriptions(channels []*models.ChannelConfiguration) error {
// 	errorMsgs := []string{}

// 	channelsMeta := make(map[string]*channelMeta)
// 	for _, ch := range channels {
// 		if ch.Pattern == PatternSubscriber || ch.Pattern == PatternService || ch.Pattern == PatternConsumer {
// 			channelsMeta[ch.Id] = &channelMeta{
// 				Channel: ch.Id,
// 				Topic:   ch.Topic,
// 				Pattern: ch.Pattern,
// 			}
// 		}
// 	}

// 	// Collect channel IDs to remove and to update
// 	channelIdsToRemove := make([]string, 0)
// 	channelIdsToUpdate := make(map[string]*channelMeta)

// 	for channelId := range b.subscriptions {
// 		if meta, found := channelsMeta[channelId]; !found {
// 			// Channel no longer exists, mark for removal
// 			channelIdsToRemove = append(channelIdsToRemove, channelId)
// 		} else if b.hasTopicChanged(channelId, meta.Topic) {
// 			// Channel exists but topic has changed, mark for update
// 			channelIdsToUpdate[channelId] = meta
// 		}
// 	}

// 	// Remove unused subscribers
// 	for _, channelId := range channelIdsToRemove {
// 		sub := b.subscriptions[channelId]
// 		if err := sub.Unsubscribe(); err != nil {
// 			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
// 		}
// 		delete(b.subscriptions, channelId)
// 	}

// 	// Update subscriptions with changed topics
// 	for channelId, meta := range channelIdsToUpdate {
// 		channelId := channelId // capture the loop variable by value (Go 1.22+)
// 		// Unsubscribe from old topic
// 		oldSub := b.subscriptions[channelId]
// 		if err := oldSub.Unsubscribe(); err != nil {
// 			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
// 		}

// 		// Subscribe to new topic
// 		sub, err := b.bus.Subscribe(meta.Topic, func(msg *models.BusMessage) {
// 			b.processMessage(channelId, msg)
// 		})
// 		if err == nil {
// 			b.subscriptions[channelId] = sub
// 		} else {
// 			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
// 			delete(b.subscriptions, channelId)         // Remove if re-subscription failed
// 		}
// 	}

// 	// Create new subscriptions
// 	newSubscriptions := make(map[string]models.BusSubscription)
// 	for channelId, meta := range channelsMeta {
// 		channelId := channelId // capture the loop variable by value (Go 1.22+)

// 		if _, found := b.subscriptions[channelId]; !found {

// 			sub, err := b.bus.Subscribe(meta.Topic, func(msg *models.BusMessage) {
// 				b.processMessage(channelId, msg)
// 			})
// 			if err == nil {
// 				newSubscriptions[channelId] = sub
// 			} else {
// 				errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
// 			}
// 		}
// 	}
// 	// Insert new subscriptions into main map
// 	maps.Copy(b.subscriptions, newSubscriptions)

// 	if len(errorMsgs) > 0 {
// 		return fmt.Errorf("errors during reconcile subscriptions: %s", errorMsgs)
// 	}

// 	b.configureTopicsMap(channels)

// 	return nil
// }

// func (b *channelBus) unsubscribeAll() error {
// 	errorMsgs := []string{}
// 	for channelId, sub := range b.subscriptions {
// 		if err := sub.Unsubscribe(); err != nil {
// 			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
// 		}
// 		delete(b.subscriptions, channelId)
// 	}
// 	if len(errorMsgs) > 0 {
// 		return fmt.Errorf("errors during unsubscribe: %s", errorMsgs)
// 	}
// 	return nil
// }

// type TopicMapper struct {
// 	topicToChannel map[string]string
// 	channelToTopic map[string]string
// }

// func NewTopicMapper(config *models.AgentConfiguration) *TopicMapper {
// 	mapper := &TopicMapper{
// 		topicToChannel: make(map[string]string),
// 		channelToTopic: make(map[string]string),
// 	}
// 	for _, connector := range config.Connectors {
// 		for _, ch := range connector.Channels {
// 			mapper.topicToChannel[ch.Topic] = ch.Id
// 			mapper.channelToTopic[ch.Id] = ch.Topic
// 		}
// 	}
// 	return mapper
// }

// func (tm *TopicMapper) CleanCache() {
// 	tm.topicToChannel = make(map[string]string)
// 	tm.channelToTopic = make(map[string]string)
// }

// func (tm *TopicMapper) GetChannelForTopic(topic string) (string, bool) {
// 	channelId, ok := tm.topicToChannel[topic]
// 	return channelId, ok
// }

// func (tm *TopicMapper) GetTopicForChannel(channelId string) (string, bool) {
// 	topic, ok := tm.channelToTopic[channelId]
// 	return topic, ok
// }
