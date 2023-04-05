package validation

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/cryptography"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation/validation_type"
)

type Validation struct {
	Version   validation_type.ValidationVersion `json:"version" msgpack:"version"`
	Nonce     []byte                            `json:"nonce" msgpack:"nonce"`
	Timestamp uint64                            `json:"timestamp" msgpack:"timestamp"`
	Signature []byte                            `json:"signature" msgpack:"signature"`
	Address   *addresses.Address                `json:"address" msgpack:"address"`
}

func (this *Validation) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteUvarint(uint64(this.Version))
	if this.Version == validation_type.VALIDATION_VERSION_V0 {
		w.Write(this.Nonce)
		w.WriteUvarint(this.Timestamp)
		w.Write(this.Signature)
	}
}

func (this *Validation) SerializeToBytes() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.Serialize(w)
	return w.Bytes()
}

func (this *Validation) Deserialize(r *advanced_buffers.BufferReader, getMessage, getExtraInfo func() []byte) (err error) {
	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	this.Version = validation_type.ValidationVersion(n)

	if this.Version == validation_type.VALIDATION_VERSION_V0 {
		if this.Nonce, err = r.ReadBytes(validation_type.VALIDATOR_NONCE_SIZE); err != nil {
			return
		}
		if this.Timestamp, err = r.ReadUvarint(); err != nil {
			return
		}
		if this.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
		if this.Address, err = addresses.CreateAddrFromSignature(this.GetMessageToValidator(getMessage, getExtraInfo), this.Signature); err != nil {
			return
		}
	}
	return
}

func (this *Validation) GetMessageToValidator(getMessage, getExtraInfo func() []byte) []byte {
	message := getMessage()

	w := advanced_buffers.NewBufferWriter()
	w.WriteUvarint(uint64(this.Version))
	if getExtraInfo != nil {
		w.Write(getExtraInfo())
	}
	w.Write(cryptography.SHA3(message))
	w.WriteUvarint(uint64(len(message)))
	w.Write(this.Nonce)
	w.WriteUvarint(this.Timestamp)
	return w.Bytes()
}

func (this *Validation) Verify(getMessage, getExtraInfo func() []byte) bool {
	if this == nil {
		return false
	}

	switch this.Version {
	case validation_type.VALIDATION_VERSION_V0:
		return this.Address.VerifySignedMessage(this.GetMessageToValidator(getMessage, getExtraInfo), this.Signature[0:64])
	default:
		return false
	}
}

func (this *Validation) Validate() error {
	switch this.Version {
	case validation_type.VALIDATION_VERSION_V0:
		if this.Timestamp == 0 {
			return errors.New("validation timestamp is invalid")
		}
		if len(this.Signature) != cryptography.SignatureSize {
			return errors.New("validation signature size is invalid")
		}
		if this.Address == nil {
			return errors.New("validation address is null")
		}
	default:
		return errors.New("invalid Validation Type")
	}
	return nil
}
