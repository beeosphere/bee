package core

import (
	"context"

	"github.com/beeosphere/bee/agent/internal/core/topics"
	"github.com/beeosphere/bee/agent/models"
	"github.com/nats-io/nats.go"
)

// PATTERN CONSTANTS

const (
	PatternPublisher  = "pub"
	PatternSubscriber = "sub"
	PatternService    = "srv"
	PatternClient     = "cli"
	PatternProducer   = "prod"
	PatternConsumer   = "cons"
)

func IsEmitter(pattern string) bool {
	return pattern == PatternPublisher || pattern == PatternService || pattern == PatternProducer
}

// BUS

type BusController interface {
	Connect(ctx context.Context) error
	Disconnect() error
}

type Bus interface {
	models.BusClient
	BusController
}

// SUBSCRIPTIONS

type busSubscription struct {
	subscription *nats.Subscription
}

func NewSubscription(sub *nats.Subscription) models.BusSubscription {
	return &busSubscription{
		subscription: sub,
	}
}

func (bs *busSubscription) Unsubscribe() error {
	return bs.subscription.Unsubscribe()
}

func (bs *busSubscription) Topic() string {
	return bs.subscription.Subject
}

// TOPICS

type busTopic string

func (bt busTopic) String() string {
	return string(bt)
}

func (bt busTopic) Command() string {
	return topics.FetchCommand(string(bt))
}

func NewTopic(topic string) models.BusTopic {
	return busTopic(topic)
}

// HEADERS

type busHeaders map[string]string

func NewHeaders() busHeaders {
	return make(busHeaders)
}

func (h busHeaders) Get(key string) (string, bool) {
	value, ok := h[key]
	return value, ok
}

func (h busHeaders) Set(key, value string) {
	h[key] = value
}

func (h busHeaders) Delete(key string) {
	delete(h, key)
}

func (h busHeaders) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

func (h busHeaders) ToMap() map[string]string {
	return h
}
