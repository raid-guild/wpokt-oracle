package ethereum

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

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

	ethClientMap map[uint32]eth.EthereumClient
	mailboxMap   map[uint32]eth.MailboxContract

	mintController eth.MintControllerContract
	warpISM        eth.WarpISMContract

	client        eth.EthereumClient
	confirmations uint64

	cosmosConfig models.CosmosNetworkConfig
	cosmosClient cosmos.CosmosClient

	chain models.Chain

	privateKey *ecdsa.PrivateKey

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

	fmt.Println("len(messageDoc.Signatures): ", len(messageDoc.Signatures))
	fmt.Println("x.numSigners: ", x.numSigners)

	if len(messageDoc.Signatures) >= int(x.numSigners) {
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

	supportedChainIDsEthereum := map[uint32]bool{
		uint32(x.chain.ChainDomain): true,
	}

	result, err := cosmosUtil.ValidateCosmosTx(txResponse, x.cosmosConfig, supportedChainIDsEthereum)

	if err != nil {
		logger.WithError(err).Errorf("Error validating tx response")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if result.TxStatus != models.TransactionStatusPending {
		logger.Debugf("Found tx with status %s", result.TxStatus)
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if result.NeedsRefund {
		logger.Debugf("Found tx with needs refund")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if messageDoc.Content.MessageBody.Amount != result.Amount.Amount.Uint64() {
		logger.Debugf("Found tx with amount mismatch")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if !strings.EqualFold(messageDoc.Content.MessageBody.SenderAddress, common.HexFromBytes(result.SenderAddress)) {
		logger.Debugf("Found tx with sender mismatch")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if !strings.EqualFold(messageDoc.Content.MessageBody.RecipientAddress, result.Memo.Address) {
		logger.Debugf("Found tx with recipient mismatch")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	confirmations := x.currentCosmosBlockHeight - uint64(txResponse.Height)

	if confirmations < x.cosmosConfig.Confirmations {
		logger.WithField("confirmations", confirmations).Debugf("Found tx with not enough confirmations")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusPending})
	}

	log.Debugf("Found valid tx with message to be signed")

	return x.SignMessage(messageDoc)
}

func (x *MessageSignerRunner) ValidateEthereumTxAndSignMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-ethereum-message")
	logger.Debugf("Signing ethereum message")

	ethClient, ok := x.ethClientMap[messageDoc.Content.OriginDomain]

	if !ok {
		logger.Errorf("Ethereum client not found")
		return false
	}

	receipt, err := ethClient.GetTransactionReceipt(messageDoc.OriginTransactionHash)
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

	messageIDBytes, err := common.BytesFromHex(messageDoc.MessageID)
	if err != nil {
		logger.WithError(err).Errorf("Error decoding message ID")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	confirmations := x.currentEthereumBlockHeight - receipt.BlockNumber.Uint64()
	if confirmations < uint64(x.confirmations) {
		logger.WithField("confirmations", confirmations).Debugf("Found tx with not enough confirmations")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusPending})
	}

	mailbox, ok := x.mailboxMap[messageDoc.Content.OriginDomain]
	if !ok {
		logger.Errorf("Mailbox not found")
		return false
	}

	var dispatchEvent *autogen.MailboxDispatchId
	for _, log := range receipt.Logs {
		if log.Address == mailbox.Address() {
			event, err := mailbox.ParseDispatchId(*log)
			if err != nil {
				logger.WithError(err).Errorf("Error parsing DispatchId event")
				continue
			}
			if !bytes.Equal(event.MessageId[:], messageIDBytes) {
				logger.Infof("Message does not match")
				continue
			}
			dispatchEvent = event
			break
		}
	}
	if dispatchEvent == nil {
		logger.Debugf("DispatchId event not found")
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
	messages, err := db.GetPendingMessages(common.Ensure0xPrefix(addressHex), x.chain)

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
	ethNetworks []models.EthereumNetworkConfig,
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

	ethClientMap := make(map[uint32]eth.EthereumClient)
	mailboxMap := make(map[uint32]eth.MailboxContract)

	for _, ethConfig := range ethNetworks {
		var ethClient eth.EthereumClient
		var err error
		if ethConfig.ChainID == config.ChainID {
			ethClient = client
		} else {
			ethClient, err = eth.NewClient(ethConfig)
			if err != nil {
				logger.WithError(err).
					Fatalf("Error creating ethereum client for chain ID %s", ethConfig.ChainID)
			}
		}
		mailbox, err := eth.NewMailboxContract(common.HexToAddress(ethConfig.MailboxAddress), ethClient.GetClient())
		if err != nil {
			logger.WithError(err).
				Fatalf("Error creating mailbox contract for chain ID %s", ethConfig.ChainID)
		}
		chainDomain := ethClient.Chain().ChainDomain
		ethClientMap[chainDomain] = ethClient
		mailboxMap[chainDomain] = mailbox
	}

	x := &MessageSignerRunner{
		timeout: time.Duration(config.TimeoutMS) * time.Millisecond,

		currentEthereumBlockHeight: 0,

		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,

		mintController: mintController,
		warpISM:        warpISM,

		privateKey: privateKey,

		chain: util.ParseChain(config),

		client:        client,
		confirmations: uint64(config.Confirmations),

		cosmosClient: cosmosClient,
		cosmosConfig: cosmosConfig,

		numSigners: 0,

		logger: logger,
	}

	x.UpdateCurrentBlockHeight()

	x.UpdateValidatorCount()

	if x.numSigners != int64(len(config.OracleAddresses)) {
		x.logger.Fatalf("Invalid number of signers")
	}

	x.UpdateDomainData()

	chainID := big.NewInt(int64(config.ChainID))

	if x.domain.ChainId.Cmp(chainID) != 0 {
		x.logger.Fatalf("Invalid chain ID")
	}

	if !strings.EqualFold(x.domain.VerifyingContract.Hex(), config.WarpISMAddress) {
		x.logger.Fatalf("Invalid verifying address in domain data")
	}

	x.UpdateMaxMintLimit()

	logger.Infof("Initialized")

	return x
}
