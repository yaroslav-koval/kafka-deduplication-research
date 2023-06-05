package minio

import (
	"context"
	"fmt"
	"io"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/mediameta/provider"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Options struct {
	Endpoint           string
	AccessKeyID        string
	SecretAccessKey    string
	Token              string
	Region             string
	UseSSL             bool
	ResponseExpiredMin int64
}

type Minio struct {
	c    *minio.Client
	opts *Options
}

func New(opts *Options) *Minio {
	return &Minio{
		opts: opts,
	}
}

// SetClient set minio client
func (m *Minio) SetClient(c *minio.Client) {
	m.c = c
}

// RemoveObject remove current object
func (m *Minio) RemoveObject(ctx context.Context, bucketName, objName string) error {
	err := m.c.RemoveObject(ctx, bucketName, objName, minio.RemoveObjectOptions{ForceDelete: true})

	if err != nil {
		return cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::RemoveObject %s", m.Type(), err).LogError()
	}

	return nil
}

// UpdateUserMetadata change user metadata
func (m *Minio) UpdateUserMetadata(ctx context.Context,
	bucketName, objName string, um provider.UserMetadata) (*provider.UserMetadataInfo, error) {
	// Source object
	srcOpts := minio.CopySrcOptions{
		Bucket: bucketName,
		Object: objName,
	}

	// Destination object
	dstOpts := minio.CopyDestOptions{
		Bucket:          bucketName,
		Object:          objName,
		UserMetadata:    um,
		ReplaceMetadata: true,
	}

	objInfo, err := m.c.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::UpdateUserMetadata %s", m.Type(), err).LogError()
	}

	return &provider.UserMetadataInfo{
		Bucket:       objInfo.Bucket,
		Key:          objInfo.Key,
		ETag:         objInfo.ETag,
		LastModified: objInfo.LastModified,
	}, nil
}

// UpdateUserTags change user tags
func (m *Minio) UpdateUserTags(ctx context.Context,
	bucketName, objName string, ut provider.UserTags) (*provider.UserMetadataInfo, error) {
	// Source object
	srcOpts := minio.CopySrcOptions{
		Bucket: bucketName,
		Object: objName,
	}

	// Destination object
	dstOpts := minio.CopyDestOptions{
		Bucket:          bucketName,
		Object:          objName,
		UserTags:        ut,
		ReplaceMetadata: false,
		ReplaceTags:     true,
	}

	objInfo, err := m.c.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::UpdateUserTags %s", m.Type(), err).LogError()
	}

	return &provider.UserMetadataInfo{
		Bucket:       objInfo.Bucket,
		Key:          objInfo.Key,
		ETag:         objInfo.ETag,
		LastModified: objInfo.LastModified,
	}, nil
}

// GetObject download current object
func (m *Minio) GetObject(ctx context.Context, bucketName, objName string) (*provider.GetObject, error) {
	reader, err := m.c.GetObject(ctx, bucketName, objName, minio.GetObjectOptions{})

	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::GetObject %s", m.Type(), err).LogError()
	}

	defer func() {
		_ = reader.Close()
	}()

	objInfo, err := reader.Stat()
	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::GetObject get stat %s", m.Type(), err).LogError()
	}

	return &provider.GetObject{
		UserMetadata: provider.UserMetadata(objInfo.UserMetadata),
		ETag:         objInfo.ETag,
		ContentType:  objInfo.ContentType,
		Body:         reader,
	}, nil
}

// GetObjectTagging get tags current object
func (m *Minio) GetObjectTagging(ctx context.Context, bucketName, objName string) (provider.UserTags, error) {
	ut := make(provider.UserTags)

	tags, err := m.c.GetObjectTagging(ctx, bucketName, objName, minio.GetObjectTaggingOptions{})

	if err != nil {
		return ut, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::GetObjectTagging %s", m.Type(), err).LogError()
	}

	return ut.ParseQuery(tags.String())
}

// PutObject upload new object
func (m *Minio) PutObject(ctx context.Context, bucketName, objName string, r io.Reader, rSize int64, opts provider.PutOptions) error {
	mOpts := minio.PutObjectOptions{
		UserMetadata:    opts.UserMetadata,
		UserTags:        opts.UserTags,
		ContentType:     opts.ContentType,
		ContentEncoding: opts.ContentEncoding,
		ContentLanguage: opts.ContentLanguage,
	}
	_, err := m.c.PutObject(ctx, bucketName, objName, r, rSize, mOpts)

	if err != nil {
		return cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::PutObject %s", m.Type(), err).LogError()
	}

	return nil
}

// ObjectExists check current object in bucket
func (m *Minio) ObjectExists(ctx context.Context, bucketName, objName string) (bool, error) {
	_, err := m.c.StatObject(ctx, bucketName, objName, minio.StatObjectOptions{})
	if err != nil {
		return false, cerror.NewF(ctx,
			cerror.KindMinioNotExist,
			"%s::ObjectExists %s", m.Type(), err).LogError()
	}

	return true, nil
}

// BucketExists check current bucket
func (m *Minio) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	ok, err := m.c.BucketExists(ctx, bucketName)
	if err != nil {
		return ok, cerror.New(ctx, cerror.MinioToKind(err), err).LogError()
	}

	return ok, nil
}

// PresignGetObject generate expires URL
func (m *Minio) PresignGetObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	urlObj, err := m.c.PresignedGetObject(ctx,
		bucketName,
		objName,
		m.getExpiredTime(expireMin),
		m.getReqParams(objName))

	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::PresignedGetObject %s", m.Type(), err).LogWarn()
	}

	return converto.StringPointer(urlObj.String()), nil
}

// PresignHeadObject generate expires URL
func (m *Minio) PresignHeadObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	urlObj, err := m.c.PresignedHeadObject(ctx,
		bucketName,
		objName,
		m.getExpiredTime(expireMin),
		m.getReqParams(objName))

	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::PresignedHeadObject %s", m.Type(), err).LogWarn()
	}

	return converto.StringPointer(urlObj.String()), nil
}

// PresignPutObject generate expires URL
func (m *Minio) PresignPutObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	urlObj, err := m.c.PresignedPutObject(
		ctx,
		bucketName,
		objName,
		m.getExpiredTime(expireMin))

	if err != nil {
		return nil, cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::PresignedPutObject %s", m.Type(), err).LogError()
	}

	return converto.StringPointer(urlObj.String()), nil
}

// Connected of minio
func (m *Minio) Connected(ctx context.Context) error {
	if m.c != nil {
		return nil
	}

	var (
		endpoint,
		accessKeyID,
		secretAccessKey,
		token,
		region string
		useSSL bool
	)

	if m.opts != nil {
		endpoint = m.opts.Endpoint
		accessKeyID = m.opts.AccessKeyID
		secretAccessKey = m.opts.SecretAccessKey
		token = m.opts.Token
		useSSL = m.opts.UseSSL
		region = m.opts.Region
	}

	clOpts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, token),
		Secure: useSSL,
	}

	if region != "" {
		clOpts.Region = region
	}

	// Initialize minio client object.
	client, err := minio.New(endpoint, clOpts)

	if err != nil {
		return cerror.NewF(ctx,
			cerror.MinioToKind(err),
			"%s::Connected %s", m.Type(), err).LogError()
	}

	m.c = client

	return nil
}

// Ping of minio
func (m *Minio) Ping(ctx context.Context) error {
	_, err := m.c.BucketExists(ctx, "probe-minio-health")

	if !minio.IsNetworkOrHostDown(err, false) {
		switch minio.ToErrorResponse(err).Code {
		case "NoSuchBucket", "":
			return nil
		}
	}

	return cerror.NewF(ctx,
		cerror.MinioToKind(err),
		"ping minio to message %s", err).LogError()
}

// Type of minio
func (m *Minio) Type() provider.Type {
	return provider.Minio
}

func (m *Minio) getExpiredTime(expireMin *int64) time.Duration {
	if expireMin != nil {
		return time.Duration(converto.Int64Value(expireMin)) * time.Minute
	}

	if m.opts != nil && m.opts.ResponseExpiredMin > 0 {
		return time.Duration(m.opts.ResponseExpiredMin) * time.Minute
	}

	return time.Duration(provider.DefaultResponseExpiredMin) * time.Minute
}

func (m *Minio) getReqParams(objName string) url.Values {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=%q", objName))

	return reqParams
}
