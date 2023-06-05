package cerror

import "net/http"

type ElasticKind uint8

const (
	_ ElasticKind = iota
	KindElasticOther
	KindElasticPermission // KindPermission denied
	KindElasticIO         // External I/O error such as network failure

	elasticUnknown          = "elastic_unknown_error"
	elasticOther            = "elastic_other_error" // king message
	elasticPermissionDenied = "elastic_permission_denied"
	elasticIO               = "elastic_io_error"
)

var (
	_elasticPermissionCodes = []string{
		"Access Denied",
	}

	_elasticIOCodes = []string{
		"failed to compress request body",
		"failed to compress request body (during close)",
		"connection refused",
	}
)

func (k ElasticKind) String() string {
	switch k {
	case KindElasticPermission:
		return elasticPermissionDenied
	case KindElasticIO:
		return elasticIO
	case KindElasticOther:
		return elasticOther
	}

	return elasticUnknown
}

func (k ElasticKind) HTTPCode() int {
	return http.StatusInternalServerError
}

func (k ElasticKind) Group() KindGroup {
	return GroupElastic
}

func ElasticToKind(err error) ElasticKind {
	switch {
	case checkCode(err.Error(), _elasticPermissionCodes):
		return KindElasticPermission
	case checkCode(err.Error(), _elasticIOCodes):
		return KindElasticIO
	default:
		return KindElasticOther
	}
}
