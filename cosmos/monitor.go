package cosmos

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/app/service"
	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"
)

type MessageMonitorRunner struct {
	startHeight     int64
	currentHeight   int64
	name            string
	multisigAddress string
	bech32Prefix       string
	coinDenom             string

	client        cosmos.CosmosClient
	multisigPk    *multisig.LegacyAminoPubKey
	minimumAmount sdk.Coin
}

func (x *MessageMonitorRunner) Run() {
	x.UpdateCurrentHeight()
	x.SyncTxs()
}

func (x *MessageMonitorRunner) Height() uint64 {
	return uint64(x.currentHeight)
}

func (x *MessageMonitorRunner) UpdateCurrentHeight() {
	height, err := x.client.GetLatestBlockHeight()
	if err != nil {
		log.Errorf("[%s] Error getting latest block: %s", x.name, err)
		return
	}
	x.currentHeight = height

	log.Infof("[%s] Current height: %d", x.name, x.currentHeight)
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
	if x.currentHeight <= x.startHeight {
		log.Infof("[%s] No new blocks to sync", x.name)
		return true
	}

	txResponses, err := x.client.GetTxsSentToAddressAfterHeight(x.multisigAddress, x.startHeight)
	if err != nil {
		log.Errorf("[%s] Error getting txs: %s", x.name, err)
		return false
	}
	log.Infof("[%s] Found %d txs to sync", x.name, len(txResponses))
	success := true
	for _, txResponse := range txResponses {
		if txResponse.Code != 0 {
			log.Infof("[%s] Found tx with error: %s", x.name, txResponse.TxHash)
			continue
		}
		log.Debugf("[%s] Found tx: %s", x.name, txResponse.TxHash)

		tx := &tx.Tx{}
		err := tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			log.Errorf("[%s] Error unmarshalling tx: %s", x.name, err)
			continue
		}

		log.Infof("[%s] Found tx memo: %s", x.name, tx.Body.Memo)

		coinsReceived, err := util.ParseCoinsReceivedEvents(x.coinDenom, x.multisigAddress, txResponse.Events)
		if err != nil {
			log.Errorf("[%s] Error parsing coins received events: %s", x.name, err)
			continue
		}

		log.Infof("[%s] Found tx coins received: %v", x.name, coinsReceived)

		coinsSpentSender, coinsSpent, err := util.ParseCoinsSpentEvents(x.coinDenom, txResponse.Events)
		if err != nil {
			log.Errorf("[%s] Error parsing coins spent events: %s", x.name, err)
			continue
		}

		log.Infof("[%s] Found tx coins spent: %v", x.name, coinsSpent)
		log.Infof("[%s] Found tx coins spent sender: %s", x.name, coinsSpentSender)

		if coinsReceived.IsZero() || coinsSpent.IsZero() {
			log.Infof("[%s] Found tx with zero coins: %s", x.name, txResponse.TxHash)
			continue
		}

		if coinsReceived.IsLTE(x.minimumAmount) {
			log.Infof("[%s] Found tx with too low amount: %s", x.name, txResponse.TxHash)
			continue
		}

		if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
			log.Infof("[%s] Found tx with invalid coins: %s", x.name, txResponse.TxHash)
		}

		// amount, ok := new(big.Int).SetString(tx.StdTx.Msg.Value.Amount, 10)
		// if tx.Tx == "" || tx.TxResult.Code != 0 || !strings.EqualFold(tx.TxResult.Recipient, x.vaultAddress) || tx.TxResult.MessageType != "send" || !ok || amount.Cmp(x.minimumAmount) != 1 {
		// 	log.Info("[MINT MONITOR] Found failed mint tx: ", tx.Hash, " with code: ", tx.TxResult.Code)
		// 	success = x.HandleFailedMint(tx) && success
		// 	continue
		// }
		// memo, ok := util.ValidateMemo(tx.StdTx.Memo)
		// if !ok {
		// 	log.Info("[MINT MONITOR] Found invalid mint tx: ", tx.Hash, " with memo: ", "\""+tx.StdTx.Memo+"\"")
		// 	success = x.HandleInvalidMint(tx) && success
		// 	continue
		// }
		//
		// log.Info("[MINT MONITOR] Found valid mint tx: ", tx.Hash, " with memo: ", tx.StdTx.Memo)
		// success = x.HandleValidMint(tx, memo) && success
	}

	if success {
		x.startHeight = x.currentHeight
	}

	return success
}

func (x *MessageMonitorRunner) InitStartHeight(lastHealth models.ChainServiceHealth) {
	// startHeight := (app.Config.Pocket.StartHeight)
	//
	// if (lastHealth.PoktHeight) != "" {
	// 	if lastHeight, err := strconv.ParseInt(lastHealth.PoktHeight, 10, 64); err == nil {
	// 		startHeight = lastHeight
	// 	}
	// }
	// if startHeight > 0 {
	// 	x.startHeight = startHeight
	// } else {
	// 	log.Info("[MINT MONITOR] Found invalid start height, using current height")
	// 	x.startHeight = x.currentHeight
	// }
	// log.Info("[MINT MONITOR] Start height: ", x.startHeight)
}

func NewMessageMonitor(config models.CosmosNetworkConfig, lastHealth models.ChainServiceHealth) service.Runner {

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

	feeAmount := sdk.NewCoin("upokt", math.NewInt(config.TxFee))

	x := &MessageMonitorRunner{
		name:       name,
		multisigPk: multisigPk,
		// wpoktAddress:  strings.ToLower(app.Config.Ethereum.WrappedPocketAddress),
		multisigAddress: multisigAddress,
		startHeight:     0,
		currentHeight:   0,
		client:          client,
		minimumAmount:   feeAmount,

		bech32Prefix: config.Bech32Prefix,
		coinDenom:       config.CoinDenom,
	}

	x.UpdateCurrentHeight()

	// x.InitStartHeight(lastHealth)

	log.Infof("[%s] Initialized", name)

	return x
}
