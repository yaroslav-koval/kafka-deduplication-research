package test

import (
	"fmt"
	pkgFHIR "kafka-polygon/pkg/fhir"
	"net/http"
	"testing"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/edenlabllc/go-fhir-adapter/model/fhir/consts"
	"github.com/tj/assert"
)

func TestCodingFromCodeableConceptSafe(t *testing.T) {
	type args struct {
		c *fhir.CodeableConcept
		i int
	}

	tests := []struct {
		name string
		args args
		want fhir.Coding
	}{
		{
			name: "checks if CodeableConcept is nil",
			args: args{c: nil, i: 0},
			want: fhir.Coding{},
		},
		{
			name: "checks if CodeableConcept is not nil and takes the first coding element",
			args: args{c: &fhir.CodeableConcept{
				Coding: []*fhir.Coding{
					{
						System:  "system1",
						Code:    "code1",
						Display: "display1",
					},
					{
						System:  "system2",
						Code:    "code2",
						Display: "display2",
					},
				},
			}, i: 0},
			want: fhir.Coding{
				System:  "system1",
				Code:    "code1",
				Display: "display1",
			},
		},
		{
			name: "checks if CodeableConcept is not nil and takes the second coding element",
			args: args{c: &fhir.CodeableConcept{
				Coding: []*fhir.Coding{
					{
						System:  "system1",
						Code:    "code1",
						Display: "display1",
					},
					{
						System:  "system2",
						Code:    "code2",
						Display: "display2",
					},
				},
			}, i: 1},
			want: fhir.Coding{
				System:  "system2",
				Code:    "code2",
				Display: "display2",
			},
		},
		{
			name: "checks if CodeableConcept is not nil but takes out of range coding element",
			args: args{c: &fhir.CodeableConcept{
				Coding: []*fhir.Coding{
					{
						System:  "system1",
						Code:    "code1",
						Display: "display1",
					},
				},
			}, i: 10},
			want: fhir.Coding{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkgFHIR.CodingFromCodeableConceptSafe(tt.args.c, tt.args.i)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscapeFHIRSearchCharacters(t *testing.T) {
	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "checks string without specific character",
			args: args{"AMOXICILLIN CLAVULANIC ACID 1.2g"},
			want: "AMOXICILLIN CLAVULANIC ACID 1.2g",
		},
		{
			name: "checks string wit specific character",
			args: args{"some text, that contains $ specific | character for FHIR search"},
			want: `some text\, that contains \$ specific \| character for FHIR search`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkgFHIR.EscapeFHIRSearchCharacters(tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenFieldPath(t *testing.T) {
	actualResult := pkgFHIR.GenFieldPath(pkgFHIR.PathBundleEntry, 1, consts.FieldNameRequest, consts.FieldNameName)
	expectedResult := fmt.Sprintf("%s[%d].%s.%s", pkgFHIR.PathBundleEntry, 1, consts.FieldNameRequest, consts.FieldNameName)

	assert.Equal(t, expectedResult, actualResult)
}

func TestValidateResourceProfileWithPath(t *testing.T) {
	t.Parallel()

	b := &fhir.Bundle{}
	path := "path.to.attributes"

	// meta shouldn't be nil
	err := pkgFHIR.ValidateResourceProfileWithPath(_cb, b, []string{}, path)
	assertValidationError(t, err, path+".meta", pkgFHIR.ErrTextMustNotBeEmpty)

	// profile should be filled
	meta := &fhir.Meta{}
	b.Meta = meta
	err = pkgFHIR.ValidateResourceProfileWithPath(_cb, b, []string{}, path)
	assertValidationError(t, err, path+".meta.profile", pkgFHIR.ErrTextMustHaveExactlyOneItem)

	// if allowed profiles is not passed and profile exists: validation is success
	meta.Profile = []string{""}
	err = pkgFHIR.ValidateResourceProfileWithPath(_cb, b, []string{}, path)
	assertEmptyError(t, err)

	// meta profile should be in expected
	meta.Profile = []string{"profile"}
	err = pkgFHIR.ValidateResourceProfileWithPath(_cb, b, []string{""}, path)
	assertValidationError(t, err, path+".meta.profile[0]", pkgFHIR.ErrTextNotSupportedProfile)

	meta.Profile = []string{""}
	err = pkgFHIR.ValidateResourceProfileWithPath(_cb, b, []string{"profile"}, path)
	assertValidationError(t, err, path+".meta.profile[0]", pkgFHIR.ErrTextNotSupportedProfile)

	// success
	meta.Profile = []string{"profile"}
	err = pkgFHIR.ValidateResourceProfileWithPath(_cb, b, []string{"profile"}, path)
	assertEmptyError(t, err)
}

func TestValidateResourceProfile(t *testing.T) {
	t.Parallel()

	b := &fhir.Bundle{}

	// check that ValidateResourceProfileWithPath is called with b.ResourceName()
	// more details tests are in TestValidateResourceProfile
	err := pkgFHIR.ValidateResourceProfile(_cb, b, []string{})
	assertValidationError(t, err, b.ResourceName()+".meta", pkgFHIR.ErrTextMustNotBeEmpty)
}

func TestExtractTelecom(t *testing.T) {
	telecom := []*fhir.ContactPoint{
		{
			System: fhir.ContactPointSystemPointerFromString(consts.ContactPointSystemEmail),
			Value:  "email@email.com",
		},
	}

	res, ok := pkgFHIR.ExtractTelecom(telecom, consts.ContactPointSystemEmail)
	assert.True(t, ok)
	assert.Equal(t, telecom[0].Value, res.Value)

	res, ok = pkgFHIR.ExtractTelecom(telecom, consts.ContactPointSystemSms)
	assert.Nil(t, res)
	assert.False(t, ok)
}

func TestNewTaskInputReference(t *testing.T) {
	t.Parallel()

	resName := "resName"
	resID := "resID"

	expTi := &fhir.TaskInput{
		Type: &fhir.CodeableConcept{
			Coding: []*fhir.Coding{
				{
					Code:   resName,
					System: pkgFHIR.CodingSystemResourceTypes,
				},
			},
		},
		ValueReference: &fhir.Reference{
			Reference: fmt.Sprintf("%s/%s", resName, resID),
		},
	}

	ti := pkgFHIR.NewTaskInputReference(resName, resID)
	assert.EqualValues(t, expTi, ti)
}

func TestNewTaskOutputReference(t *testing.T) {
	t.Parallel()

	resName := "resName"
	resID := "resID"

	expTo := &fhir.TaskOutput{
		Type: &fhir.CodeableConcept{
			Coding: []*fhir.Coding{
				{
					Code:   resName,
					System: pkgFHIR.CodingSystemResourceTypes,
				},
			},
		},
		ValueReference: &fhir.Reference{
			Reference: fmt.Sprintf("%s/%s", resName, resID),
		},
	}

	to := pkgFHIR.NewTaskOutputReference(resName, resID)
	assert.EqualValues(t, expTo, to)
}

func TestNewBundleEntryPost(t *testing.T) {
	t.Parallel()

	res := &fhir.Medication{ID: "123"}
	resByte, _ := res.MarshalJSON()

	expBe := &fhir.BundleEntry{
		Resource: resByte,
		Request: &fhir.BundleEntryRequest{
			Method: fhir.HTTPVerbPointerFromString(http.MethodPost),
			URL:    res.ResourceName(),
		},
	}

	be := pkgFHIR.NewBundleEntryPost(res, consts.ResourceMedication)
	assert.EqualValues(t, expBe, be)
}

func TestNewBundleEntryPut(t *testing.T) {
	t.Parallel()

	res := &fhir.Medication{ID: "123"}
	resByte, _ := res.MarshalJSON()

	expBe := &fhir.BundleEntry{
		Resource: resByte,
		Request: &fhir.BundleEntryRequest{
			Method: fhir.HTTPVerbPointerFromString(http.MethodPut),
			URL:    res.ResourceName() + "/" + "123",
		},
	}

	be := pkgFHIR.NewBundleEntryPut(res, consts.ResourceMedication, res.ID)
	assert.EqualValues(t, expBe, be)
}

func TestNewBundleTransaction(t *testing.T) {
	t.Parallel()

	res := &fhir.Medication{ID: "123"}
	resByte, _ := res.MarshalJSON()

	expBe := &fhir.BundleEntry{
		Resource: resByte,
		Request: &fhir.BundleEntryRequest{
			Method: fhir.HTTPVerbPointerFromString(http.MethodPost),
			URL:    res.ResourceName(),
		},
	}

	b := pkgFHIR.NewBundleTransaction(
		pkgFHIR.NewBundleEntryPost(res, res.ResourceName()))
	assert.NotEqual(t, "", b.ID)
	assert.Equal(t, consts.BundleTypeTransaction, b.Type.String())
	assert.EqualValues(t, expBe, b.Entry[0])
}

func TestValidateNationalIDValue(t *testing.T) {
	t.Parallel()

	for _, i := range []struct {
		value   string
		isValid bool
	}{
		{"1058529940", true},
		{"1018521940", false},
		{"1001244019", true},
		{"2134952183", false},
		{"2986457121", true},
	} {
		isValidActual := pkgFHIR.ValidateNationalIDValue(i.value)

		assert.Equal(t, i.isValid, isValidActual)
	}
}
