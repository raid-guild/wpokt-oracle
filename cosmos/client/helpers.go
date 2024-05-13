package client

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"

	testutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"context"
)

var txDecoder sdk.TxDecoder

func init() {

	codecOptions := testutil.CodecOptions{
		AccAddressPrefix: "pokt",
	}

	txConfig := authtx.NewTxConfig(codec.NewProtoCodec(codecOptions.NewInterfaceRegistry()), authtx.DefaultSignModes)

	txDecoder = txConfig.TxDecoder()
}

func getBlocksForTxResults(node *rpchttp.HTTP, resTxs []*rpctypes.ResultTx) (map[int64]*rpctypes.ResultBlock, error) {
	resBlocks := make(map[int64]*rpctypes.ResultBlock)

	for _, resTx := range resTxs {
		resTx := resTx

		if _, ok := resBlocks[resTx.Height]; !ok {
			resBlock, err := node.Block(context.Background(), &resTx.Height)
			if err != nil {
				return nil, err
			}

			resBlocks[resTx.Height] = resBlock
		}
	}

	return resBlocks, nil
}

func formatTxResults(resTxs []*rpctypes.ResultTx, resBlocks map[int64]*rpctypes.ResultBlock) ([]*sdk.TxResponse, error) {
	var err error
	out := make([]*sdk.TxResponse, len(resTxs))
	for i := range resTxs {
		out[i], err = mkTxResult(resTxs[i], resBlocks[resTxs[i].Height])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func mkTxResult(resTx *rpctypes.ResultTx, resBlock *rpctypes.ResultBlock) (*sdk.TxResponse, error) {
	txb, err := txDecoder(resTx.Tx)
	if err != nil {
		return nil, err
	}
	p, ok := txb.(intoAny)
	if !ok {
		return nil, fmt.Errorf("expecting a type implementing intoAny, got: %T", txb)
	}
	any := p.AsAny()
	return sdk.NewResponseResultTx(resTx, any, resBlock.Block.Time.Format(time.RFC3339)), nil
}

// Deprecated: this interface is used only internally for scenario we are
// deprecating (StdTxConfig support)
type intoAny interface {
	AsAny() *codectypes.Any
}
