package config

import (
	"fmt"
	"strings"

	"github.com/cosmos/go-bip39"
	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

// ValidateConfig validates the config
func ValidateConfig(config models.Config) error {
	log.Debug("[CONFIG] Validating config")

	// mongodb
	if config.MongoDB.URI == "" {
		return fmt.Errorf("MongoDB.URI is required")

	}
	if config.MongoDB.Database == "" {
		return fmt.Errorf("MongoDB.Database is required")
	}
	if config.MongoDB.TimeoutMS == 0 {
		return fmt.Errorf("MongoDB.TimeoutMS is required")
	}

	log.Debug("[CONFIG] MongoDB validated")

	// Mnemonic for both Ethereum and Cosmos networks
	if config.Mnemonic == "" {
		return fmt.Errorf("Mnemonic is required")
	}
	if !bip39.IsMnemonicValid(config.Mnemonic) {
		return fmt.Errorf("Mnemonic is invalid")
	}

	cosmosPubKey, err := CosmosPublicKeyFromMnemonic(config.Mnemonic)
	if err != nil {
		return fmt.Errorf("Failed to generate Cosmos public key from mnemonic: %s", err)
	}
	if !IsValidCosmosPublicKey(cosmosPubKey) {
		return fmt.Errorf("Cosmos public key is invalid")
	}

	ethAddress, err := EthereumAddressFromMnemonic(config.Mnemonic)
	if err != nil {
		return fmt.Errorf("Failed to generate Ethereum address from mnemonic: %s", err)
	}
	if !IsValidEthereumAddress(ethAddress) {
		return fmt.Errorf("Ethereum address is invalid")
	}

	log.Debug("[CONFIG] Mnemonic validated")

	// ethereum
	for i, ethNetwork := range config.EthereumNetworks {
		if ethNetwork.StartBlockHeight < 0 {
			return fmt.Errorf("EthereumNetworks[%d].StartBlockHeight is invalid", i)
		}
		if ethNetwork.Confirmations < 0 {
			return fmt.Errorf("EthereumNetworks[%d].Confirmations is invalid", i)
		}
		if ethNetwork.RPCURL == "" {
			return fmt.Errorf("EthereumNetworks[%d].RPCURL is required", i)
		}
		if ethNetwork.RPCTimeoutMS <= 0 {
			return fmt.Errorf("EthereumNetworks[%d].RPCTimeoutMS is required", i)
		}
		if ethNetwork.ChainId <= 0 {
			return fmt.Errorf("EthereumNetworks[%d].ChainId is required", i)
		}
		if ethNetwork.ChainName == "" {
			return fmt.Errorf("EthereumNetworks[%d].ChainName is required", i)
		}
		if !IsValidEthereumAddress(ethNetwork.MailboxAddress) {
			return fmt.Errorf("EthereumNetworks[%d].MailboxAddress is invalid", i)
		}
		if !IsValidEthereumAddress(ethNetwork.MintControllerAddress) {
			return fmt.Errorf("EthereumNetworks[%d].MintControllerAddress is invalid", i)
		}
		if ethNetwork.OracleAddresses == nil || len(ethNetwork.OracleAddresses) <= 1 {
			return fmt.Errorf("EthereumNetworks[%d].OracleAddresses is required and must have at least 2 addresses", i)
		}
		foundAddress := false
		for j, oracleAddress := range ethNetwork.OracleAddresses {
			if !IsValidEthereumAddress(oracleAddress) {
				return fmt.Errorf("EthereumNetworks[%d].OracleAddresses[%d] is invalid", i, j)
			}
			if strings.EqualFold(oracleAddress, ethAddress) {
				foundAddress = true
			}
		}
		if !foundAddress {
			return fmt.Errorf("EthereumNetworks[%d].OracleAddresses must contain the address of this oracle", i)
		}
		if err := validateServiceConfig("EthereumNetworks[%d].MessageMonitor", ethNetwork.MessageMonitor); err != nil {
			return err
		}
		if err := validateServiceConfig("EthereumNetworks[%d].MessageSigner", ethNetwork.MessageSigner); err != nil {
			return err
		}
		if err := validateServiceConfig("EthereumNetworks[%d].MessageRelayer", ethNetwork.MessageRelayer); err != nil {
			return err
		}
	}

	log.Debug("[CONFIG] Ethereum validated")

	// cosmos
	for i, cosmosNetwork := range config.CosmosNetworks {
		if cosmosNetwork.StartBlockHeight < 0 {
			return fmt.Errorf("CosmosNetworks[%d].StartBlockHeight is invalid", i)
		}
		if cosmosNetwork.Confirmations < 0 {
			return fmt.Errorf("CosmosNetworks[%d].Confirmations is invalid", i)
		}
		if cosmosNetwork.RPCURL == "" {
			return fmt.Errorf("CosmosNetworks[%d].RPCURL is required", i)
		}
		if cosmosNetwork.RPCTimeoutMS <= 0 {
			return fmt.Errorf("CosmosNetworks[%d].RPCTimeoutMS is required", i)
		}
		if cosmosNetwork.ChainId == "" {
			return fmt.Errorf("CosmosNetworks[%d].ChainId is required", i)
		}
		if cosmosNetwork.ChainName == "" {
			return fmt.Errorf("CosmosNetworks[%d].ChainName is required", i)
		}
		if cosmosNetwork.TxFee < 0 {
			return fmt.Errorf("CosmosNetworks[%d].TxFee is invalid", i)
		}
		if cosmosNetwork.Bech32Prefix == "" {
			return fmt.Errorf("CosmosNetworks[%d].Bech32Prefix is required", i)
		}
		if !IsValidBech32Address(cosmosNetwork.Bech32Prefix, cosmosNetwork.MultisigAddress) {
			return fmt.Errorf("CosmosNetworks[%d].MultisigAddress is invalid", i)
		}
		if cosmosNetwork.MultisigPublicKeys == nil || len(cosmosNetwork.MultisigPublicKeys) <= 1 {
			return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys is required and must have at least 2 public keys", i)
		}
		foundPublicKey := false
		for j, publicKey := range cosmosNetwork.MultisigPublicKeys {
			if !IsValidCosmosPublicKey(publicKey) {
				return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys[%d] is invalid", i, j)
			}
			if strings.EqualFold(publicKey, cosmosPubKey) {
				foundPublicKey = true
			}
		}
		if !foundPublicKey {
			return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys must contain the public key of this oracle", i)
		}
		if cosmosNetwork.MultisigThreshold <= 0 || cosmosNetwork.MultisigThreshold > int64(len(cosmosNetwork.MultisigPublicKeys)) {
			return fmt.Errorf("CosmosNetworks[%d].MultisigThreshold is invalid", i)
		}
		if err := validateServiceConfig("CosmosNetworks[%d].MessageMonitor", cosmosNetwork.MessageMonitor); err != nil {
			return err
		}
		if err := validateServiceConfig("CosmosNetworks[%d].MessageSigner", cosmosNetwork.MessageSigner); err != nil {
			return err
		}
		if err := validateServiceConfig("CosmosNetworks[%d].MessageRelayer", cosmosNetwork.MessageRelayer); err != nil {
			return err
		}
	}

	log.Debug("[CONFIG] Cosmos validated")

	if config.HealthCheck.IntervalMS == 0 {
		return fmt.Errorf("HealthCheck.Interval is required")
	}

	log.Debug("[CONFIG] HealthCheck validated")

	log.Debug("[CONFIG] config validated")
	return nil
}

func validateServiceConfig(label string, config models.ServiceConfig) error {
	if config.Enabled {
		if config.IntervalMS <= 0 {
			return fmt.Errorf("%s.IntervalMS is required", label)
		}
	}
	return nil
}
