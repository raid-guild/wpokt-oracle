package ethereum

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type EthMessageMonitorRunnable struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	mintControllerMap map[uint32][]byte

	mailbox eth.MailboxContract
	client  eth.EthereumClient

	confirmations uint64

	chain models.Chain

	logger *log.Entry

	db db.DB
}

func (x *EthMessageMonitorRunnable) Run() {
	x.UpdateCurrentBlockHeight()
	x.SyncNewBlocks()
	x.ConfirmDispatchTxs()
	x.CreateMessagesForTxs()
}

func (x *EthMessageMonitorRunnable) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *EthMessageMonitorRunnable) UpdateCurrentBlockHeight() {
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

func (x *EthMessageMonitorRunnable) UpdateTransaction(
	tx *models.Transaction,
	update bson.M,
) bool {
	err := x.db.UpdateTransaction(tx.ID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating transaction")
		return false
	}
	return true
}

func (x *EthMessageMonitorRunnable) IsValidEvent(event *autogen.MailboxDispatch) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	{ // validate sender

		mintController, ok := x.mintControllerMap[x.chain.ChainDomain]
		if !ok {
			return fmt.Errorf("mint controller not found for chain domain: %d", x.chain.ChainDomain)
		}

		if !bytes.Equal(event.Sender.Bytes(), mintController) {
			return fmt.Errorf("sender does not match mint controller for chain domain: %d", x.chain.ChainDomain)
		}

	}

	{ // validate recipient

		destMintController, ok := x.mintControllerMap[event.Destination]
		if !ok {
			return fmt.Errorf("mint controller not found for destination domain: %d", event.Destination)
		}

		if !bytes.Equal(destMintController, []byte(event.Recipient[12:32])) {
			return fmt.Errorf("recipient does not match mint controller for destination domain: %d", event.Destination)
		}

	}

	{ // validate message content

		var messageContent models.MessageContent

		err := messageContent.DecodeFromBytes(event.Message)
		if err != nil {
			return fmt.Errorf("error decoding message content: %w", err)
		}

		if messageContent.OriginDomain != x.chain.ChainDomain {
			return fmt.Errorf("invalid origin domain: %d", x.chain.ChainDomain)
		}

		if messageContent.DestinationDomain != event.Destination {
			return fmt.Errorf("invalid destination domain: %d", event.Destination)
		}

		if messageContent.Version != common.HyperlaneVersion {
			return fmt.Errorf("invalid version: %d", messageContent.Version)
		}

		if !strings.EqualFold(messageContent.Recipient, common.HexFromBytes(event.Recipient[12:32])) {
			return fmt.Errorf("invalid recipient: %s", messageContent.Recipient)
		}

		if !strings.EqualFold(messageContent.Sender, common.HexFromBytes(event.Sender[:])) {
			return fmt.Errorf("invalid sender: %s", messageContent.Sender)
		}

	}

	return nil
}

func (x *EthMessageMonitorRunnable) CreateTxForDispatchEvent(event *autogen.MailboxDispatch) bool {
	txHash := event.Raw.TxHash.String()
	logger := x.logger.WithField("tx_hash", txHash).WithField("section", "CreateTxForDispatchEvent")

	if err := x.IsValidEvent(event); err != nil {
		logger.WithError(err).Errorf("Invalid event")
		return false
	}

	result, err := ethValidateTransactionByHash(x.client, txHash)
	if err != nil {
		logger.WithError(err).Errorf("Error validating transaction")
		return false
	}

	txDoc, err := x.db.NewEthereumTransaction(
		result.Tx,
		x.mailbox.Address().Bytes(),
		result.Receipt,
		x.chain,
		models.TransactionStatusPending,
	)
	if err != nil {
		x.logger.WithError(err).Errorf("Error creating transaction")
		return false
	}

	_, err = x.db.InsertTransaction(txDoc)
	if err != nil {
		x.logger.WithError(err).Errorf("Error inserting transaction")
		return false
	}

	return true
}

type ValidateTransactionAndParseDispatchEventsResult struct {
	Events        []*autogen.MailboxDispatch
	Confirmations uint64
	TxStatus      models.TransactionStatus
}

func (x *EthMessageMonitorRunnable) ValidateTransactionAndParseDispatchEvents(txHash string) (*ValidateTransactionAndParseDispatchEventsResult, error) {
	receipt, err := x.client.GetTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction receipt: %w", err)
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return &ValidateTransactionAndParseDispatchEventsResult{
			TxStatus: models.TransactionStatusFailed,
		}, nil
	}
	var events []*autogen.MailboxDispatch
	for _, log := range receipt.Logs {
		if log.Address == x.mailbox.Address() {
			event, err := x.mailbox.ParseDispatch(*log)
			if err != nil {
				continue
			}
			events = append(events, event)
		}
	}
	result := &ValidateTransactionAndParseDispatchEventsResult{
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

func (x *EthMessageMonitorRunnable) ConfirmTx(txDoc *models.Transaction) bool {
	if txDoc == nil {
		x.logger.Error("ConfirmTx: txDoc is nil")
		return false
	}

	result, err := x.ValidateTransactionAndParseDispatchEvents(txDoc.Hash)
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

func (x *EthMessageMonitorRunnable) CreateMessagesForTx(txDoc *models.Transaction) bool {
	if txDoc == nil {
		x.logger.Error("CreateMessagesForTx: txDoc is nil")
		return false
	}

	logger := x.logger.WithField("tx_hash", txDoc.Hash).WithField("section", "CreateMessagesForTx")

	result, err := x.ValidateTransactionAndParseDispatchEvents(txDoc.Hash)

	if err != nil {
		logger.WithError(err).Error("Error validating transaction and parsing dispatch events")
		return false
	}

	if result.TxStatus != models.TransactionStatusConfirmed {
		logger.Errorf("Transaction not confirmed")
		x.UpdateTransaction(txDoc, bson.M{"status": result.TxStatus})
		return false
	}

	if lockID, err := x.db.LockWriteTransaction(txDoc); err != nil {
		logger.WithError(err).Error("Error locking transaction")
		return false
	} else {
		//nolint:errcheck
		defer x.db.Unlock(lockID)
	}

	success := true

	for _, event := range result.Events {
		var messageContent models.MessageContent

		//nolint:errcheck
		messageContent.DecodeFromBytes(event.Message) // event was validated already

		message, err := x.db.NewMessage(txDoc, messageContent, models.MessageStatusPending)
		if err != nil {
			logger.WithError(err).Errorf("Error creating message")
			success = false
			continue
		}

		messageID, err := x.db.InsertMessage(message)
		if err != nil {
			logger.WithError(err).Errorf("Error inserting message")
			success = false
			continue
		}

		logger.WithField("message_id", messageID.Hex()).Info("Message created")

		txDoc.Messages = append(txDoc.Messages, messageID)
	}

	if !success {
		return false
	}

	return x.UpdateTransaction(txDoc, bson.M{"messages": common.RemoveDuplicates(txDoc.Messages)})
}

func (x *EthMessageMonitorRunnable) SyncBlocks(startBlockHeight uint64, endBlockHeight uint64) bool {
	mintController, ok := x.mintControllerMap[x.chain.ChainDomain]
	if !ok {
		x.logger.Errorf("Mint controller not found for chain domain: %d", x.chain.ChainDomain)
		return false
	}

	mintControllerAddress := ethcommon.BytesToAddress(mintController)

	filter, err := x.mailbox.FilterDispatch(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, []ethcommon.Address{mintControllerAddress}, []uint32{}, [][32]byte{})

	if filter != nil {
		defer filter.Close()
	}

	if err != nil {
		x.logger.WithError(err).Error("Error creating filter for dispatch events")
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

		success = x.CreateTxForDispatchEvent(event) && success
	}

	if err := filter.Error(); err != nil {
		x.logger.WithError(err).Error("Error processing dispatch events")
		return false
	}

	return success
}

func (x *EthMessageMonitorRunnable) SyncNewBlocks() bool {
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

func (x *EthMessageMonitorRunnable) ConfirmDispatchTxs() bool {
	logger := x.logger.WithField("section", "ConfirmDispatchTxs")

	txs, err := x.db.GetPendingTransactionsTo(x.chain, x.mailbox.Address().Bytes())
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

func (x *EthMessageMonitorRunnable) CreateMessagesForTxs() bool {
	logger := x.logger.WithField("section", "CreateMessagesForTxs")

	txs, err := x.db.GetConfirmedTransactionsTo(x.chain, x.mailbox.Address().Bytes())
	if err != nil {
		logger.WithError(err).Error("Error getting confirmed transactions")
		return false
	}

	success := true
	for _, tx := range txs {
		success = x.CreateMessagesForTx(&tx) && success
	}

	return success
}

func (x *EthMessageMonitorRunnable) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
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

func NewMessageMonitor(
	config models.EthereumNetworkConfig,
	mintControllerMap map[uint32][]byte,
	lastHealth *models.RunnerServiceStatus,
) service.Runnable {
	logger := log.
		WithField("module", "ethereum").
		WithField("service", "monitor").
		WithField("chain_name", strings.ToLower(config.ChainName)).
		WithField("chain_id", config.ChainID)

	if !config.MessageMonitor.Enabled {
		logger.Fatalf("Message monitor is not enabled")
	}

	logger.Debugf("Initializing")

	client, err := eth.NewClient(config)
	if err != nil {
		logger.Fatalf("Error creating ethereum client: %s", err)
	}

	logger.Debug("Connecting to mailbox contract at: ", config.MailboxAddress)
	mailbox, err := eth.NewMailboxContract(common.HexToAddress(config.MailboxAddress), client.GetClient())
	if err != nil {
		logger.Fatal("Error connecting to mailbox contract: ", err)
	}
	logger.Debug("Connected to mailbox contract")

	x := &EthMessageMonitorRunnable{
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,

		mintControllerMap: mintControllerMap,

		mailbox:       mailbox,
		confirmations: config.Confirmations,

		client: client,

		chain: utilParseChain(config),

		logger: logger,

		db: db.NewDB(),
	}

	x.UpdateCurrentBlockHeight()

	x.InitStartBlockHeight(lastHealth)

	logger.Infof("Initialized")

	return x
}
