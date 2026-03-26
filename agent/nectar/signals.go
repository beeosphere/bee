package nectar

import "encoding/json"

// NECTAR MESSAGING

type Signal struct {
}

type Labels map[string]string

type NectarMessage struct {
	Labels  Labels
	Signals []Signal
}

func (nm *NectarMessage) AddSignal(signal Signal) {
	if nm.Signals == nil {
		nm.Signals = make([]Signal, 0)
	}
	nm.Signals = append(nm.Signals, signal)
}

func (nm *NectarMessage) SetLabel(key, value string) {
	if nm.Labels == nil {
		nm.Labels = make(Labels)
	}
	nm.Labels[key] = value
}

// NECTAR CONFIGURATION

type NectarSpec struct {
	Options NectarOptions `json:"options,omitempty"`
	Labels  Labels        `json:"labels,omitempty"`
	Signals []*SignalSpec `json:"signals,omitempty"`
}

type NectarOptions struct {
	SamplingInterval int `json:"samplingInterval,omitempty"`
}

type SignalSpec struct {
	Id   string          `json:"id"`
	Type string          `json:"type"`
	Map  json.RawMessage `json:"map"`
}
