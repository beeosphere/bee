package runtime

import (
	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

// AGENT PROCESSOR STRATEGY

type AgentProcessor struct {
	session       *core.Session
	log           models.Logger
	busClient     models.BusClient
	commander     models.Commander
	agent         models.Agent
	agentProvider models.AgentProvider
}

func NewAgentProcessor(session *core.Session, provider models.AgentProvider, commander models.Commander, busClient models.BusClient) *AgentProcessor {
	return &AgentProcessor{
		session:       session,
		log:           session.Log,
		busClient:     busClient,
		commander:     commander,
		agentProvider: provider,
	}
}

func (p *AgentProcessor) Process(model *models.Model) error {
	if err := p.ensureAgentInitialization(); err != nil {
		return err
	}
	if err := p.agent.Configured(model); err != nil {
		return err
	}
	return nil
}

func (p *AgentProcessor) ensureAgentInitialization() error {
	if p.agent == nil {
		p.agent = p.agentProvider(p.log)

		context := &models.AgentContext{
			AgentId:    p.session.Bee,
			InstanceId: p.session.PublicKey,
			Manifest:   p.session.Manifest,
			Log:        p.log,
			Bus:        p.busClient,
			Commands:   p.commander,
		}
		err := p.agent.Started(context)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *AgentProcessor) Dispose() error {
	if p.agent != nil {
		return p.agent.Stopped()
	}
	return nil
}
