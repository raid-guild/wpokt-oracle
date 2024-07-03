package common

import (
	"context"
	"encoding/asn1"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	dcrecSecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/common"
	gax "github.com/googleapis/gax-go/v2"
)

// MockGCPKeyManagementClient is a mock implementation of GCPKeyManagementClient
type MockGCPKeyManagementClient struct {
	mock.Mock
}

func (m *MockGCPKeyManagementClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGCPKeyManagementClient) GetPublicKey(ctx context.Context, req *kmspb.GetPublicKeyRequest, opts ...gax.CallOption) (*kmspb.PublicKey, error) {
	args := m.Called(ctx, req, opts)
	return args.Get(0).(*kmspb.PublicKey), args.Error(1)
}

func (m *MockGCPKeyManagementClient) AsymmetricSign(ctx context.Context, req *kmspb.AsymmetricSignRequest, opts ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error) {
	args := m.Called(ctx, req, opts)
	return args.Get(0).(*kmspb.AsymmetricSignResponse), args.Error(1)
}

func (m *MockGCPKeyManagementClient) GetCryptoKeyVersion(ctx context.Context, req *kmspb.GetCryptoKeyVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKeyVersion, error) {
	args := m.Called(ctx, req, opts)
	return args.Get(0).(*kmspb.CryptoKeyVersion), args.Error(1)
}

func mockASN1Signature() []byte {
	r, _ := new(big.Int).SetString("0adb0cad4842abd2003c8dcf4f2663b1ba9106b208a8b9f62aaa628121d7e818", 16)
	s, _ := new(big.Int).SetString("05a2e4ace4f4dae5437804d8ff3a67151e96d0efad0427cb49844ea80c3a03a5", 16)
	return asn1Bytes(r, s)
}

func mockPublicKey() []byte {
	keyHex := "0459fe4b6a8682418cb86df571c0a36b06f1b37e6760985009599c31d783c6ccac6cb07ba459e7d4949d12b178f8468802ebf7d932007e46b02fc7a2bf90761fd6"
	return common.Hex2Bytes(keyHex)
}

func asn1Bytes(r, s *big.Int) []byte {
	signature, _ := asn1.Marshal(struct {
		R, S *big.Int
	}{R: r, S: s})
	return signature
}

// Unit tests for GcpKmsSigner
func TestNewGcpKmsSigner(t *testing.T) {
	mockClient := new(MockGCPKeyManagementClient)
	NewGCPKeyManagementClient = func(ctx context.Context) (GCPKeyManagementClient, error) {
		return mockClient, nil
	}

	keyName := "test-key"
	expectedKeyVersion := &kmspb.CryptoKeyVersion{
		Algorithm: kmspb.CryptoKeyVersion_EC_SIGN_SECP256K1_SHA256,
	}
	mockClient.On("GetCryptoKeyVersion", mock.Anything, &kmspb.GetCryptoKeyVersionRequest{Name: keyName}, mock.Anything).Return(expectedKeyVersion, nil)

	expectedPublicKey := &kmspb.PublicKey{
		Pem: "-----BEGIN PUBLIC KEY-----\nMFYwEAYHKoZIzj0CAQYFK4EEAAoDQgAEWf5LaoaCQYy4bfVxwKNrBvGzfmdgmFAJ\nWZwx14PGzKxssHukWefUlJ0SsXj4RogC6/fZMgB+RrAvx6K/kHYf1g==\n-----END PUBLIC KEY-----",
	}
	mockClient.On("GetPublicKey", mock.Anything, &kmspb.GetPublicKeyRequest{Name: keyName}, mock.Anything).Return(expectedPublicKey, nil)

	signer, err := NewGcpKmsSigner(keyName)
	assert.NoError(t, err)
	assert.NotNil(t, signer)

	mockClient.AssertExpectations(t)
}

func TestGcpKmsSigner_EthSign(t *testing.T) {
	mockClient := new(MockGCPKeyManagementClient)
	keyName := "test-key"
	ethAddress := common.HexToAddress("0xE14FE42D9Cf3C39c27E1c1Ca3dAa033C43eD4d88")

	signer := &GcpKmsSigner{
		client:     mockClient,
		keyName:    keyName,
		ethAddress: ethAddress,
	}

	data := []byte("example transaction data")
	// expectedHash := crypto.Keccak256(data)
	expectedSignature := &kmspb.AsymmetricSignResponse{
		Signature: mockASN1Signature(),
	}
	mockClient.On("AsymmetricSign", mock.Anything, mock.Anything, mock.Anything).Return(expectedSignature, nil)

	sig, err := signer.EthSign(data)
	assert.NoError(t, err)
	assert.NotNil(t, sig)

	mockClient.AssertExpectations(t)
}

func TestGcpKmsSigner_CosmosSign(t *testing.T) {
	mockClient := new(MockGCPKeyManagementClient)
	keyName := "test-key"

	pubKeyBytes := mockPublicKey()
	secp256k1PubKey, _ := dcrecSecp256k1.ParsePubKey(pubKeyBytes)
	cosmosPubKey := &secp256k1.PubKey{Key: secp256k1PubKey.SerializeCompressed()}

	signer := &GcpKmsSigner{
		client:          mockClient,
		keyName:         keyName,
		secp256k1PubKey: secp256k1PubKey,
		cosmosPubKey:    cosmosPubKey,
	}

	data := []byte("example transaction data")
	// expectedHash := crypto.Keccak256(data)
	expectedSignature := &kmspb.AsymmetricSignResponse{
		Signature: mockASN1Signature(),
	}
	mockClient.On("AsymmetricSign", mock.Anything, mock.Anything, mock.Anything).Return(expectedSignature, nil)

	sig, err := signer.CosmosSign(data)
	assert.NoError(t, err)
	assert.NotNil(t, sig)

	mockClient.AssertExpectations(t)
}

func TestGcpKmsSigner_Destroy(t *testing.T) {
	mockClient := new(MockGCPKeyManagementClient)
	keyName := "test-key"

	signer := &GcpKmsSigner{
		client:  mockClient,
		keyName: keyName,
	}

	mockClient.On("Close").Return(nil)

	signer.Destroy()
	mockClient.AssertExpectations(t)
}

func TestGcpKmsSigner_WithGCPKMS(t *testing.T) {

	keyName := os.Getenv("GCP_KMS_KEY_NAME")
	if keyName == "" {
		t.Skip("GCP KMS key name not set")
	}
	credentails := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentails == "" {
		t.Skip("GCP credentials not set")
	}

	signer, err := NewGcpKmsSigner(keyName)
	assert.NoError(t, err)

	data := []byte("example transaction data")

	// Test EthSign
	sig, err := signer.EthSign(data)
	assert.NoError(t, err)
	assert.NotNil(t, sig)

	// Test CosmosSign
	sig, err = signer.CosmosSign(data)
	assert.NoError(t, err)
	assert.NotNil(t, sig)

}
