package util

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/std"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
)

func NewTxConfig(bech32Prefix string) client.TxConfig {

	codecOptions := testutil.CodecOptions{
		AccAddressPrefix: bech32Prefix,
	}

	reg := codecOptions.NewInterfaceRegistry()

	std.RegisterInterfaces(reg)
	authtypes.RegisterInterfaces(reg)
	banktypes.RegisterInterfaces(reg)
	// TODO: add more modules' interfaces

	codec := codec.NewProtoCodec(reg)

	txConfig := authtx.NewTxConfig(codec, authtx.DefaultSignModes)

	return txConfig
}
