package topics

import "fmt"

// Operations
const (
	cmd_DEPLOY         = "DEPLOY"
	cmd_DEPLOY_REQUEST = "DEPLOY_REQ"
	cmd_DEPLOYED       = "DEPLOYED"
	cmd_DEPLOY_FAILED  = "DEPLOY_FAILED"
	cmd_STATS          = "STATS"

	// Bee Topics
	template_HIVE_COMMANDS_TOPIC  = "$HIVE.%s"
	template_AGENT_COMMANDS_TOPIC = "$AGENT.ID.%s.%s"
	template_DATA_TOPIC           = "$BEE.%s"

	// Bee topic indices
	REQUEST_BEE_IDX = 4
	REQUEST_OP_IDX  = 5
)

func DeployRequest() string {
	return fmt.Sprintf(template_HIVE_COMMANDS_TOPIC, cmd_DEPLOY_REQUEST)
}

func DeploySuccess() string {
	return fmt.Sprintf(template_HIVE_COMMANDS_TOPIC, cmd_DEPLOYED)
}

func DeployError() string {
	return fmt.Sprintf(template_HIVE_COMMANDS_TOPIC, cmd_DEPLOY_FAILED)
}

func DeployReceived(agentId string) string {
	return fmt.Sprintf(template_AGENT_COMMANDS_TOPIC, agentId, cmd_DEPLOY)
}
