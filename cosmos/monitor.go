package cosmos

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/dan13ram/wpokt-oracle/app/service"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

type MessageMonitorRunner struct {
	name   string
	client cosmos.CosmosClient
	// wpoktAddress  string
	multisigPk      *multisig.LegacyAminoPubKey
	multisigAddress string
	startHeight     int64
	currentHeight   int64
	minimumAmount   *big.Int
}

func (x *MessageMonitorRunner) Run() {
	x.UpdateCurrentHeight()
	x.SyncTxs()
}

func (x *MessageMonitorRunner) Status() models.RunnerStatus {
	return models.RunnerStatus{}
}

func (x *MessageMonitorRunner) UpdateCurrentHeight() {
	res, err := x.client.GetLatestBlock()
	if err != nil {
		log.Error("[MINT MONITOR] Error getting current height: ", err)
		return
	}
	x.currentHeight = res.Header.Height

	log.Infof("[%s] Current height: %d", x.name, x.currentHeight)
}

// func (x *MessageMonitorRunner) HandleFailedMint(tx *pokt.TxResponse) bool {
// if tx == nil {
// 	log.Debug("[MINT MONITOR] Invalid tx response")
// 	return false
// }
//
// doc := util.CreateFailedMint(tx, x.vaultAddress)
//
// log.Debug("[MINT MONITOR] Storing failed mint tx")
// err := app.DB.InsertOne(models.CollectionInvalidMints, doc)
// if err != nil {
// 	if mongo.IsDuplicateKeyError(err) {
// 		log.Info("[MINT MONITOR] Found duplicate failed mint tx")
// 		return true
// 	}
// 	log.Error("[MINT MONITOR] Error storing failed mint tx: ", err)
// 	return false
// }
//
// log.Info("[MINT MONITOR] Stored failed mint tx")
// 	return true
// }

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
	var success bool = true
	for _, txResponse := range txResponses {
		if txResponse.Code != 0 {
			log.Infof("[%s] Found tx with error: %s", x.name, txResponse.TxHash)
			continue
		}
		log.Debugf("[%s] Found tx: %s", x.name, txResponse.TxHash)
		tx := &tx.Tx{}
		// err := txResponse.Tx.
		err := tx.Unmarshal(txResponse.Tx.Value)
		if err != nil {
			log.Errorf("[%s] Error unmarshalling tx: %s", x.name, err)
			continue
		}

		log.Infof("[%s] Found tx memo: %s", x.name, tx.Body.Memo)
		// amount
		// log.Infof("[%s] Found tx amount: %s", x.name, tx.Body.

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

func (x *MessageMonitorRunner) InitStartHeight(lastHealth models.ServiceHealth) {
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

func NewMessageMonitor(config models.CosmosNetworkConfig, lastHealth models.ServiceHealth) service.Runner {

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

	x := &MessageMonitorRunner{
		name:       name,
		multisigPk: multisigPk,
		// wpoktAddress:  strings.ToLower(app.Config.Ethereum.WrappedPocketAddress),
		multisigAddress: multisigAddress,
		startHeight:     0,
		currentHeight:   0,
		client:          client,
		minimumAmount:   big.NewInt(1),
	}

	x.UpdateCurrentHeight()

	// x.InitStartHeight(lastHealth)

	log.Infof("[%s] Initialized", name)

	return x
}
