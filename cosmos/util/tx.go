package util

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/dan13ram/wpokt-oracle/common"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

const (
	SendGasLimit = 200000
)

func NewSendTx(
	bech32Prefix string,
	fromAddr []byte,
	toAddr []byte,
	amountIncludingFees sdk.Coin,
	memo string,
	feeAmount sdk.Coin,
) (string, error) {

	finalAmount := amountIncludingFees.Sub(feeAmount)

	fromAddress, err := common.Bech32FromBytes(bech32Prefix, fromAddr)
	if err != nil {
		return "", fmt.Errorf("error converting from address: %s", err)
	}
	toAddress, err := common.Bech32FromBytes(bech32Prefix, toAddr)
	if err != nil {
		return "", fmt.Errorf("error converting to address: %s", err)
	}

	msg := &banktypes.MsgSend{FromAddress: fromAddress, ToAddress: toAddress, Amount: sdk.NewCoins(finalAmount)}

	txConfig := NewTxConfig(bech32Prefix)

	refundTx := txConfig.NewTxBuilder()

	err = refundTx.SetMsgs(msg)
	if err != nil {
		return "", fmt.Errorf("error setting msg: %s", err)
	}

	refundTx.SetMemo(memo)
	refundTx.SetFeeAmount(sdk.NewCoins(feeAmount))
	refundTx.SetGasLimit(SendGasLimit)

	txEncoder := txConfig.TxJSONEncoder()

	if txEncoder == nil {
		return "", fmt.Errorf("error getting tx encoder")
	}

	txBody, err := txEncoder(refundTx.GetTx())
	if err != nil {
		return "", fmt.Errorf("error encoding tx: %s", err)
	}
	return string(txBody), nil
}

func ParseTxBody(
	bech32Prefix string,
	txBody string,
) (sdk.Tx, error) {
	txConfig := NewTxConfig(bech32Prefix)
	txDecoder := txConfig.TxJSONDecoder()

	if txDecoder == nil {
		return nil, fmt.Errorf("error getting tx decoder")
	}

	tx, err := txDecoder([]byte(txBody))
	if err != nil {
		return nil, fmt.Errorf("error decoding tx: %s", err)
	}

	return tx, nil
}
