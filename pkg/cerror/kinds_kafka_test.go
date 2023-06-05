package cerror_test

import (
	"errors"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/tj/assert"
)

type expectedKafkaKind struct {
	kind         cerror.Kind
	kindString   string
	kindHTTPCode int
}

var expectedKindKafkaTable = []expectedKafkaKind{
	{cerror.KindKafkaOther, "kafka_other_error", http.StatusInternalServerError},
	{cerror.KindKafkaPermission, "kafka_permission_denied", http.StatusInternalServerError},
	{cerror.KindKafkaIO, "kafka_io_error", http.StatusInternalServerError},
}

var errKafkaList = []struct {
	kind cerror.Kind
	err  error
}{
	{
		kind: cerror.KindKafkaOther,
		err:  errors.New("other error"),
	},
	{
		kind: cerror.KindKafkaUnknown,
		err:  fmt.Errorf("[%d] unknown_topic_or_partition", kafka.UnknownTopicOrPartition),
	},
	{
		kind: cerror.KindKafkaPermission,
		err:  fmt.Errorf("[%d] invalid_session_timeout", kafka.InvalidSessionTimeout),
	},
	{
		kind: cerror.KindKafkaIO,
		err:  fmt.Errorf("[%d] request_timedOut", kafka.RequestTimedOut),
	},
}

func TestKindKafkaString(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindKafkaTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}

func TestKindKafkaHTTPCode(t *testing.T) {
	t.Parallel()

	for _, v := range expectedKindKafkaTable {
		assert.Equal(t, v.kindHTTPCode, v.kind.HTTPCode())
	}
}

func TestKafkaToKind(t *testing.T) {
	t.Parallel()

	for _, e := range errKafkaList {
		errKind := cerror.KafkaToKind(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
