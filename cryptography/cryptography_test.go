package cryptography

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateNewPrivateKey(t *testing.T) {

	privateKey, err := GenerateKey()
	assert.NoErrorf(t, err, "No error")
	assert.Equal(t, len(FromECDSA(privateKey)), 32, "Invalid private key length")

	publicKey := CompressPubkey(&privateKey.PublicKey)
	assert.Equal(t, len(publicKey), PublicKeySize)

	signature, err := Sign(SHA3([]byte{55, 22, 33, 11}), privateKey)
	assert.NoErrorf(t, err, "No error")

	assert.Equal(t, len(signature), SignatureSize)

}
