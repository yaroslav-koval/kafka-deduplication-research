package cerror

type Kind uint8

const (
	_                    Kind = iota
	KindOther                 // Unclassified error
	KindValidation            // Validation was not passed
	KindUnauthenticated       // Client's identity is unknown
	KindForbidden             // Client's identity is known, but the client doesn't have access rights to the content
	KindExists                // Item already exists
	KindNotExists             // Item does not exist
	KindConflict              // State or logic conflict
	KindSyntax                // Data has invalid syntax
	KindExternalPkg           // Unclassified/unexpected error from external pkg
	KindIO                    // Unclassified I/O error
	KindTimeout               // Operation has timed out
	KindExternalAPI           // External API returned an error or unsuccessful status
	KindInternalAPI           // Internal API returned an error or unsuccessful status
	KindExternalResource      // Error during communication with external resource(e.g. db)
	KindInternalResource      // Error during communication with external resource(e.g. db)

	_kindOther            = "other"
	_kindValidation       = "validation"
	_kindUnauthenticated  = "unauthenticated"
	_kindForbidden        = "forbidden"
	_kindExists           = "exists"
	_kindNotExists        = "not_exists"
	_kindConflict         = "conflict"
	_kindSyntax           = "syntax"
	_kindExternalPkg      = "external_pkg"
	_kindIO               = "io"
	_kindTimeout          = "timeout"
	_kindExternalAPI      = "external_api"
	_kindInternalAPI      = "internal_api"
	_kindExternalResource = "external_resource"
	_kindInternalResource = "internal_resource"
)

func (k Kind) String() string {
	switch k {
	case KindValidation:
		return _kindValidation
	case KindUnauthenticated:
		return _kindUnauthenticated
	case KindForbidden:
		return _kindForbidden
	case KindExists:
		return _kindExists
	case KindNotExists:
		return _kindNotExists
	case KindConflict:
		return _kindConflict
	case KindSyntax:
		return _kindSyntax
	case KindExternalPkg:
		return _kindExternalPkg
	case KindIO:
		return _kindIO
	case KindTimeout:
		return _kindTimeout
	case KindExternalAPI:
		return _kindExternalAPI
	case KindInternalAPI:
		return _kindInternalAPI
	case KindExternalResource:
		return _kindExternalResource
	case KindInternalResource:
		return _kindInternalResource
	default:
		return _kindOther
	}
}
