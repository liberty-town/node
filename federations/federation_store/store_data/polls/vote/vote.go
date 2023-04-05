package vote

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type Vote struct {
	FederationIdentity *addresses.Address     `json:"federation" msgpack:"federation"` //not serialized
	Identity           *addresses.Address     `json:"identity" msgpack:"identity"`     //not serialized
	Upvotes            uint64                 `json:"up" msgpack:"up"`
	Downvotes          uint64                 `json:"down" msgpack:"down"`
	Validation         *validation.Validation `json:"validation" msgpack:"validation"`
}

func (this *Vote) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeUnserializedData, includeValidation bool) {

	if includeUnserializedData {
		this.FederationIdentity.Serialize(w)
		this.Identity.Serialize(w)
	}

	w.WriteUvarint(this.Upvotes)
	w.WriteUvarint(this.Downvotes)
	if includeValidation {
		this.Validation.Serialize(w)
	}
}

func (this *Vote) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, false, true)
}

func (this *Vote) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.Upvotes, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Downvotes, err = r.ReadUvarint(); err != nil {
		return
	}
	this.Validation = &validation.Validation{}
	return this.Validation.Deserialize(r, this.GetMessageForSigningValidator, this.GetExtraMessageForValidator)
}

func (this *Vote) GetExtraMessageForValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	w.WriteUvarint(this.Upvotes)
	w.WriteUvarint(this.Downvotes)
	return w.Bytes()
}

func (this *Vote) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.FederationIdentity.Serialize(w)
	this.Identity.Serialize(w)
	return w.Bytes()
}

func (this *Vote) IsDeletable() bool {
	return false
}

func (this *Vote) Validate() error {
	if this.FederationIdentity == nil {
		return errors.New("federation is not set")
	}
	if this.Identity == nil {
		return errors.New("identity is not set")
	}
	return nil
}

func (this *Vote) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator, this.GetExtraMessageForValidator) {
		return errors.New("vote validation signature failed")
	}
	return nil
}

func (this *Vote) GetBetterScore() uint64 {
	return this.Upvotes + this.Downvotes
}
