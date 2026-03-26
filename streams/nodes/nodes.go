package nodes

import "github.com/beeosphere/bee/streams"

// ADD

func AddNodeProvider() streams.NodeProvider {
	return streams.NewNodeProvider(func(class, subclass string, config any) (streams.Node, error) {
		return &AddNode{}, nil
	})
}

type AddNode struct {
}

func (n *AddNode) Execute(inputs []any, context streams.WorkflowContext) (outputs []any, err error) {
	output := 0
	for _, input := range inputs {
		if val, ok := input.(int); ok {
			output += val
		} else {
			return nil, streams.ErrInputNotFound
		}
	}
	outputs = []any{output}
	return outputs, nil
}
