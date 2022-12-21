package moderator

import (
	"errors"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type Moderator struct {
	Version              ModeratorVersion       `json:"version"`
	ConditionalPublicKey []byte                 `json:"conditionalPublicKey"`
	RewardAddresses      []string               `json:"rewardAddresses"`
	Fee                  uint64                 `json:"fee"` //divided by 100000
	Validation           *validation.Validation `json:"validation"`
	Ownership            *ownership.Ownership   `json:"ownership"`
}

func (this *Moderator) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeValidation, includeOwnership, includeOwnershipSignature bool) {
	w.WriteUvarint(uint64(this.Version))
	w.Write(this.ConditionalPublicKey)
	w.WriteByte(byte(len(this.RewardAddresses)))
	for i := range this.RewardAddresses {
		w.WriteString(this.RewardAddresses[i])
	}
	w.WriteUvarint(this.Fee)
	if includeValidation {
		this.Validation.Serialize(w)
	}
	if includeValidation && includeOwnership {
		this.Ownership.AdvancedSerialize(w, includeOwnershipSignature)
	}
}

func (this *Moderator) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true, true, true)
}

func (this *Moderator) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64
	var n byte

	if version, err = r.ReadUvarint(); err != nil {
		return
	}

	switch ModeratorVersion(version) {
	case MODERATOR_PANDORA:
		if this.ConditionalPublicKey, err = r.ReadBytes(33); err != nil {
			return
		}
		if n, err = r.ReadByte(); err != nil {
			return
		}
		this.RewardAddresses = make([]string, n)
		for i := range this.RewardAddresses {
			if this.RewardAddresses[i], err = r.ReadString(255); err != nil {
				return
			}
		}
		if this.Fee, err = r.ReadUvarint(); err != nil {
			return
		}
	default:
		return errors.New("invalid Moderator Version")
	}
	this.Validation = &validation.Validation{}
	if err = this.Validation.Deserialize(r, this.GetMessageForSigningValidator); err != nil {
		return
	}

	this.Ownership = &ownership.Ownership{}
	if err = this.Ownership.Deserialize(r, this.GetMessageForSigningOwnership); err != nil {
		return
	}

	return
}

func (this *Moderator) Validate() error {

	switch this.Version {
	case MODERATOR_PANDORA:
		if len(this.ConditionalPublicKey) != 33 {
			return errors.New("invalid Conditional Public Key")
		}
		if this.Fee > 100000 {
			return errors.New("invalid fee")
		}
		if this.Fee > 0 && len(this.RewardAddresses) == 0 {
			return errors.New("moderator requries reward addresses specified")
		}
	default:
		return errors.New("invalid Moderator Version")
	}

	return nil
}

func (this *Moderator) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator) {
		return errors.New("listing validation signature failed")
	}
	if !this.Ownership.Verify(this.GetMessageForSigningOwnership) {
		return errors.New("listing ownership signature failed")
	}
	return nil
}

func (this *Moderator) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false, false, false)
	return w.Bytes()
}

func (this *Moderator) GetMessageForSigningOwnership() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, false)
	return w.Bytes()
}
