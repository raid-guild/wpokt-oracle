package client

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"

	testutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	std "github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"context"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func getTxDecoder(bech32Prefix string) sdk.TxDecoder {

	codecOptions := testutil.CodecOptions{
		AccAddressPrefix: bech32Prefix,
	}

	reg := codecOptions.NewInterfaceRegistry()

	std.RegisterInterfaces(reg)
	authtypes.RegisterInterfaces(reg)
	banktypes.RegisterInterfaces(reg)
	// TODO: add more modules' interfaces

	codec := codec.NewProtoCodec(reg)

	return authtx.DefaultTxDecoder(codec)
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

func formatTxResults(bech32Prefix string, resTxs []*rpctypes.ResultTx, resBlocks map[int64]*rpctypes.ResultBlock) ([]*sdk.TxResponse, error) {
	var err error
	out := make([]*sdk.TxResponse, len(resTxs))
	for i := range resTxs {
		out[i], err = mkTxResult(bech32Prefix, resTxs[i], resBlocks[resTxs[i].Height])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func mkTxResult(bech32Prefix string, resTx *rpctypes.ResultTx, resBlock *rpctypes.ResultBlock) (*sdk.TxResponse, error) {
	txDecoder := getTxDecoder(bech32Prefix)
	txb, err := txDecoder(resTx.Tx)
	if err != nil {
		return nil, fmt.Errorf("decoding tx: %w", err)
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
