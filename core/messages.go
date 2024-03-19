package core

import (
	"fmt"
	"strconv"
	"strings"
)

type DataMessage struct {
	Data      []byte
	Route     string
	Pattern   string
	Topic     string
	Reply     string
	responder *responder
}

func (m *DataMessage) Params() Parameters {
	params := make(Parameters)
	parts := strings.Split(m.Topic, ".")
	for idx, p := range parts {
		params["topic"+strconv.Itoa(idx)] = p
	}
	return params
}

func (m *DataMessage) Respond(data interface{}) error {
	if m.responder == nil {
		m.responder = newResponder()
	}
	return m.responder.Respond(m.Reply, data)
}

type CommandParams map[string]interface{}
type Command string

const (
	Unknown Command = ""
	Deploy  Command = "DEPLOY"
	// Startup          = "STARTUP"
	// Shutdown = "SHUTDOWN"
)

// TOPICS FROM HIVE TO BEE
func SyncTopic(beeId string) string { return fmt.Sprintf("$BEEOS.BEE.%s.SYNC", beeId) }

// TOPICS FROM BEE TO HIVE
func SyncRequestTopic(hubId, beeId string) string {
	return fmt.Sprintf("$BEEOS.HUB.%s.BEE.%s.SYNC_REQ", hubId, beeId)
}

type CommandMessage struct {
	Cmd    Command
	Params CommandParams
	Data   []byte
}
