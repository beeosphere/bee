package nectar

import (
	"errors"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

type NectarHandler func(channel string, msg *NectarMessage) error

type NectarBus interface {
	EmitSignals(channel string, msg *NectarMessage) error
	OnSignalsReceived(handler NectarHandler)
}

type nectarBus struct {
	channelBus models.ChannelBus
}

func newNectarBus(bus models.ChannelBus) NectarBus {
	return &nectarBus{channelBus: bus}
}

func (nb *nectarBus) EmitSignals(channel string, msg *NectarMessage) error {
	data, err := core.Serialize(msg)
	if err != nil {
		return err
	}
	return nb.channelBus.EmitChannel(channel, data)
}

func (nb *nectarBus) OnSignalsReceived(handler NectarHandler) {
	nb.channelBus.OnChannelReceived(func(msg *models.ChannelMessage) error {

		if data := core.DirectDeserialize[NectarMessage](msg.Data); data != nil {
			return handler(msg.Channel, data)
		}
		return errors.New("incorrect format for nectar signals")
	})
}
