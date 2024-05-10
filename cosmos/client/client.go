package client

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/types"

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
	GetLatestBlock() (*cmtservice.Block, error)
	GetTxsSentFromAddressAfterHeight(address string, height int64) ([]*types.TxResponse, error)
	GetTxsSentToAddressAfterHeight(address string, height int64) ([]*types.TxResponse, error)
	// SubmitRawTx(params rpc.SendRawTxParams) (*SubmitRawTxResponse, error)
	GetTx(hash string) (*types.TxResponse, error)
	ValidateNetwork() error
}

type pocketClient struct {
	GRPCHost      string
	GRPCPort      int64
	GRPCTimeoutMS int64
	ChainId       string
	ChainName     string
	Bech32Prefix  string

	name       string
	connection *grpc.ClientConn
}

func (c *pocketClient) GetLatestBlock() (*cmtservice.Block, error) {
	// Create a new service client
	client := cmtservice.NewServiceClient(c.connection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.GRPCTimeoutMS)*time.Millisecond)
	defer cancel()

	req := &cmtservice.GetLatestBlockRequest{}

	// Query the latest block
	resp, err := client.GetLatestBlock(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.SdkBlock, nil
}

// get transactions sent from address
func (c *pocketClient) GetTxsSentToAddressAfterHeight(address string, height int64) ([]*types.TxResponse, error) {
	_, err := util.AddressBytesFromBech32(c.Bech32Prefix, address)
	if err != nil {
		return nil, fmt.Errorf("invalid bech32 address: %s", err)
	}

	query := fmt.Sprintf("transfer.recipient='%s' AND tx.height>=%d", address, height)

	return c.getTxsByEvents(query)
}

// get transactions sent to address
func (c *pocketClient) GetTxsSentFromAddressAfterHeight(address string, height int64) ([]*types.TxResponse, error) {
	_, err := util.AddressBytesFromBech32(c.Bech32Prefix, address)
	if err != nil {
		return nil, fmt.Errorf("invalid bech32 address: %s", err)
	}

	query := fmt.Sprintf("transfer.sender='%s' AND tx.height>=%d", address, height)

	return c.getTxsByEvents(query)
}

func (c *pocketClient) getTxsByEvents(query string) ([]*types.TxResponse, error) {
	var page uint64 = 1
	var txs []*types.TxResponse
	for {
		client := tx.NewServiceClient(c.connection)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.GRPCTimeoutMS)*time.Millisecond)
		defer cancel()

		req := &tx.GetTxsEventRequest{
			Query:   query,
			OrderBy: tx.OrderBy_ORDER_BY_ASC,
			Page:    page,
			Limit:   50,
		}

		resp, err := client.GetTxsEvent(ctx, req)
		if err != nil {
			return nil, err
		}

		txs = append(txs, resp.TxResponses...)

		if len(resp.Txs) == 0 || len(txs) >= int(resp.Total) || page >= maxPageDepth {
			break
		}
		page++
	}

	return txs, nil
}

func (c *pocketClient) GetTx(hash string) (*types.TxResponse, error) {
	client := tx.NewServiceClient(c.connection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.GRPCTimeoutMS)*time.Millisecond)
	defer cancel()

	req := &tx.GetTxRequest{
		Hash: hash,
	}

	resp, err := client.GetTx(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.TxResponse, nil
}

/*
	func (c *pocketClient) GetTx(hash string) (*TxResponse, error) {
		params := rpc.HashAndProveParams{Hash: hash, Prove: false}
		j, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		res, err := queryRPC(getTxPath, j)
		if err != nil {
			return nil, err
		}
		var obj TxResponse
		err = json.Unmarshal([]byte(res), &obj)
		return &obj, err
	}

	func (c *pocketClient) SubmitRawTx(params rpc.SendRawTxParams) (*SubmitRawTxResponse, error) {
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

	func (c *pocketClient) getAccountTxsPerPage(address string, page uint32) (*AccountTxsResponse, error) {
		// filter by received transactions
		params := rpc.PaginateAddrParams{
			Address:  address,
			Page:     int(page),
			PerPage:  50,
			Received: true,
			Prove:    false,
			Sort:     "asc",
		}
		j, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		res, err := queryRPC(getAccountTxsPath, j)
		if err != nil {

			return nil, err
		}
		var obj AccountTxsResponse
		err = json.Unmarshal([]byte(res), &obj)
		return &obj, err
	}

	func (c *pocketClient) GetAccountTxsByHeight(address string, height int64) ([]*TxResponse, error) {
		var txs []*TxResponse
		var page uint32 = 1
		for {
			res, err := c.getAccountTxsPerPage(address, page)
			if err != nil {
				return nil, err
			}
			// filter only type pos/Send
			for _, tx := range res.Txs {
				if tx.StdTx.Msg.Type == "pos/Send" && tx.Height >= height {
					txs = append(txs, tx)
				}
			}
			if len(res.Txs) == 0 || len(txs) >= int(res.TotalTxs) || res.Txs[len(res.Txs)-1].Height < height {
				break
			}
			page++
		}

		return txs, nil
	}
*/
func (c *pocketClient) ValidateNetwork() error {
	log.Debugf("[%s] Validating network", c.name)
	res, err := c.GetLatestBlock()
	if err != nil {
		return fmt.Errorf("failed to get latest block: %s", err)
	}
	if res.Header.ChainID != c.ChainId {
		return fmt.Errorf("failed to validate network: expected chain id %s, got %s", c.ChainId, res.Header.ChainID)
	}
	log.Debugf("[%s] Network validated", c.name)
	return nil
}

func NewClient(config models.CosmosNetworkConfig) (CosmosClient, error) {
	grpcUrl := fmt.Sprintf("%s:%d", config.GRPCHost, config.GRPCPort)
	conn, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %s", grpcUrl, err)
	}

	client := &pocketClient{
		GRPCHost:      config.GRPCHost,
		GRPCPort:      config.GRPCPort,
		GRPCTimeoutMS: config.GRPCTimeoutMS,
		ChainId:       config.ChainID,
		ChainName:     config.ChainName,
		Bech32Prefix:  config.Bech32Prefix,

		name:       strings.ToUpper(fmt.Sprintf("%s_CLIENT", config.ChainName)),
		connection: conn,
	}

	err = client.ValidateNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to validate network: %s", err)
	}

	return client, nil
}
