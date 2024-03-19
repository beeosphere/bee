package core

import (
	"encoding/json"
	"time"
)

// BEE CONFIGURATION
type BeeConfiguration struct {
	AgentType string                  `json:"agentType"` // Hive or Bee
	Channels  []*ChannelConfiguration `json:"channels"`
}
type ChannelConfiguration struct {
	Metadata *ChannelMetadata `json:"metadata"`
	Settings json.RawMessage  `json:"settings,omitempty"`
	Routes   []*ChannelRoute  `json:"routes"`
	Streams  []*ChannelStream `json:"streams,omitempty"`
}
type ChannelMetadata struct {
	ChannelId     string `json:"channelId"`
	ChannelName   string `json:"channelName"`
	ChannelType   string `json:"channelType"`
	SchemaVersion int    `json:"schemaVersion"`
}
type ChannelRoute struct {
	Metadata *RouteMetadata  `json:"metadata,omitempty"`
	Streams  []string        `json:"streams,omitempty"`
	Pattern  string          `json:"pattern"`
	Topic    string          `json:"topic"`
	Mapping  json.RawMessage `json:"mapping,omitempty"`
}
type RouteMetadata struct {
	Id          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
type ChannelStream struct {
	Id          string            `json:"id"`
	Resource    string            `json:"resource"`
	Direction   string            `json:"direction"`
	Attributres map[string]string `json:"attributes,omitempty"`
	// Settings
}

// DEPLOYMENTS
type ResourceBinding struct {
	Id   string `json:"id"`
	Hash string `json:"hash"`
}
type DeployBinding struct {
	AgentId    string            `json:"agentId"`
	HiveId     string            `json:"hiveId"`
	ConfigId   string            `json:"configId"`
	ConfigHash string            `json:"configHash"`
	Resources  []ResourceBinding `json:"resources,omitempty"`
}

func (b *DeployBinding) IsEmpty() bool {
	return b.ConfigId == "" && b.ConfigHash == ""
}

type DeployRequest struct {
	AgentId   string    `json:"agentId"`
	Timestamp time.Time `json:"timestamp"`
}

type Deployed struct {
	AgentId   string    `json:"agentId"`
	Timestamp time.Time `json:"timestamp"`
}

type DeployFailed struct {
	AgentId   string    `json:"agentId"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
}

// SHARED MEMORY
type MemoryValue struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// func newMemoryValue(key, value string) *MemoryValue {
// 	return &MemoryValue{
// 		Key:       key,
// 		Value:     value,
// 		Timestamp: time.Now(),
// 	}
// }

type Variables map[string]string

// NECTAR MESSGES
type NectarVariable struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}
type NectarMessage struct {
	Timestamp int64             `json:"ts"`
	Variables []*NectarVariable `json:"variables"`
}
