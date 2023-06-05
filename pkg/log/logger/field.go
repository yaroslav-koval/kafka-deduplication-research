package logger

// Log fields names
const (
	FieldMessage         = "message"         // general log message
	FieldRequestID       = "requestID"       // http request id or other context identifier
	FieldPayload         = "payload"         // any log related data in key/value struct
	FieldStackTrace      = "stackTrace"      // stack trace
	FieldError           = "error"           // error message
	FieldErrorGroup      = "errorGroup"      // kind group
	FieldErrorKind       = "errorKind"       // text error kind representation
	FieldErrorCode       = "errorCode"       // numeric error code representation
	FieldErrorAttributes = "errorAttributes" // error related attributes
	FieldStatus          = "status"          // response status code
	FieldMethod          = "method"          // HTTP request method
	FieldLatency         = "latency"         // operation latency
	FieldURL             = "url"             // HTTP request url
	FieldLogType         = "logType"         // type of log message (e.g. audit)
	FieldServiceName     = "serviceName"     // name of service that initiates log
	FieldClientIP        = "clientIP"        // service's client ip address
	FieldClientID        = "clientID"        // service's client identifier (e.g. ios-mobile-app)
	FieldUserID          = "userID"          // service's user identifier (e.g. 123e4567-e89b-12d3-a456-426614174000)
	FieldUserAgent       = "userAgent"       // service's client agent
	FieldResponse        = "response"        // HTTP response body
	FieldOperations      = "operations"      // log operations trace
)

// Field is a structure representation of key/value
type Field struct {
	key   string
	value interface{}
}

// NewField constructs Field struct from key and value
func NewField(key string, value interface{}) *Field {
	return &Field{key: key, value: value}
}

// Key returns Field's key
func (l *Field) Key() string {
	return l.key
}

// Value returns Field's value
func (l *Field) Value() interface{} {
	return l.value
}

// Fields represents list of fields
type Fields []*Field

// NewFields constructs Fields struct from given list of Fields
func NewFields(f ...*Field) Fields {
	if len(f) == 0 {
		return make(Fields, 0)
	}

	return Fields(f)
}

// AddValue makes copy of Fields and adds appends new value.
// If given key already exists, its value will be modified.
func (f Fields) AddValue(key string, value interface{}) Fields {
	newFields := NewFields()

	for _, v := range f {
		if v.key == key {
			continue
		}

		newFields = append(newFields, NewField(v.key, v.value))
	}

	newFields = append(newFields, NewField(key, value))

	return newFields
}

// GetValue returns value with given key.
// If value with given key doesn't exist second parameter will be false.
func (f Fields) GetValue(key string) (interface{}, bool) {
	for _, v := range f {
		if v.key == key {
			return v.value, true
		}
	}

	return nil, false
}

// Map converts Fields to map
func (f Fields) Map() map[string]interface{} {
	m := make(map[string]interface{})
	for _, v := range f {
		m[v.key] = v.value
	}

	return m
}
