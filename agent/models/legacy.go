package models

// import (
// 	"encoding/json"
// 	"strconv"
// 	"strings"
// )

// // BEE CONFIGURATION
// type BeeConfiguration struct {
// 	AgentType string                  `json:"agentType"` // Hive or Bee
// 	Channels  []*ChannelConfiguration `json:"channels"`
// }
// type ChannelConfiguration struct {
// 	Metadata *ChannelMetadata `json:"metadata"`
// 	Settings json.RawMessage  `json:"settings,omitempty"`
// 	Routes   []*ChannelRoute  `json:"routes"`
// 	Streams  []*ChannelStream `json:"streams,omitempty"`
// }
// type ChannelMetadata struct {
// 	ChannelId     string `json:"channelId"`
// 	ChannelName   string `json:"channelName"`
// 	ChannelType   string `json:"channelType"`
// 	SchemaVersion int    `json:"schemaVersion"`
// }
// type ChannelRoute struct {
// 	Metadata *RouteMetadata  `json:"metadata,omitempty"`
// 	Streams  []string        `json:"streams,omitempty"`
// 	Pattern  string          `json:"pattern"`
// 	Topic    string          `json:"topic"`
// 	Mapping  json.RawMessage `json:"mapping,omitempty"`
// }
// type RouteMetadata struct {
// 	Id          string `json:"id"`
// 	Name        string `json:"name,omitempty"`
// 	Description string `json:"description,omitempty"`
// }
// type ChannelStream struct {
// 	Id          string            `json:"id"`
// 	Resource    string            `json:"resource"`
// 	Direction   string            `json:"direction"`
// 	Attributres map[string]string `json:"attributes,omitempty"`
// 	// Settings
// }

// // PATTERNS

// const (
// 	sendCommandPrefix = "$HIVE"
// 	sendMessagePrefix = "$BEE"
// 	sendCommandIndex  = 1

// 	PatternPublisher  = "pub"
// 	PatternSubscriber = "sub"
// 	PatternService    = "srv"
// 	PatternClient     = "cli"
// 	PatternProducer   = "prod"
// 	PatternConsumer   = "cons"
// )

// // CHANNEL INFO

// type ChannelInfo struct {
// 	Settings  []byte
// 	Resources map[string][]byte
// 	Mappings  map[string][]byte
// 	Patterns  map[string]string
// 	Topics    map[string]string
// }

// // PatternPublisher  = "pub"
// // PatternSubscriber = "sub"
// // PatternService    = "srv"
// // PatternClient     = "cli"
// // PatternProducer   = "prod"
// // PatternConsumer   = "cons"

// func (c *ChannelInfo) IsEmitter(routeId string) bool {
// 	pattern := c.Patterns[routeId]
// 	return pattern == PatternPublisher || pattern == PatternProducer || pattern == PatternClient
// }

// func (c *ChannelInfo) IsReceiver(routeId string) bool {
// 	pattern := c.Patterns[routeId]
// 	return pattern == PatternSubscriber || pattern == PatternConsumer || pattern == PatternService
// }

// // CHANNELS

// type ChannelProxy interface {
// 	IsEmbedded() bool
// 	BeeId() string
// 	Metadata() *ChannelMetadata
// 	Config() *ChannelInfo
// 	DataReceiver() chan *DataMessage
// 	CommandReceiver() chan *CommandMessage
// 	Request(route string, params Parameters, data []byte) ([]byte, error)
// 	Publish(route string, params Parameters, data interface{}) error
// 	Command(op Command, params Parameters, data []byte) ([]byte, error)
// 	Resource(id string) ([]byte, error)
// 	Resources() map[string][]byte
// 	// SharedMemory() SharedMemory
// 	RegisterDataReceiver(func(*DataMessage) error)
// 	RegisterCommandReceiver(func(*CommandMessage) error)
// }

// // MESSAGES

// type Parameters map[string]string

// type DataMessage struct {
// 	subTopic string
// 	pubTopic string
// 	prefix   string
// 	// Data     []byte
// 	Data    interface{}
// 	Route   string
// 	Pattern string
// 	// Topic     string
// 	Reply string
// 	// responder *responder
// }

// func (m *DataMessage) Topic() string {
// 	return strings.TrimPrefix(m.pubTopic, m.prefix+".")
// }

// func (m *DataMessage) Params() Parameters {
// 	params := make(Parameters)
// 	parts := strings.Split(m.Topic(), ".")
// 	for idx, p := range parts {
// 		params["topic"+strconv.Itoa(idx)] = p
// 	}
// 	return params
// }

// // func (m *DataMessage) Respond(data interface{}) error {
// // 	if m.responder == nil {
// // 		m.responder = newResponder()
// // 	}
// // 	return m.responder.Respond(m.Reply, data)
// // }

// func (m *DataMessage) DataBytes() []byte {
// 	return m.Data.([]byte)
// }

// type CommandParams map[string]interface{}
// type Command string

// const (
// 	Unknown Command = ""
// 	Deploy  Command = "DEPLOY"
// 	// Startup          = "STARTUP"
// 	// Shutdown = "SHUTDOWN"
// )

// // TOPICS FROM HIVE TO BEE
// // func SyncTopic(beeId string) string { return fmt.Sprintf("$BEEOS.BEE.%s.SYNC", beeId) }

// // TOPICS FROM BEE TO HIVE
// // func SyncRequestTopic(hubId, beeId string) string {
// // 	return fmt.Sprintf("$BEEOS.HUB.%s.BEE.%s.SYNC_REQ", hubId, beeId)
// // }

// type CommandMessage struct {
// 	Cmd    Command
// 	Params CommandParams
// 	Data   []byte
// }
