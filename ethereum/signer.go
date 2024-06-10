package ethereum

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	timeout time.Duration

	currentEthereumBlockHeight uint64
	currentCosmosBlockHeight   uint64

	mintControllerMap map[uint32][]byte

	mailbox        eth.MailboxContract
	mintController eth.MintControllerContract
	warpISM        eth.WarpISMContract

	client        eth.EthereumClient
	confirmations uint64

	cosmosConfig models.CosmosNetworkConfig
	cosmosClient cosmos.CosmosClient

	chain models.Chain

	privateKey *ecdsa.PrivateKey

	minimumAmount *big.Int

	// TODO: validate maximumAmount
	maximumAmount *big.Int

	numSigners int64
	domain     util.DomainData

	logger *log.Entry
}

func (x *MessageSignerRunner) Run() {
	x.UpdateCurrentBlockHeight()
	x.UpdateMaxMintLimit()
	x.SignMessages()
}

func (x *MessageSignerRunner) Height() uint64 {
	return uint64(x.currentEthereumBlockHeight)
}

func (x *MessageSignerRunner) UpdateCurrentCosmosBlockHeight() {
	height, err := x.cosmosClient.GetLatestBlockHeight()
	if err != nil {
		x.logger.
			WithError(err).
			Error("could not get current cosmos block height")
		return
	}
	x.currentCosmosBlockHeight = uint64(height)
	x.logger.
		WithField("current_block_height", x.currentCosmosBlockHeight).
		Info("updated current cosmos block height")
}

func (x *MessageSignerRunner) UpdateCurrentEthereumBlockHeight() {
	res, err := x.client.GetBlockHeight()
	if err != nil {
		x.logger.
			WithError(err).
			Error("could not get current ethereum block height")
		return
	}
	x.currentEthereumBlockHeight = res
	x.logger.
		WithField("current_ethereum_block_height", x.currentEthereumBlockHeight).
		Info("updated current ethereum block height")
}

func (x *MessageSignerRunner) UpdateCurrentBlockHeight() {
	x.UpdateCurrentEthereumBlockHeight()
	x.UpdateCurrentCosmosBlockHeight()
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

	err := util.SignMessage(messageDoc, x.domain, x.privateKey)
	if err != nil {
		logger.WithError(err).Errorf("Error signing message")
		return false
	}

	logger.Infof("Signed message")

	status := models.MessageStatusPending

	if len(messageDoc.Signatures) == int(x.numSigners) {
		status = models.MessageStatusSigned
	}

	update := bson.M{
		"status":     status,
		"signatures": messageDoc.Signatures,
	}

	return x.UpdateMessage(messageDoc, update)
}

func (x *MessageSignerRunner) ValidateCosmosTxAndSignMessage(messageDoc *models.Message) bool {
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

	feeAmount := sdk.NewCoin("upokt", math.NewInt(x.minimumAmount.Int64()))

	if coinsReceived.IsLTE(feeAmount) {
		logger.
			Debugf("Found tx with amount too low")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	txHeight := txResponse.Height
	if txHeight <= 0 || uint64(txHeight) > x.currentCosmosBlockHeight {
		logger.WithField("tx_height", txHeight).Debugf("Found tx with invalid height")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	confirmations := x.currentCosmosBlockHeight - uint64(txHeight)

	if confirmations < x.cosmosConfig.Confirmations {
		logger.WithField("confirmations", confirmations).Debugf("Found tx with not enough confirmations")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusPending})

	}

	if !coinsSpent.Amount.Equal(coinsReceived.Amount) {
		logger.Debugf("Found tx with invalid coins")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	supportedChainIDs := make(map[uint64]bool)
	supportedChainIDs[uint64(x.chain.ChainDomain)] = true

	memo, err := cosmosUtil.ValidateMemo(tx.Body.Memo, supportedChainIDs)
	if err != nil {
		logger.WithError(err).WithField("memo", tx.Body.Memo).Debugf("Found invalid memo")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})

	}

	logger.WithField("memo", memo).Errorf("Found message with a valid memo")
	return x.SignMessage(messageDoc)
}

func (x *MessageSignerRunner) ValidateEthereumTxAndSignMessage(messageDoc *models.Message) bool {
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

	confirmations := x.currentEthereumBlockHeight - receipt.BlockNumber.Uint64()
	if confirmations < uint64(x.confirmations) {
		logger.WithField("confirmations", confirmations).Debugf("Found tx with not enough confirmations")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusPending})
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

		if messageDoc.Content.OriginDomain == x.cosmosClient.Chain().ChainDomain {
			success = x.ValidateCosmosTxAndSignMessage(&messageDoc) && success
			continue
		}

		success = x.ValidateEthereumTxAndSignMessage(&messageDoc) && success
	}

	return success
}

func (x *MessageSignerRunner) UpdateValidatorCount() {
	x.logger.Debug("Fetching validator count")
	ctx, cancel := context.WithTimeout(context.Background(), x.timeout)
	defer cancel()
	opts := &bind.CallOpts{Context: ctx, Pending: false}
	count, err := x.warpISM.ValidatorCount(opts)

	if err != nil {
		x.logger.WithError(err).Error("Error fetching validator count")
		return
	}
	x.logger.Debug("Fetched validator count")
	x.numSigners = count.Int64()
}

func (x *MessageSignerRunner) UpdateDomainData() {
	x.logger.Debug("Fetching domain data")
	ctx, cancel := context.WithTimeout(context.Background(), x.timeout)
	defer cancel()
	opts := &bind.CallOpts{Context: ctx, Pending: false}
	domain, err := x.warpISM.Eip712Domain(opts)

	if err != nil {
		x.logger.WithError(err).Error("Error fetching domain data")
		return
	}
	x.logger.Debug("Fetched domain data")
	x.domain = domain
}

func (x *MessageSignerRunner) UpdateMaxMintLimit() {
	x.logger.Debug("Fetching max mint limit")
	ctx, cancel := context.WithTimeout(context.Background(), x.timeout)
	defer cancel()
	opts := &bind.CallOpts{Context: ctx, Pending: false}
	mintLimit, err := x.mintController.MaxMintLimit(opts)

	if err != nil {
		x.logger.WithError(err).Error("Error fetching max mint limit")
		return
	}
	x.logger.Debug("Fetched max mint limit")
	x.maximumAmount = mintLimit
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

	logger.Debug("Connecting to mint controller contract at: ", config.MintControllerAddress)
	mintController, err := eth.NewMintControllerContract(common.HexToAddress(config.MintControllerAddress), client.GetClient())
	if err != nil {
		logger.Fatal("Error connecting to mint controller contract: ", err)
	}
	logger.Debug("Connected to mint controller contract")

	logger.Debug("Connecting to warp ism contract at: ", config.WarpISMAddress)
	warpISM, err := eth.NewWarpISMContract(common.HexToAddress(config.WarpISMAddress), client.GetClient())
	if err != nil {
		logger.Fatal("Error connecting to warp ism contract: ", err)
	}
	logger.Debug("Connected to warp ism contract")

	privateKey, err := common.EthereumPrivateKeyFromMnemonic(mnemonic)
	if err != nil {
		logger.Fatalf("Error getting private key from mnemonic: %s", err)
	}

	cosmosClient, err := cosmos.NewClient(cosmosConfig)
	if err != nil {
		logger.Fatalf("Error creating cosmos client: %s", err)
	}

	minimumAmount := big.NewInt(int64(cosmosConfig.TxFee))

	x := &MessageSignerRunner{
		timeout: time.Duration(config.TimeoutMS) * time.Millisecond,

		currentEthereumBlockHeight: 0,

		mintControllerMap: mintControllerMap,

		mailbox:        mailbox,
		mintController: mintController,
		warpISM:        warpISM,

		privateKey: privateKey,

		chain: util.ParseChain(config),

		client:        client,
		confirmations: uint64(config.Confirmations),

		cosmosClient: cosmosClient,
		cosmosConfig: cosmosConfig,

		numSigners:    0,
		minimumAmount: minimumAmount,

		logger: logger,
	}

	x.UpdateCurrentBlockHeight()

	x.UpdateValidatorCount()

	if x.numSigners != int64(len(config.OracleAddresses)) {
		x.logger.Fatalf("Invalid number of signers")
	}

	x.UpdateDomainData()

	chainId := big.NewInt(int64(config.ChainID))

	if x.domain.ChainId.Cmp(chainId) != 0 {
		x.logger.Fatalf("Invalid chain ID")
	}

	if !strings.EqualFold(x.domain.VerifyingContract.Hex(), config.WarpISMAddress) {
		x.logger.Fatalf("Invalid verifying address in domain data")
	}

	x.UpdateMaxMintLimit()

	if x.maximumAmount == nil || x.maximumAmount.Cmp(x.minimumAmount) != 1 {
		x.logger.Fatalf("Invalid max mint limit")
	}

	logger.Infof("Initialized")

	return x
}
