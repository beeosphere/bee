package core

import (
	"encoding/json"
	"time"
)

// DEPLOYMENT COMMANDS

type DeploySignalCommand struct {
	AgentId   string    `json:"agentId"`
	ModelId   string    `json:"modelId"`
	ModelHash string    `json:"modelHash"`
	Timestamp time.Time `json:"timestamp"`
}

type DeployRequestCommand struct {
	AgentId          string    `json:"agentId"`
	PubKey           string    `json:"pubKey"`
	CurrentModelId   string    `json:"currentModelId,omitempty"`
	CurrentModelHash string    `json:"currentModelHash,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

type DeployResponseCommand struct {
	AgentId   string          `json:"agentId"`
	ModelId   string          `json:"modelId"`
	ModelHash string          `json:"modelHash"`
	Model     json.RawMessage `json:"model,omitempty"`
	Cached    bool            `json:"cached"`
	Synced    bool            `json:"synced"`
	Timestamp time.Time       `json:"timestamp"`
}

type DeployedCommand struct {
	AgentId   string    `json:"agentId"`
	PubKey    string    `json:"pubKey"`
	ModelId   string    `json:"modelId"`
	ModelHash string    `json:"modelHash"`
	Failed    bool      `json:"failed"`
	Errors    []string  `json:"errors,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (c *DeployedCommand) AddError(error string) {
	c.Errors = append(c.Errors, error)
	c.Failed = true
}

type DeployAckCommand struct {
	AgentId   string    `json:"agentId"`
	PubKey    string    `json:"pubKey"`
	Timestamp time.Time `json:"timestamp"`
}
