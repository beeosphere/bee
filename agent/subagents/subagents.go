package subagents

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/beeosphere/bee/agent/models"
)

type SubmodelResolver func(name string, model models.Model) (submodel []byte, err error)

type Manager interface {
	models.Agent
	RegisterSubagent(name string, subagent models.Agent, modelProvider SubmodelResolver) error
}

type subagentManager struct {
	log               models.Logger
	agentId           string
	subagents         map[string]models.Agent
	submodelProviders map[string]SubmodelResolver
	submodelHashes    map[string]string
}

func NewManager(agentId string, logger models.Logger) *subagentManager {
	return &subagentManager{
		log:               logger,
		agentId:           agentId,
		subagents:         make(map[string]models.Agent),
		submodelProviders: make(map[string]SubmodelResolver),
		submodelHashes:    make(map[string]string),
	}
}

func (m *subagentManager) RegisterSubagent(name string, subagent models.Agent, modelProvider SubmodelResolver) error {
	m.subagents[name] = subagent
	m.submodelProviders[name] = modelProvider
	m.log.Debugf("Registered subagent '%s' for agent '%s'", name, m.agentId)
	return nil
}

func (m *subagentManager) Started(ac *models.AgentContext) error {
	// Call Started on all subagents
	for name, subagent := range m.subagents {

		if err := subagent.Started(ac); err != nil {
			m.log.Warnf("Failed to start subagent %s: %v", name, err)
		}
	}
	return nil
}

func (m *subagentManager) Configured(model *models.Model) error {
	// Call Configured on all subagents with their respective submodels if they have changed since the last configuration
	for name, subagent := range m.subagents {

		submodelData, err := m.getSubmodel(name, *model)
		if err != nil {
			m.log.Warnf("Failed to get submodel for subagent '%s': %v", name, err)
			continue
		}
		hash := fmt.Sprintf("%x", sha256.Sum256(submodelData))
		if lastHash, seen := m.submodelHashes[name]; seen && lastHash == hash {
			continue
		}
		m.log.Debugf("Submodel received for subagent '%s' (hash: %s)", name, m.shortValue(hash))
		m.submodelHashes[name] = hash
		submodel := &models.Model{
			Id:        model.Id,
			Hash:      model.Hash,
			Data:      submodelData,
			Resources: model.Resources,
		}
		if err = subagent.Configured(submodel); err != nil {
			m.log.Warnf("Failed to configure subagent %s: %v", name, err)
			continue
		}
	}
	return nil
}

func (m *subagentManager) Stopped() error {
	// Call Stopped on all subagents
	for name, subagent := range m.subagents {

		if err := subagent.Stopped(); err != nil {
			m.log.Warnf("Failed to stop subagent '%s': %v", name, err)
		}
	}
	return nil
}

func (m *subagentManager) getSubmodel(name string, model models.Model) ([]byte, error) {
	provider, ok := m.submodelProviders[name]
	if !ok {
		return nil, errors.New("no provider for subagent " + name)
	}
	return provider(name, model)
}

func (m *subagentManager) shortValue(value string) string {
	if len(value) > 8 {
		return ".." + value[len(value)-8:]
	}
	return value
}
