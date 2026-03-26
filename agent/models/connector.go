package models

import (
	"encoding/json"
)

type ConnectorSetup struct {
	Config *ConnectorConfiguration
}

type ChannelMessage struct {
	Channel string
	Data    []byte
}

type ChannelHandler func(msg *ChannelMessage) error

type ChannelBus interface {
	EmitChannel(channel string, data []byte) error
	OnChannelReceived(handler ChannelHandler)
}

type ConnectorContext struct {
	AgentId    string
	InstanceId string
	Manifest   *AgentManifest
	Log        Logger
	Channels   ChannelBus
}

type Connector interface {
	Started(cctx *ConnectorContext) error
	Configured(setup *ConnectorSetup) error
	Stopped() error
}

type ConnectorProvider func(connectorType string, logger Logger) Connector

// AGENT CONFIGURATION

type AgentConfiguration struct {
	AgentType  string                    `json:"agentType"` // Hive or Bee
	Connectors []*ConnectorConfiguration `json:"connectors"`
}

type ConnectorConfiguration struct {
	Id            string `json:"id"`
	Name          string `json:"name,omitempty"`
	ConnectorType string `json:"type,omitempty"`
	SchemaVersion int    `json:"schemaVersion"`

	Settings json.RawMessage         `json:"settings,omitempty"`
	Channels []*ChannelConfiguration `json:"channels,omitempty"`
}

type ChannelConfiguration struct {
	Id      string          `json:"id"`
	Name    string          `json:"name,omitempty"`
	Pattern string          `json:"pattern"`
	Topic   string          `json:"topic"`
	Parser  string          `json:"parser,omitempty"`
	Spec    json.RawMessage `json:"spec,omitempty"`
}

// type channelBus struct {
// 	bus BusClient
// }

// func newChannelBus(bus BusClient) *channelBus {
// 	return &channelBus{bus: bus}
// }

// func (cb *channelBus) Emit(channel string, data []byte) error {
// 	return nil
// }

// func (cb *channelBus) OnReceived(handler ChannelHandler) {
// }

// // AGENT CONFIGURATION

// type AgentConfiguration struct {
// 	AgentType  string                    `json:"agentType"` // Hive or Bee
// 	Connectors []*ConnectorConfiguration `json:"connectors"`
// }
// type ConnectorConfiguration struct {
// 	Metadata *ConnectorMetadata  `json:"metadata"`
// 	Settings json.RawMessage     `json:"settings,omitempty"`
// 	Channels []*ConnectorChannel `json:"channels"`
// 	Streams  []*ConnectorStream  `json:"streams,omitempty"`
// }
// type ConnectorMetadata struct {
// 	ChannelId     string `json:"channelId"`
// 	ChannelName   string `json:"channelName"`
// 	ChannelType   string `json:"channelType"`
// 	SchemaVersion int    `json:"schemaVersion"`
// }
// type ConnectorChannel struct {
// 	Metadata *ChannelMetadata `json:"metadata,omitempty"`
// 	Streams  []string         `json:"streams,omitempty"`
// 	Pattern  string           `json:"pattern"`
// 	Topic    string           `json:"topic"`
// 	Mapping  json.RawMessage  `json:"mapping,omitempty"`
// }
// type ChannelMetadata struct {
// 	Id          string `json:"id"`
// 	Name        string `json:"name,omitempty"`
// 	Description string `json:"description,omitempty"`
// }
// type ConnectorStream struct {
// 	Id          string            `json:"id"`
// 	Resource    string            `json:"resource"`
// 	Direction   string            `json:"direction"`
// 	Attributres map[string]string `json:"attributes,omitempty"`
// 	// Settings
// }
