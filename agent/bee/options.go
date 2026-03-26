package bee

import (
	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/internal/runtime"
	"github.com/beeosphere/bee/agent/models"
)

// Opt is a configuration option to initialize a client
type Opt func(*agentEngine) error

func WithEnvironment(environment *models.AgentOptions) Opt {
	return func(c *agentEngine) error {
		c.environment = environment // Temporal storage, will be cleared after engine initialization
		return nil
	}
}

func WithManifest(manifest models.AgentManifest) Opt {
	return func(c *agentEngine) error {
		c.manifest = manifest
		return nil
	}
}

func WithAuthenticator(authenticator runtime.Authenticator) Opt {
	return func(c *agentEngine) error {
		c.authenticator = authenticator
		return nil
	}
}

func WithConnectorProvider(provider models.ConnectorProvider) Opt {
	return func(c *agentEngine) error {
		c.connectorProvider = provider
		return nil
	}
}

func WithAgentProvider(provider models.AgentProvider) Opt {
	return func(c *agentEngine) error {
		c.agentProvider = provider
		return nil
	}
}

func WithBus(bus core.Bus) Opt {
	return func(c *agentEngine) error {
		c.bus = bus
		return nil
	}
}

func WithStore(store models.Store) Opt {
	return func(c *agentEngine) error {
		c.store = store
		return nil
	}
}

func WithLogger(logger models.Logger) Opt {
	return func(c *agentEngine) error {
		c.log = logger
		return nil
	}
}

// TODO: Implement bus interceptors...
// func WithBusInterceptors(interceptors ...bus.BusInterceptor) Opt {
// 	return func(c *agentEngine) error {
// 		if c.bus != nil {
// 			for _, interceptor := range interceptors {
// 				c.bus.AddInterceptor(interceptor)
// 			}
// 		}
// 		return nil
// 	}
// }
