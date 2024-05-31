package cosmos

import (
	"bytes"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/anypb"

	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

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

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type MessageSignerRunner struct {
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
	x.BroadcastRefunds()
	x.SignMessages()
	x.BroadcastMessages()
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
	refund *models.Refund,
	update bson.M,
) bool {
	err := util.UpdateRefund(refund.ID, update)
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

	spenderAddress, err := util.BytesFromBech32(x.bech32Prefix, spender)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing spender address")
		return false
	}

	recipientAddress, err := util.BytesFromHex(refundDoc.Recipient)
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

	fromAddress, err := util.BytesFromBech32(x.bech32Prefix, msg.FromAddress)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing from address")
		return false
	}

	if !bytes.Equal(fromAddress, x.multisigPk.Address().Bytes()) {
		logger.Errorf("From address does not match multisig address")
		return false
	}

	toAddress, err := util.BytesFromBech32(x.bech32Prefix, msg.ToAddress)
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

	var sequence uint64

	if refundDoc.Sequence != nil {
		sequence = *refundDoc.Sequence
	} else {
		var err error
		maxSequence, err := util.FindMaxSequence(x.chain)
		if err != nil {
			logger.WithError(err).Error("Error getting sequence")
			return false
		}
		sequence = maxSequence + 1
	}

	for _, sig := range refundDoc.Signatures {
		signer, err := util.BytesFromHex(sig.Signer)
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

	account, err := x.client.GetAccount(x.multisigAddress)

	if err != nil {
		logger.WithError(err).Error("Error getting account")
		return false
	}

	pubKey := x.signerKey.PubKey()

	signerData := authsigning.SignerData{
		ChainID:       x.chain.ChainID,
		AccountNumber: account.AccountNumber,
		Sequence:      sequence,
		PubKey:        pubKey,
		Address:       sdk.AccAddress(pubKey.Address()).String(),
	}

	sigV2, err := util.SignWithPrivKey(
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

	sigV2s, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		logger.WithError(err).Error("Error getting signatures")
		return false
	}

	sigV2s = append(sigV2s, sigV2)
	err = txBuilder.SetSignatures(sigV2s...)

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
	for _, sig := range sigV2s {
		signer := util.HexFromBytes(sig.PubKey.Address().Bytes())
		signature := util.HexFromBytes(sig.Data.(*signingtypes.SingleSignatureData).Signature)

		signatures = append(signatures, models.Signature{
			Signer:    signer,
			Signature: signature,
		})
	}

	seq := uint64(sequence)

	update := bson.M{
		"status":           models.RefundStatusPending,
		"transaction_body": string(txBody),
		"signatures":       signatures,
		"sequence":         &seq,
	}

	if len(signatures) >= int(x.multisigThreshold) {
		update["status"] = models.RefundStatusSigned
	}

	err = util.UpdateRefund(refundDoc.ID, update)
	if err != nil {
		logger.WithError(err).Errorf("Error updating refund")
		return false
	}

	refundDoc.TransactionBody = string(txBody)
	refundDoc.Signatures = signatures
	refundDoc.Sequence = &seq
	refundDoc.Status = models.RefundStatusPending

	return true
}

func (x *MessageSignerRunner) BroadcastMessages() bool {
	return true
}

func (x *MessageSignerRunner) ValidateSignatures(
	refundDoc *models.Refund,
	txCfg client.TxConfig,
	txBuilder client.TxBuilder,
) bool {
	logger := x.logger.
		WithField("tx_hash", refundDoc.OriginTransactionHash).
		WithField("section", "validate-signatures")

	sigV2s, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		logger.WithError(err).Error("Error getting signatures")
		return false
	}

	if len(sigV2s) < int(x.multisigThreshold) {
		logger.Errorf("Not enough signatures")
		return false
	}

	account, err := x.client.GetAccount(x.multisigAddress)

	if err != nil {
		logger.WithError(err).Error("Error getting account")
		return false
	}

	multisigPub := x.multisigPk
	multisigSig := multisigtypes.NewMultisig(len(multisigPub.PubKeys))

	// read each signature and add it to the multisig if valid

	for _, sig := range sigV2s {
		anyPk, err := codectypes.NewAnyWithValue(sig.PubKey)
		if err != nil {
			logger.WithError(err).Error("Error creating any pubkey")
			return false
		}
		txSignerData := txsigning.SignerData{
			ChainID:       x.chain.ChainID,
			AccountNumber: account.AccountNumber,
			Sequence:      *refundDoc.Sequence,
			Address:       sdk.AccAddress(sig.PubKey.Address()).String(),
			PubKey: &anypb.Any{
				TypeUrl: anyPk.TypeUrl,
				Value:   anyPk.Value,
			},
		}
		builtTx := txBuilder.GetTx()
		adaptableTx, ok := builtTx.(authsigning.V2AdaptableTx)
		if !ok {
			// return fmt.Errorf("expected Tx to be signing.V2AdaptableTx, got %T", builtTx)
			logger.Errorf("expected Tx to be signing.V2AdaptableTx, got %T", builtTx)
			return false
		}
		txData := adaptableTx.GetSigningTxData()

		err = authsigning.VerifySignature(context.Background(), sig.PubKey, txSignerData, sig.Data,
			txCfg.SignModeHandler(), txData)
		if err != nil {
			addr, _ := sdk.AccAddressFromHexUnsafe(sig.PubKey.Address().String())
			// return fmt.Errorf("couldn't verify signature for address %s", addr)
			logger.Errorf("couldn't verify signature for address %s", addr)
			return false
		}

		if err := multisigtypes.AddSignatureV2(multisigSig, sig, multisigPub.GetPubKeys()); err != nil {
			// return err
			logger.WithError(err).Error("Error adding signature")
			return false
		}
	}

	sigV2 := signingtypes.SignatureV2{
		PubKey:   multisigPub,
		Data:     multisigSig,
		Sequence: *refundDoc.Sequence,
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		logger.WithError(err).Error("Error setting signatures")
		return false
	}

	// TODO: add more validation
	return true
}

func (x *MessageSignerRunner) BroadcastRefund(
	txResponse *sdk.TxResponse,
	refundDoc *models.Refund,
	spender string,
	amount sdk.Coin,
) bool {

	logger := x.logger.
		WithField("tx_hash", refundDoc.OriginTransactionHash).
		WithField("section", "broadcast-refund")

	valid := x.ValidateRefund(txResponse, refundDoc, spender, amount)
	if !valid {
		return x.UpdateRefund(refundDoc, bson.M{"status": models.RefundStatusInvalid})
	}

	tx, err := util.ParseTxBody(x.bech32Prefix, refundDoc.TransactionBody)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing tx body")
		return false
	}

	txCfg := util.NewTxConfig(x.bech32Prefix)

	txBuilder, err := txCfg.WrapTxBuilder(tx)
	if err != nil {
		logger.WithError(err).Errorf("Error wrapping tx builder")
		return false
	}

	valid = x.ValidateSignatures(refundDoc, txCfg, txBuilder)
	if !valid {
		err := txBuilder.SetSignatures()
		if err != nil {
			logger.WithError(err).Errorf("Error setting signatures")
			return false
		}
		txBody, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			logger.WithError(err).Errorf("Error encoding tx")
			return false
		}
		update := bson.M{
			"status":           models.RefundStatusPending,
			"transaction_body": string(txBody),
			"signatures":       []models.Signature{},
		}
		return x.UpdateRefund(refundDoc, update)
	}

	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		logger.WithError(err).Errorf("Error encoding tx")
	}

	txBytes, err := txCfg.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		logger.WithError(err).Errorf("Error encoding tx")
		return false
	}

	txHash, err := x.client.BroadcastTx(txBytes)
	if err != nil {
		logger.WithError(err).Errorf("Error broadcasting tx")
		return false
	}

	txHash0x := util.Ensure0xPrefix(txHash)

	update := bson.M{
		"status":           models.RefundStatusBroadcasted,
		"transaction_body": string(txJSON),
		"transaction_hash": txHash0x,
	}

	success := x.UpdateRefund(refundDoc, update)

	if success {
		refundDoc.TransactionHash = txHash0x
		refundDoc.TransactionBody = string(txJSON)
		refundDoc.Status = models.RefundStatusBroadcasted
	}

	return true
}

func (x *MessageSignerRunner) BroadcastRefunds() bool {
	x.logger.Infof("Broadcasting refunds")
	refunds, err := util.GetSignedRefunds()
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting signed refunds")
		return false
	}
	x.logger.Infof("Found %d signed refunds", len(refunds))
	success := true
	for _, refundDoc := range refunds {
		logger := x.logger.WithField("tx_hash", refundDoc.OriginTransactionHash).WithField("section", "broadcast-refunds")
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
			success = success && x.BroadcastRefund(txResponse, &refundDoc, coinsSpentSender, coinsSpent)
			continue
		}

		memo, err := util.ValidateMemo(tx.Body.Memo)
		if err != nil {
			logger.WithError(err).WithField("memo", tx.Body.Memo).Debugf("Found invalid memo")
			success = success && x.BroadcastRefund(txResponse, &refundDoc, coinsSpentSender, coinsSpent)
			continue
		}

		logger.WithField("memo", memo).Errorf("Found refund with a valid memo")
		success = success && x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
	}

	return success
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

	return success
}

func NewMessageSigner(mnemonic string, config models.CosmosNetworkConfig) service.Runner {
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
	multisigAddress, err := util.Bech32FromBytes(config.Bech32Prefix, multisigPk.Address().Bytes())
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

	logger.Infof("Initialized")

	return x
}
