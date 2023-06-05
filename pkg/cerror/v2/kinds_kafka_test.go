package cerror_test

import (
	"errors"
	"fmt"
	"kafka-polygon/pkg/cerror/v2"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/tj/assert"
)

type expectedKafkaKind struct {
	kind       cerror.KindKafka
	kindString string
}

var expectedKindKafkaTable = []expectedKafkaKind{
	{cerror.KindKafkaOther, "other"},
	{cerror.KindKafkaPermission, "permission"},
	{cerror.KindKafkaIO, "io"},
}

var errKafkaList = []struct {
	kind cerror.KindKafka
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

func TestKindFromKafka(t *testing.T) {
	t.Parallel()

	for _, e := range errKafkaList {
		errKind := cerror.KindFromKafka(e.err)
		assert.NotNil(t, errKind)
		assert.Equal(t, e.kind, errKind)
	}
}
