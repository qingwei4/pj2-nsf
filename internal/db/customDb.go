package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	mongoUri        = "mongodb://127.0.0.1:27017"
	mongoCollection = "services"
	mongoTimeout    = 5 * time.Second
)

var (
	// A var, not a const, so tests can switch to a separate database.
	NSF_MONGO_DB_NAME = "nsf-mongodb"
)

type mongoDb struct {
	client *mongo.Client
	coll   *mongo.Collection
}

// mongoEntry is one stored key/value pair. The key (service name) is the
// document _id, so it is unique and indexed.
type mongoEntry struct {
	Key   string `bson:"_id"`
	Value string `bson:"value"`
}

func newMongoDb() (*mongoDb, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %v", err)
	}

	// mongo.Connect does not do any I/O, so ping to make sure the server is reachable.
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping mongo: %v", err)
	}

	return &mongoDb{
		client: client,
		coll:   client.Database(NSF_MONGO_DB_NAME).Collection(mongoCollection),
	}, nil
}

func (m *mongoDb) Release() error {
	if m.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *mongoDb) Save(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	_, err := m.coll.ReplaceOne(ctx, bson.M{"_id": key}, mongoEntry{Key: key, Value: value},
		options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save key %s: %v", key, err)
	}
	return nil
}

// SaveAll is not atomic: a standalone mongod has no transactions, so if the
// write fails part way, some keys may already be saved.
func (m *mongoDb) SaveAll(data map[string]string) error {
	if len(data) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(data))
	for key, value := range data {
		models = append(models, mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": key}).
			SetReplacement(mongoEntry{Key: key, Value: value}).
			SetUpsert(true))
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	if _, err := m.coll.BulkWrite(ctx, models); err != nil {
		return fmt.Errorf("failed to save data: %v", err)
	}
	return nil
}

func (m *mongoDb) Exist(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	n, err := m.coll.CountDocuments(ctx, bson.M{"_id": key}, options.Count().SetLimit(1))
	if err != nil {
		return false, fmt.Errorf("failed to check key %s: %v", key, err)
	}
	return n > 0, nil
}

func (m *mongoDb) Load(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	var entry mongoEntry
	err := m.coll.FindOne(ctx, bson.M{"_id": key}).Decode(&entry)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", fmt.Errorf("key %s not found", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to load key %s: %v", key, err)
	}
	return entry.Value, nil
}

func (m *mongoDb) LoadAll() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cursor, err := m.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to load all data: %v", err)
	}
	var entries []mongoEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode all data: %v", err)
	}

	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		result[entry.Key] = entry.Value
	}
	return result, nil
}

func (m *mongoDb) Remove(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	if _, err := m.coll.DeleteOne(ctx, bson.M{"_id": key}); err != nil {
		return fmt.Errorf("failed to remove key %s: %v", key, err)
	}
	return nil
}

func (m *mongoDb) RemoveAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	if _, err := m.coll.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("failed to remove all data: %v", err)
	}
	return nil
}
