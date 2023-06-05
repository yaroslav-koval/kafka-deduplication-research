package cerror

type KindMinio uint8

const (
	_ KindMinio = iota
	KindMinioOther
	KindMinioIO
	KindMinioNotExists
	KindMinioConflict
	KindMinioPermission
	KindMinioPreconditionFailed

	_kindMinioOther              = "other"
	_kindMinioPermission         = "permission"
	_kindMinioPreconditionFailed = "precondition_failed"
	_kindMinioNotExists          = "not_exists"
	_kindMinioConflict           = "conflict"
	_kindMinioIO                 = "io"
)

var (
	_codesMinioIO = []string{
		"connection refused",
	}

	_codesMinioNotExists = []string{
		"The specified bucket does not exist",
		"The specified key does not exist",
	}

	_codesMinioPermission = []string{
		"Access Denied",
	}

	_codesMinioPreconditionFailed = []string{
		"At least one of the pre-conditions you specified did not hold",
	}

	_codesMinioConflict = []string{
		"Bucket not empty",
	}
)

func (k KindMinio) String() string {
	switch k {
	case KindMinioPermission:
		return _kindMinioPermission
	case KindMinioPreconditionFailed:
		return _kindMinioPreconditionFailed
	case KindMinioConflict:
		return _kindMinioConflict
	case KindMinioNotExists:
		return _kindMinioNotExists
	case KindMinioIO:
		return _kindMinioIO
	default:
		return _kindMinioOther
	}
}

func KindFromMinio(err error) KindMinio {
	switch {
	case isStrContainsInsensitive(err.Error(), _codesMinioIO):
		return KindMinioIO
	case isStrContainsInsensitive(err.Error(), _codesMinioConflict):
		return KindMinioConflict
	case isStrContainsInsensitive(err.Error(), _codesMinioNotExists):
		return KindMinioNotExists
	case isStrContainsInsensitive(err.Error(), _codesMinioPermission):
		return KindMinioPermission
	case isStrContainsInsensitive(err.Error(), _codesMinioPreconditionFailed):
		return KindMinioPreconditionFailed
	default:
		return KindMinioOther
	}
}
