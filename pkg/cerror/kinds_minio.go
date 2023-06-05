package cerror

import (
	"net/http"
)

type MinioKind uint8

const (
	_ MinioKind = iota
	KindMinioOther
	KindMinioIO
	KindMinioNotExist
	KindMinioConflict
	KindMinioPermission         // KindPermission denied
	KindMinioPreconditionFailed // External I/O error such as network failure

	minioOther              = "minio_other_error"
	minioPermissionDenied   = "minio_permission_denied"
	minioPreconditionFailed = "minio_precondition_failed"
	minioUnknown            = "minio_unknown"
	minioNotExist           = "minio_not_exist"
	minioConflict           = "minio_conflict"
	minioIO                 = "minio_io_error"
)

var (
	_minioIOCodes = []string{
		"connection refused",
	}

	_minioNotExistCodes = []string{
		"The specified bucket does not exist",
		"The specified key does not exist",
	}

	_minioPermissionCodes = []string{
		"Access Denied",
	}

	_minioPreconditionFailedCodes = []string{
		"At least one of the pre-conditions you specified did not hold",
	}

	_minioConflictCodes = []string{
		"Bucket not empty",
	}
)

func (k MinioKind) String() string {
	switch k {
	case KindMinioPermission:
		return minioPermissionDenied
	case KindMinioPreconditionFailed:
		return minioPreconditionFailed
	case KindMinioOther:
		return minioOther
	case KindMinioConflict:
		return minioConflict
	case KindMinioNotExist:
		return minioNotExist
	case KindMinioIO:
		return minioIO
	}

	return minioUnknown
}

func (k MinioKind) HTTPCode() int {
	switch k {
	case KindMinioNotExist:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (k MinioKind) Group() KindGroup {
	return GroupMinio
}

func MinioToKind(err error) MinioKind {
	switch {
	case checkCode(err.Error(), _minioIOCodes):
		return KindMinioIO
	case checkCode(err.Error(), _minioConflictCodes):
		return KindMinioConflict
	case checkCode(err.Error(), _minioNotExistCodes):
		return KindMinioNotExist
	case checkCode(err.Error(), _minioPermissionCodes):
		return KindMinioPermission
	case checkCode(err.Error(), _minioPreconditionFailedCodes):
		return KindMinioPreconditionFailed
	default:
		return KindMinioOther
	}
}
