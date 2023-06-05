package entity

type ID string

func (i ID) String() string {
	return string(i)
}

func PointerID(s string) *ID {
	id := ID(s)
	return &id
}

type SearchWorkflowParams struct {
	ID     *ID             `form:"id" json:"id"`
	Status *WorkflowStatus `form:"status" json:"status"`
	*Paging
}

type SearchWorkflowResult struct {
	Workflows []*Workflow
	Paging    Paging
}

type Paging struct {
	Limit  int `form:"limit" json:"limit"`
	Offset int `form:"offset" json:"offset"`
}

// UpdateWorkflowForceParams is a model for workflow field values to update.
// Each workflow field corresponding to this model will be updated, even with nil value.
type UpdateWorkflowForceParams struct {
	Status    WorkflowStatus
	Error     *WorkflowErrorMsg
	ErrorKind *WorkflowErrorKind
}

// UpdateWorkflowNotNilParams is a model for workflow field values to update.
// Only fields that are not nil will be updated.
type UpdateWorkflowNotNilParams struct {
	Status    *WorkflowStatus
	Steps     []*WorkflowStep
	Error     *WorkflowErrorMsg
	ErrorKind *WorkflowErrorKind
}
