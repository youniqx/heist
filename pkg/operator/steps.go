package operator

type StepStatus string

const (
	// StepStatusSuccess indicates that a setup step was successful.
	StepStatusSuccess = "success"

	// StepStatusInProgress indicates that a setup step is in progress.
	StepStatusInProgress = "running"

	// StepStatusFailed indicates that a setup step has failed.
	StepStatusFailed = "failed"
)

type Step struct {
	Name   string
	Status StepStatus
}

type stepManager struct {
	CurrentStepName   string
	Channel           chan *Step
	CurrentStepStatus string
}

func (s *stepManager) NextStep(name string) {
	if s.Channel == nil {
		return
	}
	if s.CurrentStepStatus == StepStatusInProgress && s.CurrentStepName != "" {
		s.StepSuccess()
	}
	s.CurrentStepStatus = StepStatusInProgress
	s.CurrentStepName = name
	s.Channel <- &Step{
		Name:   name,
		Status: StepStatusInProgress,
	}
}

func (s *stepManager) StepSuccess() {
	if s.Channel == nil {
		return
	}
	s.CurrentStepStatus = StepStatusSuccess
	s.Channel <- &Step{
		Name:   s.CurrentStepName,
		Status: StepStatusSuccess,
	}
}

func (s *stepManager) StepFailed() {
	if s.Channel == nil {
		return
	}
	s.CurrentStepStatus = StepStatusFailed
	s.Channel <- &Step{
		Name:   s.CurrentStepName,
		Status: StepStatusFailed,
	}
}

func (s *stepManager) Complete() {
	if s.Channel == nil {
		return
	}
	if s.CurrentStepStatus == StepStatusInProgress && s.CurrentStepName != "" {
		s.StepSuccess()
	}
	close(s.Channel)
}
