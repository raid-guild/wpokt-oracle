package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/models"
)

// Function to merge two Config structs, prioritizing non-empty configurations from the envConfig
func MergeConfigs(yamlConfig models.Config, envConfig models.Config) models.Config {

	log.Debug("[CONFIG] Merging configs from YAML and ENV")
	// Create a new Config instance to store the merged values
	mergedConfig := yamlConfig

	// Merge HealthCheck
	if envConfig.HealthCheck.IntervalMS != 0 {
		mergedConfig.HealthCheck.IntervalMS = envConfig.HealthCheck.IntervalMS
	}
	if envConfig.HealthCheck.ReadLastHealth {
		mergedConfig.HealthCheck.ReadLastHealth = envConfig.HealthCheck.ReadLastHealth
	}

	// Merge Logger
	if envConfig.Logger.Level != "" {
		mergedConfig.Logger.Level = envConfig.Logger.Level
	}

	// Merge MongoDB
	if envConfig.MongoDB.URI != "" {
		mergedConfig.MongoDB.URI = envConfig.MongoDB.URI
	}
	if envConfig.MongoDB.Database != "" {
		mergedConfig.MongoDB.Database = envConfig.MongoDB.Database
	}
	if envConfig.MongoDB.TimeoutMS != 0 {
		mergedConfig.MongoDB.TimeoutMS = envConfig.MongoDB.TimeoutMS
	}

	// Merge EthereumNetworks
	for i, envEthNet := range envConfig.EthereumNetworks {
		if i < len(mergedConfig.EthereumNetworks) {
			if envEthNet.StartBlockNumber != 0 {
				mergedConfig.EthereumNetworks[i].StartBlockNumber = envEthNet.StartBlockNumber
			}
			if envEthNet.Confirmations != 0 {
				mergedConfig.EthereumNetworks[i].Confirmations = envEthNet.Confirmations
			}
			if envEthNet.PrivateKey != "" {
				mergedConfig.EthereumNetworks[i].PrivateKey = envEthNet.PrivateKey
			}
			if envEthNet.RPCURL != "" {
				mergedConfig.EthereumNetworks[i].RPCURL = envEthNet.RPCURL
			}
			if envEthNet.RPCTimeoutMS != 0 {
				mergedConfig.EthereumNetworks[i].RPCTimeoutMS = envEthNet.RPCTimeoutMS
			}
			if envEthNet.ChainId != "" {
				mergedConfig.EthereumNetworks[i].ChainId = envEthNet.ChainId
			}
			if envEthNet.MailboxAddress != "" {
				mergedConfig.EthereumNetworks[i].MailboxAddress = envEthNet.MailboxAddress
			}
			if envEthNet.MintControllerAddress != "" {
				mergedConfig.EthereumNetworks[i].MintControllerAddress = envEthNet.MintControllerAddress
			}
			if len(envEthNet.OracleAddresses) != 0 {
				mergedConfig.EthereumNetworks[i].OracleAddresses = envEthNet.OracleAddresses
			}
			if envEthNet.MessageMonitor.Enabled {
				mergedConfig.EthereumNetworks[i].MessageMonitor.Enabled = envEthNet.MessageMonitor.Enabled
			}
			if envEthNet.MessageMonitor.IntervalMS != 0 {
				mergedConfig.EthereumNetworks[i].MessageMonitor.IntervalMS = envEthNet.MessageMonitor.IntervalMS
			}
			if envEthNet.MessageSigner.Enabled {
				mergedConfig.EthereumNetworks[i].MessageSigner.Enabled = envEthNet.MessageSigner.Enabled
			}
			if envEthNet.MessageSigner.IntervalMS != 0 {
				mergedConfig.EthereumNetworks[i].MessageSigner.IntervalMS = envEthNet.MessageSigner.IntervalMS
			}
			if envEthNet.MessageProcessor.Enabled {
				mergedConfig.EthereumNetworks[i].MessageProcessor.Enabled = envEthNet.MessageProcessor.Enabled
			}
			if envEthNet.MessageProcessor.IntervalMS != 0 {
				mergedConfig.EthereumNetworks[i].MessageProcessor.IntervalMS = envEthNet.MessageProcessor.IntervalMS
			}
		}
	}

	// Merge CosmosNetworks
	for i, envCosmosNet := range envConfig.CosmosNetworks {
		if i < len(mergedConfig.CosmosNetworks) {
			if envCosmosNet.StartHeight != 0 {
				mergedConfig.CosmosNetworks[i].StartHeight = envCosmosNet.StartHeight
			}
			if envCosmosNet.Confirmations != 0 {
				mergedConfig.CosmosNetworks[i].Confirmations = envCosmosNet.Confirmations
			}
			if envCosmosNet.PrivateKey != "" {
				mergedConfig.CosmosNetworks[i].PrivateKey = envCosmosNet.PrivateKey
			}
			if envCosmosNet.RPCURL != "" {
				mergedConfig.CosmosNetworks[i].RPCURL = envCosmosNet.RPCURL
			}
			if envCosmosNet.RPCTimeoutMS != 0 {
				mergedConfig.CosmosNetworks[i].RPCTimeoutMS = envCosmosNet.RPCTimeoutMS
			}
			if envCosmosNet.ChainId != "" {
				mergedConfig.CosmosNetworks[i].ChainId = envCosmosNet.ChainId
			}
			if envCosmosNet.TxFee != 0 {
				mergedConfig.CosmosNetworks[i].TxFee = envCosmosNet.TxFee
			}
			if envCosmosNet.MultisigAddress != "" {
				mergedConfig.CosmosNetworks[i].MultisigAddress = envCosmosNet.MultisigAddress
			}
			if len(envCosmosNet.MultisigPublicKeys) != 0 {
				mergedConfig.CosmosNetworks[i].MultisigPublicKeys = envCosmosNet.MultisigPublicKeys
			}
			if envCosmosNet.MultisigThreshold != 0 {
				mergedConfig.CosmosNetworks[i].MultisigThreshold = envCosmosNet.MultisigThreshold
			}
			if envCosmosNet.MessageMonitor.Enabled {
				mergedConfig.CosmosNetworks[i].MessageMonitor.Enabled = envCosmosNet.MessageMonitor.Enabled
			}
			if envCosmosNet.MessageMonitor.IntervalMS != 0 {
				mergedConfig.CosmosNetworks[i].MessageMonitor.IntervalMS = envCosmosNet.MessageMonitor.IntervalMS
			}
			if envCosmosNet.MessageSigner.Enabled {
				mergedConfig.CosmosNetworks[i].MessageSigner.Enabled = envCosmosNet.MessageSigner.Enabled
			}
			if envCosmosNet.MessageSigner.IntervalMS != 0 {
				mergedConfig.CosmosNetworks[i].MessageSigner.IntervalMS = envCosmosNet.MessageSigner.IntervalMS
			}
			if envCosmosNet.MessageProcessor.Enabled {
				mergedConfig.CosmosNetworks[i].MessageProcessor.Enabled = envCosmosNet.MessageProcessor.Enabled
			}
			if envCosmosNet.MessageProcessor.IntervalMS != 0 {
				mergedConfig.CosmosNetworks[i].MessageProcessor.IntervalMS = envCosmosNet.MessageProcessor.IntervalMS
			}
		}
	}

	log.Debug("[CONFIG] Config merged successfully")
	return mergedConfig
}
