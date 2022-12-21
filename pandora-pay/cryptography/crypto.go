package cryptography

import (
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
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

func GetChecksum(b []byte) []byte {
	return RIPEMD(b)[:ChecksumSize]
}
