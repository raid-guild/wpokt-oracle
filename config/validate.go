package config

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func validateAndCreateSigner(config models.SignerConfig) (common.Signer, error) {
	logger.Debug("Initializing signer")

	// Mnemonic for both Ethereum and Cosmos networks
	if config.Mnemonic == "" && config.GcpKmsKeyName == "" {
		return nil, fmt.Errorf("Mnemonic or GcpKmsKeyName is required")
	}
	if config.Mnemonic != "" {
		if !bip39.IsMnemonicValid(config.Mnemonic) {
			return nil, fmt.Errorf("Mnemonic is invalid")
		}

		return common.NewMnemonicSigner(config.Mnemonic)
	}

	return common.NewGcpKmsSigner(config.GcpKmsKeyName)

}

// ValidateConfig validates the config
func validateConfig(config models.Config) (common.Signer, error) {
	logger.Debug("Validating config")

	// signer
	signer, err := validateAndCreateSigner(config.Signer)
	if err != nil {
		return nil, fmt.Errorf("Signer: %w", err)
	}

	logger.Debug("Signer validated")

	// ethereum address
	ethAddressHex := signer.EthAddress().Hex()

	// cosmos public key
	cosmosPubKeyHex := hex.EncodeToString(signer.CosmosPublicKey().Bytes())

	// mongodb
	if config.MongoDB.URI == "" {
		return nil, fmt.Errorf("MongoDB.URI is required")
	}
	if config.MongoDB.Database == "" {
		return nil, fmt.Errorf("MongoDB.Database is required")
	}
	if config.MongoDB.TimeoutMS == 0 {
		return nil, fmt.Errorf("MongoDB.TimeoutMS is required")
	}

	logger.Debug("MongoDB validated")

	if len(config.EthereumNetworks) == 0 {
		return nil, fmt.Errorf("at least one ethereum network must be configured")
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
			return nil, fmt.Errorf("EthereumNetworks[%d].RPCURL is required", i)
		}
		if ethNetwork.TimeoutMS == 0 {
			return nil, fmt.Errorf("EthereumNetworks[%d].TimeoutMS is required", i)
		}
		if ethNetwork.ChainID == 0 {
			return nil, fmt.Errorf("EthereumNetworks[%d].ChainId is required", i)
		}
		if ethNetwork.ChainName == "" {
			return nil, fmt.Errorf("EthereumNetworks[%d].ChainName is required", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.MailboxAddress) {
			return nil, fmt.Errorf("EthereumNetworks[%d].MailboxAddress is invalid", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.MintControllerAddress) {
			return nil, fmt.Errorf("EthereumNetworks[%d].MintControllerAddress is invalid", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.OmniTokenAddress) {
			return nil, fmt.Errorf("EthereumNetworks[%d].OmniTokenAddress is invalid", i)
		}
		if !common.IsValidEthereumAddress(ethNetwork.WarpISMAddress) {
			return nil, fmt.Errorf("EthereumNetworks[%d].WarpISMAddress is invalid", i)
		}
		if ethNetwork.OracleAddresses == nil || len(ethNetwork.OracleAddresses) <= 1 {
			return nil, fmt.Errorf("EthereumNetworks[%d].OracleAddresses is required and must have at least 2 addresses", i)
		}
		foundAddress := false
		seen := make(map[string]bool)
		for j, oracleAddress := range ethNetwork.OracleAddresses {
			if !common.IsValidEthereumAddress(oracleAddress) {
				return nil, fmt.Errorf("EthereumNetworks[%d].OracleAddresses[%d] is invalid", i, j)
			}
			if strings.EqualFold(oracleAddress, ethAddressHex) {
				foundAddress = true
			}
			if seen[oracleAddress] {
				return nil, fmt.Errorf("EthereumNetworks[%d].OracleAddresses[%d] is duplicated", i, j)
			}
			seen[oracleAddress] = true
		}
		if !foundAddress {
			return nil, fmt.Errorf("EthereumNetworks[%d].OracleAddresses must contain the address of this oracle", i)
		}
		if err := validateServiceConfig(fmt.Sprintf("EthereumNetworks[%d].MessageMonitor", i), ethNetwork.MessageMonitor); err != nil {
			return nil, err
		}
		if err := validateServiceConfig(fmt.Sprintf("EthereumNetworks[%d].MessageSigner", i), ethNetwork.MessageSigner); err != nil {
			return nil, err
		}
		if err := validateServiceConfig(fmt.Sprintf("EthereumNetworks[%d].MessageRelayer", i), ethNetwork.MessageRelayer); err != nil {
			return nil, err
		}
	}

	logger.Debug("Ethereum validated")

	// cosmos
	if config.CosmosNetwork.StartBlockHeight == 0 {
		logger.Warn("CosmosNetwork.StartBlockHeight is 0")
	}
	if config.CosmosNetwork.Confirmations == 0 {
		logger.Warn("CosmosNetwork.Confirmations is 0")
	}
	if config.CosmosNetwork.GRPCEnabled {
		if config.CosmosNetwork.GRPCHost == "" {
			return nil, fmt.Errorf("CosmosNetwork.GRPCHost is required when GRPCEnabled is true")
		}
		if config.CosmosNetwork.GRPCPort == 0 {
			return nil, fmt.Errorf("CosmosNetwork.GRPCPort is required when GRPCEnabled is true")
		}
	} else {
		if config.CosmosNetwork.RPCURL == "" {
			return nil, fmt.Errorf("CosmosNetwork.RPCURL is required when GRPCEnabled is false")
		}
	}
	if config.CosmosNetwork.TimeoutMS == 0 {
		return nil, fmt.Errorf("CosmosNetwork.TimeoutMS is required")
	}
	if config.CosmosNetwork.ChainID == "" {
		return nil, fmt.Errorf("CosmosNetwork.ChainId is required")
	}
	if config.CosmosNetwork.ChainName == "" {
		return nil, fmt.Errorf("CosmosNetwork.ChainName is required")
	}
	if config.CosmosNetwork.TxFee == 0 {
		logger.Warn("CosmosNetwork.TxFee is 0")
	}
	if config.CosmosNetwork.Bech32Prefix == "" {
		return nil, fmt.Errorf("CosmosNetwork.Bech32Prefix is required")
	}
	if config.CosmosNetwork.CoinDenom == "" {
		return nil, fmt.Errorf("CosmosNetwork.CoinDenom is required")
	}
	if !common.IsValidBech32Address(config.CosmosNetwork.Bech32Prefix, config.CosmosNetwork.MultisigAddress) {
		return nil, fmt.Errorf("CosmosNetwork.MultisigAddress is invalid")
	}
	if config.CosmosNetwork.MultisigPublicKeys == nil || len(config.CosmosNetwork.MultisigPublicKeys) <= 1 {
		return nil, fmt.Errorf("CosmosNetwork.MultisigPublicKeys is required and must have at least 2 public keys")
	}
	foundPublicKey := false
	seen := make(map[string]bool)
	var pKeys []crypto.PubKey
	for j, publicKey := range config.CosmosNetwork.MultisigPublicKeys {
		publicKey = strings.ToLower(publicKey)
		if !common.IsValidCosmosPublicKey(publicKey) {
			return nil, fmt.Errorf("CosmosNetwork.MultisigPublicKeys[%d] is invalid", j)
		}
		if strings.EqualFold(publicKey, cosmosPubKeyHex) {
			foundPublicKey = true
		}
		pKey, _ := common.CosmosPublicKeyFromHex(publicKey) // cannot fail because public key is valid
		pKeys = append(pKeys, pKey)
		if seen[publicKey] {
			return nil, fmt.Errorf("CosmosNetwork.MultisigPublicKeys[%d] is duplicated", j)
		}
		seen[publicKey] = true
	}
	if !foundPublicKey {
		return nil, fmt.Errorf("CosmosNetwork.MultisigPublicKeys must contain the public key of this oracle")
	}
	if config.CosmosNetwork.MultisigThreshold == 0 || config.CosmosNetwork.MultisigThreshold > uint64(len(config.CosmosNetwork.MultisigPublicKeys)) {
		return nil, fmt.Errorf("CosmosNetwork.MultisigThreshold is invalid")
	}
	multisigPk := multisig.NewLegacyAminoPubKey(int(config.CosmosNetwork.MultisigThreshold), pKeys)
	multisigBech32, _ := bech32.ConvertAndEncode(config.CosmosNetwork.Bech32Prefix, multisigPk.Address().Bytes()) // no reason it should fail
	if !strings.EqualFold(config.CosmosNetwork.MultisigAddress, multisigBech32) {
		return nil, fmt.Errorf("CosmosNetwork.MultisigAddress is not valid for the given public keys and threshold")
	}
	if err := validateServiceConfig("CosmosNetwork.MessageMonitor", config.CosmosNetwork.MessageMonitor); err != nil {
		return nil, err
	}
	if err := validateServiceConfig("CosmosNetwork.MessageSigner", config.CosmosNetwork.MessageSigner); err != nil {
		return nil, err
	}
	if err := validateServiceConfig("CosmosNetwork.MessageRelayer", config.CosmosNetwork.MessageRelayer); err != nil {
		return nil, err
	}

	logger.Debug("Cosmos validated")

	if config.HealthCheck.IntervalMS == 0 {
		return nil, fmt.Errorf("HealthCheck.Interval is required")
	}

	logger.Debug("HealthCheck validated")

	logger.Debug("Config validated")
	return signer, nil
}

func validateServiceConfig(label string, config models.ServiceConfig) error {
	if config.Enabled {
		if config.IntervalMS == 0 {
			return fmt.Errorf("%s.IntervalMS is required", label)
		}
	}
	return nil
}
