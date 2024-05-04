package config

import (
	"fmt"
	"strings"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

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

	// ethereum
	for i, ethNetwork := range config.EthereumNetworks {
		if ethNetwork.StartBlockNumber <= 0 {
			return fmt.Errorf("EthereumNetworks[%d].StartBlockNumber is required", i)
		}
		if ethNetwork.Confirmations < 0 {
			return fmt.Errorf("EthereumNetworks[%d].Confirmations is required", i)
		}
		if ethNetwork.PrivateKey == "" {
			return fmt.Errorf("EthereumNetworks[%d].PrivateKey is required", i)
		}
		if has0xPrefix(ethNetwork.PrivateKey) {
			ethNetwork.PrivateKey = ethNetwork.PrivateKey[2:]
		}
		if ethNetwork.RPCURL == "" {
			return fmt.Errorf("EthereumNetworks[%d].RPCURL is required", i)
		}
		if ethNetwork.RPCTimeoutMS <= 0 {
			return fmt.Errorf("EthereumNetworks[%d].RPCTimeoutMS is required", i)
		}
		if ethNetwork.ChainId == "" {
			return fmt.Errorf("EthereumNetworks[%d].ChainId is required", i)
		}
		if !has0xPrefix(ethNetwork.MailboxAddress) || !common.IsHexAddress(ethNetwork.MailboxAddress) {
			return fmt.Errorf("EthereumNetworks[%d].MailboxAddress is invalid", i)
		}
		if !has0xPrefix(ethNetwork.MintControllerAddress) || !common.IsHexAddress(ethNetwork.MintControllerAddress) {
			return fmt.Errorf("EthereumNetworks[%d].MintControllerAddress is invalid", i)
		}
		if ethNetwork.OracleAddresses == nil || len(ethNetwork.OracleAddresses) == 0 {
			return fmt.Errorf("EthereumNetworks[%d].OracleAddresses is required", i)
		}
		for j, oracleAddress := range ethNetwork.OracleAddresses {
			if !has0xPrefix(oracleAddress) ||
				!common.IsHexAddress(oracleAddress) {
				return fmt.Errorf("EthereumNetworks[%d].OracleAddresses[%d] is invalid", i, j)
			}
		}
		if err := validateServiceConfig("EthereumNetworks[%d].MessageMonitor", ethNetwork.MessageMonitor); err != nil {
			return err
		}
		if err := validateServiceConfig("EthereumNetworks[%d].MessageSigner", ethNetwork.MessageSigner); err != nil {
			return err
		}
		if err := validateServiceConfig("EthereumNetworks[%d].MessageProcessor", ethNetwork.MessageProcessor); err != nil {
			return err
		}
	}

	// cosmos
	for i, cosmosNetwork := range config.CosmosNetworks {
		if cosmosNetwork.StartBlockHeight <= 0 {
			return fmt.Errorf("CosmosNetworks[%d].StartBlockHeight is required", i)
		}
		if cosmosNetwork.Confirmations < 0 {
			return fmt.Errorf("CosmosNetworks[%d].Confirmations is required", i)
		}
		if cosmosNetwork.PrivateKey == "" {
			return fmt.Errorf("CosmosNetworks[%d].PrivateKey is required", i)
		}
		if strings.HasPrefix(cosmosNetwork.PrivateKey, "0x") {
			cosmosNetwork.PrivateKey = cosmosNetwork.PrivateKey[2:]
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
		if cosmosNetwork.TxFee < 0 {
			return fmt.Errorf("CosmosNetworks[%d].TxFee is required", i)
		}
		if !common.IsHexAddress(cosmosNetwork.MailboxAddress) {
			return fmt.Errorf("CosmosNetworks[%d].MailboxAddress is invalid", i)
		}
		if cosmosNetwork.MintControllerAddress == "" {
			return fmt.Errorf("CosmosNetworks[%d].MintControllerAddress is required", i)
		}
		if !common.IsHexAddress(cosmosNetwork.MintControllerAddress) {
			return fmt.Errorf("CosmosNetworks[%d].MintControllerAddress is invalid", i)
		}
		if cosmosNetwork.OracleAddresses == nil || len(cosmosNetwork.OracleAddresses) == 0 {
			return fmt.Errorf("CosmosNetworks[%d].OracleAddresses is required", i)
		}
		for j, oracleAddress := range cosmosNetwork.OracleAddresses {
			if !common.IsHexAddress(oracleAddress) {
				return fmt.Errorf("CosmosNetworks[%d].OracleAddresses[%d] is invalid", i, j)
			}
		}
		if err := validateServiceConfig("CosmosNetworks[%d].MessageMonitor", cosmosNetwork.MessageMonitor); err != nil {
			return err
		}
		if err := validateServiceConfig("CosmosNetworks[%d].MessageSigner", cosmosNetwork.MessageSigner); err != nil {
			return err
		}
		if err := validateServiceConfig("CosmosNetworks[%d].MessageProcessor", cosmosNetwork.MessageProcessor); err != nil {
			return err
		}

	}

	if config.HealthCheck.IntervalMS == 0 {
		return fmt.Errorf("HealthCheck.Interval is required")
	}

	log.Debug("[CONFIG] config validated")
}

func validateServiceConfig(label string, config models.ServiceConfig) error {
	if config.Enabled {
		if config.IntervalMS <= 0 {
			return fmt.Errorf("%s.IntervalMS is required", label)
		}
	}
	return nil
}
