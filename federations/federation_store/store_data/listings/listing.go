package listings

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/listings/listing_type"
	"liberty-town/node/federations/federation_store/store_data/listings/offer"
	"liberty-town/node/federations/federation_store/store_data/listings/shipping"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/advanced_strings"
	"liberty-town/node/validator/validation"
	"net/url"
	"strings"
)

type Listing struct {
	Version            ListingVersion           `json:"version" msgpack:"version"`
	FederationIdentity *addresses.Address       `json:"federation" msgpack:"federation"`
	Nonce              []byte                   `json:"nonce" msgpack:"nonce"`
	Identity           *addresses.Address       `json:"identity" msgpack:"identity"`
	Type               listing_type.ListingType `json:"type" msgpack:"type"`
	Title              string                   `json:"title" msgpack:"title"`
	Description        string                   `json:"description" msgpack:"description"`
	Categories         []uint64                 `json:"categories" msgpack:"categories"`
	Images             []string                 `json:"images" msgpack:"images"`
	QuantityUnlimited  bool                     `json:"quantityUnlimited" msgpack:"quantityUnlimited"`
	QuantityAvailable  uint64                   `json:"quantityAvailable" msgpack:"quantityAvailable"`
	ShipsFrom          uint64                   `json:"shipsFrom" msgpack:"shipsFrom"`
	ShipsTo            []uint64                 `json:"shipsTo" msgpack:"shipsTo"`
	Offers             []*offer.Offer           `json:"offers" msgpack:"offers"`
	Shipping           []*shipping.Shipping     `json:"shipping"  msgpack:"shipping"`
	Validation         *validation.Validation   `json:"validation" msgpack:"validation"`
	Publisher          *ownership.Ownership     `json:"publisher" msgpack:"publisher"`
	Ownership          *ownership.Ownership     `json:"ownership" msgpack:"ownership"`
}

func (this *Listing) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeValidation, includePublisher, includePublisherSignature bool, includeOwnership, includeOwnershipSignature bool) {

	w.WriteUvarint(uint64(this.Version))

	this.FederationIdentity.Serialize(w)
	w.Write(this.Nonce)
	this.Identity.Serialize(w)
	w.WriteByte(byte(this.Type))

	w.WriteString(this.Title)
	w.WriteString(this.Description)
	w.WriteByte(byte(len(this.Categories)))

	for i := range this.Categories {
		w.WriteUvarint(this.Categories[i])
	}

	w.WriteByte(byte(len(this.Images)))
	for i := range this.Images {
		w.WriteString(this.Images[i])
	}

	w.WriteBool(this.QuantityUnlimited)
	if !this.QuantityUnlimited {
		w.WriteUvarint(this.QuantityAvailable)
	}
	w.WriteUvarint(this.ShipsFrom)
	w.WriteByte(byte(len(this.ShipsTo)))
	for i := range this.ShipsTo {
		w.WriteUvarint(this.ShipsTo[i])
	}

	w.WriteByte(byte(len(this.Offers)))
	for i := range this.Offers {
		this.Offers[i].Serialize(w)
	}

	w.WriteByte(byte(len(this.Shipping)))
	for i := range this.Shipping {
		this.Shipping[i].Serialize(w)
	}

	if includeValidation {
		this.Validation.Serialize(w)
	}

	if includeValidation && includePublisher {
		this.Publisher.AdvancedSerialize(w, includePublisherSignature)
	}

	if includeValidation && includePublisher && includeOwnership {
		this.Ownership.AdvancedSerialize(w, includeOwnershipSignature)
	}

}

func (this *Listing) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true, true, true, true, true)
}

func (this *Listing) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64
	var b byte

	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = ListingVersion(version)

	switch this.Version {
	case LISTING_VERSION:
	default:
		return errors.New("invalid listing version")
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
	if b, err = r.ReadByte(); err != nil {
		return
	}

	switch listing_type.ListingType(b) {
	case listing_type.LISTING_BUY, listing_type.LISTING_SELL:
	default:
		return errors.New("invalid listing type")
	}

	this.Type = listing_type.ListingType(b)

	if this.Title, err = r.ReadString(config.LISTING_TITLE_MAX_LENGTH); err != nil {
		return
	}

	if this.Description, err = r.ReadString(config.LISTING_DESCRIPTION_MAX_LENGTH); err != nil {
		return
	}

	if b, err = r.ReadByte(); err != nil {
		return
	}
	if b > config.LISTING_CATEGORIES_MAX_COUNT {
		return errors.New("invalid number of categories")
	}
	this.Categories = make([]uint64, b)
	for i := range this.Categories {
		if this.Categories[i], err = r.ReadUvarint(); err != nil {
			return
		}
	}

	if b, err = r.ReadByte(); err != nil {
		return
	}
	if b > config.LISTING_IMAGES_MAX_COUNT {
		return errors.New("invalid number of images")
	}

	this.Images = make([]string, b)
	for i := range this.Images {
		if this.Images[i], err = r.ReadString(config.LISTING_IMAGE_MAX_LENGTH); err != nil {
			return
		}
	}

	if this.QuantityUnlimited, err = r.ReadBool(); err != nil {
		return
	}

	if this.QuantityUnlimited {
		this.QuantityAvailable = 0
	} else {
		if this.QuantityAvailable, err = r.ReadUvarint(); err != nil {
			return
		}
	}

	if this.ShipsFrom, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.ShipsFrom >= config.COUNTRY_CODE_MAX {
		return errors.New("shipping from code is invalid")
	}

	if b, err = r.ReadByte(); err != nil {
		return
	}
	if b == 0 || b > config.LISTING_SHIPPING_TO_MAX_COUNT {
		return errors.New("listing shipping to length is invalid")
	}
	this.ShipsTo = make([]uint64, b)
	for i := range this.ShipsTo {
		if this.ShipsTo[i], err = r.ReadUvarint(); err != nil {
			return
		}
		if this.ShipsTo[i] >= config.COUNTRY_CODE_MAX {
			return errors.New("shipping to code is invalid")
		}
	}

	if b, err = r.ReadByte(); err != nil {
		return
	}
	if b > config.LISTING_OFFERS_MAX_COUNT {
		return errors.New("listings invalid number of offers")
	}
	this.Offers = make([]*offer.Offer, b)
	for i := range this.Offers {
		this.Offers[i] = &offer.Offer{}
		if err = this.Offers[i].Deserialize(r); err != nil {
			return
		}
	}

	if b, err = r.ReadByte(); err != nil {
		return
	}
	if b > config.LISTING_SHIPPING_MAX_COUNT {
		return errors.New("listings invalid number of shipping options")
	}
	this.Shipping = make([]*shipping.Shipping, b)
	for i := range this.Shipping {
		this.Shipping[i] = &shipping.Shipping{}
		if err = this.Shipping[i].Deserialize(r); err != nil {
			return
		}
	}

	this.Validation = &validation.Validation{}
	if err = this.Validation.Deserialize(r, this.GetMessageForSigningValidator, nil); err != nil {
		return
	}

	this.Publisher = &ownership.Ownership{}
	if err = this.Publisher.Deserialize(r, this.GetMessageForSigningPublisher); err != nil {
		return
	}

	this.Ownership = &ownership.Ownership{}
	if err = this.Ownership.Deserialize(r, this.GetMessageForSigningOwnership); err != nil {
		return
	}

	return
}

func (this *Listing) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false, false, false, false, false)
	return w.Bytes()
}

func (this *Listing) GetMessageForSigningPublisher() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, false, false, false)
	return w.Bytes()
}

func (this *Listing) GetMessageForSigningOwnership() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, true, true, false)
	return w.Bytes()
}

func (this *Listing) IsDeletable() bool {
	return false
}

func (this *Listing) Validate() error {

	switch this.Version {
	case LISTING_VERSION:
	default:
		return errors.New("invalid listing")
	}

	switch this.Type {
	case listing_type.LISTING_BUY, listing_type.LISTING_SELL:
	default:
		return errors.New("invalid type")
	}

	if len(this.Nonce) != cryptography.HashSize {
		return errors.New("invalid listing nonce")
	}
	if len(this.Title) > config.LISTING_TITLE_MAX_LENGTH {
		return errors.New("listing title is too large")
	}
	if len(this.Title) < config.LISTING_TITLE_MIN_LENGTH {
		return errors.New("listing title is too short")
	}
	if len(this.Description) > config.LISTING_DESCRIPTION_MAX_LENGTH {
		return errors.New("listing description is invalid")
	}
	if len(this.Categories) > config.LISTING_CATEGORIES_MAX_COUNT {
		return errors.New("listing contains too many categories")
	}

	dict := make(map[uint64]bool)
	for _, x := range this.Categories {
		dict[x] = true
	}
	if len(dict) != len(this.Categories) {
		return errors.New("listing categories repeated itself")
	}

	if len(this.Images) > config.LISTING_IMAGES_MAX_COUNT {
		return errors.New("listing contains too many images")
	}
	for i := range this.Images {
		if len(this.Images[i]) > config.LISTING_IMAGE_MAX_LENGTH {
			return errors.New("listing image invalid length")
		}
		if _, err := url.ParseRequestURI(this.Images[i]); err != nil {
			return errors.New("invalid image url")
		}
	}

	if this.QuantityUnlimited && this.QuantityAvailable != 0 {
		return errors.New("listing invalid quantity available")
	}

	if this.ShipsFrom >= config.COUNTRY_CODE_MAX {
		return errors.New("listing shipping from code is invalid")
	}
	if len(this.ShipsTo) == 0 || len(this.ShipsTo) > config.LISTING_SHIPPING_TO_MAX_COUNT {
		return errors.New("listing shipping to length is invalid")
	}

	dict = make(map[uint64]bool)
	for i, x := range this.ShipsTo {
		if x > config.COUNTRY_CODE_MAX {
			return errors.New("listing shipping to code is invalid")
		}

		if (x == 243 || x == 244) && (i > 0) {
			return errors.New("invalid shipping to")
		} else if dict[243] || dict[244] {
			return errors.New("don't specify other countries")
		}

		dict[x] = true

	}
	if len(dict) != len(this.ShipsTo) {
		return errors.New("listing ships to repeated itself")
	}

	if len(this.Offers) == 0 || len(this.Offers) > config.LISTING_OFFERS_MAX_COUNT {
		return errors.New("listing offers invalid length")
	}
	for i := range this.Offers {
		if err := this.Offers[i].Validate(); err != nil {
			return err
		}
	}

	if len(this.Shipping) > config.LISTING_SHIPPING_TO_MAX_COUNT {
		return errors.New("listing shipping offers length is invalid")
	}

	if len(this.Shipping) == 0 && this.Type == listing_type.LISTING_SELL {
		return errors.New("listing should have at least 1 shipping offer")
	}

	for i := range this.Shipping {
		if err := this.Shipping[i].Validate(); err != nil {
			return err
		}
	}

	if this.Ownership == nil || !this.Ownership.Address.Equals(this.Identity) {
		return errors.New("listing ownership identity does not match")
	}

	return nil
}

func (this *Listing) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator, nil) {
		return errors.New("listing validation signature failed")
	}
	if !this.Publisher.Verify(this.GetMessageForSigningPublisher) {
		return errors.New("listing publisher signature failed")
	}
	if !this.Ownership.Verify(this.GetMessageForSigningOwnership) {
		return errors.New("listing ownership signature failed")
	}
	return nil
}

func (this *Listing) IsBetter(old *Listing) bool {
	return old == nil || old.GetBetterScore() < this.GetBetterScore()
}

func (this *Listing) GetBetterScore() uint64 {
	return this.Validation.Timestamp
}

func (this *Listing) GetWords() []string {

	var words []string
	if this.Ownership.Timestamp < 1668391857 {
		words = advanced_strings.Splitter(this.Title, " ")
	} else {
		words = advanced_strings.SplitterSeparators(this.Title)
	}

	list := []string{}
	for _, word := range words {
		if len(word) < 4 {
			continue
		}
		if len(list) > 10 {
			break
		}
		list = append(list, strings.ToLower(word))
	}
	return list
}

func GetScore(listingSummaryScore, accountSummaryScore float64) float64 {
	return (float64(3)*listingSummaryScore + accountSummaryScore) / 4
}
