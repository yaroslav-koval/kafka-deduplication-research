package awssdk

import (
	"context"
	"fmt"
	"io"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/mediameta/provider"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Options struct {
	AccessKeyID        string
	SecretAccessKey    string
	Token              string
	Region             string
	ResponseExpiredMin int64
}

type S3 struct {
	c       *s3.Client
	opts    *Options
	preSign *s3.PresignClient
}

func New(opts *Options) *S3 {
	return &S3{
		opts: opts,
	}
}

// SetClient set aws.s3 client
func (s *S3) SetClient(c *s3.Client) {
	s.c = c
	s.preSign = s3.NewPresignClient(c)
}

// RemoveObject remove current object
func (s *S3) RemoveObject(ctx context.Context, bucketName, objName string) error {
	_, err := s.c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &objName,
	})

	if err != nil {
		return cerror.NewF(ctx, cerror.KindOther, "%s::RemoveObject %s", s.Type(), err).LogError()
	}

	return nil
}

// UpdateUserMetadata change user metadata
func (s *S3) UpdateUserMetadata(ctx context.Context,
	bucketName, objName string, um provider.UserMetadata) (*provider.UserMetadataInfo, error) {
	copySource := url.PathEscape(fmt.Sprintf("%s/%s", bucketName, objName))
	objInfo, err := s.c.CopyObject(context.TODO(), &s3.CopyObjectInput{
		CopySource:        &copySource,
		Bucket:            &bucketName,
		Key:               &objName,
		Metadata:          um,
		MetadataDirective: "REPLACE", // COPY | REPLACE
	})

	if err != nil {
		return nil, cerror.NewF(ctx, cerror.KindOther, "%s::UpdateUserMetadata %s", s.Type(), err).LogError()
	}

	var (
		eTag         string
		lastModified time.Time
	)

	if objInfo != nil && objInfo.CopyObjectResult != nil {
		eTag = converto.StringValue(objInfo.CopyObjectResult.ETag)
		lastModified = converto.TimeValue(objInfo.CopyObjectResult.LastModified)
	}

	return &provider.UserMetadataInfo{
		Bucket:       bucketName,
		Key:          objName,
		ETag:         eTag,
		LastModified: lastModified,
	}, nil
}

// UpdateUserTags change user tags
func (s *S3) UpdateUserTags(ctx context.Context, bucketName, objName string, ut provider.UserTags) (*provider.UserMetadataInfo, error) {
	copySource := url.PathEscape(fmt.Sprintf("%s/%s", bucketName, objName))
	objInfo, err := s.c.CopyObject(context.TODO(), &s3.CopyObjectInput{
		CopySource:       &copySource,
		Bucket:           &bucketName,
		Key:              &objName,
		TaggingDirective: "REPLACE",
		Tagging:          ut.EncodeQuery(),
	})

	if err != nil {
		return nil, cerror.NewF(ctx, cerror.KindOther, "%s::UpdateUserTags %s", s.Type(), err).LogError()
	}

	var (
		eTag         string
		lastModified time.Time
	)

	if objInfo != nil && objInfo.CopyObjectResult != nil {
		eTag = converto.StringValue(objInfo.CopyObjectResult.ETag)
		lastModified = converto.TimeValue(objInfo.CopyObjectResult.LastModified)
	}

	return &provider.UserMetadataInfo{
		Bucket:       bucketName,
		Key:          objName,
		ETag:         eTag,
		LastModified: lastModified,
	}, nil
}

// GetObject download current object
func (s *S3) GetObject(ctx context.Context, bucketName, objName string) (*provider.GetObject, error) {
	objOutput, err := s.c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objName,
	})

	if err != nil {
		return nil, cerror.NewF(ctx, cerror.KindOther, "%s::GetObject %s", s.Type(), err).LogError()
	}

	return &provider.GetObject{
		UserMetadata: objOutput.Metadata,
		ETag:         aws.ToString(objOutput.ETag),
		ContentType:  aws.ToString(objOutput.ContentType),
		Body:         objOutput.Body,
	}, nil
}

// GetObjectTagging get tags current object
func (s *S3) GetObjectTagging(ctx context.Context, bucketName, objName string) (provider.UserTags, error) {
	ut := make(provider.UserTags)
	objOutput, err := s.c.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: &bucketName,
		Key:    &objName,
	})

	if err != nil {
		return ut, cerror.NewF(ctx, cerror.KindOther, "%s::GetObjectTagging %s", s.Type(), err).LogError()
	}

	if len(objOutput.TagSet) == 0 {
		return ut, cerror.NewF(ctx, cerror.KindOther,
			"%s::GetObjectTagging not found any tags %s",
			s.Type(), err).LogError()
	}

	for _, t := range objOutput.TagSet {
		ut[converto.StringValue(t.Key)] = converto.StringValue(t.Value)
	}

	return ut, err
}

// PutObject upload new object
func (s *S3) PutObject(ctx context.Context, bucketName, objName string, r io.Reader, rSize int64, opts provider.PutOptions) error {
	poi := &s3.PutObjectInput{
		Bucket:          &bucketName,
		Key:             &objName,
		Body:            r,
		ContentLength:   rSize,
		Metadata:        opts.UserMetadata,
		ContentType:     &opts.ContentType,
		ContentLanguage: &opts.ContentLanguage,
		ContentEncoding: &opts.ContentEncoding,
		Tagging:         opts.UserTags.EncodeQuery(),
	}

	_, err := s.c.PutObject(ctx, poi)
	if err != nil {
		return cerror.NewF(ctx, cerror.KindOther, "%s::PutObject %s", s.Type(), err).LogError()
	}

	return nil
}

// ObjectExists check current object in bucket
func (s *S3) ObjectExists(ctx context.Context, bucketName, objName string) (bool, error) {
	hoi := &s3.HeadObjectInput{
		Bucket: &bucketName,
		Key:    &objName,
	}
	_, err := s.c.HeadObject(ctx, hoi)

	if err != nil {
		return false, cerror.NewF(ctx, cerror.KindNotExist, "%s::ObjectExists %s", s.Type(), err).LogWarn()
	}

	return true, nil
}

// BucketExists check current bucket
func (s *S3) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	hbi := &s3.HeadBucketInput{
		Bucket: &bucketName,
	}
	_, err := s.c.HeadBucket(ctx, hbi)

	if err != nil {
		return false, cerror.NewF(ctx, cerror.KindNotExist, "%s::BucketExists %s", s.Type(), err).LogError()
	}

	return true, nil
}

// PresignGetObject generate expires URL
func (s *S3) PresignGetObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	urlObj, err := s.preSign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objName,
	}, s3.WithPresignExpires(s.getExpiredTime(expireMin)))

	if err != nil {
		return nil, cerror.NewF(ctx, cerror.KindOther, "%s::PresignGetObject %s", s.Type(), err).LogError()
	}

	return aws.String(urlObj.URL), nil
}

// PresignHeadObject generate expires URL
func (s *S3) PresignHeadObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	urlObj, err := s.preSign.PresignHeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucketName,
		Key:    &objName,
	}, s3.WithPresignExpires(s.getExpiredTime(expireMin)))

	if err != nil {
		return nil, cerror.NewF(ctx, cerror.KindOther, "%s::PresignHeadObject %s", s.Type(), err).LogError()
	}

	return aws.String(urlObj.URL), nil
}

// PresignPutObject generate expires URL
func (s *S3) PresignPutObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	urlObj, err := s.preSign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &objName,
	}, s3.WithPresignExpires(s.getExpiredTime(expireMin)))

	if err != nil {
		return nil, cerror.NewF(ctx, cerror.KindOther, "%s::PresignPutObject %s", s.Type(), err).LogError()
	}

	return aws.String(urlObj.URL), nil
}

// Connected of aws.s3
func (s *S3) Connected(ctx context.Context) error {
	if s.c != nil {
		return nil
	}

	var (
		secretAccessKey,
		accessKeyID,
		token,
		region string
		optsS3 config.LoadOptionsFunc
	)

	if s.opts != nil {
		region = s.opts.Region
		token = s.opts.Token
		accessKeyID = s.opts.AccessKeyID
		secretAccessKey = s.opts.SecretAccessKey
	}

	if accessKeyID != "" && secretAccessKey != "" {
		optsS3 = config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, token))
	}

	// Load the SDK's configuration from environment and shared config, and
	// create the client with this.
	cfg, err := config.LoadDefaultConfig(ctx, optsS3)
	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "%s::Connected %s", s.Type(), err).LogError()
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = region
	})
	s.c = s3Client
	s.preSign = s3.NewPresignClient(s3Client)

	return nil
}

// Type of aws.s3
func (s *S3) Type() provider.Type {
	return provider.AwsS3
}

func (s *S3) getExpiredTime(expireMin *int64) time.Duration {
	if expireMin != nil {
		return time.Duration(converto.Int64Value(expireMin)) * time.Minute
	}

	if s.opts != nil && s.opts.ResponseExpiredMin > 0 {
		return time.Duration(s.opts.ResponseExpiredMin) * time.Minute
	}

	return time.Duration(provider.DefaultResponseExpiredMin) * time.Minute
}

func (s *S3) Ping(ctx context.Context) error {
	_, err := s.BucketExists(ctx, "probe-s3-health")

	if err != nil {
		checkStatusCode := []int{
			http.StatusForbidden,
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		}

		for _, code := range checkStatusCode {
			if strings.Contains(err.Error(), fmt.Sprintf("StatusCode: %d", code)) {
				return cerror.NewF(ctx, cerror.KindInternal, "s3 ping to message %s", err).LogError()
			}
		}
	}

	return nil
}
