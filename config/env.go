package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func loadConfigFromEnv(envFile string) models.Config {

	// Load environment variables from a .env file if needed
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			logger.
				WithFields(log.Fields{"error": err}).
				Warn("Error loading env file")
		} else {
			logger.Debug("Loading env file")
		}
	} else {
		logger.Debug("No env file provided")
	}

	logger.Debug("Loading config from env")

	var config models.Config

	// Set values from environment variables
	config.HealthCheck.IntervalMS = getUint64Env("HEALTH_CHECK_INTERVAL_MS")
	config.HealthCheck.ReadLastHealth = getBoolEnv("HEALTH_CHECK_READ_LAST_HEALTH")
	config.Logger.Level = getStringEnv("LOGGER_LEVEL")
	config.Logger.Format = getStringEnv("LOGGER_FORMAT")
	config.MongoDB.URI = getStringEnv("MONGODB_URI")
	config.MongoDB.Database = getStringEnv("MONGODB_DATABASE")
	config.MongoDB.TimeoutMS = getUint64Env("MONGODB_TIMEOUT_MS")

	// Mnemonic for both Ethereum and Cosmos networks
	config.Mnemonic = getStringEnv("MNEMONIC")

	// Ethereum Networks
	numEthereumNetworks := getArrayLengthEnv("ETHEREUM_NETWORKS")
	config.EthereumNetworks = make([]models.EthereumNetworkConfig, numEthereumNetworks)
	for i := 0; i < numEthereumNetworks; i++ {
		config.EthereumNetworks[i] = models.EthereumNetworkConfig{
			StartBlockHeight:      getUint64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_START_BLOCK_HEIGHT"),
			Confirmations:         getUint64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_CONFIRMATIONS"),
			RPCURL:                getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_RPC_URL"),
			TimeoutMS:             getUint64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_TIMEOUT_MS"),
			ChainID:               getUint64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_ID"),
			ChainName:             getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_NAME"),
			MailboxAddress:        getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MAILBOX_ADDRESS"),
			MintControllerAddress: getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MINT_CONTROLLER_ADDRESS"),
			OmniTokenAddress:      getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_OMNI_TOKEN_ADDRESS"),
			WarpISMAddress:        getStringEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_WARP_ISM_ADDRESS"),
			OracleAddresses:       getStringArrayEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_ORACLE_ADDRESSES"),
			MessageMonitor: models.ServiceConfig{
				Enabled:    getBoolEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_ENABLED"),
				IntervalMS: getUint64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_INTERVAL_MS"),
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    getBoolEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_ENABLED"),
				IntervalMS: getUint64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_INTERVAL_MS"),
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    getBoolEnv("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_ENABLED"),
				IntervalMS: getUint64Env("ETHEREUM_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_INTERVAL_MS"),
			},
		}
	}

	// Cosmos Networks
	numCosmosNetworks := getArrayLengthEnv("COSMOS_NETWORKS")
	config.CosmosNetworks = make([]models.CosmosNetworkConfig, numCosmosNetworks)
	for i := 0; i < numCosmosNetworks; i++ {
		config.CosmosNetworks[i] = models.CosmosNetworkConfig{
			StartBlockHeight:   getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_START_BLOCK_HEIGHT"),
			Confirmations:      getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_CONFIRMATIONS"),
			RPCURL:             getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_RPC_URL"),
			GRPCEnabled:        getBoolEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_GRPC_ENABLED"),
			GRPCHost:           getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_GRPC_HOST"),
			GRPCPort:           getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_GRPC_PORT"),
			TimeoutMS:          getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_TIMEOUT_MS"),
			ChainID:            getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_ID"),
			ChainName:          getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_CHAIN_NAME"),
			TxFee:              getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_TX_FEE"),
			Bech32Prefix:       getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_BECH32_PREFIX"),
			CoinDenom:          getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_COIN_DENOM"),
			MultisigAddress:    getStringEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MULTISIG_ADDRESS"),
			MultisigPublicKeys: getStringArrayEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MULTISIG_PUBLIC_KEYS"),
			MultisigThreshold:  getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MULTISIG_THRESHOLD"),
			MessageMonitor: models.ServiceConfig{
				Enabled:    getBoolEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_ENABLED"),
				IntervalMS: getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_MONITOR_INTERVAL_MS"),
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    getBoolEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_ENABLED"),
				IntervalMS: getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_SIGNER_INTERVAL_MS"),
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    getBoolEnv("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_ENABLED"),
				IntervalMS: getUint64Env("COSMOS_NETWORKS_" + strconv.Itoa(i) + "_MESSAGE_RELAYER_INTERVAL_MS"),
			},
		}
	}

	logger.Debug("Config loaded from env")

	return config
}

// Helper functions to retrieve environment variables with fallback values
func getUint64Env(key string) uint64 {
	valStr := os.Getenv(key)
	val, err := strconv.ParseUint(valStr, 10, 64)
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
