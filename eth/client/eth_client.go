package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"math/big"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	log "github.com/sirupsen/logrus"
)

const (
	MaxQueryBlocks int64 = 100000
)

type EthereumClient interface {
	ValidateNetwork() error
	GetBlockHeight() (uint64, error)
	GetChainID() (*big.Int, error)
	GetClient() *ethclient.Client
	GetTransactionByHash(txHash string) (*types.Transaction, bool, error)
	GetTransactionReceipt(txHash string) (*types.Receipt, error)
}

type ethereumClient struct {
	Timeout   time.Duration
	ChainID   int64
	ChainName string

	name   string
	client *ethclient.Client
}

func (c *ethereumClient) GetClient() *ethclient.Client {
	return c.client
}

func (c *ethereumClient) GetBlockHeight() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

func (c *ethereumClient) GetChainID() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	chainId, err := c.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	return chainId, nil
}

func (c *ethereumClient) ValidateNetwork() error {
	log.Debugf("[%s] Validating network", c.name)

	chainID, err := c.GetChainID()
	if err != nil {
		return fmt.Errorf("failed to validate network: %s", err)
	}
	if chainID.Cmp(big.NewInt(c.ChainID)) != 0 {
		return fmt.Errorf("failed to validate network: expected chain id %d, got %s", c.ChainID, chainID)
	}

	log.Debugf("[%s] Network validated", c.name)
	return nil
}

func (c *ethereumClient) GetTransactionByHash(txHash string) (*types.Transaction, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	tx, isPending, err := c.client.TransactionByHash(ctx, common.HexToHash(txHash))
	return tx, isPending, err
}

func (c *ethereumClient) GetTransactionReceipt(txHash string) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	receipt, err := c.client.TransactionReceipt(ctx, common.HexToHash(txHash))
	return receipt, err
}

func NewClient(config models.EthereumNetworkConfig) (EthereumClient, error) {
	client, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rpc: %s", err)
	}

	ethclient := &ethereumClient{
		Timeout:   time.Duration(config.TimeoutMS) * time.Millisecond,
		ChainID:   config.ChainID,
		ChainName: config.ChainName,

		name:   strings.ToUpper(fmt.Sprintf("%s_CLIENT", config.ChainName)),
		client: client,
	}

	err = ethclient.ValidateNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to validate network: %s", err)
	}

	return ethclient, err
}
