package cosmos

import (
	"fmt"
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
	name string

	startBlockHeight   uint64
	currentBlockHeight uint64

	multisigAddress string
	multisigPk      *multisig.LegacyAminoPubKey

	bech32Prefix  string
	coinDenom     string
	minimumAmount sdk.Coin

	chain  models.Chain
	client cosmos.CosmosClient
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
		log.Errorf("[%s] Error getting latest block: %s", x.name, err)
		return
	}
	x.currentBlockHeight = uint64(height)

	log.Infof("[%s] Current height: %d", x.name, x.currentBlockHeight)
}

// transaction has failed
func (x *MessageMonitorRunner) HandleFailedTransaction(tx *sdk.TxResponse) bool {
	log.Debugf("[%s] Handling failed tx: %s", x.name, tx.TxHash)
	return true
}

// transaction was successful but invalid
func (x *MessageMonitorRunner) HandleInvalidTransaction(tx *sdk.TxResponse) bool {
	log.Debugf("[%s] Handling invalid tx: %s", x.name, tx.TxHash)
	return true
}

// transaction was successful but cannot be processed and needs to be refunded
func (x *MessageMonitorRunner) HandleRefundTransaction(tx *sdk.TxResponse) bool {
	log.Debugf("[%s] Handling refund tx: %s", x.name, tx.TxHash)
	return true
}

// transaction was successful and valid and can be processed
func (x *MessageMonitorRunner) HandleValidTransaction(tx *sdk.TxResponse, memo models.MintMemo) bool {
	log.Debugf("[%s] Handling valid tx: %s", x.name, tx.TxHash)
	return true
}

// func (x *MessageMonitorRunner) HandleInvalidMint(tx *pokt.TxResponse) bool {
// if tx == nil {
// 	log.Debug("[MINT MONITOR] Invalid tx response")
// 	return false
// }
//
// doc := util.CreateInvalidMint(tx, x.vaultAddress)
//
// log.Debug("[MINT MONITOR] Storing invalid mint tx")
// err := app.DB.InsertOne(models.CollectionInvalidMints, doc)
// if err != nil {
// 	if mongo.IsDuplicateKeyError(err) {
// 		log.Info("[MINT MONITOR] Found duplicate invalid mint tx")
// 		return true
// 	}
// 	log.Error("[MINT MONITOR] Error storing invalid mint tx: ", err)
// 	return false
// }
//
// log.Info("[MINT MONITOR] Stored invalid mint tx")
// 	return true
// }

// func (x *MessageMonitorRunner) HandleValidMint(tx *pokt.TxResponse, memo models.MintMemo) bool {
// if tx == nil {
// 	log.Debug("[MINT MONITOR] Invalid tx response")
// 	return false
// }
//
// doc := util.CreateMint(tx, memo, x.wpoktAddress, x.vaultAddress)
//
// log.Debug("[MINT MONITOR] Storing mint tx")
// err := app.DB.InsertOne(models.CollectionMints, doc)
// if err != nil {
// 	if mongo.IsDuplicateKeyError(err) {
// 		log.Info("[MINT MONITOR] Found duplicate mint tx")
// 		return true
// 	}
// 	log.Error("[MINT MONITOR] Error storing mint tx: ", err)
// 	return false
// }
//
// log.Info("[MINT MONITOR] Stored mint tx")
// 	return true
// }

func (x *MessageMonitorRunner) SyncTxs() bool {
	log.Infof("[%s] Syncing txs", x.name)
	if x.currentBlockHeight <= x.startBlockHeight {
		log.Infof("[%s] No new blocks to sync", x.name)
		return true
	}

	txResponses, err := x.client.GetTxsSentToAddressAfterHeight(x.multisigAddress, x.startBlockHeight)
	if err != nil {
		log.Errorf("[%s] Error getting txs: %s", x.name, err)
		return false
	}
	log.Infof("[%s] Found %d txs to sync", x.name, len(txResponses))
	success := true
	for _, txResponse := range txResponses {
		if txResponse.Code != 0 {
			log.Infof("[%s] Found tx with error: %s", x.name, txResponse.TxHash)
			success = x.HandleFailedTransaction(txResponse) && success
			continue
		}
		log.Debugf("[%s] Found successful tx: %s", x.name, txResponse.TxHash)

		tx := &tx.Tx{}
		err := tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			log.Errorf("[%s] Error unmarshalling tx: %s", x.name, err)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		coinsReceived, err := util.ParseCoinsReceivedEvents(x.coinDenom, x.multisigAddress, txResponse.Events)
		if err != nil {
			log.Errorf("[%s] Error parsing coins received events: %s", x.name, err)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		log.Debugf("[%s] Found tx coins received: %v", x.name, coinsReceived)

		coinsSpentSender, coinsSpent, err := util.ParseCoinsSpentEvents(x.coinDenom, txResponse.Events)
		if err != nil {
			log.Errorf("[%s] Error parsing coins spent events: %s", x.name, err)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		log.Debugf("[%s] Found tx coins spent: %v", x.name, coinsSpent)
		log.Debugf("[%s] Found tx coins spent sender: %s", x.name, coinsSpentSender)

		if coinsReceived.IsZero() || coinsSpent.IsZero() {
			log.Debugf("[%s] Found tx with zero coins: %s", x.name, txResponse.TxHash)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		if coinsReceived.IsLTE(x.minimumAmount) {
			log.Debugf("[%s] Found tx with too low amount: %s", x.name, txResponse.TxHash)
			success = x.HandleInvalidTransaction(txResponse) && success
			continue
		}

		if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
			log.Debugf("[%s] Found tx with invalid coins: %s", x.name, txResponse.TxHash)
			success = x.HandleRefundTransaction(txResponse) && success
			continue
		}

		log.Debugf("[%s] Found tx memo: %s", x.name, tx.Body.Memo)

		memo, err := util.ValidateMemo(tx.Body.Memo)
		if err != nil {
			log.Debugf("[%s] Found invalid memo: %s", x.name, err)
			success = x.HandleRefundTransaction(txResponse) && success
			continue
		}

		log.Infof("[%s] Found tx with valid memo: %v", x.name, memo)
		success = x.HandleValidTransaction(txResponse, memo) && success
	}

	if success {
		x.startBlockHeight = x.currentBlockHeight
	}

	return success
}

func (x *MessageMonitorRunner) InitStartBlockHeight(lastHealth *models.RunnerServiceStatus) {
	if lastHealth == nil || lastHealth.BlockHeight == 0 {
		log.Infof("[%s] Invalid last health", x.name)
	} else {
		log.Debugf("[%s] Last block height: %d", x.name, lastHealth.BlockHeight)
		x.startBlockHeight = lastHealth.BlockHeight
	}
	if x.startBlockHeight == 0 || x.startBlockHeight > x.currentBlockHeight {
		log.Infof("[%s] Start block height is greater than current block height", x.name)
		x.startBlockHeight = x.currentBlockHeight
	}
	log.Infof("[%s] Initialized start block height: %d", x.name, x.startBlockHeight)
}

func NewMessageMonitor(config models.CosmosNetworkConfig, lastHealth *models.RunnerServiceStatus) service.Runner {

	name := strings.ToUpper(fmt.Sprintf("%s_Monitor", config.ChainName))

	if !config.MessageMonitor.Enabled {
		log.Fatalf("[%s] Message monitor is not enabled", name)
	}

	log.Debugf("[%s] Initializing", name)

	var pks []crypto.PubKey
	for _, pk := range config.MultisigPublicKeys {
		pKey, err := util.PubKeyFromHex(pk)
		if err != nil {
			log.Fatalf("[%s] Error parsing public key: %s", name, err)
		}
		pks = append(pks, pKey)
	}

	multisigPk := multisig.NewLegacyAminoPubKey(int(config.MultisigThreshold), pks)
	multisigAddress, err := util.Bech32FromAddressBytes(config.Bech32Prefix, multisigPk.Address().Bytes())
	if err != nil {
		log.Fatalf("[%s] Error creating multisig address: %s", name, err)
	}

	if !strings.EqualFold(multisigAddress, config.MultisigAddress) {
		log.Fatalf("[%s] Multisig address does not match config", name)
	}

	client, err := cosmos.NewClient(config)
	if err != nil {
		log.Fatalf("[%s] Error creating cosmos client: %s", name, err)
	}

	feeAmount := sdk.NewCoin("upokt", math.NewInt(int64(config.TxFee)))

	x := &MessageMonitorRunner{
		name:       name,
		multisigPk: multisigPk,

		multisigAddress:    multisigAddress,
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		client:             client,
		minimumAmount:      feeAmount,

		bech32Prefix: config.Bech32Prefix,
		coinDenom:    config.CoinDenom,
	}

	x.UpdateCurrentHeight()

	x.InitStartBlockHeight(lastHealth)

	log.Infof("[%s] Initialized", name)

	return x
}
