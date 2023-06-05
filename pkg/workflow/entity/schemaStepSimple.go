package entity

// WorkflowSchemaSimpleStep represents a simple linear workflow's schema step.
type WorkflowSchemaSimpleStep struct {
	name   WorkflowSchemaStepName
	topic  WorkflowSchemaStepTopic
	worker WorkflowSchemaStepWorker
}

var _ WorkflowSchemaStep = (*WorkflowSchemaSimpleStep)(nil)

func (w *WorkflowSchemaSimpleStep) Name() WorkflowSchemaStepName {
	return w.name
}

func (w *WorkflowSchemaSimpleStep) Topic() WorkflowSchemaStepTopic {
	return w.topic
}

func (w *WorkflowSchemaSimpleStep) Worker() WorkflowSchemaStepWorker {
	return w.worker
}

func NewWorkflowSchemaSimpleStep(
	sn WorkflowSchemaStepName, st WorkflowSchemaStepTopic, w WorkflowSchemaStepWorker) *WorkflowSchemaSimpleStep {
	return &WorkflowSchemaSimpleStep{
		name:   sn,
		topic:  st,
		worker: w,
	}
}
