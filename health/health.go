package health

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dan13ram/wpokt-oracle/app"
	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"

	cosmosUtil "github.com/dan13ram/wpokt-oracle/cosmos/util"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type HealthCheckRunner struct {
	cosmosAddress []byte
	ethAddress    []byte
	hostname      string
	oracleId      string
	services      []service.ChainServiceInterface

	logger *log.Entry
}

func (x *HealthCheckRunner) Run() {
	x.PostHealth()
}

func (x *HealthCheckRunner) AddServices(services []service.ChainServiceInterface) {
	x.services = services
}

func (x *HealthCheckRunner) GetLastHealth() (models.Node, error) {
	var health models.Node
	filter := bson.M{
		"cosmos_address": x.cosmosAddress,
		"eth_address":    x.ethAddress,
		"hostname":       x.hostname,
		"oracle_id":      x.oracleId,
	}
	err := app.DB.FindOne(common.CollectionNodes, filter, &health)
	return health, err
}

func (x *HealthCheckRunner) ServiceHealths() []models.ChainServiceHealth {
	var serviceHealths []models.ChainServiceHealth
	for _, service := range x.services {
		serviceHealth := service.Health()
		serviceHealths = append(serviceHealths, serviceHealth)
	}
	return serviceHealths
}

func (x *HealthCheckRunner) PostHealth() bool {
	x.logger.Debug("Posting health")

	filter := bson.M{
		"cosmos_address": x.cosmosAddress,
		"eth_address":    x.ethAddress,
		"hostname":       x.hostname,
		"oracle_id":      x.oracleId,
	}

	onInsert := bson.M{
		"cosmos_address": x.cosmosAddress,
		"eth_address":    x.ethAddress,
		"hostname":       x.hostname,
		"oracle_id":      x.oracleId,
		"created_at":     time.Now(),
	}

	onUpdate := bson.M{
		"healthy":         true,
		"service_healths": x.ServiceHealths(),
		"updated_at":      time.Now(),
	}

	update := bson.M{"$set": onUpdate, "$setOnInsert": onInsert}

	err := app.DB.UpsertOne(common.CollectionNodes, filter, update)

	if err != nil {
		x.logger.Error("Error posting health: ", err)
		return false
	}

	x.logger.Info("Posted health")
	return true
}

func newHealthCheck(config models.Config) *HealthCheckRunner {
	logger := log.WithFields(log.Fields{
		"module": "health",
		"runner": "health",
	})
	logger.Debug("Initializing health")

	ethAddressHex, _ := common.EthereumAddressFromMnemonic(config.Mnemonic)

	ethAddress, _ := hex.DecodeString(ethAddressHex[2:])

	logger.Debugf("ETH Address: %s", ethAddressHex)

	cosmosPubKeyHex, _ := common.CosmosPublicKeyFromMnemonic(config.Mnemonic)

	cosmosPubKey, _ := cosmosUtil.PubKeyFromHex(cosmosPubKeyHex)

	cosmosAddress := cosmosPubKey.Address().Bytes()

	cosmosAddressHex := hex.EncodeToString(cosmosAddress)

	logger.Debugf("Cosmos Address: 0x%s", cosmosAddressHex)

	signerIndex := -1
	for i, pk := range config.CosmosNetworks[0].MultisigPublicKeys {
		if strings.EqualFold(pk, cosmosPubKeyHex) {
			signerIndex = i
		}
	}

	if signerIndex == -1 {
		logger.Fatal("Multisig public keys do not contain signer")
	}

	oracleId := "oracle-" + fmt.Sprintf("%02d", signerIndex)

	hostname, err := os.Hostname()
	if err != nil {
		logger.Fatal("Error getting hostname: ", err)
	}

	x := &HealthCheckRunner{
		cosmosAddress: cosmosAddress,
		ethAddress:    ethAddress,
		hostname:      hostname,
		oracleId:      oracleId,
		logger:        logger,
	}

	x.logger.Info("Initialized health")

	return x
}
