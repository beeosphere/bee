package bee

import (
	"context"
	"os"
	"os/signal"
)

type Bee interface {
	Buzz() error
}

type beeEngine struct {
	agent Agent
}

func NewBee(ops ...Opt) Bee {
	agent := NewAgent(ops...)
	return &beeEngine{agent: agent}
}

func (b *beeEngine) Buzz() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := b.agent.Run(); err != nil {
		return err
	}

	<-ctx.Done()

	if err := b.agent.Stop(); err != nil {
		return err
	}
	return nil
}
