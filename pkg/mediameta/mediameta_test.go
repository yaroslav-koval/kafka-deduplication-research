package mediameta_test

import (
	"bytes"
	"context"
	"io"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/mediameta"
	"kafka-polygon/pkg/mediameta/provider"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var (
	bgCtx      = context.Background()
	provType   = provider.Minio
	bucketName = "dev-bucket-name"
	objName    = "dev-object-name"
)

type MockedProviderMinio struct {
	mock.Mock
}

func (mpm *MockedProviderMinio) Connected(ctx context.Context) error {
	args := mpm.Called(ctx)

	return args.Error(0)
}

func (mpm *MockedProviderMinio) Type() provider.Type {
	args := mpm.Called()
	t := args.String(0)

	return provider.Type(t)
}

func (mpm *MockedProviderMinio) Ping(ctx context.Context) error {
	args := mpm.Called(ctx)

	return args.Error(0)
}

func (mpm *MockedProviderMinio) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := mpm.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

func (mpm *MockedProviderMinio) ObjectExists(ctx context.Context, bucketName, objName string) (bool, error) {
	args := mpm.Called(ctx, bucketName, objName)
	return args.Bool(0), args.Error(1)
}

func (mpm *MockedProviderMinio) PutObject(ctx context.Context,
	bucketName, objName string, r io.Reader, rSize int64, opts provider.PutOptions) error {
	args := mpm.Called(ctx, bucketName, objName, r, rSize, opts)
	return args.Error(0)
}

func (mpm *MockedProviderMinio) UpdateUserMetadata(
	ctx context.Context,
	bucketName, objName string,
	um provider.UserMetadata) (*provider.UserMetadataInfo, error) {
	args := mpm.Called(ctx, bucketName, objName, um)
	return args.Get(0).(*provider.UserMetadataInfo), args.Error(1)
}

func (mpm *MockedProviderMinio) UpdateUserTags(
	ctx context.Context,
	bucketName, objName string,
	um provider.UserTags) (*provider.UserMetadataInfo, error) {
	args := mpm.Called(ctx, bucketName, objName, um)
	return args.Get(0).(*provider.UserMetadataInfo), args.Error(1)
}

func (mpm *MockedProviderMinio) RemoveObject(ctx context.Context, bucketName, objName string) error {
	args := mpm.Called(ctx, bucketName, objName)

	return args.Error(0)
}

func (mpm *MockedProviderMinio) PresignGetObject(ctx context.Context,
	bucketName, objName string, expireMin *int64) (*string, error) {
	args := mpm.Called(ctx, bucketName, objName, expireMin)
	bn := args.String(0)

	return converto.StringPointer(bn), args.Error(1)
}

func (mpm *MockedProviderMinio) PresignHeadObject(ctx context.Context,
	bucketName, objName string, expireMin *int64) (*string, error) {
	args := mpm.Called(ctx, bucketName, objName, expireMin)
	return converto.StringPointer(args.String(0)), args.Error(1)
}

func (mpm *MockedProviderMinio) PresignPutObject(ctx context.Context,
	bucketName, objName string, expireMin *int64) (*string, error) {
	args := mpm.Called(ctx, bucketName, objName, expireMin)

	return converto.StringPointer(args.String(0)), args.Error(1)
}

func (mpm *MockedProviderMinio) GetObject(ctx context.Context, bucketName, objName string) (*provider.GetObject, error) {
	args := mpm.Called(ctx, bucketName, objName)
	r := io.NopCloser(bytes.NewReader([]byte(args.String(0))))

	return &provider.GetObject{
		ETag:        "etag",
		ContentType: "text/plain",
		Body:        r,
	}, args.Error(1)
}

func (mpm *MockedProviderMinio) GetObjectTagging(ctx context.Context, bucketName, objName string) (provider.UserTags, error) {
	args := mpm.Called(ctx, bucketName, objName)

	return args.Get(0).(provider.UserTags), args.Error(1)
}

func TestMediaMetaConnected(t *testing.T) {
	t.Parallel()

	mockProv := &MockedProviderMinio{}
	mockProv.On("Type").Return(provType.ToString())
	mockProv.On("Connected", bgCtx).Return(nil)

	mm, err := mediameta.New(bgCtx, provType, mockProv)
	require.NoError(t, err)

	err = mm.Connected(bgCtx)
	require.NoError(t, err)
}

func TestMediaMetaPresign(t *testing.T) {
	t.Parallel()

	mockProv := &MockedProviderMinio{}
	mockProv.On("Type").Return(provType.ToString())
	mockProv.On("BucketExists", bgCtx, bucketName).Return(true, nil)
	mockProv.On("ObjectExists", bgCtx, bucketName, objName).Return(true, nil)
	mockProv.On("PresignGetObject", bgCtx, bucketName, objName, (*int64)(nil)).
		Return("http://localhost:900/pre-sign-get-object", nil)
	mockProv.On("PresignHeadObject", bgCtx, bucketName, objName, (*int64)(nil)).
		Return("http://localhost:900/pre-sign-head-object", nil)
	mockProv.On("PresignPutObject", bgCtx, bucketName, objName, (*int64)(nil)).
		Return("http://localhost:900/pre-sign-put-object", nil)

	mm, err := mediameta.New(bgCtx, provType, mockProv)
	require.NoError(t, err)

	expected := []struct {
		method, url string
		fn          func(bucketName, objName string) (*string, error)
	}{
		{
			method: "PresignGetObject",
			url:    "http://localhost:900/pre-sign-get-object",
			fn: func(bucketName, objName string) (*string, error) {
				return mm.PresignGetObject(bgCtx, bucketName, objName, nil)
			},
		},
		{
			method: "PresignHeadObject",
			url:    "http://localhost:900/pre-sign-head-object",
			fn: func(bucketName, objName string) (*string, error) {
				return mm.PresignHeadObject(bgCtx, bucketName, objName, nil)
			},
		},
		{
			method: "PresignPutObject",
			url:    "http://localhost:900/pre-sign-put-object",
			fn: func(bucketName, objName string) (*string, error) {
				return mm.PresignPutObject(bgCtx, bucketName, objName, nil)
			},
		},
	}

	for _, data := range expected {
		genURL, err := data.fn(bucketName, objName)
		require.NoError(t, err, data.method)
		assert.NotNil(t, genURL)
		assert.Equal(t, &data.url, genURL)
	}
}

func TestMediaMetaUpload(t *testing.T) {
	t.Parallel()

	buff := bytes.NewBuffer([]byte(`Hello World!! test2`))

	mockProv := &MockedProviderMinio{}
	mockProv.On("Type").Return(provType.ToString())
	mockProv.On("PutObject", bgCtx, bucketName, objName, buff, int64(buff.Len()), provider.PutOptions{}).Return(nil)
	mockProv.On("BucketExists", bgCtx, bucketName).Return(true, nil)

	mm, err := mediameta.New(bgCtx, provType, mockProv)
	require.NoError(t, err)

	err = mm.Upload(bgCtx, bucketName, objName, buff, int64(buff.Len()), provider.PutOptions{})
	require.NoError(t, err)
}

func TestMediaMetaDelete(t *testing.T) {
	t.Parallel()

	mockProv := &MockedProviderMinio{}
	mockProv.On("Type").Return(provType.ToString())
	mockProv.On("RemoveObject", bgCtx, bucketName, objName).Return(nil)
	mockProv.On("BucketExists", bgCtx, bucketName).Return(true, nil)
	mockProv.On("ObjectExists", bgCtx, bucketName, objName).Return(true, nil)

	mm, err := mediameta.New(bgCtx, provType, mockProv)
	require.NoError(t, err)

	err = mm.Delete(bgCtx, bucketName, objName)
	require.NoError(t, err)
}

//nolint:dupl
func TestMediaMetaUpdateUserMetadata(t *testing.T) {
	t.Parallel()

	um := provider.UserMetadata{"test-key": "test-value"}

	expInfo := &provider.UserMetadataInfo{
		Bucket: bucketName,
		Key:    objName,
	}

	mockProv := &MockedProviderMinio{}
	mockProv.On("Type").Return(provType.ToString())
	mockProv.On("UpdateUserMetadata", bgCtx, bucketName, objName, um).Return(expInfo, nil)
	mockProv.On("BucketExists", bgCtx, bucketName).Return(true, nil)
	mockProv.On("ObjectExists", bgCtx, bucketName, objName).Return(true, nil)

	mm, err := mediameta.New(bgCtx, provType, mockProv)
	require.NoError(t, err)

	uInfo, err := mm.UpdateUserMetadata(bgCtx, bucketName, objName, um)
	require.NoError(t, err)
	assert.Equal(t, uInfo.Bucket, bucketName)
	assert.Equal(t, uInfo.Key, objName)
}

//nolint:dupl
func TestMediaMetaUpdateUserTags(t *testing.T) {
	t.Parallel()

	ut := provider.UserTags{"test-tag-key": "test-tag-value"}

	expInfo := &provider.UserMetadataInfo{
		Bucket: bucketName,
		Key:    objName,
	}

	mockProv := &MockedProviderMinio{}
	mockProv.On("Type").Return(provType.ToString())
	mockProv.On("UpdateUserTags", bgCtx, bucketName, objName, ut).Return(expInfo, nil)
	mockProv.On("BucketExists", bgCtx, bucketName).Return(true, nil)
	mockProv.On("ObjectExists", bgCtx, bucketName, objName).Return(true, nil)

	mm, err := mediameta.New(bgCtx, provType, mockProv)
	require.NoError(t, err)

	uInfo, err := mm.UpdateUserTags(bgCtx, bucketName, objName, ut)
	require.NoError(t, err)
	assert.Equal(t, uInfo.Bucket, bucketName)
	assert.Equal(t, uInfo.Key, objName)
}
