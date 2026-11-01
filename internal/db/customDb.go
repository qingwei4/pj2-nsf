package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	mongoUri        = "mongodb://127.0.0.1:27017"
	mongoDbName     = "nsf-mongodb"
	mongoCollection = "services"
	mongoTimeout    = 5 * time.Second
)

var errMongoNotImplemented = errors.New("mongoDb: not implemented")

type mongoDb struct {
	client *mongo.Client
	coll   *mongo.Collection
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
		coll:   client.Database(mongoDbName).Collection(mongoCollection),
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

// TODO: implement the methods below.

func (m *mongoDb) Save(key, value string) error {
	return errMongoNotImplemented
}

func (m *mongoDb) SaveAll(data map[string]string) error {
	return errMongoNotImplemented
}

func (m *mongoDb) Exist(key string) (bool, error) {
	return false, errMongoNotImplemented
}

func (m *mongoDb) Load(key string) (string, error) {
	return "", errMongoNotImplemented
}

func (m *mongoDb) LoadAll() (map[string]string, error) {
	return nil, errMongoNotImplemented
}

func (m *mongoDb) Remove(key string) error {
	return errMongoNotImplemented
}

func (m *mongoDb) RemoveAll() error {
	return errMongoNotImplemented
}
