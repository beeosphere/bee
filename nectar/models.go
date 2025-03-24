package nectar

import "time"

type Signal struct {
	Id        string    `json:"id"`
	Group     string    `json:"group"`
	Value     any       `json:"value"`
	Type      string    `json:"type"`
	State     int       `json:"state"`
	Timestamp time.Time `json:"ts"`
}

type Message struct {
	Id      string    `json:"id"`
	Signals []*Signal `json:"signals"`
}

func NewMessage(id string, signals []*Signal) *Message {
	return &Message{
		Id:      id,
		Signals: signals,
	}
}

type SignalsFunc func([]*Signal) error

type SignalsProcessor interface {
	UpdateSignals(values []*Signal) error
	OnSignalsChanged(handler SignalsFunc)
}
