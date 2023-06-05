package minio_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/mediameta/provider"
	pkgMinio "kafka-polygon/pkg/mediameta/provider/minio"
	"kafka-polygon/pkg/testutil"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/suite"
)

type minioTestSuite struct {
	suite.Suite
	mCont *testutil.DockerMinioContainer
	cl    *minio.Client

	bucketName string
	objects    []string
}

func TestRepoTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(minioTestSuite))
}

func (ms *minioTestSuite) SetupSuite() {
	bName := "dev-mediastorage"
	ms.mCont = testutil.NewDockerUtilInstance().
		InitMinio().
		ConnectToMino(bName)
	ms.cl = ms.mCont.GetClient()
	ms.bucketName = bName
	ms.objects = []string{
		"test-1",
		"test-2",
		"test-3",
		"test-4",
	}
}

func (ms *minioTestSuite) TearDownSuite() {
	ms.mCont.Close()
}

func (ms *minioTestSuite) TestMinio() {
	ctx := context.Background()

	pm := pkgMinio.New(&pkgMinio.Options{
		Endpoint:        fmt.Sprintf("%s:%s", ms.mCont.Host, ms.mCont.Port),
		AccessKeyID:     "admin",
		SecretAccessKey: "password",
		Token:           "",
		Region:          "",
		UseSSL:          false,
	})
	err := pm.Connected(ctx)
	ms.NoError(err)

	exists, err := pm.BucketExists(ctx, ms.bucketName)
	ms.NoError(err)
	ms.Equal(true, exists)
}

func (ms *minioTestSuite) TestPing() {
	ctx := context.Background()

	pm := pkgMinio.New(&pkgMinio.Options{
		Endpoint:        fmt.Sprintf("%s:%s", ms.mCont.Host, ms.mCont.Port),
		AccessKeyID:     "admin",
		SecretAccessKey: "password",
		Token:           "",
		Region:          "",
		UseSSL:          false,
	})
	err := pm.Connected(ctx)
	ms.NoError(err)

	err = pm.Ping(ctx)
	ms.NoError(err)
}

func (ms *minioTestSuite) TestPingError() {
	ctx := context.Background()

	pm := pkgMinio.New(&pkgMinio.Options{
		Endpoint:        fmt.Sprintf("%s:%s", ms.mCont.Host, ms.mCont.Port),
		AccessKeyID:     "admin1",
		SecretAccessKey: "password1",
		Token:           "",
		Region:          "",
		UseSSL:          false,
	})
	err := pm.Connected(ctx)
	ms.NoError(err)

	err = pm.Ping(ctx)
	ms.Error(err)
	ms.Equal(cerror.KindMinioOther, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioPutObject() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	buff := bytes.NewBuffer([]byte(`Hello World!! test1`))
	err := pm.PutObject(ctx, ms.bucketName, ms.objects[0], buff, int64(buff.Len()), provider.PutOptions{
		UserMetadata: provider.UserMetadata{"metadata-key": "metadata-value"},
		UserTags:     provider.UserTags{"tag-key": "tag-value"},
		ContentType:  "text/plain",
	})
	ms.NoError(err)

	exists, err := pm.ObjectExists(ctx, ms.bucketName, ms.objects[0])
	ms.NoError(err)
	ms.Equal(true, exists)

	err = pm.RemoveObject(ctx, ms.bucketName, ms.objects[0])
	ms.NoError(err)

	err = pm.PutObject(ctx, "fake_bucket", ms.objects[0], buff, int64(buff.Len()), provider.PutOptions{
		UserMetadata: provider.UserMetadata{"metadata-key": "metadata-value"},
		UserTags:     provider.UserTags{"tag-key": "tag-value"},
		ContentType:  "text/plain",
	})
	ms.Error(err)
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioRemoveObject() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	buff := bytes.NewBuffer([]byte(`Hello World!! test2`))
	err := pm.PutObject(ctx, ms.bucketName, ms.objects[1], buff, int64(buff.Len()), provider.PutOptions{
		UserMetadata: provider.UserMetadata{"metadata-key": "metadata-value"},
		UserTags:     map[string]string{"tag-key": "tag-value"},
		ContentType:  "text/plain",
	})
	ms.NoError(err)

	err = pm.RemoveObject(ctx, ms.bucketName, ms.objects[1])
	ms.NoError(err)

	exists, err := pm.ObjectExists(ctx, ms.bucketName, ms.objects[1])
	ms.Error(err)
	ms.Equal(false, exists)
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioGetObject() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	expected := struct {
		um provider.UserMetadata
		ct string
	}{
		um: provider.UserMetadata{"metadata-key": "metadata-value"},
		ct: "text/plain",
	}

	buff := bytes.NewBuffer([]byte(`Hello World!! test3`))
	err := pm.PutObject(ctx, ms.bucketName, ms.objects[2], buff, int64(buff.Len()), provider.PutOptions{
		UserMetadata: expected.um,
		ContentType:  expected.ct,
	})
	ms.NoError(err)

	obj, err := pm.GetObject(ctx, ms.bucketName, ms.objects[2])
	ms.NotNil(obj)
	ms.NoError(err)

	b := streamToByte(obj.Body)

	defer func() {
		_ = obj.Body.Close()
	}()

	ms.Equal(buff.Len(), len(b))
	ms.Equal(expected.ct, obj.ContentType)
	ms.Equal(provider.UserMetadata{"Metadata-Key": "metadata-value"}, obj.UserMetadata)
}

func (ms *minioTestSuite) TestMinioGetObjectNotFound() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	obj, err := pm.GetObject(ctx, ms.bucketName, "not_found_obj")
	ms.Error(err)
	ms.Nil(obj)
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioObjectTagging() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	expected := struct {
		ut provider.UserTags
		ct string
	}{
		ut: provider.UserTags{"tag-key": "tag-value"},
		ct: "text/plain",
	}

	buff := bytes.NewBuffer([]byte(`Hello World!! test3`))
	err := pm.PutObject(ctx, ms.bucketName, ms.objects[2], buff, int64(buff.Len()), provider.PutOptions{
		UserTags:    expected.ut,
		ContentType: expected.ct,
	})
	ms.NoError(err)

	ut, err := pm.GetObjectTagging(ctx, ms.bucketName, ms.objects[2])
	ms.NoError(err)
	ms.Equal(provider.UserTags{"tag-key": "tag-value"}, ut)

	ut, err = pm.GetObjectTagging(ctx, ms.bucketName, "not_found_obj")
	ms.Error(err)
	ms.Equal(provider.UserTags{}, ut)
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioUpdateUserMetadata() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	exists, err := pm.ObjectExists(ctx, ms.bucketName, ms.objects[2])
	ms.NoError(err)
	ms.Equal(true, exists)

	uInfo, err := pm.UpdateUserMetadata(ctx, ms.bucketName,
		ms.objects[2],
		provider.UserMetadata{"new-key": "new-value", "content-type": "text/plain"})
	ms.NoError(err)
	ms.Equal(uInfo.Bucket, ms.bucketName)
	ms.Equal(uInfo.Key, ms.objects[2])

	obj, err := pm.GetObject(ctx, ms.bucketName, ms.objects[2])
	ms.NoError(err)

	defer func() {
		_ = obj.Body.Close()
	}()

	ms.Equal("text/plain", obj.ContentType)
	ms.Equal(
		provider.UserMetadata{"New-Key": "new-value"},
		obj.UserMetadata)

	err = pm.RemoveObject(ctx, ms.bucketName, ms.objects[2])
	ms.NoError(err)

	uInfo, err = pm.UpdateUserMetadata(ctx, ms.bucketName,
		"not_found_obj",
		provider.UserMetadata{"new-key": "new-value", "content-type": "text/plain"})
	ms.Error(err)
	ms.Nil(uInfo)
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioUpdateUserTags() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	buff := bytes.NewBuffer([]byte(`Hello World!! test4`))
	err := pm.PutObject(ctx, ms.bucketName, ms.objects[3], buff, int64(buff.Len()), provider.PutOptions{
		UserMetadata: provider.UserMetadata{"metadata-key": "metadata-value"},
		UserTags:     provider.UserTags{"tag-key": "tag-value"},
		ContentType:  "text/plain",
	})
	ms.NoError(err)

	exists, err := pm.ObjectExists(ctx, ms.bucketName, ms.objects[3])
	ms.NoError(err)
	ms.Equal(true, exists)

	uInfo, err := pm.UpdateUserTags(ctx, ms.bucketName,
		ms.objects[3],
		provider.UserTags{"tag-key-1": "tag-value-1"})
	ms.NoError(err)
	ms.Equal(uInfo.Bucket, ms.bucketName)
	ms.Equal(uInfo.Key, ms.objects[3])

	obj, err := pm.GetObject(ctx, ms.bucketName, ms.objects[3])
	ms.NoError(err)

	defer func() {
		_ = obj.Body.Close()
	}()
	ms.Equal(
		provider.UserMetadata{"Metadata-Key": "metadata-value"},
		obj.UserMetadata)

	ut, err := pm.GetObjectTagging(ctx, ms.bucketName, ms.objects[3])
	ms.NoError(err)
	ms.Equal(provider.UserTags{"tag-key-1": "tag-value-1"}, ut)

	uInfo, err = pm.UpdateUserTags(ctx, ms.bucketName,
		ms.objects[3],
		provider.UserTags{"tag-key-1": "tag-value-new-value", "tag-key-2": "tag-value-2"})
	ms.NoError(err)
	ms.Equal(uInfo.Bucket, ms.bucketName)
	ms.Equal(uInfo.Key, ms.objects[3])

	ut, err = pm.GetObjectTagging(ctx, ms.bucketName, ms.objects[3])
	ms.NoError(err)
	ms.Equal(provider.UserTags{"tag-key-1": "tag-value-new-value", "tag-key-2": "tag-value-2"}, ut)

	err = pm.RemoveObject(ctx, ms.bucketName, ms.objects[3])
	ms.NoError(err)

	uInfo, err = pm.UpdateUserTags(ctx, ms.bucketName,
		"not_found_obj",
		provider.UserTags{"tag-key-1": "tag-value-1"})
	ms.Error(err)
	ms.Nil(uInfo)
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioObjectExists() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	exists, err := pm.ObjectExists(ctx, ms.bucketName, ms.objects[2])
	ms.NoError(err)
	ms.Equal(true, exists)

	exists, err = pm.ObjectExists(ctx, ms.bucketName, ms.objects[1])
	ms.Error(err)
	ms.Equal(false, exists)
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioBucketExists() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	exists, err := pm.BucketExists(ctx, ms.bucketName)
	ms.NoError(err)
	ms.Equal(true, exists)

	exists, err = pm.BucketExists(ctx, "not-found-bucket")
	ms.NoError(err)
	ms.Equal(false, exists)

	exists, err = pm.BucketExists(ctx, "no")
	ms.Error(err)
	ms.Equal(false, exists)
	ms.Equal(cerror.KindMinioOther, cerror.ErrKind(err))
}

func (ms *minioTestSuite) TestMinioPresign() {
	ctx := context.Background()
	pm := pkgMinio.New(nil)
	pm.SetClient(ms.cl)

	url, err := pm.PresignGetObject(ctx, ms.bucketName, ms.objects[0], nil)
	ms.NoError(err)
	ms.NotNil(url)

	url, err = pm.PresignGetObject(ctx, "not-found-bucket", "not-found-obj", nil)
	ms.Error(err)
	ms.Nil(url)
	ms.Equal("minio::PresignedGetObject The specified bucket does not exist", err.Error())
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))

	url, err = pm.PresignHeadObject(ctx, ms.bucketName, ms.objects[0], nil)
	ms.NoError(err)
	ms.NotNil(url)

	url, err = pm.PresignHeadObject(ctx, "not-found-bucket", "not-found-obj", nil)
	ms.Error(err)
	ms.Nil(url)
	ms.Equal("minio::PresignedHeadObject The specified bucket does not exist", err.Error())
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))

	url, err = pm.PresignPutObject(ctx, ms.bucketName, ms.objects[0], nil)
	ms.NoError(err)
	ms.NotNil(url)

	url, err = pm.PresignPutObject(ctx, "not-found-bucket", "not-found-obj", nil)
	ms.Error(err)
	ms.Nil(url)
	ms.Equal("minio::PresignedPutObject The specified bucket does not exist", err.Error())
	ms.Equal(cerror.KindMinioNotExist, cerror.ErrKind(err))
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(stream)

	return buf.Bytes()
}
