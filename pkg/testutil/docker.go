package testutil

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/testutil/config"
	"sync"
)

var (
	once     sync.Once
	instance *DockerContainerTestUtils
)

type DockerContainerTestUtils struct {
	pg         *DockerPGContainer
	minio      *DockerMinioContainer
	zk         *DockerZKContainer
	k          *DockerKafkaContainer
	elastic    *DockerElasticContainer
	ch         *DockerCHContainer
	rc         *DockerRedisContainer
	containers map[string]*DockerTestContainer
	cfg        *config.TestContainerConfig
}

func NewDockerUtilInstance() *DockerContainerTestUtils {
	if instance == nil {
		once.Do(initNewInstance)
	}

	return instance
}

func initNewInstance() {
	ctx := context.Background()
	log.Log(logger.NewEventF(ctx, logger.LevelDebug, "init new instance"))

	pool, err := dockertest.NewPool("")

	if err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "failed to create new docker pool").LogError()
	}

	cfg := &config.TestContainerConfig{
		Pool: pool,
	}

	instance = &DockerContainerTestUtils{
		containers: make(map[string]*DockerTestContainer),
		cfg:        cfg,
	}
	instance.cfg.FromEnv()
}

// RunDockerContainer - run specific docker container with parameters:
// n - container name
// r - image repository
// t - image tag
// p - container port
// env - container environment variables
func (u *DockerContainerTestUtils) RunDockerContainer(n, r, t, p string, env []string) *DockerTestContainer {
	cont, ok := u.containers[n]
	if ok {
		return cont
	}

	cont = &DockerTestContainer{
		Name:                n,
		Repository:          r,
		Tag:                 t,
		Port:                p,
		Env:                 env,
		TestContainerConfig: u.cfg,
	}
	cont = cont.RunWithOptions(cont.GetDockerRunOptions())
	u.containers[n] = cont

	return cont
}

func (u *DockerContainerTestUtils) InitPG() *DockerPGContainer {
	if u.pg != nil {
		return u.pg
	}

	pg := NewDockerPGContainer(u.cfg)
	pg = pg.RunWithOptions(pg.GetDockerRunOptions())
	u.pg = pg

	return u.pg
}

func (u *DockerContainerTestUtils) InitMinio() *DockerMinioContainer {
	if u.minio != nil {
		return u.minio
	}

	mp := NewDockerMinioContainer(u.cfg)
	mp = mp.RunWithOptions(mp.GetDockerRunOptions())
	u.minio = mp

	return u.minio
}

func (u *DockerContainerTestUtils) InitElastic() *DockerElasticContainer {
	if u.minio != nil {
		return u.elastic
	}

	ec := NewDockerElasticContainer(u.cfg)
	ec = ec.RunWithOptions(ec.GetDockerRunOptions())
	u.elastic = ec

	return u.elastic
}

func (u *DockerContainerTestUtils) InitCH() *DockerCHContainer {
	if u.ch != nil {
		return u.ch
	}

	ch := NewDockerCHContainer(u.cfg)
	ch = ch.RunWithOptions(ch.GetDockerRunOptions())
	u.ch = ch

	return u.ch
}

func (u *DockerContainerTestUtils) InitRedis() *DockerRedisContainer {
	if u.rc != nil {
		return u.rc
	}

	rc := NewDockerRedisContainer(u.cfg)
	u.rc = rc.RunWithOptions(rc.GetDockerRunOptions())

	return u.rc
}

func (u *DockerContainerTestUtils) InitZK() *DockerZKContainer {
	if u.zk != nil {
		return u.zk
	}

	zk := NewDockerZKContainer(u.cfg)
	zk = zk.RunWithOptions(zk.GetDockerRunOptions())
	u.zk = zk

	return u.zk
}

type KafkaParams struct {
	DefaultTopic   []string
	ZKHost, ZKPort string
}

func (u *DockerContainerTestUtils) InitKafka(param KafkaParams) *DockerKafkaContainer {
	if u.k != nil {
		return u.k
	}

	k := NewDockerKafkaContainer(u.cfg, param)
	k = k.RunWithOptions(k.GetDockerRunOptions())
	u.k = k

	return u.k
}

func (u *DockerContainerTestUtils) GetContainerByName(name string) (*DockerTestContainer, error) {
	cont, ok := u.containers[name]
	if !ok {
		return nil, fmt.Errorf("failed to get docker container by name %v", name)
	}

	return cont, nil
}
