package util

import (
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
)

func TestParseChain(t *testing.T) {
	config := models.EthereumNetworkConfig{
		ChainName: "Ethereum Mainnet",
		ChainID:   1,
	}

	expectedChain := models.Chain{
		ChainName:   "Ethereum Mainnet",
		ChainID:     "1",
		ChainDomain: uint32(1),
		ChainType:   models.ChainTypeEthereum,
	}

	chain := ParseChain(config)
	assert.Equal(t, expectedChain, chain)
}
