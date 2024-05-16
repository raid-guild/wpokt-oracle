package app

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	lock "github.com/square/mongo-lock"
)

const (
	CollectionLocks        = "locks"
	CollectionTransactions = "transactions"
	CollectionMessages     = "messages"
	CollectionNodes        = "nodes"
)

type Database interface {
	Connect() error
	Disconnect() error

	InsertOne(collection string, data interface{}) error
	FindOne(collection string, filter interface{}, result interface{}) error
	FindMany(collection string, filter interface{}, result interface{}) error
	UpdateOne(collection string, filter interface{}, update interface{}) error
	UpsertOne(collection string, filter interface{}, update interface{}) error

	XLock(resourceId string) (string, error)
	SLock(resourceId string) (string, error)
	Unlock(lockId string) error
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

var (
	DB Database
)

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
	/* TODO: setup the locker
	d.logger.Debug("Setting up locker")

	var locker *lock.Client

	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	locker = lock.NewClient(d.db.Collection("locks"))
	err := locker.CreateIndexes(ctx)
	if err != nil {
		return err
	}

	d.locker = locker

	d.logger.Info("Locker setup")
	*/
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
func (d *MongoDatabase) XLock(resourceId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	lockId, err := randomString(32)
	if err != nil {
		return "", err
	}
	err = d.locker.XLock(ctx, resourceId, lockId, lock.LockDetails{
		TTL: 60, // locks expire in 60 seconds
	})
	return lockId, err
}

// SLock locks a resource for shared access
func (d *MongoDatabase) SLock(resourceId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	lockId, err := randomString(32)
	if err != nil {
		return "", err
	}
	err = d.locker.SLock(ctx, resourceId, lockId, lock.LockDetails{
		TTL: 60, // locks expire in 60 seconds
	}, -1)
	return lockId, err
}

// Unlock unlocks a resource
func (d *MongoDatabase) Unlock(lockId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	_, err := d.locker.Unlock(ctx, lockId)
	return err
}

// Setup Indexes
func (d *MongoDatabase) SetupIndexes() error {
	/* TODO: setup the right indexes
	d.logger.Debug("Setting up indexes")

	// setup unique index for mints
	d.logger.Debug("Setting up indexes for mints")
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err := d.db.Collection(models.CollectionMints).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "transaction_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// setup unique index for invalid mints
	d.logger.Debug("Setting up indexes for invalid mints")
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(models.CollectionInvalidMints).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "transaction_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// setup unique index for burns
	d.logger.Debug("Setting up indexes for burns")
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(models.CollectionBurns).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "transaction_hash", Value: 1}, {Key: "log_index", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// setup unique index for healthchecks
	d.logger.Debug("Setting up indexes for healthchecks")
	ctx, cancel = context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err = d.db.Collection(models.CollectionHealthChecks).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "validator_id", Value: 1}, {Key: "hostname", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	d.logger.Info("Indexes setup")
	*/

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
func (d *MongoDatabase) InsertOne(collection string, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err := d.db.Collection(collection).InsertOne(ctx, data)
	return err
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
	return err
}

// method for update single value in a collection
func (d *MongoDatabase) UpdateOne(collection string, filter interface{}, update interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	_, err := d.db.Collection(collection).UpdateOne(ctx, filter, update)
	return err
}

// method for upsert single value in a collection
func (d *MongoDatabase) UpsertOne(collection string, filter interface{}, update interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	_, err := d.db.Collection(collection).UpdateOne(ctx, filter, update, opts)
	return err
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
		d.logger.Fatal("Failed to connect to database: ", err)
	}
	err = d.SetupIndexes()
	if err != nil {
		d.logger.Fatal("Failed to setup indexes: ", err)
	}
	err = d.SetupLocker()
	if err != nil {
		d.logger.Fatal("Failed to setup locker: ", err)
	}
	d.logger.Info("Database initialized")

	DB = d
}
