package host

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/beeosphere/bee/core"
	"github.com/beeosphere/bee/core/ebus"
	"golang.org/x/sync/errgroup"
)

type Swarm struct {
	bees []*core.BeeEngine
}

func NewSwarm(bees ...*core.BeeEngine) *Swarm {
	return &Swarm{
		bees: bees,
	}
}

func (s *Swarm) Buzz() {
	bus := ebus.New()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	g, gctx := errgroup.WithContext(ctx)

	for _, bee := range s.bees {
		currentBee := bee
		g.Go(func() error {
			return currentBee.EmbededBuzz(gctx, bus)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("Swarm error: %v\n", err)
	}
}

// func (s *Swarm) Buzz() {

// 	bus := ebus.New()

// 	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
// 	defer stop()

// 	for _, bee := range s.bees {

// 		currentBee := bee

// 		go func() {
// 			err := currentBee.EmbededBuzz(ctx, bus)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 		}()
// 	}

// 	<-ctx.Done()
// }

// --------

// func (s *Swarm) Buzz() {

// 	bus := ebus.New()
// 	// bus := core.NewEmbedBus()

// 	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
// 	defer stop()

// 	for _, bee := range s.bees {
// 		// currentBee := bee
// 		err := bee.StartBuzzing(ctx, bus)
// 		if err != nil {
// 			fmt.Println(err)
// 		}

// 		// delay of 2 seconds between each bee
// 		time.Sleep(5 * time.Second)
// 	}

// 	fmt.Println("Swarm is buzzing...")

// 	<-ctx.Done()

// 	for _, bee := range s.bees {
// 		err := bee.StopBuzzing()
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 	}
// }
