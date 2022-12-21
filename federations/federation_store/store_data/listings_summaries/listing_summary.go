package listings_summaries

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type ListingSummary struct {
	Version            ListingSummaryVersion  `json:"version"`
	FederationIdentity *addresses.Address     `json:"federationIdentity"`
	ListingIdentity    *addresses.Address     `json:"listingIdentity"`
	Total              uint64                 `json:"total"`
	Count              uint64                 `json:"count"`
	Amount             uint64                 `json:"amount"`
	Validation         *validation.Validation `json:"validation"`
	Signer             *ownership.Ownership   `json:"signer"`
}

func (this *ListingSummary) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeValidation, includeOwnership, includeOwnershipSignature bool) {

	w.WriteUvarint(uint64(this.Version))
	this.FederationIdentity.Serialize(w)
	this.ListingIdentity.Serialize(w)
	w.WriteUvarint(this.Total)
	w.WriteUvarint(this.Count)
	w.WriteUvarint(this.Amount)

	if includeValidation {
		this.Validation.Serialize(w)
	}

	if includeOwnership {
		this.Signer.AdvancedSerialize(w, includeOwnershipSignature)
	}

}

func (this *ListingSummary) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true, true, true)
}

func (this *ListingSummary) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64
	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = ListingSummaryVersion(version)

	switch this.Version {
	case LISTING_SUMMARY_VERSION:
	default:
		return errors.New("INVALID LISTING SUMMARY VERSION")
	}

	this.FederationIdentity = &addresses.Address{}
	if err = this.FederationIdentity.Deserialize(r); err != nil {
		return
	}

	this.ListingIdentity = &addresses.Address{}
	if err = this.ListingIdentity.Deserialize(r); err != nil {
		return
	}

	if this.Total, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Count, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Amount, err = r.ReadUvarint(); err != nil {
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

func (this *ListingSummary) IsDeletable() bool {
	return false
}

func (this *ListingSummary) Validate() error {

	switch this.Version {
	case LISTING_SUMMARY_VERSION:
	default:
		return errors.New("INVALID ACCOUNT SUMMARY VERSION")
	}

	return nil
}

func (this *ListingSummary) IsBetter(old *ListingSummary) bool {
	return old == nil || old.GetBetterScore() < this.GetBetterScore()
}

func (this *ListingSummary) GetBetterScore() uint64 {
	return this.Validation.Timestamp
}

func (this *ListingSummary) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false, false, false)
	return w.Bytes()
}

func (this *ListingSummary) GetMessageForSigningSigner() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, false)
	return w.Bytes()
}

func (this *ListingSummary) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator) {
		return errors.New("LISTING SUMMARY VALIDATION FAILED")
	}
	if !this.Signer.Verify(this.GetMessageForSigningSigner) {
		return errors.New("LISTING SUMMARY OWNERSHIP FAILED")
	}
	return nil
}

func (this *ListingSummary) GetScore() float64 {
	if this == nil || this.Count == 0 {
		return 0
	}
	return float64(this.Total) / float64(this.Count) * float64(this.Amount)
}
