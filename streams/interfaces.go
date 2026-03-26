package streams

import "time"

type StreamEngine interface {
	Initialize() error
	Finalize() error

	Configure(config *StreamConfig) error

	ExecuteWorkflows(signals []string) error
	ExecuteWorkflow(workflowId string, inputs []InputValue, context WorkflowContext) (outputs []OutputValue, err error)
}

type ValuesProvider interface {
	GetValue(id string) (any, error)
}

type WorkflowContext interface {
	GetMetadata(key string) (any, bool)
	SetMetadata(key string, value any)
	GetTimestamp() time.Time
}

type Node interface {
	Execute(inputs []any, context WorkflowContext) (outputs []any, err error)
}

// type NodeFactory interface {
// 	CreateNode(class, subclass string) (Node, error)
// 	RegisterProvider(class, subclass string, node Node) error // Register Node instance provider (NodeProvider interface)
// }

type NodeProviderFunc func(class, subclass string, config any) (Node, error)

type NodeProvider interface {
	ProvideNode(class, subclass string, config any) (Node, error)
}

type NodeProviderBase struct {
	fn NodeProviderFunc
}

func NewNodeProvider(fn NodeProviderFunc) NodeProvider {
	return &NodeProviderBase{
		fn: fn,
	}
}

func (pb *NodeProviderBase) ProvideNode(class, subclass string, config any) (Node, error) {
	return pb.fn(class, subclass, config)
}

type NodeProviderRegistry interface {
	Register(class, subclass string, provider NodeProvider) error
	Find(class, subclass string) (provider NodeProvider, err error)
}
