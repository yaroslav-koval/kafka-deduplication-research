package cerror_test

import (
	"context"
	"errors"
	"kafka-polygon/pkg/cerror"
	"testing"

	"github.com/tj/assert"
)

func TestIsNotExist(t *testing.T) {
	ctx := context.Background()
	expS := []struct {
		exist bool
		err   error
	}{
		{
			err:   cerror.NewF(ctx, cerror.KindDBNoRows, "db no rows"),
			exist: true,
		},
		{
			err:   cerror.NewF(ctx, cerror.KindNotExist, "no exists"),
			exist: true,
		},
		{
			err:   cerror.NewF(ctx, cerror.KindMinioNotExist, "minio no exists"),
			exist: true,
		},
		{
			err:   errors.New("another err"),
			exist: false,
		},
		{
			err:   cerror.NewF(ctx, cerror.KindInternal, "internal"),
			exist: false,
		},
	}

	for _, item := range expS {
		assert.NotNil(t, item.err)
		assert.Equal(t, item.exist, cerror.IsNotExist(item.err))
	}
}
