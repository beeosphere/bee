package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	streamsPrefix = "$BEEOS.MSG"
)

type ConsumerState int

const (
	ConsumerPaused ConsumerState = iota
	ConsumerPlaying
)

type ConsumerPlayer interface {
	Play() error
	Pause() error
	State() ConsumerState
}

type consumer struct {
	sync.Mutex
	consumer       jetstream.Consumer
	consumeContext jetstream.ConsumeContext
	state          ConsumerState
	replayCounter  int
	topic          string
	route          string
	hub            string
	bee            string
	js             jetstream.JetStream
	handler        messageHandler
}

func newStreamConsumer(topic, route, hub, bee string, h messageHandler) (Subscriber, error) {
	js, err := jetstream.New(busClient.conn)
	if err != nil {
		return nil, err
	}
	cons := &consumer{
		state:         ConsumerPaused,
		replayCounter: 0,
		topic:         topic,
		route:         route,
		hub:           hub,
		bee:           bee,
		js:            js,
		handler:       h,
	}
	return cons, nil
}

// Subscriber interface implementation

func (c *consumer) Route() string {
	return c.route
}

func (c *consumer) Topic() string {
	return c.topic
}

func (c *consumer) Subscribe() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	streamName := c.hub
	consumerName := fmt.Sprintf("%s_%s", c.bee, c.route)
	consumerDescription := fmt.Sprintf("Consumer_%s_%s", c.bee, c.route)

	// Subject format: "$BEEOS.MSG.<hub>.<topic>"
	subject := fmt.Sprintf("%s.%s.%s", streamsPrefix, c.hub, c.topic)

	consumer, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Name:              consumerName,
		Description:       consumerDescription, // TODO: Fixed: (Updatable)
		Durable:           consumerName,
		AckPolicy:         jetstream.AckNonePolicy,          // TODO: Configurable
		FilterSubjects:    []string{subject},                // TODO: Review this (Updatable)
		AckWait:           15 * time.Second,                 // TODO: Configurable (Updatable)
		DeliverPolicy:     jetstream.DeliverAllPolicy,       // TODO: Configurable
		InactiveThreshold: 2 * 24 * time.Hour,               // TODO: Configurable (Updatable)
		MaxAckPending:     1000,                             // TODO: Configurable (Updatable)
		MaxDeliver:        -1,                               // TODO: Configurable (Updatable) (-1: Redeliver until acknowledged)
		BackOff:           []time.Duration{5 * time.Second}, // TODO: Configurable (Updatable) (Linked to MaxDeliver)
		ReplayPolicy:      jetstream.ReplayInstantPolicy,    // TODO: Fixed
		MemoryStorage:     true,                             // TODO: Configurable (Updatable) (Depends on the stream config)
		Metadata: map[string]string{ // TODO: Fixed
			"bee": c.bee,
			"hub": c.hub,
		},
	})
	if err != nil {
		return err
	}

	c.setState(ConsumerPaused)

	info, err := consumer.Info(ctx)
	if err != nil {
		return err
	}
	if info.NumPending > 0 {
		c.replayCounter = int(info.NumPending)
	}
	c.consumer = consumer
	return nil
}

func (c *consumer) Unsubscribe() error {
	return nil
}

// ConsumerPlayer interface implementation

func (c *consumer) State() ConsumerState {
	c.Lock()
	defer c.Unlock()
	return c.state
}

func (c *consumer) Play() error {
	if c.State() == ConsumerPaused {

		// Consume messages from the consumer
		consumeContext, err := c.consumer.Consume(func(msg jetstream.Msg) {

			// TODO: Take into account the replay counter. Decrease it until it reaches zero when a message is consumed.
			// TODO: Call a handler to process the message...

			message := &DataMessage{
				subTopic: c.topic,
				pubTopic: msg.Subject(),
				prefix:   streamsPrefix,
				Data:     msg.Data(),
				Route:    c.route,
				// Pattern: "", // Attach info in handler (channel/controller)
			}
			c.handler.processMessage(message)

			msg.Ack()
		})
		if err != nil {
			return err
		}
		c.consumeContext = consumeContext
		c.setState(ConsumerPlaying)
	}
	return nil
}

func (c *consumer) Pause() error {
	if c.State() == ConsumerPlaying {
		if c.consumeContext != nil {
			c.consumeContext.Stop()
		}
		c.setState(ConsumerPaused)
	}
	return nil
}

// Private methods

func (c *consumer) setState(state ConsumerState) {
	c.Lock()
	defer c.Unlock()
	c.state = state
}
