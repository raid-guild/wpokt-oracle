package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"

	log "github.com/sirupsen/logrus"
)

const (
	MaxQueryBlocks uint64 = 100000
)

type EthereumClient interface {
	Chain() models.Chain
	Confirmations() uint64
	ValidateNetwork() error
	GetBlockHeight() (uint64, error)
	GetChainID() (*big.Int, error)
	GetClient() EthclientClient
	GetTransactionByHash(txHash string) (*types.Transaction, bool, error)
	GetTransactionReceipt(txHash string) (*types.Receipt, error)
}

type EthclientClient interface {
	bind.ContractBackend
	BlockNumber(ctx context.Context) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type ethereumClient struct {
	chain         models.Chain
	confirmations uint64

	timeout   time.Duration
	chainID   uint64
	chainName string

	client EthclientClient

	logger *log.Entry
}

func (c *ethereumClient) Chain() models.Chain {
	return c.chain
}

func (c *ethereumClient) Confirmations() uint64 {
	return c.confirmations
}

func (c *ethereumClient) GetClient() EthclientClient {
	return c.client
}

func (c *ethereumClient) GetBlockHeight() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

func (c *ethereumClient) GetChainID() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	return chainID, nil
}

func (c *ethereumClient) ValidateNetwork() error {
	c.logger.Debugf("Validating network")

	chainID, err := c.GetChainID()
	if err != nil {
		return err
	}
	if chainID.Cmp(big.NewInt(int64(c.chainID))) != 0 {
		return fmt.Errorf("expected chain id %d, got %s", c.chainID, chainID)
	}

	c.logger.Debugf("Validated network")
	return nil
}

func (c *ethereumClient) GetTransactionByHash(txHash string) (*types.Transaction, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	tx, isPending, err := c.client.TransactionByHash(ctx, common.HexToHash(txHash))
	return tx, isPending, err
}

func (c *ethereumClient) GetTransactionReceipt(txHash string) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	receipt, err := c.client.TransactionReceipt(ctx, common.HexToHash(txHash))
	return receipt, err
}

type EthclientDial func(url string) (EthclientClient, error)

var ethclientDial EthclientDial = func(url string) (EthclientClient, error) {
	return ethclient.Dial(url)
}

func NewClient(config models.EthereumNetworkConfig) (EthereumClient, error) {
	logger := log.
		WithField("module", "ethereum").
		WithField("package", "client").
		WithField("chain_name", strings.ToLower(config.ChainName)).
		WithField("chain_id", config.ChainID)
	client, err := ethclientDial(config.RPCURL)
	if err != nil {
		logger.WithError(err).Error("failed to connect to rpc")
		return nil, fmt.Errorf("failed to connect to rpc")
	}

	ethclient := &ethereumClient{
		chain: util.ParseChain(config),

		timeout:       time.Duration(config.TimeoutMS) * time.Millisecond,
		chainID:       config.ChainID,
		chainName:     config.ChainName,
		confirmations: config.Confirmations,

		client: client,

		logger: logger,
	}

	err = ethclient.ValidateNetwork()
	if err != nil {
		logger.WithError(err).Error("failed to validate network")
		return nil, fmt.Errorf("failed to validate network")
	}

	return ethclient, err
}
