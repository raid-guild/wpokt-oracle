package db

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	lock "github.com/square/mongo-lock"
)

var (
	mongoDB Database
)

type Database interface {
	Connect() error
	Disconnect() error

	InsertOne(collection string, data interface{}) (primitive.ObjectID, error)
	FindOne(collection string, filter interface{}, result interface{}) error
	FindMany(collection string, filter interface{}, result interface{}) error
	FindManySorted(collection string, filter interface{}, sort interface{}, result interface{}) error
	AggregateOne(collection string, pipeline interface{}, result interface{}) error
	AggregateMany(collection string, pipeline interface{}, result interface{}) error

	UpdateOne(collection string, filter interface{}, update interface{}) (primitive.ObjectID, error)
	UpsertOne(collection string, filter interface{}, update interface{}) (primitive.ObjectID, error)

	XLock(resourceID string) (string, error)
	SLock(resourceID string) (string, error)
	Unlock(lockID string) error
}

// MongoDatabase is a wrapper around the mongo database
type MongoDatabase struct {
	db       *mongo.Database
	uri      string
	database string
	locker   *lock.Client
	timeout  time.Duration

	logger *log.Entry
}

// Connect connects to the database
func (d *MongoDatabase) Connect() error {
	d.logger.Debug("Connecting to database")
	wcMajority := writeconcern.Majority()

	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(d.uri).SetWriteConcern(wcMajority))
	if err != nil {
		return err
	}
	d.db = client.Database(d.database)

	d.logger.Info("Connected to mongo database: ", d.database)
	return nil
}

// SetupLocker sets up the locker
func (d *MongoDatabase) SetupLocker() error {
	d.logger.Debug("Setting up locker")

	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	locker := lock.NewClient(d.db.Collection(common.CollectionLocks))
	err := locker.CreateIndexes(ctx)
	if err != nil {
		return err
	}

	d.locker = locker

	d.logger.Info("Locker setup")
	return nil
}

func randomString(n int) (string, error) {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes), nil
}

// XLock locks a resource for exclusive access
func (d *MongoDatabase) XLock(resourceID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	lockID, err := randomString(32)
	if err != nil {
		return "", err
	}
	err = d.locker.XLock(ctx, resourceID, lockID, lock.LockDetails{
		TTL: 60, // locks expire in 60 seconds
	})
	return lockID, err
}

// SLock locks a resource for shared access
func (d *MongoDatabase) SLock(resourceID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	lockID, err := randomString(32)
	if err != nil {
		return "", err
	}
	err = d.locker.SLock(ctx, resourceID, lockID, lock.LockDetails{
		TTL: 60, // locks expire in 60 seconds
	}, -1)
	return lockID, err
}

// Unlock unlocks a resource
func (d *MongoDatabase) Unlock(lockID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	_, err := d.locker.Unlock(ctx, lockID)
	return err
}

// Setup Indexes
func (d *MongoDatabase) SetupIndexes() error {
	d.logger.Debug("Setting up indexes")

	d.logger.Debug("Setting up indexes for transactions")
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err := d.db.Collection(common.CollectionTransactions).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "hash", Value: 1}, {Key: "chain.chain_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	d.logger.Debug("Setting up indexes for refunds")
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(common.CollectionRefunds).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "origin_transaction_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(common.CollectionRefunds).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "sequence", Value: 1}},
		Options: options.Index().SetUnique(true).
			SetPartialFilterExpression(bson.D{{Key: "sequence", Value: bson.D{{Key: "$exists", Value: true}, {Key: "$type", Value: "long"}}}}),
	})
	if err != nil {
		return err
	}

	d.logger.Debug("Setting up indexes for messages")
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(common.CollectionMessages).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "message_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(common.CollectionMessages).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "sequence", Value: 1}},
		Options: options.Index().SetUnique(true).
			SetPartialFilterExpression(bson.D{{Key: "sequence", Value: bson.D{{Key: "$exists", Value: true}, {Key: "$type", Value: "long"}}}}),
	})
	if err != nil {
		return err
	}

	d.logger.Debug("Setting up indexes for nodes")
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(common.CollectionNodes).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "hostname", Value: 1}, {Key: "oracle_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	d.logger.Info("Indexes setup")

	return nil
}

// Disconnect disconnects from the database
func (d *MongoDatabase) Disconnect() error {
	d.logger.Debug("Disconnecting from database")
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	err := d.db.Client().Disconnect(ctx)
	d.logger.Info("Disconnected from database")
	return err
}

// method for insert single value in a collection
func (d *MongoDatabase) InsertOne(collection string, data interface{}) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	result, err := d.db.Collection(collection).InsertOne(ctx, data)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), err
}

// method for find single value in a collection
func (d *MongoDatabase) FindOne(collection string, filter interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	err := d.db.Collection(collection).FindOne(ctx, filter).Decode(result)
	return err
}

// method for find multiple values in a collection
func (d *MongoDatabase) FindMany(collection string, filter interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	cursor, err := d.db.Collection(collection).Find(ctx, filter)
	if err != nil {
		return err
	}
	err = cursor.All(ctx, result)

	if err != nil {
		return err
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

// method for find and sort multiple values in a collection
func (d *MongoDatabase) FindManySorted(collection string, filter interface{}, sort interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	opts := options.Find().SetSort(sort)
	cursor, err := d.db.Collection(collection).Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	err = cursor.All(ctx, result)

	if err != nil {
		return err
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

// Aggregate One
func (d *MongoDatabase) AggregateOne(collection string, pipeline interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	cursor, err := d.db.Collection(collection).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	var num int
	for cursor.Next(ctx) {
		num++
		err := cursor.Decode(result)
		if err != nil {
			return err
		}
	}

	if num > 1 {
		return fmt.Errorf("expected 1 result, got %d", num)
	}
	if num == 0 {
		return mongo.ErrNoDocuments
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

// Aggregate Many
func (d *MongoDatabase) AggregateMany(collection string, pipeline interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	cursor, err := d.db.Collection(collection).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	err = cursor.All(ctx, result)

	if err != nil {
		return err
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

// method for update single value in a collection
func (d *MongoDatabase) UpdateOne(collection string, filter interface{}, update interface{}) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetProjection(bson.M{"_id": 1})

	var updatedDocument bson.M
	err := d.db.Collection(collection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDocument)

	if err != nil {
		return primitive.NilObjectID, err
	}

	updatedID, ok := updatedDocument["_id"].(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("failed to get updated id")
	}

	return updatedID, nil
}

// method for upsert single value in a collection
func (d *MongoDatabase) UpsertOne(collection string, filter interface{}, update interface{}) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After).SetProjection(bson.M{"_id": 1})

	var upsertedDocument bson.M
	err := d.db.Collection(collection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&upsertedDocument)

	if err != nil {
		return primitive.NilObjectID, err
	}

	upsertedID, ok := upsertedDocument["_id"].(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("failed to get upserted id")
	}

	return upsertedID, nil
}

// InitDB creates a new database wrapper
func InitDB(config models.MongoConfig) {
	d := &MongoDatabase{
		uri:      config.URI,
		database: config.Database,
		timeout:  time.Duration(config.TimeoutMS) * time.Millisecond,
		logger:   log.WithFields(log.Fields{"module": "database"}),
	}

	err := d.Connect()
	if err != nil {
		d.logger.
			WithError(err).
			Fatal("Failed to connect to database")
	}
	err = d.SetupIndexes()
	if err != nil {
		d.logger.
			WithError(err).
			Fatal("Failed to setup indexes")
	}
	err = d.SetupLocker()
	if err != nil {
		d.logger.
			WithError(err).
			Fatal("Failed to setup locker")
	}
	d.logger.Info("Database initialized")

	mongoDB = d
}

func DisconnectDB() {
	if mongoDB != nil {
		mongoDB.Disconnect()
	}
}
