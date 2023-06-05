package cerror

import (
	"fmt"

	"github.com/segmentio/kafka-go"
)

type KindKafka uint8

const (
	_ KindKafka = iota
	KindKafkaOther
	KindKafkaPermission
	KindKafkaIO
	KindKafkaUnknown // unknown topic, member, etc.

	_kindKafkaOther      = "other"
	_kindKafkaPermission = "permission"
	_kindKafkaIO         = "io"
	_kindKafkaUnknown    = "unknown"
)

var (
	_codesKafkaUnknown = []string{
		fmt.Sprintf("[%d]", kafka.UnknownTopicOrPartition),
		fmt.Sprintf("[%d]", kafka.Unknown),
		fmt.Sprintf("[%d]", kafka.UnknownMemberId),
		fmt.Sprintf("[%d]", kafka.UnknownProducerId),
		fmt.Sprintf("[%d]", kafka.UnknownLeaderEpoch),
		fmt.Sprintf("[%d]", kafka.UnknownTopicID),
	}

	_codesKafkaPermission = []string{
		fmt.Sprintf("[%d]", kafka.InvalidSessionTimeout),
		fmt.Sprintf("[%d]", kafka.TopicAuthorizationFailed),
		fmt.Sprintf("[%d]", kafka.GroupAuthorizationFailed),
		fmt.Sprintf("[%d]", kafka.GroupAuthorizationFailed),
		fmt.Sprintf("[%d]", kafka.UnsupportedSASLMechanism),
		fmt.Sprintf("[%d]", kafka.IllegalSASLState),
		fmt.Sprintf("[%d]", kafka.PolicyViolation),
		fmt.Sprintf("[%d]", kafka.SecurityDisabled),
		fmt.Sprintf("[%d]", kafka.BrokerAuthorizationFailed),
		fmt.Sprintf("[%d]", kafka.SASLAuthenticationFailed),
	}

	_codesKafkaIO = []string{
		"no such host",
		fmt.Sprintf("[%d]", kafka.RequestTimedOut),
		fmt.Sprintf("[%d]", kafka.BrokerNotAvailable),
		fmt.Sprintf("[%d]", kafka.ReplicaNotAvailable),
		fmt.Sprintf("[%d]", kafka.MessageSizeTooLarge),
		fmt.Sprintf("[%d]", kafka.NetworkException),
		fmt.Sprintf("[%d]", kafka.InvalidRequiredAcks),
		fmt.Sprintf("[%d]", kafka.InvalidCommitOffsetSize),
		fmt.Sprintf("[%d]", kafka.KafkaStorageError),
	}
)

func (k KindKafka) String() string {
	switch k {
	case KindKafkaPermission:
		return _kindKafkaPermission
	case KindKafkaIO:
		return _kindKafkaIO
	case KindKafkaUnknown:
		return _kindKafkaUnknown
	default:
		return _kindKafkaOther
	}
}

func KindFromKafka(err error) KindKafka {
	switch {
	case isStrContainsInsensitive(err.Error(), _codesKafkaPermission):
		return KindKafkaPermission
	case isStrContainsInsensitive(err.Error(), _codesKafkaIO):
		return KindKafkaIO
	case isStrContainsInsensitive(err.Error(), _codesKafkaUnknown):
		return KindKafkaUnknown
	}

	return KindKafkaOther
}
