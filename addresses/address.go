package addresses

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	base58 "github.com/mr-tron/base58"
	msgpack "github.com/vmihailenco/msgpack/v5"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/cryptography/ecies"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type Address struct {
	Network        uint64         `json:"network" msgpack:"network"`
	Version        AddressVersion `json:"version" msgpack:"version"`
	PublicKey      []byte         `json:"publicKey" msgpack:"publicKey"`
	Encoded        string         `json:"encoded" msgpack:"encoded"`
	ecdsaPublicKey *ecdsa.PublicKey
}

func newAddr(network uint64, version AddressVersion, publicKey []byte) (*Address, error) {
	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Invalid PublicKey size")
	}

	ecdsaPublicKey, err := cryptography.DecompressPubkey(publicKey)
	if err != nil {
		return nil, err
	}

	a := &Address{network, version, publicKey, "", ecdsaPublicKey}
	a.Encoded = a.EncodeAddr()
	return a, nil
}

func CreateAddr(publicKey []byte) (*Address, error) {
	return newAddr(config.NETWORK_SELECTED, SIMPLE_PUBLIC_KEY, publicKey)
}

func CreateAddrFromSignature(message, signature []byte) (*Address, error) {
	pub, err := cryptography.SigToPub(cryptography.SHA3(message), signature)
	if err != nil {
		return nil, err
	}

	publicKey := cryptography.CompressPubkey(pub)
	a := &Address{config.NETWORK_SELECTED, SIMPLE_PUBLIC_KEY, publicKey, "", pub}
	a.Encoded = a.EncodeAddr()
	return a, nil
}

func (a *Address) EncodeAddr() string {
	if a == nil {
		return ""
	}

	writer := advanced_buffers.NewBufferWriter()

	var prefix string
	switch a.Network {
	case config.MAIN_NET_NETWORK_BYTE:
		prefix = config.MAIN_NET_NETWORK_BYTE_PREFIX
	case config.TEST_NET_NETWORK_BYTE:
		prefix = config.TEST_NET_NETWORK_BYTE_PREFIX
	case config.DEV_NET_NETWORK_BYTE:
		prefix = config.DEV_NET_NETWORK_BYTE_PREFIX
	default:
		panic("Invalid network")
	}

	writer.WriteUvarint(uint64(a.Version))

	writer.Write(a.PublicKey)

	buffer := writer.Bytes()

	checksum := cryptography.GetChecksum(buffer)
	buffer = append(buffer, checksum...)
	ret := base58.Encode(buffer)

	return prefix + ret
}

func DecodeAddr(input string) (*Address, error) {

	addr := &Address{}

	if len(input) < config.NETWORK_BYTE_PREFIX_LENGTH {
		return nil, errors.New("Invalid Address length")
	}

	prefix := input[0:config.NETWORK_BYTE_PREFIX_LENGTH]

	switch prefix {
	case config.MAIN_NET_NETWORK_BYTE_PREFIX:
		addr.Network = config.MAIN_NET_NETWORK_BYTE
	case config.TEST_NET_NETWORK_BYTE_PREFIX:
		addr.Network = config.TEST_NET_NETWORK_BYTE
	case config.DEV_NET_NETWORK_BYTE_PREFIX:
		addr.Network = config.DEV_NET_NETWORK_BYTE
	default:
		return nil, errors.New("Invalid Address Network PREFIX!")
	}

	if addr.Network != config.NETWORK_SELECTED {
		return nil, errors.New("Address network is invalid")
	}

	buf, err := base58.Decode(input[config.NETWORK_BYTE_PREFIX_LENGTH:])
	if err != nil {
		return nil, err
	}

	checksum := cryptography.GetChecksum(buf[:len(buf)-cryptography.ChecksumSize])

	if !bytes.Equal(checksum, buf[len(buf)-cryptography.ChecksumSize:]) {
		return nil, errors.New("Invalid Checksum")
	}

	buf = buf[0 : len(buf)-cryptography.ChecksumSize] // remove the checksum

	reader := advanced_buffers.NewBufferReader(buf)

	version, err := reader.ReadUvarint()
	if err != nil {
		return nil, err
	}
	addr.Version = AddressVersion(version)

	switch addr.Version {
	case SIMPLE_PUBLIC_KEY:
		if addr.PublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
			return nil, err
		}
		if addr.ecdsaPublicKey, err = cryptography.DecompressPubkey(addr.PublicKey); err != nil {
			return nil, err
		}

		addr.Encoded = input
	default:
		return nil, errors.New("Invalid Address Version")
	}

	return addr, nil
}

func (a *Address) EncryptMessage(message []byte) ([]byte, error) {
	return ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(a.ecdsaPublicKey), message, nil, nil)
}

//need hash
func (a *Address) VerifySignedMessage(message, signature []byte) bool {
	return cryptography.VerifySignature(a.PublicKey, cryptography.SHA3(message), signature)
}

func (a *Address) Equals(a2 *Address) bool {
	return a.Version == a2.Version && a.Network == a2.Network && bytes.Equal(a.PublicKey, a2.PublicKey)
}

func (a *Address) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteUvarint(a.Network)
	w.WriteUvarint(uint64(a.Version))
	w.Write(a.PublicKey)
}

func (a *Address) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if a.Network, err = r.ReadUvarint(); err != nil {
		return
	}
	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	a.Version = AddressVersion(n)
	switch a.Version {
	case SIMPLE_PUBLIC_KEY:
		if a.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if a.ecdsaPublicKey, err = cryptography.DecompressPubkey(a.PublicKey); err != nil {
			return
		}
	default:
		return errors.New("address invalid version")
	}

	a.Encoded = a.EncodeAddr()

	return
}

func (a *Address) MarshalJSON() ([]byte, error) {
	x := "\"" + a.Encoded + "\""
	return []byte(x), nil
}

func (a *Address) UnmarshalJSON(data []byte) error {
	b, err := DecodeAddr(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	a.Network = b.Network
	a.Version = b.Version
	a.PublicKey = b.PublicKey
	a.Encoded = b.Encoded
	a.ecdsaPublicKey = b.ecdsaPublicKey
	return nil
}

func (a *Address) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(helpers.SerializeToBytes(a))
}

func (a *Address) DecodeMsgpack(dec *msgpack.Decoder) error {
	var b []byte
	if err := dec.Decode(&b); err != nil {
		return err
	}
	return a.Deserialize(advanced_buffers.NewBufferReader(b))
}
