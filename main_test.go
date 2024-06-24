package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"

	cosmosUtil "github.com/dan13ram/wpokt-oracle/cosmos/util"

	log "github.com/sirupsen/logrus"
)

func TestNewMintControllerMap_ValidConfig(t *testing.T) {
	config := models.Config{
		EthereumNetworks: []models.EthereumNetworkConfig{
			{
				ChainID:               1,
				MintControllerAddress: "0x1234567890abcdef1234567890abcdef12345678",
			},
			{
				ChainID:               2,
				MintControllerAddress: "0xabcdef1234567890abcdef1234567890abcdef12",
			},
		},
		CosmosNetwork: models.CosmosNetworkConfig{
			ChainID:         "poktroll",
			Bech32Prefix:    "pokt",
			MultisigAddress: "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		},
	}

	cosmosChainDomain := cosmosUtil.ParseChain(config.CosmosNetwork).ChainDomain

	mintControllerMap := NewMintControllerMap(config)
	require.NotNil(t, mintControllerMap)

	expectedEthController1, _ := common.BytesFromAddressHex("0x1234567890abcdef1234567890abcdef12345678")
	expectedEthController2, _ := common.BytesFromAddressHex("0xabcdef1234567890abcdef1234567890abcdef12")
	expectedCosmosController, _ := common.AddressBytesFromBech32("pokt", "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar")

	assert.Equal(t, expectedEthController1, mintControllerMap[1])
	assert.Equal(t, expectedEthController2, mintControllerMap[2])
	assert.Equal(t, expectedCosmosController, mintControllerMap[cosmosChainDomain])
}

func TestNewMintControllerMap_InvalidMintControllerAddress(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	config := models.Config{
		EthereumNetworks: []models.EthereumNetworkConfig{
			{
				ChainID:               1,
				MintControllerAddress: "invalid-address",
			},
		},
		CosmosNetwork: models.CosmosNetworkConfig{
			ChainID:         "poktroll",
			Bech32Prefix:    "cosmos",
			MultisigAddress: "cosmos1c9p58y5sh7fnz0uh23znrl5x98y7m6jgpw9keq",
		},
	}

	assert.Panics(t, func() { NewMintControllerMap(config) })
}

func TestNewMintControllerMap_InvalidCosmosMintControllerAddress(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	config := models.Config{
		EthereumNetworks: []models.EthereumNetworkConfig{
			{
				ChainID:               1,
				MintControllerAddress: "0x1234567890abcdef1234567890abcdef12345678",
			},
		},
		CosmosNetwork: models.CosmosNetworkConfig{
			ChainID:         "poktroll",
			Bech32Prefix:    "cosmos",
			MultisigAddress: "invalid-address",
		},
	}

	assert.Panics(t, func() { NewMintControllerMap(config) })
}

func TestNewMintControllerMap_EmptyConfig(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	config := models.Config{}

	assert.Panics(t, func() { NewMintControllerMap(config) })
}
