package health

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type HealthCheckRunner struct {
	cosmosAddress string
	ethAddress    string
	hostname      string
	oracleID      string
	services      []service.ChainServiceInterface

	logger *log.Entry
}

func (x *HealthCheckRunner) Run() {
	x.PostHealth()
}

func (x *HealthCheckRunner) AddServices(services []service.ChainServiceInterface) {
	x.services = services
}

func (x *HealthCheckRunner) GetLastHealth() (*models.Node, error) {
	filter := bson.M{
		"cosmos_address": x.cosmosAddress,
		"eth_address":    x.ethAddress,
		"hostname":       x.hostname,
		"oracle_id":      x.oracleID,
	}
	health, err := db.FindNode(filter)
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
		"oracle_id":      x.oracleID,
	}

	onInsert := bson.M{
		"cosmos_address": x.cosmosAddress,
		"eth_address":    x.ethAddress,
		"hostname":       x.hostname,
		"oracle_id":      x.oracleID,
		"created_at":     time.Now(),
	}

	onUpdate := bson.M{
		"healthy":         true,
		"service_healths": x.ServiceHealths(),
		"updated_at":      time.Now(),
	}

	err := db.UpsertNode(filter, onUpdate, onInsert)

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

	logger.
		WithField("eth_address", ethAddressHex).
		Debugf("Initialized ethereum address")

	cosmosPubKey, _ := common.CosmosPublicKeyFromMnemonic(config.Mnemonic)

	cosmosPubKeyHex := hex.EncodeToString(cosmosPubKey.Bytes())

	cosmosAddress := cosmosPubKey.Address().Bytes()

	cosmosAddressHex := hex.EncodeToString(cosmosAddress)

	logger.
		WithField("cosmos_address", cosmosAddressHex).
		Debug("Initialized cosmos address")

	signerIndex := -1
	for i, pk := range config.CosmosNetwork.MultisigPublicKeys {
		if strings.EqualFold(pk, cosmosPubKeyHex) {
			signerIndex = i
		}
	}

	if signerIndex == -1 {
		logger.Fatal("Multisig public keys do not contain signer")
	}

	oracleID := "oracle-" + fmt.Sprintf("%02d", signerIndex)

	hostname, err := os.Hostname()
	if err != nil {
		logger.Fatal("Error getting hostname: ", err)
	}

	x := &HealthCheckRunner{
		cosmosAddress: common.Ensure0xPrefix(cosmosAddressHex),
		ethAddress:    common.Ensure0xPrefix(ethAddressHex),
		hostname:      hostname,
		oracleID:      oracleID,
		logger:        logger,
	}

	x.logger.Info("Initialized health")

	return x
}
