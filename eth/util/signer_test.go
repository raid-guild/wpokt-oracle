package util

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/dan13ram/wpokt-oracle/app"
	"github.com/dan13ram/wpokt-oracle/eth/autogen"
	eth "github.com/dan13ram/wpokt-oracle/eth/client"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestUpdateStatusAndConfirmationsForMint(t *testing.T) {

	testCases := []struct {
		name                  string
		initialMint           models.Mint
		poktHeight            int64
		requiredConfirmations int64
		expectedStatus        string
		expectedConfs         string
		expectedErr           bool
	}{
		{
			name: "Status is pending and confirmations are 0",
			initialMint: models.Mint{
				Status:        models.StatusPending,
				Confirmations: "0",
				Height:        "1000",
			},
			poktHeight:            1000,
			requiredConfirmations: 5,
			expectedStatus:        models.StatusPending,
			expectedConfs:         "0",
			expectedErr:           false,
		},
		{
			name: "Status is pending and confirmations are 0 but less than required confirmations",
			initialMint: models.Mint{
				Status:        models.StatusPending,
				Confirmations: "0",
				Height:        "1000",
			},
			poktHeight:            1005,
			requiredConfirmations: 10,
			expectedStatus:        models.StatusPending,
			expectedConfs:         "5",
			expectedErr:           false,
		},
		{
			name: "Status is pending and confirmations are greater than 0",
			initialMint: models.Mint{
				Status:        models.StatusPending,
				Confirmations: "10",
				Height:        "1000",
			},
			poktHeight:            1020,
			requiredConfirmations: 10,
			expectedStatus:        models.StatusConfirmed,
			expectedConfs:         "20",
			expectedErr:           false,
		},
		{
			name: "Status is confirmed",
			initialMint: models.Mint{
				Status:        models.StatusPending,
				Confirmations: "3",
				Height:        "1000",
			},
			poktHeight:            1024,
			requiredConfirmations: 2,
			expectedStatus:        models.StatusConfirmed,
			expectedConfs:         "24",
			expectedErr:           false,
		},
		{
			name: "Status is confirmed and required confirmations are 0",
			initialMint: models.Mint{
				Status:        models.StatusConfirmed,
				Confirmations: "3",
				Height:        "1000",
			},
			poktHeight:            1024,
			requiredConfirmations: 0,
			expectedStatus:        models.StatusConfirmed,
			expectedConfs:         "3",
			expectedErr:           false,
		},
		{
			name: "Status is pending and required confirmations are 0",
			initialMint: models.Mint{
				Status:        models.StatusPending,
				Confirmations: "-3",
				Height:        "1000",
			},
			poktHeight:            1024,
			requiredConfirmations: 0,
			expectedStatus:        models.StatusConfirmed,
			expectedConfs:         "0",
			expectedErr:           false,
		},
		{
			name: "Invalid Height",
			initialMint: models.Mint{
				Status:        models.StatusPending,
				Confirmations: "3",
				Height:        "random",
			},
			poktHeight:            1024,
			requiredConfirmations: 5,
			expectedStatus:        models.StatusConfirmed,
			expectedConfs:         "3",
			expectedErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app.Config.Pocket.Confirmations = tc.requiredConfirmations

			result, err := UpdateStatusAndConfirmationsForMint(&tc.initialMint, tc.poktHeight)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if result.Status != tc.expectedStatus {
					t.Errorf("Expected status %s, got %s", tc.expectedStatus, result.Status)
				}

				if result.Confirmations != tc.expectedConfs {
					t.Errorf("Expected confirmations %s, got %s", tc.expectedConfs, result.Confirmations)
				}
			}
		})
	}

}

func TestSignMint(t *testing.T) {
	ZERO_ADDRESS := "0x0000000000000000000000000000000000000000"

	testDomain := eth.DomainData{
		Name:              "Test",
		Version:           "1",
		ChainId:           big.NewInt(1),
		VerifyingContract: common.HexToAddress(ZERO_ADDRESS),
	}

	testPrivateKey, _ := crypto.HexToECDSA("1395eeb9c36ef43e9e05692c9ee34034c00a9bef301135a96d082b2a65fd1680")
	testSignature := "0x6b170e88743324cb571398f279d58a235e41d16efb7b4a90db7e86a6ddf5eb472d5b791c00aaa59c7755305cc3fb20c407a8d1fbbeacdd7032d509ce7c48cebd1b"
	testData := autogen.MintControllerMintData{
		Recipient: common.HexToAddress(ZERO_ADDRESS),
		Amount:    big.NewInt(100),
		Nonce:     big.NewInt(1),
	}

	testAddress := strings.ToLower(crypto.PubkeyToAddress(testPrivateKey.PublicKey).Hex())

	testCases := []struct {
		name         string
		initialMint  models.Mint
		numSigners   int
		expectedMint models.Mint
		expectedErr  bool
		data         autogen.MintControllerMintData
		domain       eth.DomainData
		privateKey   *ecdsa.PrivateKey
	}{
		{
			name: "Single signer, mint not signed",
			initialMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: nil,
				Signers:    nil,
			},
			numSigners: 1,
			expectedMint: models.Mint{
				Status:     models.StatusSigned,
				Signatures: []string{testSignature},
				Signers:    []string{testAddress},
			},
			expectedErr: false,
			data:        testData,
			domain:      testDomain,
			privateKey:  testPrivateKey,
		},
		{
			name: "Multiple signers, mint not signed",
			initialMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: nil,
				Signers:    nil,
			},
			numSigners: 2,
			expectedMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{testSignature},
				Signers:    []string{testAddress},
			},
			expectedErr: false,
			data:        testData,
			domain:      testDomain,
			privateKey:  testPrivateKey,
		},
		{
			name: "Multiple signers, mint signed",
			initialMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{"0x..."},
				Signers:    []string{ZERO_ADDRESS},
			},
			numSigners: 2,
			expectedMint: models.Mint{
				Status:     models.StatusSigned,
				Signatures: []string{"0x...", testSignature},
				Signers:    []string{ZERO_ADDRESS, testAddress},
			},
			expectedErr: false,
			data:        testData,
			domain:      testDomain,
			privateKey:  testPrivateKey,
		},
		{
			name: "Multiple signers, mint signed",
			initialMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{"0x..."},
				Signers:    []string{ZERO_ADDRESS},
			},
			numSigners: 3,
			expectedMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{"0x...", testSignature},
				Signers:    []string{ZERO_ADDRESS, testAddress},
			},
			expectedErr: false,
			data:        testData,
			domain:      testDomain,
			privateKey:  testPrivateKey,
		},
		{
			name: "Invalid domain",
			initialMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{"0x..."},
				Signers:    []string{ZERO_ADDRESS},
			},
			numSigners: 3,
			expectedMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{"0x...", testSignature},
				Signers:    []string{ZERO_ADDRESS, testAddress},
			},
			expectedErr: true,
			data:        testData,
			domain: eth.DomainData{
				ChainId: big.NewInt(1),
			},
			privateKey: testPrivateKey,
		},
		{
			name: "Invalid privateKey",
			initialMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{"0x..."},
				Signers:    []string{ZERO_ADDRESS},
			},
			numSigners: 3,
			expectedMint: models.Mint{
				Status:     models.StatusConfirmed,
				Signatures: []string{"0x...", testSignature},
				Signers:    []string{ZERO_ADDRESS, testAddress},
			},
			expectedErr: true,
			data: autogen.MintControllerMintData{
				Recipient: common.HexToAddress(ZERO_ADDRESS),
				Amount:    big.NewInt(100),
			},
			domain:     testDomain,
			privateKey: testPrivateKey,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SignMint(&tc.initialMint, &tc.data, tc.domain, tc.privateKey, tc.numSigners)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMint, *result)
			}

		})
	}

}
