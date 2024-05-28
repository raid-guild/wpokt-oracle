package cosmos

import (
	"bytes"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"go.mongodb.org/mongo-driver/bson"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/common"
	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"

	"context"

	"github.com/cosmos/cosmos-sdk/client"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type MessageSignerRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	multisigAddress   string
	multisigThreshold uint64
	multisigPk        *multisig.LegacyAminoPubKey

	signerKey crypto.PrivKey

	bech32Prefix string
	coinDenom    string
	feeAmount    sdk.Coin

	confirmations uint64

	chain  models.Chain
	client cosmos.CosmosClient

	logger *log.Entry
}

func (x *MessageSignerRunner) Run() {
	x.UpdateCurrentHeight()
	x.SignRefunds()
	x.SignMessages()
}

func (x *MessageSignerRunner) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *MessageSignerRunner) UpdateCurrentHeight() {
	height, err := x.client.GetLatestBlockHeight()
	if err != nil {
		x.logger.
			WithError(err).
			Error("could not get current block height")
		return
	}
	x.currentBlockHeight = uint64(height)
	x.logger.
		WithField("current_block_height", x.currentBlockHeight).
		Info("updated current block height")
}

func (x *MessageSignerRunner) SignMessages() bool {
	return true
}

func (x *MessageSignerRunner) UpdateRefund(
	tx *models.Refund,
	update bson.M,
) bool {
	err := util.UpdateRefund(tx, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating refund")
		return false
	}
	return true
}

func (x *MessageSignerRunner) ValidateRefund(
	txResponse *sdk.TxResponse,
	refundDoc *models.Refund,
	spender string,
	amount sdk.Coin,
) bool {
	logger := x.logger.
		WithField("tx_hash", refundDoc.OriginTransactionHash).
		WithField("section", "validate-refund")

	spenderAddress, err := util.AddressBytesFromBech32(x.bech32Prefix, spender)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing spender address")
		return false
	}

	recipientAddress, err := util.AddressBytesFromHexString(refundDoc.Recipient)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing recipient address")
		return false
	}

	if !bytes.Equal(spenderAddress, recipientAddress) {
		logger.Errorf("Spender address does not match recipient address")
		return false
	}

	refundAmount := sdk.NewCoin(x.coinDenom, math.NewInt(int64(refundDoc.Amount)))
	if !amount.IsEqual(refundAmount) {
		logger.Errorf("Amount does not match refund amount")
		return false
	}

	tx, err := util.ParseTxBody(x.bech32Prefix, refundDoc.TransactionBody)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing tx body")
		return false
	}

	msgs := tx.GetMsgs()

	msg := msgs[0].(*banktypes.MsgSend)

	if len(msg.Amount) != 1 {
		logger.Errorf("Invalid amount")
		return false
	}

	refundFinalAmount := refundAmount.Sub(x.feeAmount)

	if !msg.Amount[0].IsEqual(refundFinalAmount) {
		logger.Errorf("Amount does not match refund final amount")
		return false
	}

	fromAddress, err := util.AddressBytesFromBech32(x.bech32Prefix, msg.FromAddress)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing from address")
		return false
	}

	if !bytes.Equal(fromAddress, x.multisigPk.Address().Bytes()) {
		logger.Errorf("From address does not match multisig address")
		return false
	}

	toAddress, err := util.AddressBytesFromBech32(x.bech32Prefix, msg.ToAddress)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing to address")
		return false
	}

	if !bytes.Equal(toAddress, recipientAddress) {
		logger.Errorf("To address does not match recipient address")
		return false
	}

	return true
}

func isTxSigner(user []byte, signers [][]byte) bool {
	for _, s := range signers {
		if bytes.Equal(user, s) {
			return true
		}
	}

	return false
}

func (x *MessageSignerRunner) SignRefund(
	txResponse *sdk.TxResponse,
	refundDoc *models.Refund,
	spender string,
	amount sdk.Coin,
) bool {

	logger := x.logger.
		WithField("tx_hash", refundDoc.OriginTransactionHash).
		WithField("section", "sign-refund")

	valid := x.ValidateRefund(txResponse, refundDoc, spender, amount)
	if !valid {
		return x.UpdateRefund(refundDoc, bson.M{"status": models.RefundStatusInvalid})
	}

	for _, sig := range refundDoc.Signatures {
		signer, err := util.AddressBytesFromHexString(sig.Signer)
		if err != nil {
			logger.WithError(err).Errorf("Error parsing signer")
			return false
		}
		if bytes.Equal(signer, x.signerKey.PubKey().Address().Bytes()) {
			logger.Infof("Already signed")
			return true
		}
	}

	// can ignore error here because we already validated
	tx, _ := util.ParseTxBody(x.bech32Prefix, refundDoc.TransactionBody)

	txConfig := util.NewTxConfig(x.bech32Prefix)

	txBuilder, err := txConfig.WrapTxBuilder(tx)
	if err != nil {
		logger.WithError(err).Error("Error wrapping tx builder")
		return false
	}

	// check whether the address is a signer
	signers, err := txBuilder.GetTx().GetSigners()
	if err != nil {
		logger.WithError(err).Error("Error getting signers")
		return false
	}

	if !isTxSigner(x.multisigPk.Address().Bytes(), signers) {
		logger.Errorf("Address is not a signer")
		return false
	}

	sequence, err := util.FindMaxSequence(x.chain)

	if err != nil {
		logger.WithError(err).Error("Error getting sequence")
		return false
	}

	account, err := x.client.GetAccount(x.multisigAddress)

	if err != nil {
		logger.WithError(err).Error("Error getting account")
		return false
	}

	overwriteSig := false

	pubKey := x.signerKey.PubKey()

	signerData := authsigning.SignerData{
		ChainID:       x.chain.ChainID,
		AccountNumber: account.AccountNumber,
		Sequence:      sequence,
		PubKey:        pubKey,
		Address:       sdk.AccAddress(pubKey.Address()).String(),
	}

	var prevSignatures []signingtypes.SignatureV2
	if !overwriteSig {
		prevSignatures, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			logger.WithError(err).Error("Error getting signatures")
			return false
		}
	}

	sig, err := SignWithPrivKey(
		context.Background(),
		signerData,
		txBuilder,
		x.signerKey,
		txConfig,
		sequence,
	)
	if err != nil {
		logger.WithError(err).Error("Error signing")
		return false
	}

	if overwriteSig {
		err = txBuilder.SetSignatures(sig)
	} else {
		prevSignatures = append(prevSignatures, sig)
		err = txBuilder.SetSignatures(prevSignatures...)
	}

	if err != nil {
		logger.WithError(err).Errorf("unable to set signatures on payload")
		return false
	}

	txBody, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		logger.WithError(err).Errorf("unable to encode tx")
		return false
	}

	signatures := []models.Signature{}

	for _, sig := range prevSignatures {
		signer := util.HexFromBytes(sig.PubKey.Address().Bytes())
		signature := util.HexFromBytes(sig.Data.(*signingtypes.SingleSignatureData).Signature)

		signatures = append(signatures, models.Signature{
			Signer:    signer,
			Signature: signature,
		})
	}

	update := bson.M{
		"status":           models.RefundStatusPending,
		"transaction_body": string(txBody),
		"signatures":       signatures,
	}

	if len(signatures) >= int(x.multisigThreshold) {
		update["status"] = models.RefundStatusSigned
	}

	err = util.UpdateRefund(refundDoc, update)
	if err != nil {
		logger.WithError(err).Errorf("Error updating refund")
		return false
	}

	return true
}

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

func (x *MessageSignerRunner) SignRefunds() bool {
	x.logger.Infof("Signing refunds")
	addressHex := util.HexFromBytes(x.signerKey.PubKey().Address().Bytes())
	refunds, err := util.GetPendingRefunds(addressHex)
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting pending refunds")
		return false
	}
	x.logger.Infof("Found %d pending refunds", len(refunds))
	success := true
	for _, refundDoc := range refunds {
		logger := x.logger.WithField("tx_hash", refundDoc.OriginTransactionHash).WithField("section", "sign-refunds")
		txResponse, err := x.client.GetTx(refundDoc.OriginTransactionHash)
		if err != nil {
			logger.WithError(err).Errorf("Error getting tx")
			success = false
			continue
		}
		if txResponse.Code != 0 {
			logger.Infof("Found tx with error")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
			continue
		}

		tx := &tx.Tx{}
		err = tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			logger.Errorf("Error unmarshalling tx")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
			continue
		}

		coinsReceived, err := util.ParseCoinsReceivedEvents(x.coinDenom, x.multisigAddress, txResponse.Events)
		if err != nil {
			logger.WithError(err).Errorf("Error parsing coins received events")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
			continue
		}

		coinsSpentSender, coinsSpent, err := util.ParseCoinsSpentEvents(x.coinDenom, txResponse.Events)
		if err != nil {
			logger.WithError(err).Errorf("Error parsing coins spent events")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
			continue
		}

		if coinsReceived.IsZero() || coinsSpent.IsZero() {
			logger.
				Debugf("Found tx with zero coins")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
			continue
		}

		if coinsReceived.IsLTE(x.feeAmount) {
			logger.
				Debugf("Found tx with amount too low")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
			continue
		}

		txHeight := txResponse.Height
		if txHeight <= 0 || uint64(txHeight) > x.currentBlockHeight {
			logger.WithField("tx_height", txHeight).Debugf("Found tx with invalid height")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
			continue
		}

		confirmations := x.currentBlockHeight - uint64(txHeight)

		if confirmations < x.confirmations {
			logger.WithField("confirmations", confirmations).Debugf("Found tx with not enough confirmations")
			success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusPending})
			continue
		}

		if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
			logger.Debugf("Found tx with invalid coins")
			success = success && x.SignRefund(txResponse, &refundDoc, coinsSpentSender, coinsSpent)
			continue
		}

		memo, err := util.ValidateMemo(tx.Body.Memo)
		if err != nil {
			logger.WithError(err).WithField("memo", tx.Body.Memo).Debugf("Found invalid memo")
			success = success && x.SignRefund(txResponse, &refundDoc, coinsSpentSender, coinsSpent)

			continue
		}

		logger.WithField("memo", memo).Errorf("Found refund with a valid memo")
		success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
	}

	if success {
		x.startBlockHeight = x.currentBlockHeight
	}

	return success
}

func (x *MessageSignerRunner) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
	if lastHealth == nil || lastHealth.BlockHeight == 0 {
		x.logger.Debugf("Invalid last health")
	} else {
		x.logger.Debugf("Last block height: %d", lastHealth.BlockHeight)
		x.startBlockHeight = lastHealth.BlockHeight
	}
	if x.startBlockHeight == 0 {
		x.logger.Debugf("Start block height is zero")
		x.startBlockHeight = x.currentBlockHeight
	} else if x.startBlockHeight > x.currentBlockHeight {
		x.logger.Debugf("Start block height is greater than current block height")
		x.startBlockHeight = x.currentBlockHeight
	}
	x.logger.Infof("Initialized start block height: %d", x.startBlockHeight)
}

func NewMessageSigner(mnemonic string, config models.CosmosNetworkConfig, lastHealth *models.RunnerServiceStatus) service.Runner {
	logger := log.
		WithField("module", "cosmos").
		WithField("service", "signer").
		WithField("chain_name", strings.ToLower(config.ChainName)).
		WithField("chain_id", strings.ToLower(config.ChainID))

	if !config.MessageSigner.Enabled {
		logger.Fatalf("Message signer is not enabled")
	}

	logger.Debugf("Initializing")

	var pks []crypto.PubKey
	for _, pk := range config.MultisigPublicKeys {
		pKey, err := util.PubKeyFromHex(pk)
		if err != nil {
			logger.WithError(err).Fatalf("Error parsing public key")
		}
		pks = append(pks, pKey)
	}

	multisigPk := multisig.NewLegacyAminoPubKey(int(config.MultisigThreshold), pks)
	multisigAddress, err := util.Bech32FromAddressBytes(config.Bech32Prefix, multisigPk.Address().Bytes())
	if err != nil {
		logger.WithError(err).Fatalf("Error creating multisig address")
	}

	if !strings.EqualFold(multisigAddress, config.MultisigAddress) {
		logger.Fatalf("Multisig address does not match config")
	}

	client, err := cosmos.NewClient(config)
	if err != nil {
		logger.WithError(err).Errorf("Error creating cosmos client")
	}

	feeAmount := sdk.NewCoin("upokt", math.NewInt(int64(config.TxFee)))

	privKey, err := common.CosmosPrivateKeyFromMnemonic(mnemonic)
	if err != nil {
		logger.WithError(err).Fatalf("Error getting private key from mnemonic")
	}

	x := &MessageSignerRunner{
		multisigPk:        multisigPk,
		multisigThreshold: config.MultisigThreshold,
		multisigAddress:   multisigAddress,

		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		client:             client,
		feeAmount:          feeAmount,

		signerKey: privKey,

		chain:         util.ParseChain(config),
		confirmations: config.Confirmations,

		bech32Prefix: config.Bech32Prefix,
		coinDenom:    config.CoinDenom,

		logger: logger,
	}

	x.UpdateCurrentHeight()

	x.InitStartBlockHeight(lastHealth)

	logger.Infof("Initialized")

	return x
}
