package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/models"
)

func LoadConfigFromEnv(envFile string) models.Config {

	// Load environment variables from a .env file if needed
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			log.Warn("[CONFIG] Error loading env file: ", err.Error())
		} else {
			log.Debug("[CONFIG] Loading env file: ", envFile)
		}
	} else {
		log.Debug("[CONFIG] No env file provided")
	}

	log.Debug("[CONFIG] Loading config from env")

	var config models.Config

	// Set values from environment variables
	config.HealthCheck.IntervalMS = getInt64Env("HEALTH_CHECK_INTERVAL_MS")
	config.HealthCheck.ReadLastHealth = getBoolEnv("HEALTH_CHECK_READ_LAST_HEALTH")
	config.Logger.Level = getStringEnv("LOGGER_LEVEL")
	config.MongoDB.URI = getStringEnv("MONGODB_URI")
	config.MongoDB.Database = getStringEnv("MONGODB_DATABASE")
	config.MongoDB.TimeoutMS = getInt64Env("MONGODB_TIMEOUT_MS")

	// Mnemonic for both Ethereum and Cosmos networks
	config.Mnemonic = getStringEnv("MNEMONIC")

	// Ethereum Networks
	numEthereumNetworks := getArrayLengthEnv("ETHEREUM_NETWORKS")
	config.EthereumNetworks = make([]models.EthereumNetworkConfig, numEthereumNetworks)
	for i := 0; i < numEthereumNetworks; i++ {
		config.EthereumNetworks[i] = models.EthereumNetworkConfig{
			StartBlockHeight:      getInt64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_START_BLOCK_HEIGHT"),
			Confirmations:         getInt64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_CONFIRMATIONS"),
			RPCURL:                getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_RPC_URL"),
			RPCTimeoutMS:          getInt64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_RPC_TIMEOUT_MS"),
			ChainID:               getInt64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_ID"),
			ChainName:             getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_NAME"),
			MailboxAddress:        getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MAILBOX_ADDRESS"),
			MintControllerAddress: getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MINT_CONTROLLER_ADDRESS"),
			OracleAddresses:       getStringArrayEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_ORACLE_ADDRESSES"),
			MessageMonitor: models.ServiceConfig{
				Enabled:    getBoolEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_ENABLED"),
				IntervalMS: getInt64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_INTERVAL_MS"),
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    getBoolEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_ENABLED"),
				IntervalMS: getInt64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_INTERVAL_MS"),
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    getBoolEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_ENABLED"),
				IntervalMS: getInt64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_INTERVAL_MS"),
			},
		}
	}

	// Cosmos Networks
	numCosmosNetworks := getArrayLengthEnv("COSMOS_NETWORKS")
	config.CosmosNetworks = make([]models.CosmosNetworkConfig, numCosmosNetworks)
	for i := 0; i < numCosmosNetworks; i++ {
		config.CosmosNetworks[i] = models.CosmosNetworkConfig{
			StartBlockHeight:   getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_START_BLOCK_HEIGHT"),
			Confirmations:      getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_CONFIRMATIONS"),
			GRPCHost:           getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_GRPC_HOST"),
			GRPCPort:           getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_GRPC_PORT"),
			GRPCTimeoutMS:      getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_GRPC_TIMEOUT_MS"),
			ChainID:            getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_ID"),
			ChainName:          getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_NAME"),
			TxFee:              getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_TX_FEE"),
			Bech32Prefix:       getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_BECH32_PREFIX"),
			MultisigAddress:    getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MULTISIG_ADDRESS"),
			MultisigPublicKeys: getStringArrayEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MULTISIG_PUBLIC_KEYS"),
			MultisigThreshold:  getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MULTISIG_THRESHOLD"),
			MessageMonitor: models.ServiceConfig{
				Enabled:    getBoolEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_ENABLED"),
				IntervalMS: getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_INTERVAL_MS"),
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    getBoolEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_ENABLED"),
				IntervalMS: getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_INTERVAL_MS"),
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    getBoolEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_ENABLED"),
				IntervalMS: getInt64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_INTERVAL_MS"),
			},
		}
	}

	log.Debug("[CONFIG] Config loaded from env")

	return config
}

// Helper functions to retrieve environment variables with fallback values
func getInt64Env(key string) int64 {
	valStr := os.Getenv(key)
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0 // Default value
	}
	return val
}

func getBoolEnv(key string) bool {
	valStr := os.Getenv(key)
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false // Default value
	}
	return val
}

func getStringEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		return "" // Default value
	}
	return val
}

func getStringArrayEnv(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return []string{} // Default value
	}
	return strings.Split(val, ",")
}

func getArrayLengthEnv(key string) int {
	val := os.Getenv(key)
	if val == "" {
		return 0 // Default value
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return num
}
