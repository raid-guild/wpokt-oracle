package ethereum

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	cosmosMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/db/mocks"
	clientMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
)

func TestNewEthereumService(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	mockMailbox := clientMocks.NewMockMailboxContract(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	mockWarpISM := clientMocks.NewMockWarpISMContract(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.EthereumNetworkConfig{
		StartBlockHeight:      1,
		Confirmations:         1,
		RPCURL:                "http://localhost:8545",
		TimeoutMS:             1000,
		ChainID:               1,
		ChainName:             "Ethereum",
		MailboxAddress:        "0x0000000000000000000000000000000000000000",
		MintControllerAddress: "0x0000000000000000000000000000000000000000",
		OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
		WarpISMAddress:        "0x0000000000000000000000000000000000000000",
		OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	cosmosNetwork := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	ethNetworks := []models.EthereumNetworkConfig{
		config,
	}

	ethNewClient = func(models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		return mockClient, nil
	}

	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return mockMailbox, nil
	}

	ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
		return mockMintController, nil
	}

	ethNewWarpISMContract = func(ethcommon.Address, bind.ContractBackend) (eth.WarpISMContract, error) {
		return mockWarpISM, nil
	}

	cosmosNewClient = func(models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockCosmosClient, nil
	}

	dbNewDB = func() db.DB {
		return mockDB
	}

	defer func() {
		ethNewClient = eth.NewClient
		ethNewMailboxContract = eth.NewMailboxContract
		ethNewMintControllerContract = eth.NewMintControllerContract
		ethNewWarpISMContract = eth.NewWarpISMContract
		cosmosNewClient = cosmos.NewClient
		dbNewDB = db.NewDB
	}()

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mockCosmosClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	mockClient.EXPECT().GetClient().Return(nil)

	mockClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(1)})

	mockWarpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(3), nil)
	mockWarpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(2), nil)
	mockWarpISM.EXPECT().Eip712Domain(mock.Anything).Return(util.DomainData{ChainId: big.NewInt(1), VerifyingContract: ethcommon.HexToAddress(config.WarpISMAddress)}, nil)
	mockMintController.EXPECT().MaxMintLimit(mock.Anything).Return(big.NewInt(100), nil)

	mintControllerMap := map[uint32][]byte{
		1: ethcommon.FromHex("0x01"),
		2: ethcommon.FromHex("0x02"),
	}

	assert.Panics(t, func() {
		NewEthereumChainService(
			config,
			cosmosNetwork,
			mintControllerMap,
			ethNetworks,
			mnemonic,
			nil,
			nil,
		)
	})

	wg := &sync.WaitGroup{}
	lastHealth := models.ChainServiceHealth{
		Chain: models.Chain{
			ChainID:     "1",
			ChainDomain: 1,
			ChainName:   "Ethereum",
			ChainType:   models.ChainTypeEthereum,
		},
		MessageMonitor: &models.RunnerServiceStatus{BlockHeight: 100},
		MessageSigner:  &models.RunnerServiceStatus{BlockHeight: 100},
		MessageRelayer: &models.RunnerServiceStatus{BlockHeight: 100},
	}

	node := &models.Node{
		Health: []models.ChainServiceHealth{lastHealth},
	}

	service := NewEthereumChainService(
		config,
		cosmosNetwork,
		mintControllerMap,
		ethNetworks,
		mnemonic,
		wg,
		node,
	)
	assert.NotNil(t, service)

}
