package config

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"
	cosmosUtil "github.com/dan13ram/wpokt-oracle/cosmos/util"
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

	if len(config.EthereumNetworks) == 0 {
		return fmt.Errorf("At least one Ethereum network must be configured")
	}

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
		if ethNetwork.TimeoutMS <= 0 {
			return fmt.Errorf("EthereumNetworks[%d].TimeoutMS is required", i)
		}
		if ethNetwork.ChainID <= 0 {
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
		seen := make(map[string]bool)
		for j, oracleAddress := range ethNetwork.OracleAddresses {
			if !IsValidEthereumAddress(oracleAddress) {
				return fmt.Errorf("EthereumNetworks[%d].OracleAddresses[%d] is invalid", i, j)
			}
			if strings.EqualFold(oracleAddress, ethAddress) {
				foundAddress = true
			}
			if seen[oracleAddress] {
				return fmt.Errorf("EthereumNetworks[%d].OracleAddresses[%d] is duplicated", i, j)
			}
			seen[oracleAddress] = true
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
	if len(config.CosmosNetworks) == 0 {
		return fmt.Errorf("At least one Cosmos network must be configured")
	}
	if len(config.CosmosNetworks) > 1 {
		return fmt.Errorf("Only one Cosmos network is supported")
	}

	// cosmos
	for i, cosmosNetwork := range config.CosmosNetworks {
		if cosmosNetwork.StartBlockHeight < 0 {
			return fmt.Errorf("CosmosNetworks[%d].StartBlockHeight is invalid", i)
		}
		if cosmosNetwork.Confirmations < 0 {
			return fmt.Errorf("CosmosNetworks[%d].Confirmations is invalid", i)
		}
		if cosmosNetwork.GRPCEnabled {
			if cosmosNetwork.GRPCHost == "" {
				return fmt.Errorf("CosmosNetworks[%d].GRPCHost is required when GRPCEnabled is true", i)
			}
			if cosmosNetwork.GRPCPort == 0 {
				return fmt.Errorf("CosmosNetworks[%d].GRPCPort is required when GRPCEnabled is true", i)
			}
		} else {
			if cosmosNetwork.RPCURL == "" {
				return fmt.Errorf("CosmosNetworks[%d].RPCURL is required when GRPCEnabled is false", i)
			}
		}
		if cosmosNetwork.TimeoutMS <= 0 {
			return fmt.Errorf("CosmosNetworks[%d].TimeoutMS is required", i)
		}
		if cosmosNetwork.ChainID == "" {
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
		if cosmosNetwork.CoinDenom == "" {
			return fmt.Errorf("CosmosNetworks[%d].CoinDenom is required", i)
		}
		if !IsValidBech32Address(cosmosNetwork.Bech32Prefix, cosmosNetwork.MultisigAddress) {
			return fmt.Errorf("CosmosNetworks[%d].MultisigAddress is invalid", i)
		}
		if cosmosNetwork.MultisigPublicKeys == nil || len(cosmosNetwork.MultisigPublicKeys) <= 1 {
			return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys is required and must have at least 2 public keys", i)
		}
		foundPublicKey := false
		seen := make(map[string]bool)
		var pKeys []crypto.PubKey
		for j, publicKey := range cosmosNetwork.MultisigPublicKeys {
			if !IsValidCosmosPublicKey(publicKey) {
				return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys[%d] is invalid", i, j)
			}
			if strings.EqualFold(publicKey, cosmosPubKey) {
				foundPublicKey = true
			}
			pKey, err := cosmosUtil.PubKeyFromHex(publicKey)
			if err != nil {
				return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys[%d] is invalid", i, j)
			}
			pKeys = append(pKeys, pKey)
			if seen[publicKey] {
				return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys[%d] is duplicated", i, j)
			}
			seen[publicKey] = true
		}
		if !foundPublicKey {
			return fmt.Errorf("CosmosNetworks[%d].MultisigPublicKeys must contain the public key of this oracle", i)
		}
		if cosmosNetwork.MultisigThreshold <= 0 || cosmosNetwork.MultisigThreshold > int64(len(cosmosNetwork.MultisigPublicKeys)) {
			return fmt.Errorf("CosmosNetworks[%d].MultisigThreshold is invalid", i)
		}
		multisigPk := multisig.NewLegacyAminoPubKey(int(cosmosNetwork.MultisigThreshold), pKeys)
		multisigBech32, err := bech32.ConvertAndEncode(cosmosNetwork.Bech32Prefix, multisigPk.Address().Bytes())
		if err != nil {
			log.Fatalf("CosmosNetworks[%d].MultisigAddress could not be converted to bech32: %s", i, err)
		}
		if !strings.EqualFold(cosmosNetwork.MultisigAddress, multisigBech32) {
			return fmt.Errorf("CosmosNetworks[%d].MultisigAddress is not valid for the given public keys and threshold", i)
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
