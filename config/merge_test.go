package config

import (
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
)

func TestMergeConfigs(t *testing.T) {
	t.Run("Merge HealthCheck", func(t *testing.T) {
		yamlConfig := models.Config{}
		envConfig := models.Config{HealthCheck: models.HealthCheckConfig{IntervalMS: 100, ReadLastHealth: true}}

		mergedConfig := mergeConfigs(yamlConfig, envConfig)

		assert.Equal(t, uint64(100), mergedConfig.HealthCheck.IntervalMS)
		assert.True(t, mergedConfig.HealthCheck.ReadLastHealth)
	})

	t.Run("Merge Logger", func(t *testing.T) {
		yamlConfig := models.Config{}
		envConfig := models.Config{Logger: models.LoggerConfig{Level: "info", Format: "json"}}

		mergedConfig := mergeConfigs(yamlConfig, envConfig)

		assert.Equal(t, "info", mergedConfig.Logger.Level)
		assert.Equal(t, "json", mergedConfig.Logger.Format)
	})

	t.Run("Merge MongoDB", func(t *testing.T) {
		yamlConfig := models.Config{}
		envConfig := models.Config{
			MongoDB: models.MongoConfig{
				URI:       "mongodb://localhost:27017",
				Database:  "mydb",
				TimeoutMS: 5000,
			},
		}

		mergedConfig := mergeConfigs(yamlConfig, envConfig)

		assert.Equal(t, "mongodb://localhost:27017", mergedConfig.MongoDB.URI)
		assert.Equal(t, "mydb", mergedConfig.MongoDB.Database)
		assert.Equal(t, uint64(5000), mergedConfig.MongoDB.TimeoutMS)
	})

	t.Run("Merge Mnemonic", func(t *testing.T) {
		yamlConfig := models.Config{}
		envConfig := models.Config{Mnemonic: "my_mnemonic"}

		mergedConfig := mergeConfigs(yamlConfig, envConfig)

		assert.Equal(t, "my_mnemonic", mergedConfig.Mnemonic)
	})

	t.Run("Merge EthereumNetworks", func(t *testing.T) {
		yamlConfig := models.Config{
			EthereumNetworks: []models.EthereumNetworkConfig{
				{ChainID: 1},
			},
		}
		envConfig := models.Config{
			EthereumNetworks: []models.EthereumNetworkConfig{
				{
					StartBlockHeight:      100,
					Confirmations:         12,
					RPCURL:                "http://localhost:8545",
					TimeoutMS:             3000,
					ChainID:               1,
					ChainName:             "Ethereum",
					MailboxAddress:        "0xMailboxAddress",
					MintControllerAddress: "0xMintControllerAddress",
					OmniTokenAddress:      "0xOmniTokenAddress",
					WarpISMAddress:        "0xWarpISMAddress",
					OracleAddresses:       []string{"0xOracle1", "0xOracle2"},
					MessageMonitor: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 1000,
					},
					MessageSigner: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 2000,
					},
					MessageRelayer: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 3000,
					},
				},
				{
					StartBlockHeight:      100,
					Confirmations:         12,
					RPCURL:                "http://localhost:8645",
					TimeoutMS:             3000,
					ChainID:               2,
					ChainName:             "Ethereum",
					MailboxAddress:        "0xMailboxAddress",
					MintControllerAddress: "0xMintControllerAddress",
					OmniTokenAddress:      "0xOmniTokenAddress",
					WarpISMAddress:        "0xWarpISMAddress",
					OracleAddresses:       []string{"0xOracle1", "0xOracle2"},
					MessageMonitor: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 1000,
					},
					MessageSigner: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 2000,
					},
					MessageRelayer: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 3000,
					},
				},
			},
		}

		mergedConfig := mergeConfigs(yamlConfig, envConfig)

		assert.Equal(t, uint64(100), mergedConfig.EthereumNetworks[0].StartBlockHeight)
		assert.Equal(t, uint64(12), mergedConfig.EthereumNetworks[0].Confirmations)
		assert.Equal(t, "http://localhost:8545", mergedConfig.EthereumNetworks[0].RPCURL)
		assert.Equal(t, uint64(3000), mergedConfig.EthereumNetworks[0].TimeoutMS)
		assert.Equal(t, uint64(1), mergedConfig.EthereumNetworks[0].ChainID)
		assert.Equal(t, "Ethereum", mergedConfig.EthereumNetworks[0].ChainName)
		assert.Equal(t, "0xMailboxAddress", mergedConfig.EthereumNetworks[0].MailboxAddress)
		assert.Equal(t, "0xMintControllerAddress", mergedConfig.EthereumNetworks[0].MintControllerAddress)
		assert.Equal(t, "0xOmniTokenAddress", mergedConfig.EthereumNetworks[0].OmniTokenAddress)
		assert.Equal(t, "0xWarpISMAddress", mergedConfig.EthereumNetworks[0].WarpISMAddress)
		assert.Equal(t, []string{"0xOracle1", "0xOracle2"}, mergedConfig.EthereumNetworks[0].OracleAddresses)
		assert.True(t, mergedConfig.EthereumNetworks[0].MessageMonitor.Enabled)
		assert.Equal(t, uint64(1000), mergedConfig.EthereumNetworks[0].MessageMonitor.IntervalMS)
		assert.True(t, mergedConfig.EthereumNetworks[0].MessageSigner.Enabled)
		assert.Equal(t, uint64(2000), mergedConfig.EthereumNetworks[0].MessageSigner.IntervalMS)
		assert.True(t, mergedConfig.EthereumNetworks[0].MessageRelayer.Enabled)
		assert.Equal(t, uint64(3000), mergedConfig.EthereumNetworks[0].MessageRelayer.IntervalMS)
		assert.Equal(t, 2, len(mergedConfig.EthereumNetworks))
		assert.Equal(t, uint64(2), mergedConfig.EthereumNetworks[1].ChainID)
	})

	t.Run("Merge CosmosNetwork", func(t *testing.T) {
		yamlConfig := models.Config{}
		envConfig := models.Config{
			CosmosNetwork: models.CosmosNetworkConfig{
				StartBlockHeight:   100,
				Confirmations:      12,
				RPCURL:             "http://localhost:26657",
				GRPCEnabled:        true,
				GRPCHost:           "localhost",
				GRPCPort:           9090,
				TimeoutMS:          3000,
				ChainID:            "cosmoshub-4",
				ChainName:          "Cosmos Hub",
				TxFee:              5000,
				Bech32Prefix:       "cosmos",
				CoinDenom:          "atom",
				MultisigAddress:    "cosmos1multisigaddress",
				MultisigPublicKeys: []string{"cosmospub1", "cosmospub2"},
				MultisigThreshold:  2,
				MessageMonitor: models.ServiceConfig{
					Enabled:    true,
					IntervalMS: 1000,
				},
				MessageSigner: models.ServiceConfig{
					Enabled:    true,
					IntervalMS: 2000,
				},
				MessageRelayer: models.ServiceConfig{
					Enabled:    true,
					IntervalMS: 3000,
				},
			},
		}

		mergedConfig := mergeConfigs(yamlConfig, envConfig)

		assert.Equal(t, uint64(100), mergedConfig.CosmosNetwork.StartBlockHeight)
		assert.Equal(t, uint64(12), mergedConfig.CosmosNetwork.Confirmations)
		assert.Equal(t, "http://localhost:26657", mergedConfig.CosmosNetwork.RPCURL)
		assert.True(t, mergedConfig.CosmosNetwork.GRPCEnabled)
		assert.Equal(t, "localhost", mergedConfig.CosmosNetwork.GRPCHost)
		assert.Equal(t, uint64(9090), mergedConfig.CosmosNetwork.GRPCPort)
		assert.Equal(t, uint64(3000), mergedConfig.CosmosNetwork.TimeoutMS)
		assert.Equal(t, "cosmoshub-4", mergedConfig.CosmosNetwork.ChainID)
		assert.Equal(t, "Cosmos Hub", mergedConfig.CosmosNetwork.ChainName)
		assert.Equal(t, uint64(5000), mergedConfig.CosmosNetwork.TxFee)
		assert.Equal(t, "cosmos", mergedConfig.CosmosNetwork.Bech32Prefix)
		assert.Equal(t, "atom", mergedConfig.CosmosNetwork.CoinDenom)
		assert.Equal(t, "cosmos1multisigaddress", mergedConfig.CosmosNetwork.MultisigAddress)
		assert.Equal(t, []string{"cosmospub1", "cosmospub2"}, mergedConfig.CosmosNetwork.MultisigPublicKeys)
		assert.Equal(t, uint64(2), mergedConfig.CosmosNetwork.MultisigThreshold)
		assert.True(t, mergedConfig.CosmosNetwork.MessageMonitor.Enabled)
		assert.Equal(t, uint64(1000), mergedConfig.CosmosNetwork.MessageMonitor.IntervalMS)
		assert.True(t, mergedConfig.CosmosNetwork.MessageSigner.Enabled)
		assert.Equal(t, uint64(2000), mergedConfig.CosmosNetwork.MessageSigner.IntervalMS)
		assert.True(t, mergedConfig.CosmosNetwork.MessageRelayer.Enabled)
		assert.Equal(t, uint64(3000), mergedConfig.CosmosNetwork.MessageRelayer.IntervalMS)
	})
}
