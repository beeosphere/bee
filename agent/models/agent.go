package models

// CONFIG OPTIONS

type ConfigOptions struct {
	Agent  *AgentOptions            `koanf:"agent"`
	Agents map[string]*AgentOptions `koanf:"agents"`
}

type AgentOptions struct {
	Id       string         `koanf:"id"`
	Key      string         `koanf:"key"`
	HiveUri  string         `koanf:"hive"`
	LogLevel string         `koanf:"log"`
	Path     string         `koanf:"path"`
	Source   *SourceOptions `koanf:"config"`
}

type SourceOptions struct {
	Path  string `koanf:"path"`
	Watch bool   `koanf:"watch"`
}

// MANIFEST

type AgentManifest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

// AGENT

type Model struct {
	Id        string
	Hash      string
	Data      []byte
	Resources map[string][]byte
}

type AgentContext struct {
	AgentId    string
	InstanceId string
	Manifest   AgentManifest
	Log        Logger
	Commands   Commander
	Bus        BusClient
}

type Agent interface {
	Started(actx *AgentContext) error
	Configured(model *Model) error
	Stopped() error
}

type AgentProvider func(logger Logger) Agent
