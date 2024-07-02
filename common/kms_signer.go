package common

import (
	"context"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	btcecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	dcrecSecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/api/option"
)

// Interface Definition
type KmsSigner interface {
	EthSignHash(hash common.Hash) ([]byte, error)
	CosmosSignHash(hash [32]byte) ([]byte, error)
	EthAddress() common.Address
	CosmosPublicKey() *secp256k1.PubKey
}

// Struct Definition
type GcpKmsSigner struct {
	client          *kms.KeyManagementClient
	keyName         string
	ethAddress      common.Address
	cosmosPubKey    *secp256k1.PubKey
	secp256k1PubKey *dcrecSecp256k1.PublicKey
}

// Constructor Function
func NewGcpKmsSigner(credsFilePath, keyName string) (*GcpKmsSigner, error) {
	client, err := kms.NewKeyManagementClient(context.Background(), option.WithCredentialsFile(credsFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %v", err)
	}

	// verify key algorithm
	keyVersionDetails, err := getKeyVersionDetails(client, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get key version details: %v", err)
	}

	if keyVersionDetails.Algorithm != kmspb.CryptoKeyVersion_EC_SIGN_SECP256K1_SHA256 {
		return nil, fmt.Errorf("key algorithm is not EC_SIGN_P256_SHA256")
	}

	ethAddress, err := resolveEthAddr(client, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Ethereum address: %v", err)
	}

	secp256k1PubKey, err := resolveSecp256k1PubKey(client, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Secp256k1 public key: %v", err)
	}

	cosmosPubKey := &secp256k1.PubKey{Key: secp256k1PubKey.SerializeCompressed()}

	return &GcpKmsSigner{
		client:          client,
		keyName:         keyName,
		ethAddress:      ethAddress,
		cosmosPubKey:    cosmosPubKey,
		secp256k1PubKey: secp256k1PubKey,
	}, nil
}

// Method Implementations
func (s *GcpKmsSigner) EthSignHash(hash common.Hash) ([]byte, error) {
	return ethSignHash(hash, s.client, s.keyName, s.ethAddress)
}

func (s *GcpKmsSigner) CosmosSignHash(hash common.Hash) ([]byte, error) {
	return cosmosSignHash(s.client, s.keyName, hash, s.secp256k1PubKey)
}

func (s *GcpKmsSigner) EthAddress() common.Address {
	return s.ethAddress
}

func (s *GcpKmsSigner) CosmosPublicKey() *secp256k1.PubKey {
	return s.cosmosPubKey
}

// Helper Functions (same as before)
func resolveEthAddr(client *kms.KeyManagementClient, keyName string) (common.Address, error) {
	resp, err := client.GetPublicKey(context.Background(), &kmspb.GetPublicKeyRequest{Name: keyName})
	if err != nil {
		return common.Address{}, fmt.Errorf("public key %q lookup: %w", keyName, err)
	}

	block, _ := pem.Decode([]byte(resp.Pem))
	if block == nil {
		return common.Address{}, fmt.Errorf("public key %q PEM empty: %.130q", keyName, resp.Pem)
	}

	var info struct {
		AlgID pkix.AlgorithmIdentifier
		Key   asn1.BitString
	}
	_, err = asn1.Unmarshal(block.Bytes, &info)
	if err != nil {
		return common.Address{}, fmt.Errorf("public key %q PEM block %q: %v", keyName, block.Type, err)
	}

	wantAlg := asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	if gotAlg := info.AlgID.Algorithm; !gotAlg.Equal(wantAlg) {
		return common.Address{}, fmt.Errorf("public key %q ASN.1 algorithm %s intead of %s", keyName, gotAlg, wantAlg)
	}

	return ethPubKeyAddr(info.Key.Bytes), nil
}

// PubKeyAddr returns the Ethereum address for (uncompressed-)key bytes.
func ethPubKeyAddr(bytes []byte) common.Address {
	digest := crypto.Keccak256(bytes[1:])
	var addr common.Address
	copy(addr[:], digest[12:])
	return addr
}

func ethSignHash(hash common.Hash, client *kms.KeyManagementClient, keyName string, ethAddress common.Address) ([]byte, error) {
	// Resolve a signature
	req := kmspb.AsymmetricSignRequest{
		Name: keyName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: hash[:],
			},
		},
	}
	resp, err := client.AsymmetricSign(context.Background(), &req)
	if err != nil {
		return nil, fmt.Errorf("asymmetric sign operation: %w", err)
	}

	// Parse signature
	var params struct{ R, S *big.Int }
	_, err = asn1.Unmarshal(resp.Signature, &params)
	if err != nil {
		return nil, fmt.Errorf("asymmetric signature encoding: %w", err)
	}
	var rLen, sLen int // byte size
	if params.R != nil {
		rLen = (params.R.BitLen() + 7) / 8
	}
	if params.S != nil {
		sLen = (params.S.BitLen() + 7) / 8
	}
	if rLen == 0 || rLen > 32 || sLen == 0 || sLen > 32 {
		return nil, fmt.Errorf("asymmetric signature with %d-byte r and %d-byte s denied on size", rLen, sLen)
	}

	// Need uncompressed signature with "recovery ID" at end:
	// https://bitcointalk.org/index.php?topic=5249677.0
	// https://ethereum.stackexchange.com/a/53182/39582
	var sig [66]byte // + 1-byte header + 1-byte tailer
	params.R.FillBytes(sig[33-rLen : 33])
	params.S.FillBytes(sig[65-sLen : 65])

	// Brute force try includes KMS verification
	var recoverErr error
	for recoveryID := byte(0); recoveryID < 2; recoveryID++ {
		sig[0] = recoveryID + 27 // BitCoin header
		btcsig := sig[:65]       // Exclude Ethereum 'v' parameter
		pubKey, _, err := btcecdsa.RecoverCompact(btcsig, hash[:])
		if err != nil {
			recoverErr = err
			continue
		}

		if ethPubKeyAddr(pubKey.SerializeUncompressed()) == ethAddress {
			// Sign the transaction
			sig[65] = recoveryID // Ethereum 'v' parameter
			return sig[1:], nil  // Exclude BitCoin header
		}
	}
	// RecoverErr can be nil, but that's OK
	return nil, fmt.Errorf("asymmetric signature address recovery mis: %w", recoverErr)
}

func cosmosSignHash(client *kms.KeyManagementClient, keyName string, hash [32]byte, pubKey *dcrecSecp256k1.PublicKey) ([]byte, error) {
	// Sign the hash using KMS
	req := &kmspb.AsymmetricSignRequest{
		Name: keyName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: hash[:],
			},
		},
	}

	resp, err := client.AsymmetricSign(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %v", err)
	}

	signature := resp.Signature

	// Extract r and s values from the signature
	var params struct{ R, S *big.Int }
	_, err = asn1.Unmarshal(signature, &params)
	if err != nil {
		return nil, fmt.Errorf("asymmetric signature encoding: %w", err)
	}

	rBytes := params.R.Bytes()
	sBytes := params.S.Bytes()

	// Ensure r and s are 32 bytes each
	rPadded := make([]byte, 32)
	sPadded := make([]byte, 32)
	copy(rPadded[32-len(rBytes):], rBytes)
	copy(sPadded[32-len(sBytes):], sBytes)

	finalSig := append(rPadded, sPadded...)

	sig, err := signatureFromBytes(finalSig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature: %v", err)
	}

	if !sig.Verify(hash[:], pubKey) {
		return nil, fmt.Errorf("signature verification failed")
	}

	return finalSig, nil
}

func signatureFromBytes(sigStr []byte) (*btcecdsa.Signature, error) {
	if len(sigStr) != 64 {
		return nil, fmt.Errorf("signature length is not 64 bytes")
	}

	var r dcrecSecp256k1.ModNScalar
	r.SetByteSlice(sigStr[:32])
	var s dcrecSecp256k1.ModNScalar
	s.SetByteSlice(sigStr[32:64])
	if s.IsOverHalfOrder() {
		return nil, fmt.Errorf("signature is not in lower-S form")
	}

	return btcecdsa.NewSignature(&r, &s), nil
}

// func resolveCosmosPubKey(client *kms.KeyManagementClient, keyName string) (*secp256k1.PubKey, error) {
// 	pubkeyObject, err := resolveSecp256k1PubKey(client, keyName)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to resolve public key: %v", err)
// 	}
//
// 	pk := pubkeyObject.SerializeCompressed()
//
// 	return &secp256k1.PubKey{Key: pk}, nil
// }

func resolveSecp256k1PubKey(client *kms.KeyManagementClient, keyName string) (*dcrecSecp256k1.PublicKey, error) {
	publicKeyResp, err := client.GetPublicKey(context.Background(), &kmspb.GetPublicKeyRequest{Name: keyName})
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %v", err)
	}

	publicKeyPem := publicKeyResp.Pem

	block, _ := pem.Decode([]byte(publicKeyPem))
	if block == nil {
		return nil, fmt.Errorf("public key %q PEM empty: %.130q", keyName, publicKeyPem)
	}

	var info struct {
		AlgID pkix.AlgorithmIdentifier
		Key   asn1.BitString
	}
	_, err = asn1.Unmarshal(block.Bytes, &info)
	if err != nil {
		return nil, fmt.Errorf("public key %q PEM block %q: %v", keyName, block.Type, err)
	}

	wantAlg := asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	if gotAlg := info.AlgID.Algorithm; !gotAlg.Equal(wantAlg) {
		return nil, fmt.Errorf("public key %q ASN.1 algorithm %s instead of %s", keyName, gotAlg, wantAlg)
	}

	pubkeyObject, err := dcrecSecp256k1.ParsePubKey(info.Key.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	return pubkeyObject, nil
}

func getKeyVersionDetails(client *kms.KeyManagementClient, keyName string) (*kmspb.CryptoKeyVersion, error) {
	// Request the key version details
	req := &kmspb.GetCryptoKeyVersionRequest{
		Name: keyName,
	}

	resp, err := client.GetCryptoKeyVersion(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to get key version details: %v", err)
	}

	return resp, nil
}
