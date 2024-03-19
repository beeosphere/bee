package core

import (
	"context"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

type BeeOptions struct {
	Id      string
	Key     string
	HiveUri string
}

type BeeEngine struct {
	session     *session
	provisioner *provisioner
	controller  *controller
}

var busClient *bus

func NewBee(options *BeeOptions, provider ChannelProvider) *BeeEngine {

	session := newSession(options.Id, options.Key, options.HiveUri)
	http := NewHttpClient(options, session)

	busClient = newBus(session)

	return &BeeEngine{
		session:     session,
		provisioner: newProvisioner(session, http),
		controller:  newController(session, http, provider),
	}
}

func (e *BeeEngine) Buzz() error {

	log.SetLevel(log.DebugLevel) // TODO: Set by options

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := e.connect(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	<-ctx.Done()

	err = e.disconnect()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (e *BeeEngine) connect(ctx context.Context) error {

	// Step 1. Connects to BeeOS Server and authenticates using NKeys
	if err := e.provisioner.openSession(ctx); err != nil {
		return err
	}

	// Step 2. Subscribes to connector system topics
	if err := e.controller.startup(); err != nil {
		return err
	}

	return nil
}

func (e *BeeEngine) disconnect() error {
	defer log.Info("Bee disconnected")

	return e.controller.shutdown()
}
