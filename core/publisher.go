package core

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

const (
	sendCommandPrefix = "$HIVE"
	sendMessagePrefix = "$BEE"
	sendCommandIndex  = 1
)

type publisher struct {
	prefix string
	conn   *nats.Conn
}

func newCommandPublisher() *publisher {
	return &publisher{
		prefix: sendCommandPrefix,
		conn:   busClient.conn,
	}
}

func newDataPublisher() *publisher {
	return &publisher{
		prefix: sendMessagePrefix,
		conn:   busClient.conn,
	}
}

func (p *publisher) Publish(topic string, data interface{}) error {
	var err error
	bytes, ok := data.([]byte)
	if !ok {
		bytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	var subject string
	if p.prefix == sendMessagePrefix {
		subject = sendMessagePrefix + "." + topic
		log.Tracef("Message <- bee: %s (%d bytes)", topic, len(bytes))

	} else if p.prefix == sendCommandPrefix {
		subject = topicCommandToHive + "." + topic
		cmd := topic
		log.Tracef("Command <- bee: %s", cmd)

	} else {
		return errors.New("Invalid prefix")
	}
	return p.conn.Publish(subject, bytes)
}

func (p *publisher) Request(topic string, request interface{}, response interface{}) error { // TODO Adapt Request to work as Publish does
	var err error
	reqBytes, ok := request.([]byte)
	if !ok {
		reqBytes, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}
	msg, err := p.conn.Request(topic, reqBytes, 20*time.Second)
	if err != nil {
		return err
	}
	_, ok = response.([]byte)
	if !ok {
		err = json.Unmarshal(msg.Data, response)
		if err != nil {
			return err
		}
	} else {
		response = msg.Data
	}
	return nil

	// bytes, err := json.Marshal(request)
	// if err != nil {
	// 	return err
	// }
	// msg, err := p.conn.Request(topic, bytes, 20*time.Second)
	// if err != nil {
	// 	return err
	// }
	// err = json.Unmarshal(msg.Data, response)
	// if err != nil {
	// 	return err
	// }
	// return nil
}

// func getSendCommand(topic string) Command {
// 	// Command format: $HIVE.[cmd]
// 	parts := strings.Split(topic, ".")
// 	if parts[0] == sendCommandPrefix && len(parts) >= sendCommandIndex+1 {
// 		return Command(parts[sendCommandIndex])
// 	}
// 	return ""
// }
