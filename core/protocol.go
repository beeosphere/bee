package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type ProtoMessage []byte

func ProtoMap[T any](message ProtoMessage) *T {
	var data T
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		return nil
	}
	return &data
}

type Protocol struct {
	conn          *nats.Conn
	subscriptions []*nats.Subscription
	cancels       map[string]context.CancelFunc
}

func NewProtocol() *Protocol {
	return &Protocol{
		conn:    busClient.conn,
		cancels: make(map[string]context.CancelFunc),
	}
}

type EmitOptions struct {
	Interval int
	Duration int
}

func (p *Protocol) Emit(signal string, data interface{}, options *EmitOptions) {
	// fmt.Println("start emitting data")
	if options == nil {
		options = &EmitOptions{
			Interval: 5,
			Duration: 60,
		}
	}
	// cancellable context
	ctx, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(options.Duration)*time.Second)
	defer cancelTimeout()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	p.cancels[signal] = cancel

	ticker := time.NewTicker(time.Duration(options.Interval) * time.Second)
	p.Send(signal, data)

	go func() {
	loop:
		for {
			select {
			case <-ticker.C:
				// Publishes again
				p.Send(signal, data)

			case <-ctx.Done():
				ticker.Stop()
				// delete(p.cancels, signal)
				break loop // Exit the loop
			}
		}
		// fmt.Println("Emitter stopped")
	}()
}

func (p *Protocol) AbortEmit(signal string) {
	if cancel, ok := p.cancels[signal]; ok {
		cancel()
		delete(p.cancels, signal)

		// fmt.Println("emit aborted")
	}
}

func (p *Protocol) Send(signal string, data interface{}) error {
	var err error
	bytes, ok := data.([]byte)
	if !ok {
		bytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	// fmt.Println("sending data")
	return p.conn.Publish(signal, bytes)
}

func (p *Protocol) On(signal string, handler func(protocol *Protocol, message ProtoMessage)) error {
	subs, err := p.conn.Subscribe(signal, func(msg *nats.Msg) {
		handler(p, ProtoMessage(msg.Data))
	})
	if err != nil {
		return fmt.Errorf("error subscribing to signal: %w", err)
	}
	p.subscriptions = append(p.subscriptions, subs)
	return nil
}

// func (p *Protocol) On(signal string, handler func(protocol *Protocol, data interface{})) error {
// 	subs, err := p.conn.Subscribe(signal, func(msg *nats.Msg) {
// 		var data interface{}
// 		err := json.Unmarshal(msg.Data, &data)
// 		if err != nil {
// 			fmt.Println("error deserializing data: ", err)
// 		}
// 		handler(p, &data)
// 	})
// 	if err != nil {
// 		return fmt.Errorf("error subscribing to signal: %w", err)
// 	}
// 	p.subscriptions = append(p.subscriptions, subs)
// 	return nil
// }

func (p *Protocol) Dispose() error {
	for _, sub := range p.subscriptions {
		err := sub.Unsubscribe()
		if err != nil {
			return fmt.Errorf("error unsubscribing from signal: %w", err)
		}
	}
	return nil
}
