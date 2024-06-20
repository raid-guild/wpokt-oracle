package ethereum

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type EthMessageRelayerRunnable struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	mintControllerMap map[uint32][]byte

	mintController eth.MintControllerContract
	client         eth.EthereumClient

	confirmations uint64

	chain models.Chain

	logger *log.Entry
}

func (x *EthMessageRelayerRunnable) Run() {
	x.UpdateCurrentBlockHeight()
	x.SyncNewBlocks()
	x.ConfirmFulfillmentTxs()
	x.ConfirmMessages()
}

func (x *EthMessageRelayerRunnable) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *EthMessageRelayerRunnable) UpdateCurrentBlockHeight() {
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

func (x *EthMessageRelayerRunnable) CreateTxForFulfillmentEvent(event *autogen.MintControllerFulfillment) bool {
	if event == nil {
		return false
	}

	txHash := event.Raw.TxHash.String()
	logger := x.logger.WithField("tx_hash", txHash).WithField("section", "CreateTxForFulfillmentEvent")

	result, err := ValidateTransactionByHash(x.client, txHash)
	if err != nil {
		logger.WithError(err).Errorf("Error validating transaction")
		return false
	}

	txDoc, err := db.NewEthereumTransaction(result.tx, x.mintController.Address().Bytes(), result.receipt, x.chain, models.TransactionStatusPending)
	if err != nil {
		logger.WithError(err).Errorf("Error creating transaction")
		return false
	}

	_, err = db.InsertTransaction(txDoc)
	if err != nil {
		logger.WithError(err).Errorf("Error inserting transaction")
		return false
	}

	return true
}

type ValidateTransactionAndParseFulfillmentEventsResult struct {
	Events        []*autogen.MintControllerFulfillment
	Confirmations uint64
	TxStatus      models.TransactionStatus
}

func (x *EthMessageRelayerRunnable) ValidateTransactionAndParseFulfillmentEvents(txHash string) (*ValidateTransactionAndParseFulfillmentEventsResult, error) {
	receipt, err := x.client.GetTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction receipt: %w", err)
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return &ValidateTransactionAndParseFulfillmentEventsResult{
			TxStatus: models.TransactionStatusFailed,
		}, nil
	}
	var events []*autogen.MintControllerFulfillment
	for _, log := range receipt.Logs {
		if log.Address == x.mintController.Address() {
			event, err := x.mintController.ParseFulfillment(*log)
			if err != nil {
				continue
			}
			events = append(events, event)
		}
	}
	result := &ValidateTransactionAndParseFulfillmentEventsResult{
		Events:        events,
		Confirmations: x.currentBlockHeight - receipt.BlockNumber.Uint64(),
		TxStatus:      models.TransactionStatusPending,
	}
	if result.Confirmations >= x.confirmations {
		result.TxStatus = models.TransactionStatusConfirmed
	}
	if len(events) == 0 {
		result.TxStatus = models.TransactionStatusInvalid
	}
	return result, nil
}

func (x *EthMessageRelayerRunnable) UpdateTransaction(
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

func (x *EthMessageRelayerRunnable) ConfirmTx(txDoc *models.Transaction) bool {
	if txDoc == nil {
		x.logger.Error("ConfirmTx: txDoc is nil")
		return false
	}

	result, err := x.ValidateTransactionAndParseFulfillmentEvents(txDoc.Hash)
	if err != nil {
		x.logger.WithError(err).Error("Error validating transaction and parsing dispatch events")
		return false
	}

	update := bson.M{
		"confirmations": result.Confirmations,
		"status":        result.TxStatus,
	}
	return x.UpdateTransaction(txDoc, update)
}

func (x *EthMessageRelayerRunnable) ConfirmMessagesForTx(txDoc *models.Transaction) bool {
	if txDoc == nil {
		x.logger.Error("ConfirmMessagesForTx: txDoc is nil")
		return false
	}

	logger := x.logger.WithField("tx_hash", txDoc.Hash).WithField("section", "ConfirmMessagesForTx")

	result, err := x.ValidateTransactionAndParseFulfillmentEvents(txDoc.Hash)
	if err != nil {
		logger.WithError(err).Error("Error validating transaction and parsing dispatch events")
		return false
	}

	if result.TxStatus != models.TransactionStatusConfirmed {
		logger.Errorf("Transaction not confirmed")
		return x.UpdateTransaction(txDoc, bson.M{"status": result.TxStatus})
	}

	if lockID, err := db.LockWriteTransaction(txDoc); err != nil {
		logger.WithError(err).Error("Error locking transaction")
		return false
	} else {
		defer db.Unlock(lockID)
	}

	update := bson.M{
		"status":           models.MessageStatusSuccess,
		"transaction":      txDoc.ID,
		"transaction_hash": txDoc.Hash,
	}

	success := true

	for _, event := range result.Events {
		docID, err := db.UpdateMessageByMessageID(event.OrderId, update)
		if err != nil {
			logger.WithError(err).Errorf("Error updating message")
			success = false
			continue
		}
		txDoc.Messages = append(txDoc.Messages, docID)
	}

	if !success {
		return false
	}

	return x.UpdateTransaction(txDoc, bson.M{"messages": common.RemoveDuplicates(txDoc.Messages)})
}

func (x *EthMessageRelayerRunnable) SyncBlocks(startBlockHeight uint64, endBlockHeight uint64) bool {
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

		success = x.CreateTxForFulfillmentEvent(event) && success
	}

	if err := filter.Error(); err != nil {
		x.logger.WithError(err).Error("Error processing fulfillment events")
		return false
	}

	return success
}

func (x *EthMessageRelayerRunnable) SyncNewBlocks() bool {
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

func (x *EthMessageRelayerRunnable) ConfirmFulfillmentTxs() bool {
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

func (x *EthMessageRelayerRunnable) ConfirmMessages() bool {
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

func (x *EthMessageRelayerRunnable) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
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
) service.Runnable {
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

	x := &EthMessageRelayerRunnable{
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
