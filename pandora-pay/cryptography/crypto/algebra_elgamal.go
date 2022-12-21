package crypto

import (
	"liberty-town/node/pandora-pay/cryptography/bn256"
	"math/big"
)

// a ZERO
var ElGamal_ZERO *bn256.G1
var ElGamal_ZERO_string string
var ElGamal_BASE_G *bn256.G1

type ElGamal struct {
	G          *bn256.G1
	Randomness *big.Int
	Left       *bn256.G1
	Right      *bn256.G1
}
