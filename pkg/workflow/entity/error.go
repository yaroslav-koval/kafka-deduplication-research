package entity

// ProcessingError represents an error that occured during workflo processing.
// It wraps an original error and gives an opportunity to set a flag
// that indicates wether a retry attempt should be performed for this error.
type ProcessingError struct {
	err   error
	retry bool
}

func NewProcessingError(err error) *ProcessingError {
	return &ProcessingError{err: err}
}

func (p *ProcessingError) Error() string {
	return p.err.Error()
}

func (p *ProcessingError) OriginalError() error {
	return p.err
}

func (p *ProcessingError) SetRetry(v bool) *ProcessingError {
	p.retry = v

	return p
}

func (p *ProcessingError) Retry() bool {
	return p.retry
}
