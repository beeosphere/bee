package busnats

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/internal/core/topics"
	"github.com/beeosphere/bee/agent/models"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

type busNats struct {
	session *core.Session
	log     models.Logger
	conn    *nats.Conn
}

func NewBus(session *core.Session) core.Bus {
	return &busNats{session: session, log: session.Log}
}

func (b *busNats) Connect(ctx context.Context) error {

	log := b.log

	// UserJWTHandler is used to fetch and return the account signed JWT for this user.
	userJWTHandler := func() (string, error) {
		return b.session.BusToken(), nil
	}

	// SignatureHandler is used to sign a nonce from the server while authenticating with nkeys.
	// The user should sign the nonce and return the raw signature. The client will base64 encode this to send to the server.
	signatureHandler := func(nonce []byte) ([]byte, error) {
		seed := []byte(b.session.Seed())
		return signNonce(seed, nonce)
	}

	// Connect to a server
	urls := strings.Join(b.session.BusAddresses, ",")

	type result struct {
		conn *nats.Conn
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		con, err := nats.Connect(
			urls,
			nats.MaxReconnects(100000), // 1000 -> 30 mins aprox
			nats.ReconnectWait(10*time.Second),
			nats.RetryOnFailedConnect(true),
			nats.Name(b.session.Bee),
			nats.UserJWT(userJWTHandler, signatureHandler),
			nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
				// log.Warnf("Disconnected. Reason: %q\n", err)
				log.Warn("Connection lost")
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				log.Infof("Reconnected to %v\n", nc.ConnectedUrl())
			}),
			nats.ClosedHandler(func(nc *nats.Conn) {
				log.Info("Connection closed")
			}),
			nats.ConnectHandler(func(nc *nats.Conn) {
				s := b.session
				log.Infof("Connected to hive: %s (hub: %s) as %s (pubKey: %s)\n", urls, s.Hub, s.Bee, core.ShortValue(s.PublicKey))
			}),
		)
		ch <- result{con, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			log.Errorf("Error while connecting to Hive")
			return r.err
		}
		b.conn = r.conn
		return nil
	case <-ctx.Done():
		return fmt.Errorf("connection cancelled: %w", ctx.Err())
	}
}

func (b *busNats) Disconnect() error {
	defer b.log.Info("Bee disconnected")

	return b.conn.Drain()
}

func (b *busNats) Publish(topic string, data interface{}, headers models.BusHeaders) error {
	var err error
	// Serialize data if needed (if data is of type []byte, no serialization is done)
	bytes, ok := data.([]byte)
	if !ok {
		bytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	// Prepare message subject
	var subject string
	if strings.HasPrefix(topic, topics.CMD_PREFIX) {
		subject = topic
		b.log.Tracef("Command [hive <- bee: %s]", subject)
	} else {
		subject = topics.MessageSubject(topic)
		b.log.Tracef("Message [hive <- bee: %s (%d bytes)]", topic, len(bytes))
	}
	// Add headers if any
	header := nats.Header{}
	if headers != nil {
		for k, v := range headers.ToMap() {
			header.Add(k, v)
		}
	}
	// Publish message
	return b.conn.PublishMsg(&nats.Msg{
		Subject: subject,
		Data:    bytes,
		Header:  header,
	})
}

func (b *busNats) Subscribe(topic string, handler models.BusHandler) (models.BusSubscription, error) {
	// Prepare message subject
	var subject string
	if strings.HasPrefix(topic, topics.CMD_PREFIX) {
		subject = topic
	} else {
		subject = topics.MessageSubject(topic)
	}

	subs, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {

		var subscribedTopic string
		var receivedTopic string
		if strings.HasPrefix(msg.Subject, topics.CMD_PREFIX) {
			receivedTopic = msg.Subject
			subscribedTopic = msg.Sub.Subject
			b.log.Tracef("Command [hive -> bee: %s]", receivedTopic)
		} else {
			receivedTopic = topics.FetchMessagePart(msg.Subject)
			subscribedTopic = topics.FetchMessagePart(msg.Sub.Subject)
			b.log.Tracef("Message [hive -> bee: %s (%d bytes)]", receivedTopic, len(msg.Data))
		}

		headers := core.NewHeaders()
		for k, v := range msg.Header {
			headers.Set(k, v[0])
		}
		message := &models.BusMessage{
			Topic:           core.NewTopic(receivedTopic),
			SubscribedTopic: subscribedTopic,
			Data:            msg.Data,
			Headers:         headers,
		}

		handler(message)
	})
	if err != nil {
		return nil, err
	}
	return core.NewSubscription(subs), nil
}

func signNonce(seed, nonce []byte) ([]byte, error) {
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	return kp.Sign(nonce)
}
