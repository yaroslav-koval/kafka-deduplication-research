package redis_test

import (
	"context"
	"kafka-polygon/pkg/broker/store"
	stRedis "kafka-polygon/pkg/broker/store/redis"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"testing"

	"github.com/go-redis/redis/v8"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
)

var _bgCtx = context.Background()

type storeTestSuite struct {
	suite.Suite
	rcCont *testutil.DockerRedisContainer
	rc     *redis.Client
	rcCfg  *env.Redis
	st     *stRedis.Store
}

func TestStoreTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(storeTestSuite))
}

func (s *storeTestSuite) SetupSuite() {
	rcCont := testutil.NewDockerUtilInstance().InitRedis().ConnectToRedis()
	s.rcCont = rcCont
	s.rcCfg = rcCont.GetRedisEnvConfig()
	s.rcCfg.KeyPrefix = "test_broker"
	s.rc = rcCont.GetClient()

	s.st = stRedis.NewStore(s.rc, stRedis.Settings{
		KeyPrefix: s.rcCfg.KeyPrefix,
		TTL:       s.rcCfg.TTL,
	})
}

func (s *storeTestSuite) TearDownTest() {
}

func (s *storeTestSuite) TearDownSuite() {
	_ = s.rc.Close()
	s.rcCont.Close()
}

func (s *storeTestSuite) TestGetEventInfoByID() {
	uID := uuid.NewV4()

	expectedData := store.EventProcessData{Status: "test"}

	_, err := s.st.GetEventInfoByID(_bgCtx, uID.String())
	s.Error(err)

	cErr, ok := err.(*cerror.CError)
	s.True(ok)
	s.Equal(cerror.KindNotExist, cErr.Kind())

	err = s.st.PutEventInfo(_bgCtx, uID.String(), expectedData)
	s.NoError(err)

	data, err := s.st.GetEventInfoByID(_bgCtx, uID.String())
	s.NoError(err)
	s.Equal(expectedData, data)
}

func (s *storeTestSuite) TestPutEventInfo() {
	uID := uuid.NewV4()

	expectedData := store.EventProcessData{Status: "test"}

	err := s.st.PutEventInfo(_bgCtx, uID.String(), expectedData)
	s.NoError(err)

	data, err := s.st.GetEventInfoByID(_bgCtx, uID.String())
	s.NoError(err)
	s.Equal(expectedData, data)
}
