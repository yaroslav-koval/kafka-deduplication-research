package provider_test

import (
	"kafka-polygon/pkg/mediameta/provider"
	"kafka-polygon/pkg/mediameta/provider/awssdk"
	"kafka-polygon/pkg/mediameta/provider/minio"
	"testing"

	"github.com/tj/assert"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	pAws := awssdk.New(nil)
	pMinio := minio.New(nil)

	assert.Equal(t, provider.AwsS3, provider.Type("aws"))
	assert.Equal(t, pAws, provider.AwsS3.GetProvider(pAws, pMinio))
	assert.NoError(t, provider.AwsS3.Validate())

	assert.Equal(t, provider.Minio, provider.Type("minio"))
	assert.Equal(t, pMinio, provider.Minio.GetProvider(pAws, pMinio))
	assert.NoError(t, provider.Minio.Validate())
}

func TestProviderUserTags(t *testing.T) {
	t.Parallel()

	expQuery := "key1-tag=value1-tag&key2-tag=value2-tag"

	ut := provider.UserTags{
		"key1-tag": "value1-tag",
		"key2-tag": "value2-tag",
	}

	eq := ut.EncodeQuery()
	assert.NotNil(t, eq)
	assert.Equal(t, expQuery, *eq)

	nUt := make(provider.UserTags)
	nUt, err := nUt.ParseQuery(expQuery)
	assert.NoError(t, err)
	assert.Equal(t, ut, nUt)
}
