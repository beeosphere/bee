package nectar

import (
	"fmt"

	"github.com/beeosphere/bee/core"
)

var ErrSignalNotFound = fmt.Errorf("Variable not found")

type RoutingTable interface {
	ClearRoutes()
	MapSignal(routeId string, signalId string)
	ResolveRoute(signalId string) (string, error)
}

type routingTable struct {
	table map[string]string
}

func NewRoutingTable() RoutingTable {
	return &routingTable{
		table: make(map[string]string),
	}
}

func (rt *routingTable) ClearRoutes() {
	rt.table = make(map[string]string)
}

func (rt *routingTable) MapSignal(routeId string, signalId string) {
	rt.table[signalId] = routeId
}

func (rt *routingTable) ResolveRoute(signalId string) (string, error) {
	route, ok := rt.table[signalId]
	if !ok {
		return "", ErrSignalNotFound
	}
	return route, nil
}

// type routingTable struct {
// 	table sync.Map
// }

// func NewRoutingTable() RoutingTable {
// 	return &routingTable{}
// }

// func (rt *routingTable) ClearRoutes() {
// 	rt.table = sync.Map{}
// }

// func (rt *routingTable) MapSignal(routeId string, signalId string) {
// 	rt.table.Store(signalId, routeId)
// }

// func (rt *routingTable) ResolveRoute(signalId string) (string, error) {
// 	route, ok := rt.table.Load(signalId)
// 	if !ok {
// 		return "", ErrSignalNotFound
// 	}
// 	return route.(string), nil
// }

func CreateMultiRouteMessages(routingTable RoutingTable, signals []*Signal) []*Message {
	messagesMap := make(map[string][]*Signal)

	for _, v := range signals {
		routeId, err := routingTable.ResolveRoute(v.Id)
		if err != nil && err != ErrSignalNotFound {
			continue
		}
		if _, ok := messagesMap[routeId]; !ok {
			messagesMap[routeId] = make([]*Signal, 0)
		}
		messagesMap[routeId] = append(messagesMap[routeId], v)
	}

	messages := make([]*Message, 0)
	for routeId, signals := range messagesMap {
		messages = append(messages, NewMessage(routeId, signals))
	}
	return messages
}

func CreateRouteMessages(routeId string, signals []*Signal) []*Message {
	return []*Message{
		NewMessage(routeId, signals),
	}
}

func PublishMessages(proxy core.ChannelProxy, messages []*Message) error {

	for _, msg := range messages {
		message := &Message{
			Id:      msg.Id,
			Signals: msg.Signals,
		}
		// topic := proxy.Config().Topics[msg.Id]
		err := proxy.Publish(msg.Id, nil, message)
		if err != nil {
			return err
		}
	}
	return nil

	// if proxy.IsEmbedded() {

	// 	for _, msg := range messages {
	// 		message := &Message{
	// 			Id:      msg.Id,
	// 			Signals: msg.Signals,
	// 		}
	// 		err := proxy.Publish(msg.Id, nil, message)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// for _, msg := range messages {
	// 	bytes, err := Encode(msg.Id, msg.Signals)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = proxy.Publish(msg.Id, nil, bytes)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// return nil
}
