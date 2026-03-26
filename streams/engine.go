package streams

type streamEngine struct {
	provider ValuesProvider
}

func NewStreamEngine(opts ...Option) *streamEngine {
	eng := &streamEngine{}
	for _, opt := range opts {
		opt(eng)
	}
	return eng
}

type Option func(*streamEngine)

func WithValuesProvider(provider ValuesProvider) Option {
	return func(s *streamEngine) {
		s.provider = provider
	}
}

// Implement the StreamEngine interface
func (se *streamEngine) Initialize(provider ValuesProvider) error {
	// Initialize the stream engine with the provided values provider
	return nil
}
func (se *streamEngine) Finalize() error {
	// Finalize the stream engine
	return nil
}
func (se *streamEngine) Configure(config *StreamConfig) error {
	// Configure the stream engine with the provided configuration
	return nil
}
func (se *streamEngine) ExecuteWorkflows(signals []string) error {
	// Execute the workflows based on the provided signals
	return nil
}
func (se *streamEngine) ExecuteWorkflow(workflowId string, values []InputValue, context WorkflowContext) error {
	// Execute a specific workflow with the provided values and context
	return nil
}
