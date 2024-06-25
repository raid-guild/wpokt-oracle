package cosmos

import (
	"bytes"
	"fmt"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/dan13ram/wpokt-oracle/common"
	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/models"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"

	"context"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var ErrAlreadySigned = fmt.Errorf("already signed")

var CosmosSignTx = cosmosSignTx

func cosmosSignTx(
	signerKey crypto.PrivKey,
	config models.CosmosNetworkConfig,
	client cosmos.CosmosClient,
	sequence uint64,
	signatures []models.Signature,
	transactionBody string,
	toAddress []byte,
	amount sdk.Coin,
	memo string,
) (string, []models.Signature, error) {

	for _, sig := range signatures {
		signer, err := common.BytesFromAddressHex(sig.Signer)
		if err != nil {
			return "", nil, fmt.Errorf("error parsing signer: %w", err)
		}
		if bytes.Equal(signer, signerKey.PubKey().Address().Bytes()) {
			return "", nil, ErrAlreadySigned
		}
	}

	multisigAddressBytes, err := common.AddressBytesFromBech32(config.Bech32Prefix, config.MultisigAddress)
	if err != nil {
		return "", nil, fmt.Errorf("error parsing multisig address: %w", err)
	}

	if transactionBody == "" {
		txBody, err := utilNewSendTx(
			config.Bech32Prefix,
			multisigAddressBytes,
			toAddress,
			amount,
			memo,
			sdk.NewCoin(config.CoinDenom, math.NewIntFromUint64(config.TxFee)),
		)
		if err != nil {
			return "", nil, fmt.Errorf("error creating tx body: %w", err)
		}

		transactionBody = txBody
	}

	txBuilder, txConfig, err := utilWrapTxBuilder(config.Bech32Prefix, transactionBody)
	if err != nil {
		return "", nil, fmt.Errorf("error wrapping tx builder: %w", err)
	}

	// check whether the address is a signer
	signers, err := txBuilder.GetTx().GetSigners()
	if err != nil {
		return "", nil, fmt.Errorf("error getting signers: %w", err)
	}

	if !isTxSigner(multisigAddressBytes, signers) {
		return "", nil, fmt.Errorf("address is not a signer")
	}

	account, err := client.GetAccount(config.MultisigAddress)

	if err != nil {
		return "", nil, fmt.Errorf("error getting account: %w", err)
	}

	pubKey := signerKey.PubKey()

	signerData := authsigning.SignerData{
		ChainID:       config.ChainID,
		AccountNumber: account.AccountNumber,
		Sequence:      sequence,
		PubKey:        pubKey,
		Address:       sdk.AccAddress(pubKey.Address()).String(),
	}

	sigV2, _, err := utilSignWithPrivKey(
		context.Background(),
		signerData,
		txBuilder,
		signerKey,
		txConfig,
		sequence,
	)
	if err != nil {
		return "", nil, fmt.Errorf("error signing: %w", err)
	}

	var sigV2s []signingtypes.SignatureV2

	if len(signatures) > 0 {
		sigV2s, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			return "", nil, fmt.Errorf("error getting signatures: %w", err)
		}
	}

	sigV2s = append(sigV2s, sigV2)
	err = txBuilder.SetSignatures(sigV2s...)

	if err != nil {
		return "", nil, fmt.Errorf("error setting signatures: %w", err)
	}

	txBody, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", nil, fmt.Errorf("error encoding tx: %w", err)
	}

	finalSignatures := []models.Signature{}
	for _, sig := range sigV2s {
		signer, _ := common.AddressHexFromBytes(sig.PubKey.Address().Bytes())

		signature := common.HexFromBytes(sig.Data.(*signingtypes.SingleSignatureData).Signature)

		finalSignatures = append(finalSignatures, models.Signature{
			Signer:    signer,
			Signature: signature,
		})
	}

	return string(txBody), finalSignatures, nil
}
