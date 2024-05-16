package app

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dan13ram/wpokt-oracle/app/service"
	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"

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
	err := DB.FindOne(CollectionNodes, filter, &health)
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
	log.Debug("[HEALTH] Posting health")

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

	err := DB.UpsertOne(CollectionNodes, filter, update)

	if err != nil {
		log.Error("[HEALTH] Error posting health: ", err)
		return false
	}

	log.Info("[HEALTH] Posted health")
	return true
}

func newHealthCheck(config models.Config) *HealthCheckRunner {
	log.Debug("[HEALTH] Initializing health")

	ethAddressHex, _ := common.EthereumAddressFromMnemonic(config.Mnemonic)

	ethAddress, _ := hex.DecodeString(ethAddressHex[2:])

	log.Debugf("[HEALTH] ETH Address: %s", ethAddressHex)

	cosmosPubKeyHex, _ := common.CosmosPublicKeyFromMnemonic(config.Mnemonic)

	cosmosPubKey, _ := cosmosUtil.PubKeyFromHex(cosmosPubKeyHex)

	cosmosAddress := cosmosPubKey.Address().Bytes()

	cosmosAddressHex := hex.EncodeToString(cosmosAddress)

	log.Debugf("[HEALTH] Cosmos Address: 0x%s", cosmosAddressHex)

	signerIndex := -1
	for i, pk := range config.CosmosNetworks[0].MultisigPublicKeys {
		if strings.EqualFold(pk, cosmosPubKeyHex) {
			signerIndex = i
		}
	}

	if signerIndex == -1 {
		log.Fatal("[HEALTH] Multisig public keys do not contain signer")
	}

	oracleId := "oracle-" + fmt.Sprintf("%02d", signerIndex)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("[HEALTH] Error getting hostname: ", err)
	}

	x := &HealthCheckRunner{
		cosmosAddress: cosmosAddress,
		ethAddress:    ethAddress,
		hostname:      hostname,
		oracleId:      oracleId,
	}

	log.Info("[HEALTH] Initialized health")

	return x
}
