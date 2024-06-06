package ethereum

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type MessageMonitorRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	mintControllerMap map[uint32][]byte

	mailbox eth.MailboxContract
	client  eth.EthereumClient

	confirmations uint64

	chain models.Chain

	minimumAmount *big.Int

	logger *log.Entry
}

func (x *MessageMonitorRunner) Run() {
	x.UpdateCurrentBlockHeight()
	x.SyncNewBlocks()
}

func (x *MessageMonitorRunner) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *MessageMonitorRunner) UpdateCurrentBlockHeight() {
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

func (x *MessageMonitorRunner) HandleDispatchEvent(event *autogen.MailboxDispatch) bool {
	if event == nil {
		x.logger.Error("HandleDispatchEvent: event is nil")
		return false
	}

	mintController, ok := x.mintControllerMap[event.Destination]
	if !ok {
		x.logger.Errorf("Mint controller not found for destination domain: %d", event.Destination)
		return false
	}

	if !bytes.Equal(mintController, []byte(event.Recipient[12:32])) {
		x.logger.Errorf("Recipient does not match mint controller for destination domain: %d", event.Destination)
		return false
	}

	var messageContent models.MessageContent

	err := messageContent.DecodeFromBytes(event.Message)
	if err != nil {
		x.logger.WithError(err).Error("Error decoding message content")
		return false
	}

	if messageContent.DestinationDomain != event.Destination {
		x.logger.Errorf("Destination domain does not match message content destination domain: %d", event.Destination)
		return false
	}

	recipientHex := "0x" + hex.EncodeToString(event.Recipient[12:32])

	if !strings.EqualFold(messageContent.Recipient, recipientHex) {
		x.logger.Errorf("Recipient does not match message content recipient: %s", recipientHex)
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

	txDoc, err := db.NewEthereumTransaction(tx, receipt, x.chain, models.TransactionStatusPending)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_hash", txHash).
			Errorf("Error creating transaction")
		return false
	}

	confirmations := x.currentBlockHeight - txDoc.BlockHeight
	if confirmations < x.confirmations {
		x.logger.Infof("Transaction has not enough confirmations: %d", confirmations)
		return false
	}

	txDoc.Confirmations = confirmations
	txDoc.Status = models.TransactionStatusConfirmed

	txID, err := db.InsertTransaction(txDoc)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_hash", txHash).
			Errorf("Error inserting transaction")
		return false
	}
	txDoc.ID = &txID

	message, err := db.NewMessage(&txDoc, messageContent, models.MessageStatusPending)
	if err != nil {
		x.logger.WithError(err).Errorf("Error creating message")
		return false
	}

	messageID, err := db.InsertMessage(message)
	if err != nil {
		x.logger.WithError(err).Errorf("Error inserting message")
		return false
	}

	x.logger.
		WithField("tx_hash", txHash).
		WithField("message_id", messageID.Hex()).
		Info("Message created")

	update := bson.M{
		"message": messageID,
	}

	err = db.UpdateTransaction(&txDoc, update)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_hash", txHash).
			Errorf("Error updating transaction")
		return false
	}

	return true
}

func (x *MessageMonitorRunner) SyncBlocks(startBlockHeight uint64, endBlockHeight uint64) bool {
	filter, err := x.mailbox.FilterDispatch(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, []common.Address{}, []uint32{}, [][32]byte{})

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

		success = x.HandleDispatchEvent(event) && success
	}

	if err := filter.Error(); err != nil {
		x.logger.Error("[BURN MONITOR] Error while syncing burn events: ", err)
		return false
	}

	return success
}

func (x *MessageMonitorRunner) SyncNewBlocks() bool {
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
			success = success && x.SyncBlocks(uint64(i), uint64(endBlockHeight))
		}
	} else {
		x.logger.Info("Syncing blocks from blockNumber: ", x.startBlockHeight, " to blockNumber: ", x.currentBlockHeight)
		success = success && x.SyncBlocks(uint64(x.startBlockHeight), uint64(x.currentBlockHeight))
	}

	if success {
		x.startBlockHeight = x.currentBlockHeight
	}

	return success
}

func (x *MessageMonitorRunner) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
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

func NewMessageMonitor(config models.EthereumNetworkConfig, mintControllerMap map[uint32][]byte, lastHealth *models.RunnerServiceStatus) service.Runner {
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

	x := &MessageMonitorRunner{
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,

		mintControllerMap: mintControllerMap,

		mailbox:       mailbox,
		confirmations: config.Confirmations,

		client: client,

		chain: util.ParseChain(config),

		logger: logger,
	}

	x.UpdateCurrentBlockHeight()

	x.InitStartBlockHeight(lastHealth)

	logger.Infof("Initialized")

	return x
}
