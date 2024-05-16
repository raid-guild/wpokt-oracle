package util

import (
	"fmt"

	"github.com/dan13ram/wpokt-oracle/models"
)

func ParseChain(config models.EthereumNetworkConfig) models.Chain {
	return models.Chain{
		ChainName:   config.ChainName,
		ChainID:     fmt.Sprintf("%d", config.ChainID),
		ChainDomain: config.ChainID,
		ChainType:   models.ChainTypeEthereum,
	}
}
