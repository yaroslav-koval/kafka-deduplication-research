package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"kafka-polygon/pkg/converto"
	"net/url"
	"time"
)

const (
	AwsS3 Type = "aws"
	Minio Type = "minio"

	DefaultResponseExpiredMin = 7 * 24 * 60 // day * hours * minutes
)

type Provider interface {
	Connected(ctx context.Context) error
	Ping(ctx context.Context) error
	Type() Type
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	ObjectExists(ctx context.Context, bucketName, objName string) (bool, error)
	PutObject(ctx context.Context, bucketName, objName string, r io.Reader, rSize int64, opts PutOptions) error
	UpdateUserMetadata(ctx context.Context, bucketName, objName string, um UserMetadata) (*UserMetadataInfo, error)
	UpdateUserTags(ctx context.Context, bucketName, objName string, ut UserTags) (*UserMetadataInfo, error)
	RemoveObject(ctx context.Context, bucketName, objName string) error
	PresignGetObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error)
	PresignHeadObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error)
	PresignPutObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error)
	GetObject(ctx context.Context, bucketName, objName string) (*GetObject, error)
	GetObjectTagging(ctx context.Context, bucketName, objName string) (UserTags, error)
}

type UserMetadataInfo struct {
	Bucket       string
	Key          string
	ETag         string
	LastModified time.Time
}

type Type string

func (t Type) ToString() string {
	return string(t)
}

func (t Type) Validate() error {
	if t == "" {
		return errors.New("not empty type")
	}

	supports := []Type{AwsS3, Minio}
	for _, prov := range supports {
		if prov == t {
			return nil
		}
	}

	return fmt.Errorf("not support provider %s", t)
}

func (t Type) GetProvider(list ...Provider) Provider {
	for _, currProv := range list {
		if currProv.Type() == t {
			return currProv
		}
	}

	return nil
}

type UserMetadata map[string]string
type UserTags map[string]string

func (ut UserTags) ParseQuery(q string) (UserTags, error) {
	urlValues, err := url.ParseQuery(q)
	if err != nil {
		return ut, err
	}

	for key, vals := range urlValues {
		if len(vals) > 0 {
			ut[key] = vals[0]
		} else {
			ut[key] = ""
		}
	}

	return ut, nil
}

func (ut UserTags) EncodeQuery() *string {
	if len(ut) == 0 {
		return nil
	}

	values := url.Values{}
	for key, val := range ut {
		values.Add(key, val)
	}

	return converto.StringPointer(values.Encode())
}

type PutOptions struct {
	UserMetadata    UserMetadata
	UserTags        UserTags
	ContentType     string
	ContentEncoding string
	ContentLanguage string
}

type GetObject struct {
	UserMetadata UserMetadata
	ETag         string
	ContentType  string
	Body         io.ReadCloser
}
