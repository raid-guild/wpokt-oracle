package util

import (
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"

	"context"

	"github.com/cosmos/cosmos-sdk/client"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func SignWithPrivKey(
	ctx context.Context,
	signerData authsigning.SignerData,
	txBuilder client.TxBuilder,
	priv crypto.PrivKey,
	txConfig client.TxConfig,
	accSeq uint64,
) (signingtypes.SignatureV2, error) {
	signMode := signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON

	var sigV2 signingtypes.SignatureV2

	// Generate the bytes to be signed.
	signBytes, err := authsigning.GetSignBytesAdapter(
		ctx, txConfig.SignModeHandler(), signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return sigV2, err
	}

	// Sign those bytes
	signature, err := priv.Sign(signBytes)
	if err != nil {
		return sigV2, err
	}

	// Construct the SignatureV2 struct
	sigData := signingtypes.SingleSignatureData{
		SignMode:  signMode,
		Signature: signature,
	}

	sigV2 = signingtypes.SignatureV2{
		PubKey:   priv.PubKey(),
		Data:     &sigData,
		Sequence: accSeq,
	}

	return sigV2, nil
}
