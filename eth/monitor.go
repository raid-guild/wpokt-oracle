package eth

import (
	"fmt"
	"math/big"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/app/service"
	"github.com/dan13ram/wpokt-oracle/eth/autogen"
	eth "github.com/dan13ram/wpokt-oracle/eth/client"
	"github.com/dan13ram/wpokt-oracle/models"
)

type MessageMonitorRunner struct {
	name               string
	startBlockHeight   int64
	currentBlockHeight int64
	wpoktContract      eth.WrappedPocketContract
	client             eth.EthereumClient
	minimumAmount      *big.Int
}

func (x *MessageMonitorRunner) Run() {
	x.UpdateCurrentBlockHeight()
	x.SyncTxs()
}

func (x *MessageMonitorRunner) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *MessageMonitorRunner) UpdateCurrentBlockHeight() {
	res, err := x.client.GetBlockHeight()
	if err != nil {
		log.Errorf("[%s] Error getting latest block: %s", x.name, err)
		return
	}
	x.currentBlockHeight = int64(res)
	log.Infof("[%s] Current block number: %d", x.name, x.currentBlockHeight)
}

func (x *MessageMonitorRunner) HandleBurnEvent(event *autogen.WrappedPocketBurnAndBridge) bool {
	// if event == nil {
	// 	log.Error("[BURN MONITOR] Error while handling burn event: event is nil")
	// 	return false
	// }
	//
	// doc := util.CreateBurn(event)
	//
	// // each event is a combination of transaction hash and log index
	// log.Debug("[BURN MONITOR] Handling burn event: ", event.Raw.TxHash, " ", event.Raw.Index)
	//
	// err := app.DB.InsertOne(models.CollectionBurns, doc)
	// if err != nil {
	// 	if mongo.IsDuplicateKeyError(err) {
	// 		log.Info("[BURN MONITOR] Found duplicate burn event: ", event.Raw.TxHash, " ", event.Raw.Index)
	// 		return true
	// 	}
	// 	log.Error("[BURN MONITOR] Error while storing burn event in db: ", err)
	// 	return false
	// }
	//
	// log.Info("[BURN MONITOR] Stored burn event: ", event.Raw.TxHash, " ", event.Raw.Index)
	return true
}

func (x *MessageMonitorRunner) SyncBlocks(startBlockHeight uint64, endBlockHeight uint64) bool {
	// filter, err := x.wpoktContract.FilterBurnAndBridge(&bind.FilterOpts{
	// 	Start:   startBlockHeight,
	// 	End:     &endBlockHeight,
	// 	Context: context.Background(),
	// }, []*big.Int{}, []common.Address{}, []common.Address{})
	//
	// if filter != nil {
	// 	defer filter.Close()
	// }
	//
	// if err != nil {
	// 	log.Error("[BURN MONITOR] Error while syncing burn events: ", err)
	// 	return false
	// }

	var success bool = true
	// for filter.Next() {
	// 	if err := filter.Error(); err != nil {
	// 		success = false
	// 		break
	// 	}
	//
	// 	event := filter.Event()
	//
	// 	if event == nil {
	// 		success = false
	// 		continue
	// 	}
	//
	// 	if event.Raw.Removed || event.Amount.Cmp(x.minimumAmount) != 1 {
	// 		continue
	// 	}
	//
	// 	success = x.HandleBurnEvent(event) && success
	// }
	//
	// if err := filter.Error(); err != nil {
	// 	log.Error("[BURN MONITOR] Error while syncing burn events: ", err)
	// 	return false
	// }

	return success
}

func (x *MessageMonitorRunner) SyncTxs() bool {
	if x.currentBlockHeight <= x.startBlockHeight {
		log.Infof("[%s] No new blocks to sync", x.name)
		return true
	}

	var success bool = true
	// if (x.currentBlockHeight - x.startBlockHeight) > eth.MAX_QUERY_BLOCKS {
	// 	log.Debug("[BURN MONITOR] Syncing burn txs in chunks")
	// 	for i := x.startBlockHeight; i < x.currentBlockHeight; i += eth.MAX_QUERY_BLOCKS {
	// 		endBlockHeight := i + eth.MAX_QUERY_BLOCKS
	// 		if endBlockHeight > x.currentBlockHeight {
	// 			endBlockHeight = x.currentBlockHeight
	// 		}
	// 		log.Info("[BURN MONITOR] Syncing burn txs from blockNumber: ", i, " to blockNumber: ", endBlockHeight)
	// 		success = success && x.SyncBlocks(uint64(i), uint64(endBlockHeight))
	// 	}
	// } else {
	// 	log.Info("[BURN MONITOR] Syncing burn txs from blockNumber: ", x.startBlockHeight, " to blockNumber: ", x.currentBlockHeight)
	// 	success = success && x.SyncBlocks(uint64(x.startBlockHeight), uint64(x.currentBlockHeight))
	// }
	//
	// if success {
	// 	x.startBlockHeight = x.currentBlockHeight
	// }

	return success
}

func (x *MessageMonitorRunner) InitStartBlockHeight(lastHealth models.ChainServiceHealth) {
	// startBlockHeight := int64(app.Config.Ethereum.StartBlockHeight)
	//
	// if lastBlockHeight, err := strconv.ParseInt(lastHealth.EthBlockHeight, 10, 64); err == nil {
	// 	startBlockHeight = lastBlockHeight
	// }
	//
	// if startBlockHeight > 0 {
	// 	x.startBlockHeight = startBlockHeight
	// } else {
	// 	log.Warn("Found invalid start block number, updating to current block number")
	// 	x.startBlockHeight = x.currentBlockHeight
	// }
	//
	// log.Info("[BURN MONITOR] Start block number: ", x.startBlockHeight)
}

func NewMessageMonitor(config models.EthereumNetworkConfig, lastHealth models.ChainServiceHealth) service.Runner {

	name := strings.ToUpper(fmt.Sprintf("%s_Monitor", config.ChainName))

	if !config.MessageMonitor.Enabled {
		log.Fatalf("[%s] Message monitor is not enabled", name)
	}

	log.Debugf("[%s] Initializing", name)

	client, err := eth.NewClient(config)
	if err != nil {
		log.Fatalf("[%s] Error creating ethereum client: %s", name, err)
	}
	// log.Debug("[BURN MONITOR] Connecting to wpokt contract at: ", app.Config.Ethereum.WrappedPocketAddress)
	// contract, err := autogen.NewWrappedPocket(common.HexToAddress(app.Config.Ethereum.WrappedPocketAddress), client.GetClient())
	// if err != nil {
	// 	log.Fatal("[BURN MONITOR] Error connecting to wpokt contract: ", err)
	// }
	//
	// log.Debug("[BURN MONITOR] Connected to wpokt contract")

	x := &MessageMonitorRunner{
		name:               name,
		startBlockHeight:   0,
		currentBlockHeight: 0,
		// wpoktContract:      eth.NewWrappedPocketContract(contract),
		client: client,
		// minimumAmount:      big.NewInt(app.Config.Pocket.TxFee),
	}

	x.UpdateCurrentBlockHeight()

	// x.InitStartBlockHeight(lastHealth)

	log.Infof("[%s] Initialized", name)

	return x
}
