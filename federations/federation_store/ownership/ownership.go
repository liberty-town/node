package ownership

import (
	"liberty-town/node/addresses"
	"liberty-town/node/cryptography"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"time"
)

type Ownership struct {
	Timestamp uint64             `json:"timestamp" msgpack:"timestamp"`
	Signature []byte             `json:"signature" msgpack:"signature"`
	Address   *addresses.Address `json:"address" msgpack:"address"`
}

func (this *Ownership) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeSignature bool) {
	w.WriteUvarint(this.Timestamp)
	if includeSignature {
		w.Write(this.Signature)
	}
}

func (this *Ownership) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true)
}

func (this *Ownership) SerializeToBytes() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.Serialize(w)
	return w.Bytes()
}

func (this *Ownership) Deserialize(r *advanced_buffers.BufferReader, getMessage func() []byte) (err error) {
	if this.Timestamp, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	if this.Address, err = addresses.CreateAddrFromSignature(getMessage(), this.Signature); err != nil {
		return
	}
	return
}

func (this *Ownership) Sign(privKey *addresses.PrivateKey, getMessage func() []byte) (err error) {
	this.Timestamp = uint64(time.Now().Unix())
	if this.Address, err = privKey.GenerateAddress(); err != nil {
		return
	}
	if this.Signature, err = privKey.Sign(getMessage()); err != nil {
		return
	}
	return
}

func (this *Ownership) Verify(getMessage func() []byte) bool {
	return this.Address.VerifySignedMessage(getMessage(), this.Signature[0:64])
}
