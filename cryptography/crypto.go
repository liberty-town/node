package cryptography

import (
	"crypto/rand"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"math/big"
)

// SignatureLength indicates the byte length required to carry a signature with recovery id.
const SignatureLength = 64 + 1 // 64 bytes ECDSA signature + 1 byte recovery id

// RecoveryIDOffset points to the byte offset within the signature that contains the recovery id.
const RecoveryIDOffset = 64

var (
	secp256k1N, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
)

func SHA3(b []byte) []byte {
	h := sha3.New256()
	h.Write(b)
	return h.Sum(nil)
}

func RIPEMD(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	return h.Sum(nil)
}

func RandomHash() (hash []byte) {
	a := make([]byte, 32)
	rand.Read(a)
	return a
}

func RandomBytes(length int) []byte {
	a := make([]byte, length)
	rand.Read(a)
	return a
}

func GetChecksum(b []byte) []byte {
	return RIPEMD(b)[:ChecksumSize]
}
