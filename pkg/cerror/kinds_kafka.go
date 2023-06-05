package cerror

import (
	"fmt"
	"net/http"

	"github.com/segmentio/kafka-go"
)

type KafkaKind uint8

const (
	_ KafkaKind = iota
	KindKafkaOther
	KindKafkaPermission // KindPermission denied
	KindKafkaIO         // External I/O error such as network failure
	KindKafkaUnknown

	kafkaUnknown          = "kafka_unknown_error"
	kafkaOther            = "kafka_other_error" // king message
	kafkaPermissionDenied = "kafka_permission_denied"
	kafkaIO               = "kafka_io_error"
)

var (
	_kafkaUnknownCodes = []string{
		fmt.Sprintf("[%d]", kafka.UnknownTopicOrPartition),
		fmt.Sprintf("[%d]", kafka.Unknown),
		fmt.Sprintf("[%d]", kafka.UnknownMemberId),
		fmt.Sprintf("[%d]", kafka.UnknownProducerId),
		fmt.Sprintf("[%d]", kafka.UnknownLeaderEpoch),
		fmt.Sprintf("[%d]", kafka.UnknownTopicID),
	}

	_kafkaPermissionCodes = []string{
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

	_kafkaIOCodes = []string{
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

func (k KafkaKind) String() string {
	switch k {
	case KindKafkaPermission:
		return kafkaPermissionDenied
	case KindKafkaIO:
		return kafkaIO
	case KindKafkaUnknown:
		return kafkaUnknown
	case KindKafkaOther:
		return kafkaOther
	}

	return kafkaUnknown
}

func (k KafkaKind) HTTPCode() int {
	return http.StatusInternalServerError
}

func (k KafkaKind) Group() KindGroup {
	return GroupKafka
}

func KafkaToKind(err error) KafkaKind {
	switch {
	case checkCode(err.Error(), _kafkaPermissionCodes):
		return KindKafkaPermission
	case checkCode(err.Error(), _kafkaIOCodes):
		return KindKafkaIO
	case checkCode(err.Error(), _kafkaUnknownCodes):
		return KindKafkaUnknown
	default:
		return KindKafkaOther
	}
}
