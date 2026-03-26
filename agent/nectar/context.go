package nectar

import "github.com/beeosphere/bee/agent/models"

type NectarContext struct {
	AgentId    string
	InstanceId string
	Manifest   *models.AgentManifest
	Log        models.Logger
	Messaging  NectarBus
}

func ToNectarContext(cctx *models.ConnectorContext) *NectarContext {
	messaging := newNectarBus(cctx.Channels)

	return &NectarContext{
		AgentId:    cctx.AgentId,
		InstanceId: cctx.InstanceId,
		Manifest:   cctx.Manifest,
		Log:        cctx.Log,
		Messaging:  messaging,
	}
}
