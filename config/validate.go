package config

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

// ValidateConfig validates the config
func validateConfig(config models.Config) error {
	logger.Debug("Validating config")

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

	logger.Debug("MongoDB validated")

	// Mnemonic for both Ethereum and Cosmos networks
	if config.Mnemonic == "" {
		return fmt.Errorf("Mnemonic is required")
	}
	if !bip39.IsMnemonicValid(config.Mnemonic) {
		return fmt.Errorf("Mnemonic is invalid")
	}

	cosmosPubKey, err := common.CosmosPublicKeyFromMnemonic(config.Mnemonic)
	if err != nil {
		logger.WithError(err).Error("Failed to generate Cosmos public key from mnemonic")
		return fmt.Errorf("failed to generate Cosmos public key from mnemonic")
	}
	cosmosPubKeyHex := hex.EncodeToString(cosmosPubKey.Bytes())
	if !common.IsValidCosmosPublicKey(cosmosPubKeyHex) {
		return fmt.Errorf("cosmos public key is invalid")
	}

	ethAddress, err := common.EthereumAddressFromMnemonic(config.Mnemonic)
	if err != nil {
		logger.WithError(err).Error("Failed to generate Ethereum address from mnemonic")
		return fmt.Errorf("failed to generate Ethereum address from mnemonic")
	}
	if !common.IsValidEthereumAddress(ethAddress) {
		return fmt.Errorf("ethereum address is invalid")
	}

	logger.Debug("Mnemonic validated")

	if len(config.EthereumNetworks) == 0 {
		return fmt.Errorf("at least one ethereum network must be configured")
	}

	// ethereum
	for i, ethNetwork := range config.EthereumNetworks {
		if ethNetwork.StartBlockHeight == 0 {
			logger.Warnf("EthereumNetworks[%d].StartBlockHeight is 0", i)
		}
		if ethNetwork.Confirmations == 0 {
			logger.Warnf("EthereumNetworks[%d].Confirmations is 0", i)
		}
		if ethNetwork.RPCURL == "" {
			return fmt.Errorf("EthereumNetworks[%d].RPCURL is required", i)
		}
		if ethNetwork.TimeoutMS == 0 {
			return fmt.Errorf("EthereumNetworks[%d].TimeoutMS is required", i)
		}
		if ethNetwork.ChainID == 0 {
			return fmt.Errorf("EthereumNetworks[%d].ChainId is required", i)
		}
		if ethNetwork.ChainName == "" {
			return fmt.Errorf("EthereumNetworks[%d].ChainName is required", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.MailboxAddress) {
			return fmt.Errorf("EthereumNetworks[%d].MailboxAddress is invalid", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.MintControllerAddress) {
			return fmt.Errorf("EthereumNetworks[%d].MintControllerAddress is invalid", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.OmniTokenAddress) {
			return fmt.Errorf("EthereumNetworks[%d].OmniTokenAddress is invalid", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.WarpISMAddress) {
			return fmt.Errorf("EthereumNetworks[%d].WarpISMAddress is invalid", i)
		}
		if ethNetwork.OracleAddresses == nil || len(ethNetwork.OracleAddresses) <= 1 {
			return fmt.Errorf("EthereumNetworks[%d].OracleAddresses is required and must have at least 2 addresses", i)
		}
		foundAddress := false
		seen := make(map[string]bool)
		for j, oracleAddress := range ethNetwork.OracleAddresses {
			if !common.IsValidEthereumAddress(oracleAddress) {
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
		if err := validateServiceConfig(fmt.Sprintf("EthereumNetworks[%d].MessageMonitor", i), ethNetwork.MessageMonitor); err != nil {
			return err
		}
		if err := validateServiceConfig(fmt.Sprintf("EthereumNetworks[%d].MessageSigner", i), ethNetwork.MessageSigner); err != nil {
			return err
		}
		if err := validateServiceConfig(fmt.Sprintf("EthereumNetworks[%d].MessageRelayer", i), ethNetwork.MessageRelayer); err != nil {
			return err
		}
	}

	logger.Debug("Ethereum validated")

	// cosmos
	if config.CosmosNetwork.StartBlockHeight == 0 {
		logger.Warnf("CosmosNetwork.StartBlockHeight is 0")
	}
	if config.CosmosNetwork.Confirmations == 0 {
		logger.Warnf("CosmosNetwork.Confirmations is 0")
	}
	if config.CosmosNetwork.GRPCEnabled {
		if config.CosmosNetwork.GRPCHost == "" {
			return fmt.Errorf("CosmosNetwork.GRPCHost is required when GRPCEnabled is true")
		}
		if config.CosmosNetwork.GRPCPort == 0 {
			return fmt.Errorf("CosmosNetwork.GRPCPort is required when GRPCEnabled is true")
		}
	} else {
		if config.CosmosNetwork.RPCURL == "" {
			return fmt.Errorf("CosmosNetwork.RPCURL is required when GRPCEnabled is false")
		}
	}
	if config.CosmosNetwork.TimeoutMS == 0 {
		return fmt.Errorf("CosmosNetwork.TimeoutMS is required")
	}
	if config.CosmosNetwork.ChainID == "" {
		return fmt.Errorf("CosmosNetwork.ChainId is required")
	}
	if config.CosmosNetwork.ChainName == "" {
		return fmt.Errorf("CosmosNetwork.ChainName is required")
	}
	if config.CosmosNetwork.TxFee == "" {
		return fmt.Errorf("CosmosNetwork.TxFee is empty")
	}
	if _, ok := big.NewInt(0).SetString(config.CosmosNetwork.TxFee, 10); !ok {
		return fmt.Errorf("CosmosNetwork.TxFee is invalid")
	}
	if config.CosmosNetwork.Bech32Prefix == "" {
		return fmt.Errorf("CosmosNetwork.Bech32Prefix is required")
	}
	if config.CosmosNetwork.CoinDenom == "" {
		return fmt.Errorf("CosmosNetwork.CoinDenom is required")
	}
	if !common.IsValidBech32Address(config.CosmosNetwork.Bech32Prefix, config.CosmosNetwork.MultisigAddress) {
		return fmt.Errorf("CosmosNetwork.MultisigAddress is invalid")
	}
	if config.CosmosNetwork.MultisigPublicKeys == nil || len(config.CosmosNetwork.MultisigPublicKeys) <= 1 {
		return fmt.Errorf("CosmosNetwork.MultisigPublicKeys is required and must have at least 2 public keys")
	}
	foundPublicKey := false
	seen := make(map[string]bool)
	var pKeys []crypto.PubKey
	for j, publicKey := range config.CosmosNetwork.MultisigPublicKeys {
		publicKey = strings.ToLower(publicKey)
		if !common.IsValidCosmosPublicKey(publicKey) {
			return fmt.Errorf("CosmosNetwork.MultisigPublicKeys[%d] is invalid", j)
		}
		if strings.EqualFold(publicKey, cosmosPubKeyHex) {
			foundPublicKey = true
		}
		pKey, err := common.CosmosPublicKeyFromHex(publicKey)
		if err != nil {
			return fmt.Errorf("CosmosNetwork.MultisigPublicKeys[%d] is invalid", j)
		}
		pKeys = append(pKeys, pKey)
		if seen[publicKey] {
			return fmt.Errorf("CosmosNetwork.MultisigPublicKeys[%d] is duplicated", j)
		}
		seen[publicKey] = true
	}
	if !foundPublicKey {
		return fmt.Errorf("CosmosNetwork.MultisigPublicKeys must contain the public key of this oracle")
	}
	if config.CosmosNetwork.MultisigThreshold == 0 || config.CosmosNetwork.MultisigThreshold > uint64(len(config.CosmosNetwork.MultisigPublicKeys)) {
		return fmt.Errorf("CosmosNetwork.MultisigThreshold is invalid")
	}
	multisigPk := multisig.NewLegacyAminoPubKey(int(config.CosmosNetwork.MultisigThreshold), pKeys)
	multisigBech32, err := bech32.ConvertAndEncode(config.CosmosNetwork.Bech32Prefix, multisigPk.Address().Bytes())
	if err != nil {
		logger.Fatalf("CosmosNetwork.MultisigAddress could not be converted to bech32: %s", err)
	}
	if !strings.EqualFold(config.CosmosNetwork.MultisigAddress, multisigBech32) {
		return fmt.Errorf("CosmosNetwork.MultisigAddress is not valid for the given public keys and threshold")
	}
	if err := validateServiceConfig("CosmosNetwork.MessageMonitor", config.CosmosNetwork.MessageMonitor); err != nil {
		return err
	}
	if err := validateServiceConfig("CosmosNetwork.MessageSigner", config.CosmosNetwork.MessageSigner); err != nil {
		return err
	}
	if err := validateServiceConfig("CosmosNetwork.MessageRelayer", config.CosmosNetwork.MessageRelayer); err != nil {
		return err
	}

	logger.Debug("Cosmos validated")

	if config.HealthCheck.IntervalMS == 0 {
		return fmt.Errorf("HealthCheck.Interval is required")
	}

	logger.Debug("HealthCheck validated")

	logger.Debug("Config validated")
	return nil
}

func validateServiceConfig(label string, config models.ServiceConfig) error {
	if config.Enabled {
		if config.IntervalMS == 0 {
			return fmt.Errorf("%s.IntervalMS is required", label)
		}
	}
	return nil
}
