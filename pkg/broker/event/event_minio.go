package event

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/cmd/metadata"
	"time"

	"github.com/minio/minio-go/v7/pkg/notification"
)

type MinioEvent interface {
	BaseEvent
	GetEventName() string
	GetRecords() []*Record
}

type UserMetadata map[string]string

type PrincipalID struct {
	PrincipalID string `json:"principalId"`
}

type RequestParameters struct {
	PrincipalID     string `json:"principalId"`
	Region          string `json:"region"`
	SourceIPAddress string `json:"sourceIPAddress"`
}

type ResponseElements struct {
	ContentLength        string `json:"content-length,omitempty"`
	XAmzRequestID        string `json:"x-amz-request-id,omitempty"`
	XMinioDeploymentID   string `json:"x-minio-deployment-id,omitempty"`
	XMinioOriginEndpoint string `json:"x-minio-origin-endpoint,omitempty"`
}

type Bucket struct {
	Name          string      `json:"name"`
	OwnerIdentity PrincipalID `json:"ownerIdentity"`
	Arn           string      `json:"arn"`
}

type Object struct {
	Key          string       `json:"key"`
	Size         int          `json:"size"`
	ETag         string       `json:"eTag"`
	ContentType  string       `json:"contentType"`
	UserMetadata UserMetadata `json:"userMetadata"`
	Sequencer    string       `json:"sequencer"`
}

type ObjS3 struct {
	S3SchemaVersion string `json:"s3SchemaVersion"`
	ConfigurationID string `json:"configurationId"`
	Bucket          Bucket `json:"bucket"`
	Object          Object `json:"object"`
}

type Source struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	UserAgent string `json:"userAgent"`
}

type Record struct {
	EventVersion      string                 `json:"eventVersion"`
	EventSource       string                 `json:"eventSource"`
	AwsRegion         string                 `json:"awsRegion"`
	EventTime         time.Time              `json:"eventTime"`
	EventName         notification.EventType `json:"eventName"`
	UserIdentity      PrincipalID            `json:"userIdentity"`
	RequestParameters RequestParameters      `json:"requestParameters"`
	ResponseElements  ResponseElements       `json:"responseElements"`
	S3                ObjS3                  `json:"s3"`
	Source            Source                 `json:"source"`
}

type MinioData struct {
	EventName string    `json:"EventName"`
	Key       string    `json:"Key"`
	Records   []*Record `json:"Records"`
	// Header set header field with context
	Header Header `json:"header"`
	// Debug is set to true if this is a debugging event.
	Debug    bool
	Metadata metadata.Meta `json:"metadata"`
}

// GetRecords of MinioData
func (md *MinioData) GetRecords() []*Record {
	return md.Records
}

// GetEventName of MinioData
func (md *MinioData) GetEventName() string {
	return md.EventName
}

// GetID of MinioData
func (md *MinioData) GetID() string {
	return md.Key
}

// GetDebug of MinioData
func (md *MinioData) GetDebug() bool {
	return md.Debug
}

// ToByte of MinioData
func (md *MinioData) ToByte() []byte {
	b, err := json.Marshal(md)
	if err != nil {
		return nil
	}

	return b
}

// Unmarshal of MinioData
func (md *MinioData) Unmarshal(msg Message) error {
	return json.Unmarshal(msg.Value, md)
}

func (md *MinioData) GetHeader() Header {
	return md.Header
}

func (md *MinioData) WithHeader(ctx context.Context) {
	md.Header.XRequestIDFromContext(ctx)
}

func (md *MinioData) GetMeta() metadata.Meta {
	return md.Metadata
}

func (md *MinioData) WithMeta(meta metadata.Meta) {
	md.Metadata = meta
}
