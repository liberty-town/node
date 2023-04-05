package reviews

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type Review struct {
	Version            ReviewVersion          `json:"version" msgpack:"version"`
	FederationIdentity *addresses.Address     `json:"federation" msgpack:"federation"`
	Nonce              []byte                 `json:"nonce" msgpack:"nonce"`
	Identity           *addresses.Address     `json:"identity" msgpack:"identity"`
	ListingIdentity    *addresses.Address     `json:"listing" msgpack:"listing"`
	AccountIdentity    *addresses.Address     `json:"account" msgpack:"account"`
	Text               string                 `json:"text" msgpack:"text"`
	Score              byte                   `json:"score" msgpack:"score"`
	Amount             uint64                 `json:"amount" msgpack:"amount"`
	Validation         *validation.Validation `json:"validation" msgpack:"validation"`
	Ownership          *ownership.Ownership   `json:"ownership" msgpack:"ownership"`
	Signer             *ownership.Ownership   `json:"signer" msgpack:"signer"`
}

func (this *Review) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeValidation, includeOwnership, includeOwnershipSignature, includeSigner, includeSignerSignature bool) {
	w.WriteUvarint(uint64(this.Version))
	this.FederationIdentity.Serialize(w)
	w.Write(this.Nonce)
	this.Identity.Serialize(w)
	this.ListingIdentity.Serialize(w)
	this.AccountIdentity.Serialize(w)
	w.WriteString(this.Text)
	w.WriteByte(this.Score)
	w.WriteUvarint(this.Amount)
	if includeValidation {
		this.Validation.Serialize(w)
	}
	if includeOwnership {
		this.Ownership.AdvancedSerialize(w, includeOwnershipSignature)
	}
	if includeSigner {
		this.Signer.AdvancedSerialize(w, includeSignerSignature)
	}
}

func (this *Review) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true, true, true, true, true)
}

func (this *Review) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	var version uint64
	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = ReviewVersion(version)

	switch this.Version {
	case REVIEW_VERSION:
	default:
		return errors.New("review version")
	}

	this.FederationIdentity = &addresses.Address{}
	if err = this.FederationIdentity.Deserialize(r); err != nil {
		return
	}
	if this.Nonce, err = r.ReadBytes(cryptography.HashSize); err != nil {
		return
	}

	this.Identity = &addresses.Address{}
	if err = this.Identity.Deserialize(r); err != nil {
		return
	}

	this.ListingIdentity = &addresses.Address{}
	if err = this.ListingIdentity.Deserialize(r); err != nil {
		return
	}

	this.AccountIdentity = &addresses.Address{}
	if err = this.AccountIdentity.Deserialize(r); err != nil {
		return
	}
	if this.Text, err = r.ReadString(config.REVIEW_TITLE_MAX_LENGTH); err != nil {
		return
	}
	if this.Score, err = r.ReadByte(); err != nil {
		return
	}
	if this.Score > 5 {
		return errors.New("invalid review score")
	}
	if this.Amount, err = r.ReadUvarint(); err != nil {
		return
	}

	this.Validation = &validation.Validation{}
	if err = this.Validation.Deserialize(r, this.GetMessageForSigningValidator, nil); err != nil {
		return
	}

	this.Ownership = &ownership.Ownership{}
	if err = this.Ownership.Deserialize(r, this.GetMessageForSigningOwnership); err != nil {
		return
	}

	if this.Ownership.Address.Encoded != this.Identity.Encoded {
		return errors.New("review ownership identity does not match")
	}

	this.Signer = &ownership.Ownership{}
	if err = this.Signer.Deserialize(r, this.GetMessageForSigningSigner); err != nil {
		return
	}

	return
}

func (this *Review) IsDeletable() bool {
	return false
}

func (this *Review) Validate() error {

	switch this.Version {
	case REVIEW_VERSION:
	default:
		return errors.New("invalid review version")
	}

	if len(this.Text) > config.REVIEW_TITLE_MAX_LENGTH {
		return errors.New("invalid review text")
	}
	if this.Score > 5 {
		return errors.New("invalid review score")
	}

	if this.Ownership == nil || this.Ownership.Address.Encoded != this.Identity.Encoded {
		return errors.New("invalid review identity - mismatch")
	}

	return nil
}

func (this *Review) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false, false, false, false, false)
	return w.Bytes()
}

func (this *Review) GetMessageForSigningOwnership() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, false, false, false)
	return w.Bytes()
}

func (this *Review) GetMessageForSigningSigner() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, true, true, false)
	return w.Bytes()
}

func (this *Review) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator, nil) {
		return errors.New("ACCOUNT VALIDATION FAILED")
	}
	if !this.Signer.Verify(this.GetMessageForSigningSigner) {
		return errors.New("ACCOUNT OWNERSHIP FAILED")
	}
	return nil
}

func (this *Review) IsBetter(old *Review) bool {
	return old == nil || old.GetBetterScore() < this.GetBetterScore()
}

func (this *Review) GetBetterScore() uint64 {
	return this.Validation.Timestamp
}
