package cerror

type KindElastic uint8

const (
	_ KindElastic = iota
	KindElasticOther
	KindElasticPermission
	KindElasticIO

	_kindElasticOther      = "other"
	_kindElasticPermission = "permission"
	_kindElasticIO         = "io"
)

var (
	_codesElasticPermission = []string{
		"Access Denied",
	}

	_codesElasticIO = []string{
		"failed to compress request body",
		"failed to compress request body (during close)",
		"connection refused",
	}
)

func (k KindElastic) String() string {
	switch k {
	case KindElasticPermission:
		return _kindElasticPermission
	case KindElasticIO:
		return _kindElasticIO
	default:
		return _kindElasticOther
	}
}

func KindFromElastic(err error) KindElastic {
	switch {
	case isStrContainsInsensitive(err.Error(), _codesElasticPermission):
		return KindElasticPermission
	case isStrContainsInsensitive(err.Error(), _codesElasticIO):
		return KindElasticIO
	}

	return KindElasticOther
}
