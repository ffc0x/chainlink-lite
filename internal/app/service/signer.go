package service

import (
	"encoding/hex"

	"github.com/libp2p/go-libp2p/core/crypto"
)

type SignerService struct {
	key crypto.PrivKey
}

func NewSignerService() (*SignerService, error) {
	key, _, err := crypto.GenerateKeyPair(crypto.ECDSA, -1)
	if err != nil {
		return nil, err
	}
	return &SignerService{key: key}, nil
}

// SignMessage signs a message with the private key
func (s *SignerService) SignMessage(message string) (string, error) {
	signature, err := s.key.Sign([]byte(message))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil
}

func (s *SignerService) GetPublicKey() crypto.PubKey {
	return s.key.GetPublic()
}

func (s *SignerService) GetPrivateKey() crypto.PrivKey {
	return s.key
}

// VerifySignature verifies the signature of a message
func VerifySignature(message string, hexSignature string, pub crypto.PubKey) (bool, error) {
	sigBytes, err := hex.DecodeString(hexSignature)
	if err != nil {
		return false, err
	}
	return pub.Verify([]byte(message), sigBytes)
}
