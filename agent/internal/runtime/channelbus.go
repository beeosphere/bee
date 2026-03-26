package runtime

import (
	"fmt"
	"maps"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

type channelBus struct {
	bus           models.BusClient
	Log           models.Logger
	subscriptions map[string]models.BusSubscription
	handler       models.ChannelHandler
	mapper        *TopicChannelMapper
}

func NewChannelsBus(bus models.BusClient, log models.Logger) *channelBus {
	return &channelBus{
		bus:           bus,
		Log:           log,
		subscriptions: make(map[string]models.BusSubscription),
		mapper:        &TopicChannelMapper{},
	}
}

func (b *channelBus) EmitChannel(channel string, data []byte) error {
	topic := b.mapper.GetTopic(channel)
	if topic == "" {
		return fmt.Errorf("unknown topic for channel: %s", channel)
	}
	headers := core.NewHeaders()
	headers.Set("beeos-channel", channel)
	// headers.Set("beeos-agent-id"), "")

	return b.bus.Publish(topic, data, headers)
}

func (b *channelBus) OnChannelReceived(handler models.ChannelHandler) {
	b.handler = handler
}

func (b *channelBus) hasTopicChanged(channelId string, newTopic string) bool {
	sub, exists := b.subscriptions[channelId]
	return exists && sub.Topic() != newTopic
}

func (b *channelBus) configureTopicsMap(channelConfigs []*models.ChannelConfiguration) {
	b.mapper.SetMappings(channelConfigs)
}

type channelMeta struct {
	Channel string
	Topic   string
	Pattern string
}

func (b *channelBus) processMessage(channelId string, msg *models.BusMessage) {
	// Process message
	if b.handler != nil {
		channelMsg := &models.ChannelMessage{
			Channel: channelId,
			// Topic:   msg.Topic.String(),
			Data: msg.Data.([]byte),
		}
		if err := b.handler(channelMsg); err != nil {
			b.Log.Errorf("channel bus handler error (topic: %s): %v", msg.Topic, err)
		}
	}
}

func (b *channelBus) reconcileSubscriptions(channels []*models.ChannelConfiguration) error {
	errorMsgs := []string{}

	channelsMeta := make(map[string]*channelMeta)
	for _, ch := range channels {
		if ch.Pattern == core.PatternSubscriber || ch.Pattern == core.PatternService || ch.Pattern == core.PatternConsumer {
			channelsMeta[ch.Id] = &channelMeta{
				Channel: ch.Id,
				Topic:   ch.Topic,
				Pattern: ch.Pattern,
			}
		}
	}

	// Collect channel IDs to remove and to update
	channelIdsToRemove := make([]string, 0)
	channelIdsToUpdate := make(map[string]*channelMeta)

	for channelId := range b.subscriptions {
		if meta, found := channelsMeta[channelId]; !found {
			// Channel no longer exists, mark for removal
			channelIdsToRemove = append(channelIdsToRemove, channelId)
		} else if b.hasTopicChanged(channelId, meta.Topic) {
			// Channel exists but topic has changed, mark for update
			channelIdsToUpdate[channelId] = meta
		}
	}

	// Remove unused subscribers
	for _, channelId := range channelIdsToRemove {
		sub := b.subscriptions[channelId]
		if err := sub.Unsubscribe(); err != nil {
			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
		}
		delete(b.subscriptions, channelId)
	}

	// Update subscriptions with changed topics
	for _, meta := range channelIdsToUpdate {
		channelId := meta.Channel // capture the loop variable by value (Go 1.22+)
		// Unsubscribe from old topic
		oldSub := b.subscriptions[channelId]
		if err := oldSub.Unsubscribe(); err != nil {
			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
		}

		// Subscribe to new topic
		sub, err := b.bus.Subscribe(meta.Topic, func(msg *models.BusMessage) {
			b.processMessage(channelId, msg)
		})
		if err == nil {
			b.subscriptions[channelId] = sub
		} else {
			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
			delete(b.subscriptions, channelId)         // Remove if re-subscription failed
		}
	}

	// Create new subscriptions
	newSubscriptions := make(map[string]models.BusSubscription)
	for _, meta := range channelsMeta {
		channelId := meta.Channel // capture the loop variable by value (Go 1.22+)

		if _, found := b.subscriptions[channelId]; !found {

			sub, err := b.bus.Subscribe(meta.Topic, func(msg *models.BusMessage) {
				b.processMessage(channelId, msg)
			})
			if err == nil {
				newSubscriptions[channelId] = sub
			} else {
				errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
			}
		}
	}
	// Insert new subscriptions into main map
	maps.Copy(b.subscriptions, newSubscriptions)

	if len(errorMsgs) > 0 {
		return fmt.Errorf("errors during reconcile subscriptions: %s", errorMsgs)
	}

	b.configureTopicsMap(channels)

	return nil
}

func (b *channelBus) unsubscribeAll() error {
	errorMsgs := []string{}
	for channelId, sub := range b.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
		}
		delete(b.subscriptions, channelId)
	}
	if len(errorMsgs) > 0 {
		return fmt.Errorf("errors during unsubscribe: %s", errorMsgs)
	}
	return nil
}

// TOPIC-CHANNEL MAPPER

// TODO: Debe ser utilizado también (GetChannel) desde dentro del método processorConnector en lugar de pasar el channelId al método

type TopicChannelMapper struct {
	topicMap   map[string]string
	channelMap map[string]string
}

func (m *TopicChannelMapper) SetMappings(configs []*models.ChannelConfiguration) {
	m.topicMap = make(map[string]string)
	m.channelMap = make(map[string]string)
	for _, c := range configs {
		m.topicMap[c.Topic] = c.Id
		m.channelMap[c.Id] = c.Topic
	}
}

func (m *TopicChannelMapper) GetTopic(channel string) string {
	topic, ok := m.channelMap[channel]
	if !ok {
		return ""
	}
	return topic
}

func (m *TopicChannelMapper) GetChannel(topic string) string {
	channel, ok := m.topicMap[topic]
	if !ok {
		return ""
	}
	return channel
}
