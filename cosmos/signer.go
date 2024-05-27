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

type MessageSignerRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	multisigAddress string
	multisigPk      *multisig.LegacyAminoPubKey

	bech32Prefix  string
	coinDenom     string
	minimumAmount sdk.Coin

	confirmations uint64

	chain  models.Chain
	client cosmos.CosmosClient

	logger *log.Entry
}

func (x *MessageSignerRunner) Run() {
	x.UpdateCurrentHeight()
	x.SignTxs()
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

func (x *MessageSignerRunner) CreateTransactionWithSpender(
	tx *sdk.TxResponse,
	txStatus models.TransactionStatus,
	coinsSpentSender string,
) bool {

	sender, err := util.ParseMessageSenderEvent(tx.Events)
	if err != nil {
		x.logger.WithError(err).Errorf("Error parsing message sender")
		return false
	}
	senderAddress, err := util.AddressBytesFromBech32(x.bech32Prefix, sender)
	if err != nil {
		x.logger.WithError(err).Errorf("Error parsing sender address")
		return false
	}

	if coinsSpentSender != "" {
		spenderAddress, err := util.AddressBytesFromBech32(x.bech32Prefix, coinsSpentSender)
		if err != nil {
			x.logger.WithError(err).Errorf("Error parsing spender address")
			return false
		}
		if bytes.Compare(senderAddress, spenderAddress) != 0 {
			x.logger.Errorf("Sender address does not match spender address")
			txStatus = models.TransactionStatusInvalid
		}
	}

	transaction, err := util.CreateTransaction(tx, x.chain, senderAddress, txStatus)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_status", txStatus).
			WithField("tx_hash", tx.TxHash).
			Errorf("Error creating transaction")
		return false
	}
	err = util.InsertTransaction(transaction)
	if err != nil {
		x.logger.WithError(err).
			WithField("tx_status", txStatus).
			WithField("tx_hash", tx.TxHash).
			Errorf("Error inserting transaction")
		return false
	}
	return true
}

func (x *MessageSignerRunner) CreateTransaction(
	tx *sdk.TxResponse,
	txStatus models.TransactionStatus,
) bool {
	return x.CreateTransactionWithSpender(tx, txStatus, "")
}

func (x *MessageSignerRunner) UpdateTransaction(
	tx *models.Transaction,
	update bson.M,
) bool {
	err := util.UpdateTransaction(tx, update)
	if err != nil {
		x.logger.Errorf("Error updating transaction: %s", err)
		return false
	}
	return true
}

func (x *MessageSignerRunner) SignTxs() bool {
	x.logger.Infof("Confirming txs")
	txs, err := util.GetPendingTransactions(x.chain)
	if err != nil {
		x.logger.Errorf("Error getting pending txs: %s", err)
		return false
	}
	x.logger.Infof("Found %d pending txs", len(txs))
	success := true
	for _, txDoc := range txs {
		logger := x.logger.WithField("tx_hash", txDoc.Hash)
		txResponse, err := x.client.GetTx(txDoc.Hash)
		if err != nil {
			logger.WithError(err).Errorf("Error getting tx")
			success = false
			continue
		}
		if txResponse.Code != 0 {
			x.logger.Infof("Found tx with error: %s", txResponse.TxHash)
			success = success && x.UpdateTransaction(&txDoc, bson.M{"tx_status": models.TransactionStatusFailed})
			continue
		}
		x.logger.Debugf("Found successful tx: %s", txResponse.TxHash)

		tx := &tx.Tx{}
		err = tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			x.logger.Errorf("Error unmarshalling tx: %s", err)
			success = success && x.UpdateTransaction(&txDoc, bson.M{"tx_status": models.TransactionStatusInvalid})
			continue
		}

		coinsReceived, err := util.ParseCoinsReceivedEvents(x.coinDenom, x.multisigAddress, txResponse.Events)
		if err != nil {
			x.logger.Errorf("Error parsing coins received events: %s", err)
			success = success && x.UpdateTransaction(&txDoc, bson.M{"tx_status": models.TransactionStatusInvalid})
			continue
		}

		x.logger.Debugf("Found tx coins received: %v", coinsReceived)

		coinsSpentSender, coinsSpent, err := util.ParseCoinsSpentEvents(x.coinDenom, txResponse.Events)
		if err != nil {
			x.logger.Errorf("Error parsing coins spent events: %s", err)
			success = success && x.UpdateTransaction(&txDoc, bson.M{"tx_status": models.TransactionStatusInvalid})
			continue
		}

		x.logger.Debugf("Found tx coins spent: %v", coinsSpent)
		x.logger.Debugf("Found tx coins spent sender: %s", coinsSpentSender)

		if coinsReceived.IsZero() || coinsSpent.IsZero() {
			x.logger.Debugf("Found tx with zero coins: %s", txResponse.TxHash)
			success = success && x.UpdateTransaction(&txDoc, bson.M{"tx_status": models.TransactionStatusInvalid})
			continue
		}

		if coinsReceived.IsLTE(x.minimumAmount) {
			x.logger.Debugf("Found tx with too low amount: %s", txResponse.TxHash)
			success = success && x.UpdateTransaction(&txDoc, bson.M{"tx_status": models.TransactionStatusInvalid})
			continue
		}

		txHeight := txResponse.Height
		if txHeight <= 0 || uint64(txHeight) > x.currentBlockHeight {
			x.logger.Debugf("Found tx with invalid height: %s", txResponse.TxHash)
			success = success && x.UpdateTransaction(&txDoc, bson.M{"tx_status": models.TransactionStatusInvalid})
			continue
		}

		confirmations := x.currentBlockHeight - uint64(txHeight)

		update := bson.M{
			"tx_status":     models.TransactionStatusPending,
			"confirmations": confirmations,
		}
		if confirmations >= x.confirmations {
			update["tx_status"] = models.TransactionStatusConfirmed
		}

		success = success && x.UpdateTransaction(&txDoc, update)
	}

	if success {
		x.startBlockHeight = x.currentBlockHeight
	}

	return success
}

func (x *MessageSignerRunner) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
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

func NewMessageSigner(config models.CosmosNetworkConfig, lastHealth *models.RunnerServiceStatus) service.Runner {
	logger := log.
		WithField("module", "ethereum").
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
			logger.Fatalf("Error parsing public key: %s", err)
		}
		pks = append(pks, pKey)
	}

	multisigPk := multisig.NewLegacyAminoPubKey(int(config.MultisigThreshold), pks)
	multisigAddress, err := util.Bech32FromAddressBytes(config.Bech32Prefix, multisigPk.Address().Bytes())
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

	x := &MessageSignerRunner{
		multisigPk: multisigPk,

		multisigAddress:    multisigAddress,
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		client:             client,
		minimumAmount:      feeAmount,

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
