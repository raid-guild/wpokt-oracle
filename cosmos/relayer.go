package cosmos

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	log "github.com/sirupsen/logrus"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type MessageRelayerRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	multisigAddress   string
	multisigThreshold uint64
	multisigPk        *multisig.LegacyAminoPubKey

	bech32Prefix string
	coinDenom    string

	confirmations uint64

	chain  models.Chain
	client cosmos.CosmosClient

	logger *log.Entry
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

func (x *MessageRelayerRunner) SyncMessages() bool {
	return true
}

func (x *MessageRelayerRunner) UpdateRefund(
	refundID *primitive.ObjectID,
	update bson.M,
) bool {
	err := util.UpdateRefund(refundID, update)
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
	err := util.UpdateMessage(messageID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating message")
		return false
	}
	return true
}

func (x *MessageRelayerRunner) CreateTransaction(
	refundDoc *models.Refund,
) bool {

	logger := x.logger.
		WithField("tx_hash", refundDoc.TransactionHash).
		WithField("section", "create-transaction")

	txStatus := models.TransactionStatusPending

	toAddress, err := util.BytesFromHex(refundDoc.Recipient)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing recipient address")
		return false
	}

	txHash := util.Ensure0xPrefix(refundDoc.TransactionHash)

	tx, err := x.client.GetTx(txHash)
	if err != nil {
		logger.WithError(err).Errorf("Error getting tx")
		return false
	}

	transaction, err := util.NewTransaction(tx, x.chain, x.multisigPk.Address().Bytes(), toAddress, txStatus)
	if err != nil {
		x.logger.WithError(err).
			Errorf("Error creating transaction")
		return false
	}

	transaction.Refund = refundDoc.ID
	insertedID, err := util.InsertTransaction(transaction)
	if err != nil {
		x.logger.WithError(err).
			Errorf("Error inserting transaction")
		return false
	}

	return x.UpdateRefund(refundDoc.ID, bson.M{"transaction": insertedID})
}

func (x *MessageRelayerRunner) SyncRefunds() bool {
	x.logger.Infof("Relaying refunds")
	refunds, err := util.GetBroadcastedRefunds()
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting broadcasted refunds")
		return false
	}
	x.logger.Infof("Found %d broadcasted refunds", len(refunds))
	success := true
	for _, refundDoc := range refunds {
		success = success && x.CreateTransaction(&refundDoc)
	}

	if success {
		x.startBlockHeight = x.currentBlockHeight
	}

	return success
}

func (x *MessageRelayerRunner) UpdateTransaction(
	tx *models.Transaction,
	update bson.M,
) bool {
	err := util.UpdateTransaction(tx, update)
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

	return true
	// update := bson.M{
	// 	"status": models.MessageStatusPending,
	// 	"signatures": []models.Signature{},
	// 	"transaction": nil,
	// 	"transaction_hash": "",
	// }
	//
	// return x.UpdateMessage(messageID, update)
}

func (x *MessageRelayerRunner) RelayTransactions() bool {
	x.logger.Infof("Relaying transactions")
	txs, err := util.GetPendingTransactionsFrom(x.chain, x.multisigPk.Address().Bytes())
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting pending txs")
		return false
	}
	x.logger.Infof("Found %d pending txs", len(txs))
	success := true
	for _, txDoc := range txs {
		logger := x.logger.WithField("tx_hash", txDoc.Hash).WithField("section", "confirm")
		txResponse, err := x.client.GetTx(txDoc.Hash)
		if err != nil {
			logger.WithError(err).Errorf("Error getting tx")
			success = false
			continue
		}
		if txResponse.Code != 0 {
			logger.Infof("Found tx with error")
			failed := x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusFailed})
			if failed {
				if txDoc.Refund != nil {
					success = success && x.ResetRefund(txDoc.Refund)
				} else if txDoc.Message != nil {
					success = success && x.ResetMessage(txDoc.Message)
				} else {
					success = false
				}
			} else {
				success = false
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

		if confirmations < x.confirmations {
			success = success && x.UpdateTransaction(&txDoc, update)
			continue
		}

		update["status"] = models.TransactionStatusConfirmed
		confirmed := x.UpdateTransaction(&txDoc, update)

		if confirmed {
			if txDoc.Refund != nil {
				success = success && x.UpdateRefund(txDoc.Refund, bson.M{"status": models.RefundStatusSuccess})
			} else if txDoc.Message != nil {
				success = success && x.UpdateMessage(txDoc.Message, bson.M{"status": models.MessageStatusSuccess})
			} else {
				success = false
			}
		} else {
			success = false
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

	x := &MessageRelayerRunner{
		multisigPk:        multisigPk,
		multisigThreshold: config.MultisigThreshold,
		multisigAddress:   multisigAddress,

		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		client:             client,

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
