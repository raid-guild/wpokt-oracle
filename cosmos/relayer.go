package cosmos

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/common"
	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type MessageRelayerRunner struct {
	multisigPk *multisig.LegacyAminoPubKey

	config models.CosmosNetworkConfig
	chain  models.Chain
	client cosmos.CosmosClient

	logger *log.Entry

	startBlockHeight   uint64
	currentBlockHeight uint64
}

func (x *MessageRelayerRunner) Run() {
	x.UpdateCurrentHeight()
	x.SyncRefunds()
	x.SyncMessages()
	x.RelayTransactions()
}

func (x *MessageRelayerRunner) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *MessageRelayerRunner) UpdateCurrentHeight() {
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

func (x *MessageRelayerRunner) UpdateRefund(
	refundID *primitive.ObjectID,
	update bson.M,
) bool {
	err := db.UpdateRefund(refundID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating refund")
		return false
	}
	return true
}

func (x *MessageRelayerRunner) UpdateMessage(
	messageID *primitive.ObjectID,
	update bson.M,
) bool {
	err := db.UpdateMessage(messageID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating message")
		return false
	}
	return true
}

func (x *MessageRelayerRunner) CreateMessageTransaction(
	messageDoc *models.Message,
) bool {
	logger := x.logger.
		WithField("tx_hash", messageDoc.TransactionHash).
		WithField("section", "create-transaction")

	txStatus := models.TransactionStatusPending

	toAddress, err := common.BytesFromAddressHex(messageDoc.Content.MessageBody.RecipientAddress)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing recipient address")
		return false
	}

	txHash := common.Ensure0xPrefix(messageDoc.TransactionHash)

	tx, err := x.client.GetTx(txHash)
	if err != nil {
		logger.WithError(err).Errorf("Error getting tx")
		return false
	}

	transaction, err := db.NewCosmosTransaction(tx, x.chain, x.multisigPk.Address().Bytes(), toAddress, txStatus)
	if err != nil {
		x.logger.WithError(err).
			Errorf("Error creating transaction")
		return false
	}

	transaction.Messages = append(transaction.Messages, *messageDoc.ID)

	insertedID, err := db.InsertTransaction(transaction)
	if err != nil {
		x.logger.WithError(err).
			Errorf("Error inserting transaction")
		return false
	}

	return x.UpdateMessage(messageDoc.ID, bson.M{"transaction": insertedID})
}

func (x *MessageRelayerRunner) CreateRefundTransaction(
	refundDoc *models.Refund,
) bool {

	logger := x.logger.
		WithField("tx_hash", refundDoc.TransactionHash).
		WithField("section", "create-transaction")

	txStatus := models.TransactionStatusPending

	toAddress, err := common.BytesFromAddressHex(refundDoc.Recipient)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing recipient address")
		return false
	}

	tx, err := x.client.GetTx(refundDoc.TransactionHash)
	if err != nil {
		logger.WithError(err).Errorf("Error getting tx")
		return false
	}

	transaction, err := db.NewCosmosTransaction(tx, x.chain, x.multisigPk.Address().Bytes(), toAddress, txStatus)
	if err != nil {
		x.logger.WithError(err).
			Errorf("Error creating transaction")
		return false
	}

	transaction.Refund = refundDoc.ID
	insertedID, err := db.InsertTransaction(transaction)
	if err != nil {
		x.logger.WithError(err).
			Errorf("Error inserting transaction")
		return false
	}

	return x.UpdateRefund(refundDoc.ID, bson.M{"transaction": insertedID})
}

func (x *MessageRelayerRunner) SyncRefunds() bool {
	x.logger.Infof("Relaying refunds")
	refunds, err := db.GetBroadcastedRefunds()
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting broadcasted refunds")
		return false
	}
	x.logger.Infof("Found %d broadcasted refunds", len(refunds))
	success := true
	for _, refundDoc := range refunds {
		success = success && x.CreateRefundTransaction(&refundDoc)
	}

	return success
}

func (x *MessageRelayerRunner) SyncMessages() bool {
	x.logger.Infof("Relaying messages")
	messages, err := db.GetBroadcastedMessages(x.chain)
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting broadcasted messages")
		return false
	}
	x.logger.Infof("Found %d broadcasted messages", len(messages))
	success := true
	for _, messageDoc := range messages {
		success = success && x.CreateMessageTransaction(&messageDoc)
	}

	return success
}

func (x *MessageRelayerRunner) UpdateTransaction(
	tx *models.Transaction,
	update bson.M,
) bool {
	err := db.UpdateTransaction(tx.ID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating transaction")
		return false
	}
	return true
}

func (x *MessageRelayerRunner) ResetRefund(
	refundID *primitive.ObjectID,
) bool {
	if refundID == nil {
		x.logger.Errorf("Refund ID is nil")
		return false
	}
	update := bson.M{
		"status":           models.RefundStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}

	return x.UpdateRefund(refundID, update)
}

func (x *MessageRelayerRunner) ResetMessage(
	messageID *primitive.ObjectID,
) bool {
	if messageID == nil {
		x.logger.Errorf("Message ID is nil")
		return false
	}

	update := bson.M{
		"status":           models.MessageStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}

	return x.UpdateMessage(messageID, update)
}

func (x *MessageRelayerRunner) RelayTransactions() bool {
	x.logger.Infof("Relaying transactions")
	txs, err := db.GetPendingTransactionsFrom(x.chain, x.multisigPk.Address().Bytes())
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting pending txs")
		return false
	}
	x.logger.Infof("Found %d pending txs", len(txs))
	success := true
	for _, txDoc := range txs {
		logger := x.logger.WithField("tx_hash", txDoc.Hash).WithField("section", "confirm")

		if (txDoc.Refund == nil && len(txDoc.Messages) == 0) || (txDoc.Refund != nil && len(txDoc.Messages) != 0) {
			logger.Errorf("Invalid transaction")
			success = false
			continue
		}

		txResponse, err := x.client.GetTx(txDoc.Hash)
		if err != nil {
			logger.WithError(err).Errorf("Error getting tx")
			success = false
			continue
		}

		if txResponse.Code != 0 {
			logger.Infof("Found tx with error")
			updateSuccesful := x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusFailed})
			if !updateSuccesful {
				success = false
				continue
			}

			if txDoc.Refund != nil {
				success = success && x.ResetRefund(txDoc.Refund)
				continue
			}
			for _, messageID := range txDoc.Messages {
				success = success && x.ResetMessage(&messageID)
			}
			continue
		}
		logger.
			Debugf("Found pending tx")

		confirmations := x.currentBlockHeight - uint64(txResponse.Height)

		update := bson.M{
			"status":        models.TransactionStatusPending,
			"confirmations": confirmations,
		}

		if confirmations < x.config.Confirmations {
			success = success && x.UpdateTransaction(&txDoc, update)
			continue
		}

		update["status"] = models.TransactionStatusConfirmed
		confirmed := x.UpdateTransaction(&txDoc, update)

		// TODO: Handle updating refund and message status in a separate function ?
		if !confirmed {
			success = false
			continue
		}

		update = bson.M{
			"status":           models.MessageStatusSuccess,
			"transaction":      txDoc.ID,
			"transaction_hash": txDoc.Hash,
		}

		if txDoc.Refund != nil {
			success = x.UpdateRefund(txDoc.Refund, update) && success
			continue
		}
		for _, messageObjectID := range txDoc.Messages {
			success = x.UpdateMessage(&messageObjectID, update) && success
		}
	}

	return success
}

func (x *MessageRelayerRunner) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
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

func NewMessageRelayer(config models.CosmosNetworkConfig, lastHealth *models.RunnerServiceStatus) service.Runner {
	logger := log.
		WithField("module", "cosmos").
		WithField("service", "relayer").
		WithField("chain_name", strings.ToLower(config.ChainName)).
		WithField("chain_id", strings.ToLower(config.ChainID))

	if !config.MessageRelayer.Enabled {
		logger.Fatalf("Message relayer is not enabled")
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
	multisigAddress, err := common.Bech32FromBytes(config.Bech32Prefix, multisigPk.Address().Bytes())
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

	x := &MessageRelayerRunner{
		multisigPk: multisigPk,

		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		client:             client,

		chain:  util.ParseChain(config),
		config: config,

		logger: logger,
	}

	x.UpdateCurrentHeight()

	x.InitStartBlockHeight(lastHealth)

	logger.Infof("Initialized")

	return x
}
