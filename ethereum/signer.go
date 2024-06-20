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

type EthMessageSignerRunnable struct {
	timeout time.Duration

	currentEthereumBlockHeight uint64
	currentCosmosBlockHeight   uint64

	ethClientMap map[uint32]eth.EthereumClient
	mailboxMap   map[uint32]eth.MailboxContract

	mintController eth.MintControllerContract
	warpISM        eth.WarpISMContract

	client eth.EthereumClient

	cosmosConfig models.CosmosNetworkConfig
	cosmosClient cosmos.CosmosClient

	chain models.Chain

	privateKey *ecdsa.PrivateKey

	// TODO: validate maximumAmount
	maximumAmount *big.Int

	numSigners      int64
	signerThreshold int64
	domain          util.DomainData

	logger *log.Entry
}

func (x *EthMessageSignerRunnable) Run() {
	x.UpdateCurrentBlockHeight()
	x.UpdateMaxMintLimit()
	x.SignMessages()
}

func (x *EthMessageSignerRunnable) Height() uint64 {
	return uint64(x.currentEthereumBlockHeight)
}

func (x *EthMessageSignerRunnable) UpdateCurrentCosmosBlockHeight() {
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

func (x *EthMessageSignerRunnable) UpdateCurrentEthereumBlockHeight() {
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

func (x *EthMessageSignerRunnable) UpdateCurrentBlockHeight() {
	x.UpdateCurrentEthereumBlockHeight()
	x.UpdateCurrentCosmosBlockHeight()
}

func (x *EthMessageSignerRunnable) UpdateMessage(
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

func (x *EthMessageSignerRunnable) SignMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-message")
	logger.Debugf("Signing message")

	if lockID, err := db.LockWriteMessage(messageDoc); err != nil {
		logger.WithError(err).Errorf("Error locking message")
		return false
	} else {
		defer db.Unlock(lockID)
	}

	if err := util.SignMessage(messageDoc, x.domain, x.privateKey); err != nil {
		logger.WithError(err).Errorf("Error signing message")
		return false
	}

	update := bson.M{
		"status":     models.MessageStatusPending,
		"signatures": messageDoc.Signatures,
	}
	if len(messageDoc.Signatures) >= int(x.signerThreshold) {
		update["status"] = models.MessageStatusSigned
	}

	logger.Debugf("Message signed")
	return x.UpdateMessage(messageDoc, update)
}

func (x *EthMessageSignerRunnable) ValidateCosmosMessage(messageDoc *models.Message) (confirmed bool, err error) {
	txResponse, err := x.cosmosClient.GetTx(messageDoc.OriginTransactionHash)
	if err != nil {
		return false, fmt.Errorf("error getting tx: %w", err)
	}

	supportedChainIDsEthereum := map[uint32]bool{uint32(x.chain.ChainDomain): true}

	result, err := cosmosUtil.ValidateTxToCosmosMultisig(txResponse, x.cosmosConfig, supportedChainIDsEthereum, x.currentCosmosBlockHeight)
	if err != nil {
		return false, fmt.Errorf("error validating tx response: %w", err)
	}

	if result.NeedsRefund {
		return false, fmt.Errorf("tx needs refund")
	}

	amount, ok := new(big.Int).SetString(messageDoc.Content.MessageBody.Amount, 10)

	if ok && amount.Cmp(result.Amount.Amount.BigInt()) != 0 {
		return false, fmt.Errorf("amount mismatch")
	}

	if !strings.EqualFold(messageDoc.Content.MessageBody.SenderAddress, common.HexFromBytes(result.SenderAddress)) {
		return false, fmt.Errorf("sender mismatch")
	}

	if !strings.EqualFold(messageDoc.Content.MessageBody.RecipientAddress, result.Memo.Address) {
		return false, fmt.Errorf("recipient mismatch")
	}

	if result.TxStatus == models.TransactionStatusPending {
		return false, nil
	}

	if result.TxStatus != models.TransactionStatusConfirmed {
		return false, fmt.Errorf("tx is invalid")
	}

	return true, nil
}

func (x *EthMessageSignerRunnable) ValidateCosmosTxAndSignMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-cosmos-message")
	logger.Debugf("Signing cosmos message")

	confirmed, err := x.ValidateCosmosMessage(messageDoc)

	if err != nil {
		logger.WithError(err).Errorf("Error validating cosmos message")
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	if !confirmed {
		logger.Debugf("Found pending tx")
		return false
	}

	return x.SignMessage(messageDoc)
}

type ValidateTransactionAndParseDispatchIdEventsResult struct {
	Event         *autogen.MailboxDispatchId
	Confirmations uint64
	TxStatus      models.TransactionStatus
}

func (x *EthMessageSignerRunnable) ValidateAndFindDispatchIdEvent(messageDoc *models.Message) (*ValidateTransactionAndParseDispatchIdEventsResult, error) {
	chainDomain := messageDoc.Content.OriginDomain
	txHash := messageDoc.OriginTransactionHash
	messageIDBytes, err := common.BytesFromHex(messageDoc.MessageID)
	if err != nil {
		return nil, fmt.Errorf("error getting message ID bytes: %w", err)
	}

	ethClient, ok := x.ethClientMap[chainDomain]
	if !ok {
		return nil, fmt.Errorf("ethereum client not found")
	}
	mailbox, ok := x.mailboxMap[chainDomain]
	if !ok {
		return nil, fmt.Errorf("mailbox not found")
	}

	receipt, err := ethClient.GetTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction receipt: %w", err)
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return &ValidateTransactionAndParseDispatchIdEventsResult{
			TxStatus: models.TransactionStatusFailed,
		}, nil
	}
	var dispatchEvent *autogen.MailboxDispatchId
	for _, log := range receipt.Logs {
		if log.Address == mailbox.Address() {
			event, err := mailbox.ParseDispatchId(*log)
			if err != nil {
				continue
			}
			if bytes.Equal(event.MessageId[:], messageIDBytes) {
				dispatchEvent = event
				break
			}
		}
	}

	currentBlockHeight, err := ethClient.GetBlockHeight()
	if err != nil {
		return nil, fmt.Errorf("error getting current block height: %w", err)
	}

	result := &ValidateTransactionAndParseDispatchIdEventsResult{
		Event:         dispatchEvent,
		Confirmations: currentBlockHeight - receipt.BlockNumber.Uint64(),
		TxStatus:      models.TransactionStatusPending,
	}
	if result.Confirmations >= ethClient.Confirmations() {
		result.TxStatus = models.TransactionStatusConfirmed
	}
	if dispatchEvent == nil {
		result.TxStatus = models.TransactionStatusInvalid
	}
	if messageIdFromContent, err := common.BytesFromHex(messageDoc.MessageID); err != nil || !bytes.Equal(messageIdFromContent, dispatchEvent.MessageId[:]) {
		result.TxStatus = models.TransactionStatusInvalid
	}
	return result, nil
}

func (x *EthMessageSignerRunnable) ValidateEthereumTxAndSignMessage(messageDoc *models.Message) bool {
	logger := x.logger.WithField("tx_hash", messageDoc.OriginTransactionHash).WithField("section", "sign-ethereum-message")
	logger.Debugf("Signing ethereum message")

	result, err := x.ValidateAndFindDispatchIdEvent(messageDoc)
	if err != nil {
		x.logger.WithError(err).Error("Error validating transaction and parsing DispatchId events")
		return false
	}

	if result.TxStatus == models.TransactionStatusPending {
		logger.Debugf("Found pending tx")
		return false
	}

	if result.TxStatus != models.TransactionStatusConfirmed {
		logger.Debugf("Found tx with status %s", result.TxStatus)
		return x.UpdateMessage(messageDoc, bson.M{"status": models.MessageStatusInvalid})
	}

	return x.SignMessage(messageDoc)
}

func (x *EthMessageSignerRunnable) SignMessages() bool {
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

func (x *EthMessageSignerRunnable) UpdateValidatorCountAndSignerThreshold() {
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

	x.logger.Debug("Fetching signer threshold")
	ctx, cancel = context.WithTimeout(context.Background(), x.timeout)
	defer cancel()
	opts = &bind.CallOpts{Context: ctx, Pending: false}
	threshold, err := x.warpISM.SignerThreshold(opts)
	if err != nil {
		x.logger.WithError(err).Error("Error fetching signer threshold")
		return
	}
	x.logger.Debug("Fetched signer threshold")

	x.numSigners = count.Int64()
	x.signerThreshold = threshold.Int64()
}

func (x *EthMessageSignerRunnable) UpdateDomainData() {
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

func (x *EthMessageSignerRunnable) UpdateMaxMintLimit() {
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
) service.Runnable {
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
	if privateKey == nil {
		logger.Fatalf("Private key is nil")
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
				logger.WithError(err).WithField("eth_chain_id", ethConfig.ChainID).
					Fatalf("Error creating ethereum client")
			}
		}
		mailbox, err := eth.NewMailboxContract(common.HexToAddress(ethConfig.MailboxAddress), ethClient.GetClient())
		if err != nil {
			logger.WithError(err).WithField("eth_chain_id", ethConfig.ChainID).
				Fatalf("Error creating mailbox contract")
		}
		chainDomain := ethClient.Chain().ChainDomain
		ethClientMap[chainDomain] = ethClient
		mailboxMap[chainDomain] = mailbox
	}

	x := &EthMessageSignerRunnable{
		timeout: time.Duration(config.TimeoutMS) * time.Millisecond,

		currentEthereumBlockHeight: 0,

		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,

		mintController: mintController,
		warpISM:        warpISM,

		privateKey: privateKey,

		chain: util.ParseChain(config),

		client: client,

		cosmosClient: cosmosClient,
		cosmosConfig: cosmosConfig,

		numSigners:      0,
		signerThreshold: 0,

		logger: logger,
	}

	x.UpdateCurrentBlockHeight()

	x.UpdateValidatorCountAndSignerThreshold()

	if x.numSigners != int64(len(config.OracleAddresses)) {
		x.logger.Fatalf("Invalid number of signers")
	}

	if x.signerThreshold < 1 || x.signerThreshold > x.numSigners {
		x.logger.Fatalf("Invalid signer threshold")
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
