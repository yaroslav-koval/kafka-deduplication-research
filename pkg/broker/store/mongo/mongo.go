package mongo

import (
	"context"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cerror"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const defCollBroker = "broker"

// Settings Endpoint are parameters for the MongoDB event store
// to use when initializing.
type Settings struct {
	DatabaseName   string // DatabaseName is the database to create/connect to.
	CollectionName string // CollectionName is the collection name to put new documents in to
}

type Store struct {
	settings Settings
	cl       *mongo.Client
}

var _ store.Store = (*Store)(nil)

func NewStore(client *mongo.Client, settings Settings) *Store {
	if settings.CollectionName == "" {
		settings.CollectionName = defCollBroker
	}

	return &Store{
		cl:       client,
		settings: settings,
	}
}

func (s *Store) GetEventInfoByID(ctx context.Context, id string) (store.EventProcessData, error) {
	var dst store.EventProcessData

	err := s.getCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&dst)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return store.EventProcessData{}, cerror.New(ctx, cerror.KindNotExist, err) //nolint:cerrl
		}

		return store.EventProcessData{}, cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	return dst, nil
}

func (s *Store) PutEventInfo(ctx context.Context, id string, data store.EventProcessData) error {
	var dst *store.EventProcessData

	err := s.getCollection().
		FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": data.Status}}).
		Decode(&dst)
	if err != nil && err != mongo.ErrNoDocuments {
		return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	if dst == nil {
		_, err = s.getCollection().InsertOne(ctx, bson.M{"_id": id, "status": data.Status})
		if err != nil {
			return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
		}
	}

	return nil
}

func (s *Store) getCollection() *mongo.Collection {
	return s.cl.Database(s.settings.DatabaseName).Collection(s.settings.CollectionName)
}
