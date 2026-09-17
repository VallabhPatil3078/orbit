package runner

type MockReporter struct {
	SkippedTasks   map[string]bool
	SucceededTasks map[string]bool
	FailedTasks    map[string]bool
}

func NewMockReporter() *MockReporter {
	return &MockReporter{
		SkippedTasks:   make(map[string]bool),
		SucceededTasks: make(map[string]bool),
		FailedTasks:    make(map[string]bool),
	}
}

func (m *MockReporter) PipelineStarted()                               {}
func (m *MockReporter) TierStarted(tierIndex int, taskNames []string) {}
func (m *MockReporter) TaskStarted(task string)                       {}

func (m *MockReporter) TaskSkipped(task string, reason string) {
	m.SkippedTasks[task] = true
}

func (m *MockReporter) TaskSucceeded(task string) {
	m.SucceededTasks[task] = true
}

func (m *MockReporter) TaskFailed(task string, output string) {
	m.FailedTasks[task] = true
}

func (m *MockReporter) PipelineFinished(success bool) {}
