package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"strconv"
	"strings"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/edenlabllc/go-fhir-adapter/model/fhir/consts"
	uuid "github.com/satori/go.uuid"
)

type BaseResourceModel struct {
	ResourceType string `json:"resourceType"`
}

// Resource is a basic FHIR resource interface.
// Each FHIR resource has some base fields.
type Resource interface {
	ResourceName() string
	ResourceMeta() *fhir.Meta
	ResourceID() string
}

// CodingFromCodeableConceptSafe returns coding with a given index.
// If CodeableConcept is nil or coding with the given index doesn't exist
// returns a default struct.
func CodingFromCodeableConceptSafe(c *fhir.CodeableConcept, i int) fhir.Coding {
	var coding fhir.Coding
	if c == nil {
		return coding
	}

	if i > len(c.Coding)-1 {
		return coding
	}

	return *c.Coding[i]
}

// EscapeFHIRSearchCharacters escapes specific characters that FHIR use in search
// https://www.hl7.org/fhir/search.html#escaping
func EscapeFHIRSearchCharacters(name string) string {
	specialChars := []string{",", "$", "|"}

	for i := range specialChars {
		name = strings.ReplaceAll(name, specialChars[i], fmt.Sprintf(`\%s`, specialChars[i]))
	}

	return name
}

// GenFieldPath creates error path for cerror.NewValidationError method
// if pathParts receives string value it concatenates with `.value`
// if pathParts receives int value it concatenates with `[value]`
// GetFieldPath("Bundle.entry", 1, "request", 2, "value") --> "Bundle.Entry[1].request[2].value"
func GenFieldPath(pathParts ...interface{}) string {
	var errPath strings.Builder

	for _, path := range pathParts {
		switch p := path.(type) {
		case string:
			errPath.WriteString(fmt.Sprintf(".%s", p))
		case int:
			errPath.WriteString(fmt.Sprintf("[%d]", p))
		}
	}

	return strings.TrimLeft(errPath.String(), ".")
}

// ValidateResourceProfile checks that resource has a profile that is in a given
// list of allowed profiles.
func ValidateResourceProfile(ctx context.Context, res Resource, allowedProfiles []string) error {
	return ValidateResourceProfileWithPath(ctx, res, allowedProfiles, res.ResourceName())
}

func ValidateResourceProfileWithPath(ctx context.Context, res Resource, allowedProfiles []string, attributePath string) error {
	meta := res.ResourceMeta()

	if meta == nil {
		return cerror.NewValidationError(ctx, map[string]string{
			GenFieldPath(attributePath, consts.FieldNameMeta): ErrTextMustNotBeEmpty}).LogError()
	}

	if len(meta.Profile) != 1 {
		return cerror.NewValidationError(ctx, map[string]string{
			GenFieldPath(attributePath, consts.FieldNameMeta, consts.FieldNameProfile): ErrTextMustHaveExactlyOneItem}).LogError()
	}

	if len(allowedProfiles) == 0 {
		return nil
	}

	actualP := meta.Profile[0]
	for _, p := range allowedProfiles {
		if p == actualP {
			return nil
		}
	}

	return cerror.NewValidationError(ctx, map[string]string{
		GenFieldPath(attributePath, consts.FieldNameMeta, consts.FieldNameProfile, 0): ErrTextNotSupportedProfile}).LogError()
}

// ExtractTelecom extracts a telecom object by ContactPointSystemType
// https://www.hl7.org/fhir/valueset-contact-point-system.html
func ExtractTelecom(t []*fhir.ContactPoint, system string) (*fhir.ContactPoint, bool) {
	for _, telecom := range t {
		if telecom.System != nil && telecom.System.String() == system {
			return telecom, true
		}
	}

	return nil, false
}

// NewTaskInputReference creates TaskInput with specified resource name and ID.
func NewTaskInputReference(resName, resID string) *fhir.TaskInput {
	return &fhir.TaskInput{
		Type: &fhir.CodeableConcept{
			Coding: []*fhir.Coding{
				{
					Code:   resName,
					System: CodingSystemResourceTypes,
				},
			},
		},
		ValueReference: &fhir.Reference{
			Reference: fmt.Sprintf("%s/%s", resName, resID),
		},
	}
}

// NewTaskOutputReference creates TaskOutput with specified resource name and ID.
func NewTaskOutputReference(resName, resID string) *fhir.TaskOutput {
	return &fhir.TaskOutput{
		Type: &fhir.CodeableConcept{
			Coding: []*fhir.Coding{
				{
					Code:   resName,
					System: CodingSystemResourceTypes,
				},
			},
		},
		ValueReference: &fhir.Reference{
			Reference: fmt.Sprintf("%s/%s", resName, resID),
		},
	}
}

// NewBundleEntryPost creates BundleEntry with URL method POST via marshalling resource with specified resource name.
func NewBundleEntryPost(resource interface{}, resourceName string) *fhir.BundleEntry {
	//nolint:errchkjson
	resByte, _ := json.Marshal(resource)

	return &fhir.BundleEntry{
		Resource: resByte,
		Request: &fhir.BundleEntryRequest{
			Method: fhir.HTTPVerbPointerFromString(http.MethodPost),
			URL:    resourceName,
		},
	}
}

// NewBundleEntryPut creates BundleEntry with URL method PUT via marshalling resource with specified resource name.
func NewBundleEntryPut(resource interface{}, resourceName, resID string) *fhir.BundleEntry {
	//nolint:errchkjson
	resByte, _ := json.Marshal(resource)

	return &fhir.BundleEntry{
		Resource: resByte,
		Request: &fhir.BundleEntryRequest{
			Method: fhir.HTTPVerbPointerFromString(http.MethodPut),
			URL:    fmt.Sprintf("%s/%s", resourceName, resID),
		},
	}
}

// NewBundleTransaction creates Bundle with random ID, type 'transaction' and specified entries.
func NewBundleTransaction(entries ...*fhir.BundleEntry) *fhir.Bundle {
	return &fhir.Bundle{
		ID:    uuid.NewV4().String(),
		Type:  fhir.BundleTypePointerFromString(consts.BundleTypeTransaction),
		Entry: entries,
	}
}

// NewBundleCollection creates Bundle with random ID, type 'collection' and specified entries.
func NewBundleCollection(entries ...*fhir.BundleEntry) *fhir.Bundle {
	return &fhir.Bundle{
		ID:    uuid.NewV4().String(),
		Type:  fhir.BundleTypePointerFromString(consts.BundleTypeCollection),
		Entry: entries,
	}
}

// ValidateNationalIDValue validates input value
// it must be with length 10 and starts with digits 1 or 2
func ValidateNationalIDValue(value string) bool {
	validLen := 10
	value = strings.ReplaceAll(value, " ", "")

	if _, err := strconv.Atoi(value); err != nil {
		return false
	}

	if len(value) != validLen {
		return false
	}

	idType, _ := strconv.Atoi(value[0:1])

	if idType != 1 && idType != 2 {
		return false
	}

	idArr := make([]int, len(value))
	for c := 0; c < 10; c++ {
		idArr[c], _ = strconv.Atoi(value[c : c+1])
	}

	sum := 0

	for c := 0; c < validLen; c++ {
		if c%2 == 0 {
			dd := fmt.Sprintf("%02d", idArr[c]*2) //nolint:gomnd
			fvalue, _ := strconv.Atoi(dd[0:1])
			svalue, _ := strconv.Atoi(dd[1:2])
			sum += fvalue + svalue
		} else {
			sum += idArr[c]
		}
	}

	return sum%10 == 0
}

func NewOperationOutcome(severity fhir.IssueSeverity, code fhir.IssueType, details string) *fhir.OperationOutcome {
	return &fhir.OperationOutcome{
		Issue: []*fhir.OperationOutcomeIssue{
			{
				Severity: &severity,
				Code:     &code,
				Details:  &fhir.CodeableConcept{Text: details},
			},
		},
	}
}
