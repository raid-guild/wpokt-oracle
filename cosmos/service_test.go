package cosmos

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/db"
	dbMocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	ethMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"

	log "github.com/sirupsen/logrus"
)

func TestNewCosmosService(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
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

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
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
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	assert.Panics(t, func() {
		NewCosmosChainService(
			config,
			mintControllerMap,
			mnemonic,
			ethNetworks,
			nil,
			nil,
		)
	})

	wg := &sync.WaitGroup{}
	lastHealth := models.ChainServiceHealth{
		Chain: models.Chain{
			ChainID:   "poktroll",
			ChainName: "Poktroll",
			ChainType: models.ChainTypeCosmos,
		},
		MessageMonitor: &models.RunnerServiceStatus{BlockHeight: 100},
		MessageSigner:  &models.RunnerServiceStatus{BlockHeight: 100},
		MessageRelayer: &models.RunnerServiceStatus{BlockHeight: 100},
	}

	node := &models.Node{
		Health: []models.ChainServiceHealth{lastHealth},
	}

	service := NewCosmosChainService(
		config,
		mintControllerMap,
		mnemonic,
		ethNetworks,
		wg,
		node,
	)
	assert.NotNil(t, service)

}
