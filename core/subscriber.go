package core

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

const (
	recvCommandPrefix = "$AGENT"
	recvMessagePrefix = "$BEE"
	recvCommandIndex  = 3

	topicCommandToHive = "$HIVE"
	topicCommandToBee  = "$AGENT.ID"
	topicMessage       = "$BEE"
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

//	func newCommandSubscriber(topic string, h messageHandler) (*subscriber, error) {
//		subscriber := &subscriber{
//			prefix:  recvCommandPrefix,
//			topic:   topic,
//			conn:    busClient.conn,
//			handler: h,
//		}
//		err := subscriber.subscribe()
//		return subscriber, err
//	}
func newCommandSubscriber(beeId, command string, h messageHandler) (*subscriber, error) {
	subscriber := &subscriber{
		prefix:  recvCommandPrefix,
		topic:   fmt.Sprintf("%s.%s", beeId, command),
		conn:    busClient.conn,
		handler: h,
	}
	err := subscriber.subscribe()
	return subscriber, err
}

func newDataSubscriber(topic, route string, h messageHandler) (*subscriber, error) {
	subscriber := &subscriber{
		prefix:  recvMessagePrefix,
		route:   route,
		topic:   topic,
		conn:    busClient.conn,
		handler: h,
	}
	err := subscriber.subscribe()
	return subscriber, err
}

func (s *subscriber) Topic() string {
	return s.topic // strings.TrimPrefix(s.topic, recvMessagePrefix+".")
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
	var subject string
	topic := s.topic

	if s.prefix == recvMessagePrefix {
		subject = recvMessagePrefix + "." + s.topic

	} else if s.prefix == recvCommandPrefix {
		subject = topicCommandToBee + "." + s.topic

	} else {
		subject = s.topic
	}
	subs, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {

		// log.Debugf("Message from '%s'\n", msg.Subject)

		if command := getRecvCommand(msg.Subject); command != "" {
			// BeeOS Command
			log.Tracef("Command -> bee: %s", command)

			cmd := &CommandMessage{
				Cmd:    command, // TODO: Determinar cómo se extrae la información del NATS Message... (determinar formato comandos: headers, struct...)
				Params: make(CommandParams),
				Data:   msg.Data,
			}
			s.handler.executeCommand(cmd)
		} else {
			// Data message
			msgTopic := strings.TrimPrefix(msg.Sub.Subject, recvMessagePrefix+".")
			log.Tracef("Message -> bee: %s (%d bytes) Subscribed as: %s", msgTopic, len(msg.Data), topic)

			message := &DataMessage{
				Topic: msgTopic,
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

	log.Tracef("Subscribed to (%s)\n", subject)
	return nil
}

func getRecvCommand(topic string) Command {
	// Command format: $AGENT.ID.[bee_id].[cmd]
	parts := strings.Split(topic, ".")
	if parts[0] == recvCommandPrefix && len(parts) >= recvCommandIndex+1 {
		return Command(parts[recvCommandIndex])
	}
	return ""
}
