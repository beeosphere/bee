package core

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type responder struct {
	prefix string
	conn   *nats.Conn
}

func newResponder() *responder {
	return &responder{
		conn: busClient.conn,
	}
}

func (p *responder) Respond(replyTopic string, data interface{}) error {
	var err error
	bytes, ok := data.([]byte)
	if !ok {
		bytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	return p.conn.Publish(replyTopic, bytes)
}
