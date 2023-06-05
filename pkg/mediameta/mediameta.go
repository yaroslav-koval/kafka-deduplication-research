package mediameta

import (
	"context"
	"errors"
	"io"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/mediameta/provider"
)

var (
	errNotEmptyProvider = errors.New("not empty provider")
)

type MediaStorage interface {
	PresignGetObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error)
	PresignHeadObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error)
	PresignPutObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error)
	Upload(ctx context.Context, bucketName, objName string, r io.Reader, rSize int64, opts provider.PutOptions) error
	Delete(ctx context.Context, bucketName, objName string) error
	UpdateUserMetadata(ctx context.Context, bucketName, objName string, um provider.UserMetadata) (*provider.UserMetadataInfo, error)
	UpdateUserTags(ctx context.Context, bucketName, objName string, ut provider.UserTags) (*provider.UserMetadataInfo, error)
	Connected(ctx context.Context) error
	Ping(ctx context.Context) error
}

type mediaStorage struct {
	p provider.Provider
}

func New(
	ctx context.Context,
	useProvider provider.Type,
	list ...provider.Provider) (MediaStorage, error) {
	if len(list) == 0 {
		return nil, cerror.New(ctx, cerror.KindBadParams, errNotEmptyProvider).LogError()
	}

	log.InfoF(ctx, "[mediameta] useProvider: %s", useProvider)

	// validate used provider
	if err := useProvider.Validate(); err != nil {
		return nil, cerror.New(ctx, cerror.KindBadParams, err).LogError()
	}

	return &mediaStorage{
		p: useProvider.GetProvider(list...),
	}, nil
}

// PresignPutObject returns presign URL and validate bucketName of mediastorage
func (ms *mediaStorage) PresignPutObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	if err := ms.validateBucket(ctx, bucketName); err != nil {
		return nil, err
	}

	return ms.p.PresignPutObject(ctx, bucketName, objName, expireMin)
}

// PresignHeadObject returns presign URL and validate input params of mediastorage
func (ms *mediaStorage) PresignHeadObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	if err := ms.validateAll(ctx, bucketName, objName); err != nil {
		return nil, err
	}

	return ms.p.PresignHeadObject(ctx, bucketName, objName, expireMin)
}

// PresignGetObject returns presign URL and validate input params of mediastorage
func (ms *mediaStorage) PresignGetObject(ctx context.Context, bucketName, objName string, expireMin *int64) (*string, error) {
	if err := ms.validateAll(ctx, bucketName, objName); err != nil {
		return nil, err
	}

	return ms.p.PresignGetObject(ctx, bucketName, objName, expireMin)
}

// Upload validate input params and upload current file in mediaStorage
func (ms *mediaStorage) Upload(ctx context.Context, bucketName, objName string, r io.Reader, rSize int64, opts provider.PutOptions) error {
	if err := ms.validateBucket(ctx, bucketName); err != nil {
		return err
	}

	return ms.p.PutObject(ctx, bucketName, objName, r, rSize, opts)
}

// Delete remove current object
func (ms *mediaStorage) Delete(ctx context.Context, bucketName, objName string) error {
	if err := ms.validateAll(ctx, bucketName, objName); err != nil {
		return err
	}

	return ms.p.RemoveObject(ctx, bucketName, objName)
}

// UpdateUserMetadata update metadata current object in bucket
func (ms *mediaStorage) UpdateUserMetadata(ctx context.Context,
	bucketName, objName string, um provider.UserMetadata) (*provider.UserMetadataInfo, error) {
	if err := ms.validateAll(ctx, bucketName, objName); err != nil {
		return nil, err
	}

	return ms.p.UpdateUserMetadata(ctx, bucketName, objName, um)
}

// UpdateUserTags update tags current object in bucket
func (ms *mediaStorage) UpdateUserTags(ctx context.Context,
	bucketName, objName string, ut provider.UserTags) (*provider.UserMetadataInfo, error) {
	if err := ms.validateAll(ctx, bucketName, objName); err != nil {
		return nil, err
	}

	return ms.p.UpdateUserTags(ctx, bucketName, objName, ut)
}

// Connected initializes connection of mediastorage
func (ms *mediaStorage) Connected(ctx context.Context) error {
	return ms.p.Connected(ctx)
}

// Ping check connected
func (ms *mediaStorage) Ping(ctx context.Context) error {
	return ms.p.Ping(ctx)
}

func (ms *mediaStorage) validateAll(ctx context.Context, bucketName, objName string) error {
	err := ms.validateBucket(ctx, bucketName)
	if err != nil {
		return err
	}

	err = ms.validateObject(ctx, bucketName, objName)
	if err != nil {
		return err
	}

	return nil
}

func (ms *mediaStorage) validateBucket(ctx context.Context, bucketName string) error {
	if exists, err := ms.p.BucketExists(ctx, bucketName); !exists || err != nil {
		if err != nil {
			return cerror.NewF(
				ctx,
				cerror.KindNotExist,
				"bucket [%s] %s", bucketName, err).LogError()
		}

		return cerror.NewF(
			ctx,
			cerror.KindNotExist,
			"not exists bucket [%s]", bucketName).LogError()
	}

	return nil
}

func (ms *mediaStorage) validateObject(ctx context.Context, bucketName, objName string) error {
	if exists, err := ms.p.ObjectExists(ctx, bucketName, objName); !exists || err != nil {
		if err != nil {
			return cerror.NewF(
				ctx,
				cerror.KindNotExist,
				"bucket %s (ObjName: %s) %s", bucketName, objName, err).LogError()
		}

		return cerror.NewF(
			ctx,
			cerror.KindNotExist,
			"not exists object in bucket %s (ObjName: %s)", bucketName, objName).LogError()
	}

	return nil
}
