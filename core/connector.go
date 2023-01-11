package core

// type ChannelFactory interface {
// 	CreateChannel(proxy ChannelProxy) error
// 	RemoveChannel(channelId string) error
// }

// type Channel interface {
// 	Start(proxy ChannelProxy) error
// 	Restart() error
// 	Stop() error
// }

// type ChannelProvider func(channelType string) Channel

// type Connector struct {
// 	channels        map[string]Channel
// 	channelSelector ChannelProvider
// }

// func NewConnector(selector ChannelProvider) *Connector {
// 	return &Connector{
// 		channels:        make(map[string]Channel),
// 		channelSelector: selector,
// 	}
// }

// func (c *Connector) CreateChannel(proxy ChannelProxy) error {

// 	// TODO: If channel id exists remove from channel list?

// 	channel := c.channelSelector(proxy.Metadata().ChannelType)
// 	if channel == nil {
// 		return errors.New("Incorrect channel type")
// 	}

// 	c.channels[proxy.Metadata().ChannelId] = channel
// 	return channel.Start(proxy)
// }

// func (c *Connector) RemoveChannel(channelId string) error {

// 	defer delete(c.channels, channelId)
// 	return c.channels[channelId].Stop()
// }
