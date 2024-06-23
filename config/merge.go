package config

import (
	"github.com/dan13ram/wpokt-oracle/models"
)

// Function to merge two Config structs, prioritizing non-empty configurations from the envConfig
func mergeConfigs(yamlConfig models.Config, envConfig models.Config) models.Config {

	logger.Debug("Merging configs from YAML and ENV")
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
	if envConfig.Logger.Format != "" {
		mergedConfig.Logger.Format = envConfig.Logger.Format
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
			if envEthNet.TimeoutMS != 0 {
				mergedConfig.EthereumNetworks[i].TimeoutMS = envEthNet.TimeoutMS
			}
			if envEthNet.ChainID != 0 {
				mergedConfig.EthereumNetworks[i].ChainID = envEthNet.ChainID
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
			if envEthNet.OmniTokenAddress != "" {
				mergedConfig.EthereumNetworks[i].OmniTokenAddress = envEthNet.OmniTokenAddress
			}
			if envEthNet.WarpISMAddress != "" {
				mergedConfig.EthereumNetworks[i].WarpISMAddress = envEthNet.WarpISMAddress
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
		} else {
			mergedConfig.EthereumNetworks = append(mergedConfig.EthereumNetworks, envEthNet)
		}
	}

	// Merge CosmosNetworks
	if envConfig.CosmosNetwork.StartBlockHeight != 0 {
		mergedConfig.CosmosNetwork.StartBlockHeight = envConfig.CosmosNetwork.StartBlockHeight
	}
	if envConfig.CosmosNetwork.Confirmations != 0 {
		mergedConfig.CosmosNetwork.Confirmations = envConfig.CosmosNetwork.Confirmations
	}
	if envConfig.CosmosNetwork.RPCURL != "" {
		mergedConfig.CosmosNetwork.RPCURL = envConfig.CosmosNetwork.RPCURL
	}
	if envConfig.CosmosNetwork.GRPCEnabled {
		mergedConfig.CosmosNetwork.GRPCEnabled = envConfig.CosmosNetwork.GRPCEnabled
	}
	if envConfig.CosmosNetwork.GRPCHost != "" {
		mergedConfig.CosmosNetwork.GRPCHost = envConfig.CosmosNetwork.GRPCHost
	}
	if envConfig.CosmosNetwork.GRPCPort != 0 {
		mergedConfig.CosmosNetwork.GRPCPort = envConfig.CosmosNetwork.GRPCPort
	}
	if envConfig.CosmosNetwork.TimeoutMS != 0 {
		mergedConfig.CosmosNetwork.TimeoutMS = envConfig.CosmosNetwork.TimeoutMS
	}
	if envConfig.CosmosNetwork.ChainID != "" {
		mergedConfig.CosmosNetwork.ChainID = envConfig.CosmosNetwork.ChainID
	}
	if envConfig.CosmosNetwork.ChainName != "" {
		mergedConfig.CosmosNetwork.ChainName = envConfig.CosmosNetwork.ChainName
	}
	if envConfig.CosmosNetwork.TxFee != 0 {
		mergedConfig.CosmosNetwork.TxFee = envConfig.CosmosNetwork.TxFee
	}
	if envConfig.CosmosNetwork.Bech32Prefix != "" {
		mergedConfig.CosmosNetwork.Bech32Prefix = envConfig.CosmosNetwork.Bech32Prefix
	}
	if envConfig.CosmosNetwork.CoinDenom != "" {
		mergedConfig.CosmosNetwork.CoinDenom = envConfig.CosmosNetwork.CoinDenom
	}
	if envConfig.CosmosNetwork.MultisigAddress != "" {
		mergedConfig.CosmosNetwork.MultisigAddress = envConfig.CosmosNetwork.MultisigAddress
	}
	if len(envConfig.CosmosNetwork.MultisigPublicKeys) != 0 {
		mergedConfig.CosmosNetwork.MultisigPublicKeys = envConfig.CosmosNetwork.MultisigPublicKeys
	}
	if envConfig.CosmosNetwork.MultisigThreshold != 0 {
		mergedConfig.CosmosNetwork.MultisigThreshold = envConfig.CosmosNetwork.MultisigThreshold
	}
	if envConfig.CosmosNetwork.MessageMonitor.Enabled {
		mergedConfig.CosmosNetwork.MessageMonitor.Enabled = envConfig.CosmosNetwork.MessageMonitor.Enabled
	}
	if envConfig.CosmosNetwork.MessageMonitor.IntervalMS != 0 {
		mergedConfig.CosmosNetwork.MessageMonitor.IntervalMS = envConfig.CosmosNetwork.MessageMonitor.IntervalMS
	}
	if envConfig.CosmosNetwork.MessageSigner.Enabled {
		mergedConfig.CosmosNetwork.MessageSigner.Enabled = envConfig.CosmosNetwork.MessageSigner.Enabled
	}
	if envConfig.CosmosNetwork.MessageSigner.IntervalMS != 0 {
		mergedConfig.CosmosNetwork.MessageSigner.IntervalMS = envConfig.CosmosNetwork.MessageSigner.IntervalMS
	}
	if envConfig.CosmosNetwork.MessageRelayer.Enabled {
		mergedConfig.CosmosNetwork.MessageRelayer.Enabled = envConfig.CosmosNetwork.MessageRelayer.Enabled
	}
	if envConfig.CosmosNetwork.MessageRelayer.IntervalMS != 0 {
		mergedConfig.CosmosNetwork.MessageRelayer.IntervalMS = envConfig.CosmosNetwork.MessageRelayer.IntervalMS
	}

	logger.Debug("Config merged successfully")
	return mergedConfig
}
