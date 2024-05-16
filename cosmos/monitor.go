package cosmos

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

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

	bech32Prefix  string
	coinDenom     string
	minimumAmount sdk.Coin

	chain  models.Chain
	client cosmos.CosmosClient

	logger *log.Entry
}

func (x *MessageMonitorRunner) Run() {
	x.UpdateCurrentHeight()
	x.SyncTxs()
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

// transaction has failed
func (x *MessageMonitorRunner) HandleFailedTransaction(tx *sdk.TxResponse) bool {
	x.logger.Debugf("Handling failed tx: %s", tx.TxHash)
	return true
}

// transaction was successful but invalid
func (x *MessageMonitorRunner) HandleInvalidTransaction(tx *sdk.TxResponse) bool {
	x.logger.Debugf("Handling invalid tx: %s", tx.TxHash)
	return true
}

// transaction was successful but cannot be processed and needs to be refunded
func (x *MessageMonitorRunner) HandleRefundTransaction(tx *sdk.TxResponse) bool {
	x.logger.Debugf("Handling refund tx: %s", tx.TxHash)
	return true
}

// transaction was successful and valid and can be processed
func (x *MessageMonitorRunner) HandleValidTransaction(tx *sdk.TxResponse, memo models.MintMemo) bool {
	x.logger.Debugf("Handling valid tx: %s", tx.TxHash)
	return true
}

// func (x *MessageMonitorRunner) HandleInvalidMint(tx *pokt.TxResponse) bool {
// if tx == nil {
// 	x.logger.Debug("[MINT MONITOR] Invalid tx response")
// 	return false
// }
//
// doc := util.CreateInvalidMint(tx, x.vaultAddress)
//
// x.logger.Debug("[MINT MONITOR] Storing invalid mint tx")
// err := app.DB.InsertOne(models.CollectionInvalidMints, doc)
// if err != nil {
// 	if mongo.IsDuplicateKeyError(err) {
// 		x.logger.Info("[MINT MONITOR] Found duplicate invalid mint tx")
// 		return true
// 	}
// 	x.logger.Error("[MINT MONITOR] Error storing invalid mint tx: ", err)
// 	return false
// }
//
// x.logger.Info("[MINT MONITOR] Stored invalid mint tx")
// 	return true
// }

// func (x *MessageMonitorRunner) HandleValidMint(tx *pokt.TxResponse, memo models.MintMemo) bool {
// if tx == nil {
// 	x.logger.Debug("[MINT MONITOR] Invalid tx response")
// 	return false
// }
//
// doc := util.CreateMint(tx, memo, x.wpoktAddress, x.vaultAddress)
//
// x.logger.Debug("[MINT MONITOR] Storing mint tx")
// err := app.DB.InsertOne(models.CollectionMints, doc)
// if err != nil {
// 	if mongo.IsDuplicateKeyError(err) {
// 		x.logger.Info("[MINT MONITOR] Found duplicate mint tx")
// 		return true
// 	}
// 	x.logger.Error("[MINT MONITOR] Error storing mint tx: ", err)
// 	return false
// }
//
// x.logger.Info("[MINT MONITOR] Stored mint tx")
// 	return true
// }

func (x *MessageMonitorRunner) SyncTxs() bool {
	x.logger.Infof("Syncing txs")
	if x.currentBlockHeight <= x.startBlockHeight {
		x.logger.Infof("No new blocks to sync")
		return true
	}

	txResponses, err := x.client.GetTxsSentToAddressAfterHeight(x.multisigAddress, x.startBlockHeight)
	if err != nil {
		x.logger.Errorf("Error getting txs: %s", err)
		return false
	}
	x.logger.Infof("Found %d txs to sync", len(txResponses))
	success := true
	for _, txResponse := range txResponses {
		if txResponse.Code != 0 {
			x.logger.Infof("Found tx with error: %s", txResponse.TxHash)
			success = x.HandleFailedTransaction(txResponse) && success
			continue
		}
		x.logger.Debugf("Found successful tx: %s", txResponse.TxHash)

		tx := &tx.Tx{}
		err := tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			x.logger.Errorf("Error unmarshalling tx: %s", err)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		coinsReceived, err := util.ParseCoinsReceivedEvents(x.coinDenom, x.multisigAddress, txResponse.Events)
		if err != nil {
			x.logger.Errorf("Error parsing coins received events: %s", err)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		x.logger.Debugf("Found tx coins received: %v", coinsReceived)

		coinsSpentSender, coinsSpent, err := util.ParseCoinsSpentEvents(x.coinDenom, txResponse.Events)
		if err != nil {
			x.logger.Errorf("Error parsing coins spent events: %s", err)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		x.logger.Debugf("Found tx coins spent: %v", coinsSpent)
		x.logger.Debugf("Found tx coins spent sender: %s", coinsSpentSender)

		if coinsReceived.IsZero() || coinsSpent.IsZero() {
			x.logger.Debugf("Found tx with zero coins: %s", txResponse.TxHash)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		if coinsReceived.IsLTE(x.minimumAmount) {
			x.logger.Debugf("Found tx with too low amount: %s", txResponse.TxHash)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
			x.logger.Debugf("Found tx with invalid coins: %s", txResponse.TxHash)
			success = x.HandleRefundTransaction(txResponse) && success
			continue
		}

		x.logger.Debugf("Found tx memo: %s", tx.Body.Memo)

		memo, err := util.ValidateMemo(tx.Body.Memo)
		if err != nil {
			x.logger.Debugf("Found invalid memo: %s", err)
			success = x.HandleRefundTransaction(txResponse) && success
			continue
		}

		x.logger.Infof("Found tx with valid memo: %v", memo)
		success = x.HandleValidTransaction(txResponse, memo) && success
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

func NewMessageMonitor(config models.CosmosNetworkConfig, lastHealth *models.RunnerServiceStatus) service.Runner {
	logger := log.
		WithField("module", "ethereum").
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

	x := &MessageMonitorRunner{
		multisigPk: multisigPk,

		multisigAddress:    multisigAddress,
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		client:             client,
		minimumAmount:      feeAmount,

		bech32Prefix: config.Bech32Prefix,
		coinDenom:    config.CoinDenom,

		logger: logger,
	}

	x.UpdateCurrentHeight()

	x.InitStartBlockHeight(lastHealth)

	logger.Infof("Initialized")

	return x
}
