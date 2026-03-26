package streams

import (
	"errors"
)

// type NodeWrapper interface {
// 	Reset()
// 	Inject(input int, value any) error
// }

type nodeWrapper struct {
	mask        uint64
	execMask    uint64
	inputValues []any
	context     WorkflowContext
	node        Node
	workflow    *workflow
	config      *NodeConfig
}

func newNodeWrapper(nodeId string, workflow *workflow) (*nodeWrapper, error) {
	config, ok := workflow.config.Nodes[nodeId]
	if !ok {
		return nil, ErrNodeNotFound(nodeId)
	}
	provider, err := registry.Find(config.Class, config.Subclass)
	if err != nil {
		return nil, err
	}
	node, err := provider.ProvideNode(config.Class, config.Subclass, config.Settings)
	if err != nil {
		return nil, err
	}
	// Fill the mask: 3 inputs -> ...0000111
	// 3 inputs with 2nd input disabled -> ...0000101
	execMask := uint64(0)
	for idx, input := range config.Inputs {
		if input.Enabled {
			execMask |= (1 << idx)
		}
	}
	// Create the node wrapper
	wrapper := &nodeWrapper{
		workflow:    workflow,
		config:      config,
		node:        node,
		execMask:    execMask,
		mask:        uint64(0),
		inputValues: make([]any, len(config.Inputs)),
		context:     nil,
	}
	return wrapper, nil
}

func (n *nodeWrapper) Reset() {
	n.mask = 0
	n.context = nil
	// TODO: Evaluate if the input values should be reset to their default values or is it ok to leave them as is?
}

func (n *nodeWrapper) SetContext(context WorkflowContext) {
	n.context = context
}

func (n *nodeWrapper) Inject(input int, value any) error {
	// Store the input value
	if input < 0 || input >= len(n.inputValues) {
		return errors.New("input index out of range")
	}
	n.inputValues[input] = value

	// Update the inputs mask and check if all inputs are set
	n.mask |= (1 << input)
	if n.execMask&n.mask != n.execMask {
		return nil
	}
	// Execute the node with the inputs
	outputValues, err := n.node.Execute(n.inputValues, n.context)
	if err != nil {
		return err
	}
	// Execute the next nodes using the output values
	for idx, output := range n.config.Outputs {
		if output.Enabled {
			// Iterate over the output targets
			for _, target := range output.Targets {
				// Find the next node
				nextNode, ok := n.workflow.nodes[target.Node]
				if !ok {
					return ErrNodeNotFound(target.Node)
				}
				// Pass the context to the next node
				nextNode.SetContext(n.context)
				// Inject the output value into the next node
				err = nextNode.Inject(target.Input, outputValues[idx])
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
