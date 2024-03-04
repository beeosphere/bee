package plugins

import (
	"context"
	"errors"
	"fmt"
	"time"

	extism "github.com/extism/go-sdk"
)

var ErrPluginNotFound = errors.New("plugin not found")

type Manager interface {
	DownloadPlugin(id, uri string) error
	LoadPlugin(id string, data []byte) error
	UnloadPlugin(id string)
	GetPlugin(id string) (BeePlugin, error)
}

type manager struct {
	plugins map[string]*beePlugin
}

func NewManager() *manager {
	return &manager{
		plugins: make(map[string]*beePlugin),
	}
}

func (m *manager) GetPlugin(id string) (BeePlugin, error) {
	if plugin, ok := m.plugins[id]; ok {
		return plugin, nil
	}
	return nil, ErrPluginNotFound
}

func (m *manager) LoadPlugin(id string, data []byte) error {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: data,
			},
		},
	}
	ctx := context.Background()
	config := extism.PluginConfig{
		EnableWasi: true,
	}
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})

	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %v", err)
	}
	m.plugins[id] = newPlugin(id, plugin)
	return nil
}

func (m *manager) UnloadPlugin(id string) {
	delete(m.plugins, id)
}

func (m *manager) DownloadPlugin(id, uri string) error {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmUrl{
				Url: uri,
			},
		},
	}
	ctx := context.Background()
	config := extism.PluginConfig{
		EnableWasi: true,
	}

	sleepFunc := extism.NewHostFunctionWithStack("sleep", func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
		time.Sleep(4 * time.Second)
	}, []extism.ValueType{}, []extism.ValueType{})

	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{sleepFunc})

	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %v", err)
	}
	m.plugins[id] = newPlugin(id, plugin)
	return nil
}
