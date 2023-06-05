package testutil

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/testutil/config"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type DockerMinioContainer struct {
	*DockerTestContainer
	Username string
	Password string
	mClient  *minio.Client
}

func NewDockerMinioContainer(cfg *config.TestContainerConfig) *DockerMinioContainer {
	c := &DockerMinioContainer{
		DockerTestContainer: &DockerTestContainer{
			Name:                GenerateCorrectName("minio"),
			Repository:          "bitnami/minio",
			Tag:                 "2022.4.16",
			Port:                "9000",
			TestContainerConfig: cfg,
			Host:                "localhost",
			activeConn:          0,
		},
		Username: "admin",
		Password: "password",
	}
	c.Env = []string{
		fmt.Sprintf("MINIO_ROOT_USER=%v", c.Username),
		fmt.Sprintf("MINIO_ROOT_PASSWORD=%v", c.Password),
	}

	return c
}

func (c *DockerMinioContainer) GetDockerRunOptions() *dockertest.RunOptions {
	return c.DockerTestContainer.GetDockerRunOptions()
}

func (c *DockerMinioContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerMinioContainer {
	c.DockerTestContainer.RunWithOptions(opt)
	return c
}

func (c *DockerMinioContainer) ConnectToMino(bucketName string) *DockerMinioContainer {
	ctx := context.Background()
	endpoint := fmt.Sprintf("%s:%s", c.Host, c.Port)
	checkMinioURL := fmt.Sprintf("http://%s/minio/health/live", endpoint)
	httpClient := &http.Client{}

	if c.mClient != nil {
		return c
	}

	if err := c.Pool.Retry(func() error {
		reqMinio, err := http.NewRequestWithContext(ctx, http.MethodGet, checkMinioURL, http.NoBody)
		if err != nil {
			return err
		}
		resp, err := httpClient.Do(reqMinio)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK")
		}

		minioClient, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(c.Username, c.Password, ""),
			Secure: false,
		})
		if err != nil {
			return err
		}

		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}

		c.mClient = minioClient

		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not connect to minio in docker with port %v", c.Port).
			LogFatal()
		return nil
	}
	c.activeConn++

	return c
}

func (c *DockerMinioContainer) GetClient() *minio.Client {
	return c.mClient
}

func (c *DockerMinioContainer) Close() {
	c.closeConnection()
}
