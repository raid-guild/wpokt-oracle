package util

import (
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"

	"context"

	"github.com/cosmos/cosmos-sdk/client"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var signMode = signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON

func SignWithPrivKey(
	ctx context.Context,
	signerData authsigning.SignerData,
	txBuilder client.TxBuilder,
	priv crypto.PrivKey,
	txConfig client.TxConfig,
	accSeq uint64,
) (sigV2 signingtypes.SignatureV2, msg []byte, err error) {

	// Generate the bytes to be signed.
	msg, err = authsigning.GetSignBytesAdapter(ctx, txConfig.SignModeHandler(), signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return sigV2, msg, err
	}

	// Sign those bytes
	signature, err := priv.Sign(msg)
	if err != nil {
		return sigV2, msg, err
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

	return sigV2, msg, nil
}
