package accounts_summaries

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/listings/listing_type"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type AccountSummary struct {
	Version            AccountSummaryVersion  `json:"version"`
	FederationIdentity *addresses.Address     `json:"federationIdentity"`
	AccountIdentity    *addresses.Address     `json:"accountIdentity"`
	SalesTotal         uint64                 `json:"salesTotal"`
	SalesCount         uint64                 `json:"salesCount"`
	SalesAmount        uint64                 `json:"salesAmount"`
	PurchasesTotal     uint64                 `json:"purchasesTotal"`
	PurchasesCount     uint64                 `json:"purchasesCount"`
	PurchasesAmount    uint64                 `json:"purchasesAmount"`
	Validation         *validation.Validation `json:"validation"`
	Signer             *ownership.Ownership   `json:"signer"`
}

func (this *AccountSummary) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeValidation, includeOwnership, includeOwnershipSignature bool) {

	w.WriteUvarint(uint64(this.Version))
	this.FederationIdentity.Serialize(w)
	this.AccountIdentity.Serialize(w)
	w.WriteUvarint(this.SalesTotal)
	w.WriteUvarint(this.SalesCount)
	w.WriteUvarint(this.SalesAmount)

	w.WriteUvarint(this.PurchasesTotal)
	w.WriteUvarint(this.PurchasesCount)
	w.WriteUvarint(this.PurchasesAmount)

	if includeValidation {
		this.Validation.Serialize(w)
	}

	if includeOwnership {
		this.Signer.AdvancedSerialize(w, includeOwnershipSignature)
	}

}

func (this *AccountSummary) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true, true, true)
}

func (this *AccountSummary) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64
	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = AccountSummaryVersion(version)

	switch this.Version {
	case ACCOUNT_SUMMARY_VERSION:
	default:
		return errors.New("INVALID ACCOUNT VERSION")
	}

	this.FederationIdentity = &addresses.Address{}
	if err = this.FederationIdentity.Deserialize(r); err != nil {
		return
	}

	this.AccountIdentity = &addresses.Address{}
	if err = this.AccountIdentity.Deserialize(r); err != nil {
		return
	}

	if this.SalesTotal, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.SalesCount, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.SalesAmount, err = r.ReadUvarint(); err != nil {
		return
	}

	if this.PurchasesTotal, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.PurchasesCount, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.PurchasesAmount, err = r.ReadUvarint(); err != nil {
		return
	}

	this.Validation = &validation.Validation{}
	if err = this.Validation.Deserialize(r, this.GetMessageForSigningValidator); err != nil {
		return
	}

	this.Signer = &ownership.Ownership{}
	if err = this.Signer.Deserialize(r, this.GetMessageForSigningSigner); err != nil {
		return
	}

	return
}

func (this *AccountSummary) IsDeletable() bool {
	return false
}

func (this *AccountSummary) Validate() error {

	switch this.Version {
	case ACCOUNT_SUMMARY_VERSION:
	default:
		return errors.New("INVALID ACCOUNT SUMMARY VERSION")
	}

	return nil
}

func (this *AccountSummary) IsBetter(old *AccountSummary) bool {
	return old == nil || old.GetBetterScore() < this.GetBetterScore()
}

func (this *AccountSummary) GetBetterScore() uint64 {
	return this.Validation.Timestamp
}

func (this *AccountSummary) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false, false, false)
	return w.Bytes()
}

func (this *AccountSummary) GetMessageForSigningSigner() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, false)
	return w.Bytes()
}

func (this *AccountSummary) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator) {
		return errors.New("ACCOUNT SUMMARY VALIDATION FAILED")
	}
	if !this.Signer.Verify(this.GetMessageForSigningSigner) {
		return errors.New("ACCOUNT SUMMARY OWNERSHIP FAILED")
	}
	return nil
}

func (this *AccountSummary) GetScore(t listing_type.ListingType) float64 {
	if this == nil {
		return 0
	}
	if t == listing_type.LISTING_BUY {
		if this.PurchasesCount == 0 {
			return 0
		}
		return float64(this.PurchasesTotal) / float64(this.PurchasesCount) * float64(this.PurchasesAmount)
	} else {
		if this.SalesCount == 0 {
			return 0
		}
		return float64(this.SalesTotal) / float64(this.SalesCount) * float64(this.SalesAmount)
	}
}
