package jwe

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/square/go-jose"
)

// Manager provides encrypt & decrypt interface
type Manager interface {
	GenerateToken(marshaled []byte) (string, error)
	GetDataByToken(token string) ([]byte, error)
}

type manager struct {
	secretKey  []byte
	expiration time.Duration
	crypter    jose.Encrypter
}

type claim struct {
	Data      []byte `json:"data"`
	ExpiresAt int64  `json:"exp"`
}

// NewManager returns JWE manager
// it uses A128GCMKW as a key encryption algoritm and A128CBC-HS256 as a content encryption algorithm
func NewManager(secretKey string, expiration time.Duration) (Manager, error) {
	if len(secretKey) != 32 {
		return nil, errors.New("Length of secretKey is not 32")
	}

	alg := jose.KeyAlgorithm(jose.A128GCMKW)
	enc := jose.ContentEncryption(jose.A128CBC_HS256)

	crypter, err := jose.NewEncrypter(enc, jose.Recipient{Algorithm: alg, Key: secretKey}, nil)
	if err != nil {
		return nil, err
	}

	return &manager{
		secretKey:  []byte(secretKey),
		expiration: expiration,
		crypter:    crypter,
	}, nil
}

func (m *manager) GenerateToken(marshaled []byte) (string, error) {
	c := claim{
		Data:      marshaled,
		ExpiresAt: time.Now().Add(m.expiration).Unix(),
	}
	serialized, err := json.Marshal(&c)
	if err != nil {
		return "", err
	}

	cipher, err := m.crypter.Encrypt(serialized)
	if err != nil {
		return "", err
	}

	return cipher.FullSerialize(), nil
}

func (m *manager) GetDataByToken(tokenString string) ([]byte, error) {
	cipher, err := jose.ParseEncrypted(tokenString)
	if err != nil {
		return nil, err
	}

	serialized, err := cipher.Decrypt(m.secretKey)
	if err != nil {
		return nil, err
	}

	c := &claim{}
	err = json.Unmarshal(serialized, &c)
	if err != nil {
		return nil, err
	}

	// Ignore expired token temporarily REWARD-261
	// if c.ExpiresAt < time.Now().Unix() {
	// 	return nil, errors.New("expired token")
	// }

	return c.Data, nil
}
