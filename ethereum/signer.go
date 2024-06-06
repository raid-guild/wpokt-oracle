package ethereum

import (
	"bytes"
	"crypto/ecdsa"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"

	"github.com/dan13ram/wpokt-oracle/common"
	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	cosmosUtil "github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type MessageSignerRunner struct {
	startBlockHeight   uint64
	currentBlockHeight uint64

	mintControllerMap map[uint32][]byte

	mailbox eth.MailboxContract
	client  eth.EthereumClient

	cosmosConfig models.CosmosNetworkConfig
	cosmosClient cosmos.CosmosClient

	chain models.Chain

	privateKey *ecdsa.PrivateKey

	minimumAmount *big.Int

	logger *log.Entry
}

func (x *MessageSignerRunner) Run() {
	x.UpdateCurrentBlockHeight()
	x.SignMessages()
}

func (x *MessageSignerRunner) Height() uint64 {
	return uint64(x.currentBlockHeight)
}

func (x *MessageSignerRunner) UpdateCurrentBlockHeight() {
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

func (x *MessageSignerRunner) UpdateMessage(
	message *models.Message,
	update bson.M,
) bool {
	err := db.UpdateMessage(message.ID, update)
	if err != nil {
		x.logger.WithError(err).Errorf("Error updating message")
		return false
	}
	return true
}

func (x *MessageSignerRunner) SignMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-message")
	logger.Debugf("Signing message")
	return false
}

func (x *MessageSignerRunner) ValidateAndSignCosmosMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-cosmos-message")
	logger.Debugf("Signing cosmos message")

	txResponse, err := x.cosmosClient.GetTx(messageDoc.OriginTransactionHash)
	if err != nil {
		logger.WithError(err).Errorf("Error getting tx")
		return false
	}
	if txResponse.Code != 0 {
		logger.Infof("Found tx with error")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	tx := &tx.Tx{}
	err = tx.Unmarshal(txResponse.Tx.Value)
	if err != nil {
		logger.Errorf("Error unmarshalling tx")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	coinsReceived, err := cosmosUtil.ParseCoinsReceivedEvents(x.cosmosConfig.CoinDenom, x.cosmosConfig.MultisigAddress, txResponse.Events)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing coins received events")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	_, coinsSpent, err := cosmosUtil.ParseCoinsSpentEvents(x.cosmosConfig.CoinDenom, txResponse.Events)
	if err != nil {
		logger.WithError(err).Errorf("Error parsing coins spent events")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	if coinsReceived.IsZero() || coinsSpent.IsZero() {
		logger.
			Debugf("Found tx with zero coins")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	feeAmount := sdk.NewCoin("upokt", math.NewInt(int64(x.cosmosConfig.TxFee)))

	if coinsReceived.IsLTE(feeAmount) {
		logger.
			Debugf("Found tx with amount too low")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	txHeight := txResponse.Height
	if txHeight <= 0 || uint64(txHeight) > x.currentBlockHeight {
		logger.WithField("tx_height", txHeight).Debugf("Found tx with invalid height")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	confirmations := x.currentBlockHeight - uint64(txHeight)

	if confirmations < x.cosmosConfig.Confirmations {
		logger.WithField("confirmations", confirmations).Debugf("Found tx with not enough confirmations")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusPending})

	}

	if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
		logger.Debugf("Found tx with invalid coins")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	memo, err := cosmosUtil.ValidateMemo(tx.Body.Memo)
	if err != nil {
		logger.WithError(err).WithField("memo", tx.Body.Memo).Debugf("Found invalid memo")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	logger.WithField("memo", memo).Errorf("Found message with a valid memo")
	return x.SignMessage(messageDoc)
}

func (x *MessageSignerRunner) ValidateAndSignEthereumMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-ethereum-message")
	logger.Debugf("Signing ethereum message")

	receipt, err := x.client.GetTransactionReceipt(messageDoc.OriginTransactionHash)
	if err != nil {
		logger.WithError(err).Errorf("Error getting tx receipt")
		return false
	}

	if receipt == nil {
		logger.Infof("Tx receipt not found")
		return false
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		logger.Infof("Tx receipt failed")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	messageBytes, err := messageDoc.Content.EncodeToBytes()
	if err != nil {
		logger.WithError(err).Errorf("Error encoding message to bytes")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	mintController, ok := x.mintControllerMap[x.chain.ChainDomain]
	if !ok {
		logger.Infof("Mint controller not found")
		return false
	}

	var dispatchEvent *autogen.MailboxDispatch
	for _, log := range receipt.Logs {
		if log.Address == x.mailbox.Address() {
			event, err := x.mailbox.ParseDispatch(*log)
			if err != nil {
				logger.WithError(err).Errorf("Error parsing dispatch event")
				continue
			}
			if event.Destination != x.chain.ChainDomain {
				logger.Infof("Event destination is not this chain")
				continue
			}
			if !bytes.Equal(event.Recipient[12:32], mintController) {
				logger.Infof("Event recipient is not mint controller")
				continue
			}
			if !bytes.Equal(event.Message, messageBytes) {
				logger.Infof("Message does not match")
				continue
			}
			dispatchEvent = event
			break
		}
	}
	if dispatchEvent == nil {
		logger.Infof("Dispatch event not found")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	return x.SignMessage(messageDoc)
}

func (x *MessageSignerRunner) SignMessages() bool {
	x.logger.Infof("Signing messages")
	addressHex, err := common.EthereumPrivateKeyToAddressHex(x.privateKey)
	if err != nil {
		x.logger.WithError(err).Errorf("Error getting address hex")
	}
	messages, err := db.GetPendingMessages(addressHex, x.chain)

	if err != nil {
		x.logger.WithError(err).Errorf("Error getting pending messages")
		return false
	}
	x.logger.Infof("Found %d pending messages", len(messages))
	success := true
	for _, messageDoc := range messages {

		if messageDoc.Content.DestinationDomain == x.cosmosClient.Chain().ChainDomain {
			success = success && x.ValidateAndSignCosmosMessage(&messageDoc)
			continue
		}

		success = success && x.ValidateAndSignEthereumMessage(&messageDoc)
	}

	return true
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

func NewMessageSigner(
	mnemonic string,
	config models.EthereumNetworkConfig,
	cosmosConfig models.CosmosNetworkConfig,
	mintControllerMap map[uint32][]byte,
) service.Runner {
	logger := log.
		WithField("module", "ethereum").
		WithField("service", "signer").
		WithField("chain_name", strings.ToLower(config.ChainName)).
		WithField("chain_id", config.ChainID)

	if !config.MessageSigner.Enabled {
		logger.Fatalf("Message signer is not enabled")
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

	privateKey, err := common.EthereumPrivateKeyFromMnemonic(mnemonic)
	if err != nil {
		logger.Fatalf("Error getting private key from mnemonic: %s", err)
	}

	cosmosClient, err := cosmos.NewClient(cosmosConfig)
	if err != nil {
		logger.Fatalf("Error creating cosmos client: %s", err)
	}

	x := &MessageSignerRunner{
		startBlockHeight:   config.StartBlockHeight,
		currentBlockHeight: 0,

		mintControllerMap: mintControllerMap,

		mailbox:    mailbox,
		privateKey: privateKey,

		chain: util.ParseChain(config),

		client:       client,
		cosmosClient: cosmosClient,
		cosmosConfig: cosmosConfig,

		logger: logger,
	}

	x.UpdateCurrentBlockHeight()

	logger.Infof("Initialized")

	return x
}
