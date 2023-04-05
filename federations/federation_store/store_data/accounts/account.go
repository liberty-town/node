package accounts

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type Account struct {
	Version            AccountVersion         `json:"version" msgpack:"version"`
	FederationIdentity *addresses.Address     `json:"federation"  msgpack:"federation"`
	Identity           *addresses.Address     `json:"identity" msgpack:"identity"`
	Description        string                 `json:"description" msgpack:"description"`
	Country            uint64                 `json:"country" msgpack:"country"`
	Validation         *validation.Validation `json:"validation" msgpack:"validation"`
	Ownership          *ownership.Ownership   `json:"ownership" msgpack:"ownership"`
}

func (this *Account) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeValidation, includeOwnership, includeOwnershipSignature bool) {

	w.WriteUvarint(uint64(this.Version))
	this.FederationIdentity.Serialize(w)
	this.Identity.Serialize(w)

	w.WriteString(this.Description)
	w.WriteUvarint(this.Country)

	if includeValidation {
		this.Validation.Serialize(w)
	}

	if includeValidation && includeOwnership {
		this.Ownership.AdvancedSerialize(w, includeOwnershipSignature)
	}

}

func (this *Account) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true, true, true)
}

func (this *Account) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64
	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = AccountVersion(version)

	switch this.Version {
	case ACCOUNT_VERSION:
	default:
		return errors.New("invalid account version")
	}

	this.FederationIdentity = &addresses.Address{}
	if err = this.FederationIdentity.Deserialize(r); err != nil {
		return
	}

	this.Identity = &addresses.Address{}
	if err = this.Identity.Deserialize(r); err != nil {
		return
	}

	if this.Description, err = r.ReadString(5 * 1024); err != nil {
		return
	}

	if this.Country, err = r.ReadUvarint(); err != nil {
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

	if !this.Identity.Equals(this.Ownership.Address) {
		return errors.New("identity does not match")
	}

	return
}

func (this *Account) IsDeletable() bool {
	return false
}

func (this *Account) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false, false, false)
	return w.Bytes()
}

func (this *Account) GetMessageForSigningOwnership() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, false)
	return w.Bytes()
}

func (this *Account) Validate() error {

	switch this.Version {
	case ACCOUNT_VERSION:
	default:
		return errors.New("invalid account version")
	}

	if len(this.Description) > 5*1024 {
		return errors.New("invalid account description")
	}

	if this.Country > config.COUNTRY_CODE_MAX {
		return errors.New("invalid account country")
	}

	if this.Ownership == nil || !this.Ownership.Address.Equals(this.Identity) {
		return errors.New("invalid account identity")
	}

	return nil
}

func (this *Account) IsBetter(old *Account) bool {
	return old == nil || old.GetBetterScore() < this.GetBetterScore()
}

func (this *Account) GetBetterScore() uint64 {
	return this.Validation.Timestamp
}

func (this *Account) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator, nil) {
		return errors.New("account validation failed")
	}
	if !this.Ownership.Verify(this.GetMessageForSigningOwnership) {
		return errors.New("account ownership failed")
	}
	return nil
}
