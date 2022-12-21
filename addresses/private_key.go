package addresses

import (
	"crypto/ecdsa"
	"errors"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/cryptography/ecies"
)

type PrivateKey struct {
	KeyWIF
	ecdsaKey *ecdsa.PrivateKey
}

func (pk *PrivateKey) GeneratePublicKey() []byte {
	return cryptography.CompressPubkey(&pk.ecdsaKey.PublicKey)
}

func (pk *PrivateKey) GenerateAddress() (*Address, error) {
	return CreateAddr(pk.GeneratePublicKey())
}

//need hash
func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	return cryptography.Sign(cryptography.SHA3(message), pk.ecdsaKey)
}

func (pk *PrivateKey) Decrypt(encrypted []byte) ([]byte, error) {
	return ecies.ImportECDSA(pk.ecdsaKey).Decrypt(encrypted, nil, nil)
}

func (pk *PrivateKey) Deserialize(buffer []byte) (err error) {
	if err = pk.deserialize(buffer, cryptography.PrivateKeySize); err != nil {
		return
	}
	if pk.ecdsaKey, err = cryptography.ToECDSA(pk.Key); err != nil {
		return
	}
	return
}

func GenerateNewPrivateKey() *PrivateKey {
	for {
		key, err := cryptography.GenerateKey()
		if err != nil {
			continue
		}

		privateKey, err := NewPrivateKey(cryptography.FromECDSA(key))
		if err != nil {
			continue
		}
		return privateKey
	}
}

func NewPrivateKey(key []byte) (*PrivateKey, error) {

	if len(key) != cryptography.PrivateKeySize {
		return nil, errors.New("Private Key length is invalid")
	}

	ecdsaKey, err := cryptography.ToECDSA(key)
	if err != nil {
		return nil, err
	}

	privateKey := &PrivateKey{
		KeyWIF{
			SIMPLE_PRIVATE_KEY_WIF,
			config.NETWORK_SELECTED,
			key,
			nil,
		},
		ecdsaKey,
	}

	privateKey.Checksum = privateKey.computeCheckSum()

	return privateKey, nil
}
