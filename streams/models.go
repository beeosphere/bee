package streams

import "time"

type StreamConfig struct {
	Workflows map[string]*WorkflowConfig
	Instances []*WorkflowInstance
}

type WorkflowConfig struct {
	Id    string
	Name  string
	Nodes map[string]*NodeConfig
}

type NodeConfig struct {
	Id       string
	Name     string
	Class    string
	Subclass string
	Settings any
	Inputs   []Input
	Outputs  []Output
}

type Input struct {
	Id      string
	Type    string
	Enabled bool
}

type Output struct {
	Id      string
	Type    string
	Enabled bool
	Targets []OutputTarget
}

type OutputTarget struct {
	Node  string
	Input int
}

type WorkflowInstance struct {
	WorkflowId string
	Sources    map[string]string // Maps of signal to source
	Metadata   map[string]any
}

type InputValue struct {
	SourceId string
	Value    any
}

type OutputValue struct {
	SinkId string
	Value  any
}

type workflowContext struct {
	metadata  map[string]any
	timestamp time.Time
}

func newWorkflowContext(timestamp time.Time) WorkflowContext {
	wc := &workflowContext{
		metadata: make(map[string]any),
	}
	if timestamp.IsZero() {
		wc.timestamp = time.Now()
	} else {
		wc.timestamp = timestamp
	}
	return wc
}

func (wc *workflowContext) GetMetadata(key string) (any, bool) {
	return wc.metadata[key], wc.metadata != nil
}

func (wc *workflowContext) SetMetadata(key string, value any) {
	wc.metadata[key] = value
}
func (wc *workflowContext) GetTimestamp() time.Time {
	return wc.timestamp
}
