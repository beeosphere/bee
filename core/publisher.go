package core

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type publisher struct {
	prefix string
	conn   *nats.Conn
}

func newCommandPublisher() *publisher {
	return &publisher{
		prefix: beeosCommandPrefix,
		conn:   busClient.conn,
	}
}

func newDataPublisher() *publisher {
	return &publisher{
		prefix: beeosDataPrefix,
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
	if p.prefix == beeosDataPrefix {
		topic = beeosDataPrefix + topic
	}

	log.Debugf("Publish to '%s'\n", topic)

	return p.conn.Publish(topic, bytes)
}

func (p *publisher) Request(topic string, request interface{}, response interface{}) error {
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
