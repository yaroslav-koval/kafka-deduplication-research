package fhir

import (
	"time"
)

// Identifier
const (
	IdentNationalID                   = "NI"
	IdentPassport                     = "PPN"
	IdentBirthCertificate             = "BCT"
	IdentCitizenshipCard              = "CZ"
	IdentPermanentResidentCardNumber  = "PRC"
	IdentBorderNumber                 = "BN"
	IdentDisplacedPerson              = "DP"
	IdentGulfCooperationCouncilNumber = "GCC"
	IdentJurisdictionalHealthNumber   = "JHN"
	IdentVisa                         = "VS"
)

// Format
const (
	FormatDateTime = time.RFC3339
	FormatInstant  = time.RFC3339Nano
	FormatDate     = "2006-01-02"
	FormatTime     = "15:04:05"
)

const (
	NationalityCodeSA = "SA"
)

// coding system
const (
	CodingSystemResourceTypes = "http://hl7.org/fhir/resource-types"
)

// parameter names
const (
	ParamNameRelatedPerson      = "relatedPerson"
	ParamNameDocumentReference  = "documentReference"
	ParamNamePatient            = "patient"
	ParamNameOrganization       = "organization"
	ParamNamePractitioner       = "practitioner"
	ParamNamePractitionerRole   = "practitionerRole"
	ParamNameLocation           = "location"
	ParamNameReviewer           = "reviewer"
	ParamNameReasonCode         = "reasonCode"
	ParamNameMedicationDispense = "medicationDispense"
	ParamNameConfirmationMethod = "confirmationMethod"
	ParamNameMedicationRequest  = "medicationRequest"
	ParamNameComposition        = "composition"
	ParamNameTaskID             = "task_id"
	ParamNameInvoiceReference   = "invoice_reference"
	ParamNameOTP                = "otp"
	ParamNameStatusReason       = "statusReason"
)

// media storage
const (
	StorageTypeFile = "file"
)

// paths
const (
	PathParameter   = "Parameters.parameter"
	PathBundleEntry = "Bundle.entry"
)

// common error messages
const (
	ErrTextWrongReference                   = "wrong reference"
	ErrTextMustContainX                     = "must contain %q"
	ErrTextMustBeEmpty                      = "must be empty"
	ErrTextMustBeX                          = "must be %q"
	ErrTextMustNotBeEmpty                   = "must not be empty"
	ErrTextMustHaveXValue                   = "must have %q value"
	ErrTextMustHaveXParam                   = "must have %q param"
	ErrTextSuchXAlreadyExists               = "such %s already exists"
	ErrTextNotSupportedProfile              = "profile is not supported"
	ErrTextMustBeReferenceToX               = "must be reference to %q"
	ErrTextSuchXDoesNotExists               = "such %s does not exist"
	ErrTextXDoesNotExists                   = "%s does not exist"
	ErrTextSuchXIsNotActive                 = "such %s is not active"
	ErrTextXIsNotActive                     = "%s is not active"
	ErrTextParameterXIsEmpty                = "parameter %q is empty"
	ErrTextXIsEmpty                         = "%s is empty"
	ErrTextMoreThanXResourcesWithIdentifier = "found more than %d %#q with such identifier"
	ErrTextInvalidXFormat                   = "invalid %q format"
	ErrTextDifferentXIDAndURL               = "url id and %s id must be equal"
	ErrTextMustHaveExactlyOneItem           = "must have exactly 1 item"
	ErrTextMustHaveExactlyXItems            = "must have exactly %d items"
	ErrTextMustHaveMoreThanXItems           = "must have %d or more items"
	ErrTextMustNotBeEqualToReferencedX      = "must not be equal to the referenced %s"
	ErrTextMustBeActive                     = "must be active"
	ErrTextMustNotBeDeceased                = "must not be deceased"
	ErrTextAgeMustNotBeLessThanX            = "age must not be less than %d"
	ErrTextParametersForSuchXDoNotExist     = "parameters for such %s do not exist"
	ErrTextCouldNotParseXInX                = "couldn't parse %s in %s"
	ErrTextXNotFound                        = "%s not found"
	ErrTextXPeriodsCrossEachOther           = "%s periods cross each other"
	ErrTextNotValidPeriodRange              = "start of period must be before end of period"
	ErrTextGivenXIsNotValid                 = "given %s is not valid"
	ErrTextGivenXIsNotCorrect               = "%s is not correct"
	ErrTextReferencedXMustContainX          = "referenced %s must contain %s"
)

const (
	OpCreate  = "create"
	OpUpdate  = "update"
	OpReplace = "replace"
)

const (
	HumanNameValueCodeEn = "en"
	HumanNameValueCodeAr = "ar"
)

const (
	HeaderXAccessToken = "X-Access-Token"
)
