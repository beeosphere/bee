package topics

import (
	"fmt"
	"strings"
)

const (
	COMMAND_INDEX = 2

	MSG_PREFIX = "beeos.msg."
	CMD_PREFIX = "beeos.cmd."
)

// COMMAND MANAGEMENT

func FetchCommand(topic string) string {
	// If topic starts with "beeos.cmd.", extract the command from part 2 of the topic
	var command string
	if strings.HasPrefix(topic, CMD_PREFIX) {
		parts := strings.Split(topic, ".")
		if len(parts) < 4 {
			return "" //, fmt.Errorf("invalid topic format")
		}
		command = parts[COMMAND_INDEX]
		return command
	}
	return ""
}

func FetchMessagePart(subject string) string {
	if strings.HasPrefix(subject, MSG_PREFIX) {
		return strings.TrimPrefix(subject, MSG_PREFIX)
	}
	return ""
}

// AGENT MANAGED TOPICS

// Topic for sending a command to a specific hive
func CommandToHive(command, hivePath string) string {
	return fmt.Sprintf("beeos.cmd.%s.hive.%s", command, hivePath)
}

// Topic for sending a command to the root hive
func CommandToRootHive(command string) string {
	return fmt.Sprintf("beeos.cmd.%s.hive", command)
}

// Topic for subscribing to all commands for a specific agent
func CommandFromHive(agentId string) string {
	return fmt.Sprintf("beeos.cmd.*.agent.%s.hive.>", agentId)
}

func MessageSubject(topic string) string {
	return fmt.Sprintf("beeos.msg.%s", topic)
}

// HIVE MANAGED TOPICS

// Topic for sending a command to an agent
func CommandToAgent(agentId, command, hivePath string) string {
	return fmt.Sprintf("beeos.cmd.%s.agent.%s.hive.%s", command, agentId, hivePath)
}

// Topic for subscribing to all commands from agents to upstream hive
func CommandFromAgentToRootHive() string {
	return "beeos.cmd.*.hive"
}

// Topic for subscribing to all commands from agents to hive
func CommandFromAgentToHive() string {
	return "beeos.cmd.*.hive.>"
}
