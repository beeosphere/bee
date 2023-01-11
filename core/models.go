package core

import "encoding/json"

// BEE CONFIGURATION
type BeeConfiguration struct {
	Channels []*ChannelConfiguration `json:"channels"`
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
