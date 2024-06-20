package health

import (
	"os"
	"testing"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mocks "github.com/dan13ram/wpokt-oracle/db/mocks"

	log "github.com/sirupsen/logrus"
)

type mockChainService struct {
	health models.ChainServiceHealth
}

func (m *mockChainService) Start() {}
func (m *mockChainService) Stop()  {}
func (m *mockChainService) Health() models.ChainServiceHealth {
	return m.health
}

func TesthealthCheckRunnable_Run(t *testing.T) {
	healthCheck := &healthCheckRunnable{
		logger: log.NewEntry(log.New()),
	}
	oldMongoDB := db.MongoDB
	mockDB := mocks.NewMockDatabase(t)
	db.MongoDB = mockDB
	defer func() {
		mockDB.AssertExpectations(t)
		db.MongoDB = oldMongoDB
	}()

	mockDB.On("UpsertOne", common.CollectionNodes, mock.Anything, mock.Anything).Return(primitive.ObjectID{}, nil).Once()

	healthCheck.Run()
}

func TesthealthCheckRunnable_AddServices(t *testing.T) {
	healthCheck := &healthCheckRunnable{}
	services := []service.ChainService{
		&mockChainService{},
	}
	healthCheck.AddServices(services)
	assert.Equal(t, services, healthCheck.services)
}

func TesthealthCheckRunnable_GetLastHealth(t *testing.T) {
	healthCheck := &healthCheckRunnable{
		cosmosAddress: "cosmosAddress",
		ethAddress:    "ethAddress",
		hostname:      "hostname",
		oracleID:      "oracleID",
	}

	oldMongoDB := db.MongoDB
	mockDB := mocks.NewMockDatabase(t)
	db.MongoDB = mockDB
	defer func() {
		mockDB.AssertExpectations(t)
		db.MongoDB = oldMongoDB
	}()

	expectedFilter := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
	}
	mockDB.On("FindOne", common.CollectionNodes, expectedFilter, &models.Node{}).Return(nil).Once()

	health, err := healthCheck.GetLastHealth()
	assert.NoError(t, err)
	assert.NotNil(t, health)
}

func TesthealthCheckRunnable_ServiceHealths(t *testing.T) {
	serviceHealth := models.ChainServiceHealth{
		Chain: models.Chain{
			ChainName: "TestChain",
		},
	}
	healthCheck := &healthCheckRunnable{
		services: []service.ChainService{
			&mockChainService{health: serviceHealth},
		},
	}
	healths := healthCheck.ServiceHealths()
	assert.Equal(t, 1, len(healths))
	assert.Equal(t, serviceHealth, healths[0])
}

func TesthealthCheckRunnable_PostHealth(t *testing.T) {
	healthCheck := &healthCheckRunnable{
		cosmosAddress: "cosmosAddress",
		ethAddress:    "ethAddress",
		hostname:      "hostname",
		oracleID:      "oracleID",
		logger:        log.NewEntry(log.New()),
	}

	oldMongoDB := db.MongoDB
	mockDB := mocks.NewMockDatabase(t)
	db.MongoDB = mockDB
	defer func() {
		mockDB.AssertExpectations(t)
		db.MongoDB = oldMongoDB
	}()

	expectedFilter := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
	}

	onInsert := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
		"created_at":     nil,
	}

	onUpdate := bson.M{
		"healthy":         true,
		"service_healths": nil,
		"updated_at":      nil,
	}

	expectedUpdate := bson.M{"$set": onUpdate, "$setOnInsert": onInsert}

	mockDB.On("UpsertOne", common.CollectionNodes, expectedFilter, mock.Anything).Return(primitive.ObjectID{}, nil).Once().Run(func(args mock.Arguments) {
		update := args.Get(2).(bson.M)
		assert.NotNil(t, update["$set"])
		assert.NotNil(t, update["$setOnInsert"])

		update["$setOnInsert"].(bson.M)["created_at"] = nil
		update["$set"].(bson.M)["service_healths"] = nil
		update["$set"].(bson.M)["updated_at"] = nil

		assert.Equal(t, expectedUpdate, update)
	})

	success := healthCheck.PostHealth()
	assert.True(t, success)
}

func TestNewHealthCheck(t *testing.T) {
	config := models.Config{
		Mnemonic: "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve",
		CosmosNetwork: models.CosmosNetworkConfig{
			MultisigPublicKeys: []string{
				"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9",
				"02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2",
				"02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df",
			},
		},
	}

	healthCheck := newHealthCheck(config)
	assert.NotNil(t, healthCheck)
	assert.Equal(t, common.Ensure0xPrefix("0x0e90a32df6f6143f1a91c25d9552dcbc789c34eb"), healthCheck.ethAddress)
	assert.Equal(t, common.Ensure0xPrefix("0x3f23b2b1de52d246657a4ec3ca69c7b04b3c739d"), healthCheck.cosmosAddress)
	assert.Equal(t, "oracle-00", healthCheck.oracleID)

	hostname, err := os.Hostname()
	assert.NoError(t, err)
	assert.Equal(t, hostname, healthCheck.hostname)
}

// func TestNewHealthCheck_MissingSigner(t *testing.T) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			assert.Equal(t, "Multisig public keys do not contain signer", r)
// 		}
// 	}()
//
// 	config := models.Config{
// 		Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
// 		CosmosNetwork: models.CosmosNetworkConfig{
// 			MultisigPublicKeys: []string{
// 				"cosmos1nxyyrxs69w4qf9cwt8r0w9pw4z5uzhrx38p2s5",
// 			},
// 		},
// 	}
//
// 	newHealthCheck(config)
// 	t.Fail() // Should not reach here
// }
