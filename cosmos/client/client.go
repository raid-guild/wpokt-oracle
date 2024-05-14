package client

import (
	"encoding/hex"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	maxPageDepth = 100
)

type CosmosClient interface {
	GetLatestBlockHeight() (int64, error)
	GetChainID() (string, error)
	GetTxsSentFromAddressAfterHeight(address string, height int64) ([]*sdk.TxResponse, error)
	GetTxsSentToAddressAfterHeight(address string, height int64) ([]*sdk.TxResponse, error)
	// SubmitRawTx(params rpc.SendRawTxParams) (*SubmitRawTxResponse, error)
	GetTx(hash string) (*sdk.TxResponse, error)
	ValidateNetwork() error
}

type cosmosClient struct {
	GRPCEnabled bool

	Timeout       time.Duration
	ChainID      string
	ChainName    string
	Bech32Prefix string
	CoinDenom    string

	name      string
	grpcConn  *grpc.ClientConn
	rpcClient *rpchttp.HTTP
}

func (c *cosmosClient) getLatestBlockGRPC() (*cmtservice.Block, error) {
	if !c.GRPCEnabled {
		return nil, fmt.Errorf("grpc disabled")
	}

	client := cmtservice.NewServiceClient(c.grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req := &cmtservice.GetLatestBlockRequest{}

	resp, err := client.GetLatestBlock(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.SdkBlock, nil
}

func (c *cosmosClient) getStatusRPC() (*rpctypes.ResultStatus, error) {
	if c.GRPCEnabled {
		return nil, fmt.Errorf("grpc enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	res, err := c.rpcClient.Status(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *cosmosClient) GetLatestBlockHeight() (int64, error) {
	if c.GRPCEnabled {
		block, err := c.getLatestBlockGRPC()
		if err != nil {
			return 0, err
		}
		return block.Header.Height, nil
	}

	status, err := c.getStatusRPC()

	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil

}

func (c *cosmosClient) GetTxsSentToAddressAfterHeight(address string, height int64) ([]*sdk.TxResponse, error) {
	_, err := util.AddressBytesFromBech32(c.Bech32Prefix, address)
	if err != nil {
		return nil, fmt.Errorf("invalid bech32 address: %s", err)
	}

	query := fmt.Sprintf("transfer.recipient='%s' AND tx.height>=%d", address, height)

	return c.getTxsByEvents(query)
}

func (c *cosmosClient) GetTxsSentFromAddressAfterHeight(address string, height int64) ([]*sdk.TxResponse, error) {
	_, err := util.AddressBytesFromBech32(c.Bech32Prefix, address)
	if err != nil {
		return nil, fmt.Errorf("invalid bech32 address: %s", err)
	}

	query := fmt.Sprintf("transfer.sender='%s' AND tx.height>=%d", address, height)

	return c.getTxsByEvents(query)
}

func (c *cosmosClient) getTxsByEventsPerPageGRPC(query string, page uint64) ([]*sdk.TxResponse, uint64, error) {
	if !c.GRPCEnabled {
		return nil, 0, fmt.Errorf("grpc disabled")
	}
	client := tx.NewServiceClient(c.grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req := &tx.GetTxsEventRequest{
		Query:   query,
		OrderBy: tx.OrderBy_ORDER_BY_ASC,
		Page:    page,
		Limit:   50,
	}

	resp, err := client.GetTxsEvent(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	return resp.TxResponses, resp.Total, nil
}

func (c *cosmosClient) getTxsByEventsPerPageRPC(query string, page uint64) ([]*sdk.TxResponse, uint64, error) {
	if c.GRPCEnabled {
		return nil, 0, fmt.Errorf("grpc enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	limit := 50
	pageint := int(page)

	resTxs, err := c.rpcClient.TxSearch(ctx, query, false, &pageint, &limit, "asc")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get txs: %s", err)
	}

	resBlocks, err := getBlocksForTxResults(c.rpcClient, resTxs.Txs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get blocks for txs: %s", err)
	}

	txs, err := formatTxResults(c.Bech32Prefix, resTxs.Txs, resBlocks)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to format tx results: %s", err)
	}

	return txs, uint64(resTxs.TotalCount), err
}

func (c *cosmosClient) getTxsByEvents(query string) ([]*sdk.TxResponse, error) {
	var page uint64 = 1
	var txs []*sdk.TxResponse
	for {

		var respTxs []*sdk.TxResponse
		var err error
		var total uint64

		if c.GRPCEnabled {
			respTxs, total, err = c.getTxsByEventsPerPageGRPC(query, page)
		} else {
			respTxs, total, err = c.getTxsByEventsPerPageRPC(query, page)
		}

		if err != nil {
			return nil, err
		}

		txs = append(txs, respTxs...)

		if len(respTxs) == 0 || len(txs) >= int(total) || page >= maxPageDepth {
			break
		}
		page++
	}

	return txs, nil
}

func (c *cosmosClient) getTxGRPC(hash string) (*sdk.TxResponse, error) {
	if !c.GRPCEnabled {
		return nil, fmt.Errorf("grpc disabled")
	}
	client := tx.NewServiceClient(c.grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req := &tx.GetTxRequest{
		Hash: hash,
	}

	resp, err := client.GetTx(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx: %s", err)
	}

	return resp.TxResponse, nil
}

func (c *cosmosClient) getTxRPC(hash string) (*sdk.TxResponse, error) {
	if c.GRPCEnabled {
		return nil, fmt.Errorf("grpc enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hash: %s", err)
	}

	resTx, err := c.rpcClient.Tx(ctx, hashBytes, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx: %s", err)
	}

	resBlocks, err := getBlocksForTxResults(c.rpcClient, []*rpctypes.ResultTx{resTx})
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks for tx: %s", err)
	}

	out, err := mkTxResult(c.Bech32Prefix, resTx, resBlocks[resTx.Height])
	if err != nil {
		return nil, fmt.Errorf("failed to format tx result: %s", err)
	}

	return out, nil
}

func (c *cosmosClient) GetTx(hash string) (*sdk.TxResponse, error) {
	if c.GRPCEnabled {
		return c.getTxGRPC(hash)
	}
	return c.getTxRPC(hash)
}

/*

	func (c *cosmosClient) SubmitRawTx(params rpc.SendRawTxParams) (*SubmitRawTxResponse, error) {
		j, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		res, err := queryRPC(sendRawTxPath, j)
		if err != nil {
			return nil, err
		}
		var obj SubmitRawTxResponse
		err = json.Unmarshal([]byte(res), &obj)
		return &obj, err
	}

*/

func (c *cosmosClient) GetChainID() (string, error) {
	var chainID string
	if c.GRPCEnabled {
		res, err := c.getLatestBlockGRPC()
		if err != nil {
			return "", fmt.Errorf("failed to get latest block: %s", err)
		}
		chainID = res.Header.ChainID
	} else {
		status, err := c.getStatusRPC()
		if err != nil {
			return "", fmt.Errorf("failed to get status: %s", err)
		}
		chainID = status.NodeInfo.Network
	}

	return chainID, nil
}

func (c *cosmosClient) ValidateNetwork() error {
	log.Debugf("[%s] Validating network", c.name)
	chainID, err := c.GetChainID()
	if err != nil {
		return fmt.Errorf("failed to validate network: %s", err)
	}
	if chainID != c.ChainID {
		return fmt.Errorf("failed to validate network: expected chain id %s, got %s", c.ChainID, chainID)
	}
	log.Debugf("[%s] Network validated", c.name)
	return nil
}

func NewClient(config models.CosmosNetworkConfig) (CosmosClient, error) {
	var connection *grpc.ClientConn
	var client *rpchttp.HTTP

	if config.GRPCEnabled {
		grpcURL := fmt.Sprintf("%s:%d", config.GRPCHost, config.GRPCPort)
		conn, err := grpc.Dial(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to grpc: %s", err)
		}
		connection = conn
		client = nil
	} else {
		c, err := rpchttp.New(config.RPCURL, "/websocket")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to rpc: %s", err)
		}
		client = c
		connection = nil
	}

	c := &cosmosClient{
		GRPCEnabled: config.GRPCEnabled,

		Timeout:        time.Duration(config.TimeoutMS) * time.Millisecond,
		ChainID:      config.ChainID,
		ChainName:    config.ChainName,
		Bech32Prefix: config.Bech32Prefix,
		CoinDenom:    config.CoinDenom,

		name:      strings.ToUpper(fmt.Sprintf("%s_CLIENT", config.ChainName)),
		grpcConn:  connection,
		rpcClient: client,
	}

	err := c.ValidateNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to validate network: %s", err)
	}

	return c, nil
}
