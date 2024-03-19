package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	log "github.com/sirupsen/logrus"
)

type bus struct {
	session *session
	conn    *nats.Conn
}

func newBus(session *session) *bus {
	return &bus{session: session}
}

func (b *bus) Connect() error {

	// UserJWTHandler is used to fetch and return the account signed JWT for this user.
	userJWTHandler := func() (string, error) {
		return b.session.BusToken(), nil
	}

	// SignatureHandler is used to sign a nonce from the server while authenticating with nkeys.
	// The user should sign the nonce and return the raw signature. The client will base64 encode this to send to the server.
	signatureHandler := func(nonce []byte) ([]byte, error) {
		seed := []byte(b.session.Seed())
		return b.SignNonce(seed, nonce)
	}

	// Connect to a server
	urls := strings.Join(b.session.busAddresses, ",")

	con, err := nats.Connect(
		urls,
		nats.MaxReconnects(100000), // 1000 -> 30 mins aprox
		nats.ReconnectWait(10*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.Name(b.session.bee),
		nats.UserJWT(userJWTHandler, signatureHandler),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Warnf("Disconnected. Reason: %q\n", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Infof("Reconnected to %v\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Errorf("Connection closed. Reason: %q\n", nc.LastError())
		}),
	)
	if err != nil {
		fmt.Printf("Error while connecting to Hive")
		return nil
	}

	s := b.session
	log.Infof("Connected to Hive: %s. Hub:%s. Bee:%s. PubKey:%s\n", urls, s.hub, s.bee, s.publicKey)

	// msg := &nats.Msg{Subject: fmt.Sprintf("$BEEOS.HUB.%s.BEE.%s.SYNC_REQ", b.session.hub, b.session.bee), Data: []byte("Message from bee")}
	// res, err := con.RequestMsg(msg, 5*time.Second)
	// if err != nil {
	// 	fmt.Printf("Error while requesting hive (%s)", err)
	// } else {
	// 	fmt.Println(string(res.Data))
	// }
	// con.Publish("$BEEOS.BEE.bee2", []byte("Hello from bee2!"))

	// con.QueueSubscribe("$BEEOS.BEE.bee2.SYNC", "bee2", func(msg *nats.Msg) {
	// 	fmt.Println(string(msg.Data))
	// })

	b.conn = con
	return nil

	// ------

	// // nats.TokenHandler(func() string {
	// // 	return ""
	// // })

	// nc, err := nats.Connect(nats.DefaultURL)
	// nc.Close()

	// return err
}

func (*bus) SignNonce(seed, nonce []byte) ([]byte, error) {

	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	return kp.Sign(nonce)
}
