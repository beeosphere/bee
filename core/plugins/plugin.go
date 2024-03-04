package plugins

import (
	"fmt"

	extism "github.com/extism/go-sdk"
)

type BeePlugin interface {
	Id() string
	Execute(method string, data []byte) (result []byte, err error)
}

type beePlugin struct {
	id   string // 7a336165-d59c-4bea-8a1b-3a892fdd1461
	wasm *extism.Plugin
}

func newPlugin(id string, plugin *extism.Plugin) *beePlugin {
	return &beePlugin{
		id:   id,
		wasm: plugin,
	}
}

func (p *beePlugin) Id() string {
	return p.id
}

func (p *beePlugin) Execute(method string, data []byte) (result []byte, err error) {
	_, out, err := p.wasm.Call(method, data)
	if err != nil {
		return nil, fmt.Errorf("Failed to execute plugin: %v\n", err)
	}
	return out, nil
}
