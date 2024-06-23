package config

import (
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus/hooks/test"

	log "github.com/sirupsen/logrus"
)

func validConfig() models.Config {
	return models.Config{
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
}

func TestValidateConfig(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		config := validConfig()
		err := validateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Invalid mongodb uri", func(t *testing.T) {
		config := validConfig()
		config.MongoDB.URI = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MongoDB.URI")
	})

	t.Run("Invalid mongodb database", func(t *testing.T) {
		config := validConfig()
		config.MongoDB.Database = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MongoDB.Database")
	})

	t.Run("Invalid mongodb timeout", func(t *testing.T) {
		config := validConfig()
		config.MongoDB.TimeoutMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MongoDB.TimeoutMS")
	})

	t.Run("Invalid mnemonic", func(t *testing.T) {
		config := validConfig()
		config.Mnemonic = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mnemonic is required")

		config.Mnemonic = "invalid mnemonic"
		err = validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mnemonic is invalid")
	})

	t.Run("Invalid ethereum networks length", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks = []models.EthereumNetworkConfig{}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one ethereum network must be configured")
	})

	t.Run("Invalid ethereum network rpc url", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].RPCURL = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].RPCURL")
	})

	t.Run("Invalid ethereum network start block height", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].StartBlockHeight = 0
		config.EthereumNetworks[0].RPCURL = ""

		testLogger, hook := test.NewNullLogger()
		oldLogger := logger
		logger = testLogger.WithField("test", "validateConfig")
		defer func() {
			logger = oldLogger
		}()
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].RPCURL")
		assert.Contains(t, hook.LastEntry().Message, "EthereumNetworks[0].StartBlockHeight")
		assert.Equal(t, hook.LastEntry().Level, log.WarnLevel)
	})

	t.Run("Invalid ethereum network confirmations", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].Confirmations = 0
		config.EthereumNetworks[0].RPCURL = ""

		testLogger, hook := test.NewNullLogger()
		oldLogger := logger
		logger = testLogger.WithField("test", "validateConfig")
		defer func() {
			logger = oldLogger
		}()
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].RPCURL")
		assert.Contains(t, hook.LastEntry().Message, "EthereumNetworks[0].Confirmations")
		assert.Equal(t, hook.LastEntry().Level, log.WarnLevel)
	})

	t.Run("Invalid ethereum network timeout", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].TimeoutMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].TimeoutMS")
	})

	t.Run("Invalid ethereum network chain id", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].ChainID = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].ChainId")
	})

	t.Run("Invalid ethereum network chain name", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].ChainName = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].ChainName")
	})

	t.Run("Invalid ethereum network mailbox address", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].MailboxAddress = "invalid address"
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].MailboxAddress")
	})

	t.Run("Invalid ethereum network mint controller address", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].MintControllerAddress = "invalid address"
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].MintControllerAddress")
	})

	t.Run("Invalid ethereum network omni token address", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].OmniTokenAddress = "invalid address"
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].OmniTokenAddress")
	})

	t.Run("Invalid ethereum network warp ism address", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].WarpISMAddress = "invalid address"
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].WarpISMAddress")
	})

	t.Run("Invalid ethereum network oracle addresses", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].OracleAddresses = []string{}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].OracleAddresses")
	})

	t.Run("Invalid ethereum network oracle addresses with invalid address", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].OracleAddresses = []string{"0xinvalid", "0xinvalid"}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].OracleAddresses[0] is invalid")
	})

	t.Run("Invalid ethereum network oracle addresses with duplicated address", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].OracleAddresses = []string{
			"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb",
			"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb",
		}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].OracleAddresses[1] is duplicated")
	})

	t.Run("Invalid ethereum network oracle addresses without oracle address", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].OracleAddresses = []string{
			"0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6902",
			"0x0E90A32Df6f6143F1A91c25d9552dCbc789C3401",
		}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].OracleAddresses")
	})

	t.Run("Invalid ethereum network message monitor", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].MessageMonitor.IntervalMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].MessageMonitor")
	})

	t.Run("Invalid ethereum network message signer", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].MessageSigner.IntervalMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].MessageSigner")
	})

	t.Run("Invalid ethereum network message relayer", func(t *testing.T) {
		config := validConfig()
		config.EthereumNetworks[0].MessageRelayer.IntervalMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EthereumNetworks[0].MessageRelayer")
	})

	t.Run("Invalid cosmos network rpc url", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.GRPCEnabled = false
		config.CosmosNetwork.RPCURL = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.RPCURL")
	})

	t.Run("Invalid cosmos network start block height", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.StartBlockHeight = 0
		config.CosmosNetwork.GRPCEnabled = false
		config.CosmosNetwork.RPCURL = ""

		testLogger, hook := test.NewNullLogger()
		oldLogger := logger
		logger = testLogger.WithField("test", "validateConfig")
		defer func() {
			logger = oldLogger
		}()

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.RPCURL")
		assert.Contains(t, hook.LastEntry().Message, "CosmosNetwork.StartBlockHeight")
		assert.Equal(t, hook.LastEntry().Level, log.WarnLevel)
	})

	t.Run("Invalid cosmos network confirmations", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.Confirmations = 0
		config.CosmosNetwork.GRPCEnabled = false
		config.CosmosNetwork.RPCURL = ""

		testLogger, hook := test.NewNullLogger()
		oldLogger := logger
		logger = testLogger.WithField("test", "validateConfig")
		defer func() {
			logger = oldLogger
		}()

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.RPCURL")
		assert.Contains(t, hook.LastEntry().Message, "CosmosNetwork.Confirmations")
		assert.Equal(t, hook.LastEntry().Level, log.WarnLevel)
	})

	t.Run("Invalid cosmos network grpc host", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.GRPCEnabled = true
		config.CosmosNetwork.GRPCHost = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.GRPCHost")
	})

	t.Run("Invalid cosmos network grpc port", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.GRPCEnabled = true
		config.CosmosNetwork.GRPCPort = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.GRPCPort")
	})

	t.Run("Invalid cosmos network timeout", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.TimeoutMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.TimeoutMS")
	})

	t.Run("Invalid cosmos network chain id", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.ChainID = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.ChainId")
	})

	t.Run("Invalid cosmos network chain name", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.ChainName = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.ChainName")
	})

	t.Run("Invalid cosmos network bech32 prefix", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.Bech32Prefix = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.Bech32Prefix")
	})

	t.Run("Invalid cosmos network tx fee", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.TxFee = 0
		config.CosmosNetwork.Bech32Prefix = ""

		testLogger, hook := test.NewNullLogger()
		oldLogger := logger
		logger = testLogger.WithField("test", "validateConfig")
		defer func() {
			logger = oldLogger
		}()

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.Bech32Prefix")
		assert.Contains(t, hook.LastEntry().Message, "CosmosNetwork.TxFee")
		assert.Equal(t, hook.LastEntry().Level, log.WarnLevel)
	})

	t.Run("Invalid cosmos network coin denom", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.CoinDenom = ""
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.CoinDenom")
	})

	t.Run("Invalid cosmos network multisig address", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MultisigAddress = "invalid address"
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MultisigAddress")
	})

	t.Run("Incorrect cosmos network multisig address", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MultisigAddress = "pokt1mrqt5f7qh8uxs27cjm9t7v9e74a9vvdnq5jva4"
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MultisigAddress")
	})

	t.Run("Invalid cosmos network multisig public keys", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MultisigPublicKeys = []string{}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MultisigPublicKeys")
	})

	t.Run("Invalid cosmos network multisig public keys with invalid address", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MultisigPublicKeys = []string{"invalid", "invalid"}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MultisigPublicKeys[0] is invalid")
	})

	t.Run("Invalid cosmos network multisig public keys with duplicated address", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MultisigPublicKeys = []string{
			"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9",
			"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9",
		}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MultisigPublicKeys[1] is duplicated")
	})

	t.Run("Invalid cosmos network multisig public keys without public key", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MultisigPublicKeys = []string{
			"02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6d9",
			"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d8",
		}
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MultisigPublicKeys")
	})

	t.Run("Invalid cosmos network multisig threshold", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MultisigThreshold = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MultisigThreshold")
	})

	t.Run("Invalid cosmos network message monitor", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MessageMonitor.IntervalMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MessageMonitor")
	})

	t.Run("Invalid cosmos network message signer", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MessageSigner.IntervalMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MessageSigner")
	})

	t.Run("Invalid cosmos network message relayer", func(t *testing.T) {
		config := validConfig()
		config.CosmosNetwork.MessageRelayer.IntervalMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CosmosNetwork.MessageRelayer")
	})

	t.Run("Invalid health check interval", func(t *testing.T) {
		config := validConfig()
		config.HealthCheck.IntervalMS = 0
		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HealthCheck.Interval")
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
