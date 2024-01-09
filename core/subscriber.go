package core

import (
	"strings"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

const (
	beeosCommandPrefix        = "$AGENT"
	beeosDataPrefix           = "$BEE"
	beeosCommandPrefixWithDot = beeosCommandPrefix + "."
	beeosDataPrefixWithDot    = beeosDataPrefix + "."
	beeosCommandIndex         = 3
)

type subscriber struct {
	prefix       string
	route        string
	topic        string
	conn         *nats.Conn
	subscription *nats.Subscription
	handler      messageHandler
}

type messageHandler interface {
	processMessage(msg *DataMessage) error
	executeCommand(cmd *CommandMessage) error
}

func newCommandSubscriber(topic string, h messageHandler) (*subscriber, error) {
	subscriber := &subscriber{
		prefix:  beeosCommandPrefix,
		topic:   topic,
		conn:    busClient.conn,
		handler: h,
	}
	err := subscriber.subscribe()
	return subscriber, err
}

func newDataSubscriber(topic, route string, h messageHandler) (*subscriber, error) {
	subscriber := &subscriber{
		prefix:  beeosDataPrefix,
		route:   route,
		topic:   topic,
		conn:    busClient.conn,
		handler: h,
	}
	err := subscriber.subscribe()
	return subscriber, err
}

func (s *subscriber) Topic() string {
	return strings.TrimPrefix(s.topic, beeosDataPrefixWithDot)
}

func (s *subscriber) Unsubscribe() error {
	if err := s.subscription.Unsubscribe(); err != nil {
		return err
	}
	s.subscription = nil
	s.topic = ""
	return nil
}

func (s *subscriber) subscribe() error { // TODO: Determine queue group subscriptions... take into account...
	var topic string
	if s.prefix == beeosDataPrefix {
		topic = beeosDataPrefixWithDot + s.topic
	} else {
		topic = s.topic
	}
	subs, err := s.conn.Subscribe(topic, func(msg *nats.Msg) {

		log.Debugf("Message from '%s'\n", msg.Subject)

		if command := getCommand(msg.Subject); command != "" {
			// BeeOS Command
			cmd := &CommandMessage{
				Cmd:    command, // TODO: Determinar cómo se extrae la información del NATS Message... (determinar formato comandos: headers, struct...)
				Params: make(CommandParams),
				Data:   msg.Data,
			}
			s.handler.executeCommand(cmd)
		} else {
			// Data message
			message := &DataMessage{
				Topic: strings.TrimPrefix(msg.Subject, beeosDataPrefixWithDot),
				Data:  msg.Data,
				Route: s.route,
				// Pattern: "", // Attach info in handler (channel/controller)
			}
			s.handler.processMessage(message)
		}
	})
	if err != nil {
		return err
	}
	s.subscription = subs

	if strings.HasPrefix(topic, beeosCommandPrefix) {
		log.Debugf("Subscribed to command '%s'\n", topic)
	} else {
		log.Infof("Subscribed to '%s'\n", topic)
	}
	return nil
}

func getCommand(topic string) Command {
	// Command format: $BEEOS.BEE.[bee_id].[cmd]
	parts := strings.Split(topic, ".")
	if parts[0] == beeosCommandPrefix && len(parts) >= 4 {
		return Command(parts[beeosCommandIndex])
	}
	return ""
}

// ----------------------------------------------

// package core

// import (
// 	"strings"

// 	"github.com/nats-io/nats.go"
// 	log "github.com/sirupsen/logrus"
// )

// const (
// 	beeosCommandPrefix        = "$AGENT"
// 	beeosDataPrefix           = "$BEE"
// 	beeosCommandPrefixWithDot = beeosCommandPrefix + "."
// 	beeosDataPrefixWithDot    = beeosDataPrefix + "."
// 	beeosCommandIndex         = 3
// )

// type subscriber struct {
// 	prefix       string
// 	route        string
// 	topic        string
// 	conn         *nats.Conn
// 	subscription *nats.Subscription
// 	handler      messageHandler
// }

// type messageHandler interface {
// 	processMessage(msg *DataMessage) error
// 	executeCommand(cmd *CommandMessage) error
// }

// func newCommandSubscriber(topic string, h messageHandler) (*subscriber, error) {
// 	subscriber := &subscriber{
// 		prefix:  beeosCommandPrefix,
// 		topic:   topic,
// 		conn:    busClient.conn,
// 		handler: h,
// 	}
// 	err := subscriber.subscribe()
// 	return subscriber, err
// }

// func newDataSubscriber(topic, route string, h messageHandler) (*subscriber, error) {
// 	subscriber := &subscriber{
// 		prefix:  beeosDataPrefix,
// 		route:   route,
// 		topic:   topic,
// 		conn:    busClient.conn,
// 		handler: h,
// 	}
// 	err := subscriber.subscribe()
// 	return subscriber, err
// }

// func (s *subscriber) Topic() string {
// 	return strings.TrimPrefix(s.topic, beeosDataPrefixWithDot)
// }

// func (s *subscriber) Unsubscribe() error {
// 	if err := s.subscription.Unsubscribe(); err != nil {
// 		return err
// 	}
// 	s.subscription = nil
// 	s.topic = ""
// 	return nil
// }

// func (s *subscriber) subscribe() error { // TODO: Determine queue group subscriptions... take into account...
// 	var topic string
// 	if s.prefix == beeosDataPrefix {
// 		topic = beeosDataPrefixWithDot + s.topic
// 	} else {
// 		topic = s.topic
// 	}
// 	subs, err := s.conn.Subscribe(topic, func(msg *nats.Msg) {

// 		log.Debugf("Message from '%s'\n", msg.Subject)

// 		if command := getCommand(msg.Subject); command != "" {
// 			// BeeOS Command
// 			cmd := &CommandMessage{
// 				Cmd:    command, // TODO: Determinar cómo se extrae la información del NATS Message... (determinar formato comandos: headers, struct...)
// 				Params: make(CommandParams),
// 				Data:   msg.Data,
// 			}
// 			s.handler.executeCommand(cmd)
// 		} else {
// 			// Data message
// 			message := &DataMessage{
// 				Topic: strings.TrimPrefix(msg.Subject, beeosDataPrefixWithDot),
// 				Data:  msg.Data,
// 				Route: s.route,
// 				// Pattern: "", // Attach info in handler (channel/controller)
// 			}
// 			s.handler.processMessage(message)
// 		}
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	s.subscription = subs

// 	if strings.HasPrefix(topic, beeosCommandPrefix) {
// 		log.Debugf("Subscribed to command '%s'\n", topic)
// 	} else {
// 		log.Infof("Subscribed to '%s'\n", topic)
// 	}
// 	return nil
// }

// func getCommand(topic string) Command {
// 	// Command format: $BEEOS.BEE.[bee_id].[cmd]
// 	parts := strings.Split(topic, ".")
// 	if parts[0] == beeosCommandPrefix && len(parts) >= 4 {
// 		return Command(parts[beeosCommandIndex])
// 	}
// 	return ""
// }
