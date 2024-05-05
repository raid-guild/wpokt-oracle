package pokt

import (
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/app"
	"github.com/dan13ram/wpokt-oracle/models"
	pokt "github.com/dan13ram/wpokt-oracle/pokt/client"
	"github.com/dan13ram/wpokt-oracle/pokt/util"
	"github.com/pokt-network/pocket-core/crypto"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MintMonitorName = "MINT MONITOR"
)

type MintMonitorRunner struct {
	client        pokt.PocketClient
	wpoktAddress  string
	vaultAddress  string
	startHeight   int64
	currentHeight int64
	minimumAmount *big.Int
}

func (x *MintMonitorRunner) Run() {
	x.UpdateCurrentHeight()
	x.SyncTxs()
}

func (x *MintMonitorRunner) Status() models.RunnerStatus {
	return models.RunnerStatus{
		PoktHeight: strconv.FormatInt(x.startHeight, 10),
	}
}

func (x *MintMonitorRunner) UpdateCurrentHeight() {
	res, err := x.client.GetHeight()
	if err != nil {
		log.Error("[MINT MONITOR] Error getting current height: ", err)
		return
	}
	x.currentHeight = res.Height
	log.Info("[MINT MONITOR] Current height: ", x.currentHeight)
}

func (x *MintMonitorRunner) HandleFailedMint(tx *pokt.TxResponse) bool {
	if tx == nil {
		log.Debug("[MINT MONITOR] Invalid tx response")
		return false
	}

	doc := util.CreateFailedMint(tx, x.vaultAddress)

	log.Debug("[MINT MONITOR] Storing failed mint tx")
	err := app.DB.InsertOne(models.CollectionInvalidMints, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Info("[MINT MONITOR] Found duplicate failed mint tx")
			return true
		}
		log.Error("[MINT MONITOR] Error storing failed mint tx: ", err)
		return false
	}

	log.Info("[MINT MONITOR] Stored failed mint tx")
	return true
}

func (x *MintMonitorRunner) HandleInvalidMint(tx *pokt.TxResponse) bool {
	if tx == nil {
		log.Debug("[MINT MONITOR] Invalid tx response")
		return false
	}

	doc := util.CreateInvalidMint(tx, x.vaultAddress)

	log.Debug("[MINT MONITOR] Storing invalid mint tx")
	err := app.DB.InsertOne(models.CollectionInvalidMints, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Info("[MINT MONITOR] Found duplicate invalid mint tx")
			return true
		}
		log.Error("[MINT MONITOR] Error storing invalid mint tx: ", err)
		return false
	}

	log.Info("[MINT MONITOR] Stored invalid mint tx")
	return true
}

func (x *MintMonitorRunner) HandleValidMint(tx *pokt.TxResponse, memo models.MintMemo) bool {
	if tx == nil {
		log.Debug("[MINT MONITOR] Invalid tx response")
		return false
	}

	doc := util.CreateMint(tx, memo, x.wpoktAddress, x.vaultAddress)

	log.Debug("[MINT MONITOR] Storing mint tx")
	err := app.DB.InsertOne(models.CollectionMints, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Info("[MINT MONITOR] Found duplicate mint tx")
			return true
		}
		log.Error("[MINT MONITOR] Error storing mint tx: ", err)
		return false
	}

	log.Info("[MINT MONITOR] Stored mint tx")
	return true
}

func (x *MintMonitorRunner) SyncTxs() bool {

	if x.currentHeight <= x.startHeight {
		log.Info("[MINT MONITOR] No new blocks to sync")
		return true
	}

	txs, err := x.client.GetAccountTxsByHeight(x.vaultAddress, x.startHeight)
	if err != nil {
		log.Error("[MINT MONITOR] Error getting txs: ", err)
		return false
	}
	log.Info("[MINT MONITOR] Found ", len(txs), " txs to sync")
	var success bool = true
	for i := range txs {
		tx := txs[i]

		amount, ok := new(big.Int).SetString(tx.StdTx.Msg.Value.Amount, 10)
		if tx.Tx == "" || tx.TxResult.Code != 0 || !strings.EqualFold(tx.TxResult.Recipient, x.vaultAddress) || tx.TxResult.MessageType != "send" || !ok || amount.Cmp(x.minimumAmount) != 1 {
			log.Info("[MINT MONITOR] Found failed mint tx: ", tx.Hash, " with code: ", tx.TxResult.Code)
			success = x.HandleFailedMint(tx) && success
			continue
		}
		memo, ok := util.ValidateMemo(tx.StdTx.Memo)
		if !ok {
			log.Info("[MINT MONITOR] Found invalid mint tx: ", tx.Hash, " with memo: ", "\""+tx.StdTx.Memo+"\"")
			success = x.HandleInvalidMint(tx) && success
			continue
		}

		log.Info("[MINT MONITOR] Found valid mint tx: ", tx.Hash, " with memo: ", tx.StdTx.Memo)
		success = x.HandleValidMint(tx, memo) && success
	}

	if success {
		x.startHeight = x.currentHeight
	}

	return success
}

func (x *MintMonitorRunner) InitStartHeight(lastHealth models.ServiceHealth) {
	startHeight := (app.Config.Pocket.StartHeight)

	if (lastHealth.PoktHeight) != "" {
		if lastHeight, err := strconv.ParseInt(lastHealth.PoktHeight, 10, 64); err == nil {
			startHeight = lastHeight
		}
	}
	if startHeight > 0 {
		x.startHeight = startHeight
	} else {
		log.Info("[MINT MONITOR] Found invalid start height, using current height")
		x.startHeight = x.currentHeight
	}
	log.Info("[MINT MONITOR] Start height: ", x.startHeight)
}

func NewMintMonitor(wg *sync.WaitGroup, lastHealth models.ServiceHealth) app.Service {
	if !app.Config.MintMonitor.Enabled {
		log.Debug("[MINT MONITOR] Disabled")
		return app.NewEmptyService(wg)
	}

	log.Debug("[MINT MONITOR] Initializing")
	var pks []crypto.PublicKey
	for _, pk := range app.Config.Pocket.MultisigPublicKeys {
		p, err := crypto.NewPublicKey(pk)
		if err != nil {
			log.Fatal("[MINT MONITOR] Error parsing multisig public key: ", err)
		}
		pks = append(pks, p)
	}

	vaultPk := crypto.PublicKeyMultiSignature{PublicKeys: pks}
	vaultAddress := vaultPk.Address().String()
	log.Debug("[MINT MONITOR] Vault address: ", vaultAddress)
	if !strings.EqualFold(vaultAddress, app.Config.Pocket.VaultAddress) {
		log.Fatal("[MINT MONITOR] Multisig address does not match vault address")
	}

	x := &MintMonitorRunner{
		vaultAddress:  strings.ToLower(vaultAddress),
		wpoktAddress:  strings.ToLower(app.Config.Ethereum.WrappedPocketAddress),
		startHeight:   0,
		currentHeight: 0,
		client:        pokt.NewClient(),
		minimumAmount: big.NewInt(app.Config.Pocket.TxFee),
	}

	x.UpdateCurrentHeight()

	x.InitStartHeight(lastHealth)

	log.Info("[MINT MONITOR] Initialized")

	return app.NewRunnerService(MintMonitorName, x, wg, time.Duration(app.Config.MintMonitor.IntervalMillis)*time.Millisecond)
}
