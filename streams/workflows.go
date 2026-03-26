package streams

import "time"

// workflow represents a single workflow instance.
// It contains the logic to execute the workflow based on the provided signals.
// The workflow can be configured with different parameters and can be reset to its initial state.
// The workflow is responsible for managing its own state and execution flow.
// It can be reused for different executions, allowing for efficient resource management.
// The workflow can be built using a builder pattern, allowing for flexible configuration and instantiation.
// The workflow can also be pooled for reuse, reducing the overhead of creating new instances.
// The workflow is designed to be lightweight and efficient, making it suitable for high-performance applications.

type workflow struct {
	config *WorkflowConfig
	nodes  map[string]*nodeWrapper
}

// type source struct {
// 	id      string
// 	value   any
// 	targets []OutputTarget
// }

func (w *workflow) execute(inputs []InputValue, context WorkflowContext) error {
	if context == nil {
		context = newWorkflowContext(time.Now())
	}
	for _, in := range inputs {
		// Find the source nodes by their IDs
		node, ok := w.nodes[in.SourceId]
		if !ok {
			return ErrNodeNotFound(in.SourceId)
		}
		node.SetContext(context)
		if err := node.Inject(0, in.Value); err != nil {
			return err
		}
	}
	return nil
}

func (w *workflow) reset() {
	for _, node := range w.nodes {
		node.Reset()
	}
}

// WorkflowBuilder is responsible for building workflows based on the provided configuration.
// It allows for flexible configuration and instantiation of workflows.
// The builder pattern is used to create workflows with different parameters.

type workflowBuilder struct {
}

func (wb *workflowBuilder) buildWorkflow(config *WorkflowConfig) (*workflow, error) {
	wf := &workflow{
		config: config,
		nodes:  make(map[string]*nodeWrapper),
	}
	for id := range config.Nodes {

		node, err := newNodeWrapper(id, wf)
		if err != nil {
			return nil, err
		}
		wf.nodes[id] = node
	}
	return wf, nil
}

// WorkflowPool is responsible for managing a pool of workflow instances.
// It allows for efficient reuse of workflow instances, reducing the overhead of creating new instances.
// Contains a pool of workflow instances that can be reused for different executions.

type workflowPool struct {
}

func (wp *workflowPool) getWorkflow() (*workflow, error) {
	return &workflow{}, nil
}

func (wp *workflowPool) recycleWorkflow(wf *workflow) error {
	return nil
}
