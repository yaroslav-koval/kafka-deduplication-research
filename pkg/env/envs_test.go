package env_test

import (
	"fmt"
	pkgenv "kafka-polygon/pkg/env"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tj/assert"
)

const trueString = "true"

type testConfig struct {
	TestMap map[string]string `env:"TEST_MAP,required"`
}

type requiredFields []string

func (fr requiredFields) hasLast(id int) bool {
	return id == len(fr)-1
}

func (fr requiredFields) firstKey() string {
	firstKey, _ := fr.getData(0)
	return firstKey
}

func (fr requiredFields) nextKey(id int) string {
	nextKey, _ := fr.getData(id + 1)
	return nextKey
}

func (fr requiredFields) getData(id int) (key, val string) {
	data := fr[id]
	key = data
	val = fmt.Sprintf("%s_test", key)

	if strings.Contains(data, ":") {
		sliceD := strings.Split(data, ":")
		key = sliceD[0]
		val = sliceD[1]
	}

	return
}

func checkRequiredFields(t *testing.T, cfgStr interface{}, sliceKeys requiredFields) {
	os.Clearenv()

	t.Run("not set interface{} values", func(t *testing.T) {
		err := pkgenv.ParseCfg(cfgStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("env: required environment variable %q is not set", sliceKeys.firstKey()))
	})

	t.Run("check required fields", func(t *testing.T) {
		for id := range sliceKeys {
			keyEnv, valEnv := sliceKeys.getData(id)
			err := os.Setenv(keyEnv, valEnv)
			assert.NoError(t, err)
			err = pkgenv.ParseCfg(cfgStr)
			if sliceKeys.hasLast(id) {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), fmt.Sprintf("env: required environment variable %q is not set", sliceKeys.nextKey(id)))
			}
		}
	})
}

func TestServiceEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["SERVICE_NAME"] = "service_name_test"
	mapConfigs["SERVICE_LOG_LEVEL"] = "log_level_test"
	mapConfigs["DEBUG_MODE"] = "true"
	mapConfigs["PPROF_MODE"] = "true"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgService := pkgenv.Service{}
	err := pkgenv.ParseCfg(&cfgService)
	assert.NoError(t, err)

	expDebugMode, _ := strconv.ParseBool(mapConfigs["DEBUG_MODE"])
	expPprofMode, _ := strconv.ParseBool(mapConfigs["PPROF_MODE"])

	assert.Equal(t, mapConfigs["SERVICE_NAME"], cfgService.Name)
	assert.Equal(t, mapConfigs["SERVICE_LOG_LEVEL"], cfgService.LogLevel)
	assert.Equal(t, expDebugMode, cfgService.DebugMode)
	assert.Equal(t, expPprofMode, cfgService.PProfMode)
}

func TestServiceEnvErr(t *testing.T) {
	cfgService := &pkgenv.Service{}
	sliceKeys := []string{
		"SERVICE_NAME",
	}

	checkRequiredFields(t, cfgService, sliceKeys)
}

func TestHTTPServerEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["HTTP_SERVER_PORT"] = "http_server_port_test"
	mapConfigs["HTTP_SERVER_HOST"] = "http_server_host_test"
	mapConfigs["HTTP_SERVER_LOG_RESPONSE"] = trueString

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgHTTPServer := pkgenv.HTTPServer{}
	err := pkgenv.ParseCfg(&cfgHTTPServer)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["HTTP_SERVER_PORT"], cfgHTTPServer.Port)
	assert.Equal(t, mapConfigs["HTTP_SERVER_HOST"], cfgHTTPServer.Host)
	assert.Equal(t, true, cfgHTTPServer.LogResponse)

	assert.Equal(t, time.Duration(60)*time.Second, cfgHTTPServer.ReadTimeout())
	assert.Equal(t, time.Duration(60)*time.Second, cfgHTTPServer.WriteTimeout())
	assert.Equal(t, time.Duration(0)*time.Second, cfgHTTPServer.IdleTimeout())
	assert.Equal(t, true, cfgHTTPServer.CloseOnShutdown)
}

func TestHTTPServerEnvErr(t *testing.T) {
	cfgHTTPServer := &pkgenv.HTTPServer{}
	sliceKeys := []string{
		"HTTP_SERVER_PORT",
	}

	checkRequiredFields(t, cfgHTTPServer, sliceKeys)
}

func TestMongoEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["MONGODB_URL"] = "mongodb_host_test"
	mapConfigs["MONGODB_DB_NAME"] = "mongodb_schema_db_name_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgMongo := pkgenv.Mongo{}
	err := pkgenv.ParseCfg(&cfgMongo)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["MONGODB_URL"], cfgMongo.Host)
	assert.Equal(t, mapConfigs["MONGODB_DB_NAME"], cfgMongo.SchemaName)
}

func TestMongoEnvErr(t *testing.T) {
	cfgMongo := &pkgenv.Mongo{}
	sliceKeys := []string{
		"MONGODB_URL",
		"MONGODB_DB_NAME",
	}

	checkRequiredFields(t, cfgMongo, sliceKeys)
}

func TestPostgresEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["POSTGRES_HOST"] = "postgres_host_test"
	mapConfigs["POSTGRES_PORT"] = "postgres_port_test"
	mapConfigs["POSTGRES_USERNAME"] = "postgres_username_test"
	mapConfigs["POSTGRES_PASSWORD"] = "postgres_password_test"
	mapConfigs["POSTGRES_DBNAME"] = "postgres_dbname_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgPostgres := pkgenv.Postgres{}
	err := pkgenv.ParseCfg(&cfgPostgres)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["POSTGRES_HOST"], cfgPostgres.Host)
	assert.Equal(t, mapConfigs["POSTGRES_PORT"], cfgPostgres.Port)
	assert.Equal(t, mapConfigs["POSTGRES_USERNAME"], cfgPostgres.Username)
	assert.Equal(t, mapConfigs["POSTGRES_PASSWORD"], cfgPostgres.Password)
	assert.Equal(t, mapConfigs["POSTGRES_DBNAME"], cfgPostgres.DBName)
	assert.Equal(t, "disable", cfgPostgres.SSLMode)
	assert.Equal(t, "none", cfgPostgres.DBQueryLogLevel)
}

func TestPostgresEnvErr(t *testing.T) {
	cfgPostgres := &pkgenv.Postgres{}
	sliceKeys := []string{
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_USERNAME",
		"POSTGRES_PASSWORD",
		"POSTGRES_DBNAME",
	}

	checkRequiredFields(t, cfgPostgres, sliceKeys)
}

func TestKafkaEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["KAFKA_BOOTSTRAP_SERVERS"] = "kafka_hosts_test1"
	mapConfigs["KAFKA_SASL_USERNAME"] = "kafka_sasl_username_test"
	mapConfigs["KAFKA_SASL_PASSWORD"] = "kafka_sasl_password_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgKafka := pkgenv.Kafka{}
	err := pkgenv.ParseCfg(&cfgKafka)
	assert.NoError(t, err)

	assert.Equal(t, []string{mapConfigs["KAFKA_BOOTSTRAP_SERVERS"]}, cfgKafka.Hosts)
	assert.Equal(t, mapConfigs["KAFKA_SASL_USERNAME"], cfgKafka.SASLUsername)
	assert.Equal(t, mapConfigs["KAFKA_SASL_PASSWORD"], cfgKafka.SASLPassword)
	assert.Equal(t, true, cfgKafka.Enabled)
	assert.Equal(t, false, cfgKafka.LoggerEnabled)
	assert.Equal(t, true, cfgKafka.UseKeyDoubleQuote)
	assert.Equal(t, "", cfgKafka.SASLSecurityProtocol)
	assert.Equal(t, "", cfgKafka.SASLAlgorithm)
	assert.Equal(t, false, cfgKafka.CommitOnError)
	assert.Equal(t, 1, cfgKafka.CommitOnErrorMessagesCount)
	assert.Equal(t, 30*time.Second, cfgKafka.RerunDelay)
	assert.Equal(t, false, cfgKafka.TLSEnabled)
	assert.Equal(t, false, cfgKafka.TLSInsecureSkipVerify)
	assert.Equal(t, "", cfgKafka.TLSClientCertFile)
	assert.Equal(t, "", cfgKafka.TLSClientKeyFile)
	assert.Equal(t, "", cfgKafka.TLSRootCACertFile)
	assert.Equal(t, false, cfgKafka.AutoCreateTopic)
}

func TestKafkaEnvErr(t *testing.T) {
	cfgKafka := &pkgenv.Kafka{}
	sliceKeys := []string{
		"KAFKA_BOOTSTRAP_SERVERS",
		"KAFKA_SASL_USERNAME",
		"KAFKA_SASL_PASSWORD",
	}

	checkRequiredFields(t, cfgKafka, sliceKeys)
}

func TestKafkaProducerEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["KAFKA_PRODUCER_WRITE_TIMEOUT"] = "10ms"
	mapConfigs["KAFKA_PRODUCER_BATCH_SIZE"] = "11"
	mapConfigs["KAFKA_PRODUCER_WRITER_BATCH_TIMEOUT"] = "12s"
	mapConfigs["KAFKA_PRODUCER_MAX_ATTEMPTS"] = "13"
	mapConfigs["KAFKA_PRODUCER_MAX_RETRY"] = "10"
	mapConfigs["KAFKA_PRODUCER_MAX_ATTEMPTS_DELAY"] = "14s"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgKafkaProducer := pkgenv.KafkaProducer{}
	err := pkgenv.ParseCfg(&cfgKafkaProducer)
	assert.NoError(t, err)

	assert.Equal(t, 10*time.Millisecond, cfgKafkaProducer.WriteTimeout)
	assert.Equal(t, 11, cfgKafkaProducer.WriterBatchSize)
	assert.Equal(t, 12*time.Second, cfgKafkaProducer.WriterBatchTimeout)
	assert.Equal(t, 13, cfgKafkaProducer.MaxAttempts)
	assert.Equal(t, 10, cfgKafkaProducer.MaxRetry)
	assert.Equal(t, 14*time.Second, cfgKafkaProducer.MaxAttemptsDelay)
}

func TestKafkaConsumerEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["KAFKA_CONSUMER_GROUP_ID"] = "kafka_consumer_group_id_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgKafkaConsumer := pkgenv.KafkaConsumer{}
	err := pkgenv.ParseCfg(&cfgKafkaConsumer)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["KAFKA_CONSUMER_GROUP_ID"], cfgKafkaConsumer.GroupID)
	assert.Equal(t, 1000, cfgKafkaConsumer.MinBytes)
	assert.Equal(t, 1000000, cfgKafkaConsumer.MaxBytes)
	assert.Equal(t, 10*time.Second, cfgKafkaConsumer.MaxWait)
	assert.Equal(t, -1*time.Second, cfgKafkaConsumer.ReadLagInterval)
	assert.Equal(t, false, cfgKafkaConsumer.CancelIfOneFailed)
	assert.Equal(t, 3, cfgKafkaConsumer.MaxAttempts)
}

func TestKafkaConsumerErr(t *testing.T) {
	os.Clearenv()
	t.Run("not set interface{} values", func(t *testing.T) {
		cfgKafkaConsumer := pkgenv.KafkaConsumer{}
		err := pkgenv.ParseCfg(&cfgKafkaConsumer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "env: required environment variable \"KAFKA_CONSUMER_GROUP_ID\" is not set")
	})
}

func TestRedisEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["REDIS_HOST"] = "redis_host_test"
	mapConfigs["REDIS_DB_NUMBER"] = "1"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgRedis := pkgenv.Redis{}
	err := pkgenv.ParseCfg(&cfgRedis)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["REDIS_HOST"], cfgRedis.Host)
	assert.Equal(t, 1, cfgRedis.DBNumber)
	assert.Equal(t, "", cfgRedis.KeyPrefix)
}

func TestRedisEnvErr(t *testing.T) {
	cfgRedis := &pkgenv.Redis{}
	sliceKeys := []string{
		"REDIS_HOST:redis_host_test",
		"REDIS_DB_NUMBER:1",
	}

	checkRequiredFields(t, cfgRedis, sliceKeys)
}

func TestTraceEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["TRACE_ENVIRONMENT"] = "development"
	mapConfigs["TRACE_AGENT_HOST"] = "localhost"
	mapConfigs["TRACE_AGENT_PORT"] = "6831"
	mapConfigs["TRACE_AGENT_RECONNECTION_INTERVAL"] = "30"
	mapConfigs["TRACE_URL"] = "url"
	mapConfigs["TRACE_USE_AGENT"] = trueString

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.Trace{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["TRACE_URL"], cfg.URL)
	assert.Equal(t, true, cfg.UseAgent)
	assert.Equal(t, mapConfigs["TRACE_ENVIRONMENT"], cfg.Environment)
	assert.Equal(t, mapConfigs["TRACE_AGENT_HOST"], cfg.AgentHost)
	assert.Equal(t, mapConfigs["TRACE_AGENT_PORT"], cfg.AgentPort)
	assert.Equal(t, int64(30), cfg.AgentReconnectingInterval)
}

func TestPagingEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["PAGING_DEFAULT_LIMIT"] = "999"
	mapConfigs["PAGING_MAX_LIMIT"] = "999"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.Paging{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["PAGING_DEFAULT_LIMIT"], strconv.Itoa(int(cfg.DefaultLimit)))
	assert.Equal(t, mapConfigs["PAGING_MAX_LIMIT"], strconv.Itoa(int(cfg.MaxLimit)))
}

func TestMigrationEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["MIGRATION_VERSION"] = "999"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.Migration{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["MIGRATION_VERSION"], strconv.Itoa(int(cfg.Version)))
}

func TestMigrationEnvErr(t *testing.T) {
	cfg := &pkgenv.Migration{}
	sliceKeys := []string{"MIGRATION_VERSION:999"}

	checkRequiredFields(t, cfg, sliceKeys)
}

func TestMigrationWorkflowEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["MIGRATION_WORKFLOW_VERSION"] = "999"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.MigrationWorkflow{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["MIGRATION_WORKFLOW_VERSION"], strconv.Itoa(int(cfg.Version)))
	assert.Equal(t, "schema_migrations_workflow", cfg.SchemaTable)
}

func TestMigrationWorkflowEnvErr(t *testing.T) {
	cfg := &pkgenv.MigrationWorkflow{}
	sliceKeys := []string{"MIGRATION_WORKFLOW_VERSION:999"}

	checkRequiredFields(t, cfg, sliceKeys)
}

func TestWorkflowEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["WORKFLOW_NO_RETRY_ON_ERROR"] = "true"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.Workflow{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, true, cfg.NoRetryOnError)
}

func TestFHIREnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["FHIR_SERVER_API_URL"] = "fhir-host"
	mapConfigs["FHIR_SERVER_SEARCH_URL"] = "host-search"
	mapConfigs["FHIR_SERVER_API_CONSUMER"] = "user"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.FHIR{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["FHIR_SERVER_API_URL"], cfg.HostAPI)
	assert.Equal(t, mapConfigs["FHIR_SERVER_SEARCH_URL"], cfg.HostSearch)
	assert.Equal(t, mapConfigs["FHIR_SERVER_API_CONSUMER"], cfg.User)
	assert.Equal(t, 30*time.Second, cfg.RequestTimeout)
	assert.Equal(t, 1000, cfg.MaxSearchLimit)
}

func TestFHIREnvErr(t *testing.T) {
	cfg := &pkgenv.FHIR{}
	sliceKeys := []string{
		"FHIR_SERVER_API_URL:fhir-host",
		"FHIR_SERVER_SEARCH_URL:host-search",
		"FHIR_SERVER_API_CONSUMER:user",
	}

	checkRequiredFields(t, cfg, sliceKeys)
}

func TestParseCfgWithParsers(t *testing.T) {
	defer os.Clearenv()

	expTestMap := map[string]string{
		"test1": "sets_sub1",
		"test2": "sets_sub2",
	}

	t.Setenv("TEST_MAP", "test1:sets_sub1,test2:sets_sub2")

	cfgParsers := map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(map[string]string{}): func(v string) (interface{}, error) {
			result := map[string]string{}

			items := strings.Split(v, ",")
			for i := 0; i < len(items); i++ {
				pair := strings.Split(items[i], ":")

				result[pair[0]] = pair[1]
			}

			return result, nil
		},
	}
	c := &testConfig{}

	err := pkgenv.ParseCfgWithParsers(c, cfgParsers)
	assert.NoError(t, err)
	assert.Equal(t, expTestMap, c.TestMap)
}

func TestOTPClientEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["OTP_SERVICE_HOST"] = "otp-host"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.OTPClient{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["OTP_SERVICE_HOST"], cfg.Host)
}

func TestOTPClientEnvErr(t *testing.T) {
	cfg := &pkgenv.OTPClient{}
	sliceKeys := []string{
		"OTP_SERVICE_HOST:otp-host",
	}

	checkRequiredFields(t, cfg, sliceKeys)
}

func TestClickHouseEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["CH_HOST"] = "ch_host_test"
	mapConfigs["CH_PORT"] = "ch_port_test"
	mapConfigs["CH_USERNAME"] = "ch_username_test"
	mapConfigs["CH_PASSWORD"] = "ch_password_test"
	mapConfigs["CH_DBNAME"] = "ch_dbname_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgCH := pkgenv.ClickHouse{}
	err := pkgenv.ParseCfg(&cfgCH)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["CH_HOST"], cfgCH.Host)
	assert.Equal(t, mapConfigs["CH_PORT"], cfgCH.Port)
	assert.Equal(t, mapConfigs["CH_USERNAME"], cfgCH.Username)
	assert.Equal(t, mapConfigs["CH_PASSWORD"], cfgCH.Password)
	assert.Equal(t, mapConfigs["CH_DBNAME"], cfgCH.DBName)
	assert.Equal(t, "20s", cfgCH.ConnTimeout)
	assert.Equal(t, false, cfgCH.Compress)
	assert.Equal(t, false, cfgCH.Debug)
	assert.Equal(t, 10, cfgCH.MaxOpenConns)
	assert.Equal(t, 3, cfgCH.MaxIdleConns)
	assert.Equal(t, int64(10), cfgCH.MaxConnLifetime)
}

func TestClickHouseEnvErr(t *testing.T) {
	cfgCH := &pkgenv.ClickHouse{}
	sliceKeys := []string{
		"CH_HOST",
		"CH_PORT",
		"CH_USERNAME",
		"CH_PASSWORD",
		"CH_DBNAME",
	}

	checkRequiredFields(t, cfgCH, sliceKeys)
}

func TestMinioEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["MINIO_HOST"] = "minio_host_test"
	mapConfigs["MINIO_ACCESS_KEY_ID"] = "minio_access_key_id_test"
	mapConfigs["MINIO_SECRET_ACCESS_KEY"] = "minio_secret_access_key_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgMinio := pkgenv.Minio{}
	err := pkgenv.ParseCfg(&cfgMinio)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["MINIO_HOST"], cfgMinio.Host)
	assert.Equal(t, mapConfigs["MINIO_ACCESS_KEY_ID"], cfgMinio.AccessKeyID)
	assert.Equal(t, mapConfigs["MINIO_SECRET_ACCESS_KEY"], cfgMinio.SecretAccessKey)
	assert.Equal(t, "", cfgMinio.Region)
	assert.Equal(t, "", cfgMinio.Token)
	assert.Equal(t, false, cfgMinio.UseSSL)
	assert.Equal(t, int64(10), cfgMinio.ResponseExpiredMin)
}

func TestMinioEnvErr(t *testing.T) {
	cfgMinio := &pkgenv.Minio{}
	sliceKeys := []string{
		"MINIO_HOST",
		"MINIO_ACCESS_KEY_ID",
		"MINIO_SECRET_ACCESS_KEY",
	}

	checkRequiredFields(t, cfgMinio, sliceKeys)
}

func TestAWSEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["AWS_REGION"] = "aws_region_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgAWS := pkgenv.AWS{}
	err := pkgenv.ParseCfg(&cfgAWS)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["AWS_REGION"], cfgAWS.Region)
	assert.Equal(t, "", cfgAWS.AccessKeyID)
	assert.Equal(t, "", cfgAWS.SecretAccessKey)
	assert.Equal(t, "", cfgAWS.Token)
}

func TestAWSS3Env(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["AWSS3_EXPIRED_MIN"] = "10"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.AWSS3{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), cfg.ResponseExpiredMin)
}

func TestMediaStorageEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["MEDIA_STORAGE_TYPE"] = "media_storage_type_test"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgMediaStorage := pkgenv.MediaStorage{}
	err := pkgenv.ParseCfg(&cfgMediaStorage)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["MEDIA_STORAGE_TYPE"], cfgMediaStorage.UseStorage)
	assert.Equal(t, "", cfgMediaStorage.PublicPresignURL)
}

func TestMediaStorageEnvErr(t *testing.T) {
	cfgMediaStorage := &pkgenv.MediaStorage{}
	sliceKeys := []string{
		"MEDIA_STORAGE_TYPE",
	}

	checkRequiredFields(t, cfgMediaStorage, sliceKeys)
}

func TestMediaStorageClientEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["MEDIA_STORAGE_SERVICE_HOST"] = "mediastorage-host"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.MediaStorageClient{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["MEDIA_STORAGE_SERVICE_HOST"], cfg.Host)
}

func TestMediaStorageClientEnvErr(t *testing.T) {
	cfg := &pkgenv.MediaStorageClient{}
	sliceKeys := []string{
		"MEDIA_STORAGE_SERVICE_HOST:mediastorage-host",
	}

	checkRequiredFields(t, cfg, sliceKeys)
}

func TestElasticEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["ELASTIC_HOSTS"] = "elastic_host_test1,elastic_host_test2"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgES := pkgenv.Elastic{}
	err := pkgenv.ParseCfg(&cfgES)
	assert.NoError(t, err)

	assert.Equal(t, []string{"elastic_host_test1", "elastic_host_test2"}, cfgES.Hosts)
	assert.Equal(t, "", cfgES.User)
	assert.Equal(t, "", cfgES.Pass)
	assert.Equal(t, true, cfgES.EnableRetryOnTimeout)
	assert.Equal(t, 5, cfgES.MaxRetries)
	assert.Equal(t, false, cfgES.CompressRequestBody)
	assert.Equal(t, false, cfgES.EnableMetrics)
	assert.Equal(t, false, cfgES.Debug)
	assert.Equal(t, false, cfgES.EnableCompatibilityMode)
	assert.Equal(t, 10, cfgES.MaxIdleConnsPerHost)
	assert.Equal(t, 60, cfgES.ResponseHeaderTimeoutSec)
}

func TestElasticEnvErr(t *testing.T) {
	cfgES := &pkgenv.Elastic{}
	sliceKeys := []string{
		"ELASTIC_HOSTS",
	}

	checkRequiredFields(t, cfgES, sliceKeys)
}

func TestNotifyEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["NOTIFY_MAX_ATTEMPTS"] = "10"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.Notify{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, int64(10), cfg.MaxAttempts)
}

func TestNotifyPostmarkEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["NOTIFY_POSTMARK_SERVER_TOKEN"] = "serverToken"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.NotifyPostmark{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["NOTIFY_POSTMARK_SERVER_TOKEN"], cfg.ServerToken)
	assert.Equal(t, "", cfg.AccountToken)
}

func TestNotifyTiniyoEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["NOTIFY_TINIYO_HOST_API"] = "hostAPI"
	mapConfigs["NOTIFY_TINIYO_AUTH_ID"] = "authID"
	mapConfigs["NOTIFY_TINIYO_AUTH_TOKEN"] = "authToken"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.NotifyTiniyo{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["NOTIFY_TINIYO_HOST_API"], cfg.TiniyoHostAPI)
	assert.Equal(t, mapConfigs["NOTIFY_TINIYO_AUTH_ID"], cfg.TiniyoAuthID)
	assert.Equal(t, mapConfigs["NOTIFY_TINIYO_AUTH_TOKEN"], cfg.TiniyoAuthToken)
	assert.Equal(t, time.Second*30, cfg.TiniyoRequestTimeout)
}

func TestNotifyTwilioEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["NOTIFY_TWILIO_ACCOUNT_SID"] = "account-sid"
	mapConfigs["NOTIFY_TWILIO_AUTH_TOKEN"] = "auth-token"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.NotifyTwilio{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["NOTIFY_TWILIO_ACCOUNT_SID"], cfg.AccountSID)
	assert.Equal(t, mapConfigs["NOTIFY_TWILIO_AUTH_TOKEN"], cfg.AuthToken)
	assert.Equal(t, "", cfg.CallbackURL)
	assert.Equal(t, "", cfg.MessagingServiceSID)
}

func TestNotificationClientEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["NOTIFICATION_SERVICE_HOST"] = "http://test.host/notification"
	mapConfigs["NOTIFICATION_PROVIDER_EMAIL"] = "postmark"
	mapConfigs["NOTIFICATION_PROVIDER_SMS"] = "twilio"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgNC := pkgenv.NotificationClient{}
	err := pkgenv.ParseCfg(&cfgNC)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["NOTIFICATION_SERVICE_HOST"], cfgNC.Host)
	assert.Equal(t, mapConfigs["NOTIFICATION_PROVIDER_EMAIL"], cfgNC.ProviderEmail)
	assert.Equal(t, mapConfigs["NOTIFICATION_PROVIDER_SMS"], cfgNC.ProviderSMS)
}

func TestNotificationClientErr(t *testing.T) {
	cfgNC := &pkgenv.NotificationClient{}
	sliceKeys := []string{
		"NOTIFICATION_SERVICE_HOST",
	}

	checkRequiredFields(t, cfgNC, sliceKeys)
}

func TestKeycloakEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["KEYCLOAK_HOST_API"] = "http://test.host/keycloak"
	mapConfigs["KEYCLOAK_REALM"] = "ksa-ehealth"
	mapConfigs["KEYCLOAK_CLIENT_SECRET"] = "secret"
	mapConfigs["KEYCLOAK_USERNAME"] = "username"
	mapConfigs["KEYCLOAK_PASSWORD"] = "password"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfgK := pkgenv.KeycloakClient{}
	err := pkgenv.ParseCfg(&cfgK)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["KEYCLOAK_HOST_API"], cfgK.HostAPI)
	assert.Equal(t, mapConfigs["KEYCLOAK_REALM"], cfgK.Realm)
	assert.Equal(t, mapConfigs["KEYCLOAK_CLIENT_SECRET"], cfgK.ClientSecret)
	assert.Equal(t, mapConfigs["KEYCLOAK_USERNAME"], cfgK.Username)
	assert.Equal(t, mapConfigs["KEYCLOAK_PASSWORD"], cfgK.Password)
}

func TestKeycloakErr(t *testing.T) {
	cfgNC := &pkgenv.KeycloakClient{}
	sliceKeys := []string{
		"KEYCLOAK_HOST_API",
		"KEYCLOAK_REALM",
		"KEYCLOAK_CLIENT_SECRET",
		"KEYCLOAK_USERNAME",
		"KEYCLOAK_PASSWORD",
	}

	checkRequiredFields(t, cfgNC, sliceKeys)
}

func TestGeodictionaryClientEnv(t *testing.T) {
	defer os.Clearenv()

	mapConfigs := make(map[string]string)
	mapConfigs["GEODICTIONARY_SERVICE_HOST"] = "geodictionary-host"

	for key, value := range mapConfigs {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}

	cfg := pkgenv.GeodictionaryClient{}
	err := pkgenv.ParseCfg(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, mapConfigs["GEODICTIONARY_SERVICE_HOST"], cfg.Host)
}

func TestGeodictionaryClientEnvErr(t *testing.T) {
	cfg := &pkgenv.GeodictionaryClient{}
	sliceKeys := []string{
		"GEODICTIONARY_SERVICE_HOST:geodictionary-host",
	}

	checkRequiredFields(t, cfg, sliceKeys)
}
