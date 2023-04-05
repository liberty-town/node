package invoices

import (
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/config/globals"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations"
	"liberty-town/node/federations/moderator"
	pandora_pay_addresses "liberty-town/node/pandora-pay/addresses"
	pandora_pay_cryptography "liberty-town/node/pandora-pay/cryptography"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

// 发票明细
type Invoice struct {
	Version    InvoiceVersion        `json:"version" msgpack:"version"`
	Federation *addresses.Address    `json:"federation" msgpack:"federation"`
	Moderator  *addresses.Address    `json:"moderator" msgpack:"moderator"`
	Items      []*InvoiceItem        `json:"items" msgpack:"items"`
	Date       uint64                `json:"date" msgpack:"date"`
	Deadline   uint64                `json:"deadline" msgpack:"deadline"`
	Shipping   uint64                `json:"shipping" msgpack:"shipping"`
	Delivery   string                `json:"delivery" msgpack:"delivery"`
	Notes      string                `json:"notes" msgpack:"notes"`
	Buyer      *InvoiceBuyerAccount  `json:"buyer" msgpack:"buyer"`
	Seller     *InvoiceSellerAccount `json:"seller" msgpack:"seller"`
}

func (this *Invoice) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteUvarint(uint64(this.Version))
	this.Federation.Serialize(w)
	this.Moderator.Serialize(w)
	w.WriteByte(byte(len(this.Items)))
	for i := range this.Items {
		this.Items[i].Serialize(w)
	}
	w.WriteUvarint(this.Date)
	w.WriteUvarint(this.Deadline)
	w.WriteUvarint(this.Shipping)
	w.WriteString(this.Delivery)
	w.WriteString(this.Notes)
	this.Buyer.Serialize(w)
	this.Seller.Serialize(w)
}

func (this *Invoice) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	var n uint64
	var b byte

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	this.Version = InvoiceVersion(n)
	switch this.Version {
	case INVOICE_VERSION_0:
	default:
		errors.New("invalid invoice")
	}

	this.Federation = &addresses.Address{}
	if err = this.Federation.Deserialize(r); err != nil {
		return
	}

	this.Moderator = &addresses.Address{}
	if err = this.Moderator.Deserialize(r); err != nil {
		return
	}

	if b, err = r.ReadByte(); err != nil {
		return err
	}
	this.Items = make([]*InvoiceItem, b)
	for i := range this.Items {
		this.Items[i] = &InvoiceItem{}
		if err = this.Items[i].Deserialize(r); err != nil {
			return err
		}
	}

	if this.Date, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Deadline, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Shipping, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Delivery, err = r.ReadString(config.INVOICE_DELIVERY_MAX_LENGTH); err != nil {
		return
	}
	if this.Notes, err = r.ReadString(config.INVOICE_NOTES_MAX_LENGTH); err != nil {
		return
	}

	this.Buyer = &InvoiceBuyerAccount{}
	if err = this.Buyer.Deserialize(r); err != nil {
		return
	}

	this.Seller = &InvoiceSellerAccount{}
	if err = this.Seller.Deserialize(r); err != nil {
		return
	}

	return nil
}

func (this *Invoice) MessageForSignature() []byte {
	a := &Invoice{
		this.Version,
		this.Federation,
		this.Moderator,
		this.Items,
		this.Date,
		this.Deadline,
		this.Shipping,
		this.Delivery,
		this.Notes,
		&InvoiceBuyerAccount{
			this.Buyer.Address,
			this.Buyer.Nonce,
			this.Buyer.Multisig,
			this.Buyer.ConversionAsset,
			this.Buyer.ConversionAmount,
			nil,
		},
		&InvoiceSellerAccount{
			this.Seller.Address,
			this.Seller.Nonce,
			this.Seller.Multisig,
			this.Seller.Recipient,
			nil,
		},
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil
	}
	return b
}

func (this *Invoice) GetTotal() (uint64, error) {
	s := uint64(0)
	for i := range this.Items {
		s += this.Items[i].Price * this.Items[i].Quantity
	}
	return s, nil
}

func (this *Invoice) GetModerator() (*moderator.Moderator, error) {

	fed, _ := federations.FederationsDict.Load(this.Federation.Encoded)
	if fed == nil {
		return nil, errors.New("invoice federation was not found ")
	}

	mod := fed.FindModerator(this.Moderator.Encoded)
	if mod == nil {
		return nil, errors.New("invoice moderator was not found in federation")
	}

	return mod, nil
}

func (this *Invoice) ValidateBuyer(signature bool) error {

	if this.Buyer == nil {
		return errors.New("invoice buyer is null")
	}

	if len(this.Buyer.Nonce) != cryptography.HashSize {
		return errors.New("invoice buyer nonce is invalid")
	}
	if len(this.Buyer.Multisig) != pandora_pay_cryptography.PublicKeySize {
		return errors.New("invoice buyer multisig is invalid")
	}

	if signature {

		addr, err := addresses.CreateAddrFromSignature(this.MessageForSignature(), this.Buyer.Signature)
		if err != nil {
			return err
		}

		if !addr.Equals(this.Buyer.Address) {
			return errors.New("buyer address does not match")
		}

		if globals.Assets.Assets[this.Buyer.ConversionAsset] == nil {
			return errors.New("conversion asset was not found")
		}

		if this.Buyer.ConversionAmount == 0 {
			return errors.New("conversion amount is zero")
		}

	}

	return nil
}

func (this *Invoice) ValidateSeller(signature bool) error {

	if this.Seller == nil {
		return errors.New("invoice seller is null")
	}
	if len(this.Seller.Nonce) != cryptography.HashSize {
		return errors.New("invoice seller nonce is invalid")
	}

	recipientAddr, err := pandora_pay_addresses.DecodeAddr(this.Seller.Recipient)
	if err != nil {
		return err
	}
	if recipientAddr.Network != config.NETWORK_SELECTED {
		return errors.New("invoice invalid seller address network")
	}

	if len(this.Seller.Multisig) != pandora_pay_cryptography.PublicKeySize {
		return errors.New("invoice seller multisig is invalid")
	}

	if signature {

		addr, err := addresses.CreateAddrFromSignature(this.MessageForSignature(), this.Seller.Signature)
		if err != nil {
			return err
		}

		if !addr.Equals(this.Seller.Address) {
			return errors.New("seller address does not match")
		}
	}

	return nil
}

func (this *Invoice) Validate() error {

	switch this.Version {
	case INVOICE_VERSION_0:
	default:
		return errors.New("invoice version is invalid")
	}

	if fed, _ := federations.FederationsDict.Load(this.Federation.Encoded); fed == nil {
		return errors.New("invoice federation was not found")
	}

	if this.Moderator.Network != config.NETWORK_SELECTED {
		return errors.New("invoice moderator network is invalid")
	}

	if _, err := this.GetModerator(); err != nil {
		return err
	}

	if len(this.Items) == 0 {
		return errors.New("invoice does not contain any item")
	}

	if this.Deadline < 100 {
		return errors.New("deadline is less than 100 blocks")
	}
	if this.Deadline > 100000 {
		return errors.New("deadline can not exceed 100000 blocks")
	}

	for _, it := range this.Items {
		if err := it.Validate(); err != nil {
			return err
		}
	}

	if len(this.Delivery) < config.INVOICE_DELIVERY_MIN_LENGTH || len(this.Delivery) > config.INVOICE_DELIVERY_MAX_LENGTH {
		return errors.New("invoice delivery is invalid")
	}

	if len(this.Notes) > config.INVOICE_NOTES_MAX_LENGTH {
		return errors.New("invoice note is too large")
	}

	if this.Buyer == nil || this.Seller == nil {
		return errors.New("invoice has invalid buyer/seller data")
	}

	if this.Buyer.Address.Encoded == this.Seller.Address.Encoded {
		return errors.New("invoice contains identical buyer and seller")
	}

	total, err := this.GetTotal()
	if err != nil {
		return err
	}

	if total == 0 {
		return errors.New("total should be greater than zero")
	}

	if this.Buyer.Address.Network != config.NETWORK_SELECTED || this.Seller.Address.Network != config.NETWORK_SELECTED {
		return errors.New("invoice invalid network")
	}

	return nil
}
