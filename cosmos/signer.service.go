package cosmos

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/bson"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/common"
	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type CosmosMessageSignerRunnable struct {
	multisigPk *multisig.LegacyAminoPubKey
	signerKey  crypto.PrivKey

	mintControllerMap         map[uint32][]byte
	ethClientMap              map[uint32]eth.EthereumClient
	mailboxMap                map[uint32]eth.MailboxContract
	supportedChainIDsEthereum map[uint32]bool

	config models.CosmosNetworkConfig
	chain  models.Chain
	client cosmos.CosmosClient

	logger *log.Entry

	currentBlockHeight uint64

	db db.DB
}

func (x *CosmosMessageSignerRunnable) Run() {
	x.UpdateCurrentHeight()
	x.SignRefunds()
	x.BroadcastRefunds()
	x.SignMessages()
	x.BroadcastMessages()
}

func (x *CosmosMessageSignerRunnable) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *CosmosMessageSignerRunnable) UpdateCurrentHeight() {
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

func (x *CosmosMessageSignerRunnable) UpdateMessage(
	message *models.Message,
	update bson.M,
) bool {
	err := x.db.UpdateMessage(message.ID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating message")
		return false
	}
	return true
}
func (x *CosmosMessageSignerRunnable) Sign(
	sequence *uint64,
	signatures []models.Signature,
	transactionBody string,
	toAddress []byte,
	amount sdk.Coin,
	memo string,
) (bson.M, error) {

	if sequence == nil {
		gotSequence, err := x.FindMaxSequence()
		if err != nil {
			return nil, fmt.Errorf("error getting sequence: %w", err)
		}
		sequence = &gotSequence
	}

	txBody, finalSignatures, err := CosmosSignTx(
		x.signerKey,
		x.config,
		x.client,
		*sequence,
		signatures,
		transactionBody,
		toAddress,
		amount,
		memo,
	)

	if err != nil {
		return nil, err
	}

	update := bson.M{
		"status":           models.MessageStatusPending,
		"transaction_body": string(txBody),
		"signatures":       finalSignatures,
		"sequence":         sequence,
	}

	if len(finalSignatures) >= int(x.config.MultisigThreshold) {
		update["status"] = models.MessageStatusSigned
	}

	return update, nil
}

func (x *CosmosMessageSignerRunnable) SignMessage(
	messageDoc *models.Message,
) bool {

	logger := x.logger.
		WithField("tx_hash", messageDoc.OriginTransactionHash).
		WithField("section", "sign-message")

	toAddr, err := common.BytesFromAddressHex(messageDoc.Content.MessageBody.RecipientAddress)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing to address")
		return false
	}

	coinAmount, ok := math.NewIntFromString(messageDoc.Content.MessageBody.Amount)
	if !ok {
		logger.Errorf("Error parsing amount")
		return false
	}

	update, err := x.Sign(
		messageDoc.Sequence,
		messageDoc.Signatures,
		messageDoc.TransactionBody,
		toAddr,
		sdk.NewCoin(x.config.CoinDenom, coinAmount),
		"Message from "+messageDoc.OriginTransactionHash+" on "+x.chain.ChainID,
	)

	if err == ErrAlreadySigned {
		return true
	}

	if err != nil {
		logger.WithError(err).Error("Error signing")
		return false
	}

	if lockID, err := x.db.LockWriteSequence(); err != nil {
		logger.WithError(err).Error("Error locking sequence")
		return false
	} else {
		//nolint:errcheck
		defer x.db.Unlock(lockID)
	}

	err = x.db.UpdateMessage(messageDoc.ID, update)
	if err != nil {
		logger.WithError(err).Errorf("Error updating message")
		return false
	}

	return true
}

type ValidateTransactionAndParseDispatchIDEventsResult struct {
	Event         *autogen.MailboxDispatchId
	Confirmations uint64
	TxStatus      models.TransactionStatus
}

func (x *CosmosMessageSignerRunnable) ValidateAndFindDispatchIDEvent(messageDoc *models.Message) (*ValidateTransactionAndParseDispatchIDEventsResult, error) {
	chainDomain := messageDoc.Content.OriginDomain
	txHash := messageDoc.OriginTransactionHash
	messageIDBytes, err := common.BytesFromHex(messageDoc.MessageID)
	if err != nil {
		return nil, fmt.Errorf("error getting message ID bytes: %w", err)
	}

	ethClient, ok := x.ethClientMap[chainDomain]
	if !ok {
		return nil, fmt.Errorf("ethereum client not found")
	}
	mailbox, ok := x.mailboxMap[chainDomain]
	if !ok {
		return nil, fmt.Errorf("mailbox not found")
	}

	receipt, err := ethClient.GetTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction receipt: %w", err)
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return &ValidateTransactionAndParseDispatchIDEventsResult{
			TxStatus: models.TransactionStatusFailed,
		}, nil
	}
	var dispatchEvent *autogen.MailboxDispatchId
	for _, log := range receipt.Logs {
		if log.Address == mailbox.Address() {
			event, err := mailbox.ParseDispatchId(*log)
			if err != nil {
				continue
			}
			if bytes.Equal(event.MessageId[:], messageIDBytes) {
				dispatchEvent = event
				break
			}
		}
	}

	currentBlockHeight, err := ethClient.GetBlockHeight()
	if err != nil {
		return nil, fmt.Errorf("error getting current block height: %w", err)
	}

	result := &ValidateTransactionAndParseDispatchIDEventsResult{
		Event:         dispatchEvent,
		Confirmations: currentBlockHeight - receipt.BlockNumber.Uint64(),
		TxStatus:      models.TransactionStatusPending,
	}

	if result.Confirmations >= ethClient.Confirmations() {
		result.TxStatus = models.TransactionStatusConfirmed
	}
	if dispatchEvent == nil {
		result.TxStatus = models.TransactionStatusInvalid
	}
	return result, nil
}

func (x *CosmosMessageSignerRunnable) ValidateEthereumTxAndSignMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-ethereum-message")
	logger.Debugf("Signing ethereum message")

	result, err := x.ValidateAndFindDispatchIDEvent(messageDoc)
	if err != nil {
		x.logger.WithError(err).Error("Error validating transaction and parsing DispatchId events")
		return false
	}

	if result.TxStatus == models.TransactionStatusPending {
		logger.Debugf("Found pending tx")
		return false
	}

	if result.TxStatus != models.TransactionStatusConfirmed {
		logger.Debugf("Found tx with status %s", result.TxStatus)
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if lockID, err := x.db.LockWriteMessage(messageDoc); err != nil {
		logger.WithError(err).Error("Error locking message")
		return false
	} else {
		//nolint:errcheck
		defer x.db.Unlock(lockID)
	}

	return x.SignMessage(messageDoc)
}

func (x *CosmosMessageSignerRunnable) SignMessages() bool {
	x.logger.Infof("Signing messages")
	addressHex, _ := common.AddressHexFromBytes(x.signerKey.PubKey().Address().Bytes())
	messages, err := x.db.GetPendingMessages(addressHex, x.chain)

	if err != nil {
		x.logger.WithError(err).Errorf("Error getting pending messages")
		return false
	}
	x.logger.Infof("Found %d pending messages", len(messages))
	success := true
	for _, messageDoc := range messages {
		success = x.ValidateEthereumTxAndSignMessage(&messageDoc) && success
	}

	return success
}

func (x *CosmosMessageSignerRunnable) UpdateRefund(
	refund *models.Refund,
	update bson.M,
) bool {
	err := x.db.UpdateRefund(refund.ID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating refund")
		return false
	}
	return true
}

func (x *CosmosMessageSignerRunnable) ValidateRefund(
	refundDoc *models.Refund,
	spenderAddress []byte,
	amount sdk.Coin,
) bool {
	logger := x.logger.
		WithField("tx_hash", refundDoc.OriginTransactionHash).
		WithField("section", "validate-refund")

	recipientAddress, err := common.BytesFromAddressHex(refundDoc.Recipient)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing recipient address")
		return false
	}

	if !bytes.Equal(spenderAddress, recipientAddress) {
		logger.Errorf("Spender address does not match recipient address")
		return false
	}

	coinAmount, ok := math.NewIntFromString(refundDoc.Amount)
	if !ok {
		logger.Errorf("Error parsing amount")
		return false
	}

	refundAmount := sdk.NewCoin(x.config.CoinDenom, coinAmount)
	if !amount.IsEqual(refundAmount) {
		logger.Errorf("Amount does not match refund amount")
		return false
	}

	if refundDoc.TransactionBody == "" {
		return true
	}

	tx, err := utilParseTxBody(x.config.Bech32Prefix, refundDoc.TransactionBody)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing tx body")
		return false
	}

	msgs := tx.GetMsgs()

	msg, ok := msgs[0].(*banktypes.MsgSend)

	if !ok {
		logger.Errorf("Invalid message type")
		return false
	}

	if len(msg.Amount) != 1 {
		logger.Errorf("Invalid amount")
		return false
	}

	refundFinalAmount := refundAmount.Sub(sdk.NewCoin(x.config.CoinDenom, math.NewIntFromUint64(x.config.TxFee)))

	if !msg.Amount[0].IsEqual(refundFinalAmount) {
		logger.Errorf("Amount does not match refund final amount")
		return false
	}

	fromAddress, err := common.AddressBytesFromBech32(x.config.Bech32Prefix, msg.FromAddress)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing from address")
		return false
	}

	if !bytes.Equal(fromAddress, x.multisigPk.Address().Bytes()) {
		logger.Errorf("From address does not match multisig address")
		return false
	}

	toAddress, err := common.AddressBytesFromBech32(x.config.Bech32Prefix, msg.ToAddress)
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

func (x *CosmosMessageSignerRunnable) FindMaxSequence() (uint64, error) {
	lockID, err := x.db.LockReadSequences()
	if err != nil {
		return 0, fmt.Errorf("could not lock sequences: %w", err)
	}
	//nolint:errcheck
	defer x.db.Unlock(lockID)

	maxSequence, err := x.db.FindMaxSequence(x.chain)
	if err != nil {
		return 0, err
	}
	account, err := x.client.GetAccount(x.config.MultisigAddress)
	if err != nil {
		return 0, err
	}
	if maxSequence == nil {
		return account.Sequence, nil
	}
	nextSequence := *maxSequence + 1
	if nextSequence > account.Sequence {
		return nextSequence, nil
	}

	return account.Sequence, nil
}

func (x *CosmosMessageSignerRunnable) SignRefund(
	refundDoc *models.Refund,
) bool {

	logger := x.logger.
		WithField("tx_hash", refundDoc.OriginTransactionHash).
		WithField("section", "sign-refund")

	spender, err := common.BytesFromAddressHex(refundDoc.Recipient)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing spender address")
		return false
	}

	amount, ok := math.NewIntFromString(refundDoc.Amount)
	if !ok {
		logger.Errorf("Error parsing amount")
		return false
	}

	coinAmount := sdk.NewCoin(x.config.CoinDenom, amount)

	update, err := x.Sign(
		refundDoc.Sequence,
		refundDoc.Signatures,
		refundDoc.TransactionBody,
		spender,
		coinAmount,
		"Refund for "+refundDoc.OriginTransactionHash,
	)

	if err == ErrAlreadySigned {
		return true
	}

	if err != nil {
		logger.WithError(err).Error("Error signing")
		return false
	}

	if lockID, err := x.db.LockWriteSequence(); err != nil {
		logger.WithError(err).Error("Error locking sequence")
		return false
	} else {
		//nolint:errcheck
		defer x.db.Unlock(lockID)
	}

	err = x.db.UpdateRefund(refundDoc.ID, update)
	if err != nil {
		logger.WithError(err).Errorf("Error updating refund")
		return false
	}

	return true
}

func (x *CosmosMessageSignerRunnable) BroadcastMessage(messageDoc *models.Message) bool {

	logger := x.logger.
		WithField("tx_hash", messageDoc.OriginTransactionHash).
		WithField("section", "broadcast-message")

	txBuilder, txCfg, err := utilWrapTxBuilder(x.config.Bech32Prefix, messageDoc.TransactionBody)
	if err != nil {
		logger.WithError(err).Errorf("Error wrapping tx builder")
		return false
	}

	valid := x.ValidateSignaturesAndAddMultiSignatureToTxConfig(messageDoc.OriginTransactionHash, *messageDoc.Sequence, txCfg, txBuilder)
	if !valid {
		return x.ResetMessage(messageDoc)
	}

	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		logger.WithError(err).Errorf("Error encoding tx")
		return false
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

	txHash0x := common.Ensure0xPrefix(txHash)

	update := bson.M{
		"status":           models.MessageStatusBroadcasted,
		"transaction_body": string(txJSON),
		"transaction_hash": txHash0x,
	}

	return x.UpdateMessage(messageDoc, update)
}

func (x *CosmosMessageSignerRunnable) ValidateEthereumTxAndBroadcastMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "broadcast-ethereum-message")
	logger.Debugf("Broadcasting ethereum message")

	result, err := x.ValidateAndFindDispatchIDEvent(messageDoc)
	if err != nil {
		x.logger.WithError(err).Error("Error validating transaction and parsing DispatchId events")
		return false
	}

	if result.TxStatus == models.TransactionStatusPending {
		logger.Debugf("Found pending tx")
		return false
	}

	if result.TxStatus != models.TransactionStatusConfirmed {
		logger.Debugf("Found tx with status %s", result.TxStatus)
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if lockID, err := x.db.LockWriteMessage(messageDoc); err != nil {
		logger.WithError(err).Error("Error locking message")
		return false
	} else {
		//nolint:errcheck
		defer x.db.Unlock(lockID)
	}

	return x.BroadcastMessage(messageDoc)

}

func (x *CosmosMessageSignerRunnable) BroadcastMessages() bool {
	x.logger.Infof("Broadcasting messages")
	messages, err := x.db.GetSignedMessages(x.chain)
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting signed messages")
		return false
	}
	x.logger.Infof("Found %d signed messages", len(messages))
	success := true
	for _, messageDoc := range messages {
		success = x.ValidateEthereumTxAndBroadcastMessage(&messageDoc) && success
	}

	return success
}

var utilValidateSignature = util.ValidateSignature
var multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2

func (x *CosmosMessageSignerRunnable) ValidateSignaturesAndAddMultiSignatureToTxConfig(
	originTxHash string,
	sequence uint64,
	txCfg client.TxConfig,
	txBuilder client.TxBuilder,
) bool {
	logger := x.logger.
		WithField("tx_hash", originTxHash).
		WithField("section", "validate-signatures")

	sigV2s, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		logger.WithError(err).Error("Error getting signatures")
		return false
	}

	if len(sigV2s) == 0 || len(sigV2s) < int(x.config.MultisigThreshold) {
		logger.Errorf("Not enough signatures")
		return false
	}

	account, err := x.client.GetAccount(x.config.MultisigAddress)

	if err != nil {
		logger.WithError(err).Error("Error getting account")
		return false
	}

	multisigSig := multisigtypes.NewMultisig(len(x.multisigPk.PubKeys))

	// read each signature and add it to the multisig if valid
	for _, sig := range sigV2s {
		if err := utilValidateSignature(x.config, &sig, account.AccountNumber, sequence, txCfg, txBuilder); err != nil {
			logger.WithError(err).Error("Error validating signature")
			return false
		}
		if err := multisigtypesAddSignatureV2(multisigSig, sig, x.multisigPk.GetPubKeys()); err != nil {
			logger.WithError(err).Error("Error adding signature")
			return false
		}
	}

	sigV2 := signingtypes.SignatureV2{
		PubKey:   x.multisigPk,
		Data:     multisigSig,
		Sequence: sequence,
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		logger.WithError(err).Error("Error setting signatures")
		return false
	}

	// TODO: add more validation
	return true
}
func (x *CosmosMessageSignerRunnable) ResetRefund(
	refund *models.Refund,
) bool {
	update := bson.M{
		"status":           models.RefundStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}

	return x.UpdateRefund(refund, update)
}

func (x *CosmosMessageSignerRunnable) ResetMessage(
	message *models.Message,
) bool {

	update := bson.M{
		"status":           models.MessageStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}

	return x.UpdateMessage(message, update)
}

func (x *CosmosMessageSignerRunnable) BroadcastRefund(refundDoc *models.Refund) bool {
	logger := x.logger.
		WithField("tx_hash", refundDoc.OriginTransactionHash).
		WithField("section", "broadcast-refund")

	txBuilder, txCfg, err := utilWrapTxBuilder(x.config.Bech32Prefix, refundDoc.TransactionBody)
	if err != nil {
		logger.WithError(err).Error("Error wrapping tx builder")
		return false
	}

	valid := x.ValidateSignaturesAndAddMultiSignatureToTxConfig(refundDoc.OriginTransactionHash, *refundDoc.Sequence, txCfg, txBuilder)
	if !valid {
		return x.ResetRefund(refundDoc)
	}

	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		logger.WithError(err).Errorf("Error encoding tx")
		return false
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

	txHash0x := common.Ensure0xPrefix(txHash)

	update := bson.M{
		"status":           models.RefundStatusBroadcasted,
		"transaction_body": string(txJSON),
		"transaction_hash": txHash0x,
	}

	return x.UpdateRefund(refundDoc, update)
}

func (x *CosmosMessageSignerRunnable) ValidateCosmosTxAndBroadcastRefund(
	refundDoc models.Refund,
) bool {
	logger := x.logger.WithField("tx_hash", refundDoc.OriginTransactionHash).WithField("section", "validateCosmosTxAndBroadcastRefund")
	if !x.ValidateCosmosTx(refundDoc) {
		logger.Debugf("Refund is not valid")
		return false
	}

	if lockID, err := x.db.LockWriteRefund(&refundDoc); err != nil {
		logger.WithError(err).Error("Error locking refund")
		return false
	} else {
		defer x.db.Unlock(lockID)
	}

	return x.BroadcastRefund(&refundDoc)
}

func (x *CosmosMessageSignerRunnable) BroadcastRefunds() bool {
	x.logger.Infof("Broadcasting refunds")
	refunds, err := x.db.GetSignedRefunds()
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting signed refunds")
		return false
	}
	x.logger.Infof("Found %d signed refunds", len(refunds))
	success := true
	for _, refundDoc := range refunds {
		success = x.ValidateCosmosTxAndBroadcastRefund(refundDoc) && success
	}

	return success
}

func (x *CosmosMessageSignerRunnable) ValidateCosmosTx(
	refundDoc models.Refund,
) bool {
	logger := x.logger.WithField("tx_hash", refundDoc.OriginTransactionHash).WithField("section", "validateCosmosTxAndSignRefund")
	txResponse, err := x.client.GetTx(refundDoc.OriginTransactionHash)
	if err != nil {
		logger.WithError(err).Errorf("Error getting tx")
		return false
	}

	result, err := utilValidateTxToCosmosMultisig(txResponse, x.config, x.supportedChainIDsEthereum, x.currentBlockHeight)

	if err != nil {
		logger.WithError(err).Errorf("Error validating tx")
		x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
		return false
	}

	if result.TxStatus == models.TransactionStatusPending {
		logger.Debugf("Tx is pending")
		return false
	}

	if !result.NeedsRefund {
		logger.Debugf("Tx does not need refund")
		x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
		return false
	}

	if result.TxStatus != models.TransactionStatusConfirmed {
		logger.Debugf("Tx is invalid")
		x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
		return false
	}

	if !x.ValidateRefund(&refundDoc, result.SenderAddress, result.Amount) {
		logger.Warnf("Invalid refund")
		x.UpdateRefund(&refundDoc, bson.M{"status": models.RefundStatusInvalid})
		return false
	}

	return true
}

func (x *CosmosMessageSignerRunnable) ValidateCosmosTxAndSignRefund(
	refundDoc models.Refund,
) bool {
	logger := x.logger.WithField("tx_hash", refundDoc.OriginTransactionHash).WithField("section", "validateCosmosTxAndSignRefund")

	if !x.ValidateCosmosTx(refundDoc) {
		logger.Debugf("Refund is not valid")
		return false
	}

	if lockID, err := x.db.LockWriteRefund(&refundDoc); err != nil {
		logger.WithError(err).Error("Error locking refund")
		return false
	} else {
		defer x.db.Unlock(lockID)
	}

	return x.SignRefund(&refundDoc)
}

func (x *CosmosMessageSignerRunnable) SignRefunds() bool {
	x.logger.Infof("Signing refunds")
	addressHex, _ := common.AddressHexFromBytes(x.signerKey.PubKey().Address().Bytes())
	refunds, err := x.db.GetPendingRefunds(addressHex)
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting pending refunds")
		return false
	}
	x.logger.Infof("Found %d pending refunds", len(refunds))
	success := true
	for _, refundDoc := range refunds {
		success = x.ValidateCosmosTxAndSignRefund(refundDoc) && success
	}

	return success
}

var ethNewClient = eth.NewClient
var ethNewMailboxContract = eth.NewMailboxContract

func NewMessageSigner(
	mnemonic string,
	config models.CosmosNetworkConfig,
	mintControllerMap map[uint32][]byte,
	ethNetworks []models.EthereumNetworkConfig,
) service.Runnable {
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
		pKey, err := common.CosmosPublicKeyFromHex(pk)
		if err != nil {
			logger.WithError(err).Fatalf("Error parsing public key")
		}
		pks = append(pks, pKey)
	}

	multisigPk := multisig.NewLegacyAminoPubKey(int(config.MultisigThreshold), pks)
	multisigAddress, _ := common.Bech32FromBytes(config.Bech32Prefix, multisigPk.Address().Bytes())

	if !strings.EqualFold(multisigAddress, config.MultisigAddress) {
		logger.Fatalf("Multisig address does not match config")
	}

	client, err := cosmosNewClient(config)
	if err != nil {
		logger.WithError(err).Fatalf("Error creating cosmos client")
	}

	privKey, err := common.CosmosPrivateKeyFromMnemonic(mnemonic)
	if err != nil {
		logger.WithError(err).Fatalf("Error getting private key from mnemonic")
	}

	ethClientMap := make(map[uint32]eth.EthereumClient)
	mailboxMap := make(map[uint32]eth.MailboxContract)
	supportedChainIDsEthereum := make(map[uint32]bool)

	for _, ethConfig := range ethNetworks {
		ethClient, err := ethNewClient(ethConfig)
		if err != nil {
			logger.WithError(err).
				WithField("chain_name", ethConfig.ChainName).
				WithField("chain_id", ethConfig.ChainID).
				Fatalf("Error creating ethereum client")
		}
		chainDomain := ethClient.Chain().ChainDomain
		mailbox, err := ethNewMailboxContract(common.HexToAddress(ethConfig.MailboxAddress), ethClient.GetClient())
		if err != nil {
			logger.WithError(err).
				WithField("chain_name", ethConfig.ChainName).
				WithField("chain_id", ethConfig.ChainID).
				Fatalf("Error creating mailbox contract")
		}
		ethClientMap[chainDomain] = ethClient
		mailboxMap[chainDomain] = mailbox
		supportedChainIDsEthereum[chainDomain] = true
	}

	x := &CosmosMessageSignerRunnable{
		multisigPk: multisigPk,

		currentBlockHeight: 0,
		client:             client,

		signerKey: privKey,

		chain: utilParseChain(config),

		mintControllerMap:         mintControllerMap,
		ethClientMap:              ethClientMap,
		mailboxMap:                mailboxMap,
		supportedChainIDsEthereum: supportedChainIDsEthereum,

		config: config,

		logger: logger,

		db: dbNewDB(),
	}

	x.UpdateCurrentHeight()

	logger.Infof("Initialized")

	return x
}
