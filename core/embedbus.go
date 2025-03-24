package core

// import (
// 	"errors"
// 	"sync"
// )

// type EmbedBus interface {
// 	Subscribe(topic string, handler func(data interface{})) error
// 	Publish(topic string, data interface{}) error
// 	Unsubscribe(topic string) error
// }

// type embedBus struct {
// 	handlers map[string][]func(data interface{})
// 	lock     sync.RWMutex
// }

// func NewEmbedBus() *embedBus {
// 	return &embedBus{
// 		handlers: make(map[string][]func(data interface{})),
// 	}
// }

// func (eb *embedBus) Subscribe(topic string, handler func(data interface{})) error {
// 	eb.lock.Lock()
// 	defer eb.lock.Unlock()

// 	if _, ok := eb.handlers[topic]; !ok {
// 		eb.handlers[topic] = []func(data interface{}){}
// 	}
// 	eb.handlers[topic] = append(eb.handlers[topic], handler)
// 	return nil
// }

// func (eb *embedBus) Publish(topic string, data interface{}) error {
// 	eb.lock.RLock()
// 	defer eb.lock.RUnlock()

// 	if handlers, ok := eb.handlers[topic]; ok {
// 		for _, handler := range handlers {
// 			go handler(data)
// 		}
// 		return nil
// 	}
// 	return errors.New("no handlers for topic")
// }

// func (eb *embedBus) Unsubscribe(topic string) error {
// 	eb.lock.Lock()
// 	defer eb.lock.Unlock()

// 	if _, ok := eb.handlers[topic]; ok {
// 		delete(eb.handlers, topic)
// 		return nil
// 	}
// 	return errors.New("no handlers for topic")
// }
