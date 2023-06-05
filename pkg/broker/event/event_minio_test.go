package event_test

import (
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var records = []*event.Record{
	{
		EventVersion: "test-event-version",
		EventSource:  "test-event-source",
		AwsRegion:    "test-aws-region",
		EventTime:    time.Now(),
		EventName:    "test-event-name",
		UserIdentity: event.PrincipalID{
			PrincipalID: "test-principal-id",
		},
		RequestParameters: event.RequestParameters{
			PrincipalID:     "test-principal-id",
			Region:          "test-region",
			SourceIPAddress: "0.0.0.0",
		},
		ResponseElements: event.ResponseElements{
			ContentLength:        "test-content-length",
			XAmzRequestID:        "test-x-amz-request-id",
			XMinioDeploymentID:   "test-x-minio-deployment-id",
			XMinioOriginEndpoint: "test-x-minio-origin-endpoint",
		},
		S3: event.ObjS3{
			S3SchemaVersion: "test-s3-schema-version",
			ConfigurationID: "test-configuration-id",
			Bucket: event.Bucket{
				Name: "test-bucket-name",
				OwnerIdentity: event.PrincipalID{
					PrincipalID: "test-principal-id",
				},
				Arn: "test-arn",
			},
			Object: event.Object{
				Key:          "test-key",
				Size:         1,
				ETag:         "test-etag",
				ContentType:  "test-content-type",
				UserMetadata: event.UserMetadata{"key": "val"},
				Sequencer:    "test-sequencer",
			},
		},
	},
}

func TestMinioEventType(t *testing.T) {
	t.Parallel()

	e := event.MinioData{}

	var (
		i interface{} = e
		p interface{} = &e
	)

	_, ok := i.(event.BaseEvent)
	assert.Equal(t, ok, false)

	_, ok = p.(event.BaseEvent)
	assert.Equal(t, ok, true)

	_, ok = p.(event.MinioEvent)

	assert.Equal(t, ok, true)
	assert.Equal(t, "", e.GetID())
}

func TestMinioEvent(t *testing.T) {
	t.Parallel()

	e := event.MinioData{
		EventName: "test-event-name",
		Key:       "test-key",
		Records:   records,
		Header:    expHeader,
		Metadata:  expMeta,
		Debug:     true,
	}

	assert.Equal(t, "test-key", e.GetID())
	assert.Equal(t, expHeader, e.GetHeader())
	assert.Equal(t, expMeta, e.GetMeta())
	assert.Equal(t, true, e.GetDebug())
	assert.NotNil(t, e.ToByte())
	assert.Equal(t, true, len(e.ToByte()) > 0)
	assert.Equal(t, "test-event-name", e.GetEventName())
	assert.Equal(t, records, e.GetRecords())
}

func TestMinioEventUnmarshal(t *testing.T) {
	t.Parallel()

	expectedData := map[string]interface{}{
		"EventName": "test-event-name",
		"Key":       "test-key",
		"Records":   records,
		"header":    expHeader,
		"metadata":  expMeta,
	}

	b, err := json.Marshal(expectedData)
	require.NoError(t, err)

	msg := event.Message{
		Value: b,
	}
	e := event.MinioData{}
	err = e.Unmarshal(msg)
	require.NoError(t, err)

	assert.Equal(t, expectedData["header"], e.GetHeader())
	assert.Equal(t, expectedData["metadata"], e.GetMeta())
	assert.Equal(t, expectedData["Key"], e.GetID())
	assert.Equal(t, expectedData["EventName"], e.GetEventName())
	assert.Equal(t, len(expectedData["Records"].([]*event.Record)), len(e.GetRecords()))
	assert.Equal(t, false, e.GetDebug())
}

func TestMinioEventError(t *testing.T) {
	t.Parallel()

	msg := event.Message{
		Value: []byte(`{"num1":6.13,"strs1":{"a","b"}}`),
	}

	e := &event.MinioData{}
	err := e.Unmarshal(msg)
	require.Error(t, err)
}
