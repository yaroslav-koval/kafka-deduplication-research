package ferror_test

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/fhir/ferror"
	"net/http"
	"testing"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/edenlabllc/go-fhir-adapter/model/fhir/consts"
	"github.com/tj/assert"
)

func TestOutcomeFromError(t *testing.T) {
	ctx := context.Background()
	err := cerror.NewF(ctx, cerror.KindInternal, "err")
	sev := "error"
	expectedOO := fhir.OperationOutcome{
		Issue: []*fhir.OperationOutcomeIssue{{Severity: fhir.IssueSeverityPointerFromString(sev)}},
	}

	err = err.WithPayload(expectedOO)

	assert.Equal(t, &expectedOO, ferror.OutcomeFromError(err))

	err = err.WithPayload(&expectedOO)

	assert.Equal(t, &expectedOO, ferror.OutcomeFromError(err))

	err = err.WithPayload(map[string]string{"hello": "world"})
	expectedOO.Issue = []*fhir.OperationOutcomeIssue{{
		Severity:   fhir.IssueSeverityPointerFromString(consts.IssueSeverityError),
		Code:       fhir.IssueTypePointerFromString("exception"),
		Expression: []string{"hello"},
		Details:    &fhir.CodeableConcept{Text: "world"},
	}}

	assert.Equal(t, &expectedOO, ferror.OutcomeFromError(err))

	err = cerror.NewValidationError(ctx, map[string]string{"hello": "world"}).CError
	expectedOO.Issue[0].Code = fhir.IssueTypePointerFromString("business-rule")

	assert.Equal(t, &expectedOO, ferror.OutcomeFromError(err))

	err = cerror.NewF(ctx, cerror.KindInternal, "err")
	expectedOO.Issue = []*fhir.OperationOutcomeIssue{
		{
			Severity: fhir.IssueSeverityPointerFromString(consts.IssueSeverityError),
			Code:     fhir.IssueTypePointerFromString("exception"),
			Details:  &fhir.CodeableConcept{Text: "err"},
		},
	}

	assert.Equal(t, &expectedOO, ferror.OutcomeFromError(err))
	assert.Equal(t, &expectedOO, ferror.OutcomeFromError(err.Err()))
}

func TestOutcomeCodeFromHTTP(t *testing.T) {
	knownCodes := map[int]string{
		http.StatusNotFound:            consts.IssueTypeNotFound,
		http.StatusForbidden:           consts.IssueTypeForbidden,
		http.StatusGatewayTimeout:      consts.IssueTypeTimeout,
		http.StatusRequestTimeout:      consts.IssueTypeTimeout,
		http.StatusBadRequest:          consts.IssueTypeIncomplete,
		http.StatusUnprocessableEntity: consts.IssueTypeBusinessRule,
		http.StatusConflict:            consts.IssueTypeConflict,
		http.StatusInternalServerError: consts.IssueTypeException,
	}

	for k, v := range knownCodes {
		assert.Equal(t, v, ferror.OutcomeCodeFromHTTP(k))
	}

	for i := 100; i < 600; i++ {
		if _, ok := knownCodes[i]; ok {
			continue
		}

		assert.Equal(t, consts.IssueTypeUnknown, ferror.OutcomeCodeFromHTTP(i))
	}
}
