package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func loadConfigFromEnv(envFile string) (models.Config, error) {

	// Load environment variables from a .env file if needed
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			logger.
				WithFields(log.Fields{"error": err}).
				Warn("Error loading env file")
			return models.Config{}, err
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

	numEthereumNetworksEnv := getUint64Env("NUM_ETHEREUM_NETWORKS")

	// Ethereum Networks
	numEthereumNetworks := getArrayLengthEnv("ETHEREUM_NETWORKS")

	if numEthereumNetworks != int(numEthereumNetworksEnv) {
		logger.
			WithFields(log.Fields{"numEthereumNetworks": numEthereumNetworks, "numEthereumNetworksEnv": numEthereumNetworksEnv}).
			Fatal("Number of Ethereum networks does not match")
	}

	if numEthereumNetworks > 0 {
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
	}

	// Cosmos Networks
	config.CosmosNetwork = models.CosmosNetworkConfig{
		StartBlockHeight:   getUint64Env("COSMOS_NETWORK_START_BLOCK_HEIGHT"),
		Confirmations:      getUint64Env("COSMOS_NETWORK_CONFIRMATIONS"),
		RPCURL:             getStringEnv("COSMOS_NETWORK_RPC_URL"),
		GRPCEnabled:        getBoolEnv("COSMOS_NETWORK_GRPC_ENABLED"),
		GRPCHost:           getStringEnv("COSMOS_NETWORK_GRPC_HOST"),
		GRPCPort:           getUint64Env("COSMOS_NETWORK_GRPC_PORT"),
		TimeoutMS:          getUint64Env("COSMOS_NETWORK_TIMEOUT_MS"),
		ChainID:            getStringEnv("COSMOS_NETWORK_CHAIN_ID"),
		ChainName:          getStringEnv("COSMOS_NETWORK_CHAIN_NAME"),
		TxFee:              getUint64Env("COSMOS_NETWORK_TX_FEE"),
		Bech32Prefix:       getStringEnv("COSMOS_NETWORK_BECH32_PREFIX"),
		CoinDenom:          getStringEnv("COSMOS_NETWORK_COIN_DENOM"),
		MultisigAddress:    getStringEnv("COSMOS_NETWORK_MULTISIG_ADDRESS"),
		MultisigPublicKeys: getStringArrayEnv("COSMOS_NETWORK_MULTISIG_PUBLIC_KEYS"),
		MultisigThreshold:  getUint64Env("COSMOS_NETWORK_MULTISIG_THRESHOLD"),
		MessageMonitor: models.ServiceConfig{
			Enabled:    getBoolEnv("COSMOS_NETWORK_MESSAGE_MONITOR_ENABLED"),
			IntervalMS: getUint64Env("COSMOS_NETWORK_MESSAGE_MONITOR_INTERVAL_MS"),
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    getBoolEnv("COSMOS_NETWORK_MESSAGE_SIGNER_ENABLED"),
			IntervalMS: getUint64Env("COSMOS_NETWORK_MESSAGE_SIGNER_INTERVAL_MS"),
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    getBoolEnv("COSMOS_NETWORK_MESSAGE_RELAYER_ENABLED"),
			IntervalMS: getUint64Env("COSMOS_NETWORK_MESSAGE_RELAYER_INTERVAL_MS"),
		},
	}

	logger.Debug("Config loaded from env")

	return config, nil
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
		return nil // Default value
	}
	return strings.Split(val, ",")
}

func getArrayLengthEnv(key string) int {
	env := os.Environ()
	var length int
	seen := make(map[int]bool)
	for _, e := range env {
		pair := strings.Split(e, "=")
		if strings.HasPrefix(pair[0], key) {
			vals := strings.Split(strings.TrimPrefix(pair[0], key), "_")
			log.Println(vals)
			val, err := strconv.Atoi(vals[1])
			if err != nil {
				return 0
			}
			if !seen[val] {
				length++
				seen[val] = true
			}
		}
	}
	return length
}
