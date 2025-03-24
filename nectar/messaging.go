package nectar

import (
	"github.com/beeosphere/bee/core"
)

type SignalsReceiver func(route string, signals []*Signal) error

func RegisterSignalsReceiver(proxy core.ChannelProxy, receiver SignalsReceiver) {
	go func() {
		for {
			msg := <-proxy.DataReceiver()
			// fmt.Printf("Data received in agent: %s\n", proxy.BeeId())

			var err error
			var signals []*Signal
			if !proxy.IsEmbedded() {
				signals, _, err = DecodeSignals(msg.DataBytes())
				if err != nil {
					// TODO: log error
				}
			} else {
				if message, ok := msg.Data.(*Message); ok {
					signals = message.Signals
				}
			}
			receiver(msg.Route, signals)
		}
	}()

	// fmt.Printf("Data receiver registered for agent: %s\n", proxy.BeeId())
}
