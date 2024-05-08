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

	if envConfig.Mnemonic != "" {
		mergedConfig.Mnemonic = envConfig.Mnemonic
	}

	// Merge EthereumNetworks
	for i, envEthNet := range envConfig.EthereumNetworks {
		if i < len(mergedConfig.EthereumNetworks) {
			if envEthNet.StartBlockHeight != 0 {
				mergedConfig.EthereumNetworks[i].StartBlockHeight = envEthNet.StartBlockHeight
			}
			if envEthNet.Confirmations != 0 {
				mergedConfig.EthereumNetworks[i].Confirmations = envEthNet.Confirmations
			}
			if envEthNet.RPCURL != "" {
				mergedConfig.EthereumNetworks[i].RPCURL = envEthNet.RPCURL
			}
			if envEthNet.RPCTimeoutMS != 0 {
				mergedConfig.EthereumNetworks[i].RPCTimeoutMS = envEthNet.RPCTimeoutMS
			}
			if envEthNet.ChainId != 0 {
				mergedConfig.EthereumNetworks[i].ChainId = envEthNet.ChainId
			}
			if envEthNet.ChainName != "" {
				mergedConfig.EthereumNetworks[i].ChainName = envEthNet.ChainName
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
			if envEthNet.MessageRelayer.Enabled {
				mergedConfig.EthereumNetworks[i].MessageRelayer.Enabled = envEthNet.MessageRelayer.Enabled
			}
			if envEthNet.MessageRelayer.IntervalMS != 0 {
				mergedConfig.EthereumNetworks[i].MessageRelayer.IntervalMS = envEthNet.MessageRelayer.IntervalMS
			}
		}
	}

	// Merge CosmosNetworks
	for i, envCosmosNet := range envConfig.CosmosNetworks {
		if i < len(mergedConfig.CosmosNetworks) {
			if envCosmosNet.StartBlockHeight != 0 {
				mergedConfig.CosmosNetworks[i].StartBlockHeight = envCosmosNet.StartBlockHeight
			}
			if envCosmosNet.Confirmations != 0 {
				mergedConfig.CosmosNetworks[i].Confirmations = envCosmosNet.Confirmations
			}
			if envCosmosNet.GRPCHost != "" {
				mergedConfig.CosmosNetworks[i].GRPCHost = envCosmosNet.GRPCHost
			}
			if envCosmosNet.GRPCPort != 0 {
				mergedConfig.CosmosNetworks[i].GRPCPort = envCosmosNet.GRPCPort
			}
			if envCosmosNet.GRPCTimeoutMS != 0 {
				mergedConfig.CosmosNetworks[i].GRPCTimeoutMS = envCosmosNet.GRPCTimeoutMS
			}
			if envCosmosNet.ChainId != "" {
				mergedConfig.CosmosNetworks[i].ChainId = envCosmosNet.ChainId
			}
			if envCosmosNet.ChainName != "" {
				mergedConfig.CosmosNetworks[i].ChainName = envCosmosNet.ChainName
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
			if envCosmosNet.MessageRelayer.Enabled {
				mergedConfig.CosmosNetworks[i].MessageRelayer.Enabled = envCosmosNet.MessageRelayer.Enabled
			}
			if envCosmosNet.MessageRelayer.IntervalMS != 0 {
				mergedConfig.CosmosNetworks[i].MessageRelayer.IntervalMS = envCosmosNet.MessageRelayer.IntervalMS
			}
		}
	}

	log.Debug("[CONFIG] Config merged successfully")
	return mergedConfig
}
