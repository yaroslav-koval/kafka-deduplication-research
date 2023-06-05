package env

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"reflect"
	"strings"
	"time"
)

type Service struct {
	Name      string `env:"SERVICE_NAME,required"`
	LogLevel  string `env:"SERVICE_LOG_LEVEL" envDefault:"info"`
	DebugMode bool   `env:"DEBUG_MODE" envDefault:"false"`
	PProfMode bool   `env:"PPROF_MODE" envDefault:"false"`
}

type HTTPServer struct {
	Port            string `env:"HTTP_SERVER_PORT,required"`
	Host            string `env:"HTTP_SERVER_HOST" envDefault:""`
	ReadTimeoutSec  int64  `env:"HTTP_SERVER_READ_TIMEOUT_SEC" envDefault:"60"`
	WriteTimeoutSec int64  `env:"HTTP_SERVER_WRITE_TIMEOUT_SEC" envDefault:"60"`
	IdleTimeoutSec  int64  `env:"HTTP_SERVER_IDLE_TIMEOUT_SEC" envDefault:"0"`
	CloseOnShutdown bool   `env:"HTTP_SERVER_CLOSE_ON_SHUTDOWN" envDefault:"true"`
	LogResponse     bool   `env:"HTTP_SERVER_LOG_RESPONSE" envDefault:"false"`
	TraceEnabled    bool   `env:"HTTP_SERVER_TRACE_ENABLED" envDefault:"true"`
}

func (f *HTTPServer) ReadTimeout() time.Duration {
	return time.Duration(f.ReadTimeoutSec) * time.Second
}

func (f *HTTPServer) WriteTimeout() time.Duration {
	return time.Duration(f.WriteTimeoutSec) * time.Second
}

func (f *HTTPServer) IdleTimeout() time.Duration {
	return time.Duration(f.IdleTimeoutSec) * time.Second
}

type Mongo struct {
	Host       string `env:"MONGODB_URL,required"`
	SchemaName string `env:"MONGODB_DB_NAME,required"`
}

type Postgres struct {
	Host            string        `env:"POSTGRES_HOST,required"`
	Port            string        `env:"POSTGRES_PORT,required"`
	Username        string        `env:"POSTGRES_USERNAME,required"`
	Password        string        `env:"POSTGRES_PASSWORD,required"`
	DBName          string        `env:"POSTGRES_DBNAME,required"`
	SSLMode         string        `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	ConnTimeout     int           `env:"POSTGRES_CONN_TIMEOUT" envDefault:"20"`
	MaxOpenConns    int           `env:"POSTGRES_MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns    int           `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"3"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"10s"`
	DBQueryLogLevel string        `env:"POSTGRES_DB_QUERY_LOG_LEVEL" envDefault:"none"`
	TraceEnabled    bool          `env:"POSTGRES_TRACE_ENABLED" envDefault:"false"`
}

type Kafka struct {
	Enabled                    bool          `env:"KAFKA_ENABLED" envDefault:"true"`
	LoggerEnabled              bool          `env:"KAFKA_LOGGER_ENABLED" envDefault:"false"`
	UseKeyDoubleQuote          bool          `env:"KAFKA_KEY_DOUBLE_QUOTE" envDefault:"true"`
	Hosts                      []string      `env:"KAFKA_BOOTSTRAP_SERVERS,required"`
	SASLSecurityProtocol       string        `env:"KAFKA_SASL_SECURITY_PROTOCOL" envDefault:""`
	SASLAlgorithm              string        `env:"KAFKA_SASL_ALGORITHM" envDefault:""`
	SASLUsername               string        `env:"KAFKA_SASL_USERNAME,required"`
	SASLPassword               string        `env:"KAFKA_SASL_PASSWORD,required"`
	CommitOnError              bool          `env:"KAFKA_COMMIT_ON_ERROR" envDefault:"false"`
	CommitOnErrorMessagesCount int           `env:"KAFKA_COMMIT_ON_ERROR_MESSAGES_COUNT" envDefault:"1"`
	RerunDelay                 time.Duration `env:"KAFKA_RERUN_DELAY" envDefault:"30s"`
	TLSEnabled                 bool          `env:"KAFKA_TLS_ENABLED" envDefault:"false"`
	TLSInsecureSkipVerify      bool          `env:"KAFKA_TLS_INSECURE_SKIP_VERIFY" envDefault:"false"`
	TLSClientCertFile          string        `env:"KAFKA_TLS_CLIENT_CERT_FILE" envDefault:""`
	TLSClientKeyFile           string        `env:"KAFKA_TLS_CLIENT_KEY_FILE" envDefault:""`
	TLSRootCACertFile          string        `env:"KAFKA_TLS_ROOT_CA_CERT_FILE" envDefault:""`
	AutoCreateTopic            bool          `env:"KAFKA_AUTO_CREATE_TOPIC" envDefault:"false"`
	TraceEnabled               bool          `env:"KAFKA_TRACE_ENABLED" envDefault:"false"`
}

type KafkaProducer struct {
	WriterBatchSize    int           `env:"KAFKA_PRODUCER_BATCH_SIZE" envDefault:"10"`
	MaxAttempts        int           `env:"KAFKA_PRODUCER_MAX_ATTEMPTS" envDefault:"3"`
	MaxRetry           int           `env:"KAFKA_PRODUCER_MAX_RETRY" envDefault:"10"`
	WriteTimeout       time.Duration `env:"KAFKA_PRODUCER_WRITE_TIMEOUT" envDefault:"50ms"`
	WriterBatchTimeout time.Duration `env:"KAFKA_PRODUCER_WRITER_BATCH_TIMEOUT" envDefault:"2s"`
	MaxAttemptsDelay   time.Duration `env:"KAFKA_PRODUCER_MAX_ATTEMPTS_DELAY" envDefault:"5s"`
}

type KafkaConsumer struct {
	GroupID           string        `env:"KAFKA_CONSUMER_GROUP_ID,required"`
	MinBytes          int           `env:"KAFKA_CONSUMER_MIN_BYTES" envDefault:"1000"`
	MaxBytes          int           `env:"KAFKA_CONSUMER_MAX_BYTES" envDefault:"1000000"`
	QueueCapacity     int           `env:"KAFKA_CONSUMER_QUEUE_CAPACITY" envDefault:"100"`
	MaxWait           time.Duration `env:"KAFKA_CONSUMER_MAX_WAIT" envDefault:"10s"`
	ReadLagInterval   time.Duration `env:"KAFKA_CONSUMER_READ_LAG_INTERVAL" envDefault:"-1s"`
	CancelIfOneFailed bool          `env:"KAFKA_CONSUMER_CANCEL_ALL_IF_ONE_FAILED" envDefault:"false"`
	MaxAttempts       int           `env:"KAFKA_CONSUMER_MAX_ATTEMPTS" envDefault:"3"`
}

type Redis struct {
	Host      string        `env:"REDIS_HOST,required"`
	Password  string        `env:"REDIS_PASSWORD" envDefault:""`
	DBNumber  int           `env:"REDIS_DB_NUMBER,required"`
	KeyPrefix string        `env:"REDIS_KEY_PREFIX" envDefault:""`
	TTL       time.Duration `env:"REDIS_TTL" envDefault:"86400s"`
}

type Trace struct {
	Environment               string `env:"TRACE_ENVIRONMENT" envDefault:"development"`
	URL                       string `env:"TRACE_URL" envDefault:""`
	UseAgent                  bool   `env:"TRACE_USE_AGENT" envDefault:"false"`
	AgentHost                 string `env:"TRACE_AGENT_HOST" envDefault:"localhost"`
	AgentPort                 string `env:"TRACE_AGENT_PORT" envDefault:"6831"`
	AgentReconnectingInterval int64  `env:"TRACE_AGENT_RECONNECTION_INTERVAL" envDefault:"30"`
}

type Paging struct {
	DefaultLimit uint `env:"PAGING_DEFAULT_LIMIT" envDefault:"30"`
	MaxLimit     uint `env:"PAGING_MAX_LIMIT" envDefault:"1000"`
}

type Migration struct {
	Version uint `env:"MIGRATION_VERSION,required"`
}

type MigrationWorkflow struct {
	Version     uint   `env:"MIGRATION_WORKFLOW_VERSION,required"`
	SchemaTable string `env:"MIGRATION_WORKFLOW_SCHEMA_TABLE" envDefault:"schema_migrations_workflow"`
}

type MigrationErrorBrokerStore struct {
	Version uint `env:"MIGRATION_ERROR_BROKER_STORE_VERSION,required"`
}

type Workflow struct {
	NoRetryOnError          bool  `env:"WORKFLOW_NO_RETRY_ON_ERROR" envDefault:"false"`
	StepAutoRetryErrorCodes []int `env:"WORKFLOW_STEP_AUTO_RETRY_ERROR_CODES" envDefault:"429,502,503"`
}

type FHIR struct {
	HostAPI        string        `env:"FHIR_SERVER_API_URL,required"`
	HostSearch     string        `env:"FHIR_SERVER_SEARCH_URL,required"`
	User           string        `env:"FHIR_SERVER_API_CONSUMER,required"`
	RequestTimeout time.Duration `env:"FHIR_SERVER_API_REQUEST_TIMEOUT" envDefault:"30s"`
	MaxSearchLimit int           `env:"FHIR_SERVER_SEARCH_MAX_SEARCH_LIMIT" envDefault:"1000"`
}

type ClickHouse struct {
	Host            string `env:"CH_HOST,required"`
	Port            string `env:"CH_PORT,required"`
	Username        string `env:"CH_USERNAME,required"`
	Password        string `env:"CH_PASSWORD,required"`
	DBName          string `env:"CH_DBNAME,required"`
	ConnTimeout     string `env:"CH_CONN_TIMEOUT" envDefault:"20s"`
	Compress        bool   `env:"CH_COMPRESS" envDefault:"false"`
	Debug           bool   `env:"CH_DEBUG" envDefault:"false"`
	MaxOpenConns    int    `env:"CH_MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns    int    `env:"CH_MAX_IDLE_CONNS" envDefault:"3"`
	MaxConnLifetime int64  `env:"CH_MAX_CONN_LIFETIME" envDefault:"10"`
}

type Minio struct {
	Host               string `env:"MINIO_HOST,required"`
	AccessKeyID        string `env:"MINIO_ACCESS_KEY_ID,required"`
	SecretAccessKey    string `env:"MINIO_SECRET_ACCESS_KEY,required"`
	Region             string `env:"MINIO_REGION" envDefault:""`
	Token              string `env:"MINIO_TOKEN" envDefault:""`
	UseSSL             bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	ResponseExpiredMin int64  `env:"MINIO_EXPIRED_MIN" envDefault:"10"`
}

type AWS struct {
	AccessKeyID     string `env:"AWS_ACCESS_KEY_ID" envDefault:""`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY" envDefault:""`
	Region          string `env:"AWS_REGION" envDefault:""`
	Token           string `env:"AWS_TOKEN" envDefault:""`
}

type AWSS3 struct {
	ResponseExpiredMin int64 `env:"AWSS3_EXPIRED_MIN" envDefault:"10"`
}

type MediaStorage struct {
	UseStorage       string `env:"MEDIA_STORAGE_TYPE,required"`
	PublicPresignURL string `env:"MEDIA_STORAGE_PUBLIC_PRESIGN_URL" envDefault:""`
}

type MediaStorageWithBucket struct {
	MediaStorage
	BucketName string `env:"MEDIA_STORAGE_BUCKET_NAME,required"`
}

type MediaStorageClient struct {
	Host           string        `env:"MEDIA_STORAGE_SERVICE_HOST,required"`
	RequestTimeout time.Duration `env:"MEDIA_STORAGE_SERVICE_REQUEST_TIMEOUT" envDefault:"30s"`
}

type OTPClient struct {
	Host           string        `env:"OTP_SERVICE_HOST,required"`
	RequestTimeout time.Duration `env:"OTP_SERVICE_REQUEST_TIMEOUT" envDefault:"30s"`
}

type NotificationClient struct {
	Host           string        `env:"NOTIFICATION_SERVICE_HOST,required"`
	RequestTimeout time.Duration `env:"NOTIFICATION_SERVICE_REQUEST_TIMEOUT" envDefault:"30s"`
	ProviderEmail  string        `env:"NOTIFICATION_PROVIDER_EMAIL" envDefault:""`
	ProviderSMS    string        `env:"NOTIFICATION_PROVIDER_SMS" envDefault:""`
}

type Elastic struct {
	Hosts                    []string `env:"ELASTIC_HOSTS,required"`
	User                     string   `env:"ELASTIC_USER" envDefault:""`
	Pass                     string   `env:"ELASTIC_PASS" envDefault:""`
	EnableRetryOnTimeout     bool     `env:"ELASTIC_RETRY" envDefault:"true"`
	MaxRetries               int      `env:"ELASTIC_MAX_RETRIES" envDefault:"5"`
	CompressRequestBody      bool     `env:"ELASTIC_COMPRESS_REQUEST_BODY" envDefault:"false"`
	EnableMetrics            bool     `env:"ELASTIC_ENABLE_METRICS" envDefault:"false"`
	Debug                    bool     `env:"ELASTIC_DEBUG" envDefault:"false"`
	EnableCompatibilityMode  bool     `env:"ELASTIC_ENABLE_COMPATIBILITY_MODE" envDefault:"false"`
	MaxIdleConnsPerHost      int      `env:"ELASTIC_MAX_IDLE_CONNS_PER_HOST" envDefault:"10"`
	ResponseHeaderTimeoutSec int      `env:"ELASTIC_RESPONSE_HEADER_TIMEOUT_SEC" envDefault:"60"`
}

type Notify struct {
	MaxAttempts int64 `env:"NOTIFY_MAX_ATTEMPTS" envDefault:"10"`
}

type NotifyPostmark struct {
	AccountToken string `env:"NOTIFY_POSTMARK_ACCOUNT_TOKEN" envDefault:""`
	ServerToken  string `env:"NOTIFY_POSTMARK_SERVER_TOKEN"`
}

type NotifyTiniyo struct {
	TiniyoHostAPI        string        `env:"NOTIFY_TINIYO_HOST_API"`
	TiniyoAuthID         string        `env:"NOTIFY_TINIYO_AUTH_ID"`
	TiniyoAuthToken      string        `env:"NOTIFY_TINIYO_AUTH_TOKEN"`
	TiniyoRequestTimeout time.Duration `env:"NOTIFY_TINIYO_REQUEST_TIMEOUT" envDefault:"30s"`
}

type NotifyMailgun struct {
	MailgunDomain string `env:"NOTIFY_MAILGUN_DOMAIN"`
	MailgunAPIKey string `env:"NOTIFY_MAILGUN_API_KEY"`
}

type NotifyTwilio struct {
	AccountSID          string `env:"NOTIFY_TWILIO_ACCOUNT_SID"`
	AuthToken           string `env:"NOTIFY_TWILIO_AUTH_TOKEN"`
	CallbackURL         string `env:"NOTIFY_TWILIO_CALLBACK_URL" envDefault:""`
	MessagingServiceSID string `env:"NOTIFY_TWILIO_MESSAGING_SERVICE_SID" envDefault:""`
}

type KeycloakClient struct {
	HostAPI        string        `env:"KEYCLOAK_HOST_API,required"`
	Realm          string        `env:"KEYCLOAK_REALM,required"`
	RequestTimeout time.Duration `env:"KEYCLOAK_STORAGE_SERVICE_REQUEST_TIMEOUT" envDefault:"30s"`
	GrantType      string        `env:"KEYCLOAK_GRANT_TYPE" envDefault:"password"`
	ClientID       string        `env:"KEYCLOAK_CLIENT_ID" envDefault:"admin-cli"`
	ClientSecret   string        `env:"KEYCLOAK_CLIENT_SECRET,required"`
	Username       string        `env:"KEYCLOAK_USERNAME,required"`
	Password       string        `env:"KEYCLOAK_PASSWORD,required"`
}

type GeodictionaryClient struct {
	Host           string        `env:"GEODICTIONARY_SERVICE_HOST,required"`
	RequestTimeout time.Duration `env:"GEODICTIONARY_SERVICE_REQUEST_TIMEOUT" envDefault:"30s"`
}

func GetDefaultParsers() map[reflect.Type]env.ParserFunc {
	return map[reflect.Type]env.ParserFunc{
		reflect.TypeOf([]string{}): func(v string) (interface{}, error) {
			return strings.Split(v, ","), nil
		},
	}
}

func ParseCfg(c interface{}) error {
	cfgParsers := GetDefaultParsers()
	if err := env.ParseWithFuncs(c, cfgParsers); err != nil {
		return cerror.New(context.Background(), cerror.KindInternal, err).LogError()
	}

	return nil
}

func ParseCfgWithParsers(c interface{}, parsers map[reflect.Type]env.ParserFunc) error {
	cfgParsers := GetDefaultParsers()
	for k, v := range parsers {
		cfgParsers[k] = v
	}

	if err := env.ParseWithFuncs(c, cfgParsers); err != nil {
		return cerror.New(context.Background(), cerror.KindInternal, err).LogError()
	}

	return nil
}
