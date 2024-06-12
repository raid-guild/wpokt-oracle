package ethereum

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type MessageRelayerRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	mintControllerMap map[uint32][]byte

	mintController eth.MintControllerContract
	client         eth.EthereumClient

	confirmations uint64

	chain models.Chain

	logger *log.Entry
}

func (x *MessageRelayerRunner) Run() {
	x.UpdateCurrentBlockHeight()
	x.SyncNewBlocks()
	x.ConfirmFulfillmentTxs()
	x.ConfirmMessages()
}

func (x *MessageRelayerRunner) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *MessageRelayerRunner) UpdateCurrentBlockHeight() {
	res, err := x.client.GetBlockHeight()
	if err != nil {
		x.logger.
			WithError(err).
			Error("could not get current block height")
		return
	}
	x.currentBlockHeight = res
	x.logger.
		WithField("current_block_height", x.currentBlockHeight).
		Info("updated current block height")
}

func (x *MessageRelayerRunner) HandleFulfillmentEvent(event *autogen.MintControllerFulfillment) bool {
	if event == nil {
		x.logger.Error("HandleFulfillmentEvent: event is nil")
		return false
	}

	var messageContent models.MessageContent

	err := messageContent.DecodeFromBytes(event.Message)
	if err != nil {
		x.logger.WithError(err).Error("Error decoding message content")
		return false
	}

	txHash := event.Raw.TxHash.String()

	tx, isPending, err := x.client.GetTransactionByHash(txHash)
	if err != nil {
		x.logger.WithError(err).Error("Error getting transaction by hash")
		return false
	}
	if tx == nil {
		x.logger.Errorf("Transaction not found: %s", txHash)
		return false
	}
	if isPending {
		x.logger.Infof("Transaction is pending")
		return false
	}
	receipt, err := x.client.GetTransactionReceipt(txHash)
	if err != nil {
		x.logger.WithError(err).Error("Error getting transaction receipt")
		return false
	}

	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		x.logger.Infof("Transaction failed")
		return false
	}

	txDoc, err := db.NewEthereumTransaction(tx, x.mintController.Address().Bytes(), receipt, x.chain, models.TransactionStatusPending)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_hash", txHash).
			Errorf("Error creating transaction")
		return false
	}

	insertedID, err := db.InsertTransaction(txDoc)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_hash", txHash).
			Errorf("Error inserting transaction")
		return false
	}

	update := bson.M{
		"transaction":      insertedID,
		"transaction_hash": txHash,
	}

	messageID := event.OrderId

	_, err = db.UpdateMessageByMessageID(messageID, update)
	if err != nil {
		x.logger.WithError(err).
			Errorf("Error updating message")
		return false
	}

	return true
}

func (x *MessageRelayerRunner) ConfirmTx(txDoc *models.Transaction) bool {
	if txDoc == nil {
		x.logger.Error("ConfirmTx: txDoc is nil")
		return false
	}

	receipt, err := x.client.GetTransactionReceipt(txDoc.Hash)
	if err != nil {
		x.logger.WithError(err).Error("Error getting transaction receipt")
		return false
	}

	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		x.logger.Infof("Transaction failed")
		return false
	}

	confirmations := x.currentBlockHeight - txDoc.BlockHeight
	if confirmations < x.confirmations {
		x.logger.Infof("Transaction has not enough confirmations: %d", confirmations)
		return false
	}

	update := bson.M{
		"confirmations": confirmations,
		"status":        models.TransactionStatusConfirmed,
	}

	err = db.UpdateTransaction(txDoc.ID, update)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_hash", txDoc.Hash).
			Errorf("Error updating transaction")
		return false
	}

	return true
}

func (x *MessageRelayerRunner) ConfirmMessagesForTx(txDoc *models.Transaction) bool {
	if txDoc == nil {
		x.logger.Error("ConfirmMessagesForTx: txDoc is nil")
		return false
	}

	logger := x.logger.WithField("tx_hash", txDoc.Hash).WithField("section", "ConfirmMessagesForTx")

	receipt, err := x.client.GetTransactionReceipt(txDoc.Hash)
	if err != nil {
		x.logger.WithError(err).Error("Error getting transaction receipt")
		return false
	}

	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		x.logger.Infof("Transaction failed")
		return false
	}

	confirmations := x.currentBlockHeight - txDoc.BlockHeight
	if confirmations < x.confirmations {
		x.logger.Infof("Transaction has not enough confirmations: %d", confirmations)
		return false
	}

	var messageIDs [][32]byte
	for _, log := range receipt.Logs {
		if log.Address == x.mintController.Address() {
			event, err := x.mintController.ParseFulfillment(*log)
			if err != nil {
				logger.WithError(err).Errorf("Error parsing fulfillment event")
				continue
			}
			messageIDs = append(messageIDs, event.OrderId)
			break
		}
	}

	if len(messageIDs) == 0 {
		return false
	}

	update := bson.M{
		"status":           models.MessageStatusSuccess,
		"transaction":      txDoc.ID,
		"transaction_hash": txDoc.Hash,
	}

	var docIDs []primitive.ObjectID

	success := true

	for _, messageID := range messageIDs {
		docID, err := db.UpdateMessageByMessageID(messageID, update)
		if err != nil {
			logger.WithError(err).
				Errorf("Error updating message")
			success = false
		} else {
			docIDs = append(docIDs, docID)
		}
	}

	if !success {
		return false
	}

	update = bson.M{
		"messages": docIDs,
	}

	err = db.UpdateTransaction(txDoc.ID, update)
	if err != nil {
		logger.WithError(err).
			Errorf("Error updating transaction")
		return false
	}

	return true
}

func (x *MessageRelayerRunner) SyncBlocks(startBlockHeight uint64, endBlockHeight uint64) bool {
	filter, err := x.mintController.FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, [][32]byte{})

	if filter != nil {
		defer filter.Close()
	}

	if err != nil {
		x.logger.WithError(err).Error("Error creating filter for fulfillment events")
		return false
	}

	success := true
	for filter.Next() {
		if err := filter.Error(); err != nil {
			success = false
			break
		}

		event := filter.Event()

		if event == nil {
			success = false
			continue
		}

		if event.Raw.Removed {
			continue
		}

		success = x.HandleFulfillmentEvent(event) && success
	}

	if err := filter.Error(); err != nil {
		x.logger.WithError(err).Error("Error processing fulfillment events")
		return false
	}

	return success
}

func (x *MessageRelayerRunner) SyncNewBlocks() bool {
	if x.currentBlockHeight <= x.startBlockHeight {
		x.logger.Infof("No new blocks to sync")
		return true
	}

	success := true
	if (x.currentBlockHeight - x.startBlockHeight) > eth.MaxQueryBlocks {
		x.logger.Debug("Syncing blocks in chunks")
		for i := x.startBlockHeight; i < x.currentBlockHeight; i += eth.MaxQueryBlocks {
			endBlockHeight := i + eth.MaxQueryBlocks
			if endBlockHeight > x.currentBlockHeight {
				endBlockHeight = x.currentBlockHeight
			}
			x.logger.Info("Syncing blocks from blockNumber: ", i, " to blockNumber: ", endBlockHeight)
			success = x.SyncBlocks(uint64(i), uint64(endBlockHeight)) && success
		}
	} else {
		x.logger.Info("Syncing blocks from blockNumber: ", x.startBlockHeight, " to blockNumber: ", x.currentBlockHeight)
		success = x.SyncBlocks(uint64(x.startBlockHeight), uint64(x.currentBlockHeight)) && success
	}

	if success {
		x.startBlockHeight = x.currentBlockHeight
	}

	return success
}

func (x *MessageRelayerRunner) ConfirmFulfillmentTxs() bool {
	logger := x.logger.WithField("section", "ConfirmFulfillmentTxs")

	txs, err := db.GetPendingTransactionsTo(x.chain, x.mintController.Address().Bytes())
	if err != nil {
		logger.WithError(err).Error("Error getting pending transactions")
		return false
	}

	success := true
	for _, tx := range txs {
		success = x.ConfirmTx(&tx) && success
	}

	return success
}

func (x *MessageRelayerRunner) ConfirmMessages() bool {
	logger := x.logger.WithField("section", "ConfirmMessages")

	txs, err := db.GetConfirmedTransactionsTo(x.chain, x.mintController.Address().Bytes())
	if err != nil {
		logger.WithError(err).Error("Error getting confirmed transactions")
		return false
	}

	success := true
	for _, tx := range txs {
		success = x.ConfirmMessagesForTx(&tx) && success
	}

	return success
}

func (x *MessageRelayerRunner) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
	if lastHealth == nil || lastHealth.BlockHeight == 0 {
		x.logger.Infof("Invalid last health")
	} else {
		x.logger.Debugf("Last block height: %d", lastHealth.BlockHeight)
		x.startBlockHeight = lastHealth.BlockHeight
	}
	if x.startBlockHeight == 0 || x.startBlockHeight > x.currentBlockHeight {
		x.logger.Infof("Start block height is greater than current block height")
		x.startBlockHeight = x.currentBlockHeight
	}
	x.logger.Infof("Initialized start block height: %d", x.startBlockHeight)
}

func NewMessageRelayer(
	config models.EthereumNetworkConfig,
	mintControllerMap map[uint32][]byte,
	lastHealth *models.RunnerServiceStatus,
) service.Runner {
	logger := log.
		WithField("module", "ethereum").
		WithField("service", "relayer").
		WithField("chain_name", strings.ToLower(config.ChainName)).
		WithField("chain_id", config.ChainID)

	if !config.MessageRelayer.Enabled {
		logger.Fatalf("Message relayer is not enabled")
	}

	logger.Debugf("Initializing")

	client, err := eth.NewClient(config)
	if err != nil {
		logger.Fatalf("Error creating ethereum client: %s", err)
	}

	logger.Debug("Connecting to mintController contract at: ", config.MintControllerAddress)
	mintController, err := eth.NewMintControllerContract(common.HexToAddress(config.MintControllerAddress), client.GetClient())
	if err != nil {
		logger.Fatal("Error connecting to mintController contract: ", err)
	}
	logger.Debug("Connected to mintController contract")

	x := &MessageRelayerRunner{
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,

		mintControllerMap: mintControllerMap,

		mintController: mintController,
		confirmations:  config.Confirmations,

		client: client,

		chain: util.ParseChain(config),

		logger: logger,
	}

	x.UpdateCurrentBlockHeight()

	x.InitStartBlockHeight(lastHealth)

	logger.Infof("Initialized")

	return x
}
