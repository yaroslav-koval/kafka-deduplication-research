package cerror

type KindGroup uint8

const (
	_ KindGroup = iota
	GroupHTTP
	GroupDB
	GroupElastic
	GroupKafka
	GroupMinio
	GroupRedis
)

func (g KindGroup) String() string {
	switch g {
	case GroupHTTP:
		return "http"
	case GroupDB:
		return "db"
	case GroupElastic:
		return "elastic"
	case GroupKafka:
		return "kafka"
	case GroupMinio:
		return "minio"
	case GroupRedis:
		return "redis"
	default:
		return ""
	}
}
