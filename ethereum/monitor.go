package ethereum

import (
	"math/big"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type MessageMonitorRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64
	wpoktContract      eth.WrappedPocketContract
	client             eth.EthereumClient
	minimumAmount      *big.Int

	logger *log.Entry
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

func (x *MessageMonitorRunner) HandleBurnEvent(event *autogen.WrappedPocketBurnAndBridge) bool {
	// if event == nil {
	// 	x.logger.Error("[BURN MONITOR] Error while handling burn event: event is nil")
	// 	return false
	// }
	//
	// doc := util.CreateBurn(event)
	//
	// // each event is a combination of transaction hash and log index
	// x.logger.Debug("[BURN MONITOR] Handling burn event: ", event.Raw.TxHash, " ", event.Raw.Index)
	//
	// err := app.DB.InsertOne(models.CollectionBurns, doc)
	// if err != nil {
	// 	if mongo.IsDuplicateKeyError(err) {
	// 		x.logger.Info("[BURN MONITOR] Found duplicate burn event: ", event.Raw.TxHash, " ", event.Raw.Index)
	// 		return true
	// 	}
	// 	x.logger.Error("[BURN MONITOR] Error while storing burn event in db: ", err)
	// 	return false
	// }
	//
	// x.logger.Info("[BURN MONITOR] Stored burn event: ", event.Raw.TxHash, " ", event.Raw.Index)
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
	// 	x.logger.Error("[BURN MONITOR] Error while syncing burn events: ", err)
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
	// 	x.logger.Error("[BURN MONITOR] Error while syncing burn events: ", err)
	// 	return false
	// }

	return success
}

func (x *MessageMonitorRunner) SyncTxs() bool {
	if x.currentBlockHeight <= x.startBlockHeight {
		x.logger.Infof("No new blocks to sync")
		return true
	}

	var success bool = true
	// if (x.currentBlockHeight - x.startBlockHeight) > eth.MAX_QUERY_BLOCKS {
	// 	x.logger.Debug("[BURN MONITOR] Syncing burn txs in chunks")
	// 	for i := x.startBlockHeight; i < x.currentBlockHeight; i += eth.MAX_QUERY_BLOCKS {
	// 		endBlockHeight := i + eth.MAX_QUERY_BLOCKS
	// 		if endBlockHeight > x.currentBlockHeight {
	// 			endBlockHeight = x.currentBlockHeight
	// 		}
	// 		x.logger.Info("[BURN MONITOR] Syncing burn txs from blockNumber: ", i, " to blockNumber: ", endBlockHeight)
	// 		success = success && x.SyncBlocks(uint64(i), uint64(endBlockHeight))
	// 	}
	// } else {
	// 	x.logger.Info("[BURN MONITOR] Syncing burn txs from blockNumber: ", x.startBlockHeight, " to blockNumber: ", x.currentBlockHeight)
	// 	success = success && x.SyncBlocks(uint64(x.startBlockHeight), uint64(x.currentBlockHeight))
	// }
	//
	// if success {
	// 	x.startBlockHeight = x.currentBlockHeight
	// }

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

func NewMessageMonitor(config models.EthereumNetworkConfig, lastHealth *models.RunnerServiceStatus) service.Runner {
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
	// logger.Debug("[BURN MONITOR] Connecting to wpokt contract at: ", app.Config.Ethereum.WrappedPocketAddress)
	// contract, err := autogen.NewWrappedPocket(common.HexToAddress(app.Config.Ethereum.WrappedPocketAddress), client.GetClient())
	// if err != nil {
	// 	logger.Fatal("[BURN MONITOR] Error connecting to wpokt contract: ", err)
	// }
	//
	// logger.Debug("[BURN MONITOR] Connected to wpokt contract")

	x := &MessageMonitorRunner{
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,
		// wpoktContract:      eth.NewWrappedPocketContract(contract),
		client: client,
		// minimumAmount:      big.NewInt(app.Config.Pocket.TxFee),

		logger: logger,
	}

	x.UpdateCurrentBlockHeight()

	x.InitStartBlockHeight(lastHealth)

	logger.Infof("Initialized")

	return x
}
