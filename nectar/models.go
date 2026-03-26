package nectar

import "time"

const (
	// The signal is good.
	Good uint32 = 0x00000000
	// An unexpected error occurred.
	BadUnexpectedError uint32 = 0x80010000
	// An internal error occurred as a result of a programming or configuration error.
	BadInternalError uint32 = 0x80020000
	// A low level communication error occurred.
	BadCommunicationError uint32 = 0x80050000
	// Communication with the data source is defined, but not established, and there is no last known value available.
	BadNoCommunication uint32 = 0x80310000
	// Waiting for the server to obtain values from the underlying data source.
	BadWaitingForInitialData uint32 = 0x80320000
)

type Signal struct {
	Id        string    `json:"id"`
	Group     string    `json:"group"`
	Value     any       `json:"value"`
	Type      string    `json:"type"`
	State     uint32    `json:"state"`
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
