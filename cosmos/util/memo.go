package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/common/math"
)

func ValidateMemo(txMemo string, supportedChainIDsEthereum map[uint32]bool) (models.MintMemo, error) {
	var memo models.MintMemo

	err := json.Unmarshal([]byte(txMemo), &memo)
	if err != nil {
		return memo, fmt.Errorf("failed to unmarshal memo: %s", err)
	}

	memo.Address = strings.Trim(strings.ToLower(memo.Address), " ")
	memo.ChainID = strings.Trim(strings.ToLower(memo.ChainID), " ")

	if !common.IsValidEthereumAddress(memo.Address) {
		return memo, fmt.Errorf("invalid address: %s", memo.Address)
	}

	if strings.EqualFold(memo.Address, common.ZeroAddress) {
		return memo, fmt.Errorf("zero address: %s", memo.Address)
	}

	chainID, ok := math.ParseUint64(memo.ChainID)
	if !ok {
		return memo, fmt.Errorf("invalid chain id: %s", memo.ChainID)
	}

	if _, ok := supportedChainIDsEthereum[uint32(chainID)]; !ok {
		return memo, fmt.Errorf("unsupported chain id: %s", memo.ChainID)
	}

	return memo, nil
}
