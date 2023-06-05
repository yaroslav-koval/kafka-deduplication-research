package cerror

const (
	// Known values for "type" error attribute
	ErrAttributeTypeAPI          = "api"
	ErrAttributeTypeDB           = "db"
	ErrAttributeTypeMediaStorage = "mediastorage"
	ErrAttributeTypeMsgBroker    = "msgbroker"

	// Known values for "subtype" error attribute
	ErrAttributeSubtypeHTTP     = "http"
	ErrAttributeSubtypeRedis    = "redis"
	ErrAttributeSubtypePostgres = "postgres"
	ErrAttributeSubtypeMinio    = "minio"
	ErrAttributeSubtypeKafka    = "kafka"
	ErrAttributeSubtypeElastic  = "elastic"
)

// ErrAttributes represents a structured additional information about an error
type ErrAttributes struct {
	Type    string `json:"type,omitempty"`
	Subtype string `json:"subtype,omitempty"`
	Code    int    `json:"code,omitempty"`
	CodeStr string `json:"codeStr,omitempty"`
}

// ErrAttributesValuesHTTP represents information about an error during http communication
type ErrAttributesValuesHTTP struct {
	Code int
}

// ErrAttributesValuesRedis represents information about an error during redis interaction
type ErrAttributesValuesRedis struct {
	Err error
}

// ErrAttributesValuesRedis represents information about an error during minio interaction
type ErrAttributesValuesMinio struct {
	Err error
}

// ErrAttributesValuesRedis represents information about an error during kafka interaction
type ErrAttributesValuesKafka struct {
	Err error
}

// ErrAttributesValuesRedis represents information about an error during elastic interaction
type ErrAttributesValuesElastic struct {
	Err error
}

// ErrAttributesValuesRedis represents information about an error during psotgres interaction
type ErrAttributesValuesPostgres struct {
	Err error
}
