package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

type Channel interface {
	Start(proxy ChannelProxy) error
	Configure(data *ChannelInfo) error
	Stop(destroy bool) error
}

type ChannelProvider func(channelType string, logger Logger) Channel

type Parameters map[string]string

// DEPRECATED
type ChannelReceiver interface {
	MessageReceived(msg *DataMessage)
	CommandReceived(cmd *CommandMessage)
}

type ChannelBase struct {
}

// DEPRECATED
func UseChannelMessagingReceiver(proxy ChannelProxy, receiver ChannelReceiver) {
	go func() {
		for {
			select {
			case cmd := <-proxy.CommandReceiver():
				receiver.CommandReceived(cmd)
			case msg := <-proxy.DataReceiver():
				receiver.MessageReceived(msg)
			}
		}
	}()
}

func DeserializeConfig[TSettings any, TMapping any](data *ChannelInfo) (*TSettings, map[string]*TMapping, error) {
	settings, err := DeserializeSettings[TSettings](data)
	if err != nil {
		return nil, nil, err
	}
	mappings, err := DeserializeMappings[TMapping](data)
	if err != nil {
		return nil, nil, err
	}
	return settings, mappings, nil
}

func DeserializeSettings[TSettings any](data *ChannelInfo) (*TSettings, error) {
	var settings TSettings
	if err := json.Unmarshal(data.Settings, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func DeserializeMappings[TMapping any](data *ChannelInfo) (map[string]*TMapping, error) {
	mappings := make(map[string]*TMapping)
	for mappingId, mappingData := range data.Mappings {
		var mapping TMapping
		if err := json.Unmarshal(mappingData, &mapping); err != nil {
			return nil, err
		}
		mappings[mappingId] = &mapping
	}
	return mappings, nil
}

type ChannelInfo struct {
	Settings  []byte
	Resources map[string][]byte
	Mappings  map[string][]byte
	Patterns  map[string]string
	Topics    map[string]string
}

// PatternPublisher  = "pub"
// PatternSubscriber = "sub"
// PatternService    = "srv"
// PatternClient     = "cli"
// PatternProducer   = "prod"
// PatternConsumer   = "cons"

func (c *ChannelInfo) IsEmitter(routeId string) bool {
	pattern := c.Patterns[routeId]
	return pattern == PatternPublisher || pattern == PatternProducer || pattern == PatternClient
}

func (c *ChannelInfo) IsReceiver(routeId string) bool {
	pattern := c.Patterns[routeId]
	return pattern == PatternSubscriber || pattern == PatternConsumer || pattern == PatternService
}

type ChannelProxy interface {
	IsEmbedded() bool
	BeeId() string
	Metadata() *ChannelMetadata
	Config() *ChannelInfo
	DataReceiver() chan *DataMessage
	CommandReceiver() chan *CommandMessage
	Request(route string, params Parameters, data []byte) ([]byte, error)
	Publish(route string, params Parameters, data interface{}) error
	Command(op Command, params Parameters, data []byte) ([]byte, error)
	Resource(id string) ([]byte, error)
	Resources() map[string][]byte
	SharedMemory() SharedMemory
	RegisterDataReceiver(func(*DataMessage) error)
	RegisterCommandReceiver(func(*CommandMessage) error)
}

type channelProxy struct {
	channelConfig *ChannelConfiguration
	resources     map[string][]byte
	controller    *controller
	message       chan *DataMessage
	command       chan *CommandMessage
	subscribers   map[string]Subscriber
	publisher     Publisher
	// producer           *producer
	channelConfigCache *ChannelInfo
	templates          map[string]*template.Template
	dataReceiver       func(*DataMessage) error
	commandReceiver    func(*CommandMessage) error
}

type routeMeta struct {
	Route   string
	Topic   string
	Pattern string
}

func newChannelProxy(controller *controller) *channelProxy {
	// session := controller.session

	publisher := newEmbeddedPublisher()
	// publisher := newDataPublisher()
	// producer, _ := newProducer(session.hub) // TODO: Check error
	return &channelProxy{
		// channel:     channel,
		resources:   make(map[string][]byte),
		controller:  controller,
		message:     make(chan *DataMessage),
		command:     make(chan *CommandMessage),
		subscribers: make(map[string]Subscriber),
		publisher:   publisher,
		// producer:    producer,
	}
}

func (c *channelProxy) configureSubscriptions(channelConfig *ChannelConfiguration) error {

	c.channelConfig = channelConfig // TODO: Store at the beginning or at the end of method?
	c.channelConfigCache = nil

	// TEMPLATES

	c.templates = make(map[string]*template.Template)
	for _, r := range c.channelConfig.Routes {
		rid := r.Metadata.Id
		if strings.Contains(r.Topic, "{{") {

			template, err := template.New(rid).Parse(r.Topic)
			if err != nil {
				return err
			}
			c.templates[rid] = template
		}
	}

	// SUBSCRIBERS

	routesMeta := make(map[string]*routeMeta)
	// topics := []string{}
	// routes := []string{}
	for _, r := range c.channelConfig.Routes {
		if r.Pattern == PatternSubscriber || r.Pattern == PatternService || r.Pattern == PatternConsumer {
			// topics = append(topics, r.Topic)
			// routes = append(routes, r.Metadata.Id)
			routesMeta[r.Metadata.Id] = &routeMeta{
				Route:   r.Metadata.Id,
				Topic:   r.Topic,
				Pattern: r.Pattern,
			}
		}
	}

	errorMsgs := []string{}

	// Remove unused subscribers
	for _, sub := range c.subscribers {
		if _, found := routesMeta[sub.Route()]; !found {
			if err := sub.Unsubscribe(); err != nil {
				errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
			}
			delete(c.subscribers, sub.Route())
		}
	}
	// Create and insert new added subscribers
	for routeId, meta := range routesMeta {
		if _, found := c.subscribers[routeId]; !found {

			var sub Subscriber
			var err error = nil
			isEmbedded := c.controller.session.IsEmbedded()

			// Select a subscriber type based on the pattern
			switch meta.Pattern {

			case PatternSubscriber:
				if isEmbedded {
					sub = newEmbeddedSubscriber(meta.Topic, routeId, c)
				} else {
					sub = newDataSubscriber(meta.Topic, routeId, c)
				}

			case PatternService:
				sub = newDataSubscriber(meta.Topic, routeId, c)

			case PatternConsumer:
				session := c.controller.session
				sub, err = newStreamConsumer(meta.Topic, routeId, session.hub, session.bee, c)
			}
			if err == nil { // Subscribe if no error and store the subscriber in the map
				err = sub.Subscribe()
				if err == nil {
					c.subscribers[routeId] = sub
				}
			}
			if err != nil { // If error then store the error message
				errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
			}
		}
	}
	if len(errorMsgs) > 0 {
		fmt.Println("ERRORS:", errorMsgs)
		return errors.New("managing subscribers error") // TODO: Specific error wrapping all the error messages
	}
	return nil

	// // Remove unused subscribers
	// for _, sub := range c.subscribers {
	// 	if !contains(topics, sub.Topic()) {
	// 		if err := sub.Unsubscribe(); err != nil {
	// 			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
	// 		}
	// 		delete(c.subscribers, sub.Topic())
	// 	}
	// }
	// // Create and insert new added subscribers
	// for topicIdx, topic := range topics {
	// 	if _, found := c.subscribers[topic]; !found {
	// 		sub, err := newDataSubscriber(topic, routes[topicIdx], c)
	// 		if err != nil {
	// 			errorMsgs = append(errorMsgs, err.Error()) // TODO: Store current error in a formatted way...
	// 		}
	// 		c.subscribers[topic] = sub
	// 	}
	// }
	// if len(errorMsgs) > 0 {
	// 	fmt.Println("ERRORS:", errorMsgs)
	// 	return errors.New("managing subscribers error") // TODO: Specific error wrapping all the error messages
	// }
	// return nil
}

func (c *channelProxy) storeResources(resources map[string][]byte) {
	c.resources = resources
}

func (c *channelProxy) processMessage(msg *DataMessage) error {

	// TODO: Handles streaming, then delegates to connector channel implementation using DataReceiver golang channel

	// Fill in extra information to the message (pattern)
	msg.Pattern = c.Config().Patterns[msg.Route]

	c.message <- msg

	return nil
}
func (c *channelProxy) executeCommand(cmd *CommandMessage) error {

	// TODO: If command is executed then command is not sent to specific channel implementation from connector
	c.command <- cmd

	return nil
}

// CHANNEL PROXY INTERFACE IMPLEMENTATION

func (c *channelProxy) IsEmbedded() bool                      { return c.controller.session.IsEmbedded() }
func (c *channelProxy) BeeId() string                         { return c.controller.session.bee }
func (c *channelProxy) DataReceiver() chan *DataMessage       { return c.message }
func (c *channelProxy) CommandReceiver() chan *CommandMessage { return c.command }
func (c *channelProxy) Metadata() *ChannelMetadata            { return c.channelConfig.Metadata }
func (c *channelProxy) Config() *ChannelInfo {
	if c.channelConfigCache == nil {
		c.channelConfigCache = &ChannelInfo{
			Settings:  c.channelConfig.Settings,
			Resources: c.resources,
			Mappings:  make(map[string][]byte),
			Patterns:  make(map[string]string),
			Topics:    make(map[string]string),
		}
		for _, route := range c.channelConfig.Routes {
			c.channelConfigCache.Mappings[route.Metadata.Id] = route.Mapping
			c.channelConfigCache.Patterns[route.Metadata.Id] = route.Pattern
			c.channelConfigCache.Topics[route.Metadata.Id] = route.Topic
		}
	}
	return c.channelConfigCache
}
func (c *channelProxy) Request(route string, params Parameters, data []byte) ([]byte, error) {

	// TODO: Not implemented
	// topic := getTopic(route, c.channel.Routes)
	// if topic != "" {
	// 	return c.publisher.Request(topic, data)
	// }
	return nil, nil
}
func (c *channelProxy) Publish(route string, params Parameters, data interface{}) error {
	pattern, ok := c.channelConfigCache.Patterns[route]
	if !ok {
		return errors.New("route not found")
	}

	var topic string

	if pattern != PatternProducer {

		// Process topic templates using received params
		template, ok := c.templates[route]
		if ok {
			var topicBfm bytes.Buffer
			err := template.Execute(&topicBfm, params)
			if err != nil {
				return err
			}
			topic = topicBfm.String()
		} else {
			// If topic is not templated...
			// topic = getTopic(route, c.channelConfig.Routes)
			topic = c.channelConfigCache.Topics[route]
		}
		// return c.publisher.Publish(topic, data)
	} else {
		// Using streams
		topic = c.channelConfigCache.Topics[route]
		// return c.producer.Publish(topic, data)
	}

	// If embedded, publish data directly. No serialization needed
	if c.IsEmbedded() {
		return c.publisher.Publish(topic, data)
	}

	// Convert data to bytes if needed. Otherwise, use the data as bytes
	var bytes []byte
	if byteData, ok := data.([]byte); ok {
		bytes = byteData
	} else {
		var err error
		bytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	return c.publisher.Publish(topic, bytes)
}
func (c *channelProxy) Command(cmd Command, params Parameters, data []byte) ([]byte, error) {
	return nil, nil
}
func (c *channelProxy) Resource(id string) ([]byte, error) {
	if res, ok := c.resources[id]; ok {
		return res, nil
	}
	return nil, errors.New("resource not found")
}
func (c *channelProxy) Resources() map[string][]byte {
	return c.resources
}
func (c *channelProxy) SharedMemory() SharedMemory {
	return c.controller.resources
}
func (c *channelProxy) RegisterDataReceiver(handler func(*DataMessage) error) {
	c.dataReceiver = handler
}
func (c *channelProxy) RegisterCommandReceiver(handler func(*CommandMessage) error) {
	c.commandReceiver = handler
}

// // Utils
// func contains(elems []string, v string) bool {
// 	for _, s := range elems {
// 		if v == s {
// 			return true
// 		}
// 	}
// 	return false
// }

func getTopic(routeId string, routes []*ChannelRoute) string {
	var route *ChannelRoute
	for _, r := range routes {
		if r.Metadata.Id == routeId {
			route = r
			break
		}
	}
	if route != nil {
		return route.Topic
	}
	return ""
}
