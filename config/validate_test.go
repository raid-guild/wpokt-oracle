package config

import (
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		config := models.Config{
			MongoDB: models.MongoConfig{
				URI:       "mongodb://localhost:27017",
				Database:  "testdb",
				TimeoutMS: 1000,
			},
			Mnemonic: "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve",
			EthereumNetworks: []models.EthereumNetworkConfig{
				{
					StartBlockHeight:      1,
					Confirmations:         1,
					RPCURL:                "http://localhost:8545",
					TimeoutMS:             1000,
					ChainID:               1,
					ChainName:             "Ethereum",
					MailboxAddress:        "0x0000000000000000000000000000000000000000",
					MintControllerAddress: "0x0000000000000000000000000000000000000000",
					OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
					WarpISMAddress:        "0x0000000000000000000000000000000000000000",
					OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
					MessageMonitor: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 1000,
					},
					MessageSigner: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 1000,
					},
					MessageRelayer: models.ServiceConfig{
						Enabled:    true,
						IntervalMS: 1000,
					},
				},
			},
			CosmosNetwork: models.CosmosNetworkConfig{
				StartBlockHeight:   1,
				Confirmations:      1,
				RPCURL:             "http://localhost:36657",
				GRPCEnabled:        true,
				GRPCHost:           "localhost",
				GRPCPort:           9090,
				TimeoutMS:          1000,
				ChainID:            "poktroll",
				ChainName:          "Poktroll",
				TxFee:              1000,
				Bech32Prefix:       "pokt",
				CoinDenom:          "upokt",
				MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
				MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
				MultisigThreshold:  2,
				MessageMonitor: models.ServiceConfig{
					Enabled:    true,
					IntervalMS: 1000,
				},
				MessageSigner: models.ServiceConfig{
					Enabled:    true,
					IntervalMS: 1000,
				},
				MessageRelayer: models.ServiceConfig{
					Enabled:    true,
					IntervalMS: 1000,
				},
			},
			HealthCheck: models.HealthCheckConfig{
				IntervalMS:     1000,
				ReadLastHealth: true,
			},
		}
		err := validateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Invalid config", func(t *testing.T) {
		config := models.Config{}
		err := validateConfig(config)
		assert.Error(t, err)
	})
}

func TestValidateServiceConfig(t *testing.T) {
	t.Run("Valid service config", func(t *testing.T) {
		config := models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		}
		err := validateServiceConfig("TestService", config)
		assert.NoError(t, err)
	})

	t.Run("Invalid service config", func(t *testing.T) {
		config := models.ServiceConfig{
			Enabled: true,
		}
		err := validateServiceConfig("TestService", config)
		assert.Error(t, err)
	})
}
