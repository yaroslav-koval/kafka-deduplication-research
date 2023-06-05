//nolint:dupl
package awssdk_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/mediameta/provider"
	"kafka-polygon/pkg/mediameta/provider/awssdk"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

type args struct {
	bucket    string
	key       string
	eTag      string
	uMetadata provider.UserMetadata
	uTags     provider.UserTags
}

type itemData struct {
	name                     string
	args                     args
	resp                     resp
	respPresign              respPresign
	want                     []byte
	wantErr, expObjectExists bool
}

type respPresign struct {
	URL, Method string
	err         error
}

type resp struct {
	enabled bool
	body    string
	err     error
	tags    []types.Tag
}

func middlewareForPresign(r respPresign) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test-presign",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					return middleware.FinalizeOutput{
						Result: &v4.PresignedHTTPRequest{
							URL:    r.URL,
							Method: r.Method,
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForHeadBucket(r resp) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					mt := middleware.Metadata{}
					mt.Set("bucketExists", r.enabled)

					return middleware.FinalizeOutput{
						Result: &s3.HeadBucketOutput{
							ResultMetadata: mt,
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForHeadObject(r resp) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					return middleware.FinalizeOutput{
						Result: &s3.HeadObjectOutput{
							BucketKeyEnabled: r.enabled,
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForGetObject(r resp) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					return middleware.FinalizeOutput{
						Result: &s3.GetObjectOutput{
							Body: io.NopCloser(strings.NewReader(r.body)),
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForGetObjectTagging(r resp) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					return middleware.FinalizeOutput{
						Result: &s3.GetObjectTaggingOutput{
							TagSet: r.tags,
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForRemoveObject(r resp) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test-rm",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					mt := middleware.Metadata{}
					mt.Set("delete", "ok")
					return middleware.FinalizeOutput{
						Result: &s3.DeleteObjectOutput{
							DeleteMarker:   true,
							VersionId:      aws.String("0.0.1"),
							ResultMetadata: mt,
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForCopyObject(r resp) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					return middleware.FinalizeOutput{
						Result: &s3.CopyObjectOutput{
							BucketKeyEnabled: true,
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForPutObject(r resp) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					mt := middleware.Metadata{}
					mt.Set("put-obj", "ok")
					return middleware.FinalizeOutput{
						Result: &s3.PutObjectOutput{
							BucketKeyEnabled: true,
							ETag:             aws.String("put-obj-1"),
							ResultMetadata:   mt,
						},
					}, middleware.Metadata{}, r.err
				},
			),
			middleware.Before,
		)
	}
}

func middlewareForPing(err error) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				"test",
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					return middleware.FinalizeOutput{
						Result: &s3.HeadBucketOutput{},
					}, middleware.Metadata{}, err
				},
			),
			middleware.Before,
		)
	}
}

func TestS3GetObject(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForGetObject(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			obj, err := pS3.GetObject(ctx, tt.args.bucket, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			got, _ := io.ReadAll(obj.Body)
			if !bytes.Equal(tt.want, got) {
				t.Errorf("GetObject() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestS3GetObjectTagging(t *testing.T) {
	t.Parallel()

	expTags := provider.UserTags{"tag-key-1": "tag-value-1"}

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForGetObjectTagging(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			obj, err := pS3.GetObjectTagging(ctx, tt.args.bucket, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetObjectTagging() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.Equal(t, expTags, obj)
		})
	}
}

func TestS3RemoveObject(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForRemoveObject(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			err = pS3.RemoveObject(ctx, tt.args.bucket, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
		})
	}
}

func TestS3UpdateUserMetadata(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForCopyObject(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			uInfo, err := pS3.UpdateUserMetadata(ctx, tt.args.bucket, tt.args.key, tt.args.uMetadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.Equal(t, uInfo.Bucket, tt.args.bucket)
			assert.Equal(t, uInfo.Key, tt.args.key)
		})
	}
}

func TestS3UpdateUserTags(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForCopyObject(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			uInfo, err := pS3.UpdateUserTags(ctx, tt.args.bucket, tt.args.key, tt.args.uTags)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.Equal(t, uInfo.Bucket, tt.args.bucket)
			assert.Equal(t, uInfo.Key, tt.args.key)
		})
	}
}

func TestS3PutObject(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForPutObject(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			buff := bytes.NewBuffer([]byte("test body"))

			err = pS3.PutObject(ctx, tt.args.bucket, tt.args.key, buff, int64(buff.Len()), provider.PutOptions{
				UserMetadata: tt.args.uMetadata,
				ContentType:  "text/plain",
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("PutObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
		})
	}
}

func TestS3Type(t *testing.T) {
	t.Parallel()

	pS3 := awssdk.New(nil)

	assert.Equal(t, provider.AwsS3, pS3.Type())
}

func TestS3ObjectExists(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForHeadObject(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			ok, err := pS3.ObjectExists(ctx, tt.args.bucket, tt.args.key)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, cerror.KindNotExist, cerror.ErrKind(err))
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expObjectExists, ok)
		})
	}
}

func TestS3BucketExists(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForHeadBucket(tt.resp)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			ok, err := pS3.BucketExists(ctx, tt.args.bucket)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, cerror.KindNotExist, cerror.ErrKind(err))
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expObjectExists, ok)
		})
	}
}

func TestS3PreSign(t *testing.T) {
	t.Parallel()

	for _, tt := range getTestData() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForPresign(tt.respPresign)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			actual, err := pS3.PresignGetObject(ctx, tt.args.bucket, tt.args.key, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.Equal(t, converto.StringPointer(tt.respPresign.URL), actual)

			actual, err = pS3.PresignHeadObject(ctx, tt.args.bucket, tt.args.key, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.Equal(t, converto.StringPointer(tt.respPresign.URL), actual)

			actual, err = pS3.PresignPutObject(ctx, tt.args.bucket, tt.args.key, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.Equal(t, converto.StringPointer(tt.respPresign.URL), actual)
		})
	}
}

func TestS3Ping(t *testing.T) {
	t.Parallel()

	errs := map[string]error{
		"internal_server": fmt.Errorf("StatusCode: %d internal_server", http.StatusInternalServerError),
		"unauthorized":    fmt.Errorf("StatusCode: %d unauthorized", http.StatusUnauthorized),
		"forbidden":       fmt.Errorf("StatusCode: %d forbidden", http.StatusForbidden),
		"ok":              nil,
	}

	for name, ie := range errs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{middlewareForPing(ie)}),
			)
			if err != nil {
				t.Fatal(err)
			}
			client := s3.NewFromConfig(cfg)

			// used provider
			pS3 := awssdk.New(nil)
			pS3.SetClient(client)

			err = pS3.Ping(ctx)
			if ie == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, cerror.KindInternal, cerror.ErrKind(err))
			}
		})
	}
}

func getTestData() []itemData {
	um1 := make(provider.UserMetadata, 0)
	um1["u-meta-1"] = "val-1"

	ut1 := make(provider.UserTags, 0)
	ut1["u-tag-1"] = "tag-val-1"

	um2 := make(provider.UserMetadata, 0)
	um2["u-meta-2"] = "val-2"

	ut2 := make(provider.UserTags, 0)
	ut2["u-tag-2"] = "tag-val-2"

	return []itemData{
		{
			name: "success",
			args: args{bucket: "Bucket", key: "Key", eTag: "etag-1", uMetadata: um1, uTags: ut1},
			resp: resp{
				body: "ok",
				tags: []types.Tag{
					{
						Key:   converto.StringPointer("tag-key-1"),
						Value: converto.StringPointer("tag-value-1"),
					},
				},
			},
			respPresign:     respPresign{URL: "url-success", Method: http.MethodGet},
			want:            []byte("ok"),
			expObjectExists: true,
		},
		{
			name:            "failure",
			args:            args{bucket: "Bucket", key: "Key", eTag: "etag-2", uMetadata: um2, uTags: ut2},
			resp:            resp{err: errors.New("object not found")},
			respPresign:     respPresign{URL: "url-error", Method: http.MethodGet, err: errors.New("not presign err")},
			wantErr:         true,
			expObjectExists: false,
		},
	}
}
