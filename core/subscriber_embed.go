package core

import (
	// "github.com/olebedev/emitter"
	"github.com/beeosphere/bee/core/ebus"
	// log "github.com/sirupsen/logrus"
)

// type Subscriber interface {
// 	Route() string
// 	Topic() string
// 	Subscribe() error
// 	Unsubscribe() error
// }

type embeddedSubscriber struct {
	route   string
	topic   string
	handler messageHandler
	bus     ebus.Bus
	// bus     *emitter.Emitter
}

// type messageHandler interface {
// 	processMessage(msg *DataMessage) error
// 	executeCommand(cmd *CommandMessage) error
// }

func newEmbeddedSubscriber(topic, route string, h messageHandler) Subscriber {

	// e := &emitter.Emitter{}
	// e.Use("*", emitter.Void)

	return &embeddedSubscriber{
		route:   route,
		topic:   topic,
		handler: h,
		bus:     eBus,
		// bus:     e,
	}
}

func (es *embeddedSubscriber) Route() string {
	return es.route
}

func (es *embeddedSubscriber) Topic() string {
	return es.topic
}

func (es *embeddedSubscriber) Unsubscribe() error {
	// es.bus.Off("*") // unsubscribe any listeners
	es.bus.Unsubscribe(es.Topic(), es.processData)
	es.topic = ""
	return nil
}

func (es *embeddedSubscriber) Subscribe() error {
	topic := es.topic

	// Transactional determines whether subsequent callbacks for a topic are run serially (true) or concurrently(false)
	transactional := false
	es.bus.SubscribeAsync(es.Topic(), es.processData, transactional)

	// es.bus.On(topic, func(event *emitter.Event) {
	// 	msg := event.Args[0].(*DataMessage)

	// 	err := es.handler.processMessage(msg)
	// 	if err != nil {
	// 		log.Errorf("Error processing message: %s", err)
	// 	}
	// })

	log.Tracef("Subscribed (embedded) to %s\n", topic)
	return nil
}

func (es *embeddedSubscriber) processData(data interface{}) error {
	msg := &DataMessage{
		Data:     data,
		subTopic: es.Topic(),
		pubTopic: es.Topic(),
		Route:    es.Route(),
		Pattern:  "sub",
	}
	err := es.handler.processMessage(msg)
	if err != nil {
		log.Errorf("Error processing message: %s", err)
	}
	return nil
}
