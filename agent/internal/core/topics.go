package core

import (
	"fmt"
	"strings"
)

const (
	BEEOS_MSG_PREFIX = "beeos.msg."
	BEEOS_CMD_PREFIX = "beeos.cmd."
)

func RootCommandTopic(command string) string {
	return fmt.Sprintf("beeos.cmd.%s.hive", command)
}
func HiveCommandTopic(command, hivePath string) string {
	return fmt.Sprintf("beeos.cmd.%s.hive.%s", command, hivePath)
}
func CommandsReceptionTopic(agentId string) string {
	return fmt.Sprintf("beeos.cmd.*.agent.%s.>", agentId)
}

func CommandFromTopic(topic string) string {
	// If topic starts with "beeos.cmd.", extract the command from part 2 of the topic
	var command string
	if strings.HasPrefix(topic, BEEOS_CMD_PREFIX) {
		parts := strings.Split(topic, ".")
		if len(parts) < 4 {
			return "" //, fmt.Errorf("invalid topic format")
		}
		command = parts[2]
		return command
	}
	return "" //, fmt.Errorf("topic does not start with beeos.cmd.")
}
