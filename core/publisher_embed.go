package core

import "github.com/beeosphere/bee/core/ebus"

type embeddedPublisher struct {
	bus ebus.Bus
}

func newEmbeddedPublisher() Publisher {
	return &embeddedPublisher{
		bus: eBus,
	}
}

func (p *embeddedPublisher) Publish(topic string, data interface{}) error {

	p.bus.Publish(topic, data)

	return nil
}
