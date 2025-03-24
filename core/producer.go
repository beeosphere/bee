package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	// log "github.com/sirupsen/logrus"
)

type producer struct {
	js  jetstream.Publisher
	hub string
}

func newProducer(hub string) (*producer, error) {
	js, err := jetstream.New(busClient.conn)
	if err != nil {
		return nil, err
	}
	return &producer{
		js:  js,
		hub: hub,
	}, nil
}

func (p *producer) Publish(topic string, data interface{}) error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	bytes, ok := data.([]byte)
	if !ok {
		bytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	// Subject format: "$BEEOS.MSG.<hub>.<topic>"
	subject := fmt.Sprintf("%s.%s.%s", streamsPrefix, p.hub, topic)

	log.Tracef("Message <- bee: %s (%d bytes)", topic, len(bytes))

	_, err = p.js.Publish(ctx, subject, bytes)
	return err
}
