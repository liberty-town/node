//go:build !wasm
// +build !wasm

package addresses

import (
	"github.com/stretchr/testify/assert"
	"liberty-town/node/pandora-pay/cryptography"
	"liberty-town/node/pandora-pay/helpers"
	"math/rand"
	"testing"
)

func Test_VerifySignedMessage(t *testing.T) {

	for i := 0; i < 100; i++ {

		privateKey := GenerateNewPrivateKey()
		address, err := privateKey.GenerateAddress(false, nil, false, nil, 0, nil)
		assert.Nil(t, err, "Error generating key")

		message := helpers.RandomBytes(cryptography.HashSize)
		signature, err := privateKey.Sign(message)
		assert.Nil(t, err, "Error signing")

		assert.Equal(t, len(signature), cryptography.SignatureSize, "signature length is invalid")

		emptySignature := helpers.EmptyBytes(cryptography.SignatureSize)
		assert.NotEqual(t, signature, emptySignature, "Signing is empty...")

		assert.Equal(t, address.VerifySignedMessage(message, signature), true, "verification failed")

		var signature2 = helpers.CloneBytes(signature)
		copy(signature2, signature)

		value := byte(rand.Uint64() % 256)
		if signature2[2] == value {
			signature2[2] = value + byte(rand.Uint64()%255)
		} else {
			signature2[2] = value
		}

		assert.Equal(t, address.VerifySignedMessage(message, signature2), false, "Changed Signature was validated")

	}

}
