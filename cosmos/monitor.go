package cosmos

import (
	"bytes"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"go.mongodb.org/mongo-driver/bson"

	log "github.com/sirupsen/logrus"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"
)

type MessageMonitorRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	multisigAddress string
	multisigPk      *multisig.LegacyAminoPubKey

	bech32Prefix string
	coinDenom    string
	feeAmount    sdk.Coin

	confirmations uint64

	chain  models.Chain
	client cosmos.CosmosClient

	logger *log.Entry
}

func (x *MessageMonitorRunner) Run() {
	x.UpdateCurrentHeight()
	x.SyncNewTxs()
	x.ConfirmTxs()
}

func (x *MessageMonitorRunner) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *MessageMonitorRunner) UpdateCurrentHeight() {
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

func (x *MessageMonitorRunner) CreateTransactionWithSpender(
	tx *sdk.TxResponse,
	txStatus models.TransactionStatus,
	coinsSpender string,
) bool {

	sender, err := util.ParseMessageSenderEvent(tx.Events)
	if err != nil {
		x.logger.WithError(err).Errorf("Error parsing message sender")
		return false
	}
	senderAddress, err := util.BytesFromBech32(x.bech32Prefix, sender)
	if err != nil {
		x.logger.WithError(err).Errorf("Error parsing sender address")
		return false
	}

	if coinsSpender != "" {
		var spenderAddress []byte
		spenderAddress, err = util.BytesFromBech32(x.bech32Prefix, coinsSpender)
		if err != nil {
			x.logger.WithError(err).Errorf("Error parsing spender address")
			return false
		}
		if !bytes.Equal(senderAddress, spenderAddress) {
			x.logger.Errorf("Sender address does not match spender address")
			txStatus = models.TransactionStatusInvalid
		}
	}

	transaction, err := util.NewTransaction(tx, x.chain, senderAddress, x.multisigPk.Address().Bytes(), txStatus)
	if err != nil {
		x.logger.WithError(err).
			WithField("status", txStatus).
			WithField("tx_hash", tx.TxHash).
			Errorf("Error creating transaction")
		return false
	}
	_, err = util.InsertTransaction(transaction)
	if err != nil {
		x.logger.WithError(err).
			WithField("status", txStatus).
			WithField("tx_hash", tx.TxHash).
			Errorf("Error inserting transaction")
		return false
	}
	return true
}

func (x *MessageMonitorRunner) CreateTransaction(
	tx *sdk.TxResponse,
	txStatus models.TransactionStatus,
) bool {
	return x.CreateTransactionWithSpender(tx, txStatus, "")
}

func (x *MessageMonitorRunner) UpdateTransaction(
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

func (x *MessageMonitorRunner) CreateRefund(
	txRes *sdk.TxResponse,
	txDoc *models.Transaction,
	spender string,
	amount sdk.Coin,
) bool {

	toAddr, err := util.BytesFromBech32(x.bech32Prefix, spender)
	if err != nil {
		x.logger.WithError(err).Errorf("Error parsing spender address")
		return false
	}

	txBody, err := util.NewSendTx(
		x.bech32Prefix,
		x.multisigPk.Address().Bytes(),
		toAddr,
		amount,
		"Refund for "+txRes.TxHash,
		x.feeAmount,
	)

	if err != nil {
		x.logger.WithError(err).Errorf("Error creating tx body")
		return false
	}

	refund, err := util.NewRefund(txRes, txDoc, toAddr, amount, string(txBody))

	if err != nil {
		x.logger.WithError(err).Errorf("Error creating refund")
		return false
	}

	insertedID, err := util.InsertRefund(refund)
	if err != nil {
		x.logger.WithError(err).Errorf("Error inserting refund")
		return false
	}

	err = util.UpdateTransaction(txDoc, bson.M{"refund": insertedID})
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating transaction")
		return false
	}

	return true
}

func (x *MessageMonitorRunner) CreateMessage(
	tx *sdk.TxResponse,
	txDoc *models.Transaction,
	spender string,
	amountCoin sdk.Coin,
	memo models.MintMemo,
) bool {
	// senderAddr, err := util.BytesFromBech32(x.bech32Prefix, spender)
	// if err != nil {
	// 	x.logger.WithError(err).Errorf("Error parsing spender address")
	// 	return false
	// }
	// recipientAddr, err := util.BytesFromHex(memo.Address)
	// if err != nil {
	// 	x.logger.WithError(err).Errorf("Error parsing recipient address")
	// 	return false
	// }
	// amount := amountCoin.Amount.Uint64()

	return true
}

func (x *MessageMonitorRunner) SyncNewTxs() bool {
	x.logger.Infof("Syncing new txs")
	if x.currentBlockHeight <= x.startBlockHeight {
		x.logger.Infof("No new blocks to sync")
		return true
	}

	txResponses, err := x.client.GetTxsSentToAddressAfterHeight(x.multisigAddress, x.startBlockHeight)
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting new txs")
		return false
	}
	x.logger.Infof("Found %d txs to sync", len(txResponses))
	success := true
	for _, txResponse := range txResponses {
		logger := x.logger.WithField("tx_hash", txResponse.TxHash).WithField("section", "sync")

		if txResponse.Code != 0 {
			logger.Infof("Found tx with non-zero code")
			success = success && x.CreateTransaction(txResponse, models.TransactionStatusFailed)
			continue
		}
		logger.Debugf("Found successful tx")

		tx := &tx.Tx{}
		err = tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			logger.WithError(err).Errorf("Error unmarshalling tx")
			success = success && x.CreateTransaction(txResponse, models.TransactionStatusInvalid)
			continue
		}

		coinsReceived, err := util.ParseCoinsReceivedEvents(x.coinDenom, x.multisigAddress, txResponse.Events)
		if err != nil {
			logger.WithError(err).Errorf("Error parsing coins received events")
			success = x.CreateTransaction(txResponse, models.TransactionStatusInvalid) && success
			continue
		}

		coinsSpentSender, coinsSpent, err := util.ParseCoinsSpentEvents(x.coinDenom, txResponse.Events)
		if err != nil {
			logger.WithError(err).Errorf("Error parsing coins spent events")
			success = x.CreateTransaction(txResponse, models.TransactionStatusInvalid) && success
			continue
		}

		if coinsReceived.IsZero() || coinsSpent.IsZero() {
			logger.Debugf("Found tx with zero coins")
			success = x.CreateTransaction(txResponse, models.TransactionStatusInvalid) && success
			continue
		}

		if coinsReceived.IsLTE(x.feeAmount) {
			logger.Debugf("Found tx with amount too low")
			success = x.CreateTransaction(txResponse, models.TransactionStatusInvalid) && success
			continue
		}

		if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
			logger.Debugf("Found tx with invalid coins")
			// refund
			success = x.CreateTransactionWithSpender(txResponse, models.TransactionStatusPending, coinsSpentSender) && success
			continue
		}

		memo, err := util.ValidateMemo(tx.Body.Memo)
		if err != nil {
			logger.WithError(err).WithField("memo", tx.Body.Memo).Debugf("Found invalid memo")
			// refund
			success = x.CreateTransactionWithSpender(txResponse, models.TransactionStatusPending, coinsSpentSender) && success
			continue
		}

		logger.WithField("memo", memo).Debugf("Found valid memo")
		success = x.CreateTransactionWithSpender(txResponse, models.TransactionStatusPending, coinsSpentSender) && success
	}

	if success {
		x.startBlockHeight = x.currentBlockHeight
	}

	return success
}

func (x *MessageMonitorRunner) ConfirmTxs() bool {
	x.logger.Infof("Confirming txs")
	txs, err := util.GetPendingTransactionsTo(x.chain, x.multisigPk.Address().Bytes())
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
			success = success && x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusFailed})
			continue
		}
		logger.
			Debugf("Found pending tx")

		tx := &tx.Tx{}
		err = tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			logger.Errorf("Error unmarshalling tx")
			success = success && x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusInvalid})
			continue
		}

		coinsReceived, err := util.ParseCoinsReceivedEvents(x.coinDenom, x.multisigAddress, txResponse.Events)
		if err != nil {
			logger.WithError(err).Errorf("Error parsing coins received events")
			success = success && x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusInvalid})
			continue
		}

		coinsSpentSender, coinsSpent, err := util.ParseCoinsSpentEvents(x.coinDenom, txResponse.Events)
		if err != nil {
			logger.WithError(err).Errorf("Error parsing coins spent events")
			success = success && x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusInvalid})
			continue
		}

		if coinsReceived.IsZero() || coinsSpent.IsZero() {
			logger.
				Debugf("Found tx with zero coins")
			success = success && x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusInvalid})
			continue
		}

		if coinsReceived.IsLTE(x.feeAmount) {
			logger.
				Debugf("Found tx with amount too low")
			success = success && x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusInvalid})
			continue
		}

		txHeight := txResponse.Height
		if txHeight <= 0 || uint64(txHeight) > x.currentBlockHeight {
			logger.WithField("tx_height", txHeight).Debugf("Found tx with invalid height")
			success = success && x.UpdateTransaction(&txDoc, bson.M{"status": models.TransactionStatusInvalid})
			continue
		}

		confirmations := x.currentBlockHeight - uint64(txHeight)

		update := bson.M{
			"status":        models.TransactionStatusPending,
			"confirmations": confirmations,
		}

		if confirmations >= x.confirmations {
			update["status"] = models.TransactionStatusConfirmed
		} else {
			success = success && x.UpdateTransaction(&txDoc, update)
			continue
		}

		if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
			logger.Debugf("Found tx with invalid coins")
			if refundCreated := x.CreateRefund(txResponse, &txDoc, coinsSpentSender, coinsSpent); refundCreated {
				success = success && x.UpdateTransaction(&txDoc, update)
			} else {
				success = false
			}
			continue
		}

		memo, err := util.ValidateMemo(tx.Body.Memo)
		if err != nil {
			logger.WithError(err).WithField("memo", tx.Body.Memo).Debugf("Found invalid memo")
			if refundCreated := x.CreateRefund(txResponse, &txDoc, coinsSpentSender, coinsSpent); refundCreated {
				success = success && x.UpdateTransaction(&txDoc, update)
			} else {
				success = false
			}

			continue
		}

		logger.WithField("memo", memo).Debugf("Found valid memo")
		if messageCreated := x.CreateMessage(txResponse, &txDoc, coinsSpentSender, coinsSpent, memo); messageCreated {
			success = success && x.UpdateTransaction(&txDoc, update)
		} else {
			success = false
		}
	}

	return success
}

func (x *MessageMonitorRunner) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
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

func NewMessageMonitor(config models.CosmosNetworkConfig, lastHealth *models.RunnerServiceStatus) service.Runner {
	logger := log.
		WithField("module", "cosmos").
		WithField("service", "monitor").
		WithField("chain_name", strings.ToLower(config.ChainName)).
		WithField("chain_id", strings.ToLower(config.ChainID))

	if !config.MessageMonitor.Enabled {
		logger.Fatalf("Message monitor is not enabled")
	}

	logger.Debugf("Initializing")

	var pks []crypto.PubKey
	for _, pk := range config.MultisigPublicKeys {
		pKey, err := util.PubKeyFromHex(pk)
		if err != nil {
			logger.Fatalf("Error parsing public key: %s", err)
		}
		pks = append(pks, pKey)
	}

	multisigPk := multisig.NewLegacyAminoPubKey(int(config.MultisigThreshold), pks)
	multisigAddress, err := util.Bech32FromBytes(config.Bech32Prefix, multisigPk.Address().Bytes())
	if err != nil {
		logger.Fatalf("Error creating multisig address: %s", err)
	}

	if !strings.EqualFold(multisigAddress, config.MultisigAddress) {
		logger.Fatalf("Multisig address does not match config")
	}

	client, err := cosmos.NewClient(config)
	if err != nil {
		logger.Fatalf("Error creating cosmos client: %s", err)
	}

	feeAmount := sdk.NewCoin("upokt", math.NewInt(int64(config.TxFee)))

	x := &MessageMonitorRunner{
		multisigPk: multisigPk,

		multisigAddress:    multisigAddress,
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		client:             client,
		feeAmount:          feeAmount,

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
