package models

type BusClient interface {
	Publish(topic string, data any, headers BusHeaders) error
	Subscribe(topic string, handler BusHandler) (BusSubscription, error)
}

type BusHandler func(msg *BusMessage)

type BusSubscription interface {
	Unsubscribe() error
	Topic() string
}

type BusTopic interface {
	String() string
	Command() string
}

type BusHeaders interface {
	Get(key string) (string, bool)
	Set(key, value string)
	Delete(key string)
	Keys() []string
	ToMap() map[string]string
}

type BusMessage struct {
	Topic           BusTopic
	SubscribedTopic string
	Data            any
	Headers         BusHeaders
}
