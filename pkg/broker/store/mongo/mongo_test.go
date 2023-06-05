package mongo_test

import (
	"context"
	"kafka-polygon/pkg/broker/store"
	storeMongo "kafka-polygon/pkg/broker/store/mongo"
	"kafka-polygon/pkg/cerror"
	"testing"

	"github.com/tj/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var _bgCtx = context.Background()

func TestGetEventInfoByID(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			1,
			"test-db.test-collection",
			mtest.FirstBatch,
			bson.D{{Key: "status", Value: "test-status"}}))

		s := storeMongo.NewStore(mt.Client, storeMongo.Settings{
			DatabaseName:   "test-db",
			CollectionName: "test-collection",
		})

		data, err := s.GetEventInfoByID(_bgCtx, "test-id")
		assert.NoError(t, err)
		assert.Equal(t, "test-status", data.Status)
	})

	mt.Run("error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})
		s := storeMongo.NewStore(mt.Client, storeMongo.Settings{
			DatabaseName:   "test-db",
			CollectionName: "test-collection",
		})

		_, err := s.GetEventInfoByID(_bgCtx, "test-id")
		assert.Error(t, err)
		assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
	})
}

func TestPutEventInfo(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{{Key: "status", Value: "test-status"}}},
		})

		s := storeMongo.NewStore(mt.Client, storeMongo.Settings{
			DatabaseName:   "test-db",
			CollectionName: "test-collection",
		})

		data := store.EventProcessData{Status: "test-status"}

		err := s.PutEventInfo(_bgCtx, "test-id", data)
		assert.NoError(t, err)
	})

	mt.Run("error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})
		s := storeMongo.NewStore(mt.Client, storeMongo.Settings{
			DatabaseName:   "test-db",
			CollectionName: "test-collection",
		})

		data := store.EventProcessData{Status: "test-status"}

		err := s.PutEventInfo(_bgCtx, "test-id", data)
		assert.Error(t, err)
		assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
	})
}
