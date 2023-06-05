package es_test

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/db/elastic/es"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/suite"
)

type esConnectionTestSuite struct {
	suite.Suite
	ctx    context.Context
	cl     *elasticsearch.Client
	esCont *testutil.DockerElasticContainer
	esConn *es.Connection
	esCfg  *env.Elastic
}

func TestRepoTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(esConnectionTestSuite))
}

func (s *esConnectionTestSuite) SetupSuite() {
	s.ctx = context.Background()
	esCont := testutil.NewDockerUtilInstance().InitElastic().ConnectToElastic()
	s.esCont = esCont
	s.cl = esCont.GetClient()
	s.esCfg = esCont.GetElasticEnvConfig()
	s.esConn = es.NewConnection(s.ctx, s.esCfg)
}

func (s *esConnectionTestSuite) TearDownSuite() {
	s.esCont.Close()
}

func (s *esConnectionTestSuite) TestConnectAndClose() {
	s.NotNil(s.esConn)
	s.Nil(s.esConn.DB())

	err := s.esConn.Connect()
	s.NoError(err)
	s.NotNil(s.esConn.DB())

	res, err := s.esConn.DB().Info()
	s.NoError(err)

	defer func() {
		_ = res.Body.Close()
	}()

	var info es.Info

	err = json.NewDecoder(res.Body).Decode(&info)
	s.NoError(err)
	s.Equal("7.7.0", info.Version.Number)

	err = s.esConn.Close()
	s.NoError(err)
}

func (s *esConnectionTestSuite) TestConnectError() {
	conn := es.NewConnection(s.ctx, &env.Elastic{Hosts: []string{"http://fake_host:1000"}})
	err := conn.Connect()
	s.Error(err)
	s.Equal(cerror.KindElasticOther, cerror.ErrKind(err))
}
