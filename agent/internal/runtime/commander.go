package runtime

import (
	"context"
	"time"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

const rootHivePath = ""

const (
	DEPLOY_SIG = "DEPLOY_SIG"
	DEPLOY     = "DEPLOY"
	DEPLOY_ACK = "DEPLOY_ACK"
	DEPLOY_REQ = "DEPLOY_REQ"
	DEPLOYED   = "DEPLOYED"
)

// // CommandReceiver is a callback function that handles received commands.
// // It receives the command name, associated data, and bus message headers.
// type CommandReceiver func(data any, headers bus.BusHeaders)

// commander manages command sending, receiving, and periodic emission.
// It handles command subscriptions, callbacks, and active emitters.
type commander struct {
	session      *core.Session
	bus          models.BusClient
	callbacks    map[string][]models.CommandReceiver
	cancels      map[string]context.CancelFunc
	subscription models.BusSubscription
}

// NewCommander creates a new Commander instance with the given session and bus client.
// It initializes the callbacks and cancels maps for managing command handlers and emitters.
func NewCommander(session *core.Session, bus models.BusClient) *commander {
	return &commander{
		session:   session,
		bus:       bus,
		callbacks: make(map[string][]models.CommandReceiver),
		cancels:   make(map[string]context.CancelFunc),
	}
}

// SubscribeCommands subscribes to command reception topics for this bee.
// When a command is received, it invokes all registered callbacks for that command.
// Returns an error if the subscription fails.
func (c *commander) SubscribeCommands() error {
	sub, err := c.bus.Subscribe(core.CommandsReceptionTopic(c.session.Bee), func(msg *models.BusMessage) {
		if command := msg.Topic.Command(); command != "" {

			if receivers, ok := c.callbacks[command]; ok {
				for _, receiver := range receivers {
					receiver(msg.Data, msg.Headers)
				}
			}
		}
	})
	if err != nil {
		return err
	}
	c.subscription = sub
	return nil
}

// UnsubscribeCommands unsubscribes from command reception and clears all registered callbacks.
// Returns an error if the unsubscription fails.
func (c *commander) UnsubscribeCommands() error {
	if c.subscription != nil {
		if err := c.subscription.Unsubscribe(); err != nil {
			return err
		}
		c.subscription = nil
	}
	c.callbacks = make(map[string][]models.CommandReceiver)
	return nil
}

// Send publishes a command to the root hive with the specified data and headers.
// Returns an error if the publish operation fails.
func (c *commander) Send(command string, data any, headers models.BusHeaders) error {
	return c.bus.Publish(core.RootCommandTopic(command), data, headers)
}

// SendToHive publishes a command to a specific hive path with the specified data and headers.
// Returns an error if the publish operation fails.
func (c *commander) SendToHive(command, hivePath string, data any, headers models.BusHeaders) error {
	return c.bus.Publish(core.HiveCommandTopic(command, hivePath), data, headers)
}

// OnCommandReceived registers a callback to be invoked when the specified command is received.
// Multiple callbacks can be registered for the same command.
func (c *commander) OnCommandReceived(command string, receiver models.CommandReceiver) {
	if c.callbacks[command] == nil {
		c.callbacks[command] = []models.CommandReceiver{}
	}
	c.callbacks[command] = append(c.callbacks[command], receiver)
}

// Emit starts periodic emission of a command to the root hive.
// The command is sent immediately and then repeated at the specified interval (in seconds)
// for the specified duration (in seconds). Any existing emitter for the same command is cancelled.
func (c *commander) Emit(command string, data any, headers models.BusHeaders, interval, duration int) {
	c.EmitToHive(command, rootHivePath, data, headers, interval, duration)
}

// EmitToHive starts periodic emission of a command to a specific hive path.
// The command is sent immediately and then repeated at the specified interval (in seconds)
// for the specified duration (in seconds). Any existing emitter for the same command is cancelled.
func (c *commander) EmitToHive(command, hivePath string, data any, headers models.BusHeaders, interval, duration int) {
	// Cancel existing emitter if active
	if cancel, ok := c.cancels[command]; ok {
		cancel()
		delete(c.cancels, command)
	}

	// fmt.Println("start emitting data")
	sendCommand := func() {
		if hivePath == rootHivePath {
			c.Send(command, data, headers)
		} else {
			c.SendToHive(command, hivePath, data, headers)
		}
	}
	// cancellable context
	ctx, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
	defer cancelTimeout()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.cancels[command] = cancel

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	sendCommand()

	go func() {
	loop:
		for {
			select {
			case <-ticker.C:
				// Publishes again
				sendCommand()

			case <-ctx.Done():
				ticker.Stop()
				// delete(p.cancels, signal)
				break loop // Exit the loop
			}
		}
		// fmt.Println("Emitter stopped")
	}()
}

// CancelEmit stops the periodic emission of the specified command.
// If no emitter exists for the command, this is a no-op.
func (c *commander) CancelEmit(command string) {
	if cancel, ok := c.cancels[command]; ok {
		cancel()
		delete(c.cancels, command)
	}
}
