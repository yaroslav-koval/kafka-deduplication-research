package ferror

import (
	"kafka-polygon/pkg/cerror"
	"net/http"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/edenlabllc/go-fhir-adapter/model/fhir/consts"
)

// OutcomeFromError transforms error to FHIR standard error
// described by OperationOutcome FHIR resource.
func OutcomeFromError(err error) *fhir.OperationOutcome {
	oo := &fhir.OperationOutcome{}
	httpCode := cerror.ErrKind(err).HTTPCode()
	textCode := OutcomeCodeFromHTTP(httpCode)

	var (
		severity fhir.IssueSeverity
		code     fhir.IssueType
	)

	if pErr, ok := err.(cerror.PayloadError); ok {
		switch p := pErr.Payload().(type) {
		case *fhir.OperationOutcome:
			return p
		case fhir.OperationOutcome:
			return &p
		case map[string]string:
			for k, v := range p {
				severity = consts.IssueSeverityError
				code = fhir.IssueType(textCode)

				oo.Issue = append(oo.Issue, &fhir.OperationOutcomeIssue{
					Severity:   &severity,
					Code:       &code,
					Expression: []string{k},
					Details:    &fhir.CodeableConcept{Text: v},
				})
			}

			return oo
		}
	}

	if mErr, ok := err.(*cerror.MultiError); ok {
		for _, err := range mErr.Errors() {
			opOut := OutcomeFromError(err)
			oo.Issue = append(oo.Issue, opOut.Issue...)
		}

		return oo
	}

	severity = consts.IssueSeverityError
	code = fhir.IssueType(textCode)
	oo.Issue = []*fhir.OperationOutcomeIssue{
		{
			Severity: &severity,
			Code:     &code,
			Details:  &fhir.CodeableConcept{Text: err.Error()},
		},
	}

	return oo
}

// OutcomeCodeFromHTTP transforms http status code to human readable code
func OutcomeCodeFromHTTP(code int) string {
	switch code {
	case http.StatusNotFound:
		return consts.IssueTypeNotFound
	case http.StatusForbidden:
		return consts.IssueTypeForbidden
	case http.StatusConflict:
		return consts.IssueTypeConflict
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return consts.IssueTypeTimeout
	case http.StatusBadRequest:
		return consts.IssueTypeIncomplete
	case http.StatusUnprocessableEntity:
		return consts.IssueTypeBusinessRule
	case http.StatusInternalServerError:
		return consts.IssueTypeException
	}

	return consts.IssueTypeUnknown
}
