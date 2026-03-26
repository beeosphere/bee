package streams

type WorkflowProcessor interface {
	Execute() error
}

type WorkflowResolver interface {
	Resolve(signals []string) (WorkflowProcessor, error)
}

type workflowResolver struct {
	instances []*WorkflowInstance
	mapper    map[string]*WorkflowInstance
}

func NewWorkflowResolver(instances []*WorkflowInstance) WorkflowResolver {
	return &workflowResolver{
		instances: instances,
	}
}

func (wr *workflowResolver) Resolve(signals []string) (WorkflowProcessor, error) {

	return nil, nil
}

// func (wr *WorkflowReference) Source(signal string) string {
