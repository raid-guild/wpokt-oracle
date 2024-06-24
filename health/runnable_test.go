package health

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/dan13ram/wpokt-oracle/common"
	mocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"

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

func Test_HealthCheckRunnable_Run(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	healthCheck := &healthCheckRunnable{
		logger: log.NewEntry(log.New()),
		db:     mockDB,
	}

	mockDB.EXPECT().UpsertNode(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	healthCheck.Run()
	mockDB.AssertExpectations(t)
}

func Test_HealthCheckRunnable_AddServices(t *testing.T) {
	healthCheck := &healthCheckRunnable{}
	services := []service.ChainService{
		&mockChainService{},
	}
	healthCheck.AddServices(services)
	assert.Equal(t, services, healthCheck.services)
}

func Test_HealthCheckRunnable_GetLastHealth(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	healthCheck := &healthCheckRunnable{
		cosmosAddress: "cosmosAddress",
		ethAddress:    "ethAddress",
		hostname:      "hostname",
		oracleID:      "oracleID",
		db:            mockDB,
	}

	expectedFilter := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
	}
	mockDB.EXPECT().FindNode(expectedFilter).Return(&models.Node{}, nil).Once()

	health, err := healthCheck.GetLastHealth()
	assert.NoError(t, err)
	assert.NotNil(t, health)
	mockDB.AssertExpectations(t)
}

func Test_HealthCheckRunnable_ServiceHealths(t *testing.T) {
	serviceHealth := models.ChainServiceHealth{
		Chain: models.Chain{
			ChainName: "TestChain",
		},
	}
	mockDB := mocks.NewMockDB(t)
	healthCheck := &healthCheckRunnable{
		services: []service.ChainService{
			&mockChainService{health: serviceHealth},
		},
		db: mockDB,
	}
	healths := healthCheck.ServiceHealths()
	assert.Equal(t, 1, len(healths))
	assert.Equal(t, serviceHealth, healths[0])
}

func Test_HealthCheckRunnable_PostHealth(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	healthCheck := &healthCheckRunnable{
		cosmosAddress: "cosmosAddress",
		ethAddress:    "ethAddress",
		hostname:      "hostname",
		oracleID:      "oracleID",
		logger:        log.NewEntry(log.New()),
		db:            mockDB,
	}

	expectedFilter := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
	}

	expectedOnInsert := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
		"created_at":     nil,
	}

	expectedOnUpdate := bson.M{
		"healthy":         true,
		"service_healths": nil,
		"updated_at":      nil,
	}

	mockDB.EXPECT().UpsertNode(expectedFilter, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		onUpdate := args.Get(1).(bson.M)
		onInsert := args.Get(2).(bson.M)
		assert.NotNil(t, onUpdate)
		assert.NotNil(t, onInsert)

		onInsert["created_at"] = nil
		onUpdate["service_healths"] = nil
		onUpdate["updated_at"] = nil

		assert.Equal(t, expectedOnUpdate, onUpdate)
		assert.Equal(t, expectedOnInsert, onInsert)
	})

	success := healthCheck.PostHealth()
	assert.True(t, success)
	mockDB.AssertExpectations(t)
}

func Test_HealthCheckRunnable_PostHealth_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	healthCheck := &healthCheckRunnable{
		cosmosAddress: "cosmosAddress",
		ethAddress:    "ethAddress",
		hostname:      "hostname",
		oracleID:      "oracleID",
		logger:        log.NewEntry(log.New()),
		db:            mockDB,
	}

	expectedFilter := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
	}

	expectedOnInsert := bson.M{
		"cosmos_address": "cosmosAddress",
		"eth_address":    "ethAddress",
		"hostname":       "hostname",
		"oracle_id":      "oracleID",
		"created_at":     nil,
	}

	expectedOnUpdate := bson.M{
		"healthy":         true,
		"service_healths": nil,
		"updated_at":      nil,
	}

	mockDB.EXPECT().UpsertNode(expectedFilter, mock.Anything, mock.Anything).Return(errors.New("error")).Once().Run(func(args mock.Arguments) {
		onUpdate := args.Get(1).(bson.M)
		onInsert := args.Get(2).(bson.M)
		assert.NotNil(t, onUpdate)
		assert.NotNil(t, onInsert)

		onInsert["created_at"] = nil
		onUpdate["service_healths"] = nil
		onUpdate["updated_at"] = nil

		assert.Equal(t, expectedOnUpdate, onUpdate)
		assert.Equal(t, expectedOnInsert, onInsert)
	})

	success := healthCheck.PostHealth()
	assert.False(t, success)
	mockDB.AssertExpectations(t)
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

	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	osHostname = func() (string, error) {
		return "", errors.New("error")
	}
	defer func() { osHostname = os.Hostname }()

	assert.Panics(t, func() {
		newHealthCheck(config)
	})
}

func TestNewHealthCheck_MissingSigner(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	config := models.Config{
		Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		CosmosNetwork: models.CosmosNetworkConfig{
			MultisigPublicKeys: []string{
				"cosmos1nxyyrxs69w4qf9cwt8r0w9pw4z5uzhrx38p2s5",
			},
		},
	}

	assert.Panics(t, func() {
		newHealthCheck(config)
	})
}
